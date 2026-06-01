package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/iZcy/imposizcy/internal/models"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/iZcy/imposizcy/internal/services"
	"github.com/sirupsen/logrus"
)

type FontHandler struct {
	repo          *repositories.FontRepository
	uploadService *services.UploadService
	registry      *services.FontRegistry
	logger        *logrus.Logger
}

func NewFontHandler(repo *repositories.FontRepository, uploadService *services.UploadService, registry *services.FontRegistry, logger *logrus.Logger) *FontHandler {
	return &FontHandler{repo: repo, uploadService: uploadService, registry: registry, logger: logger}
}

// Upload accepts multipart form: font (file), family, scope ("global"|"template"), template_id (optional).
func (h *FontHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("font")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No font file provided: " + err.Error()})
		return
	}
	family := strings.TrimSpace(c.PostForm("family"))
	if family == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "family is required"})
		return
	}
	scope := models.FontScope(strings.TrimSpace(c.DefaultPostForm("scope", string(models.FontScopeGlobal))))
	if scope != models.FontScopeGlobal && scope != models.FontScopeTemplate {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "scope must be 'global' or 'template'"})
		return
	}
	templateID := strings.TrimSpace(c.PostForm("template_id"))
	if scope == models.FontScopeTemplate && templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "template_id is required when scope=template"})
		return
	}

	relativePath, _, err := h.uploadService.SaveUpload(file, "fonts")
	if err != nil {
		h.logger.WithError(err).Error("Failed to save font upload")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	format := models.FontFormat(strings.TrimPrefix(strings.ToLower(filepath.Ext(file.Filename)), "."))
	row := &models.Font{
		Family:     family,
		FileName:   filepath.Base(file.Filename),
		Path:       relativePath,
		Format:     format,
		Scope:      scope,
		TemplateID: templateID,
		Size:       file.Size,
	}
	if err := h.repo.Create(c.Request.Context(), row); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to persist font: " + err.Error()})
		return
	}
	if err := h.registry.Register(row); err != nil {
		h.logger.WithError(err).Warn("Failed to warm font cache; will retry on render")
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": row})
}

func (h *FontHandler) List(c *gin.Context) {
	scope := c.Query("scope")
	templateID := c.Query("template_id")
	fonts, err := h.repo.List(c.Request.Context(), scope, templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": fonts})
}

func (h *FontHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	row, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Font not found"})
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	_ = h.uploadService.DeleteUpload(row.Path)
	h.registry.Invalidate(row.Family)
	c.JSON(http.StatusOK, gin.H{"success": true})
}
