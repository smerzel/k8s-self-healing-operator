# --- Event-Driven Operator Makefile ---

.PHONY: build deploy run logs clean help

help:
    @echo "Available commands:"
    @echo "  make build  - Build Docker images for SundayApp and EtherealOperator"
    @echo "  make deploy - Deploy CRDs, Backend App, and Operator to the K8s Cluster"
    @echo "  make run    - Run the Operator locally in simulation mode (No K8s needed)"
    @echo "  make logs   - Follow the logs of the deployed Ethereal Operator"
    @echo "  make clean  - Remove all resources from the K8s cluster"

build:
    @echo "Building SundayApp Docker image..."
    docker build -t sunday-app:latest ./SundayApp
    @echo "Building EtherealOperator Docker image..."
    docker build -t ethereal-operator:latest ./EtherealOperator

deploy:
    @echo "Applying Custom Resource Definitions (CRDs)..."
    kubectl apply -f EtherealOperator/crd.yaml
    @echo "Deploying SundayApp Backend..."
    kubectl apply -f EtherealOperator/deployment.yaml
    @echo "Deploying Ethereal Operator with RBAC..."
    kubectl apply -f EtherealOperator/operator-setup.yaml
    @echo "Instantiating Custom Resource (EtherealPod)..."
    kubectl apply -f EtherealOperator/my-ghost.yaml
    @echo "Deployment complete! Run 'kubectl get pods' to view resources."

run:
    @echo "Running EtherealOperator locally in standalone mode..."
    cd EtherealOperator && go run main.go

logs:
    kubectl logs -f -l name=ethereal-operator

clean:
    @echo "Tearing down the environment..."
    kubectl delete -f EtherealOperator/my-ghost.yaml --ignore-not-found
    kubectl delete -f EtherealOperator/operator-setup.yaml --ignore-not-found
    kubectl delete -f EtherealOperator/deployment.yaml --ignore-not-found
    kubectl delete -f EtherealOperator/crd.yaml --ignore-not-found