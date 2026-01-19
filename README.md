# 👻 Ethereal Self-Healing Operator

![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)
![Kubernetes](https://img.shields.io/badge/Kubernetes-Operator-326ce5.svg)
![Docker](https://img.shields.io/badge/Container-Ready-2496ed.svg)
![Build](https://img.shields.io/badge/Build-Hermetic-success)

**A production-grade Kubernetes Operator designed for high availability, offline stability, and automated lifecycle management.**

This project demonstrates a custom Kubernetes Controller that manages the lifecycle of a backend application (`SundayApp`). It ensures the application remains in its desired state, automatically "resurrecting" it in case of failure or accidental deletion.

---

## 🚀 Key Features

### 🛡️ Self-Healing Mechanism
The Operator constantly watches the cluster state. If the managed pod is deleted or crashes, the operator detects the discrepancy and **resurrects** it immediately, ensuring 99.9% availability.

### 📦 Hermetic Builds (Offline Ready)
The project utilizes `go mod vendor` to ensure fully reproducible builds. It does not rely on external repositories during the build process, making it secure and stable even in air-gapped or restricted network environments.

### 🧠 Smart Configuration
* **Auto-Detection:** The Operator automatically detects if it's running inside a cluster or on a local machine (Windows/Mac/Linux) and adjusts its configuration accordingly.
* **Local Dev Support:** Configured with `ImagePullPolicy: IfNotPresent` to support local development workflows (Docker Desktop) without needing a remote registry.

### 📊 Observability
Implements structured JSON logging (`log/slog`) for all events, making the system ready for modern observability stacks (ELK, Grafana, Datadog).

---

## 🏗️ Architecture

1.  **Operator (Go):** The brain of the system. Watches a Custom Resource Definition (`EtherealPod`) and reconciles the state using `client-go`.
2.  **Backend App (Go/Gin/SQLite):** A RESTful API encapsulated in a lightweight Docker container (`Alpine`), managed strictly by the operator.
3.  **Database:** Uses SQLite for data persistence, managed within the application container.

---

## 🛠️ Getting Started

### Prerequisites
* **Docker Desktop** (Running)
* **Kubernetes** (Enabled in Docker Desktop or Minikube)
* **kubectl** (Command line tool)

### 📦 Installation & Deployment (Cross-Platform)

These commands work on **Windows (PowerShell)**, **Mac**, and **Linux**.

#### 1. Build the Images
Since this is a local environment, we build the images directly to your local Docker registry.

```bash
# Build the Application (Version v2)
docker build -t sunday-app:v2 ./SundayApp

# Build the Operator (Using --no-cache to ensure latest code)
docker build --no-cache -t ethereal-operator:latest ./EtherealOperator
2. Deploy CRD & Permissions
Set up the Custom Resource Definition and Role-Based Access Control (RBAC).

Bash

kubectl apply -f EtherealOperator/crd.yaml
kubectl apply -f EtherealOperator/operator-deployment.yaml
3. Run the Managed Application
Trigger the operator to create the application pod by applying the Custom Resource.

Bash

kubectl apply -f EtherealOperator/my-ghost.yaml
🧪 Testing the Self-Healing
Verify Status: Check that both the operator and the app are running.

Bash

kubectl get pods
You should see two pods with status Running.

Simulate a Disaster: Delete the application pod to test resilience.

Bash

kubectl delete pod real-sunday-server-pod
Witness the Resurrection: Immediately check the pods again. The Operator will have already created a replacement pod.

Bash

kubectl get pods
Result: You will see real-sunday-server-pod with a fresh AGE (e.g., 5s).

📜 Project Structure
Plaintext

├── EtherealOperator/
│   ├── main.go                 # Operator logic & reconciliation loop
│   ├── operator-deployment.yaml # K8s Deployment for the Operator
│   ├── crd.yaml                # Custom Resource Definition
│   ├── my-ghost.yaml           # Custom Resource Instance (The Trigger)
│   └── Dockerfile              # Multi-stage build for the Operator
├── SundayApp/
│   ├── main.go                 # Backend API (Gin + SQLite)
│   └── Dockerfile              # Multi-stage build for the App
└── README.md                   # Documentation