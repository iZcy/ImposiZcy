package handlers

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/iZcy/imposizcy/internal/models"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/iZcy/imposizcy/internal/services"
	"github.com/sirupsen/logrus"
)

type RenderHandler struct {
	templateRepo    *repositories.TemplateRepository
	renderJobRepo   *repositories.RenderJobRepository
	imageOutputRepo *repositories.ImageOutputRepository
	renderer        *services.RendererService
	validator       *services.ValidatorService
	imageGenerator  *services.ImageGeneratorService
	wsHandler       *WebSocketHandler
	validate        *validator.Validate
	logger          *logrus.Logger
}

func NewRenderHandler(
	templateRepo *repositories.TemplateRepository,
	renderJobRepo *repositories.RenderJobRepository,
	imageOutputRepo *repositories.ImageOutputRepository,
	renderer *services.RendererService,
	validatorSvc *services.ValidatorService,
	imageGenerator *services.ImageGeneratorService,
	wsHandler *WebSocketHandler,
	logger *logrus.Logger,
) *RenderHandler {
	return &RenderHandler{
		templateRepo:    templateRepo,
		renderJobRepo:   renderJobRepo,
		imageOutputRepo: imageOutputRepo,
		renderer:        renderer,
		validator:       validatorSvc,
		imageGenerator:  imageGenerator,
		wsHandler:       wsHandler,
		validate:        validator.New(),
		logger:          logger,
	}
}

func (h *RenderHandler) Render(c *gin.Context) {
	var req models.RenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	template, err := h.templateRepo.GetBySlug(c.Request.Context(), req.TemplateSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Template not found: " + req.TemplateSlug,
			Code:    http.StatusNotFound,
		})
		return
	}

	if !template.IsActive {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Template is not active",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if template.DataSchema != "" {
		valid, errors, err := h.validator.ValidateData(template.DataSchema, req.Data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success: false,
				Error:   "Schema validation error: " + err.Error(),
				Code:    http.StatusInternalServerError,
			})
			return
		}
		if !valid {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success: false,
				Error:   "Data validation failed",
				Code:    http.StatusBadRequest,
				Data:    errors,
			})
			return
		}
	}

	renderedHTML, err := h.renderer.RenderHTML(template.HTML, template.CSS, template.Variables, req.Data)
	if err != nil {
		h.logger.WithError(err).Error("Failed to render template")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to render template: " + err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	outputFormat := req.OutputFormat
	if outputFormat == "" {
		outputFormat = string(template.OutputFormat)
	}
	if outputFormat == "" {
		outputFormat = string(models.RenderFormatPNG)
	}

	width := int(template.Width)
	height := int(template.Height)
	if req.Width > 0 {
		width = req.Width
	}
	if req.Height > 0 {
		height = req.Height
	}

	imgBytes, err := h.imageGenerator.GenerateFromHTML(c.Request.Context(), renderedHTML, &models.RenderOptions{
		Width:   width,
		Height:  height,
		DPI:     template.DPI,
		Format:  outputFormat,
		Quality: template.Quality,
	})
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate image")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to generate image: " + err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	now := time.Now()
	imageOutput := &models.ImageOutput{
		TemplateID:   template.ID.Hex(),
		TemplateSlug: template.Slug,
		Format:       models.RenderOutputFormat(outputFormat),
		Width:        width,
		Height:       height,
		DPI:          template.DPI,
		FileSize:     len(imgBytes),
		Data:         req.Data,
		Base64:       base64.StdEncoding.EncodeToString(imgBytes),
		Status:       models.RenderStatusCompleted,
		CompletedAt:  &now,
	}

	if err := h.imageOutputRepo.Create(c.Request.Context(), imageOutput); err != nil {
		h.logger.WithError(err).Warn("Failed to save image output metadata")
	}

	if h.wsHandler != nil {
		h.wsHandler.Broadcast(map[string]interface{}{
			"event":       "render.completed",
			"template_id": template.ID.Hex(),
			"image_id":    imageOutput.ID.Hex(),
		})
	}

	if req.Mode == "" || req.Mode == "base64" {
		c.JSON(http.StatusOK, models.SuccessResponse{
			Success: true,
			Data: map[string]interface{}{
				"image_id":  imageOutput.ID.Hex(),
				"base64":    imageOutput.Base64,
				"format":    outputFormat,
				"width":     width,
				"height":    height,
				"dpi":       template.DPI,
				"file_size": len(imgBytes),
			},
		})
		return
	}

	contentType := "image/png"
	fileExt := ".png"
	if outputFormat == string(models.RenderFormatJPEG) {
		contentType = "image/jpeg"
		fileExt = ".jpg"
	}

	c.Header("Content-Disposition", "inline; filename=\""+template.Slug+fileExt+"\"")
	c.Data(http.StatusOK, contentType, imgBytes)
}

func (h *RenderHandler) RenderAsync(c *gin.Context) {
	var req models.RenderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	template, err := h.templateRepo.GetBySlug(c.Request.Context(), req.TemplateSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Template not found: " + req.TemplateSlug,
			Code:    http.StatusNotFound,
		})
		return
	}

	if !template.IsActive {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Template is not active",
			Code:    http.StatusBadRequest,
		})
		return
	}

	outputFormat := req.OutputFormat
	if outputFormat == "" {
		outputFormat = string(template.OutputFormat)
	}
	if outputFormat == "" {
		outputFormat = string(models.RenderFormatPNG)
	}

	job := &models.RenderJob{
		TemplateID:   template.ID,
		TemplateSlug: template.Slug,
		Data:         req.Data,
		OutputFormat: models.RenderOutputFormat(outputFormat),
		Width:        int(template.Width),
		Height:       int(template.Height),
		Status:       models.RenderStatusPending,
		Source:       "api",
	}

	if req.Width > 0 {
		job.Width = req.Width
	}
	if req.Height > 0 {
		job.Height = req.Height
	}

	if err := h.renderJobRepo.Create(c.Request.Context(), job); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to create render job: " + err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	go h.processJobAsync(job, template)

	c.JSON(http.StatusAccepted, models.SuccessResponse{
		Success: true,
		Data: map[string]interface{}{
			"job_id": job.ID.Hex(),
			"status": job.Status,
		},
	})
}

func (h *RenderHandler) processJobAsync(job *models.RenderJob, template *models.PrintTemplate) {
	if template.DataSchema != "" {
		valid, validationErrors, err := h.validator.ValidateData(template.DataSchema, job.Data)
		if err != nil || !valid {
			job.Status = models.RenderStatusFailed
			if err != nil {
				job.Error = err.Error()
			} else {
				job.Error = "Data validation failed"
				_ = validationErrors
			}
			_ = h.renderJobRepo.Update(context.Background(), job)
			return
		}
	}

	renderedHTML, err := h.renderer.RenderHTML(template.HTML, template.CSS, template.Variables, job.Data)
	if err != nil {
		job.Status = models.RenderStatusFailed
		job.Error = err.Error()
		_ = h.renderJobRepo.Update(context.Background(), job)
		return
	}

	imgBytes, err := h.imageGenerator.GenerateFromHTML(context.Background(), renderedHTML, &models.RenderOptions{
		Width:   job.Width,
		Height:  job.Height,
		DPI:     template.DPI,
		Format:  string(job.OutputFormat),
		Quality: template.Quality,
	})
	if err != nil {
		job.Status = models.RenderStatusFailed
		job.Error = err.Error()
		_ = h.renderJobRepo.Update(context.Background(), job)
		return
	}

	now := time.Now()
	imageOutput := &models.ImageOutput{
		TemplateID:   template.ID.Hex(),
		TemplateSlug: template.Slug,
		Format:       models.RenderOutputFormat(job.OutputFormat),
		Width:        job.Width,
		Height:       job.Height,
		DPI:          template.DPI,
		FileSize:     len(imgBytes),
		Data:         job.Data,
		Base64:       base64.StdEncoding.EncodeToString(imgBytes),
		Status:       models.RenderStatusCompleted,
		CompletedAt:  &now,
	}

	if err := h.imageOutputRepo.Create(context.Background(), imageOutput); err != nil {
		h.logger.WithError(err).Error("Failed to save async image output")
	}

	job.Status = models.RenderStatusCompleted
	job.OutputImageID = imageOutput.ID.Hex()
	job.OutputSize = len(imgBytes)
	_ = h.renderJobRepo.Update(context.Background(), job)

	if h.wsHandler != nil {
		h.wsHandler.Broadcast(map[string]interface{}{
			"event":       "render.completed",
			"job_id":      job.ID.Hex(),
			"template_id": template.ID.Hex(),
			"image_id":    imageOutput.ID.Hex(),
		})
	}
}

func (h *RenderHandler) GetJobStatus(c *gin.Context) {
	id := c.Param("id")

	job, err := h.renderJobRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Job not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Data:    job,
	})
}
