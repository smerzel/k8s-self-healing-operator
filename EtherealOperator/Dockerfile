# שלב 1: בנייה
# שנה את השורה הראשונה ל:
    FROM golang:1.24-alpine AS builder
    WORKDIR /app
COPY . .


RUN go build -mod=vendor -o operator .
# שלב 2: הרצה (קונטיינר קטן ונקי)
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/operator .
CMD ["./operator"]