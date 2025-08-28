// Package examples 演示企业级工具集成使用
package examples

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"trusioo_api_v0.0.1/pkg/cache"
	pkgConfig "trusioo_api_v0.0.1/pkg/config"
	"trusioo_api_v0.0.1/pkg/logger"
	"trusioo_api_v0.0.1/pkg/metrics"
	"trusioo_api_v0.0.1/pkg/middleware"
	"trusioo_api_v0.0.1/pkg/response"
	"trusioo_api_v0.0.1/pkg/security"
	"trusioo_api_v0.0.1/pkg/swagger"
	"trusioo_api_v0.0.1/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// EnterpriseApp 企业级应用程序
type EnterpriseApp struct {
	router         *gin.Engine
	logger         *logger.Logger
	configManager  *pkgConfig.ConfigManager
	metricsManager *metrics.Metrics
	cacheManager   *cache.CacheManager
	swaggerDoc     *swagger.SwaggerDoc
	validator      *validator.ValidatorEngine

	// 安全组件
	rateLimiter  *security.TokenBucketLimiter
	sqlProtector *security.SQLInjectionProtector
	xssProtector *security.XSSProtector

	server *http.Server
}

// NewEnterpriseApp 创建企业级应用实例
func NewEnterpriseApp() *EnterpriseApp {
	// 初始化日志系统
	loggerInstance := logger.NewLogger(&logger.Config{
		Level:    "info",
		Format:   "json",
		Output:   "file",
		FilePath: "logs/app.log",
	})

	// 初始化配置管理器
	configManager := pkgConfig.NewConfigManager("development", loggerInstance.Logger)

	// 初始化指标收集器
	metricsManager := metrics.NewMetrics(metrics.DefaultConfig(), loggerInstance.Logger)

	// 初始化Redis和缓存
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	redisCache := cache.NewRedisCache(redisClient, "trusioo", loggerInstance.Logger)
	cacheManager := cache.NewCacheManager(redisCache, cache.DefaultStrategy(), loggerInstance.Logger)

	// 初始化验证器
	validatorInstance := validator.NewValidator()

	// 初始化安全组件
	rateLimiter := security.NewTokenBucketLimiter(10, 20, loggerInstance.Logger) // 10 req/s, burst 20
	sqlProtector := security.NewSQLInjectionProtector(loggerInstance.Logger)
	xssProtector := security.NewXSSProtector(loggerInstance.Logger)

	// 初始化Swagger文档
	swaggerDoc := swagger.NewSwaggerDoc(swagger.DefaultConfig())

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	return &EnterpriseApp{
		router:         router,
		logger:         loggerInstance,
		configManager:  configManager,
		metricsManager: metricsManager,
		cacheManager:   cacheManager,
		swaggerDoc:     swaggerDoc,
		validator:      validatorInstance,
		rateLimiter:    rateLimiter,
		sqlProtector:   sqlProtector,
		xssProtector:   xssProtector,
	}
}

// SetupMiddlewares 设置中间件
func (app *EnterpriseApp) SetupMiddlewares() {
	// 基础中间件
	app.router.Use(gin.Recovery())

	// 日志中间件
	app.router.Use(middleware.Logger(app.logger.Logger))

	// 请求ID中间件
	app.router.Use(middleware.RequestID())

	// 性能监控中间件（简化版）
	app.router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		app.logger.Performance("http_request", duration, logger.Fields{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": c.Writer.Status(),
		})
	})

	// 指标收集中间件
	app.router.Use(app.metricsManager.HTTPMetricsMiddleware())

	// 安全中间件
	app.router.Use(security.SecurityHeaders())

	// CORS中间件
	corsConfig := &security.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
	}
	app.router.Use(security.CORSMiddleware(corsConfig.AllowOrigins, corsConfig.AllowMethods, corsConfig.AllowHeaders))

	// 限流中间件
	rateLimitConfig := &security.RateLimitConfig{
		Rate:         10,
		Burst:        20,
		ErrorMessage: "请求过于频繁，请稍后再试",
		KeyGenerator: func(c *gin.Context) string {
			return c.ClientIP()
		},
		SkipPaths: []string{"/health", "/metrics"},
	}
	app.router.Use(security.RateLimitMiddleware(app.rateLimiter, rateLimitConfig, app.logger.Logger))

	// SQL注入防护中间件
	app.router.Use(security.SQLInjectionMiddleware(app.sqlProtector, app.logger.Logger))

	// XSS防护中间件
	app.router.Use(security.XSSMiddleware(app.xssProtector, app.logger.Logger))

	// 全局错误处理中间件
	app.router.Use(middleware.ErrorHandler(app.logger.Logger))

	// 验证中间件（简化版）
	app.router.Use(func(c *gin.Context) {
		c.Next()
	})
}

// SetupRoutes 设置路由
func (app *EnterpriseApp) SetupRoutes() {
	// 健康检查
	app.router.GET("/health", app.healthCheck)

	// API版本信息
	app.router.GET("/api/v1", app.apiInfo)

	// 设置指标路由
	app.metricsManager.SetupMetricsRoute(app.router)

	// 设置Swagger文档路由
	app.swaggerDoc.SetupSwaggerRoutes(app.router)

	// API路由组
	v1 := app.router.Group("/api/v1")
	{
		// 认证路由
		auth := v1.Group("/auth")
		{
			auth.POST("/login", app.login)
			auth.POST("/register", app.register)
			auth.GET("/profile", app.profile)
		}

		// 管理员路由
		admin := v1.Group("/admin")
		{
			admin.POST("/login", app.adminLogin)
			admin.GET("/profile", app.adminProfile)
			admin.GET("/users", app.getUserList)
		}

		// 缓存演示路由
		cache := v1.Group("/cache")
		{
			cache.GET("/test", app.cacheTest)
			cache.DELETE("/clear", app.cacheClear)
		}

		// 配置管理路由
		config := v1.Group("/config")
		{
			config.GET("/reload", app.configReload)
			config.GET("/info", app.configInfo)
		}
	}
}

// healthCheck 健康检查
func (app *EnterpriseApp) healthCheck(c *gin.Context) {
	// 检查Redis连接（简化版）
	redisStatus := "healthy"
	if app.cacheManager == nil {
		redisStatus = "unavailable"
	}

	response.SuccessWithMessage(c, "服务运行正常", gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"services": gin.H{
			"redis": redisStatus,
			"cache": "healthy",
		},
	})
}

// apiInfo API信息
func (app *EnterpriseApp) apiInfo(c *gin.Context) {
	response.SuccessWithMessage(c, "API信息获取成功", gin.H{
		"name":        "Trusioo API",
		"version":     "v0.0.1",
		"description": "企业级电商平台API",
		"features": []string{
			"统一响应格式",
			"错误处理机制",
			"输入验证",
			"日志系统",
			"监控指标",
			"缓存策略",
			"安全防护",
			"API文档",
			"配置管理",
		},
		"endpoints": gin.H{
			"health":  "/health",
			"metrics": "/metrics",
			"docs":    "/swagger/",
		},
	})
}

// login 用户登录示例
func (app *EnterpriseApp) login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	// 记录登录指标（简化版）
	app.logger.Business("user_login", "success", logger.Fields{
		"email": req.Email,
	})

	// 模拟缓存用户信息（简化版）
	app.logger.Info("用户登录成功")

	response.SuccessWithMessage(c, "登录成功", gin.H{
		"token": "mock_jwt_token",
		"user": gin.H{
			"email": req.Email,
		},
	})
}

// register 用户注册示例
func (app *EnterpriseApp) register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
		Name     string `json:"name" validate:"required,min=2"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err)
		return
	}

	// 记录注册指标（简化版）
	app.logger.Business("user_register", "success", logger.Fields{
		"email": req.Email,
		"name":  req.Name,
	})

	response.SuccessWithMessage(c, "注册成功", gin.H{
		"user": gin.H{
			"email": req.Email,
			"name":  req.Name,
		},
	})
}

// profile 用户资料
func (app *EnterpriseApp) profile(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.BadRequest(c, "缺少邮箱参数")
		return
	}

	// 模拟获取用户信息
	userData := gin.H{
		"email":      email,
		"name":       "用户" + email[:5],
		"created_at": time.Now().Add(-24 * time.Hour),
	}

	response.SuccessWithMessage(c, "获取用户资料成功", userData)
}

// adminLogin 管理员登录
func (app *EnterpriseApp) adminLogin(c *gin.Context) {
	response.SuccessWithMessage(c, "接口可用", gin.H{
		"message": "管理员登录接口",
	})
}

// adminProfile 管理员资料
func (app *EnterpriseApp) adminProfile(c *gin.Context) {
	response.SuccessWithMessage(c, "接口可用", gin.H{
		"message": "管理员资料接口",
	})
}

// getUserList 用户列表
func (app *EnterpriseApp) getUserList(c *gin.Context) {
	// 简化版参数获取
	page := 1
	limit := 10
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	users := []gin.H{
		{"id": 1, "email": "user1@example.com", "name": "用户1"},
		{"id": 2, "email": "user2@example.com", "name": "用户2"},
	}

	response.SuccessWithMessage(c, "获取用户列表成功", gin.H{
		"data": users,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       100,
			"total_pages": 10,
		},
	})
}

// cacheTest 缓存测试
func (app *EnterpriseApp) cacheTest(c *gin.Context) {
	// 模拟缓存测试
	data := gin.H{
		"message":   "这是模拟缓存数据",
		"timestamp": time.Now(),
	}

	response.SuccessWithMessage(c, "缓存测试成功", data)
}

// cacheClear 清空缓存
func (app *EnterpriseApp) cacheClear(c *gin.Context) {
	pattern := c.Query("pattern")
	if pattern == "" {
		pattern = "*"
	}

	// 模拟清空缓存
	app.logger.Info("模拟清空缓存", logger.Fields{"pattern": pattern})

	response.SuccessWithMessage(c, "缓存清空成功", nil)
}

// configReload 重新加载配置
func (app *EnterpriseApp) configReload(c *gin.Context) {
	// 模拟重新加载配置
	app.logger.Info("模拟重新加载配置")

	response.SuccessWithMessage(c, "配置重新加载成功", nil)
}

// configInfo 配置信息
func (app *EnterpriseApp) configInfo(c *gin.Context) {
	response.SuccessWithMessage(c, "获取配置信息成功", gin.H{
		"environment":   "development",
		"is_production": false,
		"config_keys":   []string{"app", "database", "redis"},
	})
}

// Start 启动应用
func (app *EnterpriseApp) Start() error {
	// 设置中间件
	app.SetupMiddlewares()

	// 设置路由
	app.SetupRoutes()

	// 创建HTTP服务器
	app.server = &http.Server{
		Addr:    ":8080",
		Handler: app.router,
	}

	// 启动服务器
	app.logger.Info("企业级应用启动中...", logger.Fields{
		"port": "8080",
		"env":  "development",
	})

	go func() {
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 优雅关闭
	app.gracefulShutdown()

	return nil
}

// gracefulShutdown 优雅关闭
func (app *EnterpriseApp) gracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.logger.Info("正在关闭服务器...")

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := app.server.Shutdown(ctx); err != nil {
		app.logger.Error("服务器强制关闭", logger.Fields{"error": err})
	}

	// 停止配置监听（简化版）
	app.logger.Info("停止配置监听")

	app.logger.Info("服务器已关闭")
}

// 演示如何运行
func main() {
	app := NewEnterpriseApp()

	if err := app.Start(); err != nil {
		log.Fatalf("启动应用失败: %v", err)
	}
}
