package handler

import (
	"net/http"
	"strconv"
	"time"

	"frp-proxy/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminInviteHandler struct {
	inviteSvc *service.InviteService
}

func NewAdminInviteHandler(inviteSvc *service.InviteService) *AdminInviteHandler {
	return &AdminInviteHandler{inviteSvc: inviteSvc}
}

type CreateInviteRequest struct {
	MaxUses   int `json:"max_uses"`
	ExpiresIn int `json:"expires_in_hours"`
}

func (h *AdminInviteHandler) List(c *gin.Context) {
	codes, err := h.inviteSvc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, codes)
}

func (h *AdminInviteHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var req CreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	maxUses := req.MaxUses
	if maxUses == 0 {
		maxUses = 1
	}
	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)
		expiresAt = &t
	}
	code, err := h.inviteSvc.Create(userID, maxUses, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, code)
}

func (h *AdminInviteHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.inviteSvc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
