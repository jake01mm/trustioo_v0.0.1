package admin

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"trusioo_api_v0.0.1/internal/modules/auth"
	"trusioo_api_v0.0.1/internal/modules/auth/user"
	"trusioo_api_v0.0.1/pkg/cryptoutil"

	"github.com/sirupsen/logrus"
)

// Service 管理员认证服务
type Service struct {
	repo       *Repository
	verifyRepo *user.VerificationRepository
	encryptor  *cryptoutil.PasswordEncryptor
	logger     *logrus.Logger
}

// Admin结构体已移至model.go文件

// NewService 创建新的管理员认证服务
func NewService(repo *Repository, verifyRepo *user.VerificationRepository, encryptor *cryptoutil.PasswordEncryptor, logger *logrus.Logger) *Service {
	return &Service{
		repo:       repo,
		verifyRepo: verifyRepo,
		encryptor:  encryptor,
		logger:     logger,
	}
}

// ValidateCredentials 验证管理员凭证
func (s *Service) ValidateCredentials(ctx context.Context, email, password string) (*Admin, error) {
	// 根据邮箱获取管理员信息
	admin, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("admin not found: %w", err)
	}

	// 检查管理员是否激活
	if !admin.Active {
		return nil, errors.New("admin account is deactivated")
	}

	// 验证密码
	if err := s.encryptor.VerifyPassword(password, admin.Password); err != nil {
		return nil, errors.New("invalid password")
	}

	return admin, nil
}

// GetByID 根据ID获取管理员信息
func (s *Service) GetByID(ctx context.Context, id string) (*Admin, error) {
	admin, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin by ID: %w", err)
	}

	if !admin.Active {
		return nil, errors.New("admin account is deactivated")
	}

	return admin, nil
}

// GetByEmail 根据邮箱获取管理员信息
func (s *Service) GetByEmail(ctx context.Context, email string) (*Admin, error) {
	admin, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin by email: %w", err)
	}

	if !admin.Active {
		return nil, errors.New("admin account is deactivated")
	}

	return admin, nil
}

// UpdatePassword 更新管理员密码
func (s *Service) UpdatePassword(ctx context.Context, adminID, newPassword string) error {
	// 加密新密码
	hashedPassword, err := s.encryptor.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新数据库中的密码
	if err := s.repo.UpdatePassword(ctx, adminID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.WithField("admin_id", adminID).Info("Admin password updated")
	return nil
}

// CreateAdmin 创建新管理员（系统内部使用）
func (s *Service) CreateAdmin(ctx context.Context, email, name, password, role string) (*Admin, error) {
	// 检查邮箱是否已存在
	if _, err := s.repo.GetByEmail(ctx, email); err == nil {
		return nil, errors.New("admin with this email already exists")
	}

	// 加密密码
	hashedPassword, err := s.encryptor.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	admin := &Admin{
		Email:    email,
		Name:     name,
		Password: hashedPassword,
		Role:     role,
		Active:   true,
	}

	// 创建管理员
	if err := s.repo.Create(ctx, admin); err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"admin_id": admin.ID,
		"email":    admin.Email,
		"role":     admin.Role,
	}).Info("New admin created")

	return admin, nil
}

// UpdateAdminStatus 更新管理员状态
func (s *Service) UpdateAdminStatus(ctx context.Context, adminID string, active bool) error {
	if err := s.repo.UpdateStatus(ctx, adminID, active); err != nil {
		return fmt.Errorf("failed to update admin status: %w", err)
	}

	status := "deactivated"
	if active {
		status = "activated"
	}

	s.logger.WithFields(logrus.Fields{
		"admin_id": adminID,
		"status":   status,
	}).Info("Admin status updated")

	return nil
}

// ListAdmins 获取管理员列表
func (s *Service) ListAdmins(ctx context.Context, limit, offset int) ([]*Admin, error) {
	admins, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list admins: %w", err)
	}

	return admins, nil
}

// SendLoginVerificationCode 发送管理员登录验证码
func (s *Service) SendLoginVerificationCode(ctx context.Context, email, password, ipAddress string) (string, error) {
	// 首先验证管理员凭证
	admin, err := s.ValidateCredentials(ctx, email, password)
	if err != nil {
		return "", err
	}

	// 检查频率限制（5分钟内最多3次）
	allowed, err := s.verifyRepo.CheckRateLimit(ctx, email, "admin", "login_code", 5*time.Minute, 3)
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
	verification := &user.EmailVerification{
		Email:            email,
		UserType:         "admin",
		Type:             "login_code",
		VerificationCode: code,
		Attempts:         0,
		MaxAttempts:      3,
		Verified:         false,
		IPAddress:        &ipAddress,
		ReferenceID:      &admin.ID,
		ExpiresAt:        time.Now().Add(5 * time.Minute), // 5分钟过期
	}

	if err := s.verifyRepo.CreateVerification(ctx, verification); err != nil {
		return "", fmt.Errorf("failed to create verification: %w", err)
	}

	// TODO: 这里应该发送邮件，现在先记录日志
	s.logger.WithFields(logrus.Fields{
		"email": email,
		"code":  code, // 生产环境中不应该记录验证码
		"type":  "admin_login_code",
	}).Info("Admin login verification code generated")

	return code, nil
}

// VerifyLoginCode 验证管理员登录验证码
func (s *Service) VerifyLoginCode(ctx context.Context, email, password, code string) (*Admin, error) {
	// 首先验证管理员凭证
	admin, err := s.ValidateCredentials(ctx, email, password)
	if err != nil {
		return nil, err
	}

	// 获取活跃的验证记录
	verification, err := s.verifyRepo.GetActiveVerification(ctx, email, "admin", "login_code")
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

	s.logger.WithFields(logrus.Fields{
		"admin_id": admin.ID,
		"email":    admin.Email,
		"role":     admin.Role,
	}).Info("Admin login verification successful")

	return admin, nil
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

// ForgotPassword 忘记密码，发送密码重置验证码
func (s *Service) ForgotPassword(ctx context.Context, email, ipAddress string) (string, error) {
	// 验证邮箱是否存在
	_, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", auth.ErrAdminNotFound
	}

	// 检查频率限制（5分钟内最多3次）
	allowed, err := s.verifyRepo.CheckRateLimit(ctx, email, "admin", "password_reset", 5*time.Minute, 3)
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
	verification := &user.EmailVerification{
		Email:            email,
		UserType:         "admin",
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
		"type":  "admin_password_reset",
	}).Info("Admin password reset verification code generated")

	return code, nil
}

// ResetPassword 重置密码
func (s *Service) ResetPassword(ctx context.Context, email, code, newPassword, ipAddress string) error {
	// 验证邮箱是否存在
	admin, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return auth.ErrAdminNotFound
	}

	// 获取活跃的验证记录
	verification, err := s.verifyRepo.GetActiveVerification(ctx, email, "admin", "password_reset")
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
	if err := s.repo.UpdatePassword(ctx, admin.ID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 记录密码重置操作到password_resets表
	if err := s.repo.CreatePasswordReset(ctx, &PasswordReset{
		Email:     email,
		UserType:  "admin",
		Token:     verification.ID, // 使用验证ID作为token
		Used:      true,
		UsedAt:    verification.VerifiedAt,
		ExpiresAt: verification.ExpiresAt,
	}); err != nil {
		s.logger.WithError(err).Warn("Failed to record password reset")
		// 不返回错误，因为密码已经重置成功
	}

	s.logger.WithFields(logrus.Fields{
		"admin_id": admin.ID,
		"email":    admin.Email,
	}).Info("Admin password reset successfully")

	return nil
}
