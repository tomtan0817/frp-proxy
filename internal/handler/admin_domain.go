package handler

import (
	"net/http"
	"strconv"

	"frp-proxy/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminDomainHandler struct {
	domainSvc *service.DomainService
}

func NewAdminDomainHandler(domainSvc *service.DomainService) *AdminDomainHandler {
	return &AdminDomainHandler{domainSvc: domainSvc}
}

type AdminCreateDomainRequest struct {
	UserID    uint   `json:"user_id" binding:"required"`
	Subdomain string `json:"subdomain" binding:"required"`
}

type AdminUpdateDomainRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *AdminDomainHandler) List(c *gin.Context) {
	domains, err := h.domainSvc.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, domains)
}

func (h *AdminDomainHandler) Create(c *gin.Context) {
	var req AdminCreateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	domain, err := h.domainSvc.AdminCreate(req.UserID, req.Subdomain)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, domain)
}

func (h *AdminDomainHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req AdminUpdateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.domainSvc.AdminUpdate(uint(id), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *AdminDomainHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.domainSvc.AdminDelete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
