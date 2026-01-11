package main

import (
	"database/sql"
	"net/http"
	"os"
	"strings"

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
	var err error
	// 1. Create data directory for persistence
	os.MkdirAll("/data", 0755)

	// 2. Connect to SQLite database
	db, err = sql.Open("sqlite", "/data/sunday.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS items (user TEXT, product TEXT, amount INTEGER)`)
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	r.POST("/write", func(c *gin.Context) {
		var item Item
		if err := c.BindJSON(&item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// --- Input Validations ---
		if item.Amount <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than zero"})
			return
		}
		if item.User == "" || item.Product == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and product_name are required"})
			return
		}

		// Normalize input to lowercase to handle case-insensitive product names
		item.User = strings.ToLower(item.User)
		item.Product = strings.ToLower(item.Product)

		_, err := db.Exec("INSERT INTO items (user, product, amount) VALUES (?, ?, ?)", item.User, item.Product, item.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "created"})
	})

	r.GET("/get_product_amount", func(c *gin.Context) {
		// Normalize query param to lowercase for accurate search
		productName := strings.ToLower(c.Query("product_name"))

		rows, err := db.Query("SELECT amount FROM items WHERE product = ?", productName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	r.DELETE("/delete_product", func(c *gin.Context) {
		productName := strings.ToLower(c.Query("product_name"))
		_, err := db.Exec("DELETE FROM items WHERE product = ?", productName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "alive"})
	})

	r.Run(":8080")
}