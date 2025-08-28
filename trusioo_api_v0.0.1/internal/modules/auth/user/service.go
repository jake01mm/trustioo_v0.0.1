package user

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"trusioo_api_v0.0.1/internal/modules/auth"
	"trusioo_api_v0.0.1/pkg/cryptoutil"

	"github.com/sirupsen/logrus"
)

// Service 用户认证服务
type Service struct {
	repo       *Repository
	verifyRepo *VerificationRepository
	encryptor  *cryptoutil.PasswordEncryptor
	logger     *logrus.Logger
}

// User结构体已移至model.go文件

// NewService 创建新的用户认证服务
func NewService(repo *Repository, verifyRepo *VerificationRepository, encryptor *cryptoutil.PasswordEncryptor, logger *logrus.Logger) *Service {
	return &Service{
		repo:       repo,
		verifyRepo: verifyRepo,
		encryptor:  encryptor,
		logger:     logger,
	}
}

// CreateUser 创建新用户
func (s *Service) CreateUser(ctx context.Context, email, name, password string) (*User, error) {
	// 检查邮箱是否已存在
	if exists, err := s.repo.ExistsByEmail(ctx, email); err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	} else if exists {
		return nil, errors.New("user with this email already exists")
	}

	// 加密密码
	hashedPassword, err := s.encryptor.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Email:    email,
		Name:     name,
		Password: hashedPassword,
		Status:   "active",
	}

	// 创建用户
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("New user created")

	return user, nil
}

// ValidateCredentials 验证用户凭证
func (s *Service) ValidateCredentials(ctx context.Context, email, password string) (*User, error) {
	// 根据邮箱获取用户信息
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, fmt.Errorf("user account is %s", user.Status)
	}

	// 验证密码
	if err := s.encryptor.VerifyPassword(password, user.Password); err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// GetByID 根据ID获取用户信息
func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetByEmail 根据邮箱获取用户信息
func (s *Service) GetByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// UpdatePassword 更新用户密码
func (s *Service) UpdatePassword(ctx context.Context, userID, newPassword string) error {
	// 加密新密码
	hashedPassword, err := s.encryptor.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新数据库中的密码
	if err := s.repo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.WithField("user_id", userID).Info("User password updated")
	return nil
}

// UpdateStatus 更新用户状态
func (s *Service) UpdateStatus(ctx context.Context, userID, status string) error {
	validStatuses := map[string]bool{"active": true, "inactive": true, "suspended": true}
	if !validStatuses[status] {
		return errors.New("invalid status")
	}

	if err := s.repo.UpdateStatus(ctx, userID, status); err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"status":  status,
	}).Info("User status updated")

	return nil
}

// CreateSimpleUser 创建简单用户（仅需email和password）
func (s *Service) CreateSimpleUser(ctx context.Context, email, password string) (*User, error) {
	// 检查邮箱是否已存在
	if exists, err := s.repo.ExistsByEmail(ctx, email); err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	} else if exists {
		return nil, errors.New("user with this email already exists")
	}

	// 加密密码
	hashedPassword, err := s.encryptor.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Email:         email,
		Name:          email, // 使用邮箱作为默认名称
		Password:      hashedPassword,
		Status:        "active",
		EmailVerified: false, // 注册时邮箱未验证
	}

	// 创建用户
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("Simple user created")

	return user, nil
}

// SendLoginVerificationCode 发送登录验证码
func (s *Service) SendLoginVerificationCode(ctx context.Context, email, password, ipAddress string) (string, error) {
	// 首先验证用户凭证
	user, err := s.ValidateCredentials(ctx, email, password)
	if err != nil {
		return "", err
	}

	// 检查频率限制（5分钟内最多3次）
	allowed, err := s.verifyRepo.CheckRateLimit(ctx, email, "user", "login_code", 5*time.Minute, 3)
	if err != nil {
		return "", fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return "", auth.ErrTooManyAttempts
	}

	// 生成6位数字验证码
	code, err := s.generateVerificationCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate verification code: %w", err)
	}

	// 创建验证记录
	verification := &EmailVerification{
		Email:            email,
		UserType:         "user",
		Type:             "login_code",
		VerificationCode: code,
		Attempts:         0,
		MaxAttempts:      3,
		Verified:         false,
		IPAddress:        &ipAddress,
		ReferenceID:      &user.ID,
		ExpiresAt:        time.Now().Add(5 * time.Minute), // 5分钟过期
	}

	if err := s.verifyRepo.CreateVerification(ctx, verification); err != nil {
		return "", fmt.Errorf("failed to create verification: %w", err)
	}

	// TODO: 这里应该发送邮件，现在先记录日志
	s.logger.WithFields(logrus.Fields{
		"email": email,
		"code":  code, // 生产环境中不应该记录验证码
		"type":  "login_code",
	}).Info("Login verification code generated")

	return code, nil
}

// VerifyLoginCode 验证登录验证码
func (s *Service) VerifyLoginCode(ctx context.Context, email, password, code string) (*User, error) {
	// 首先验证用户凭证
	user, err := s.ValidateCredentials(ctx, email, password)
	if err != nil {
		return nil, err
	}

	// 获取活跃的验证记录
	verification, err := s.verifyRepo.GetActiveVerification(ctx, email, "user", "login_code")
	if err != nil {
		return nil, auth.ErrInvalidVerificationCode
	}

	// 检查验证码是否匹配
	if verification.VerificationCode != code {
		// 增加尝试次数
		if err := s.verifyRepo.IncrementAttempts(ctx, verification.ID); err != nil {
			s.logger.WithError(err).Error("Failed to increment attempts")
		}
		return nil, auth.ErrInvalidVerificationCode
	}

	// 检查是否可以继续尝试
	if !verification.CanAttempt() {
		return nil, auth.ErrTooManyAttempts
	}

	// 标记验证码为已使用
	if err := s.verifyRepo.MarkAsVerified(ctx, verification.ID); err != nil {
		return nil, fmt.Errorf("failed to mark verification as used: %w", err)
	}

	// 更新用户的邮箱验证状态（首次验证登录后）
	if !user.EmailVerified {
		if err := s.repo.UpdateEmailVerified(ctx, user.ID, true); err != nil {
			s.logger.WithError(err).Error("Failed to update email verified status")
		}
		user.EmailVerified = true
		user.VerifyEmail()
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("Login verification successful")

	return user, nil
}

// generateVerificationCode 生成6位数字验证码
func (s *Service) generateVerificationCode() (string, error) {
	// 生成6位随机数字
	code := ""
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += n.String()
	}
	return code, nil
}

// ========== 会话相关方法 ==========

// CreateUserSession 创建用户会话
func (s *Service) CreateUserSession(ctx context.Context, session *UserSession) error {
	if err := s.repo.CreateUserSession(ctx, session); err != nil {
		s.logger.WithError(err).Error("Failed to create user session")
		return fmt.Errorf("failed to create user session: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session_id": session.SessionID,
		"user_id":    session.UserID,
		"ip_address": session.IPAddress,
	}).Info("User session created")

	return nil
}

// GetUserSessions 获取用户会话列表
func (s *Service) GetUserSessions(ctx context.Context, userID string, includeInactive bool) ([]*UserSession, error) {
	sessions, err := s.repo.GetUserSessions(ctx, userID, includeInactive)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	return sessions, nil
}

// LogoutAllSessions 登出用户所有会话
func (s *Service) LogoutAllSessions(ctx context.Context, userID string) error {
	if err := s.repo.DeactivateAllUserSessions(ctx, userID); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to deactivate all user sessions")
		return fmt.Errorf("failed to logout all sessions: %w", err)
	}

	s.logger.WithField("user_id", userID).Info("All user sessions deactivated")
	return nil
}

// ========== 登录日志相关方法 ==========

// CreateLoginLog 创建登录日志
func (s *Service) CreateLoginLog(ctx context.Context, log *LoginLog) error {
	if err := s.repo.CreateLoginLog(ctx, log); err != nil {
		s.logger.WithError(err).Error("Failed to create login log")
		return fmt.Errorf("failed to create login log: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":      log.UserID,
		"email":        log.Email,
		"login_status": log.LoginStatus,
		"ip_address":   log.IPAddress,
	}).Info("Login log created")

	return nil
}

// GetUserLoginLogs 获取用户登录日志列表
func (s *Service) GetUserLoginLogs(ctx context.Context, userID string, limit, offset int) ([]*LoginLog, error) {
	logs, err := s.repo.GetUserLoginLogs(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user login logs: %w", err)
	}

	return logs, nil
}

// LogSuccessfulLogin 记录成功登录
func (s *Service) LogSuccessfulLogin(ctx context.Context, userID, email, ipAddress string, userAgent *string, deviceInfo, locationInfo *map[string]interface{}, sessionID *string) error {
	log := &LoginLog{
		UserID:       &userID,
		Email:        email,
		UserType:     "user",
		LoginStatus:  LoginStatusSuccess,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		DeviceInfo:   deviceInfo,
		LocationInfo: locationInfo,
		SessionID:    sessionID,
		RiskScore:    0, // 成功登录，风险评分为0
	}

	return s.CreateLoginLog(ctx, log)
}

// LogFailedLogin 记录失败登录
func (s *Service) LogFailedLogin(ctx context.Context, email, ipAddress string, userAgent *string, deviceInfo, locationInfo *map[string]interface{}, failureReason string, riskScore int) error {
	log := &LoginLog{
		UserID:        nil, // 失败登录可能没有用户ID
		Email:         email,
		UserType:      "user",
		LoginStatus:   LoginStatusFailed,
		FailureReason: &failureReason,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		DeviceInfo:    deviceInfo,
		LocationInfo:  locationInfo,
		RiskScore:     riskScore,
	}

	return s.CreateLoginLog(ctx, log)
}

// ForgotPassword 忘记密码，发送密码重置验证码
func (s *Service) ForgotPassword(ctx context.Context, email, ipAddress string) (string, error) {
	// 验证邮箱是否存在
	_, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", auth.ErrUserNotFound
	}

	// 检查频率限制（5分钟内最多3次）
	allowed, err := s.verifyRepo.CheckRateLimit(ctx, email, "user", "password_reset", 5*time.Minute, 3)
	if err != nil {
		return "", fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		return "", auth.ErrTooManyAttempts
	}

	// 生成6位数字验证码
	code, err := s.generateVerificationCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate verification code: %w", err)
	}

	// 创建验证记录
	verification := &EmailVerification{
		Email:            email,
		UserType:         "user",
		Type:             "password_reset",
		VerificationCode: code,
		Attempts:         0,
		MaxAttempts:      3,
		Verified:         false,
		IPAddress:        &ipAddress,
		ExpiresAt:        time.Now().Add(15 * time.Minute), // 15分钟过期
	}

	if err := s.verifyRepo.CreateVerification(ctx, verification); err != nil {
		return "", fmt.Errorf("failed to create verification: %w", err)
	}

	// TODO: 这里应该发送邮件，现在先记录日志
	s.logger.WithFields(logrus.Fields{
		"email": email,
		"code":  code, // 生产环境中不应该记录验证码
		"type":  "user_password_reset",
	}).Info("User password reset verification code generated")

	return code, nil
}

// ResetPassword 重置密码
func (s *Service) ResetPassword(ctx context.Context, email, code, newPassword, ipAddress string) error {
	// 验证邮箱是否存在
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return auth.ErrUserNotFound
	}

	// 获取活跃的验证记录
	verification, err := s.verifyRepo.GetActiveVerification(ctx, email, "user", "password_reset")
	if err != nil {
		return auth.ErrInvalidVerificationCode
	}

	// 检查验证码是否匹配
	if verification.VerificationCode != code {
		// 增加尝试次数
		if err := s.verifyRepo.IncrementAttempts(ctx, verification.ID); err != nil {
			s.logger.WithError(err).Error("Failed to increment attempts")
		}
		return auth.ErrInvalidVerificationCode
	}

	// 检查是否可以继续尝试
	if !verification.CanAttempt() {
		return auth.ErrTooManyAttempts
	}

	// 标记验证码为已使用
	if err := s.verifyRepo.MarkAsVerified(ctx, verification.ID); err != nil {
		return fmt.Errorf("failed to mark verification as used: %w", err)
	}

	// 加密新密码
	hashedPassword, err := s.encryptor.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新数据库中的密码
	if err := s.repo.UpdatePassword(ctx, user.ID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 记录密码重置操作到password_resets表
	if err := s.repo.CreatePasswordReset(ctx, &PasswordReset{
		Email:    email,
		UserType: "user",
		Token:    verification.ID, // 使用验证ID作为token
		Used:     true,
		UsedAt:   &verification.VerifiedAt,
		ExpiresAt: verification.ExpiresAt,
	}); err != nil {
		s.logger.WithError(err).Warn("Failed to record password reset")
		// 不返回错误，因为密码已经重置成功
	}

	s.logger.WithFields(logrus.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("User password reset successfully")

	return nil
}
