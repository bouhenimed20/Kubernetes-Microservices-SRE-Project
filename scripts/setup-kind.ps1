$ErrorActionPreference = "Stop"

$REGISTRY_NAME = "kind-registry"
$REGISTRY_PORT = "5001"
$CLUSTER_NAME = "sre-kind"

$commands = @("kind", "docker", "kubectl")
foreach ($cmd in $commands) {
    if (!(Get-Command $cmd -ErrorAction SilentlyContinue)) {
        Write-Error "$cmd not found"
        exit 1
    }
}

$registryExists = docker ps -a --format '{{.Names}}' | Select-String -Pattern "^${REGISTRY_NAME}$"
if (!$registryExists) {
    docker run -d -p ${REGISTRY_PORT}:5000 --restart=always --name ${REGISTRY_NAME} registry:2
}

$clusterExists = kind get clusters | Select-String -Pattern "^${CLUSTER_NAME}$"
if (!$clusterExists) {
    $kindConfig = @"
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: ${CLUSTER_NAME}
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
- role: worker
- role: worker
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${REGISTRY_PORT}"]
    endpoint = ["http://${REGISTRY_NAME}:5000"]
"@
    $kindConfig | Out-File -FilePath "$env:TEMP\kind-config.yaml" -Encoding UTF8
    kind create cluster --config="$env:TEMP\kind-config.yaml"
}

$REGISTRY_CONTAINER_ID = docker ps -q -f name="^${REGISTRY_NAME}$"
if ($REGISTRY_CONTAINER_ID) {
    $NETWORK = docker inspect $REGISTRY_CONTAINER_ID --format='{{json .HostConfig.NetworkMode}}' | ConvertFrom-Json
    $CLUSTER_NODES = kind get nodes --name ${CLUSTER_NAME}
    foreach ($node in $CLUSTER_NODES) {
        docker network connect $NETWORK $node 2>$null
    }
}

kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/kind/deploy.yaml
kubectl wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=120s

kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
Start-Sleep -Seconds 10
kubectl patch deployment metrics-server -n kube-system --type='json' -p='[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]'

kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml
Start-Sleep -Seconds 30
kubectl wait --namespace cert-manager --for=condition=ready pod --selector=app.kubernetes.io/instance=cert-manager --timeout=120s

kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.27.0/manifests/calico.yaml
