$ErrorActionPreference = "Stop"

$REGISTRY = "localhost:5001"
$CLUSTER_NAME = "sre-kind"

$clusterExists = kind get clusters | Select-String -Pattern "^${CLUSTER_NAME}$"
if (!$clusterExists) {
    Write-Error "Cluster not found"
    exit 1
}

docker build -t ${REGISTRY}/main-api:latest ./main-api
docker push ${REGISTRY}/main-api:latest

docker build -t ${REGISTRY}/auth-service:latest ./auth-service
docker push ${REGISTRY}/auth-service:latest

docker build -t ${REGISTRY}/image-service:latest ./image-service
docker push ${REGISTRY}/image-service:latest
