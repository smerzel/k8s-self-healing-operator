package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
)

type Item struct {
	User    string `json:"user_id"`
	Product string `json:"product_name"`
	Amount  int    `json:"amount"`
}

var db *sql.DB

func main() {
	// הגדרת לוגר JSON מקצועי (זהה לאופרטור)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Sunday App is starting", "port", 8080)

	var err error
	os.MkdirAll("/data", 0755)

	db, err = sql.Open("sqlite", "/data/sunday.db")
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS items (user TEXT, product TEXT, amount INTEGER)`)
	if err != nil {
		slog.Error("Failed to create table", "error", err)
		os.Exit(1)
	}

	// הגדרת Gin במצב Release כדי שלא ידפיס לוגים מיותרים
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Middleware ללוגים בפורמט JSON
	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		slog.Info("HTTP Request",
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", c.Writer.Status(),
			"duration", time.Since(start).String(),
			"ip", c.ClientIP(),
		)
	})

	r.POST("/write", func(c *gin.Context) {
		var item Item
		if err := c.BindJSON(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if item.Amount <= 0 || item.User == "" || item.Product == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input fields"})
			return
		}

		item.User = strings.ToLower(item.User)
		item.Product = strings.ToLower(item.Product)

		_, err := db.Exec("INSERT INTO items (user, product, amount) VALUES (?, ?, ?)", item.User, item.Product, item.Amount)
		if err != nil {
			slog.Error("Database write error", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "created"})
	})

	r.GET("/get_product_amount", func(c *gin.Context) {
		productName := strings.ToLower(c.Query("product_name"))
		rows, err := db.Query("SELECT amount FROM items WHERE product = ?", productName)
		if err != nil {
			slog.Error("Database read error", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		defer rows.Close()

		total := 0
		found := false
		for rows.Next() {
			var amount int
			rows.Scan(&amount)
			total += amount
			found = true
		}

		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"product": productName, "amount": total})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "alive"})
	})

	slog.Info("Server is ready and listening")
	r.Run(":8080")
}