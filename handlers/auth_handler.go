package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"queuego-backend/config"
	"queuego-backend/models"
)

// ============================================================
// REQUEST STRUCTS
// ============================================================

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=4"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ============================================================
// REGISTER HANDLER
// ============================================================

func Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Data pendaftaran tidak valid: " + err.Error(),
		})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	name := strings.TrimSpace(req.Name)
	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role == "" {
		role = "user"
	}

	// Cek apakah email sudah terdaftar di database
	var count int64
	config.DB.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Email sudah terdaftar. Silakan gunakan email lain.",
		})
		return
	}

	// Enkripsi password menggunakan bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengenkripsi kata sandi",
		})
		return
	}

	now := time.Now()
	user := models.User{
		Name:      name,
		Email:     email,
		Role:      role,
		Password:  string(hashedPassword),
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal menyimpan akun ke database",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Pendaftaran berhasil!",
		"data": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		},
	})
}

// ============================================================
// LOGIN HANDLER
// ============================================================

func Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Email/Username dan kata sandi wajib diisi",
		})
		return
	}

	identifier := strings.TrimSpace(req.Email)
	identifierLower := strings.ToLower(identifier)

	var user models.User
	result := config.DB.Where("LOWER(email) = ? OR LOWER(name) = ?", identifierLower, identifierLower).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Email / Nama Pengguna atau kata sandi salah",
		})
		return
	}

	// Verifikasi hash password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Email / Nama Pengguna atau kata sandi salah",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login berhasil!",
		"data": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}
