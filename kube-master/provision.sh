echo "Become root"
sudo su
echo "Install docker..."
apt-get update && \
apt-get install -y docker.io apt-transport-https curl && \
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main 
EOF

echo "Do autoremove..."
apt-get autoremove && \

echo "Turn off swap permanently"
swapoff -a && \
sudo sed -i '/ swap / s/^\(.*\)$/#\1/g' /etc/fstab && \

echo "Install kubernetes"
apt-get update && \
apt-get install -y kubelet kubeadm kubectl kubernetes-cni && \

IPADDR=`ip address show enp0s8 | grep 'inet ' | sed -e 's/^.*inet //' -e 's/\/.*$//'`
echo This VM has IP address $IPADDR

kubeadm init --apiserver-advertise-address=$IPADDR

exit
mkdir -p $HOME/.kube && \
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config && \
sudo chown $(id -u):$(id -g) $HOME/.kube/config && \
sysctl net.bridge.bridge-nf-call-iptables=1 && \

# Install pod network
kubectl apply -f https://cloud.weave.works/k8s/net?k8s-version=$(kubectl version | base64 | tr -d '\n') && \
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/master/src/deploy/recommended/kubernetes-dashboard.yaml && \
kubectl create serviceaccount cluster-admin-dashboard-sa && \
kubectl create clusterrolebinding cluster-admin-dashboard-sa --clusterrole=cluster-admin --serviceaccount=default:cluster-admin-dashboard-sa