package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/iZcy/imposizcy/config"
	"github.com/iZcy/imposizcy/internal/services"
	"github.com/sirupsen/logrus"
)

type UploadHandler struct {
	uploadService *services.UploadService
	cfg           *config.Config
	logger        *logrus.Logger
}

func NewUploadHandler(uploadService *services.UploadService, cfg *config.Config, logger *logrus.Logger) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
		cfg:           cfg,
		logger:        logger,
	}
}

func (h *UploadHandler) UploadBackground(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No file provided: " + err.Error(),
			"code":    http.StatusBadRequest,
		})
		return
	}

	relativePath, fullPath, err := h.uploadService.SaveUpload(file, "backgrounds")
	if err != nil {
		h.logger.WithError(err).Error("Failed to save upload")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Upload failed: " + err.Error(),
			"code":    http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"path":       relativePath,
			"full_path":  fullPath,
			"filename":   file.Filename,
			"size":       file.Size,
			"url":        "/uploads/" + relativePath,
		},
	})
}

func (h *UploadHandler) UploadTemplateAsset(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "No file provided: " + err.Error(),
			"code":    http.StatusBadRequest,
		})
		return
	}

	relativePath, fullPath, err := h.uploadService.SaveUpload(file, "assets")
	if err != nil {
		h.logger.WithError(err).Error("Failed to save upload")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Upload failed: " + err.Error(),
			"code":    http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"path":      relativePath,
			"full_path": fullPath,
			"filename":  file.Filename,
			"size":      file.Size,
			"url":       "/uploads/" + relativePath,
		},
	})
}

func (h *UploadHandler) ServeUploads(c *gin.Context) {
	c.File(filepath.Join(h.cfg.Storage.UploadDir, c.Param("filepath")))
}
