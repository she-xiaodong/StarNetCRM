package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/starnet/crm/internal/config"
	"github.com/starnet/crm/internal/handler"
	"github.com/starnet/crm/internal/middleware"
	"github.com/starnet/crm/internal/model"
	repomysql "github.com/starnet/crm/internal/repository/mysql"
	reponeo4j "github.com/starnet/crm/internal/repository/neo4j"
	"github.com/starnet/crm/pkg/cache"
	"github.com/starnet/crm/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// ─── 1. 加载配置 ───
	cfgPath := "config.yaml"
	if v := os.Getenv("CONFIG_PATH"); v != "" {
		cfgPath = v
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ─── 2. 初始化日志 ───
	if err := logger.Init(cfg.Server.Mode); err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer logger.Sync()

	// ─── 3. 初始化MySQL ───
	if err := repomysql.Init(cfg.MySQL); err != nil {
		logger.Log.Fatal("Failed to init MySQL", zap.Error(err))
	}
	defer repomysql.Close()

	if err := repomysql.AutoMigrate(); err != nil {
		logger.Log.Fatal("Failed to auto migrate", zap.Error(err))
	}
	logger.Log.Info("MySQL connected and migrated")

	// ─── 4. 初始化Neo4j ───
	if err := reponeo4j.Init(cfg.Neo4j); err != nil {
		logger.Log.Warn("Neo4j connection failed (graph features will be unavailable)", zap.Error(err))
	} else {
		defer reponeo4j.Close()
		logger.Log.Info("Neo4j connected")
	}

	// ─── 5. 初始化Redis ───
	if err := cache.Init(cfg.Redis); err != nil {
		logger.Log.Warn("Redis connection failed (cache will be unavailable)", zap.Error(err))
	} else {
		defer cache.Close()
		logger.Log.Info("Redis connected")
	}

	// ─── 6. 设置Gin模式 ───
	gin.SetMode(cfg.Server.Mode)

	// ─── 7. 创建路由 ───
	r := gin.New()

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Logger())

	// 日志中间件
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Log.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	})

	// ─── 8. 注册路由 ───
	setupRoutes(r)

	// ─── 9. 启动服务器 ───
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// 优雅启动
	go func() {
		logger.Log.Info("🚀 StarNet CRM Server starting",
			zap.String("addr", addr),
			zap.String("mode", cfg.Server.Mode),
			zap.String("deploy", string(cfg.Deploy.Mode)),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed", zap.Error(err))
		}
	}()

	// ─── 10. 优雅退出 ───
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server exited")
}

// setupRoutes 路由注册
func setupRoutes(r *gin.Engine) {
	authHandler := handler.NewAuthHandler()
	contactHandler := handler.NewContactHandler()
	graphHandler := handler.NewGraphHandler()
	tagHandler := handler.NewTagHandler()
	adminHandler := handler.NewAdminHandler()
	dashboardHandler := handler.NewDashboardHandler()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "StarNet CRM",
			"version": "1.0.0",
		})
	})

	api := r.Group("/api/v1")

	// ─── 公开接口 ───
	auth := api.Group("/auth")
	{
		// 独立模式
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		// 企微OAuth
		auth.POST("/wecom", authHandler.WecomAuth)
		auth.GET("/wecom/verify", authHandler.WecomVerify)
		auth.GET("/wecom/config", authHandler.WecomOAuthConfig)
		// 部署模式查询
		auth.GET("/mode", func(c *gin.Context) {
			c.JSON(200, gin.H{"deploy_mode": config.AppConfig.Deploy.Mode})
		})
	}

	// ─── 需要鉴权的接口 ───
	authorized := api.Group("")
	authorized.Use(middleware.JWTAuth())
	{
		// 用户
		authorized.GET("/user/me", authHandler.GetCurrentUser)
		authorized.POST("/auth/logout", authHandler.Logout)

		// 联系人
		contacts := authorized.Group("/contacts")
		{
			contacts.GET("", contactHandler.List)
			contacts.POST("", contactHandler.Create)
			contacts.GET("/:id", contactHandler.Get)
			contacts.PUT("/:id", contactHandler.Update)
			contacts.DELETE("/:id", contactHandler.Delete)
		}

		// 图谱
		graph := authorized.Group("/graph")
		{
			graph.GET("/first-degree", graphHandler.GetFirstDegree)
			graph.POST("/search-path", graphHandler.SearchPath)
			graph.POST("/relations", graphHandler.CreateRelation)
			graph.GET("/super-connectors", graphHandler.GetSuperConnectors)
		}

		// 标签管理
		tags := authorized.Group("/tags")
		{
			tags.GET("", tagHandler.List)
			tags.POST("", tagHandler.Create)
			tags.PUT("/:id", tagHandler.Update)
			tags.DELETE("/:id", tagHandler.Delete)
		}

		// 统计
		stats := authorized.Group("/stats")
		{
			stats.GET("/dashboard", dashboardHandler.Stats)
			stats.GET("/analytics", dashboardHandler.Analytics)
		}

		// 导入导出
		io := authorized.Group("/io")
		{
			io.POST("/import")
			io.GET("/export")
		}
	}

	// ─── 管理员接口 ───
	admin := api.Group("/admin")
	admin.Use(middleware.JWTAuth(), middleware.RequireRole(string(model.RoleAdmin)))
	{
		admin.GET("/stats", adminHandler.Stats)
		admin.GET("/tenants", adminHandler.ListTenants)
		admin.POST("/tenants", adminHandler.CreateTenant)
		admin.GET("/tenants/:id", adminHandler.GetTenant)
		admin.DELETE("/tenants/:id", adminHandler.DeleteTenant)
	}
}
