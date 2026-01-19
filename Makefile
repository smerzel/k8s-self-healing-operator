.PHONY: build-images deploy-operator deploy-resource clean

# 1. בניית האימג'ים של האפליקציה ושל האופרטור
build-images:
	docker build -t sunday-app:v1 ./SundayApp
	docker build -t ethereal-operator:latest ./EtherealOperator

# 2. הרמת האופרטור לקלאסטר (כולל ה-CRD)
deploy-operator: build-images
	kubectl apply -f EtherealOperator/ethereal_crd.yaml
	kubectl apply -f EtherealOperator/operator-deployment.yaml

# 3. יצירת הפוד המנוהל (הטריגר לפעולה)
deploy-resource:
	kubectl apply -f EtherealOperator/my-ghost.yaml

# 4. מחיקה וניקוי
clean:
	kubectl delete -f EtherealOperator/my-ghost.yaml --ignore-not-found
	kubectl delete -f EtherealOperator/operator-deployment.yaml --ignore-not-found
	kubectl delete -f EtherealOperator/ethereal_crd.yaml --ignore-not-found