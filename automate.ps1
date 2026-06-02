param (
    [string]$Action = "help"
)

switch ($Action) {
    "build" {
        Write-Host "=== Building SundayApp Docker image ===" -ForegroundColor Cyan
        docker build -t saramerzel/sunday-app:latest ./SundayApp

        Write-Host "=== Building EtherealOperator Docker image ===" -ForegroundColor Cyan
        docker build -t saramerzel/ethereal-operator:latest ./EtherealOperator

        Write-Host "=== Building Worker Docker image ===" -ForegroundColor Cyan
        docker build -t saramerzel/sunday-worker:latest ./Worker
        
        Write-Host "Build complete successfully!" -ForegroundColor Green
    }
    
    "deploy" {
        Write-Host "=== Deploying all numbered manifests to the K8s Cluster ===" -ForegroundColor Green
        kubectl apply -f EtherealOperator/manifests/
    }
    
    "logs" {
        Write-Host "=== Following Ethereal Operator logs ===" -ForegroundColor Yellow
        kubectl logs -f -l name=ethereal-operator
    }
    
    "clean" {
        Write-Host "=== Tearing down the environment ===" -ForegroundColor Red
        kubectl delete -f EtherealOperator/manifests/ --ignore-not-found
    }
    
    Default {
        Write-Host "Available commands:" -ForegroundColor Yellow
        Write-Host "  ./automate.ps1 build  - Build all Docker images"
        Write-Host "  ./automate.ps1 deploy - Deploy all manifests"
        Write-Host "  ./automate.ps1 logs   - Follow operator logs"
        Write-Host "  ./automate.ps1 clean  - Remove all resources"
    }
}