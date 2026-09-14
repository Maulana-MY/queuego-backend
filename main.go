package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"queuego-backend/config"
	"queuego-backend/routes"
)

func main() {

	// ========================================
	// KONEKSI DATABASE
	// ========================================

	config.ConnectDatabase()

	// ========================================
	// ROUTER
	// ========================================

	router := gin.Default()

	// ========================================
	// CORS MIDDLEWARE
	// ========================================

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// ========================================
	// TEST API
	// ========================================

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
			"message": "QueueGo API berhasil berjalan!",
		})
	})

	// ========================================
	// ROUTES
	// ========================================

	routes.SetupRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8095"
	}

	fmt.Println("========================================")
	fmt.Println("      QUEUEGO REST API")
	fmt.Println("========================================")
	fmt.Println("Server berjalan di port:", port)
	fmt.Println("========================================")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Gagal menjalankan server:", err)
	}
}