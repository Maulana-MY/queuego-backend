	package routes

import (
	"github.com/gin-gonic/gin"

	"queuego-backend/handlers"
)

func SetupRoutes(router *gin.Engine) {

	api := router.Group("/api")
	{
		// =========================
		// AUTH
		// =========================

		api.POST("/register", handlers.Register)
		api.POST("/login", handlers.Login)

		// =========================
		// COUNTERS
		// =========================

		api.GET("/counters", handlers.GetCounters)

		// =========================
		// QUEUES
		// =========================

		api.GET("/queues/today", handlers.GetTodayQueues)

		api.GET("/queues/history", handlers.GetHistory)
		
		api.GET("/queues/:id", handlers.GetQueueByID)

		api.POST("/queues", handlers.CreateQueue)

		// =========================
		// OPERATOR
		// =========================

		api.POST("/queues/:id/call", handlers.CallNext)

		api.POST("/queues/:id/recall", handlers.RecallQueue)

		api.POST("/queues/:id/serve", handlers.StartServing)

		api.POST("/queues/:id/complete", handlers.CompleteQueue)

		api.POST("/queues/:id/skip", handlers.SkipQueue)
	}
}