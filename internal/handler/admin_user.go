package handler

import (
	"net/http"
	"strconv"

	"frp-proxy/internal/model"
	"frp-proxy/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminUserHandler struct {
	userSvc *service.UserService
}

func NewAdminUserHandler(userSvc *service.UserService) *AdminUserHandler {
	return &AdminUserHandler{userSvc: userSvc}
}

type CreateUserRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required,min=6"`
	Role       string `json:"role"`
	MaxDomains int    `json:"max_domains"`
}

type UpdateUserRequest struct {
	MaxDomains *int   `json:"max_domains"`
	Status     string `json:"status"`
	Role       string `json:"role"`
}

func (h *AdminUserHandler) List(c *gin.Context) {
	status := c.Query("status")
	var users []model.User
	var err error
	if status != "" {
		users, err = h.userSvc.ListByStatus(status)
	} else {
		users, err = h.userSvc.List()
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *AdminUserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	validRoles := map[string]bool{"admin": true, "user": true}
	role := req.Role
	if role == "" {
		role = "user"
	}
	if !validRoles[role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	maxDomains := req.MaxDomains
	if maxDomains < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max_domains must be non-negative"})
		return
	}
	if maxDomains == 0 {
		maxDomains = 1
	}
	user, err := h.userSvc.Create(req.Username, req.Password, role, maxDomains)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *AdminUserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	validRoles := map[string]bool{"admin": true, "user": true}
	validStatuses := map[string]bool{"pending": true, "active": true, "disabled": true}

	if req.Role != "" && !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	if req.Status != "" && !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	updates := make(map[string]interface{})
	if req.MaxDomains != nil {
		updates["max_domains"] = *req.MaxDomains
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if err := h.userSvc.Update(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *AdminUserHandler) Activate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.userSvc.Activate(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "activated"})
}

func (h *AdminUserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.userSvc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
