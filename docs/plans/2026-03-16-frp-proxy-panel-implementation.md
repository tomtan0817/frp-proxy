# frp-proxy Web 管理面板实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 构建一个 Go Web 面板，结合 Nginx + frps，实现用户自助注册二级域名、管理员管理用户和配额、frps plugin 验证的完整系统。

**Architecture:** 单体 Go 应用（Gin + PostgreSQL），提供 REST API + frps plugin 端点。React + Ant Design 前端编译后嵌入 Go 二进制。Nginx 做反向代理分发流量。

**Tech Stack:** Go 1.22+, Gin, GORM, PostgreSQL, JWT, React 18, Ant Design 5, Vite

---

## Task 1: 项目脚手架

**Files:**
- Create: `go.mod`
- Create: `cmd/server/main.go`
- Create: `internal/config/config.go`

**Step 1: 初始化 Go module**

Run: `go mod init frp-proxy`

**Step 2: 创建配置结构**

```go
// internal/config/config.go
package config

import (
	"os"

	"github.com/BurntToast/toml"
)

type Config struct {
	Server   ServerConfig   `toml:"server"`
	Database DatabaseConfig `toml:"database"`
	JWT      JWTConfig      `toml:"jwt"`
	Domain   DomainConfig   `toml:"domain"`
}

type ServerConfig struct {
	Port int    `toml:"port"`
	Host string `toml:"host"`
}

type DatabaseConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBName   string `toml:"dbname"`
	SSLMode  string `toml:"sslmode"`
}

type JWTConfig struct {
	Secret     string `toml:"secret"`
	ExpireHour int    `toml:"expire_hour"`
}

type DomainConfig struct {
	BaseDomain string `toml:"base_domain"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
```

**Step 3: 创建配置文件示例**

```toml
# configs/app.toml
[server]
host = "127.0.0.1"
port = 5040

[database]
host = "127.0.0.1"
port = 5432
user = "frpproxy"
password = "changeme"
dbname = "frpproxy"
sslmode = "disable"

[jwt]
secret = "change-this-to-a-random-string"
expire_hour = 72

[domain]
base_domain = "example.com"
```

**Step 4: 创建最小 main.go**

```go
// cmd/server/main.go
package main

import (
	"flag"
	"fmt"
	"log"

	"frp-proxy/internal/config"
)

func main() {
	cfgPath := flag.String("config", "configs/app.toml", "config file path")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	fmt.Printf("Server starting on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
}
```

**Step 5: 安装依赖并验证编译**

Run: `go mod tidy && go build ./cmd/server`
Expected: 编译成功，无错误

**Step 6: Commit**

```bash
git add go.mod go.sum cmd/ internal/config/ configs/app.toml
git commit -m "feat: project scaffolding with config loading"
```

---

## Task 2: 数据库连接与模型定义

**Files:**
- Create: `internal/database/db.go`
- Create: `internal/model/user.go`
- Create: `internal/model/domain.go`
- Create: `internal/model/invite_code.go`

**Step 1: 创建数据库连接**

```go
// internal/database/db.go
package database

import (
	"frp-proxy/internal/config"
	"frp-proxy/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Domain{},
		&model.InviteCode{},
	)
}
```

**Step 2: 创建模型**

```go
// internal/model/user.go
package model

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         string    `gorm:"size:16;not null;default:user" json:"role"`
	Status       string    `gorm:"size:16;not null;default:pending" json:"status"`
	MaxDomains   int       `gorm:"not null;default:1" json:"max_domains"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Domains      []Domain  `gorm:"foreignKey:UserID" json:"domains,omitempty"`
}
```

```go
// internal/model/domain.go
package model

import "time"

type Domain struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Subdomain string    `gorm:"uniqueIndex;size:64;not null" json:"subdomain"`
	Token     string    `gorm:"uniqueIndex;size:128;not null" json:"token"`
	Status    string    `gorm:"size:16;not null;default:active" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
```

```go
// internal/model/invite_code.go
package model

import "time"

type InviteCode struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Code      string     `gorm:"uniqueIndex;size:64;not null" json:"code"`
	MaxUses   int        `gorm:"not null;default:1" json:"max_uses"`
	UsedCount int        `gorm:"not null;default:0" json:"used_count"`
	CreatedBy uint       `gorm:"not null" json:"created_by"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	Creator   User       `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}
```

**Step 3: 安装 GORM 依赖**

Run: `go get gorm.io/gorm gorm.io/driver/postgres`

**Step 4: 验证编译**

Run: `go build ./...`
Expected: 编译成功

**Step 5: Commit**

```bash
git add internal/database/ internal/model/ go.mod go.sum
git commit -m "feat: database connection and model definitions"
```

---

## Task 3: 写测试 — Auth Service

**Files:**
- Create: `internal/service/auth.go`
- Create: `internal/service/auth_test.go`

**Step 1: 写 auth service 的失败测试**

```go
// internal/service/auth_test.go
package service

import (
	"testing"

	"frp-proxy/internal/model"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&model.User{}, &model.Domain{}, &model.InviteCode{})
	assert.NoError(t, err)
	return db
}

func TestRegisterUser(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAuthService(db, "test-secret", 72)

	user, err := svc.Register("testuser", "password123", "")
	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "pending", user.Status)
}

func TestRegisterWithInviteCode(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAuthService(db, "test-secret", 72)

	// Create an invite code
	db.Create(&model.InviteCode{Code: "INVITE1", MaxUses: 1, UsedCount: 0, CreatedBy: 1})

	user, err := svc.Register("testuser", "password123", "INVITE1")
	assert.NoError(t, err)
	assert.Equal(t, "active", user.Status)

	// Verify invite code used_count incremented
	var code model.InviteCode
	db.Where("code = ?", "INVITE1").First(&code)
	assert.Equal(t, 1, code.UsedCount)
}

func TestRegisterDuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAuthService(db, "test-secret", 72)

	_, err := svc.Register("testuser", "password123", "")
	assert.NoError(t, err)

	_, err = svc.Register("testuser", "password456", "")
	assert.Error(t, err)
}

func TestLogin(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAuthService(db, "test-secret", 72)

	_, err := svc.Register("testuser", "password123", "")
	assert.NoError(t, err)

	// Activate user manually for login test
	db.Model(&model.User{}).Where("username = ?", "testuser").Update("status", "active")

	token, err := svc.Login("testuser", "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLoginPendingUser(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAuthService(db, "test-secret", 72)

	_, err := svc.Register("testuser", "password123", "")
	assert.NoError(t, err)

	_, err = svc.Login("testuser", "password123")
	assert.Error(t, err) // pending user cannot login
}

func TestLoginWrongPassword(t *testing.T) {
	db := setupTestDB(t)
	svc := NewAuthService(db, "test-secret", 72)

	_, err := svc.Register("testuser", "password123", "")
	assert.NoError(t, err)
	db.Model(&model.User{}).Where("username = ?", "testuser").Update("status", "active")

	_, err = svc.Login("testuser", "wrongpass")
	assert.Error(t, err)
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/service/ -v`
Expected: FAIL — NewAuthService 不存在

**Step 3: 实现 AuthService**

```go
// internal/service/auth.go
package service

import (
	"errors"
	"time"

	"frp-proxy/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db         *gorm.DB
	jwtSecret  string
	expireHour int
}

func NewAuthService(db *gorm.DB, jwtSecret string, expireHour int) *AuthService {
	return &AuthService{db: db, jwtSecret: jwtSecret, expireHour: expireHour}
}

func (s *AuthService) Register(username, password, inviteCode string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "user",
		Status:       "pending",
		MaxDomains:   1,
	}

	if inviteCode != "" {
		var code model.InviteCode
		err := s.db.Where("code = ?", inviteCode).First(&code).Error
		if err != nil {
			return nil, errors.New("invalid invite code")
		}
		if code.UsedCount >= code.MaxUses {
			return nil, errors.New("invite code exhausted")
		}
		if code.ExpiresAt != nil && code.ExpiresAt.Before(time.Now()) {
			return nil, errors.New("invite code expired")
		}
		user.Status = "active"

		if err := s.db.Create(&user).Error; err != nil {
			return nil, errors.New("username already exists")
		}

		s.db.Model(&code).Update("used_count", code.UsedCount+1)
		return &user, nil
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, errors.New("username already exists")
	}
	return &user, nil
}

func (s *AuthService) Login(username, password string) (string, error) {
	var user model.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return "", errors.New("invalid credentials")
	}

	if user.Status != "active" {
		return "", errors.New("account not activated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Duration(s.expireHour) * time.Hour).Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ParseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
```

**Step 4: 安装依赖**

Run: `go get github.com/golang-jwt/jwt/v5 golang.org/x/crypto github.com/stretchr/testify gorm.io/driver/sqlite`

**Step 5: 运行测试确认通过**

Run: `go test ./internal/service/ -v -run TestRegister && go test ./internal/service/ -v -run TestLogin`
Expected: ALL PASS

**Step 6: Commit**

```bash
git add internal/service/auth.go internal/service/auth_test.go go.mod go.sum
git commit -m "feat: auth service with register, login, JWT"
```

---

## Task 4: 写测试 — Domain Service

**Files:**
- Create: `internal/service/domain.go`
- Create: `internal/service/domain_test.go`

**Step 1: 写失败测试**

```go
// internal/service/domain_test.go
package service

import (
	"testing"

	"frp-proxy/internal/model"

	"github.com/stretchr/testify/assert"
)

func createTestUser(db *gorm.DB, username, status string, maxDomains int) *model.User {
	user := model.User{
		Username:     username,
		PasswordHash: "hash",
		Role:         "user",
		Status:       status,
		MaxDomains:   maxDomains,
	}
	db.Create(&user)
	return &user
}

func TestCreateDomain(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user := createTestUser(db, "testuser", "active", 2)

	domain, err := svc.Create(user.ID, "myapp")
	assert.NoError(t, err)
	assert.Equal(t, "myapp", domain.Subdomain)
	assert.NotEmpty(t, domain.Token)
	assert.Equal(t, "active", domain.Status)
}

func TestCreateDomainExceedQuota(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user := createTestUser(db, "testuser", "active", 1)

	_, err := svc.Create(user.ID, "first")
	assert.NoError(t, err)

	_, err = svc.Create(user.ID, "second")
	assert.Error(t, err) // quota exceeded
}

func TestCreateDomainDuplicate(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user1 := createTestUser(db, "user1", "active", 2)
	user2 := createTestUser(db, "user2", "active", 2)

	_, err := svc.Create(user1.ID, "myapp")
	assert.NoError(t, err)

	_, err = svc.Create(user2.ID, "myapp")
	assert.Error(t, err) // subdomain already taken
}

func TestListDomainsByUser(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user := createTestUser(db, "testuser", "active", 3)

	svc.Create(user.ID, "app1")
	svc.Create(user.ID, "app2")

	domains, err := svc.ListByUser(user.ID)
	assert.NoError(t, err)
	assert.Len(t, domains, 2)
}

func TestDeleteDomain(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user := createTestUser(db, "testuser", "active", 2)

	domain, _ := svc.Create(user.ID, "myapp")
	err := svc.Delete(domain.ID, user.ID)
	assert.NoError(t, err)

	domains, _ := svc.ListByUser(user.ID)
	assert.Len(t, domains, 0)
}

func TestDeleteDomainWrongUser(t *testing.T) {
	db := setupTestDB(t)
	svc := NewDomainService(db)
	user1 := createTestUser(db, "user1", "active", 2)
	user2 := createTestUser(db, "user2", "active", 2)

	domain, _ := svc.Create(user1.ID, "myapp")
	err := svc.Delete(domain.ID, user2.ID)
	assert.Error(t, err) // not owner
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/service/ -v -run TestCreateDomain`
Expected: FAIL

**Step 3: 实现 DomainService**

```go
// internal/service/domain.go
package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"frp-proxy/internal/model"

	"gorm.io/gorm"
)

type DomainService struct {
	db *gorm.DB
}

func NewDomainService(db *gorm.DB) *DomainService {
	return &DomainService{db: db}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *DomainService) Create(userID uint, subdomain string) (*model.Domain, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	var count int64
	s.db.Model(&model.Domain{}).Where("user_id = ?", userID).Count(&count)
	if int(count) >= user.MaxDomains {
		return nil, errors.New("domain quota exceeded")
	}

	// Check subdomain uniqueness
	var existing model.Domain
	if err := s.db.Where("subdomain = ?", subdomain).First(&existing).Error; err == nil {
		return nil, errors.New("subdomain already taken")
	}

	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	domain := model.Domain{
		UserID:    userID,
		Subdomain: subdomain,
		Token:     token,
		Status:    "active",
	}
	if err := s.db.Create(&domain).Error; err != nil {
		return nil, err
	}
	return &domain, nil
}

func (s *DomainService) ListByUser(userID uint) ([]model.Domain, error) {
	var domains []model.Domain
	err := s.db.Where("user_id = ?", userID).Find(&domains).Error
	return domains, err
}

func (s *DomainService) Delete(domainID, userID uint) error {
	var domain model.Domain
	if err := s.db.First(&domain, domainID).Error; err != nil {
		return errors.New("domain not found")
	}
	if domain.UserID != userID {
		return errors.New("permission denied")
	}
	return s.db.Delete(&domain).Error
}

func (s *DomainService) ListAll() ([]model.Domain, error) {
	var domains []model.Domain
	err := s.db.Preload("User").Find(&domains).Error
	return domains, err
}

func (s *DomainService) AdminCreate(userID uint, subdomain string) (*model.Domain, error) {
	return s.Create(userID, subdomain)
}

func (s *DomainService) AdminUpdate(domainID uint, status string) error {
	return s.db.Model(&model.Domain{}).Where("id = ?", domainID).Update("status", status).Error
}

func (s *DomainService) AdminDelete(domainID uint) error {
	return s.db.Delete(&model.Domain{}, domainID).Error
}

// Plugin verification methods
func (s *DomainService) VerifyToken(token string) bool {
	var domain model.Domain
	err := s.db.Where("token = ? AND status = ?", token, "active").First(&domain).Error
	return err == nil
}

func (s *DomainService) VerifyTokenSubdomain(token, subdomain string) bool {
	var domain model.Domain
	err := s.db.Where("token = ? AND subdomain = ? AND status = ?", token, subdomain, "active").First(&domain).Error
	return err == nil
}
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/service/ -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/service/domain.go internal/service/domain_test.go
git commit -m "feat: domain service with quota check and CRUD"
```

---

## Task 5: User Service + Invite Service

**Files:**
- Create: `internal/service/user.go`
- Create: `internal/service/invite.go`

**Step 1: 实现 UserService（管理员用）**

```go
// internal/service/user.go
package service

import (
	"errors"

	"frp-proxy/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) List() ([]model.User, error) {
	var users []model.User
	err := s.db.Find(&users).Error
	return users, err
}

func (s *UserService) ListByStatus(status string) ([]model.User, error) {
	var users []model.User
	err := s.db.Where("status = ?", status).Find(&users).Error
	return users, err
}

func (s *UserService) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := s.db.Preload("Domains").First(&user, id).Error
	return &user, err
}

func (s *UserService) Create(username, password, role string, maxDomains int) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		Status:       "active",
		MaxDomains:   maxDomains,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, errors.New("username already exists")
	}
	return &user, nil
}

func (s *UserService) Update(id uint, updates map[string]interface{}) error {
	return s.db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

func (s *UserService) Activate(id uint) error {
	return s.db.Model(&model.User{}).Where("id = ?", id).Update("status", "active").Error
}

func (s *UserService) Delete(id uint) error {
	// Delete user's domains first
	s.db.Where("user_id = ?", id).Delete(&model.Domain{})
	return s.db.Delete(&model.User{}, id).Error
}
```

**Step 2: 实现 InviteService**

```go
// internal/service/invite.go
package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"frp-proxy/internal/model"

	"gorm.io/gorm"
)

type InviteService struct {
	db *gorm.DB
}

func NewInviteService(db *gorm.DB) *InviteService {
	return &InviteService{db: db}
}

func (s *InviteService) Create(createdBy uint, maxUses int, expiresAt *time.Time) (*model.InviteCode, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	code := model.InviteCode{
		Code:      hex.EncodeToString(b),
		MaxUses:   maxUses,
		CreatedBy: createdBy,
		ExpiresAt: expiresAt,
	}
	if err := s.db.Create(&code).Error; err != nil {
		return nil, err
	}
	return &code, nil
}

func (s *InviteService) List() ([]model.InviteCode, error) {
	var codes []model.InviteCode
	err := s.db.Preload("Creator").Find(&codes).Error
	return codes, err
}

func (s *InviteService) Delete(id uint) error {
	result := s.db.Delete(&model.InviteCode{}, id)
	if result.RowsAffected == 0 {
		return errors.New("invite code not found")
	}
	return result.Error
}
```

**Step 3: 验证编译**

Run: `go build ./...`
Expected: 编译成功

**Step 4: Commit**

```bash
git add internal/service/user.go internal/service/invite.go
git commit -m "feat: user service and invite code service"
```

---

## Task 6: JWT 中间件

**Files:**
- Create: `internal/middleware/auth.go`
- Create: `internal/middleware/admin.go`

**Step 1: 实现 JWT 认证中间件**

```go
// internal/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"frp-proxy/internal/service"

	"github.com/gin-gonic/gin"
)

func AuthRequired(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := authSvc.ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", uint(claims["user_id"].(float64)))
		c.Set("username", claims["username"].(string))
		c.Set("role", claims["role"].(string))
		c.Next()
	}
}
```

```go
// internal/middleware/admin.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role.(string) != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
```

**Step 2: 验证编译**

Run: `go get github.com/gin-gonic/gin && go build ./...`
Expected: 编译成功

**Step 3: Commit**

```bash
git add internal/middleware/ go.mod go.sum
git commit -m "feat: JWT auth and admin middleware"
```

---

## Task 7: HTTP Handlers — Auth

**Files:**
- Create: `internal/handler/auth.go`

**Step 1: 实现 auth handler**

```go
// internal/handler/auth.go
package handler

import (
	"net/http"

	"frp-proxy/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type RegisterRequest struct {
	Username   string `json:"username" binding:"required,min=3,max=64"`
	Password   string `json:"password" binding:"required,min=6"`
	InviteCode string `json:"invite_code"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authSvc.Register(req.Username, req.Password, req.InviteCode)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "registered successfully",
		"user":    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
```

**Step 2: 验证编译**

Run: `go build ./...`
Expected: 编译成功

**Step 3: Commit**

```bash
git add internal/handler/auth.go
git commit -m "feat: auth handler (register, login)"
```

---

## Task 8: HTTP Handlers — Domain, Admin, Plugin

**Files:**
- Create: `internal/handler/domain.go`
- Create: `internal/handler/admin_user.go`
- Create: `internal/handler/admin_domain.go`
- Create: `internal/handler/admin_invite.go`
- Create: `internal/handler/plugin.go`

**Step 1: 实现 domain handler（用户端）**

```go
// internal/handler/domain.go
package handler

import (
	"net/http"
	"strconv"

	"frp-proxy/internal/service"

	"github.com/gin-gonic/gin"
)

type DomainHandler struct {
	domainSvc *service.DomainService
}

func NewDomainHandler(domainSvc *service.DomainService) *DomainHandler {
	return &DomainHandler{domainSvc: domainSvc}
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
```

**Step 2: 实现 admin handlers**

```go
// internal/handler/admin_user.go
package handler

import (
	"net/http"
	"strconv"

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
	MaxDomains *int    `json:"max_domains"`
	Status     string  `json:"status"`
	Role       string  `json:"role"`
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
	role := req.Role
	if role == "" {
		role = "user"
	}
	maxDomains := req.MaxDomains
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
```

```go
// internal/handler/admin_domain.go
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
```

```go
// internal/handler/admin_invite.go
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
	MaxUses   int    `json:"max_uses"`
	ExpiresIn int    `json:"expires_in_hours"` // hours from now, 0 = never
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
```

**Step 3: 实现 frps plugin handler**

```go
// internal/handler/plugin.go
package handler

import (
	"net/http"

	"frp-proxy/internal/service"

	"github.com/gin-gonic/gin"
)

type PluginHandler struct {
	domainSvc *service.DomainService
}

func NewPluginHandler(domainSvc *service.DomainService) *PluginHandler {
	return &PluginHandler{domainSvc: domainSvc}
}

// frps plugin request format
type PluginRequest struct {
	Version string                 `json:"version"`
	Op      string                 `json:"op"`
	Content map[string]interface{} `json:"content"`
}

func (h *PluginHandler) Login(c *gin.Context) {
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
	var req PluginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"reject": true, "reject_reason": "invalid request"})
		return
	}

	// Get user metadata from the request
	user := req.Content["user"].(map[string]interface{})
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
```

**Step 4: 验证编译**

Run: `go build ./...`
Expected: 编译成功

**Step 5: Commit**

```bash
git add internal/handler/
git commit -m "feat: all HTTP handlers (auth, domain, admin, plugin)"
```

---

## Task 9: 路由注册与 main.go 完善

**Files:**
- Modify: `cmd/server/main.go`

**Step 1: 完善 main.go，注册所有路由**

```go
// cmd/server/main.go
package main

import (
	"flag"
	"fmt"
	"log"

	"frp-proxy/internal/config"
	"frp-proxy/internal/database"
	"frp-proxy/internal/handler"
	"frp-proxy/internal/middleware"
	"frp-proxy/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfgPath := flag.String("config", "configs/app.toml", "config file path")
	initAdmin := flag.Bool("init-admin", false, "create default admin user")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Services
	authSvc := service.NewAuthService(db, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	domainSvc := service.NewDomainService(db)
	userSvc := service.NewUserService(db)
	inviteSvc := service.NewInviteService(db)

	// Init admin if requested
	if *initAdmin {
		_, err := userSvc.Create("admin", "admin123", "admin", 999)
		if err != nil {
			log.Printf("admin user may already exist: %v", err)
		} else {
			log.Println("admin user created (username: admin, password: admin123)")
		}
		return
	}

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	domainH := handler.NewDomainHandler(domainSvc)
	adminUserH := handler.NewAdminUserHandler(userSvc)
	adminDomainH := handler.NewAdminDomainHandler(domainSvc)
	adminInviteH := handler.NewAdminInviteHandler(inviteSvc)
	pluginH := handler.NewPluginHandler(domainSvc)

	r := gin.Default()

	// Public routes
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
		}

		// frps plugin endpoints (no JWT, called by frps internally)
		plugin := api.Group("/plugin")
		{
			plugin.POST("/login", pluginH.Login)
			plugin.POST("/new-proxy", pluginH.NewProxy)
		}
	}

	// Authenticated routes
	authed := api.Group("")
	authed.Use(middleware.AuthRequired(authSvc))
	{
		domains := authed.Group("/domains")
		{
			domains.GET("", domainH.List)
			domains.POST("", domainH.Create)
			domains.DELETE("/:id", domainH.Delete)
		}
	}

	// Admin routes
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(authSvc))
	admin.Use(middleware.AdminRequired())
	{
		users := admin.Group("/users")
		{
			users.GET("", adminUserH.List)
			users.POST("", adminUserH.Create)
			users.PUT("/:id", adminUserH.Update)
			users.PUT("/:id/activate", adminUserH.Activate)
			users.DELETE("/:id", adminUserH.Delete)
		}
		adminDomains := admin.Group("/domains")
		{
			adminDomains.GET("", adminDomainH.List)
			adminDomains.POST("", adminDomainH.Create)
			adminDomains.PUT("/:id", adminDomainH.Update)
			adminDomains.DELETE("/:id", adminDomainH.Delete)
		}
		invites := admin.Group("/invite-codes")
		{
			invites.GET("", adminInviteH.List)
			invites.POST("", adminInviteH.Create)
			invites.DELETE("/:id", adminInviteH.Delete)
		}
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
```

**Step 2: 验证编译**

Run: `go build ./cmd/server`
Expected: 编译成功

**Step 3: Commit**

```bash
git add cmd/server/main.go
git commit -m "feat: main.go with full route registration"
```

---

## Task 10: React 前端脚手架

**Files:**
- Create: `web/` (整个 React 项目)

**Step 1: 用 Vite 创建 React 项目**

Run: `cd web && npm create vite@latest . -- --template react-ts`

**Step 2: 安装依赖**

Run: `cd web && npm install antd @ant-design/icons axios react-router-dom`

**Step 3: 创建项目结构**

```
web/src/
├── api/
│   └── index.ts          # axios 实例 + API 函数
├── components/
│   └── Layout.tsx         # 通用布局
├── pages/
│   ├── Login.tsx
│   ├── Register.tsx
│   ├── Domains.tsx        # 用户域名管理
│   ├── AdminUsers.tsx
│   ├── AdminDomains.tsx
│   └── AdminInvites.tsx
├── App.tsx                # 路由配置
├── main.tsx
└── auth.ts                # token 存取工具
```

**Step 4: Commit**

```bash
git add web/
git commit -m "feat: react frontend scaffolding"
```

---

## Task 11: 前端 — API 层与认证

**Files:**
- Create: `web/src/api/index.ts`
- Create: `web/src/auth.ts`

**Step 1: 实现 API 客户端**

```typescript
// web/src/api/index.ts
import axios from 'axios';
import { getToken } from '../auth';

const api = axios.create({
  baseURL: '/api',
});

api.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(err);
  }
);

// Auth
export const login = (username: string, password: string) =>
  api.post('/auth/login', { username, password });
export const register = (username: string, password: string, invite_code?: string) =>
  api.post('/auth/register', { username, password, invite_code });

// Domains
export const getDomains = () => api.get('/domains');
export const createDomain = (subdomain: string) => api.post('/domains', { subdomain });
export const deleteDomain = (id: number) => api.delete(`/domains/${id}`);

// Admin Users
export const getUsers = (status?: string) =>
  api.get('/admin/users', { params: status ? { status } : {} });
export const createUser = (data: any) => api.post('/admin/users', data);
export const updateUser = (id: number, data: any) => api.put(`/admin/users/${id}`, data);
export const activateUser = (id: number) => api.put(`/admin/users/${id}/activate`);
export const deleteUser = (id: number) => api.delete(`/admin/users/${id}`);

// Admin Domains
export const getAllDomains = () => api.get('/admin/domains');
export const adminCreateDomain = (data: any) => api.post('/admin/domains', data);
export const adminUpdateDomain = (id: number, data: any) => api.put(`/admin/domains/${id}`, data);
export const adminDeleteDomain = (id: number) => api.delete(`/admin/domains/${id}`);

// Invite Codes
export const getInviteCodes = () => api.get('/admin/invite-codes');
export const createInviteCode = (data: any) => api.post('/admin/invite-codes', data);
export const deleteInviteCode = (id: number) => api.delete(`/admin/invite-codes/${id}`);

export default api;
```

```typescript
// web/src/auth.ts
export const getToken = () => localStorage.getItem('token');
export const setToken = (token: string) => localStorage.setItem('token', token);
export const removeToken = () => localStorage.removeItem('token');

export const parseJwt = (token: string) => {
  try {
    return JSON.parse(atob(token.split('.')[1]));
  } catch {
    return null;
  }
};

export const getUser = () => {
  const token = getToken();
  if (!token) return null;
  const payload = parseJwt(token);
  if (!payload || payload.exp * 1000 < Date.now()) {
    removeToken();
    return null;
  }
  return payload;
};

export const isAdmin = () => getUser()?.role === 'admin';
```

**Step 2: Commit**

```bash
git add web/src/api/ web/src/auth.ts
git commit -m "feat: frontend API client and auth utilities"
```

---

## Task 12: 前端 — 页面实现

**Files:**
- Create: `web/src/App.tsx`
- Create: `web/src/components/Layout.tsx`
- Create: `web/src/pages/Login.tsx`
- Create: `web/src/pages/Register.tsx`
- Create: `web/src/pages/Domains.tsx`
- Create: `web/src/pages/AdminUsers.tsx`
- Create: `web/src/pages/AdminDomains.tsx`
- Create: `web/src/pages/AdminInvites.tsx`

每个页面的实现细节较长，核心要点：

- **Login.tsx**: 表单 → 调用 login API → setToken → 跳转
- **Register.tsx**: 表单（含可选邀请码）→ 调用 register API → 提示结果
- **Domains.tsx**: Table 展示域名列表、Modal 创建域名、复制 token 按钮、frpc 配置示例展示
- **AdminUsers.tsx**: Table + 筛选 pending → 激活按钮、编辑 max_domains、创建/删除用户
- **AdminDomains.tsx**: Table 所有域名 → 启用/禁用/删除
- **AdminInvites.tsx**: Table 邀请码列表 → 生成/删除、复制邀请码
- **Layout.tsx**: Ant Design Layout，侧栏导航，admin 菜单仅 admin 可见
- **App.tsx**: react-router-dom 路由配置，登录/注册公开，其余需认证

**Step 1: 实现所有页面**
（每个页面按上述要点实现，使用 Ant Design 的 Table、Modal、Form、Button、message 组件）

**Step 2: 验证前端编译**

Run: `cd web && npm run build`
Expected: 编译成功，输出到 `web/dist/`

**Step 3: Commit**

```bash
git add web/src/
git commit -m "feat: all frontend pages (login, register, domains, admin)"
```

---

## Task 13: Go 嵌入前端静态文件

**Files:**
- Modify: `cmd/server/main.go`
- Create: `cmd/server/static.go`

**Step 1: 创建静态文件嵌入**

```go
// cmd/server/static.go
package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed all:../../web/dist
var webFS embed.FS

func serveStatic(r *gin.Engine) {
	distFS, _ := fs.Sub(webFS, "web/dist")
	fileServer := http.FileServer(http.FS(distFS))

	r.NoRoute(func(c *gin.Context) {
		// Try to serve static file first
		f, err := distFS.Open(c.Request.URL.Path[1:])
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		// Fallback to index.html for SPA routing
		c.FileFromFS("index.html", http.FS(distFS))
	})
}
```

**Step 2: 在 main.go 的 `r.Run()` 前添加调用**

```go
serveStatic(r)
```

**Step 3: 构建前端并编译 Go**

Run: `cd web && npm run build && cd .. && go build ./cmd/server`
Expected: 编译成功，单个二进制文件

**Step 4: Commit**

```bash
git add cmd/server/static.go cmd/server/main.go
git commit -m "feat: embed React frontend into Go binary"
```

---

## Task 14: 配置文件示例

**Files:**
- Modify: `configs/app.toml`（已有）
- Create: `configs/frps.toml`
- Create: `configs/frpc-example.toml`
- Create: `configs/nginx.conf`

**Step 1: 创建 frps 配置示例**

```toml
# configs/frps.toml
bindPort = 7000
vhostHTTPPort = 8081
subDomainHost = "example.com"

[httpPlugins]
  [httpPlugins.login]
    addr = "http://127.0.0.1:5040/api/plugin/login"
    path = "/api/plugin/login"
    ops = ["Login"]

  [httpPlugins.new-proxy]
    addr = "http://127.0.0.1:5040/api/plugin/new-proxy"
    path = "/api/plugin/new-proxy"
    ops = ["NewProxy"]
```

**Step 2: 创建 frpc 配置示例**

```toml
# configs/frpc-example.toml
serverAddr = "example.com"
serverPort = 7000
metadatas.token = "your-token-from-panel"

[[proxies]]
name = "web"
type = "http"
localPort = 3000
subdomain = "your-subdomain"
```

**Step 3: 创建 Nginx 配置示例**

```nginx
# configs/nginx.conf
# panel.example.com -> Go panel
server {
    listen 80;
    server_name panel.example.com;

    location / {
        proxy_pass http://127.0.0.1:5040;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}

# *.example.com -> frps HTTP proxy
server {
    listen 80;
    server_name *.example.com;

    location / {
        proxy_pass http://127.0.0.1:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Step 4: Commit**

```bash
git add configs/
git commit -m "feat: config examples for frps, frpc, nginx"
```

---

## Task 15: 集成测试 + 最终验证

**Step 1: 确保 PostgreSQL 运行，创建数据库**

Run: `createdb frpproxy` (或通过 psql)

**Step 2: 初始化管理员**

Run: `./server -config configs/app.toml -init-admin`
Expected: "admin user created"

**Step 3: 启动服务**

Run: `./server -config configs/app.toml`
Expected: "Server starting on 127.0.0.1:5040"

**Step 4: 测试 API**

```bash
# 登录
curl -X POST http://localhost:5040/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}'

# 创建邀请码 (用返回的 token)
curl -X POST http://localhost:5040/api/admin/invite-codes \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d '{"max_uses":10}'

# 注册新用户
curl -X POST http://localhost:5040/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"testuser","password":"pass123","invite_code":"<code>"}'

# 用新用户创建域名
curl -X POST http://localhost:5040/api/domains \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d '{"subdomain":"myapp"}'
```

**Step 5: 验证所有测试通过**

Run: `go test ./... -v`
Expected: ALL PASS

**Step 6: Final commit**

```bash
git add -A
git commit -m "feat: integration tests and final verification"
```
