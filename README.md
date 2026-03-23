# 👻 Ethereal Self-Healing Operator

![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)
![Kubernetes](https://img.shields.io/badge/Kubernetes-Operator-326ce5.svg)
![Docker](https://img.shields.io/badge/Container-Ready-2496ed.svg)
![Build](https://img.shields.io/badge/Build-Hermetic-success)

## About

A Kubernetes Operator built with Go that manages custom `EtherealPod` resources with self-healing capabilities and data persistence. This project demonstrates production-ready patterns for building Kubernetes controllers using the operator pattern.

**Key Highlights:**
- **Self-Healing:** Automatically detects and recovers from pod failures
- **Custom Resources:** Implements Custom Resource Definitions (CRD) for declarative management
- **Production Patterns:** Uses the same APIs and patterns as enterprise operators (`client-go`, RBAC, Service Accounts)
- **Local & Cloud Ready:** Designed for local development (Docker Desktop) but can be deployed to any Kubernetes cluster

This is a **development/demo project** that showcases how to build Kubernetes operators. It runs on local Kubernetes by default (Docker Desktop or Minikube) but uses the same Kubernetes APIs that production operators use, making it an excellent learning resource and a foundation for production deployments.

**A Kubernetes Operator demonstrating production-ready patterns for high availability, offline stability, and automated lifecycle management.**

This project demonstrates a custom Kubernetes Controller that manages the lifecycle of a backend application (`SundayApp`). It ensures the application remains in its desired state, automatically "resurrecting" it in case of failure or accidental deletion.

> **Note:** This is a development/demo project designed to run on local Kubernetes (Docker Desktop or Minikube) by default. The operator uses the same Kubernetes APIs and patterns as production operators, making it suitable for learning and can be deployed to cloud Kubernetes clusters when needed.

---

## 🚀 Key Features

### 🛡️ Self-Healing Mechanism
The Operator constantly watches the cluster state. If the managed pod is deleted or crashes, the operator detects the discrepancy and **resurrects** it immediately. The actual availability depends on your infrastructure - in a properly configured cloud environment, this can achieve high availability (99.9%+).

### 📦 Hermetic Builds (Offline Ready)
The project utilizes `go mod vendor` to ensure fully reproducible builds. It does not rely on external repositories during the build process, making it secure and stable even in air-gapped or restricted network environments.

### 🧠 Smart Configuration
* **Auto-Detection:** The Operator automatically detects if it's running inside a cluster or on a local machine (Windows/Mac/Linux) and adjusts its configuration accordingly.
* **Local Dev Support:** Configured with `ImagePullPolicy: IfNotPresent` to support local development workflows (Docker Desktop) without needing a remote registry.
* **Cloud Ready:** Can be easily deployed to cloud Kubernetes clusters (GKE, EKS, AKS) with minimal configuration changes.

### 📊 Observability
Implements structured JSON logging (`log/slog`) for all events, making the system ready for modern observability stacks (ELK, Grafana, Datadog).

---

## 🏗️ Architecture

1.  **Operator (Go):** The brain of the system. Watches a Custom Resource Definition (`EtherealPod`) and reconciles the state using `client-go`.
2.  **Backend App (Go/Gin/SQLite):** A RESTful API encapsulated in a lightweight Docker container (`Alpine`), managed strictly by the operator.
3.  **Database:** Uses SQLite for data persistence, managed within the application container.

---

## 🌐 Deployment Options

This operator can run on various Kubernetes environments, from local development to cloud production clusters.

### Local Development (Recommended for Getting Started)

**Free and Easy Setup:**

* **Docker Desktop** (Windows/Mac/Linux)
  * Built-in Kubernetes support
  * One-click enable in settings
  * Perfect for development and testing
  * **Cost:** Free

* **Minikube** (All platforms)
  * Lightweight local Kubernetes
  * Great for learning
  * **Cost:** Free

* **Kind** (Kubernetes in Docker)
  * Fast cluster creation
  * Ideal for CI/CD testing
  * **Cost:** Free

* **K3s / MicroK8s**
  * Lightweight distributions
  * Good for edge/IoT scenarios
  * **Cost:** Free

### Cloud Deployment Options

**Free Tier / Trial Options:**

* **Google Kubernetes Engine (GKE)**
  * $300 free credit for new users
  * Free cluster management
  * Pay only for compute resources (nodes)
  * **Best for:** Production workloads, Google Cloud users

* **Amazon EKS**
  * Free cluster management
  * Pay only for EC2 instances or Fargate
  * **Best for:** AWS ecosystem integration

* **Azure Kubernetes Service (AKS)**
  * Free cluster management
  * Pay only for VM instances
  * $200 free credit for new users
  * **Best for:** Microsoft/Azure environments

**Note:** All cloud providers offer free tiers with credits for new accounts. After credits expire, you pay only for the compute resources (nodes) you use. For small development clusters, costs can be as low as $10-30/month.

---

## ☁️ Cloud Deployment Guide

If you want to deploy this operator to a cloud Kubernetes cluster, follow these steps:

### Prerequisites for Cloud Deployment

1. **Container Registry Account** (choose one):
   * Docker Hub (free public repos)
   * Google Container Registry (GCR)
   * Amazon Elastic Container Registry (ECR)
   * Azure Container Registry (ACR)

2. **Cloud Kubernetes Cluster** (GKE, EKS, or AKS)

3. **kubectl** configured to connect to your cluster

### Steps to Deploy to Cloud

#### 1. Build and Push Images to Registry

```bash
# Tag images for your registry (example with Docker Hub)
docker build -t yourusername/sunday-app:v2 ./SundayApp
docker build -t yourusername/ethereal-operator:latest ./EtherealOperator

# Push to registry
docker push yourusername/sunday-app:v2
docker push yourusername/ethereal-operator:latest
```

#### 2. Update Deployment YAMLs

Edit `EtherealOperator/operator-deployment.yaml`:

```yaml
spec:
  containers:
    - name: operator
      image: yourusername/ethereal-operator:latest
      imagePullPolicy: Always  # Change from Never to Always
```

Edit `EtherealOperator/my-ghost.yaml`:

```yaml
spec:
  image: yourusername/sunday-app:v2
```

#### 3. Deploy to Cloud Cluster

```bash
# Ensure kubectl points to your cloud cluster
kubectl config current-context

# Deploy (same commands as local)
kubectl apply -f EtherealOperator/crd.yaml
kubectl apply -f EtherealOperator/operator-deployment.yaml
kubectl apply -f EtherealOperator/my-ghost.yaml
```

#### 4. Verify Deployment

```bash
kubectl get pods
kubectl get etherealpods
```

### Cloud-Specific Considerations

* **Image Pull Secrets:** If using private registries, configure image pull secrets
* **Resource Limits:** Adjust CPU/memory limits based on your cluster capacity
* **Network Policies:** Configure if your cluster uses network policies
* **Persistent Storage:** For production, consider using cloud-managed databases instead of SQLite

---

## 🛠️ Getting Started

### Prerequisites
* **Docker Desktop** (Running)
* **Kubernetes** (Enabled in Docker Desktop or Minikube)
* **kubectl** (Command line tool)

### 📦 Installation & Deployment (Local Development)

> **Note:** The following instructions are for **local development** using Docker Desktop. For cloud deployment, see the [Cloud Deployment Guide](#-cloud-deployment-guide) section below.

These commands work on **Windows (PowerShell)**, **Mac**, and **Linux**.

#### 1. Build the Images
Since this is a local environment, we build the images directly to your local Docker registry.

```bash
# Build the Application (Version v2)
docker build -t sunday-app:v2 ./SundayApp

# Build the Operator (Using --no-cache to ensure latest code)
docker build --no-cache -t ethereal-operator:latest ./EtherealOperator
```

#### 2. Deploy CRD & Permissions
Set up the Custom Resource Definition and Role-Based Access Control (RBAC).

```bash
kubectl apply -f EtherealOperator/crd.yaml
kubectl apply -f EtherealOperator/operator-deployment.yaml
```

#### 3. Run the Managed Application
Trigger the operator to create the application pod by applying the Custom Resource.

```bash
kubectl apply -f EtherealOperator/my-ghost.yaml
```

---

## 🧪 Testing the Self-Healing

1.  **Verify Status:** Check that both the operator and the app are running.
    ```bash
    kubectl get pods
    ```
    *You should see two pods with status `Running`.*

2.  **Simulate a Disaster:** Delete the application pod to test resilience.
    ```bash
    kubectl delete pod real-sunday-server-pod
    ```

3.  **Witness the Resurrection:** Immediately check the pods again. The Operator will have already created a replacement pod.
    ```bash
    kubectl get pods
    ```
    *Result: You will see `real-sunday-server-pod` with a fresh `AGE` (e.g., 5s).*

---

## 📜 Project Structure

```text
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
```