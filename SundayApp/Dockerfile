# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /app

# מעתיקים את כל הקבצים, כולל תיקיית ה-vendor
COPY . .

# בונים תוך שימוש ב-vendor (בלי לנסות להוריד כלום)
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o main .

# Stage 2: Runtime
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]