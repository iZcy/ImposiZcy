package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/sirupsen/logrus"
)

type SecurityHandler struct {
	apiKeyRepo   *repositories.APIKeyRepository
	settingsRepo *repositories.SettingsRepository
	logger       *logrus.Logger
}

func NewSecurityHandler(
	apiKeyRepo *repositories.APIKeyRepository,
	settingsRepo *repositories.SettingsRepository,
	logger *logrus.Logger,
) *SecurityHandler {
	return &SecurityHandler{
		apiKeyRepo:   apiKeyRepo,
		settingsRepo: settingsRepo,
		logger:       logger,
	}
}

func (h *SecurityHandler) Login(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Password required"})
		return
	}

	setting, err := h.settingsRepo.Get(c.Request.Context(), "admin_password")
	if err != nil || setting == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid credentials"})
		return
	}

	if req.Password != setting.Value {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "admin",
		"iss": "imposizcy",
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	secret, _ := h.settingsRepo.Get(c.Request.Context(), "jwt_secret")
	if secret == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "JWT secret not configured"})
		return
	}

	tokenString, err := token.SignedString([]byte(secret.Value))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"token": tokenString}})
}

func (h *SecurityHandler) CreateAPIKey(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	apiKey, rawKey, err := h.apiKeyRepo.Create(c.Request.Context(), req.Name, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":     apiKey.ID,
			"key":    rawKey,
			"prefix": apiKey.Prefix,
		},
	})
}

func (h *SecurityHandler) ListAPIKeys(c *gin.Context) {
	keys, err := h.apiKeyRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": keys})
}

func (h *SecurityHandler) DeleteAPIKey(c *gin.Context) {
	id := c.Param("id")
	if err := h.apiKeyRepo.Revoke(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "API key revoked"})
}

func generateAPIKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "imz_" + hex.EncodeToString(b)
}
