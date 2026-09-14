package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"queuego-backend/config"
	"queuego-backend/models"
)

// ============================================================
// REQUEST
// ============================================================

type CreateQueueRequest struct {
	CounterID    uint   `json:"counter_id" binding:"required"`
	CustomerName string `json:"customer_name"`
}

// ============================================================
// GET ANTREAN HARI INI
// ============================================================

func GetTodayQueues(c *gin.Context) {
	var queues []models.Queue

	startOfDay := time.Now().Truncate(24 * time.Hour)

	// Lebih aman menggunakan tanggal langsung
	today := time.Now().Format("2006-01-02")

	result := config.DB.
		Where("DATE(created_at) = ?", today).
		Order("id ASC").
		Find(&queues)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengambil antrean hari ini",
			"error":   result.Error.Error(),
		})
		return
	}

	_ = startOfDay

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    queues,
	})
}

// ============================================================
// GET HISTORY
// ============================================================

func GetHistory(c *gin.Context) {
	var queues []models.Queue

	result := config.DB.
		Where(
			"status IN ?",
			[]string{"completed", "skipped", "cancelled"},
		).
		Order("completed_at DESC").
		Find(&queues)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengambil riwayat antrean",
			"error":   result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    queues,
	})
}

// ============================================================
// AMBIL ANTREAN
// ============================================================

func CreateQueue(c *gin.Context) {
	var request CreateQueueRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Data request tidak valid",
			"error":   err.Error(),
		})
		return
	}

	// Cek loket
	var counter models.Counter

	result := config.DB.First(&counter, request.CounterID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Loket tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mencari loket",
			"error":   result.Error.Error(),
		})
		return
	}

	if !counter.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Loket sedang tidak aktif",
		})
		return
	}

	// Nama default
	customerName := request.CustomerName

	if customerName == "" {
		customerName = "Pelanggan"
	}

	// Tanggal hari ini
	today := time.Now().Format("2006-01-02")

	// Hitung nomor terakhir untuk loket tersebut
	var lastQueue models.Queue

	result = config.DB.
		Where(
			"counter_id = ? AND DATE(created_at) = ?",
			request.CounterID,
			today,
		).
		Order("id DESC").
		First(&lastQueue)

	nextNumber := 1

	if result.Error == nil {
		var lastNumber int

		_, err := fmt.Sscanf(
			lastQueue.QueueNumber,
			fmt.Sprintf("L%d-%03d", request.CounterID, &lastNumber),
		)

		if err == nil && lastNumber > 0 {
			nextNumber = lastNumber + 1
		} else {
			// Fallback: hitung jumlah antrean
			var count int64

			config.DB.
				Model(&models.Queue{}).
				Where(
					"counter_id = ? AND DATE(created_at) = ?",
					request.CounterID,
					today,
				).
				Count(&count)

			nextNumber = int(count) + 1
		}
	}

	queueNumber := fmt.Sprintf(
		"L%d-%03d",
		request.CounterID,
		nextNumber,
	)

	queue := models.Queue{
		CounterID:    request.CounterID,
		QueueNumber:  queueNumber,
		CustomerName: customerName,
		Status:       "waiting",
		CreatedAt:    time.Now(),
	}

	result = config.DB.Create(&queue)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal membuat antrean",
			"error":   result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Antrean berhasil dibuat",
		"data":    queue,
	})
}

// ============================================================
// PANGGIL ANTREAN BERIKUTNYA
// ============================================================

func CallNext(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID antrean tidak valid",
		})
		return
	}

	var queue models.Queue

	// ID di endpoint digunakan sebagai counter ID
	counterID := uint(id)

	today := time.Now().Format("2006-01-02")

	// Selesaikan antrean aktif sebelumnya
	var activeQueues []models.Queue

	config.DB.
		Where(
			"counter_id = ? AND DATE(created_at) = ? AND status IN ?",
			counterID,
			today,
			[]string{"calling", "serving"},
		).
		Find(&activeQueues)

	now := time.Now()

	for _, q := range activeQueues {
		q.Status = "completed"
		q.CompletedAt = &now
		config.DB.Save(&q)
	}

	// Cari antrean waiting paling awal
	result := config.DB.
		Where(
			"counter_id = ? AND DATE(created_at) = ? AND status = ?",
			counterID,
			today,
			"waiting",
		).
		Order("id ASC").
		First(&queue)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Tidak ada antrean yang menunggu",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengambil antrean berikutnya",
			"error":   result.Error.Error(),
		})
		return
	}

	queue.Status = "calling"
	queue.CalledAt = &now

	result = config.DB.Save(&queue)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal memanggil antrean",
			"error":   result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Antrean berhasil dipanggil",
		"data":    queue,
	})
}

// ============================================================
// PANGGIL ULANG
// ============================================================

func RecallQueue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID antrean tidak valid",
		})
		return
	}

	var queue models.Queue

	result := config.DB.First(&queue, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Antrean tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mencari antrean",
			"error":   result.Error.Error(),
		})
		return
	}

	now := time.Now()

	queue.Status = "calling"
	queue.CalledAt = &now

	result = config.DB.Save(&queue)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal memanggil ulang antrean",
			"error":   result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Antrean dipanggil ulang",
		"data":    queue,
	})
}

// ============================================================
// MULAI MELAYANI
// ============================================================

func StartServing(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID antrean tidak valid",
		})
		return
	}

	var queue models.Queue

	result := config.DB.First(&queue, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Antrean tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mencari antrean",
			"error":   result.Error.Error(),
		})
		return
	}

	if queue.Status != "calling" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Antrean harus berstatus calling",
		})
		return
	}

	queue.Status = "serving"

	result = config.DB.Save(&queue)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal memulai pelayanan",
			"error":   result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Antrean mulai dilayani",
		"data":    queue,
	})
}

// ============================================================
// SELESAI
// ============================================================

func CompleteQueue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID antrean tidak valid",
		})
		return
	}

	var queue models.Queue

	result := config.DB.First(&queue, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Antrean tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mencari antrean",
			"error":   result.Error.Error(),
		})
		return
	}

	now := time.Now()

	queue.Status = "completed"
	queue.CompletedAt = &now

	result = config.DB.Save(&queue)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal menyelesaikan antrean",
			"error":   result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Antrean selesai dilayani",
		"data":    queue,
	})
}

// ============================================================
// LEWATI
// ============================================================

func SkipQueue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID antrean tidak valid",
		})
		return
	}

	var queue models.Queue

	result := config.DB.First(&queue, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Antrean tidak ditemukan",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mencari antrean",
			"error":   result.Error.Error(),
		})
		return
	}

	now := time.Now()

	queue.Status = "skipped"
	queue.CompletedAt = &now

	result = config.DB.Save(&queue)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal melewati antrean",
			"error":   result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Antrean dilewati",
		"data":    queue,
	})
}