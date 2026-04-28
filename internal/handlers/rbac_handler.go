package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iZcy/imposizcy/internal/models"
	"github.com/iZcy/imposizcy/internal/repositories"
	"github.com/sirupsen/logrus"
)

type RBACHandler struct {
	internalRoleRepo *repositories.InternalRoleRepository
	externalRoleRepo *repositories.ExternalRoleRepository
	logger           *logrus.Logger
}

func NewRBACHandler(
	internalRoleRepo *repositories.InternalRoleRepository,
	externalRoleRepo *repositories.ExternalRoleRepository,
	logger *logrus.Logger,
) *RBACHandler {
	return &RBACHandler{
		internalRoleRepo: internalRoleRepo,
		externalRoleRepo: externalRoleRepo,
		logger:           logger,
	}
}

func (h *RBACHandler) ListInternalRoles(c *gin.Context) {
	roles, err := h.internalRoleRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": roles})
}

func (h *RBACHandler) CreateInternalRole(c *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		Role        string `json:"role" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	role := &models.InternalRole{
		Username:    req.Username,
		Role:        req.Role,
		Description: req.Description,
	}

	if err := h.internalRoleRepo.Create(c.Request.Context(), role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": role})
}

func (h *RBACHandler) UpdateInternalRole(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Role        string `json:"role"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := h.internalRoleRepo.Update(c.Request.Context(), id, req.Role, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Role updated"})
}

func (h *RBACHandler) DeleteInternalRole(c *gin.Context) {
	id := c.Param("id")
	if err := h.internalRoleRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Role deleted"})
}

func (h *RBACHandler) ListExternalRoles(c *gin.Context) {
	roles, err := h.externalRoleRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": roles})
}

func (h *RBACHandler) CreateExternalRole(c *gin.Context) {
	var req struct {
		ServiceName string   `json:"service_name" binding:"required"`
		Role        string   `json:"role" binding:"required"`
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	role := &models.ExternalRole{
		ServiceName: req.ServiceName,
		Role:        req.Role,
		Permissions: req.Permissions,
	}

	if err := h.externalRoleRepo.Create(c.Request.Context(), role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": role})
}

func (h *RBACHandler) DeleteExternalRole(c *gin.Context) {
	id := c.Param("id")
	if err := h.externalRoleRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "External role deleted"})
}
