package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"queuego-backend/config"
	"queuego-backend/models"
)

func GetCounters(c *gin.Context) {
	var counters []models.Counter

	result := config.DB.
		Where("is_active = ?", true).
		Order("id ASC").
		Find(&counters)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengambil data loket",
			"error":   result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    counters,
	})
}