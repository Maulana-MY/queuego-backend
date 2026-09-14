package handler

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"queuego-backend/config"
	"queuego-backend/routes"
)

var (
	app  *gin.Engine
	once sync.Once
)

func getEngine() *gin.Engine {
	once.Do(func() {
		gin.SetMode(gin.ReleaseMode)
		config.ConnectDatabase()
		app = gin.Default()

		app.Use(func(c *gin.Context) {
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

		app.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"success": true,
				"message": "QueueGo API Vercel Serverless berhasil berjalan!",
			})
		})

		routes.SetupRoutes(app)
	})
	return app
}

func Handler(w http.ResponseWriter, r *http.Request) {
	getEngine().ServeHTTP(w, r)
}
