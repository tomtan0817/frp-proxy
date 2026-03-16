package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"frp-proxy/internal/config"
	"frp-proxy/internal/database"
	"frp-proxy/internal/handler"
	"frp-proxy/internal/middleware"
	"frp-proxy/internal/service"
	frpweb "frp-proxy/web"

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
	domainH := handler.NewDomainHandler(domainSvc, cfg.Domain.BaseDomain)
	adminUserH := handler.NewAdminUserHandler(userSvc)
	adminDomainH := handler.NewAdminDomainHandler(domainSvc)
	adminInviteH := handler.NewAdminInviteHandler(inviteSvc)
	pluginH := handler.NewPluginHandler(domainSvc, cfg.Plugin.Secret)

	r := gin.Default()

	// Public routes
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
		}

		api.GET("/config", domainH.GetConfig)

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

	// Serve embedded frontend
	distFS, err := fs.Sub(frpweb.DistFS, "dist")
	if err != nil {
		log.Fatalf("failed to get embedded frontend: %v", err)
	}

	fileServer := http.FileServer(http.FS(distFS))
	r.NoRoute(func(c *gin.Context) {
		// Try static file first
		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		_, err := fs.Stat(distFS, path)
		if err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		// SPA fallback: serve index.html
		indexFile, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			c.String(http.StatusNotFound, "not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexFile)
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
