package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"trusioo_api_v0.0.1/internal/infrastructure/database"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Repository 用户仓储
type Repository struct {
	*database.BaseRepository
	logger *logrus.Logger
}

// NewRepository 创建新的用户仓储
func NewRepository(db *database.Database, logger *logrus.Logger) *Repository {
	return &Repository{
		BaseRepository: database.NewBaseRepository(db, logger),
		logger:         logger,
	}
}

// Create 创建用户
func (r *Repository) Create(ctx context.Context, user *User) error {
	// 生成UUID
	user.ID = uuid.New().String()

	query := `
		INSERT INTO users (id, email, name, password, status, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`

	_, err := r.GetDB().ExecContext(ctx, query,
		user.ID, user.Email, user.Name, user.Password, user.Status, user.EmailVerified)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID 根据ID获取用户
func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, name, password, status, email_verified, email_verified_at, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	user := &User{}
	err := r.GetDB().QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.Password, &user.Status, &user.EmailVerified, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, name, password, status, email_verified, email_verified_at, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	user := &User{}
	err := r.GetDB().QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.Password, &user.Status, &user.EmailVerified, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdatePassword 更新用户密码
func (r *Repository) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.GetDB().ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or no changes made")
	}

	return nil
}

// UpdateStatus 更新用户状态
func (r *Repository) UpdateStatus(ctx context.Context, userID, status string) error {
	query := `
		UPDATE users
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.GetDB().ExecContext(ctx, query, status, userID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or no changes made")
	}

	return nil
}

// ExistsByEmail 检查邮箱是否已存在
func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.GetDB().QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// UpdateEmailVerified 更新邮箱验证状态
func (r *Repository) UpdateEmailVerified(ctx context.Context, userID string, verified bool) error {
	query := `
		UPDATE users
		SET email_verified = $1, email_verified_at = CASE WHEN $1 = true THEN NOW() ELSE NULL END, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.GetDB().ExecContext(ctx, query, verified, userID)
	if err != nil {
		return fmt.Errorf("failed to update email verified status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or no changes made")
	}

	return nil
}

// List 获取用户列表
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*User, error) {
	query := `
		SELECT id, email, name, status
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.GetDB().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}

// ========== 会话相关方法 ==========

// CreateUserSession 创建用户会话
func (r *Repository) CreateUserSession(ctx context.Context, session *UserSession) error {
	// 生成UUID
	session.ID = uuid.New().String()
	session.CreatedAt = time.Now()
	session.LastActivity = time.Now()
	session.UserType = "user" // 固定为user类型

	// 序列化JSON字段
	var deviceInfoJSON, locationInfoJSON []byte
	var err error

	if session.DeviceInfo != nil {
		deviceInfoJSON, err = json.Marshal(session.DeviceInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal device info: %w", err)
		}
	}

	if session.LocationInfo != nil {
		locationInfoJSON, err = json.Marshal(session.LocationInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal location info: %w", err)
		}
	}

	query := `
		INSERT INTO user_sessions (
			id, session_id, user_id, user_type, refresh_token_id,
			ip_address, user_agent, device_info, location_info,
			is_active, last_activity, expires_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	_, err = r.GetDB().ExecContext(ctx, query,
		session.ID, session.SessionID, session.UserID, session.UserType,
		session.RefreshTokenID, session.IPAddress, session.UserAgent,
		deviceInfoJSON, locationInfoJSON, session.IsActive,
		session.LastActivity, session.ExpiresAt, session.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user session: %w", err)
	}

	return nil
}

// GetUserSessions 获取用户会话列表
func (r *Repository) GetUserSessions(ctx context.Context, userID string, includeInactive bool) ([]*UserSession, error) {
	query := `
		SELECT id, session_id, user_id, user_type, refresh_token_id,
		       ip_address, user_agent, device_info, location_info,
		       is_active, last_activity, expires_at, created_at
		FROM user_sessions
		WHERE user_id = $1 AND user_type = 'user'
	`

	args := []interface{}{userID}
	if !includeInactive {
		query += " AND is_active = true"
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.GetDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*UserSession
	for rows.Next() {
		session := &UserSession{}
		var deviceInfoJSON, locationInfoJSON sql.NullString

		err := rows.Scan(
			&session.ID, &session.SessionID, &session.UserID, &session.UserType,
			&session.RefreshTokenID, &session.IPAddress, &session.UserAgent,
			&deviceInfoJSON, &locationInfoJSON, &session.IsActive,
			&session.LastActivity, &session.ExpiresAt, &session.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan user session: %w", err)
		}

		// 反序列化JSON字段
		if deviceInfoJSON.Valid {
			var deviceInfo map[string]interface{}
			if err := json.Unmarshal([]byte(deviceInfoJSON.String), &deviceInfo); err == nil {
				session.DeviceInfo = &deviceInfo
			}
		}

		if locationInfoJSON.Valid {
			var locationInfo map[string]interface{}
			if err := json.Unmarshal([]byte(locationInfoJSON.String), &locationInfo); err == nil {
				session.LocationInfo = &locationInfo
			}
		}

		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user sessions: %w", err)
	}

	return sessions, nil
}

// DeactivateAllUserSessions 停用用户所有会话
func (r *Repository) DeactivateAllUserSessions(ctx context.Context, userID string) error {
	query := `
		UPDATE user_sessions
		SET is_active = false, last_activity = NOW()
		WHERE user_id = $1 AND user_type = 'user' AND is_active = true
	`

	_, err := r.GetDB().ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate all user sessions: %w", err)
	}

	return nil
}

// ========== 登录日志相关方法 ==========

// CreateLoginLog 创建登录日志
func (r *Repository) CreateLoginLog(ctx context.Context, log *LoginLog) error {
	// 生成UUID
	log.ID = uuid.New().String()
	log.CreatedAt = time.Now()
	log.UserType = "user" // 固定为user类型

	// 序列化JSON字段
	var deviceInfoJSON, locationInfoJSON []byte
	var err error

	if log.DeviceInfo != nil {
		deviceInfoJSON, err = json.Marshal(log.DeviceInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal device info: %w", err)
		}
	}

	if log.LocationInfo != nil {
		locationInfoJSON, err = json.Marshal(log.LocationInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal location info: %w", err)
		}
	}

	query := `
		INSERT INTO login_logs (
			id, user_id, email, user_type, login_status, failure_reason,
			ip_address, user_agent, device_info, location_info,
			session_id, risk_score, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	_, err = r.GetDB().ExecContext(ctx, query,
		log.ID, log.UserID, log.Email, log.UserType, log.LoginStatus,
		log.FailureReason, log.IPAddress, log.UserAgent,
		deviceInfoJSON, locationInfoJSON, log.SessionID,
		log.RiskScore, log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create login log: %w", err)
	}

	return nil
}

// GetUserLoginLogs 获取用户登录日志列表
func (r *Repository) GetUserLoginLogs(ctx context.Context, userID string, limit, offset int) ([]*LoginLog, error) {
	query := `
		SELECT id, user_id, email, user_type, login_status, failure_reason,
		       ip_address, user_agent, device_info, location_info,
		       session_id, risk_score, created_at
		FROM login_logs
		WHERE user_id = $1 AND user_type = 'user'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.GetDB().QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user login logs: %w", err)
	}
	defer rows.Close()

	var logs []*LoginLog
	for rows.Next() {
		log := &LoginLog{}
		var deviceInfoJSON, locationInfoJSON sql.NullString

		err := rows.Scan(
			&log.ID, &log.UserID, &log.Email, &log.UserType, &log.LoginStatus,
			&log.FailureReason, &log.IPAddress, &log.UserAgent,
			&deviceInfoJSON, &locationInfoJSON, &log.SessionID,
			&log.RiskScore, &log.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan login log: %w", err)
		}

		// 反序列化JSON字段
		if deviceInfoJSON.Valid {
			var deviceInfo map[string]interface{}
			if err := json.Unmarshal([]byte(deviceInfoJSON.String), &deviceInfo); err == nil {
				log.DeviceInfo = &deviceInfo
			}
		}

		if locationInfoJSON.Valid {
			var locationInfo map[string]interface{}
			if err := json.Unmarshal([]byte(locationInfoJSON.String), &locationInfo); err == nil {
				log.LocationInfo = &locationInfo
			}
		}

		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating login logs: %w", err)
	}

	return logs, nil
}
