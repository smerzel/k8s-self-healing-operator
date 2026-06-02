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
	docker build -t saramerzel/sunday-app:latest ./SundayApp
	@echo "Building EtherealOperator Docker image..."
	docker build -t saramerzel/ethereal-operator:latest ./EtherealOperator

deploy:
	@echo "Deploying all numbered manifests to the K8s Cluster..."
	# התיקון: החלפת הפקודות הבודדות בפקודה אחת שקוראת את כל התיקייה לפי סדר אלפביתי
	kubectl apply -f EtherealOperator/manifests/

run:
	@echo "Running EtherealOperator locally in standalone mode..."
	cd EtherealOperator && go run main.go

logs:
	# התיקון: התאמת הלייבל החיפוש ללייבל הנכון של האופרטור בקובץ 04
	kubectl logs -f -l name=ethereal-operator

clean:
	@echo "Tearing down the environment..."
	# התיקון: מחיקה נקייה של כל המשאבים שהוגדרו בתיקיית המניפסטים
	kubectl delete -f EtherealOperator/manifests/ --ignore-not-found