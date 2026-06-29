$ErrorActionPreference = "Stop"

$namespaceExists = kubectl get namespace cert-manager 2>$null
if ($namespaceExists) {
    kubectl get pods -n cert-manager
    $confirmation = Read-Host "Reinstall cert-manager? (yes/no)"
    if ($confirmation -ne "yes") {
        exit 0
    }
    kubectl delete -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml 2>$null
    Start-Sleep -Seconds 10
}

kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.3/cert-manager.yaml
kubectl wait --namespace cert-manager --for=condition=ready pod --selector=app.kubernetes.io/instance=cert-manager --timeout=180s
kubectl apply -f k8s/10-cert-manager.yaml

kubectl get pods -n cert-manager
kubectl get deployments -n cert-manager
kubectl get clusterissuer
kubectl get certificate -A
