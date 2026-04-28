package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/iZcy/imposizcy/internal/models"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/sirupsen/logrus"
)

type ImageHandler struct {
	imageRepo *repositories.ImageOutputRepository
	validate  *validator.Validate
	logger    *logrus.Logger
}

func NewImageHandler(
	imageRepo *repositories.ImageOutputRepository,
	logger *logrus.Logger,
) *ImageHandler {
	return &ImageHandler{
		imageRepo: imageRepo,
		validate:  validator.New(),
		logger:    logger,
	}
}

func (h *ImageHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	image, err := h.imageRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Image not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Data:    image,
	})
}

func (h *ImageHandler) Download(c *gin.Context) {
	id := c.Param("id")

	image, err := h.imageRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Image not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	if image.FilePath != "" {
		c.Header("Content-Disposition", "attachment; filename=\""+image.Filename+"\"")
		c.File(image.FilePath)
		return
	}

	if image.Base64 != "" {
		c.JSON(http.StatusOK, models.SuccessResponse{
			Success: true,
			Data:    image,
		})
		return
	}

	c.JSON(http.StatusNotFound, models.ErrorResponse{
		Success: false,
		Error:   "Image data not available",
		Code:    http.StatusNotFound,
	})
}

func (h *ImageHandler) List(c *gin.Context) {
	page := int64(1)
	limit := int64(20)

	images, _, err := h.imageRepo.List(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "Failed to list images",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Data:    images,
	})
}

func (h *ImageHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.imageRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "Image not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Image deleted successfully",
	})
}
