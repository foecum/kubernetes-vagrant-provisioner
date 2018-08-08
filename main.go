package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"text/template"
)

var (
	// Command line flags.
	//minions     int
	showVersion bool
	// master     int
	// masterPath string
	// masterOnly bool

	version = "devel" // for -v flag, updated during the release process with -ldflags=-X=main.version=...
)

func init() {
	//flag.IntVar(&minions, "minions", 1, "Number of minions to create")
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

	// Start master
	logs := startMaster(dir)
	if logs == "" {
		log.Fatalf("Master was not started\n")
	}

}

func startMaster(pwd string) string {
	exec.Command("cd", pwd, "/kube-master")
	vagrantUp := exec.Command("vagrant", "up")

	b, err := vagrantUp.CombinedOutput()
	if err != nil {
		log.Printf("Error: %v\n", err)
		return ""
	}
	exec.Command("cd", " ..")
	return string(b)
}

// Token for nodes to use to join cluster
type Token struct {
	JoinToken string
}

func startMinions(joinCmd, pwd string) {
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
	bash
	{{.joinToken}}`
	t := template.New("Master logs")
	t, err := t.Parse(tmpl)
	if err != nil {
		log.Fatalf("Failed to parse template.\n")
	}

	exec.Command("cd", pwd, "/kube-minions")
	f, err := os.Create("provision.sh")

	if err != nil {
		log.Fatalf("Cannot create provision file.\n")
	}
	err = t.Execute(f, Token{JoinToken: joinCmd})

	if err != nil {
		log.Fatalf("Cannot write to provision file.\n")
	}
	f.Close()

	exec.Command("cd", pwd, "/kube-minions")
	vagrantUp := exec.Command("vagrant", "up")

	_, err = vagrantUp.CombinedOutput()
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
	exec.Command("cd", " ..")
}

func getKubeClusterJoinToken(logs string) string {
	var re = regexp.MustCompile(`(?m)(kubeadm)\s(join)\s([0-9]+\.)+([0-9]+:)([0-9]+)\s(--token)\s([a-zA-Z0-9\.]+)\s(--discovery-token-ca-cert-hash)\s(sha256:[a-zA-Z0-9]+)`)
	var token string
	for _, match := range re.FindAllString(logs, -1) {
		token = match
	}
	return token
}
