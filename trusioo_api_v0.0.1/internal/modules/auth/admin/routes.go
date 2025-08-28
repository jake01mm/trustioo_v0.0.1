package admin

import (
	"trusioo_api_v0.0.1/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

// Routes 管理员认证路由
type Routes struct {
	handler    *Handler
	authMiddle *auth.AuthMiddleware
}

// NewRoutes 创建新的管理员认证路由
func NewRoutes(handler *Handler, authMiddle *auth.AuthMiddleware) *Routes {
	return &Routes{
		handler:    handler,
		authMiddle: authMiddle,
	}
}

// RegisterRoutes 注册管理员认证路由
func (r *Routes) RegisterRoutes(router *gin.RouterGroup) {
	admin := router.Group("/admin")
	{
		// 公开路由（不需要认证）
		admin.POST("/login", r.handler.Login)
		admin.POST("/verify-login", r.handler.VerifyLogin)
		admin.POST("/forgot-password", r.handler.ForgotPassword)
		admin.POST("/reset-password", r.handler.ResetPassword)

		// 需要认证的路由
		authenticated := admin.Group("")
		authenticated.Use(r.authMiddle.RequireAuth())
		authenticated.Use(r.authMiddle.RequireUserType("admin"))
		{
			// 令牌相关
			authenticated.POST("/refresh", r.handler.RefreshToken)
			authenticated.POST("/logout", r.handler.Logout)

			// 个人资料
			authenticated.GET("/profile", r.handler.GetProfile)
			authenticated.PUT("/password", r.handler.ChangePassword)
		}
	}
}
