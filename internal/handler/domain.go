package handler

import (
	"net/http"
	"strconv"

	"frp-proxy/internal/service"

	"github.com/gin-gonic/gin"
)

type DomainHandler struct {
	domainSvc  *service.DomainService
	baseDomain string
}

func NewDomainHandler(domainSvc *service.DomainService, baseDomain string) *DomainHandler {
	return &DomainHandler{domainSvc: domainSvc, baseDomain: baseDomain}
}

func (h *DomainHandler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"base_domain": h.baseDomain})
}

type CreateDomainRequest struct {
	Subdomain string `json:"subdomain" binding:"required,min=1,max=64"`
}

func (h *DomainHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	domains, err := h.domainSvc.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, domains)
}

func (h *DomainHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var req CreateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain, err := h.domainSvc.Create(userID, req.Subdomain)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, domain)
}

func (h *DomainHandler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.domainSvc.Delete(uint(id), userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
