# Event-Driven Kubernetes Operator (Go)

This project showcases a custom-built Kubernetes Operator developed entirely in Go. It represents a shift from basic infrastructure scripts to a fully-fledged, event-driven backend system. 

Unlike standard "restart-on-crash" operators, this service actively monitors real-time business metrics from a target backend application and dynamically orchestrates the cluster state to handle fluctuating loads.

## Architecture & Core Components

1.  **SundayApp (Target Backend):**
    * A RESTful API built with the **Gin** framework and backed by a **SQLite** database.
    * Acts as a task-queue manager, processing incoming jobs and exposing a critical metrics endpoint (/pending-count) that reflects the system's current load.

2.  **EtherealOperator (The Core Controller):**
    * The "brain" of the operation, utilizing K8s native libraries (client-go, dynamicClient, Informers, Workqueues).
    * **Proactive Auto-Scaling:** It continuously queries SundayApp for pending tasks. 
        * **Scale UP:** If the pending task queue exceeds 15, the operator uses the K8s API to horizontally scale the backend (increasing Replicas to 3).
        * **Scale DOWN:** Once the queue drops below 5, it scales down to 1 Replica to optimize resources.
    * **Simulation Mode:** The operator is designed for high developer velocity; it can run entirely outside a cluster (Standalone/Simulation Mode) to test business logic quickly.

## Quick Start (Local Development)

This project includes a smart Makefile to simplify the build and deployment process. Ensure you have Docker and a local K8s environment (like Minikube/MicroK8s) running.

1.  **Build the Docker Images:**
   
bash
    make build
   
2.  **Deploy Everything to the Cluster:**
    (This deploys the CRDs, the SundayApp Backend, and the Operator with required RBAC)
   
bash
    make deploy
   
3.  **View Live Operator Logs:**
   
bash
    make logs
   
4.  **Tear Down the Environment:**
   
bash
    make clean
