package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"trusioo_api_v0.0.1/internal/infrastructure/database"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// EmailVerification 邮箱验证结构
type EmailVerification struct {
	ID               string     `json:"id" db:"id"`
	Email            string     `json:"email" db:"email"`
	UserType         string     `json:"user_type" db:"user_type"`
	Type             string     `json:"type" db:"type"`
	VerificationCode string     `json:"verification_code" db:"verification_code"`
	Token            *string    `json:"token" db:"token"`
	Attempts         int        `json:"attempts" db:"attempts"`
	MaxAttempts      int        `json:"max_attempts" db:"max_attempts"`
	Verified         bool       `json:"verified" db:"verified"`
	IPAddress        *string    `json:"ip_address" db:"ip_address"`
	ReferenceID      *string    `json:"reference_id" db:"reference_id"`
	ExpiresAt        time.Time  `json:"expires_at" db:"expires_at"`
	VerifiedAt       *time.Time `json:"verified_at" db:"verified_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// VerificationRepository 验证码仓储
type VerificationRepository struct {
	*database.BaseRepository
	logger *logrus.Logger
}

// NewVerificationRepository 创建新的验证码仓储
func NewVerificationRepository(db *database.Database, logger *logrus.Logger) *VerificationRepository {
	return &VerificationRepository{
		BaseRepository: database.NewBaseRepository(db, logger),
		logger:         logger,
	}
}

// CreateVerification 创建验证码记录
func (r *VerificationRepository) CreateVerification(ctx context.Context, verification *EmailVerification) error {
	// 生成UUID
	verification.ID = uuid.New().String()
	
	query := `
		INSERT INTO email_verifications (
			id, email, user_type, type, verification_code, token, 
			attempts, max_attempts, verified, ip_address, reference_id,
			expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
	`

	_, err := r.GetDB().ExecContext(ctx, query,
		verification.ID, verification.Email, verification.UserType, verification.Type,
		verification.VerificationCode, verification.Token, verification.Attempts,
		verification.MaxAttempts, verification.Verified, verification.IPAddress,
		verification.ReferenceID, verification.ExpiresAt)

	if err != nil {
		return fmt.Errorf("failed to create verification: %w", err)
	}

	return nil
}

// GetActiveVerification 获取有效的验证码
func (r *VerificationRepository) GetActiveVerification(ctx context.Context, email, userType, verificationType string) (*EmailVerification, error) {
	query := `
		SELECT id, email, user_type, type, verification_code, token, 
			   attempts, max_attempts, verified, ip_address, reference_id,
			   expires_at, verified_at, created_at
		FROM email_verifications
		WHERE email = $1 AND user_type = $2 AND type = $3 
		  AND verified = false AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	verification := &EmailVerification{}
	err := r.GetDB().QueryRowContext(ctx, query, email, userType, verificationType).Scan(
		&verification.ID, &verification.Email, &verification.UserType, &verification.Type,
		&verification.VerificationCode, &verification.Token, &verification.Attempts,
		&verification.MaxAttempts, &verification.Verified, &verification.IPAddress,
		&verification.ReferenceID, &verification.ExpiresAt, &verification.VerifiedAt,
		&verification.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active verification found")
		}
		return nil, fmt.Errorf("failed to get verification: %w", err)
	}

	return verification, nil
}

// IncrementAttempts 增加尝试次数
func (r *VerificationRepository) IncrementAttempts(ctx context.Context, verificationID string) error {
	query := `
		UPDATE email_verifications
		SET attempts = attempts + 1
		WHERE id = $1
	`

	result, err := r.GetDB().ExecContext(ctx, query, verificationID)
	if err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("verification not found")
	}

	return nil
}

// MarkAsVerified 标记为已验证
func (r *VerificationRepository) MarkAsVerified(ctx context.Context, verificationID string) error {
	query := `
		UPDATE email_verifications
		SET verified = true, verified_at = NOW()
		WHERE id = $1
	`

	result, err := r.GetDB().ExecContext(ctx, query, verificationID)
	if err != nil {
		return fmt.Errorf("failed to mark as verified: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("verification not found")
	}

	return nil
}

// CleanupExpired 清理过期的验证码
func (r *VerificationRepository) CleanupExpired(ctx context.Context) error {
	query := `DELETE FROM email_verifications WHERE expires_at < NOW() - INTERVAL '7 days'`
	
	_, err := r.GetDB().ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired verifications: %w", err)
	}

	return nil
}

// CheckRateLimit 检查频率限制
func (r *VerificationRepository) CheckRateLimit(ctx context.Context, email, userType, verificationType string, within time.Duration, maxCount int) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM email_verifications
		WHERE email = $1 AND user_type = $2 AND type = $3 
		  AND created_at > NOW() - INTERVAL '%d seconds'
	`

	var count int
	err := r.GetDB().QueryRowContext(ctx, fmt.Sprintf(query, int(within.Seconds())), email, userType, verificationType).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	return count < maxCount, nil
}

// IsExceedingAttempts 检查是否超过最大尝试次数
func (v *EmailVerification) IsExceedingAttempts() bool {
	return v.Attempts >= v.MaxAttempts
}

// IsExpired 检查是否已过期
func (v *EmailVerification) IsExpired() bool {
	return time.Now().After(v.ExpiresAt)
}

// CanAttempt 检查是否可以尝试验证
func (v *EmailVerification) CanAttempt() bool {
	return !v.Verified && !v.IsExpired() && !v.IsExceedingAttempts()
}