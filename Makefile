.PHONY: build-app build-operator run-app run-operator docker-build

# פקודות לבנייה (Build)
build-all:
	go build -o bin/app ./SundayApp/main.go
	go build -o bin/operator ./EtherealOperator/main.go

# הרצה מקומית של האפליקציה
run-app:
	cd SundayApp && go run main.go

# הרצה מקומית של האופרטור
run-operator:
	cd EtherealOperator && go run main.go

# פקודות עזר לקוברנטיס
deploy-crd:
	kubectl apply -f EtherealOperator/ethereal_crd.yaml

deploy-resource:
	kubectl apply -f EtherealOperator/my-ghost.yaml

clean-k8s:
	kubectl delete -f EtherealOperator/my-ghost.yaml
	kubectl delete pods -l managed-by=ethereal-operator