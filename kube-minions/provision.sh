
sudo apt-get update
sudo apt-get install -y docker.io apt-transport-https curl
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main
EOF
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
apt-mark hold kubelet kubeadm kubectl
sudo su
bash
kubeadm join 192.168.33.10:6443 --token 55pd0c.pbuq5yq121owd211 --discovery-token-ca-cert-hash sha256:5c32a7324400381d036bc5ee812b6123521fc400d940e4288bf902cfeecfa1f6