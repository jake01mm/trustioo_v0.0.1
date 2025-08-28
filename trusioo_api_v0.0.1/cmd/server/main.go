package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"trusioo_api_v0.0.1/internal/config"
	"trusioo_api_v0.0.1/internal/infrastructure/database"
	"trusioo_api_v0.0.1/internal/infrastructure/redis"
	"trusioo_api_v0.0.1/internal/infrastructure/router"
	"trusioo_api_v0.0.1/pkg/cryptoutil"

	"trusioo_api_v0.0.1/internal/modules/auth"
	"trusioo_api_v0.0.1/internal/modules/auth/admin"
	"trusioo_api_v0.0.1/internal/modules/auth/user"
	"trusioo_api_v0.0.1/internal/modules/health"
	"trusioo_api_v0.0.1/internal/modules/user_management"
	"trusioo_api_v0.0.1/internal/modules/wallet"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger := setupLogger(cfg)
	logger.Info("Starting Trusioo API server...")

	// 初始化数据库
	db, err := database.New(&cfg.Database, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// 初始化Redis
	redisClient, err := redis.New(&cfg.Redis, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// 初始化路由器
	routerEngine := router.New(cfg, logger)

	// 初始化服务
	setupServices(routerEngine, db, redisClient, cfg, logger)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         cfg.App.Host + ":" + cfg.App.Port,
		Handler:      routerEngine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		logger.WithFields(logrus.Fields{
			"addr": server.Addr,
			"env":  cfg.App.Env,
		}).Info("HTTP server starting")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// 等待中断信号进行优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 创建关闭上下文，30秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	} else {
		logger.Info("Server exited gracefully")
	}
}

// setupLogger 设置日志
func setupLogger(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	if cfg.Log.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// 设置日志输出
	if cfg.Log.Output == "file" {
		// 这里可以添加文件输出逻辑
		// 目前简化为标准输出
		logger.SetOutput(os.Stdout)
	} else {
		logger.SetOutput(os.Stdout)
	}

	return logger
}

// setupServices 设置服务和路由
func setupServices(routerEngine *router.Router, db *database.Database, redisClient *redis.Client, cfg *config.Config, logger *logrus.Logger) {
	// 初始化JWT管理器（带数据库支持）
	jwtManager := auth.NewJWTManager(&cfg.JWT, db, logger)

	// 初始化密码加密器
	passwordEncryptor := cryptoutil.NewPasswordEncryptor(cfg.PasswordEncrypt.Key, cfg.PasswordEncrypt.Method)

	// 初始化认证中间件
	authMiddle := auth.NewAuthMiddleware(jwtManager, logger)

	// 设置健康检查模块
	setupHealthModule(routerEngine, db, redisClient, logger)

	// 设置认证模块
	setupAuthModules(routerEngine, db, jwtManager, authMiddle, passwordEncryptor, logger)

	// 设置用户管理模块
	setupUserManagementModule(routerEngine, db, jwtManager, authMiddle, passwordEncryptor, logger)

	// 设置钱包模块
	setupWalletModule(routerEngine, db, jwtManager, authMiddle, passwordEncryptor, logger)
}

// setupHealthModule 设置健康检查模块
func setupHealthModule(routerEngine *router.Router, db *database.Database, redisClient *redis.Client, logger *logrus.Logger) {
	// 初始化健康检查服务和处理器
	_ = health.NewService(db, redisClient, logger)
	healthHandler := health.NewHandler(db, redisClient, logger)
	healthRoutes := health.NewRoutes(healthHandler)

	// 注册健康检查路由
	healthRoutes.RegisterRoutes(routerEngine.Engine)

	logger.Info("Health check module initialized")
}

// setupAuthModules 设置认证模块
func setupAuthModules(routerEngine *router.Router, db *database.Database, jwtManager *auth.JWTManager, authMiddle *auth.AuthMiddleware, passwordEncryptor *cryptoutil.PasswordEncryptor, logger *logrus.Logger) {
	// 获取API v1路由分组
	v1Group := routerEngine.GetV1Group()
	authGroup := v1Group.Group("/auth")

	// 设置管理员认证模块
	setupAdminAuth(authGroup, db, jwtManager, authMiddle, passwordEncryptor, logger)

	// 设置用户认证模块
	setupUserAuth(authGroup, db, jwtManager, authMiddle, passwordEncryptor, logger)

	logger.Info("Auth modules initialized")
}

// setupAdminAuth 设置管理员认证模块
func setupAdminAuth(authGroup *gin.RouterGroup, db *database.Database, jwtManager *auth.JWTManager, authMiddle *auth.AuthMiddleware, passwordEncryptor *cryptoutil.PasswordEncryptor, logger *logrus.Logger) {
	adminRepo := admin.NewRepository(db, logger)
	verifyRepo := user.NewVerificationRepository(db, logger)
	adminService := admin.NewService(adminRepo, verifyRepo, passwordEncryptor, logger)
	adminHandler := admin.NewHandler(adminService, jwtManager, logger)
	adminRoutes := admin.NewRoutes(adminHandler, authMiddle)

	adminRoutes.RegisterRoutes(authGroup)
	logger.Info("Admin auth module initialized")
}

// setupUserAuth 设置用户认证模块
func setupUserAuth(authGroup *gin.RouterGroup, db *database.Database, jwtManager *auth.JWTManager, authMiddle *auth.AuthMiddleware, passwordEncryptor *cryptoutil.PasswordEncryptor, logger *logrus.Logger) {
	userRepo := user.NewRepository(db, logger)
	verifyRepo := user.NewVerificationRepository(db, logger)
	userService := user.NewService(userRepo, verifyRepo, passwordEncryptor, logger)
	userHandler := user.NewHandler(userService, jwtManager, logger)
	userRoutes := user.NewRoutes(userHandler, authMiddle)

	userRoutes.RegisterRoutes(authGroup)
	logger.Info("User auth module initialized")
}

// setupUserManagementModule 设置用户管理模块
func setupUserManagementModule(routerEngine *router.Router, db *database.Database, jwtManager *auth.JWTManager, authMiddle *auth.AuthMiddleware, passwordEncryptor *cryptoutil.PasswordEncryptor, logger *logrus.Logger) {
	// 获取API v1路由分组
	v1Group := routerEngine.GetV1Group()

	// 初始化用户管理模块的依赖
	userRepo := user.NewRepository(db, logger) // 复用用户仓储
	userMgmtRepo := user_management.NewRepository(db, logger)
	userMgmtService := user_management.NewService(userMgmtRepo, userRepo, passwordEncryptor, logger)
	userMgmtHandler := user_management.NewHandler(userMgmtService, logger)
	userMgmtRoutes := user_management.NewRoutes(userMgmtHandler, authMiddle)

	// 注册路由
	userMgmtRoutes.RegisterRoutes(v1Group)
	logger.Info("User management module initialized")
}

// setupWalletModule 设置钱包模块
func setupWalletModule(routerEngine *router.Router, db *database.Database, jwtManager *auth.JWTManager, authMiddle *auth.AuthMiddleware, passwordEncryptor *cryptoutil.PasswordEncryptor, logger *logrus.Logger) {
	// 获取API v1路由分组
	v1Group := routerEngine.GetV1Group()

	// 初始化钱包模块组件
	walletRepo := wallet.NewRepository(db, logger)
	walletService := wallet.NewService(walletRepo, passwordEncryptor, logger)
	walletHandler := wallet.NewHandler(walletService, logger)
	walletRoutes := wallet.NewRoutes(walletHandler, authMiddle)

	// 注册钱包路由
	walletRoutes.RegisterRoutes(v1Group)

	logger.Info("Wallet module initialized")
}
