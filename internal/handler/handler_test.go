package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"frp-proxy/internal/database"
	"frp-proxy/internal/handler"
	"frp-proxy/internal/middleware"
	"frp-proxy/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	testJWTSecret  = "test-secret-key"
	testExpireHour = 24
	testBaseDomain = "example.com"
	testPluginSecret = "plugin-secret-123"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.AutoMigrate(db))

	authSvc := service.NewAuthService(db, testJWTSecret, testExpireHour)
	domainSvc := service.NewDomainService(db)
	userSvc := service.NewUserService(db)
	inviteSvc := service.NewInviteService(db)

	authH := handler.NewAuthHandler(authSvc)
	domainH := handler.NewDomainHandler(domainSvc, testBaseDomain)
	adminUserH := handler.NewAdminUserHandler(userSvc)
	adminDomainH := handler.NewAdminDomainHandler(domainSvc)
	adminInviteH := handler.NewAdminInviteHandler(inviteSvc)
	pluginH := handler.NewPluginHandler(domainSvc, testPluginSecret)

	r := gin.New()

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
		}

		api.GET("/config", domainH.GetConfig)

		plugin := api.Group("/plugin")
		{
			plugin.POST("/login", pluginH.Login)
			plugin.POST("/new-proxy", pluginH.NewProxy)
		}
	}

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

	return r
}

// loginAsAdmin creates an admin user via the admin service, then logs in and returns the JWT token.
func loginAsAdmin(t *testing.T, router *gin.Engine) string {
	t.Helper()

	// First register a normal user (we need a DB-level admin, so we use admin create endpoint workaround).
	// Instead, we create admin by registering, then use a second router with direct DB access.
	// Simpler: register user, then create admin via a fresh setup. But we share the same router/db.
	// The cleanest approach: create an admin user via the UserService directly.
	// Since we can't access the DB from here, we'll use a two-step approach:
	// 1. Register a user
	// 2. The user is pending, so we need another way.

	// Actually, the simplest way is to register a user, and since we need an admin,
	// we'll create the admin using the same pattern as main.go --init-admin.
	// But we don't have direct DB access from here.

	// Let's use a helper that embeds the DB in the router setup.
	// For simplicity, we'll create a dedicated helper that sets up everything.

	// We'll use a trick: register a user, then we need to activate them and make them admin.
	// But we can't do that without an admin. Chicken-and-egg problem.

	// Solution: use setupTestRouterWithDB that returns both router and db.
	t.Fatal("use loginAsAdminWithDB instead")
	return ""
}

// setupTestRouterWithDB returns router and the underlying DB for test helpers.
func setupTestRouterWithDB(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.AutoMigrate(db))

	authSvc := service.NewAuthService(db, testJWTSecret, testExpireHour)
	domainSvc := service.NewDomainService(db)
	userSvc := service.NewUserService(db)
	inviteSvc := service.NewInviteService(db)

	authH := handler.NewAuthHandler(authSvc)
	domainH := handler.NewDomainHandler(domainSvc, testBaseDomain)
	adminUserH := handler.NewAdminUserHandler(userSvc)
	adminDomainH := handler.NewAdminDomainHandler(domainSvc)
	adminInviteH := handler.NewAdminInviteHandler(inviteSvc)
	pluginH := handler.NewPluginHandler(domainSvc, testPluginSecret)

	r := gin.New()

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
		}

		api.GET("/config", domainH.GetConfig)

		plugin := api.Group("/plugin")
		{
			plugin.POST("/login", pluginH.Login)
			plugin.POST("/new-proxy", pluginH.NewProxy)
		}
	}

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

	return r, db
}

// createAdminToken creates an admin user in the DB and logs in, returning the JWT token.
func createAdminToken(t *testing.T, router *gin.Engine, db *gorm.DB) string {
	t.Helper()
	userSvc := service.NewUserService(db)
	_, err := userSvc.Create("admin", "admin123", "admin", 999)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	body := `{"username":"admin","password":"admin123"}`
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	token, ok := resp["token"].(string)
	require.True(t, ok)
	return token
}

// createActiveUserToken registers a user, activates them via DB, logs in, and returns the JWT token.
func createActiveUserToken(t *testing.T, router *gin.Engine, db *gorm.DB, username string) string {
	t.Helper()
	userSvc := service.NewUserService(db)
	_, err := userSvc.Create(username, "password123", "user", 1)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	body := fmt.Sprintf(`{"username":"%s","password":"password123"}`, username)
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	token, ok := resp["token"].(string)
	require.True(t, ok)
	return token
}

// ======================== Auth Tests ========================

func TestRegisterSuccess_PendingStatus(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	body := `{"username":"testuser","password":"pass123456"}`
	req, _ := http.NewRequest("POST", "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	user := resp["user"].(map[string]interface{})
	assert.Equal(t, "pending", user["status"])
	assert.Equal(t, "testuser", user["username"])
}

func TestRegisterWithValidInviteCode_ActiveStatus(t *testing.T) {
	router, db := setupTestRouterWithDB(t)

	// Create an admin to own the invite code
	userSvc := service.NewUserService(db)
	admin, err := userSvc.Create("admin", "admin123", "admin", 999)
	require.NoError(t, err)

	// Create an invite code
	inviteSvc := service.NewInviteService(db)
	code, err := inviteSvc.Create(admin.ID, 5, nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	body := fmt.Sprintf(`{"username":"inviteduser","password":"pass123456","invite_code":"%s"}`, code.Code)
	req, _ := http.NewRequest("POST", "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	user := resp["user"].(map[string]interface{})
	assert.Equal(t, "active", user["status"])
}

func TestRegisterWithInvalidInviteCode_Error(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	body := `{"username":"testuser","password":"pass123456","invite_code":"invalid-code"}`
	req, _ := http.NewRequest("POST", "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "invalid invite code")
}

func TestRegisterDuplicateUsername_Error(t *testing.T) {
	router := setupTestRouter(t)

	body := `{"username":"dupuser","password":"pass123456"}`

	// First registration
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Second registration with same username
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestLoginSuccess_ActiveUser(t *testing.T) {
	router, db := setupTestRouterWithDB(t)

	// Create an active user
	userSvc := service.NewUserService(db)
	_, err := userSvc.Create("activeuser", "pass123456", "user", 1)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	body := `{"username":"activeuser","password":"pass123456"}`
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["token"])
}

func TestLoginPendingUser_Rejected(t *testing.T) {
	router := setupTestRouter(t)

	// Register a user (will be pending)
	w := httptest.NewRecorder()
	body := `{"username":"pendinguser","password":"pass123456"}`
	req, _ := http.NewRequest("POST", "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Try to login
	w = httptest.NewRecorder()
	body = `{"username":"pendinguser","password":"pass123456"}`
	req, _ = http.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "not activated")
}

func TestLoginWrongPassword(t *testing.T) {
	router, db := setupTestRouterWithDB(t)

	userSvc := service.NewUserService(db)
	_, err := userSvc.Create("wrongpwuser", "correctpass", "user", 1)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	body := `{"username":"wrongpwuser","password":"wrongpass"}`
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ======================== Domain Tests ========================

func TestDomainsRequireAuth(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/domains", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateDomainSuccess(t *testing.T) {
	router, db := setupTestRouterWithDB(t)
	token := createActiveUserToken(t, router, db, "domainuser")

	w := httptest.NewRecorder()
	body := `{"subdomain":"myapp"}`
	req, _ := http.NewRequest("POST", "/api/domains", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "myapp", resp["subdomain"])
	assert.NotEmpty(t, resp["token"])
}

func TestCreateDomainExceedQuota(t *testing.T) {
	router, db := setupTestRouterWithDB(t)
	token := createActiveUserToken(t, router, db, "quotauser")

	// Create the first domain (max_domains=1)
	w := httptest.NewRecorder()
	body := `{"subdomain":"first"}`
	req, _ := http.NewRequest("POST", "/api/domains", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Try to create a second domain
	w = httptest.NewRecorder()
	body = `{"subdomain":"second"}`
	req, _ = http.NewRequest("POST", "/api/domains", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "quota exceeded")
}

func TestCreateDomainDuplicateSubdomain(t *testing.T) {
	router, db := setupTestRouterWithDB(t)

	// Create two users, each with quota of 1
	// User1 creates "shared", User2 tries the same subdomain
	userSvc := service.NewUserService(db)
	_, err := userSvc.Create("user1", "password123", "user", 2)
	require.NoError(t, err)
	_, err = userSvc.Create("user2", "password123", "user", 2)
	require.NoError(t, err)

	// Login as user1
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"username":"user1","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	var resp1 map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp1)
	token1 := resp1["token"].(string)

	// Login as user2
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"username":"user2","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	var resp2 map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp2)
	token2 := resp2["token"].(string)

	// User1 creates "shared"
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/domains", strings.NewReader(`{"subdomain":"shared"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token1)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// User2 tries same subdomain
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/domains", strings.NewReader(`{"subdomain":"shared"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token2)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "already taken")
}

func TestCreateDomainInvalidSubdomainFormat(t *testing.T) {
	router, db := setupTestRouterWithDB(t)
	token := createActiveUserToken(t, router, db, "fmtuser")

	invalidSubdomains := []string{
		"MY APP",   // uppercase and space
		"my.app",   // dot
		"-bad",     // leading hyphen
		"bad-",     // trailing hyphen
		"MY_APP",   // uppercase and underscore
	}

	for _, sub := range invalidSubdomains {
		t.Run(sub, func(t *testing.T) {
			w := httptest.NewRecorder()
			body := fmt.Sprintf(`{"subdomain":"%s"}`, sub)
			req, _ := http.NewRequest("POST", "/api/domains", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusConflict, w.Code, "subdomain %q should be rejected", sub)
		})
	}
}

func TestDeleteDomainSuccess(t *testing.T) {
	router, db := setupTestRouterWithDB(t)
	token := createActiveUserToken(t, router, db, "deluser")

	// Create a domain
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/domains", strings.NewReader(`{"subdomain":"todelete"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var domain map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &domain)
	domainID := int(domain["id"].(float64))

	// Delete it
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/domains/%d", domainID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteDomainWrongUser_Forbidden(t *testing.T) {
	router, db := setupTestRouterWithDB(t)

	// Create two users with quota
	userSvc := service.NewUserService(db)
	_, err := userSvc.Create("owner", "password123", "user", 5)
	require.NoError(t, err)
	_, err = userSvc.Create("other", "password123", "user", 5)
	require.NoError(t, err)

	// Login as owner
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"username":"owner","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	var r1 map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &r1)
	ownerToken := r1["token"].(string)

	// Login as other
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"username":"other","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	var r2 map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &r2)
	otherToken := r2["token"].(string)

	// Owner creates a domain
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/domains", strings.NewReader(`{"subdomain":"ownerdomain"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var domain map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &domain)
	domainID := int(domain["id"].(float64))

	// Other user tries to delete it
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/domains/%d", domainID), nil)
	req.Header.Set("Authorization", "Bearer "+otherToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ======================== Plugin Tests ========================

func TestPluginLoginValidToken(t *testing.T) {
	router, db := setupTestRouterWithDB(t)

	// Create a user and a domain to get a valid token
	userSvc := service.NewUserService(db)
	user, err := userSvc.Create("pluginuser", "password123", "user", 5)
	require.NoError(t, err)

	domainSvc := service.NewDomainService(db)
	domain, err := domainSvc.Create(user.ID, "plugintest")
	require.NoError(t, err)

	w := httptest.NewRecorder()
	body := fmt.Sprintf(`{
		"version":"0.1.0",
		"op":"Login",
		"content":{
			"metas":{"token":"%s"}
		}
	}`, domain.Token)
	req, _ := http.NewRequest("POST", "/api/plugin/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Plugin-Secret", testPluginSecret)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["reject"])
}

func TestPluginLoginInvalidToken(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	body := `{
		"version":"0.1.0",
		"op":"Login",
		"content":{
			"metas":{"token":"invalid-token-value"}
		}
	}`
	req, _ := http.NewRequest("POST", "/api/plugin/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Plugin-Secret", testPluginSecret)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["reject"])
	assert.Contains(t, resp["reject_reason"], "invalid token")
}

func TestPluginNewProxyValidTokenSubdomain(t *testing.T) {
	router, db := setupTestRouterWithDB(t)

	userSvc := service.NewUserService(db)
	user, err := userSvc.Create("proxyuser", "password123", "user", 5)
	require.NoError(t, err)

	domainSvc := service.NewDomainService(db)
	domain, err := domainSvc.Create(user.ID, "proxytest")
	require.NoError(t, err)

	w := httptest.NewRecorder()
	body := fmt.Sprintf(`{
		"version":"0.1.0",
		"op":"NewProxy",
		"content":{
			"user":{"metas":{"token":"%s"}},
			"subdomain":"proxytest"
		}
	}`, domain.Token)
	req, _ := http.NewRequest("POST", "/api/plugin/new-proxy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Plugin-Secret", testPluginSecret)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["reject"])
}

func TestPluginNewProxyMismatchedTokenSubdomain(t *testing.T) {
	router, db := setupTestRouterWithDB(t)

	userSvc := service.NewUserService(db)
	user, err := userSvc.Create("mismatchuser", "password123", "user", 5)
	require.NoError(t, err)

	domainSvc := service.NewDomainService(db)
	domain, err := domainSvc.Create(user.ID, "correctsub")
	require.NoError(t, err)

	w := httptest.NewRecorder()
	body := fmt.Sprintf(`{
		"version":"0.1.0",
		"op":"NewProxy",
		"content":{
			"user":{"metas":{"token":"%s"}},
			"subdomain":"wrongsub"
		}
	}`, domain.Token)
	req, _ := http.NewRequest("POST", "/api/plugin/new-proxy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Plugin-Secret", testPluginSecret)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["reject"])
	assert.Contains(t, resp["reject_reason"], "mismatch")
}

func TestPluginSecretHeaderCheck(t *testing.T) {
	router := setupTestRouter(t)

	// No secret header
	w := httptest.NewRecorder()
	body := `{"version":"0.1.0","op":"Login","content":{"metas":{"token":"any"}}}`
	req, _ := http.NewRequest("POST", "/api/plugin/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Deliberately no X-Plugin-Secret header
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	// Wrong secret header
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/plugin/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Plugin-Secret", "wrong-secret")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ======================== Admin Tests ========================

func TestAdminEndpointsRequireAdminRole(t *testing.T) {
	router, db := setupTestRouterWithDB(t)
	userToken := createActiveUserToken(t, router, db, "normaluser")

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/admin/users"},
		{"POST", "/api/admin/users"},
		{"PUT", "/api/admin/users/1"},
		{"PUT", "/api/admin/users/1/activate"},
		{"DELETE", "/api/admin/users/1"},
		{"GET", "/api/admin/domains"},
		{"GET", "/api/admin/invite-codes"},
		{"POST", "/api/admin/invite-codes"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			var bodyReader *strings.Reader
			if ep.method == "POST" || ep.method == "PUT" {
				bodyReader = strings.NewReader(`{}`)
			} else {
				bodyReader = strings.NewReader("")
			}
			req, _ := http.NewRequest(ep.method, ep.path, bodyReader)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+userToken)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code, "expected 403 for %s %s with normal user", ep.method, ep.path)
		})
	}
}

func TestAdminListUsers(t *testing.T) {
	router, db := setupTestRouterWithDB(t)
	adminToken := createAdminToken(t, router, db)

	// Register a pending user
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", strings.NewReader(`{"username":"listeduser","password":"pass123456"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// List all users as admin
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var users []map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &users))
	assert.GreaterOrEqual(t, len(users), 2) // admin + listeduser
}

func TestAdminActivatePendingUser(t *testing.T) {
	router, db := setupTestRouterWithDB(t)
	adminToken := createAdminToken(t, router, db)

	// Register a pending user
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", strings.NewReader(`{"username":"pendingactivate","password":"pass123456"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var regResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &regResp)
	user := regResp["user"].(map[string]interface{})
	userID := int(user["id"].(float64))

	// Activate the user
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/api/admin/users/%d/activate", userID), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify user can now login
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"username":"pendingactivate","password":"pass123456"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminCreateInviteCode(t *testing.T) {
	router, db := setupTestRouterWithDB(t)
	adminToken := createAdminToken(t, router, db)

	w := httptest.NewRecorder()
	body := `{"max_uses":5,"expires_in_hours":24}`
	req, _ := http.NewRequest("POST", "/api/admin/invite-codes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["code"])
	assert.Equal(t, float64(5), resp["max_uses"])
}

func TestAdminUpdateUserInvalidRole(t *testing.T) {
	router, db := setupTestRouterWithDB(t)
	adminToken := createAdminToken(t, router, db)

	// Register a user to update
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", strings.NewReader(`{"username":"roleuser","password":"pass123456"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var regResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &regResp)
	user := regResp["user"].(map[string]interface{})
	userID := int(user["id"].(float64))

	// Try to set invalid role
	w = httptest.NewRecorder()
	body := `{"role":"superadmin"}`
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/api/admin/users/%d", userID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "invalid role")
}

func TestAdminUpdateUserInvalidStatus(t *testing.T) {
	router, db := setupTestRouterWithDB(t)
	adminToken := createAdminToken(t, router, db)

	// Register a user
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", strings.NewReader(`{"username":"statususer","password":"pass123456"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var regResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &regResp)
	user := regResp["user"].(map[string]interface{})
	userID := int(user["id"].(float64))

	// Try to set invalid status
	w = httptest.NewRecorder()
	body := `{"status":"banned"}`
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/api/admin/users/%d", userID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "invalid status")
}

// ======================== Config Tests ========================

func TestGetConfig(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/config", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, testBaseDomain, resp["base_domain"])
}
