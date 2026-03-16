package handler

import (
	"net/http"

	"frp-proxy/internal/service"

	"github.com/gin-gonic/gin"
)

type PluginHandler struct {
	domainSvc *service.DomainService
	secret    string
}

func NewPluginHandler(domainSvc *service.DomainService, secret string) *PluginHandler {
	return &PluginHandler{domainSvc: domainSvc, secret: secret}
}

type PluginRequest struct {
	Version string                 `json:"version"`
	Op      string                 `json:"op"`
	Content map[string]interface{} `json:"content"`
}

func (h *PluginHandler) Login(c *gin.Context) {
	// Check plugin secret
	if h.secret != "" {
		if c.GetHeader("X-Plugin-Secret") != h.secret {
			c.JSON(http.StatusForbidden, gin.H{"reject": true, "reject_reason": "unauthorized"})
			return
		}
	}

	var req PluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"reject": true, "reject_reason": "invalid request"})
		return
	}

	metas, ok := req.Content["metas"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusOK, gin.H{"reject": true, "reject_reason": "missing metadata"})
		return
	}

	token, ok := metas["token"].(string)
	if !ok || token == "" {
		c.JSON(http.StatusOK, gin.H{"reject": true, "reject_reason": "missing token"})
		return
	}

	if !h.domainSvc.VerifyToken(token) {
		c.JSON(http.StatusOK, gin.H{"reject": true, "reject_reason": "invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reject":   false,
		"unchange": true,
	})
}

func (h *PluginHandler) NewProxy(c *gin.Context) {
	// Check plugin secret
	if h.secret != "" {
		if c.GetHeader("X-Plugin-Secret") != h.secret {
			c.JSON(http.StatusForbidden, gin.H{"reject": true, "reject_reason": "unauthorized"})
			return
		}
	}

	var req PluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"reject": true, "reject_reason": "invalid request"})
		return
	}

	user, ok := req.Content["user"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusOK, gin.H{"reject": true, "reject_reason": "missing user info"})
		return
	}
	metas, ok := user["metas"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusOK, gin.H{"reject": true, "reject_reason": "missing metadata"})
		return
	}

	token, _ := metas["token"].(string)
	subdomain, _ := req.Content["subdomain"].(string)

	if token == "" || subdomain == "" {
		c.JSON(http.StatusOK, gin.H{"reject": true, "reject_reason": "missing token or subdomain"})
		return
	}

	if !h.domainSvc.VerifyTokenSubdomain(token, subdomain) {
		c.JSON(http.StatusOK, gin.H{"reject": true, "reject_reason": "token and subdomain mismatch"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reject":   false,
		"unchange": true,
	})
}
