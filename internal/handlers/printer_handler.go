package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/iZcy/imposizcy/internal/models"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/iZcy/imposizcy/internal/services"
	"github.com/sirupsen/logrus"
)

type PrinterHandler struct {
	repo        *repositories.PrinterRepository
	cupsService *services.CUPSService
	validate    *validator.Validate
	logger      *logrus.Logger
}

func NewPrinterHandler(repo *repositories.PrinterRepository, cupsService *services.CUPSService, logger *logrus.Logger) *PrinterHandler {
	return &PrinterHandler{
		repo:        repo,
		cupsService: cupsService,
		validate:    validator.New(),
		logger:      logger,
	}
}

func (h *PrinterHandler) Create(c *gin.Context) {
	var req models.CreatePrinterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Error: "Invalid request: " + err.Error(), Code: http.StatusBadRequest})
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Error: "Validation failed: " + err.Error(), Code: http.StatusBadRequest})
		return
	}

	printer := &models.Printer{
		Name:       req.Name,
		Location:   req.Location,
		CupsName:   req.CupsName,
		PaperSizes: req.PaperSizes,
		ColorModes: req.ColorModes,
		IsActive:   req.IsActive,
	}
	if err := h.repo.Create(c.Request.Context(), printer); err != nil {
		h.logger.WithError(err).Error("Failed to create printer")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "Failed to create printer", Code: http.StatusInternalServerError})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{Success: true, Message: "Printer created", Data: printer})
}

func (h *PrinterHandler) List(c *gin.Context) {
	items, err := h.repo.List(c.Request.Context(), 1, 100)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list printers")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "Failed to list printers", Code: http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse{Success: true, Data: items})
}

func (h *PrinterHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Error: "Printer not found", Code: http.StatusNotFound})
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse{Success: true, Data: item})
}

func (h *PrinterHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdatePrinterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Error: "Invalid request: " + err.Error(), Code: http.StatusBadRequest})
		return
	}

	item, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Error: "Printer not found", Code: http.StatusNotFound})
		return
	}

	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Location != nil {
		item.Location = *req.Location
	}
	if req.CupsName != nil {
		item.CupsName = *req.CupsName
	}
	if req.PaperSizes != nil {
		item.PaperSizes = req.PaperSizes
	}
	if req.ColorModes != nil {
		item.ColorModes = req.ColorModes
	}
	if req.IsActive != nil {
		item.IsActive = *req.IsActive
	}
	if req.Status != nil {
		item.Status = models.PrinterStatus(*req.Status)
	}

	if err := h.repo.Update(c.Request.Context(), item); err != nil {
		h.logger.WithError(err).Error("Failed to update printer")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "Failed to update printer", Code: http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse{Success: true, Message: "Printer updated", Data: item})
}

func (h *PrinterHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Error: "Printer not found", Code: http.StatusNotFound})
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse{Success: true, Message: "Printer deleted"})
}

func (h *PrinterHandler) GetStatus(c *gin.Context) {
	id := c.Param("id")
	item, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Error: "Printer not found", Code: http.StatusNotFound})
		return
	}

	status := item.Status
	if h.cupsService.Available() && item.CupsName != "" {
		cupsStatus, err := h.cupsService.GetPrinterStatus(item.CupsName)
		if err == nil {
			status = cupsStatus
			item.Status = cupsStatus
			_ = h.repo.Update(c.Request.Context(), item)
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Success: true, Data: map[string]interface{}{"status": status}})
}

func (h *PrinterHandler) Discover(c *gin.Context) {
	if !h.cupsService.Available() {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Success: false, Error: "CUPS is not available on this system", Code: http.StatusServiceUnavailable,
		})
		return
	}

	printers, err := h.cupsService.ListPrinters()
	if err != nil {
		h.logger.WithError(err).Error("Failed to discover CUPS printers")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "Failed to discover printers", Code: http.StatusInternalServerError})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Success: true, Data: printers})
}
