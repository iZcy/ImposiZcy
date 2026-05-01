package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/iZcy/imposizcy/internal/models"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/sirupsen/logrus"
)

type TemplateHandler struct {
	templateRepo *repositories.TemplateRepository
	validate     *validator.Validate
	logger       *logrus.Logger
}

func NewTemplateHandler(
	templateRepo *repositories.TemplateRepository,
	logger *logrus.Logger,
) *TemplateHandler {
	return &TemplateHandler{
		templateRepo: templateRepo,
		validate:     validator.New(),
		logger:       logger,
	}
}

func (h *TemplateHandler) Create(c *gin.Context) {
	var req models.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err := h.validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Validation failed: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	template := &models.PrintTemplate{
		Name:            req.Name,
		Slug:            req.Slug,
		HTML:            req.HTML,
		CSS:             req.CSS,
		DataSchema:      req.DataSchema,
		Variables:       req.Variables,
		FieldMapping:    req.FieldMapping,
		BackgroundImage: func() string { if req.BackgroundImage != nil { return *req.BackgroundImage }; return "" }(),
		Width:           req.Width,
		Height:          req.Height,
		DimensionUnit:   req.DimensionUnit,
		DPI:             req.DPI,
		OutputFormat:    models.OutputFormatType(req.OutputFormat),
		Quality:         req.Quality,
		Tags:            req.Tags,
		IsActive:        isActive,
	}

	if err := h.templateRepo.Create(c.Request.Context(), template); err != nil {
		h.logger.WithError(err).Error("Failed to create template")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to create template",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Success: true,
		Message: "Template created successfully",
		Data:    template,
	})
}

func (h *TemplateHandler) List(c *gin.Context) {
	tag := c.Query("tag")

	templates, err := h.templateRepo.List(c.Request.Context(), 1, 100, tag)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list templates")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to list templates",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Data:    templates,
	})
}

func (h *TemplateHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	template, err := h.templateRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Template not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Data:    template,
	})
}

func (h *TemplateHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	template, err := h.templateRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Template not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Data:    template,
	})
}

func (h *TemplateHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	template, err := h.templateRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Template not found",
			Code:    http.StatusNotFound,
		})
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
	if req.Tags != nil {
		template.Tags = req.Tags
	}
	if req.IsActive != nil {
		template.IsActive = *req.IsActive
	}
	if req.Variables != nil {
		template.Variables = req.Variables
	}
	if req.BackgroundImage != nil {
		template.BackgroundImage = *req.BackgroundImage
	}
	if req.FieldMapping != nil {
		template.FieldMapping = req.FieldMapping
	}

	if err := h.templateRepo.Update(c.Request.Context(), template); err != nil {
		h.logger.WithError(err).Error("Failed to update template")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to update template",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Template updated successfully",
		Data:    template,
	})
}

func (h *TemplateHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.templateRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Template not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Template deleted successfully",
	})
}
