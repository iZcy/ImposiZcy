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
	templateRepo     *repositories.TemplateRepository
	renderJobRepo    *repositories.RenderJobRepository
	imageRepo        *repositories.ImageOutputRepository
	settingsRepo     *repositories.SettingsRepository
	kafkaConnRepo    *repositories.KafkaConnectionRepository
	eventMappingRepo *repositories.EventMappingRepository
	kafkaLogRepo     *repositories.KafkaLogRepository
	kafkaManager     *services.KafkaManager
	renderer         *services.RendererService
	validator        *services.ValidatorService
	imageGen         *services.ImageGeneratorService
	logger           *logrus.Logger
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

func (h *DashboardHandler) SetEventMappingRepo(repo *repositories.EventMappingRepository) {
	h.eventMappingRepo = repo
}

func (h *DashboardHandler) SetKafkaLogRepo(repo *repositories.KafkaLogRepository) {
	h.kafkaLogRepo = repo
}

func (h *DashboardHandler) SetKafkaManager(mgr *services.KafkaManager) {
	h.kafkaManager = mgr
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
		Name:          req.Name,
		Slug:          req.Slug,
		HTML:          req.HTML,
		CSS:           req.CSS,
		DataSchema:    req.DataSchema,
		Width:         req.Width,
		Height:        req.Height,
		DimensionUnit: req.DimensionUnit,
		DPI:           req.DPI,
		OutputFormat:  models.OutputFormatType(req.OutputFormat),
		Quality:       req.Quality,
		Tags:          req.Tags,
		IsActive:      true,
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

	if h.kafkaManager != nil {
		statuses := h.kafkaManager.GetConnectionStatus()
		for _, conn := range conns {
			for _, st := range statuses {
				if st.ID == conn.ID.Hex() {
					// merge status into response
					break
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": conns})
}

func (h *DashboardHandler) CreateKafkaConnection(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Broker      string `json:"broker" binding:"required"`
		Topic       string `json:"topic" binding:"required"`
		GroupID     string `json:"group_id"`
		ClientID    string `json:"client_id"`
		AutoOffset  string `json:"auto_offset"`
		Enabled     *bool  `json:"enabled"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	conn := repositories.NewKafkaConnection(&repositories.KafkaConnection{
		Name:        req.Name,
		Broker:      req.Broker,
		Topic:       req.Topic,
		GroupID:     req.GroupID,
		ClientID:    req.ClientID,
		AutoOffset:  req.AutoOffset,
		Description: req.Description,
		Enabled:     true,
	})
	if req.Enabled != nil {
		conn.Enabled = *req.Enabled
	}

	if err := h.kafkaConnRepo.Create(c.Request.Context(), conn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if conn.Enabled && h.kafkaManager != nil {
		_ = h.kafkaManager.StartConnection(c.Request.Context(), conn)
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": conn})
}

func (h *DashboardHandler) UpdateKafkaConnection(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        string `json:"name"`
		Broker      string `json:"broker"`
		Topic       string `json:"topic"`
		GroupID     string `json:"group_id"`
		ClientID    string `json:"client_id"`
		AutoOffset  string `json:"auto_offset"`
		Enabled     *bool  `json:"enabled"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	update := bson.M{}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Broker != "" {
		update["broker"] = req.Broker
	}
	if req.Topic != "" {
		update["topic"] = req.Topic
	}
	if req.GroupID != "" {
		update["group_id"] = req.GroupID
	}
	if req.ClientID != "" {
		update["client_id"] = req.ClientID
	}
	if req.AutoOffset != "" {
		update["auto_offset"] = req.AutoOffset
	}
	if req.Description != "" {
		update["description"] = req.Description
	}
	if req.Enabled != nil {
		update["enabled"] = *req.Enabled
	}

	if err := h.kafkaConnRepo.Update(c.Request.Context(), id, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if h.kafkaManager != nil {
		_ = h.kafkaManager.RestartConnection(c.Request.Context(), id)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Connection updated"})
}

func (h *DashboardHandler) DeleteKafkaConnection(c *gin.Context) {
	id := c.Param("id")

	if h.kafkaManager != nil {
		_ = h.kafkaManager.StopConnection(id)
	}

	if err := h.kafkaConnRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Connection not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Connection deleted"})
}

func (h *DashboardHandler) StartKafkaConnection(c *gin.Context) {
	id := c.Param("id")
	if h.kafkaManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Kafka manager not available"})
		return
	}
	if err := h.kafkaManager.RestartConnection(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Kafka connection started"})
}

func (h *DashboardHandler) StopKafkaConnection(c *gin.Context) {
	id := c.Param("id")
	if h.kafkaManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Kafka manager not available"})
		return
	}
	if err := h.kafkaManager.StopConnection(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Kafka connection stopped"})
}

func (h *DashboardHandler) ListEventMappings(c *gin.Context) {
	if h.eventMappingRepo == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": []models.EventMapping{}})
		return
	}
	mappings, err := h.eventMappingRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": mappings})
}

func (h *DashboardHandler) CreateEventMapping(c *gin.Context) {
	if h.eventMappingRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Event mapping not available"})
		return
	}

	var req models.EventMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.ConnectionID != "" {
		if _, err := h.kafkaConnRepo.GetByID(c.Request.Context(), req.ConnectionID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Connection not found"})
			return
		}
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	mapping := &models.EventMapping{
		EventType:        req.EventType,
		TemplateSlug:    req.TemplateSlug,
		ConnectionID:    req.ConnectionID,
		Description:     req.Description,
		Active:          active,
		FilterConditions: req.FilterConditions,
	}

	if err := h.eventMappingRepo.Create(c.Request.Context(), mapping); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": mapping})
}

func (h *DashboardHandler) UpdateEventMapping(c *gin.Context) {
	if h.eventMappingRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Event mapping not available"})
		return
	}

	id := c.Param("id")
	var req models.EventMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	existing, err := h.eventMappingRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Event mapping not found"})
		return
	}

	if req.EventType != "" {
		existing.EventType = req.EventType
	}
	if req.TemplateSlug != "" {
		existing.TemplateSlug = req.TemplateSlug
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Active != nil {
		existing.Active = *req.Active
	}
	if req.ConnectionID != "" {
		existing.ConnectionID = req.ConnectionID
	}
	if req.FilterConditions != nil {
		existing.FilterConditions = req.FilterConditions
	}

	if err := h.eventMappingRepo.Update(c.Request.Context(), id, existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": existing})
}

func (h *DashboardHandler) DeleteEventMapping(c *gin.Context) {
	if h.eventMappingRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Event mapping not available"})
		return
	}

	id := c.Param("id")
	if err := h.eventMappingRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Event mapping not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Event mapping deleted"})
}

func (h *DashboardHandler) ListKafkaLogs(c *gin.Context) {
	if h.kafkaLogRepo == nil {
		c.JSON(http.StatusOK, models.KafkaLogListResponse{Success: true, Data: []*models.KafkaLog{}, Total: 0, Page: 1, PerPage: 20})
		return
	}

	var filter models.KafkaLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	logs, total, err := h.kafkaLogRepo.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	c.JSON(http.StatusOK, models.KafkaLogListResponse{
		Success: true,
		Data:    logs,
		Total:   total,
		Page:    page,
		PerPage: pageSize,
	})
}

func (h *DashboardHandler) ClearKafkaLogs(c *gin.Context) {
	if h.kafkaLogRepo == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Logs cleared"})
		return
	}
	if err := h.kafkaLogRepo.DeleteAll(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Logs cleared"})
}
