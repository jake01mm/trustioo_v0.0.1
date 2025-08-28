package router

import (
	"net/http"
	"time"

	"trusioo_api_v0.0.1/internal/config"
	"trusioo_api_v0.0.1/pkg/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Router 路由器结构
type Router struct {
	*gin.Engine
	config *config.Config
	logger *logrus.Logger
}

// New 创建新的路由器
func New(cfg *config.Config, logger *logrus.Logger) *Router {
	// 根据环境设置Gin模式
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()

	router := &Router{
		Engine: engine,
		config: cfg,
		logger: logger,
	}

	// 设置基础中间件
	router.setupMiddleware()

	// 设置基础路由
	router.setupBaseRoutes()

	return router
}

// setupMiddleware 设置中间件
func (r *Router) setupMiddleware() {
	// 自定义日志中间件
	r.Use(middleware.Logger(r.logger))

	// 恢复中间件
	r.Use(middleware.Recovery(r.logger))

	// CORS中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     r.config.Security.CORSAllowedOrigins,
		AllowMethods:     r.config.Security.CORSAllowedMethods,
		AllowHeaders:     r.config.Security.CORSAllowedHeaders,
		AllowCredentials: r.config.Security.CORSAllowCredentials,
		MaxAge:           12 * time.Hour,
	}))

	// 请求ID中间件
	r.Use(middleware.RequestID())

	// 超时中间件
	r.Use(middleware.Timeout(30*time.Second, r.logger))
}

// setupBaseRoutes 设置基础路由
func (r *Router) setupBaseRoutes() {
	// 健康检查路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 版本信息路由
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"app":     r.config.App.Name,
			"version": r.config.App.Version,
			"env":     r.config.App.Env,
		})
	})

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Route not found",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
			"message": "The requested resource was not found on this server",
		})
	})

	// 405处理
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error":   "Method not allowed",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
			"message": "The requested method is not allowed for this resource",
		})
	})
}

// RegisterHealthRoutes 注册健康检查路由
func (r *Router) RegisterHealthRoutes(healthHandler interface{}) {
	health := r.Group("/health")
	{
		// 这里将在健康检查模块创建后实现
		health.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}
}

// RegisterAPIRoutes 注册API路由
func (r *Router) RegisterAPIRoutes() {
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			// 这里将在各个模块创建后实现路由注册
			v1.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "API v1",
					"version": "1.0.0",
					"status":  "active",
				})
			})
		}
	}
}

// GetAPIGroup 获取API分组
func (r *Router) GetAPIGroup() *gin.RouterGroup {
	return r.Group("/api")
}

// GetV1Group 获取v1 API分组
func (r *Router) GetV1Group() *gin.RouterGroup {
	return r.Group("/api/v1")
}

// StartServer 启动服务器
func (r *Router) StartServer() error {
	addr := r.config.App.Host + ":" + r.config.App.Port

	r.logger.WithFields(logrus.Fields{
		"addr": addr,
		"env":  r.config.App.Env,
	}).Info("Starting HTTP server")

	server := &http.Server{
		Addr:           addr,
		Handler:        r.Engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	return server.ListenAndServe()
}

// RouteInfo 路由信息结构
type RouteInfo struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// GetRoutes 获取所有注册的路由
func (r *Router) GetRoutes() []RouteInfo {
	routes := r.Routes()
	routeInfos := make([]RouteInfo, 0, len(routes))

	for _, route := range routes {
		routeInfos = append(routeInfos, RouteInfo{
			Method: route.Method,
			Path:   route.Path,
		})
	}

	return routeInfos
}

// RegisterDocsRoutes 注册文档路由 (如果启用)
func (r *Router) RegisterDocsRoutes() {
	if r.config.App.Debug {
		docs := r.Group("/docs")
		{
			docs.GET("", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "API Documentation",
					"routes":  r.GetRoutes(),
				})
			})
		}
	}
}
