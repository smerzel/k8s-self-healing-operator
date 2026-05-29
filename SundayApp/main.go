package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"

	"SundayApp/internal/repository"
)

func main() {
	// פתיחת חיבור למסד הנתונים
	db, err := sql.Open("sqlite", "./sunday.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// התיקון: יצירת הטבלה אם היא לא קיימת
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user TEXT,
		product TEXT,
		amount INTEGER,
		status TEXT
	)`)
	if err != nil {
		log.Fatal("Failed to initialize database table:", err)
	}

	// אתחול שכבת הנתונים
	repo := repository.NewSQLiteRepo(db)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// ראוט האופרטור - סופר משימות ממתינות
	r.GET("/pending-count", func(c *gin.Context) {
		count, err := repo.GetPendingTasksCount(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch count"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"count": count})
	})

	// ראוט הזרקת הנתונים - מקבל משימות ושומר ל-DB
	r.POST("/tasks", func(c *gin.Context) {
		var task repository.Task

		if err := c.ShouldBindJSON(&task); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		if err := repo.SaveTask(c.Request.Context(), task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "Task saved successfully and is pending"})
	})

	r.GET("/readiness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ready": true})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	r.Run(":8080")
}
