package admin

import (
	"context"
	"net/http"
	"time"

	"trusioo_api_v0.0.1/internal/modules/auth"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler 管理员认证处理器
type Handler struct {
	service    *Service
	jwtManager *auth.JWTManager
	logger     *logrus.Logger
}

// NewHandler 创建新的管理员认证处理器
func NewHandler(service *Service, jwtManager *auth.JWTManager, logger *logrus.Logger) *Handler {
	return &Handler{
		service:    service,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// 请求和响应结构体已移至dto.go文件

// Login 管理员登录（发送验证码）
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid login request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 获取客户端IP地址
	ipAddress := c.ClientIP()

	// 发送管理员登录验证码
	verificationCode, err := h.service.SendLoginVerificationCode(ctx, req.Email, req.Password, ipAddress)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"email": req.Email,
			"error": err.Error(),
		}).Warn("Failed to send admin login verification code")

		statusCode := http.StatusUnauthorized
		message := "Invalid email or password"

		// 根据错误类型返回不同的响应
		switch err.Error() {
		case "too many verification attempts":
			statusCode = http.StatusTooManyRequests
			message = "Too many attempts, please try again later"
		}

		c.JSON(statusCode, gin.H{
			"error":   "Login failed",
			"message": message,
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"email": req.Email,
	}).Info("Admin login verification code sent")

	c.JSON(http.StatusOK, LoginResponse{
		Message:          "Verification code sent to your email",
		VerificationCode: verificationCode, // 仅用于测试
		ExpiresIn:        300,              // 5分钟
	})
}

// VerifyLogin 验证管理员登录验证码
func (h *Handler) VerifyLogin(c *gin.Context) {
	var req VerifyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid verify login request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 验证管理员登录验证码
	admin, err := h.service.VerifyLoginCode(ctx, req.Email, req.Password, req.LoginCode)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"email": req.Email,
			"error": err.Error(),
		}).Warn("Admin login verification failed")

		statusCode := http.StatusUnauthorized
		message := "Invalid verification code"

		// 根据错误类型返回不同的响应
		switch err.Error() {
		case "too many verification attempts":
			statusCode = http.StatusTooManyRequests
			message = "Too many attempts, please try again later"
		case "admin not found":
			message = "Invalid email or password"
		}

		c.JSON(statusCode, gin.H{
			"error":   "Verification failed",
			"message": message,
		})
		return
	}

	// 生成JWT令牌
	tokens, err := h.jwtManager.GenerateTokenPair(admin.ID, admin.Email, admin.Role, "admin")
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate tokens")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to generate authentication tokens",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"admin_id": admin.ID,
		"email":    admin.Email,
		"role":     admin.Role,
	}).Info("Admin login verification successful")

	c.JSON(http.StatusOK, VerifyLoginResponse{
		Message: "Login successful",
		Admin:   admin.ToAdminInfo(),
		Tokens:  tokens,
	})
}

// RefreshToken 刷新访问令牌
func (h *Handler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid refresh token request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// 验证刷新令牌
	claims, err := h.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.WithError(err).Warn("Invalid refresh token")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid refresh token",
			"message": "The provided refresh token is invalid or expired",
		})
		return
	}

	// 获取管理员信息
	admin, err := h.service.GetByID(ctx, claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get admin by ID")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication failed",
			"message": "Admin not found",
		})
		return
	}

	// 生成新的令牌对
	newTokens, err := h.jwtManager.RefreshTokenPair(req.RefreshToken, admin.Email, admin.Role)
	if err != nil {
		h.logger.WithError(err).Error("Failed to refresh tokens")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to refresh authentication tokens",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"admin_id": admin.ID,
		"email":    admin.Email,
	}).Info("Admin token refreshed")

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"tokens":  newTokens,
	})
}

// Logout 管理员登出
func (h *Handler) Logout(c *gin.Context) {
	// 获取当前管理员信息
	claims, err := auth.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
		return
	}

	// 记录登出
	h.logger.WithFields(logrus.Fields{
		"admin_id": claims.UserID,
		"email":    claims.Email,
	}).Info("Admin logout")

	// 在实际应用中，这里可以将令牌加入黑名单
	// 或者在数据库中标记为已登出

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// GetProfile 获取管理员资料
func (h *Handler) GetProfile(c *gin.Context) {
	claims, err := auth.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	admin, err := h.service.GetByID(ctx, claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get admin profile")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to retrieve admin profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"admin": AdminInfo{
			ID:    admin.ID,
			Email: admin.Email,
			Name:  admin.Name,
			Role:  admin.Role,
		},
	})
}

// 修改密码请求结构体已移至dto.go文件

// ChangePassword 修改密码
func (h *Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid change password request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	claims, err := auth.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 验证当前密码
	admin, err := h.service.ValidateCredentials(ctx, claims.Email, req.CurrentPassword)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"admin_id": claims.UserID,
			"email":    claims.Email,
		}).Warn("Current password validation failed")

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid current password",
			"message": "The current password is incorrect",
		})
		return
	}

	// 更新密码
	if err := h.service.UpdatePassword(ctx, admin.ID, req.NewPassword); err != nil {
		h.logger.WithError(err).Error("Failed to update admin password")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to update password",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"admin_id": admin.ID,
		"email":    admin.Email,
	}).Info("Admin password changed")

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// ForgotPassword 忘记密码，发送密码重置验证码
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid forgot password request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 获取客户端IP地址
	ipAddress := c.ClientIP()

	// 发送密码重置验证码
	verificationCode, err := h.service.ForgotPassword(ctx, req.Email, ipAddress)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"email": req.Email,
			"error": err.Error(),
		}).Warn("Failed to send password reset verification code")

		statusCode := http.StatusBadRequest
		message := ""

		// 根据错误类型返回不同的响应
		if err == auth.ErrAdminNotFound {
			message = "Please register first, email not found"
		} else if err == auth.ErrTooManyAttempts {
			statusCode = http.StatusTooManyRequests
			message = "Too many attempts, please try again later"
		} else {
			message = "Failed to send verification code"
		}

		c.JSON(statusCode, gin.H{
			"error":   "Request failed",
			"message": message,
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"email": req.Email,
	}).Info("Password reset verification code sent")

	c.JSON(http.StatusOK, ForgotPasswordResponse{
		Message:          "Password reset code sent to your email",
		Email:            req.Email,
		VerificationCode: verificationCode, // 仅用于测试
		ExpiresIn:        900,              // 15分钟
	})
}

// ResetPassword 重置密码
func (h *Handler) ResetPassword(c *gin.Context) {
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

	// 获取客户端IP地址
	ipAddress := c.ClientIP()

	// 重置密码
	err := h.service.ResetPassword(ctx, req.Email, req.VerificationCode, req.NewPassword, ipAddress)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"email": req.Email,
			"error": err.Error(),
		}).Warn("Failed to reset password")

		statusCode := http.StatusBadRequest
		message := ""

		// 根据错误类型返回不同的响应
		if err == auth.ErrAdminNotFound {
			message = "Admin not found"
		} else if err == auth.ErrInvalidVerificationCode {
			message = "Invalid or expired verification code"
		} else if err == auth.ErrTooManyAttempts {
			statusCode = http.StatusTooManyRequests
			message = "Too many attempts, please try again later"
		} else {
			message = "Failed to reset password"
		}

		c.JSON(statusCode, gin.H{
			"error":   "Reset failed",
			"message": message,
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"email": req.Email,
	}).Info("Password reset successfully")

	c.JSON(http.StatusOK, ResetPasswordResponse{
		Message: "Password reset successfully",
	})
}
