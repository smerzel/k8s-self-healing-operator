# 👻 Ethereal Self-Healing Operator

![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)
![Kubernetes](https://img.shields.io/badge/Kubernetes-Operator-326ce5.svg)
![Docker](https://img.shields.io/badge/Container-Ready-2496ed.svg)
![Build](https://img.shields.io/badge/Build-Hermetic-success)

## About

This project is a custom Kubernetes Operator built in Go designed to manage `EtherealPod` resources. It serves as a demonstration of high-performance, resilient controller design patterns for automating the lifecycle of backend applications.

The primary goal of the project was to move away from inefficient, polling-based reconciliation and implement a professional, **Event-Driven Architecture**.

**Key Technical Highlights (The "What I Did" section):**

1.  **Refactored to Event-Driven Core:** Replaced the continuous polling loop (frequent API server requests) with `client-go` **Informers** and a local **shared cache**. The operator now only reconciles when an actual event (`Add`, `Update`, `Delete`) is detected.
2.  **Workqueue Implementation:** Integrated a `RateLimitingQueue` to handle events asynchronously. This provides advanced flow control, debouncing of rapid changes, and automated, exponential backoff retries in case of transient failures, ensuring cluster stability under load.
3.  **Full Idempotency:** Optimized the reconciliation logic to strictly compare the *Desired State* (from the Custom Resource) against the *Actual State* (from the cluster's cache). It applies changes only when a drift is detected, making the operation safe to run repeatedly.
4.  **Cascading Garbage Collection:** Implemented `OwnerReferences` for every dynamically provisioned Pod. This ensures that deleting the parent `EtherealPod` Custom Resource automatically triggers the Kubernetes control plane's Garbage Collector to delete the managed Pod, preventing resource leaks.
5.  **Observability & Reliability:** Upgraded to structured JSON logging (`log/slog`) and enforced consistent `Context` usage with timeouts across all operations to ensure clean resource management and easier debugging.

---

## 🚀 Key Features

### 🛡️ Event-Driven Self-Healing
The Operator monitors the cluster state via an Informer. If the managed pod is deleted or crashes, the operator receives an event, processes it through the workqueue, and immediately **resurrects** the application, achieving high availability.

### 📦 Hermetic Builds (Offline Ready)
Utilizes `go mod vendor` to ensure fully reproducible builds that do not rely on external repositories during the build process, ideal for secure or air-gapped environments.

### 🧠 Smart Configuration
* **Auto-Detection:** Automatically detects if it's running inside a cluster or locally (Windows/Mac/Linux) and adjusts its configuration.
* **Local Dev Support:** Defaults to `ImagePullPolicy: IfNotPresent` for efficient local development workflows without needing a remote container registry.

---

## 🏗️ Architecture

1.  **Operator (Go):** The brain. Watches the Custom Resource (`EtherealPod`) and reconciles the state using advanced `client-go` patterns (Informers, Cache, Workqueue).
2.  **Backend App (Go/Gin/SQLite):** A RESTful API encapsulated in a lightweight container, strictly managed by the operator.
3.  **Database:** Uses SQLite for data persistence, managed within the application container.

---

## 🛠️ Getting Started

### Prerequisites
* **Docker Desktop** (Running with Kubernetes enabled) or **Minikube**
* **kubectl** (Command line tool)

### 📦 Installation & Deployment (Local Development)

#### 1. Build the Images
Build the application and the operator images directly to your local Docker registry.

```bash
# Build the Application (SundayApp v2)
docker build -t sunday-app:v2 ./SundayApp

# Build the Operator
docker build --no-cache -t ethereal-operator:latest ./EtherealOperator

2. Deploy CRD & Permissions
Apply the Custom Resource Definition and RBAC roles to the cluster.

Bash
kubectl apply -f EtherealOperator/crd.yaml
kubectl apply -f EtherealOperator/operator-deployment.yaml
3. Start the Managed Application
Apply the Custom Resource to trigger the operator to create the application.

Bash
kubectl apply -f EtherealOperator/my-ghost.yaml
🧪 Testing the Self-Healing
Verify Status: Check that both the operator and the app are running.

Bash
kubectl get pods
Result: You should see both pods with status Running.

Simulate a Failure: Manually delete the managed application pod.

Bash
kubectl delete pod real-sunday-server-pod
Witness the Resurrection: The Operator instantly detects the Delete event and creates a replacement.

Bash
kubectl get pods
Result: You will see a new real-sunday-server-pod with a very fresh AGE (e.g., 5s), proving the event-driven self-healing works.

🌐 Cloud Deployment
While optimized for local development, the architecture is ready for cloud environments (GKE, EKS, AKS). To deploy to the cloud, you will need to push the built images to a Container Registry (e.g., Docker Hub) and update the imagePullPolicy in your YAML files to Always.

📜 Project Structure

├── EtherealOperator/
│   ├── main.go                # Operator logic & event-driven reconciliation
│   ├── operator-deployment.yaml # Kubernetes Deployment for the Operator
│   ├── crd.yaml                # Custom Resource Definition
│   ├── my-ghost.yaml           # Custom Resource Instance (The Trigger)
│   └── Dockerfile              # Multi-stage build for the Operator
├── SundayApp/
│   ├── main.go                # Backend API (Gin + SQLite)
│   └── Dockerfile              # Multi-stage build for the App
└── README.md                  # Documentation