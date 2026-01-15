👻 Ethereal Self-Healing Operator
A production-grade Kubernetes Operator written in Go, designed for high availability, state management, and data persistence.

This project demonstrates a custom Kubernetes controller that manages the lifecycle of a backend application, ensuring it remains in the desired state while maintaining data integrity.

🚀 Key Features (Advanced)
Self-Healing & Idempotency: The Operator doesn't just "resurrect" pods; it ensures the running state matches the desired configuration (e.g., Image version) defined in the Custom Resource.

Structured JSON Logging: Implements log/slog for industry-standard, machine-readable logs, making the system ready for modern observability stacks (ELK/Grafana).

Data Persistence: Guaranteed persistence using HostPath volumes, ensuring SQLite data survives pod lifecycle events.

Custom Status Subresource: Real-time tracking of resurrection events directly in the Kubernetes API.

Developer Experience (DX): Fully automated workflow using a Makefile.

🏗️ Architecture
Operator (Go): Watches a Custom Resource Definition (CRD) and reconciles the state.

Backend App (Go/Gin): A RESTful API that handles product data with input validation and persistence.

Persistence Layer: Local storage mapping for stateful workloads in a stateless environment.

🛠️ Getting Started
The easiest way to run and test the project is using the provided Makefile.

1. Prerequisites
Kubernetes Cluster (Docker Desktop / Minikube)

Go 1.21+

2. Installation & Setup
Bash

# Deploy the CRD to your cluster
make deploy-crd

# Run the Operator locally (will output structured JSON logs)
make run-operator
3. Deploy a Managed Pod
In a new terminal:

Bash

# Create the Custom Resource
make deploy-resource

# Check the status and resurrection count
kubectl get eps
4. Test Idempotency & Self-Healing
Self-Healing: Delete the pod (kubectl delete pod <pod-name>) and watch it reappear.

Update: Change the image in my-ghost.yaml and run make deploy-resource again. The Operator will detect the mismatch and update the pod.