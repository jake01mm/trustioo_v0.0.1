package user

import (
	"trusioo_api_v0.0.1/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

// Routes 用户认证路由
type Routes struct {
	handler    *Handler
	authMiddle *auth.AuthMiddleware
}

// NewRoutes 创建新的用户认证路由
func NewRoutes(handler *Handler, authMiddle *auth.AuthMiddleware) *Routes {
	return &Routes{
		handler:    handler,
		authMiddle: authMiddle,
	}
}

// RegisterRoutes 注册用户认证路由
func (r *Routes) RegisterRoutes(router *gin.RouterGroup) {
	user := router.Group("/user")
	{
		// 公开路由（不需要认证）
		user.POST("/register", r.handler.Register)              // 简化注册：仅需email+password
		user.POST("/login", r.handler.Login)                    // 发送登录验证码
		user.POST("/verify-login", r.handler.VerifyLogin)       // 验证登录验证码并获取token
		user.POST("/forgot-password", r.handler.ForgotPassword) // 忘记密码
		user.POST("/reset-password", r.handler.ResetPassword)   // 重置密码

		// 需要认证的路由
		authenticated := user.Group("")
		authenticated.Use(r.authMiddle.RequireAuth())
		authenticated.Use(r.authMiddle.RequireUserType("user"))
		{
			authenticated.GET("/profile", r.handler.GetProfile)
			authenticated.POST("/logout", r.handler.Logout)
		}
	}
}
