# ⚙️ Event-Driven Go Backend & Custom K8s Controller

This repository showcases a distributed backend system and a custom Kubernetes controller, engineered entirely in **Go**. 

The project demonstrates advanced software engineering concepts, shifting away from basic synchronous APIs to a highly concurrent, decoupled architecture. It implements a custom control-loop to programmatically manage its own scale via the Kubernetes API, rather than relying on standard out-of-the-box infrastructure scripts.

## 💻 Software Architecture & Code Design

The system is divided into three primary Go microservices:

### 1. Message-Driven Ingress API (Go / Gin)
- **Decoupling:** Receives incoming requests and immediately pushes payloads to a **Redis** message broker. 
- **Performance:** This event-driven design ensures the API remains completely unblocked and highly responsive, delegating heavy processing to the background.

### 2. Asynchronous Worker Pool (Go)
- **Concurrency:** Background processes utilizing Go's concurrency model to continuously poll and consume tasks from the Redis queue.
- **Stateless Processing:** Workers handle business logic asynchronously, allowing for seamless horizontal scaling.

### 3. Algorithmic K8s Controller
- **K8s Native Engineering:** Programmed from scratch using the official `k8s.io/client-go` library, `dynamicClient`, and Informers.
- **Reconciliation Loop:** A continuous Go routine that queries system metrics and programmatically compares the current state (pending queue size) against desired thresholds.
- **Autonomous Scaling:** Makes direct K8s API calls to dynamically provision or terminate worker Pods based on real-time load.

## ☁️ Cloud-Native Deployment
To prove the architecture in a real-world environment, the complete microservices stack is deployed on a cloud-based **Azure Linux VM (IaaS)** running **MicroK8s**. This setup validates the system's ability to handle concurrent workloads and internal network communication in a production-like setting.

## 🛠️ Tech Stack
- **Core Languages:** Go (Golang)
- **Engineering Concepts:** Distributed Systems, Decoupling, Asynchronous Processing, Operator Pattern
- **Libraries:** `client-go` (Kubernetes Go SDK), Gin Web Framework
- **Databases & Brokers:** Redis, SQLite
- **Infrastructure:** Azure IaaS, MicroK8s, Docker

## 🚀 Developer Quick Start

A `Makefile` automates the build and deployment pipeline for rapid local testing.

**1. Compile & Build Docker Images:**
```bash
make build
