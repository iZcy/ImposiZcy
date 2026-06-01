package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/iZcy/imposizcy/internal/models"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/iZcy/imposizcy/internal/services"
	"github.com/sirupsen/logrus"
)

type PrintJobHandler struct {
	repo        *repositories.PrintJobRepository
	printerRepo *repositories.PrinterRepository
	cupsService *services.CUPSService
	validate    *validator.Validate
	logger      *logrus.Logger
	tmpDir      string
}

func NewPrintJobHandler(
	repo *repositories.PrintJobRepository,
	printerRepo *repositories.PrinterRepository,
	cupsService *services.CUPSService,
	logger *logrus.Logger,
) *PrintJobHandler {
	tmpDir := filepath.Join(os.TempDir(), "imposizcy-print")
	os.MkdirAll(tmpDir, 0755)
	return &PrintJobHandler{
		repo:        repo,
		printerRepo: printerRepo,
		cupsService: cupsService,
		validate:    validator.New(),
		logger:      logger,
		tmpDir:      tmpDir,
	}
}

func (h *PrintJobHandler) Create(c *gin.Context) {
	var req models.CreatePrintJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Error: "Invalid request: " + err.Error(), Code: http.StatusBadRequest})
		return
	}
	if err := h.validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Error: "Validation failed: " + err.Error(), Code: http.StatusBadRequest})
		return
	}

	pj := &models.PrintJob{
		TenantID:    req.TenantID,
		OrderID:     req.OrderID,
		PrinterID:   req.PrinterID,
		FileURL:     req.FileURL,
		FilePath:    req.FilePath,
		Copies:      req.Copies,
		PaperSize:   req.PaperSize,
		ColorMode:   req.ColorMode,
		Pages:       req.Pages,
		Sides:       req.Sides,
		CallbackURL: req.CallbackURL,
	}
	if err := h.repo.Create(c.Request.Context(), pj); err != nil {
		h.logger.WithError(err).Error("Failed to create print job")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "Failed to create print job", Code: http.StatusInternalServerError})
		return
	}

	go h.dispatchToCUPS(pj)

	c.JSON(http.StatusCreated, models.SuccessResponse{Success: true, Message: "Print job created", Data: pj})
}

func (h *PrintJobHandler) List(c *gin.Context) {
	items, err := h.repo.List(c.Request.Context(), 1, 100)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list print jobs")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Error: "Failed to list print jobs", Code: http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse{Success: true, Data: items})
}

func (h *PrintJobHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Error: "Print job not found", Code: http.StatusNotFound})
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse{Success: true, Data: item})
}

func (h *PrintJobHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdatePrintJobStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Error: "Invalid request: " + err.Error(), Code: http.StatusBadRequest})
		return
	}

	status := models.PrintJobStatus(req.Status)
	if err := h.repo.UpdateStatus(c.Request.Context(), id, status, req.ErrorMsg); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Error: "Print job not found", Code: http.StatusNotFound})
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse{Success: true, Message: "Status updated"})
}

func (h *PrintJobHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Error: "Print job not found", Code: http.StatusNotFound})
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse{Success: true, Message: "Print job deleted"})
}

func (h *PrintJobHandler) dispatchToCUPS(pj *models.PrintJob) {
	ctx := context.Background()
	jobID := pj.ID.Hex()

	logger := h.logger.WithFields(logrus.Fields{
		"job_id":    jobID,
		"order_id":  pj.OrderID,
		"printer_id": pj.PrinterID,
	})

	if !h.cupsService.Available() {
		h.onPrintFailed(ctx, pj, "CUPS is not available on this system")
		return
	}

	printer, err := h.printerRepo.GetByID(ctx, pj.PrinterID)
	if err != nil {
		h.onPrintFailed(ctx, pj, "Printer not found: "+err.Error())
		return
	}

	if printer.CupsName == "" {
		h.onPrintFailed(ctx, pj, "Printer has no CUPS name configured")
		return
	}

	var filePath string
	if pj.FilePath != "" {
		filePath = pj.FilePath
	} else {
		filePath, err = h.downloadFile(pj.FileURL, jobID)
		if err != nil {
			h.onPrintFailed(ctx, pj, "Failed to download file: "+err.Error())
			return
		}
		defer os.Remove(filePath)
	}

	cupsJobID, err := h.cupsService.SubmitJob(services.SubmitJobOptions{
		CUPSName:  printer.CupsName,
		FilePath:  filePath,
		Copies:    pj.Copies,
		PaperSize: pj.PaperSize,
		ColorMode: pj.ColorMode,
		Sides:     pj.Sides,
	})
	if err != nil {
		h.onPrintFailed(ctx, pj, "CUPS submit failed: "+err.Error())
		return
	}

	if err := h.repo.UpdateCUPSJobID(ctx, jobID, cupsJobID); err != nil {
		logger.WithError(err).Warn("Failed to store CUPS job ID")
	}

	logger = logger.WithField("cups_job_id", cupsJobID)
	logger.Info("Print job submitted to CUPS, polling for completion")

	if err := h.cupsService.PollJobCompletion(ctx, cupsJobID); err != nil {
		h.onPrintFailed(ctx, pj, "Print job failed: "+err.Error())
		return
	}

	h.onPrintDone(ctx, pj)
	logger.Info("Print job completed successfully")
}

func (h *PrintJobHandler) downloadFile(fileURL, jobID string) (string, error) {
	if fileURL == "" {
		return "", fmt.Errorf("no file URL provided")
	}

	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", fileURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s returned status %d", fileURL, resp.StatusCode)
	}

	ext := ".pdf"
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" {
		switch {
		case bytes.Contains([]byte(contentType), []byte("image/png")):
			ext = ".png"
		case bytes.Contains([]byte(contentType), []byte("image/jpeg")):
			ext = ".jpg"
		}
	}

	filePath := filepath.Join(h.tmpDir, jobID+ext)
	f, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("write temp file: %w", err)
	}

	return filePath, nil
}

func (h *PrintJobHandler) onPrintDone(ctx context.Context, pj *models.PrintJob) {
	_ = h.repo.UpdateStatus(ctx, pj.ID.Hex(), models.PrintJobStatusDone, "")
	h.fireCallback(pj, "done")
}

func (h *PrintJobHandler) onPrintFailed(ctx context.Context, pj *models.PrintJob, errMsg string) {
	h.logger.WithFields(logrus.Fields{
		"job_id":   pj.ID.Hex(),
		"order_id": pj.OrderID,
		"error":    errMsg,
	}).Error("Print job failed")
	_ = h.repo.UpdateStatus(ctx, pj.ID.Hex(), models.PrintJobStatusFailed, errMsg)
	h.fireCallback(pj, "failed")
}

func (h *PrintJobHandler) fireCallback(pj *models.PrintJob, status string) {
	if pj.CallbackURL == "" {
		return
	}

	payload := map[string]string{
		"order_id": pj.OrderID,
		"status":   status,
	}
	if status == "failed" {
		payload["error"] = pj.ErrorMsg
	}

	jsonBody, _ := json.Marshal(payload)
	resp, err := http.Post(pj.CallbackURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		h.logger.WithError(err).WithField("callback_url", pj.CallbackURL).Warn("Failed to fire print callback")
		return
	}
	resp.Body.Close()
	h.logger.WithFields(logrus.Fields{
		"callback_url": pj.CallbackURL,
		"status":       resp.StatusCode,
		"job_status":   status,
	}).Info("Print callback fired")
}
