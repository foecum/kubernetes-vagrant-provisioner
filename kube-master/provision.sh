apt-get update && \
apt-get install -y docker.io apt-transport-https curl && \
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
cat <<EOF >/etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main 
EOF
apt-get update && \
apt-get install -y kubelet kubeadm kubectl && \
apt-mark hold kubelet kubeadm kubectl && \
kubeadm init --pod-network-cidr=10.244.0.0/16 && \

mkdir -p $HOME/.kube && \
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config && \
chown $(id -u):$(id -g) $HOME/.kube/config && \
sysctl net.bridge.bridge-nf-call-iptables=1 && \
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml

kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/master/src/deploy/recommended/kubernetes-dashboard.yaml
kubectl create serviceaccount cluster-admin-dashboard-sa
kubectl create clusterrolebinding cluster-admin-dashboard-sa --clusterrole=cluster-admin --serviceaccount=default:cluster-admin-dashboard-sa