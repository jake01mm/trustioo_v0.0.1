package health

import (
	"github.com/gin-gonic/gin"
)

// Routes 健康检查路由
type Routes struct {
	handler *Handler
}

// NewRoutes 创建新的健康检查路由
func NewRoutes(handler *Handler) *Routes {
	return &Routes{
		handler: handler,
	}
}

// RegisterRoutes 注册健康检查路由
func (r *Routes) RegisterRoutes(router *gin.Engine) {
	health := router.Group("/health")
	{
		// 整体健康检查
		health.GET("", r.handler.OverallHealth)
		health.GET("/", r.handler.OverallHealth)

		// 具体服务健康检查
		health.GET("/database", r.handler.DatabaseHealth)
		health.GET("/redis", r.handler.RedisHealth)
		health.GET("/api/v1", r.handler.APIHealth)

		// Kubernetes风格的健康检查
		health.GET("/readiness", r.handler.Readiness)
		health.GET("/liveness", r.handler.Liveness)
	}
}
