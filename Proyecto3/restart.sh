#!/bin/bash
echo "=== Conectando al cluster ==="
gcloud container clusters get-credentials mumnk8s-cluster \
  --zone=us-central1-a \
  --project=mumnk8s-proyecto

echo "=== Reagregando labels a los nodos ==="
kubectl label nodes \
  $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') \
  $(kubectl get nodes -o jsonpath='{.items[1].metadata.name}') \
  $(kubectl get nodes -o jsonpath='{.items[2].metadata.name}') \
  node-role.kubernetes.io/master=true \
  node-role.kubernetes.io/control-plane=true \
  kubevirt.io/schedulable=true \
  --overwrite

echo "=== Reiniciando KubeVirt ==="
kubectl delete pods -n kubevirt -l kubevirt.io=virt-api --ignore-not-found
kubectl delete pods -n kubevirt -l kubevirt.io=virt-controller --ignore-not-found

echo "=== Configurando containerd en nodos ==="
for NODE in $(kubectl get nodes -o jsonpath='{.items[*].metadata.name}'); do
  echo "Configurando: $NODE"
  gcloud compute ssh $NODE \
    --zone=us-central1-a \
    --project=mumnk8s-proyecto \
    --command="
      sudo mkdir -p /etc/containerd/certs.d/34.29.32.190:5000 &&
      sudo bash -c 'cat > /etc/containerd/certs.d/34.29.32.190:5000/hosts.toml << TOML
server = \"http://34.29.32.190:5000\"
[host.\"http://34.29.32.190:5000\"]
  capabilities = [\"pull\", \"resolve\", \"push\"]
  skip_verify = true
TOML' &&
      grep -q '34.29.32.190' /etc/containerd/config.toml || sudo bash -c 'cat >> /etc/containerd/config.toml << CONF

[plugins.\"io.containerd.grpc.v1.cri\".registry.mirrors.\"34.29.32.190:5000\"]
  endpoint = [\"http://34.29.32.190:5000\"]

[plugins.\"io.containerd.grpc.v1.cri\".registry.configs.\"34.29.32.190:5000\".tls]
  insecure_skip_verify = true
CONF' &&
      sudo systemctl restart containerd &&
      echo 'Nodo $NODE configurado'
    " --quiet
done

echo "=== Esperando 60 segundos para que los pods levanten ==="
sleep 60

echo "=== Reiniciando pods ==="
kubectl delete pods -n mumnk8s --all --ignore-not-found
sleep 30

echo "=== Esperando VMs de KubeVirt ==="
kubectl wait --for=condition=ready vmi/valkey-vm -n mumnk8s --timeout=300s
kubectl wait --for=condition=ready vmi/grafana-vm -n mumnk8s --timeout=300s

echo "=== Configurando red en VM de Valkey ==="
sleep 30
virtctl ssh ubuntu@valkey-vm -n mumnk8s \
  --local-ssh-opts="-o StrictHostKeyChecking=no" \
  -- "sudo dhclient enp1s0 2>/dev/null; sudo systemctl start redis-server" || true


echo "=== Configurando red en Valkey via console ==="
kubectl exec -n mumnk8s \
  $(kubectl get pod -n mumnk8s -l app=valkey-vm -o jsonpath='{.items[0].metadata.name}') \
  -c guest-console-log -- echo "console ready" 2>/dev/null || true
sleep 15

echo "=== Obteniendo IP de Valkey ==="
sleep 10
VALKEY_IP=$(kubectl get vmi valkey-vm -n mumnk8s -o jsonpath='{.status.interfaces[0].ipAddress}')
echo "IP de Valkey: $VALKEY_IP"

echo "=== Actualizando consumer con IP de Valkey ==="
kubectl set env deployment/go-consumer \
  -n mumnk8s \
  VALKEY_ADDR=$VALKEY_IP:6379
kubectl rollout restart deployment go-consumer -n mumnk8s

echo "=== Estado final del cluster ==="
kubectl get nodes
echo "---"
kubectl get pods -n mumnk8s
echo "---"
kubectl get vmis -n mumnk8s
echo "---"
echo "IP Rust API LoadBalancer:"
kubectl get service rust-api -n mumnk8s | grep -v NAME | awk '{print $4}'
echo "IP Grafana:"
kubectl get service grafana -n mumnk8s | grep -v NAME | awk '{print $4}'
echo "=== Script completado ==="


# PASO MANUAL SI REDIS NO CONECTA:
# 1. virtctl console valkey-vm -n mumnk8s
# 2. Login: ubuntu / valkey123
# 3. sudo dhclient enp1s0
# 4. sudo systemctl start redis-server
# 5. Ctrl+] para salir
# 6. kubectl rollout restart deployment go-consumer -n mumnk8s