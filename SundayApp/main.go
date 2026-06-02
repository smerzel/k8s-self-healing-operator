package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"SundayApp/internal/repository"
)

func main() {
	// אתחול שכבת הנתונים - חיבור ישיר לתור ה-Redis בקלאסטר
	repo := repository.NewRedisRepo("redis-service:6379")
	defer repo.Close()

	// לקוח Redis ייעודי עבור מנגנון ה-Pub/Sub (שידור אירועים לאופרטור)
	redisClient := redis.NewClient(&redis.Options{Addr: "redis-service:6379"})
	defer redisClient.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// ראוט האופרטור - סופר משימות ממתינות ישירות מתוך התור
	r.GET("/pending-count", func(c *gin.Context) {
		count, err := repo.GetPendingTasksCount(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch count"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"count": count})
	})

	// ראוט הזרקת הנתונים - מקבל משימות ודוחף אותן לתור (Producer)
	r.POST("/tasks", func(c *gin.Context) {
		var task repository.Task

		if err := c.ShouldBindJSON(&task); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		// דחיפה לתור (PushTask) כדי שהפועלים יבצעו את העבודה
		if err := repo.PushTask(c.Request.Context(), task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to push task to Redis queue"})
			return
		}

		// ארכיטקטורת Push טהורה: שידור אירוע לערוץ כדי להעיר את האופרטור מידית
		redisClient.Publish(c.Request.Context(), "task_events", "new_task")

		c.JSON(http.StatusOK, gin.H{"status": "Task pushed to queue successfully"})
	})

	r.GET("/readiness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ready": true})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting gracefully")
}