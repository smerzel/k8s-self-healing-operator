package main

import (
	"database/sql"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
)

func main() {
	db, _ := sql.Open("sqlite", "./sunday.db")
	_ = db // Reserved for future persistence layer

	// העברה למצב ייצור כדי להראות לוגים נקיים
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// סימולציה של מערכת עמוסה: מחזיר מספר משימות אקראי בין 0 ל-20
	r.GET("/pending-count", func(c *gin.Context) {
		rand.Seed(time.Now().UnixNano())
		load := rand.Intn(20) 
		c.JSON(http.StatusOK, gin.H{"count": load})
	})

	// Readiness probe for Kubernetes
	r.GET("/readiness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ready": true})
	})

	r.Run(":8080")
}