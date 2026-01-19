# שלב 1: בנייה
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .


RUN go build -mod=vendor -o operator .
# שלב 2: הרצה (קונטיינר קטן ונקי)
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/operator .
CMD ["./operator"]