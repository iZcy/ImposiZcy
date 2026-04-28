package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iZcy/imposizcy/internal/models"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/iZcy/imposizcy/internal/services"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type DashboardHandler struct {
	templateRepo  *repositories.TemplateRepository
	renderJobRepo *repositories.RenderJobRepository
	imageRepo     *repositories.ImageOutputRepository
	settingsRepo  *repositories.SettingsRepository
	kafkaConnRepo *repositories.KafkaConnectionRepository
	renderer      *services.RendererService
	validator     *services.ValidatorService
	imageGen      *services.ImageGeneratorService
	logger        *logrus.Logger
}

func NewDashboardHandler(
	templateRepo *repositories.TemplateRepository,
	renderJobRepo *repositories.RenderJobRepository,
	imageRepo *repositories.ImageOutputRepository,
	settingsRepo *repositories.SettingsRepository,
	kafkaConnRepo *repositories.KafkaConnectionRepository,
	renderer *services.RendererService,
	validatorSvc *services.ValidatorService,
	imageGen *services.ImageGeneratorService,
	logger *logrus.Logger,
) *DashboardHandler {
	return &DashboardHandler{
		templateRepo:  templateRepo,
		renderJobRepo: renderJobRepo,
		imageRepo:     imageRepo,
		settingsRepo:  settingsRepo,
		kafkaConnRepo: kafkaConnRepo,
		renderer:      renderer,
		validator:     validatorSvc,
		imageGen:      imageGen,
		logger:       logger,
	}
}

func (h *DashboardHandler) ListTemplates(c *gin.Context) {
	templates, err := h.templateRepo.List(c.Request.Context(), 1, 100, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": templates})
}

func (h *DashboardHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")
	template, err := h.templateRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Template not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": template})
}

func (h *DashboardHandler) CreateTemplate(c *gin.Context) {
	var req models.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	template := &models.PrintTemplate{
		Name:         req.Name,
		Slug:         req.Slug,
		HTML:         req.HTML,
		CSS:          req.CSS,
		DataSchema:   req.DataSchema,
		Width:        req.Width,
		Height:       req.Height,
		DimensionUnit: req.DimensionUnit,
		DPI:          req.DPI,
		OutputFormat: models.OutputFormatType(req.OutputFormat),
		Quality:      req.Quality,
		Tags:         req.Tags,
		IsActive:     true,
	}

	if err := h.templateRepo.Create(c.Request.Context(), template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": template})
}

func (h *DashboardHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	template, err := h.templateRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Template not found"})
		return
	}

	var req models.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Name != nil {
		template.Name = *req.Name
	}
	if req.HTML != nil {
		template.HTML = *req.HTML
	}
	if req.CSS != nil {
		template.CSS = *req.CSS
	}
	if req.DataSchema != nil {
		template.DataSchema = *req.DataSchema
	}
	if req.Width != nil {
		template.Width = *req.Width
	}
	if req.Height != nil {
		template.Height = *req.Height
	}
	if req.DimensionUnit != nil {
		template.DimensionUnit = *req.DimensionUnit
	}
	if req.DPI != nil {
		template.DPI = *req.DPI
	}
	if req.OutputFormat != nil {
		template.OutputFormat = models.OutputFormatType(*req.OutputFormat)
	}
	if req.Quality != nil {
		template.Quality = *req.Quality
	}
	if req.Tags != nil {
		template.Tags = req.Tags
	}
	if req.IsActive != nil {
		template.IsActive = *req.IsActive
	}

	if err := h.templateRepo.Update(c.Request.Context(), template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": template})
}

func (h *DashboardHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")
	if err := h.templateRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Template not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Template deleted"})
}

func (h *DashboardHandler) ListRenderJobs(c *gin.Context) {
	jobs, err := h.renderJobRepo.List(c.Request.Context(), 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": jobs})
}

func (h *DashboardHandler) GetRenderJob(c *gin.Context) {
	id := c.Param("id")
	job, err := h.renderJobRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Render job not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": job})
}

func (h *DashboardHandler) ListImages(c *gin.Context) {
	images, _, err := h.imageRepo.List(c.Request.Context(), 1, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": images})
}

func (h *DashboardHandler) DeleteImage(c *gin.Context) {
	id := c.Param("id")
	if err := h.imageRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Image not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Image deleted"})
}

func (h *DashboardHandler) GetSettings(c *gin.Context) {
	settings, err := h.settingsRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": settings})
}

func (h *DashboardHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")
	var req struct {
		Value    string `json:"value" binding:"required"`
		Category string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := h.settingsRepo.Set(c.Request.Context(), key, req.Value, req.Category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Setting updated"})
}

func (h *DashboardHandler) ListKafkaConnections(c *gin.Context) {
	conns, err := h.kafkaConnRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": conns})
}

func (h *DashboardHandler) CreateKafkaConnection(c *gin.Context) {
	var conn repositories.KafkaConnection
	if err := c.ShouldBindJSON(&conn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err := h.kafkaConnRepo.Create(c.Request.Context(), &conn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": conn})
}

func (h *DashboardHandler) UpdateKafkaConnection(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name    string   `json:"name"`
		Brokers []string `json:"brokers"`
		Topics  []string `json:"topics"`
		GroupID string   `json:"group_id"`
		Active  *bool    `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	update := bson.M{}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Brokers != nil {
		update["brokers"] = req.Brokers
	}
	if req.Topics != nil {
		update["topics"] = req.Topics
	}
	if req.GroupID != "" {
		update["group_id"] = req.GroupID
	}
	if req.Active != nil {
		update["active"] = *req.Active
	}
	if err := h.kafkaConnRepo.Update(c.Request.Context(), id, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Connection updated"})
}

func (h *DashboardHandler) DeleteKafkaConnection(c *gin.Context) {
	id := c.Param("id")
	if err := h.kafkaConnRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Connection not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Connection deleted"})
}

func (h *DashboardHandler) StartKafkaConnection(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Kafka connection started"})
}

func (h *DashboardHandler) StopKafkaConnection(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Kafka connection stopped"})
}
