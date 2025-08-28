package user_management

import (
	"trusioo_api_v0.0.1/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

// Routes 用户管理路由
type Routes struct {
	handler    *Handler
	authMiddle *auth.AuthMiddleware
}

// NewRoutes 创建新的用户管理路由
func NewRoutes(handler *Handler, authMiddle *auth.AuthMiddleware) *Routes {
	return &Routes{
		handler:    handler,
		authMiddle: authMiddle,
	}
}

// RegisterRoutes 注册用户管理路由
func (r *Routes) RegisterRoutes(router *gin.RouterGroup) {
	// 用户管理路由组 - 需要管理员认证
	userMgmt := router.Group("/admin/user-management")
	userMgmt.Use(r.authMiddle.RequireAuth())
	userMgmt.Use(r.authMiddle.RequireUserType("admin"))
	{
		// === 用户查询接口 ===

		// 获取用户列表（支持分页、排序、筛选、搜索）
		userMgmt.GET("/users", r.handler.GetUsers)

		// 获取用户详细信息
		userMgmt.GET("/users/:user_id", r.handler.GetUserDetail)

		// 获取用户活动信息
		userMgmt.GET("/users/:user_id/activity", r.handler.GetUserActivity)

		// === 统计接口 ===

		// 获取用户统计信息
		userMgmt.GET("/statistics", r.handler.GetStatistics)

		// === 用户管理操作接口 ===

		// 更新用户状态
		userMgmt.PUT("/users/:user_id/status", r.handler.UpdateUserStatus)

		// 暂停用户
		userMgmt.POST("/users/:user_id/suspend", r.handler.SuspendUser)

		// 重新激活用户
		userMgmt.POST("/users/:user_id/reactivate", r.handler.ReactivateUser)

		// 重置用户密码
		userMgmt.POST("/users/:user_id/reset-password", r.handler.ResetUserPassword)

		// 强制用户登出
		userMgmt.POST("/users/:user_id/force-logout", r.handler.ForceLogoutUser)

		// 验证用户邮箱
		userMgmt.POST("/users/:user_id/verify-email", r.handler.VerifyUserEmail)

		// === 未来扩展接口占位 ===
		// 注意：这些接口在第一阶段不实现，仅作为路由占位

		// 用户会话管理
		// userMgmt.GET("/users/:user_id/sessions", r.handler.GetUserSessions)
		// userMgmt.DELETE("/users/:user_id/sessions/:session_id", r.handler.DeleteUserSession)

		// 更新用户邮箱
		// userMgmt.PUT("/users/:user_id/email", r.handler.UpdateUserEmail)

		// 软删除用户
		// userMgmt.DELETE("/users/:user_id", r.handler.DeleteUser)

		// 恢复已删除用户
		// userMgmt.POST("/users/:user_id/restore", r.handler.RestoreUser)

		// 获取操作日志
		// userMgmt.GET("/users/:user_id/logs", r.handler.GetUserManagementLogs)
		// userMgmt.GET("/logs", r.handler.GetAllManagementLogs)

		// 数据导出
		// userMgmt.GET("/export/users", r.handler.ExportUsers)
		// userMgmt.GET("/downloads/:file_name", r.handler.DownloadExportFile)

		// 批量操作
		// userMgmt.POST("/users/batch/status", r.handler.BatchUpdateUserStatus)
		// userMgmt.POST("/users/batch/suspend", r.handler.BatchSuspendUsers)
		// userMgmt.POST("/users/batch/reactivate", r.handler.BatchReactivateUsers)
		// userMgmt.POST("/users/batch/force-logout", r.handler.BatchForceLogoutUsers)

		// 高级统计
		// userMgmt.GET("/statistics/registration", r.handler.GetRegistrationStatistics)
		// userMgmt.GET("/statistics/verification", r.handler.GetVerificationStatistics)
		// userMgmt.GET("/statistics/activity", r.handler.GetActivityStatistics)
		// userMgmt.GET("/statistics/export", r.handler.ExportStatistics)
	}
}
