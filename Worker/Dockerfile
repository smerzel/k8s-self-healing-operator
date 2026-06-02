# שנה את השורה הראשונה ל:
    FROM golang:1.24-alpine AS builder
    WORKDIR /app

# מעתיק את כל הקוד, כולל קבצי ה-go.mod ותיקיית ה-vendor שהרגע יצרנו
COPY . .

# קימפול מתוך תיקיית ה-vendor המקומית ללא הורדות מהרשת (כדי לא להיחסם)
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o worker main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/worker .
CMD ["./worker"]