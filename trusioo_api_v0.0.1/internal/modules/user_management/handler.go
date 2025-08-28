package user_management

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler 用户管理处理器
type Handler struct {
	service *Service
	logger  *logrus.Logger
}

// NewHandler 创建新的用户管理处理器
func NewHandler(service *Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// === 用户查询接口 ===

// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取用户列表，支持分页、排序、筛选和搜索
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param sort_by query string false "排序字段" Enums(created_at,updated_at,email,name,last_login_at,login_count)
// @Param sort_dir query string false "排序方向" Enums(asc,desc)
// @Param email query string false "邮箱筛选"
// @Param name query string false "姓名筛选"
// @Param status query string false "状态筛选" Enums(active,inactive,suspended)
// @Param email_verified query bool false "邮箱验证状态筛选"
// @Param search query string false "搜索关键词"
// @Security ApiKeyAuth
// @Success 200 {object} UserListResponse
// @Failure 400 {object} object
// @Failure 401 {object} object
// @Failure 500 {object} object
// @Router /api/v1/admin/user-management/users [get]
func (h *Handler) GetUsers(c *gin.Context) {
	var req GetUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid get users request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	response, err := h.service.GetUsers(ctx, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get users")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to retrieve users",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUserDetail 获取用户详细信息
// @Summary 获取用户详细信息
// @Description 获取指定用户的详细信息，包括活动统计
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Security ApiKeyAuth
// @Success 200 {object} UserDetailResponse
// @Failure 400 {object} object
// @Failure 401 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/admin/user-management/users/{user_id} [get]
func (h *Handler) GetUserDetail(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "User ID is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	response, err := h.service.GetUserDetail(ctx, userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user detail")

		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"message": "The specified user does not exist",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to retrieve user details",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUserActivity 获取用户活动信息
// @Summary 获取用户活动信息
// @Description 获取用户的活动信息，包括登录记录和会话信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Security ApiKeyAuth
// @Success 200 {object} UserActivityResponse
// @Failure 400 {object} object
// @Failure 401 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/admin/user-management/users/{user_id}/activity [get]
func (h *Handler) GetUserActivity(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "User ID is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	response, err := h.service.GetUserActivity(ctx, userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user activity")

		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"message": "The specified user does not exist",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to retrieve user activity",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// === 统计接口 ===

// GetStatistics 获取用户统计信息
// @Summary 获取用户统计信息
// @Description 获取用户的统计信息，包括总数、新增、验证状态等
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} StatisticsResponse
// @Failure 401 {object} object
// @Failure 500 {object} object
// @Router /api/v1/admin/user-management/statistics [get]
func (h *Handler) GetStatistics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	response, err := h.service.GetStatistics(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get statistics")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to retrieve statistics",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// === 用户管理操作接口 ===

// UpdateUserStatus 更新用户状态
// @Summary 更新用户状态
// @Description 更新用户的状态（激活、停用、暂停）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param request body UpdateUserStatusRequest true "更新用户状态请求"
// @Security ApiKeyAuth
// @Success 200 {object} OperationResponse
// @Failure 400 {object} object
// @Failure 401 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/admin/user-management/users/{user_id}/status [put]
func (h *Handler) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "User ID is required",
		})
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid update user status request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 获取管理员信息
	adminInfo := h.getAdminInfoFromContext(c)
	ipAddress := c.ClientIP()

	response, err := h.service.UpdateUserStatus(ctx, userID, adminInfo.ID, adminInfo.Email, ipAddress, &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":  userID,
			"admin_id": adminInfo.ID,
			"status":   req.Status,
		}).Error("Failed to update user status")

		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"message": "The specified user does not exist",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to update user status",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SuspendUser 暂停用户
// @Summary 暂停用户
// @Description 暂停用户账户并强制登出
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param request body SuspendUserRequest true "暂停用户请求"
// @Security ApiKeyAuth
// @Success 200 {object} OperationResponse
// @Failure 400 {object} object
// @Failure 401 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/admin/user-management/users/{user_id}/suspend [post]
func (h *Handler) SuspendUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "User ID is required",
		})
		return
	}

	var req SuspendUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid suspend user request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 获取管理员信息
	adminInfo := h.getAdminInfoFromContext(c)
	ipAddress := c.ClientIP()

	response, err := h.service.SuspendUser(ctx, userID, adminInfo.ID, adminInfo.Email, ipAddress, &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":  userID,
			"admin_id": adminInfo.ID,
			"reason":   req.Reason,
		}).Error("Failed to suspend user")

		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"message": "The specified user does not exist",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to suspend user",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ReactivateUser 重新激活用户
// @Summary 重新激活用户
// @Description 重新激活已暂停或停用的用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Security ApiKeyAuth
// @Success 200 {object} OperationResponse
// @Failure 400 {object} object
// @Failure 401 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/admin/user-management/users/{user_id}/reactivate [post]
func (h *Handler) ReactivateUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "User ID is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 获取管理员信息
	adminInfo := h.getAdminInfoFromContext(c)
	ipAddress := c.ClientIP()

	response, err := h.service.ReactivateUser(ctx, userID, adminInfo.ID, adminInfo.Email, ipAddress)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":  userID,
			"admin_id": adminInfo.ID,
		}).Error("Failed to reactivate user")

		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"message": "The specified user does not exist",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to reactivate user",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ResetUserPassword 重置用户密码
// @Summary 重置用户密码
// @Description 管理员重置用户密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param request body ResetPasswordRequest true "重置密码请求"
// @Security ApiKeyAuth
// @Success 200 {object} OperationResponse
// @Failure 400 {object} object
// @Failure 401 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/admin/user-management/users/{user_id}/reset-password [post]
func (h *Handler) ResetUserPassword(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "User ID is required",
		})
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid reset password request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 获取管理员信息
	adminInfo := h.getAdminInfoFromContext(c)
	ipAddress := c.ClientIP()

	response, err := h.service.ResetUserPassword(ctx, userID, adminInfo.ID, adminInfo.Email, ipAddress, &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":  userID,
			"admin_id": adminInfo.ID,
		}).Error("Failed to reset user password")

		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"message": "The specified user does not exist",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to reset user password",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ForceLogoutUser 强制用户登出
// @Summary 强制用户登出
// @Description 强制用户登出所有或指定会话
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Param request body ForceLogoutRequest true "强制登出请求"
// @Security ApiKeyAuth
// @Success 200 {object} OperationResponse
// @Failure 400 {object} object
// @Failure 401 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/admin/user-management/users/{user_id}/force-logout [post]
func (h *Handler) ForceLogoutUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "User ID is required",
		})
		return
	}

	var req ForceLogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid force logout request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 获取管理员信息
	adminInfo := h.getAdminInfoFromContext(c)
	ipAddress := c.ClientIP()

	response, err := h.service.ForceLogoutUser(ctx, userID, adminInfo.ID, adminInfo.Email, ipAddress, &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":  userID,
			"admin_id": adminInfo.ID,
			"reason":   req.Reason,
		}).Error("Failed to force logout user")

		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"message": "The specified user does not exist",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to force logout user",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// VerifyUserEmail 验证用户邮箱
// @Summary 验证用户邮箱
// @Description 管理员手动验证用户邮箱
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user_id path string true "用户ID"
// @Security ApiKeyAuth
// @Success 200 {object} OperationResponse
// @Failure 400 {object} object
// @Failure 401 {object} object
// @Failure 404 {object} object
// @Failure 500 {object} object
// @Router /api/v1/admin/user-management/users/{user_id}/verify-email [post]
func (h *Handler) VerifyUserEmail(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": "User ID is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 获取管理员信息
	adminInfo := h.getAdminInfoFromContext(c)
	ipAddress := c.ClientIP()

	response, err := h.service.VerifyUserEmail(ctx, userID, adminInfo.ID, adminInfo.Email, ipAddress)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":  userID,
			"admin_id": adminInfo.ID,
		}).Error("Failed to verify user email")

		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "User not found",
				"message": "The specified user does not exist",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to verify user email",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// === 辅助方法 ===

// AdminInfo 管理员信息结构
type AdminInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// getAdminInfoFromContext 从上下文获取管理员信息
func (h *Handler) getAdminInfoFromContext(c *gin.Context) *AdminInfo {
	// 从JWT中间件获取管理员信息
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		return &AdminInfo{ID: "unknown", Email: "unknown", Role: "unknown"}
	}

	email, exists := c.Get("email")
	if !exists {
		h.logger.Error("Email not found in context")
		email = "unknown"
	}

	role, exists := c.Get("role")
	if !exists {
		h.logger.Error("Role not found in context")
		role = "unknown"
	}

	return &AdminInfo{
		ID:    userID.(string),
		Email: email.(string),
		Role:  role.(string),
	}
}
