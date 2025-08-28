package user

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"trusioo_api_v0.0.1/internal/modules/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Handler 用户认证处理器
type Handler struct {
	service    *Service
	jwtManager *auth.JWTManager
	logger     *logrus.Logger
}

// NewHandler 创建新的用户认证处理器
func NewHandler(service *Service, jwtManager *auth.JWTManager, logger *logrus.Logger) *Handler {
	return &Handler{
		service:    service,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// 请求和响应结构体已移至dto.go文件

// Register 用户注册（简化版，仅需email和password）
func (h *Handler) Register(c *gin.Context) {
	var req SimpleRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid register request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 创建新用户
	user, err := h.service.CreateSimpleUser(ctx, req.Email, req.Password)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create user")
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Registration failed",
			"message": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("New user registered")

	c.JSON(http.StatusCreated, SimpleRegisterResponse{
		Message: "Registration successful",
		UserID:  user.ID,
	})
}

// Login 用户登录（发送验证码）
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

	// 发送登录验证码
	verificationCode, err := h.service.SendLoginVerificationCode(ctx, req.Email, req.Password, ipAddress)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"email": req.Email,
			"error": err.Error(),
		}).Warn("Failed to send login verification code")

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
	}).Info("Login verification code sent")

	c.JSON(http.StatusOK, LoginResponse{
		Message:          "Verification code sent to your email",
		VerificationCode: verificationCode, // 仅用于测试
		ExpiresIn:        300,              // 5分钟
	})
}

// GetProfile 获取用户资料
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

	user, err := h.service.GetByID(ctx, claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user profile")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to retrieve user profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": UserInfo{
			ID:     user.ID,
			Email:  user.Email,
			Name:   user.Name,
			Status: user.Status,
		},
	})
}

// Logout 用户登出
func (h *Handler) Logout(c *gin.Context) {
	claims, err := auth.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": claims.UserID,
		"email":   claims.Email,
	}).Info("User logout")

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// VerifyLogin 验证登录验证码
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

	// 解析设备和位置信息（用于日志记录）
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	deviceInfo := h.parseDeviceInfo(userAgent)
	locationInfo := h.parseLocationInfo(ipAddress)

	// 验证登录验证码
	user, err := h.service.VerifyLoginCode(ctx, req.Email, req.Password, req.LoginCode)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"email": req.Email,
			"error": err.Error(),
		}).Warn("Login verification failed")

		// 记录失败登录日志
		failureReason := FailureReasonInvalidCode
		riskScore := 3 // 默认风险分数
		switch err.Error() {
		case "too many verification attempts":
			failureReason = FailureReasonTooManyAttempts
			riskScore = 8
		case "user not found":
			failureReason = FailureReasonInvalidCredentials
			riskScore = 5
		}

		// 异步记录失败日志（不影响响应速度）
		go func() {
			logCtx, logCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer logCancel()
			if logErr := h.service.LogFailedLogin(logCtx, req.Email, ipAddress, &userAgent, deviceInfo, locationInfo, failureReason, riskScore); logErr != nil {
				h.logger.WithError(logErr).Error("Failed to log failed login attempt")
			}
		}()

		statusCode := http.StatusUnauthorized
		message := "Invalid verification code"

		// 根据错误类型返回不同的响应
		switch err.Error() {
		case "too many verification attempts":
			statusCode = http.StatusTooManyRequests
			message = "Too many attempts, please try again later"
		case "user not found":
			message = "Invalid email or password"
		}

		c.JSON(statusCode, gin.H{
			"error":   "Verification failed",
			"message": message,
		})
		return
	}

	// 生成JWT令牌（带设备信息和IP地址）
	tokens, err := h.jwtManager.GenerateTokenPairWithContext(ctx, user.ID, user.Email, "user", "user", deviceInfo, &ipAddress)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate tokens")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to generate authentication tokens",
		})
		return
	}

	// 创建用户会话
	session := &UserSession{
		SessionID:    uuid.New().String(),
		UserID:       user.ID,
		UserType:     "user",
		IPAddress:    ipAddress,
		UserAgent:    &req.UserAgent,
		DeviceInfo:   deviceInfo,
		LocationInfo: locationInfo,
		IsActive:     true,
		ExpiresAt:    GetDefaultExpirationTime(false), // 使用默认过期时间
	}

	// 保存会话到数据库
	if err := h.service.CreateUserSession(ctx, session); err != nil {
		h.logger.WithError(err).Warn("Failed to create user session")
		// 不返回错误，因为这不影响登录成功
	}

	// 记录成功登录日志（异步处理）
	go func() {
		logCtx, logCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer logCancel()
		if logErr := h.service.LogSuccessfulLogin(logCtx, user.ID, user.Email, ipAddress, &userAgent, deviceInfo, locationInfo, &session.SessionID); logErr != nil {
			h.logger.WithError(logErr).Error("Failed to log successful login")
		}
	}()

	h.logger.WithFields(logrus.Fields{
		"user_id":    user.ID,
		"email":      user.Email,
		"session_id": session.SessionID,
	}).Info("User login verification successful")

	c.JSON(http.StatusOK, VerifyLoginResponse{
		Message: "Login successful",
		User:    user.ToUserInfo(),
		Tokens:  tokens,
		Session: session.ToUserSessionInfo(), // 返回会话信息
	})
}

// ========== 辅助方法 ==========

// parseDeviceInfo 解析设备信息
func (h *Handler) parseDeviceInfo(userAgent string) *map[string]interface{} {
	if userAgent == "" {
		return nil
	}

	deviceInfo := make(map[string]interface{})
	ua := strings.ToLower(userAgent)

	// 浏览器检测
	if strings.Contains(ua, "chrome") && !strings.Contains(ua, "edge") {
		deviceInfo["browser"] = "Chrome"
	} else if strings.Contains(ua, "firefox") {
		deviceInfo["browser"] = "Firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		deviceInfo["browser"] = "Safari"
	} else if strings.Contains(ua, "edge") {
		deviceInfo["browser"] = "Edge"
	} else {
		deviceInfo["browser"] = "Unknown"
	}

	// 操作系统检测
	if strings.Contains(ua, "windows") {
		deviceInfo["os"] = "Windows"
	} else if strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os") {
		deviceInfo["os"] = "macOS"
	} else if strings.Contains(ua, "linux") {
		deviceInfo["os"] = "Linux"
	} else if strings.Contains(ua, "android") {
		deviceInfo["os"] = "Android"
	} else if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		deviceInfo["os"] = "iOS"
	} else {
		deviceInfo["os"] = "Unknown"
	}

	// 设备类型检测
	if strings.Contains(ua, "mobile") {
		deviceInfo["device_type"] = "Mobile"
	} else if strings.Contains(ua, "tablet") {
		deviceInfo["device_type"] = "Tablet"
	} else {
		deviceInfo["device_type"] = "Desktop"
	}

	return &deviceInfo
}

// parseLocationInfo 解析位置信息
func (h *Handler) parseLocationInfo(ipAddress string) *map[string]interface{} {
	if ipAddress == "" {
		return nil
	}

	locationInfo := make(map[string]interface{})
	locationInfo["ip"] = ipAddress

	// 判断IP类型
	if ip := net.ParseIP(ipAddress); ip != nil {
		if ip.IsPrivate() {
			locationInfo["type"] = "private"
		} else if ip.IsLoopback() {
			locationInfo["type"] = "loopback"
		} else {
			locationInfo["type"] = "public"
		}
	} else {
		locationInfo["type"] = "invalid"
	}

	return &locationInfo
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
		switch err {
		case auth.ErrUserNotFound:
			message = "Please register first, email not found"
		case auth.ErrTooManyAttempts:
			statusCode = http.StatusTooManyRequests
			message = "Too many attempts, please try again later"
		default:
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
		switch err {
		case auth.ErrUserNotFound:
			message = "User not found"
		case auth.ErrInvalidVerificationCode:
			message = "Invalid or expired verification code"
		case auth.ErrTooManyAttempts:
			statusCode = http.StatusTooManyRequests
			message = "Too many attempts, please try again later"
		default:
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
