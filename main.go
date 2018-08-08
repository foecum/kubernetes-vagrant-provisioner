package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"text/template"
)

var (
	showVersion bool
	destroy     bool
	version     = "devel"
)

func init() {
	flag.BoolVar(&destroy, "destroy", false, "Destroys the cluster")
	flag.BoolVar(&showVersion, "v", false, "print version number")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] URL\n\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "OPTIONS:")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()

	if showVersion {
		fmt.Printf("%s %s (runtime: %s)\n", os.Args[0], version, runtime.Version())
		os.Exit(0)
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	if destroy {
		err := destroyNodes(dir)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	// Start master
	logs, err := startMaster(dir)
	if logs == "" {
		log.Fatalf("Master was not started\n")
	}

	joinToken := getKubeClusterJoinToken(logs)
	log.Printf("Cluster Join Token: %s\n", joinToken)

	err = startMinions(joinToken, dir)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}

func startMaster(pwd string) (string, error) {
	log.Println("Starting master...")
	err := os.Chdir(pwd + "/kube-master")
	if err != nil {
		return "", nil
	}
	log.Println("Starting vagrant...")

	logs, err := vagrantUp()
	if err != nil {
		return "", err
	}
	return logs, nil
}

// Token for nodes to use to join cluster
type Token struct {
	JoinToken string
}

func destroyNodes(pwd string) error {
	os.Chdir(pwd + "/kube-minions")
	_, err := vagrantDestroy()
	if err != nil {
		return err
	}

	os.Chdir(pwd + "/kube-master")
	_, err = vagrantDestroy()
	if err != nil {
		return err
	}
	return nil
}

func startMinions(joinCmd, pwd string) error {
	tmpl := `
sudo apt-get update
sudo apt-get install -y docker.io apt-transport-https curl
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main
EOF
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
apt-mark hold kubelet kubeadm kubectl
sudo sudo
bash\n` + joinCmd

	t := template.New("Master logs")
	t, err := t.Parse(tmpl)
	if err != nil {
		return err
	}

	err = os.Chdir(pwd + "/kube-minions")
	if err != nil {
		return err
	}
	f, err := os.Create("provision.sh")

	if err != nil {
		return err
	}
	err = t.Execute(f, Token{JoinToken: joinCmd})

	if err != nil {
		return err
	}
	f.Close()

	_, err = vagrantUp()
	if err != nil {
		return err
	}
	return nil
}

func vagrantDestroy() (string, error) {
	cmd := exec.Command("vagrant", "destroy", "-f")
	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	cmd.Start()

	go func() {
		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
	}()

	go func() {
		stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)
	}()

	err := cmd.Wait()
	if err != nil {
		return "", err
	}
	if errStdout != nil || errStderr != nil {
		return "", fmt.Errorf("failed to capture stdout or stderr")
	}
	outStr, errStr := string(stdout), string(stderr)
	log.Printf("%s", errStr)

	return outStr, nil
}

func vagrantUp() (string, error) {
	cmd := exec.Command("vagrant", "up")
	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	cmd.Start()

	go func() {
		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
	}()

	go func() {
		stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)
	}()

	err := cmd.Wait()
	if err != nil {
		return "", err
	}
	if errStdout != nil || errStderr != nil {
		return "", fmt.Errorf("failed to capture stdout or stderr")
	}
	outStr, errStr := string(stdout), string(stderr)
	log.Printf("%s", errStr)

	return outStr, nil
}

func getKubeClusterJoinToken(logs string) string {
	var re = regexp.MustCompile(`(?m)(kubeadm)\s(join)\s([0-9]+\.)+([0-9]+:)([0-9]+)\s(--token)\s([a-zA-Z0-9\.]+)\s(--discovery-token-ca-cert-hash)\s(sha256:[a-zA-Z0-9]+)`)
	var token string
	for _, match := range re.FindAllString(logs, -1) {
		token = match
	}
	return token
}

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {

			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}
