# Stage 1: Builder
FROM golang:1.25 AS builder
WORKDIR /app

# Copy source code and dependencies
COPY . .

# Build the binary using local vendor files
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o main .

# Stage 2: Runtime
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]