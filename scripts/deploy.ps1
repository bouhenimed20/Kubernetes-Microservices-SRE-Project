$ErrorActionPreference = "Stop"

$CLUSTER_NAME = "sre-kind"

$clusterExists = kind get clusters | Select-String -Pattern "^${CLUSTER_NAME}$"
if (!$clusterExists) {
    Write-Error "Cluster not found"
    exit 1
}

kubectl config use-context kind-${CLUSTER_NAME}

kubectl apply -f k8s/00-namespaces.yaml
kubectl apply -f k8s/01-secrets.yaml

kubectl apply -f k8s/02-main-api.yaml
kubectl apply -f k8s/03-auth-service.yaml
kubectl apply -f k8s/04-image-service.yaml

kubectl apply -f k8s/05-network-policies.yaml
kubectl apply -f k8s/06-ingress.yaml
kubectl apply -f k8s/07-autoscaling.yaml

kubectl apply -f k8s/08-prometheus.yaml
kubectl apply -f k8s/09-grafana.yaml

kubectl apply -f k8s/10-cert-manager.yaml

kubectl wait --for=condition=ready pod -l app=main-api -n prod-api --timeout=120s
kubectl wait --for=condition=ready pod -l app=auth-service -n prod-auth --timeout=120s
kubectl wait --for=condition=ready pod -l app=image-service -n prod-image --timeout=120s
