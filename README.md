

**A Kubernetes Operator written in Go that ensures high availability and self-healing for a backend application.**
It implements a custom controller to manage pod lifecycles and guarantees data persistence across restarts using HostPath volumes.

---

## 🚀 Key Features

* **Self-Healing:** The Operator monitors the `EtherealPod`. If the managed pod is deleted or crashes, it is immediately resurrected.
* **Data Persistence:** Uses `HostPath` volumes to ensure data (SQLite) survives pod destruction ("Never lose data").
* **Smart Backend:** The Go application includes input validation and automatic lowercase normalization.
* **Status Tracking:** Updates the `RESTARTS` counter in the Custom Resource status, visible via CLI.

## 🛠️ How to Run

1.  **Deploy the CRD:**
    ```bash
    cd EtherealOperator
    kubectl apply -f crd.yaml
    ```
2.  **Start the Operator:**
    ```bash
    go run main.go
    ```
3.  **Create the Custom Resource:**
    ```bash
    kubectl apply -f my-ghost.yaml
    ```
4.  **Verify Status:**
    ```bash
    kubectl get eps
    ```