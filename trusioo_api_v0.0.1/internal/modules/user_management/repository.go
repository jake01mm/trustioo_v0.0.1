package user_management

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"trusioo_api_v0.0.1/internal/infrastructure/database"
	"trusioo_api_v0.0.1/internal/modules/auth/user"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Repository 用户管理仓储
type Repository struct {
	*database.BaseRepository
	logger *logrus.Logger
}

// NewRepository 创建新的用户管理仓储
func NewRepository(db *database.Database, logger *logrus.Logger) *Repository {
	return &Repository{
		BaseRepository: database.NewBaseRepository(db, logger),
		logger:         logger,
	}
}

// === 用户查询方法 ===

// GetUserByID 根据ID获取用户详细信息
func (r *Repository) GetUserByID(ctx context.Context, userID string) (*UserManagementModel, error) {
	query := `
		SELECT 
			u.id, u.email, u.name, u.password, u.status, u.email_verified, u.email_verified_at,
			u.created_at, u.updated_at, u.deleted_at,
			COALESCE(stats.last_login_at, NULL) as last_login_at,
			COALESCE(stats.login_count, 0) as login_count,
			COALESCE(stats.failed_attempts, 0) as failed_attempts
		FROM users u
		LEFT JOIN (
			SELECT 
				user_id,
				MAX(created_at) as last_login_at,
				COUNT(*) as login_count,
				SUM(CASE WHEN login_status = 'failed' THEN 1 ELSE 0 END) as failed_attempts
			FROM login_logs 
			WHERE user_type = 'user' AND user_id IS NOT NULL
			GROUP BY user_id
		) stats ON u.id = stats.user_id
		WHERE u.id = $1
	`

	userModel := &UserManagementModel{User: &user.User{}}
	err := r.GetDB().QueryRowContext(ctx, query, userID).Scan(
		&userModel.ID, &userModel.Email, &userModel.Name, &userModel.Password,
		&userModel.Status, &userModel.EmailVerified, &userModel.EmailVerifiedAt,
		&userModel.CreatedAt, &userModel.UpdatedAt, &userModel.DeletedAt,
		&userModel.LastLoginAt, &userModel.LoginCount, &userModel.FailedAttempts,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return userModel, nil
}

// GetUsers 获取用户列表（支持分页和过滤）
func (r *Repository) GetUsers(ctx context.Context, filter *SearchFilter, pagination PaginationParams) ([]*UserManagementModel, int64, error) {
	// 构建WHERE条件
	whereConditions := []string{"u.deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if filter.Email != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.email ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Email+"%")
		argIndex++
	}

	if filter.Name != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Name+"%")
		argIndex++
	}

	if filter.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.status = $%d", argIndex))
		args = append(args, filter.Status.String())
		argIndex++
	}

	if filter.EmailVerified != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.email_verified = $%d", argIndex))
		args = append(args, *filter.EmailVerified)
		argIndex++
	}

	if filter.CreatedFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.created_at >= $%d", argIndex))
		args = append(args, *filter.CreatedFrom)
		argIndex++
	}

	if filter.CreatedTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.created_at <= $%d", argIndex))
		args = append(args, *filter.CreatedTo)
		argIndex++
	}

	if filter.LastLoginFrom != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("stats.last_login_at >= $%d", argIndex))
		args = append(args, *filter.LastLoginFrom)
		argIndex++
	}

	if filter.LastLoginTo != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("stats.last_login_at <= $%d", argIndex))
		args = append(args, *filter.LastLoginTo)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 先获取总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM users u
		LEFT JOIN (
			SELECT 
				user_id,
				MAX(created_at) as last_login_at,
				COUNT(*) as login_count,
				SUM(CASE WHEN login_status = 'failed' THEN 1 ELSE 0 END) as failed_attempts
			FROM login_logs 
			WHERE user_type = 'user' AND user_id IS NOT NULL
			GROUP BY user_id
		) stats ON u.id = stats.user_id
		WHERE %s
	`, whereClause)

	var total int64
	err := r.GetDB().QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 构建排序
	orderBy := fmt.Sprintf("u.%s %s", pagination.SortBy, strings.ToUpper(pagination.SortDir))

	// 获取数据
	dataQuery := fmt.Sprintf(`
		SELECT 
			u.id, u.email, u.name, u.status, u.email_verified, u.email_verified_at,
			u.created_at, u.updated_at, u.deleted_at,
			COALESCE(stats.last_login_at, NULL) as last_login_at,
			COALESCE(stats.login_count, 0) as login_count,
			COALESCE(stats.failed_attempts, 0) as failed_attempts
		FROM users u
		LEFT JOIN (
			SELECT 
				user_id,
				MAX(created_at) as last_login_at,
				COUNT(*) as login_count,
				SUM(CASE WHEN login_status = 'failed' THEN 1 ELSE 0 END) as failed_attempts
			FROM login_logs 
			WHERE user_type = 'user' AND user_id IS NOT NULL
			GROUP BY user_id
		) stats ON u.id = stats.user_id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, pagination.GetLimit(), pagination.GetOffset())

	rows, err := r.GetDB().QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*UserManagementModel
	for rows.Next() {
		userModel := &UserManagementModel{User: &user.User{}}
		err := rows.Scan(
			&userModel.ID, &userModel.Email, &userModel.Name, &userModel.Status,
			&userModel.EmailVerified, &userModel.EmailVerifiedAt,
			&userModel.CreatedAt, &userModel.UpdatedAt, &userModel.DeletedAt,
			&userModel.LastLoginAt, &userModel.LoginCount, &userModel.FailedAttempts,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, userModel)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, total, nil
}

// SearchUsers 搜索用户（支持模糊搜索）
func (r *Repository) SearchUsers(ctx context.Context, keyword string, pagination PaginationParams) ([]*UserManagementModel, int64, error) {
	whereCondition := "u.deleted_at IS NULL AND (u.email ILIKE $1 OR u.name ILIKE $1)"
	searchPattern := "%" + keyword + "%"

	// 获取总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM users u
		WHERE %s
	`, whereCondition)

	var total int64
	err := r.GetDB().QueryRowContext(ctx, countQuery, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// 获取数据
	orderBy := fmt.Sprintf("u.%s %s", pagination.SortBy, strings.ToUpper(pagination.SortDir))

	dataQuery := fmt.Sprintf(`
		SELECT 
			u.id, u.email, u.name, u.status, u.email_verified, u.email_verified_at,
			u.created_at, u.updated_at, u.deleted_at,
			COALESCE(stats.last_login_at, NULL) as last_login_at,
			COALESCE(stats.login_count, 0) as login_count,
			COALESCE(stats.failed_attempts, 0) as failed_attempts
		FROM users u
		LEFT JOIN (
			SELECT 
				user_id,
				MAX(created_at) as last_login_at,
				COUNT(*) as login_count,
				SUM(CASE WHEN login_status = 'failed' THEN 1 ELSE 0 END) as failed_attempts
			FROM login_logs 
			WHERE user_type = 'user' AND user_id IS NOT NULL
			GROUP BY user_id
		) stats ON u.id = stats.user_id
		WHERE %s
		ORDER BY %s
		LIMIT $2 OFFSET $3
	`, whereCondition, orderBy)

	rows, err := r.GetDB().QueryContext(ctx, dataQuery, searchPattern, pagination.GetLimit(), pagination.GetOffset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var users []*UserManagementModel
	for rows.Next() {
		userModel := &UserManagementModel{User: &user.User{}}
		err := rows.Scan(
			&userModel.ID, &userModel.Email, &userModel.Name, &userModel.Status,
			&userModel.EmailVerified, &userModel.EmailVerifiedAt,
			&userModel.CreatedAt, &userModel.UpdatedAt, &userModel.DeletedAt,
			&userModel.LastLoginAt, &userModel.LoginCount, &userModel.FailedAttempts,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan search result: %w", err)
		}
		users = append(users, userModel)
	}

	return users, total, nil
}

// === 统计查询方法 ===

// GetUserStatistics 获取用户统计信息
func (r *Repository) GetUserStatistics(ctx context.Context) (*UserStatistics, error) {
	stats := &UserStatistics{
		GeneratedAt:   time.Now(),
		UsersByStatus: make(map[user.UserStatus]int64),
	}

	// 基础统计查询
	basicStatsQuery := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN email_verified = true THEN 1 END) as verified_users,
			COUNT(CASE WHEN email_verified = false THEN 1 END) as unverified_users,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_users,
			COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive_users,
			COUNT(CASE WHEN status = 'suspended' THEN 1 END) as suspended_users,
			COUNT(CASE WHEN deleted_at IS NOT NULL THEN 1 END) as deleted_users,
			COUNT(CASE WHEN DATE(created_at) = CURRENT_DATE THEN 1 END) as new_users_today,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '7 days' THEN 1 END) as new_users_this_week,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as new_users_this_month
		FROM users
	`

	err := r.GetDB().QueryRowContext(ctx, basicStatsQuery).Scan(
		&stats.TotalUsers, &stats.VerifiedUsers, &stats.UnverifiedUsers,
		&stats.ActiveUsers, &stats.InactiveUsers, &stats.SuspendedUsers, &stats.DeletedUsers,
		&stats.NewUsersToday, &stats.NewUsersThisWeek, &stats.NewUsersThisMonth,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic statistics: %w", err)
	}

	// 计算验证率
	if stats.TotalUsers > 0 {
		stats.VerificationRate = float64(stats.VerifiedUsers) / float64(stats.TotalUsers) * 100
	}

	// 按状态统计
	stats.UsersByStatus[user.UserStatusActive] = stats.ActiveUsers
	stats.UsersByStatus[user.UserStatusInactive] = stats.InactiveUsers
	stats.UsersByStatus[user.UserStatusSuspended] = stats.SuspendedUsers

	// 活跃用户统计（基于登录日志）
	activityStatsQuery := `
		SELECT 
			COUNT(DISTINCT CASE WHEN DATE(created_at) = CURRENT_DATE THEN user_id END) as active_users_today,
			COUNT(DISTINCT CASE WHEN created_at >= CURRENT_DATE - INTERVAL '7 days' THEN user_id END) as active_users_this_week,
			COUNT(DISTINCT CASE WHEN created_at >= CURRENT_DATE - INTERVAL '30 days' THEN user_id END) as active_users_this_month
		FROM login_logs 
		WHERE user_type = 'user' AND login_status = 'success' AND user_id IS NOT NULL
	`

	err = r.GetDB().QueryRowContext(ctx, activityStatsQuery).Scan(
		&stats.ActiveUsersToday, &stats.ActiveUsersThisWeek, &stats.ActiveUsersThisMonth,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity statistics: %w", err)
	}

	// 获取注册趋势（最近30天）
	registrationTrendQuery := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM users 
		WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
		AND deleted_at IS NULL
		GROUP BY DATE(created_at)
		ORDER BY date DESC
		LIMIT 30
	`

	rows, err := r.GetDB().QueryContext(ctx, registrationTrendQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get registration trend: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var daily DailyCount
		var date time.Time
		err := rows.Scan(&date, &daily.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registration trend: %w", err)
		}
		daily.Date = date.Format("2006-01-02")
		stats.RegistrationTrend = append(stats.RegistrationTrend, daily)
	}

	// 获取验证趋势（最近30天）
	verificationTrendQuery := `
		SELECT DATE(email_verified_at) as date, COUNT(*) as count
		FROM users 
		WHERE email_verified_at >= CURRENT_DATE - INTERVAL '30 days'
		AND email_verified = true
		AND deleted_at IS NULL
		GROUP BY DATE(email_verified_at)
		ORDER BY date DESC
		LIMIT 30
	`

	rows, err = r.GetDB().QueryContext(ctx, verificationTrendQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get verification trend: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var daily DailyCount
		var date time.Time
		err := rows.Scan(&date, &daily.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan verification trend: %w", err)
		}
		daily.Date = date.Format("2006-01-02")
		stats.VerificationTrend = append(stats.VerificationTrend, daily)
	}

	return stats, nil
}

// === 用户管理操作方法 ===

// UpdateUserStatus 更新用户状态
func (r *Repository) UpdateUserStatus(ctx context.Context, userID string, status user.UserStatus, reason *string) error {
	tx, err := r.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 更新用户状态
	updateQuery := `
		UPDATE users 
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := tx.ExecContext(ctx, updateQuery, status.String(), userID)
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

	// 如果是暂停状态，记录暂停信息
	if status == user.UserStatusSuspended && reason != nil {
		suspendQuery := `
			UPDATE users 
			SET suspended_at = NOW(), suspended_reason = $1
			WHERE id = $2
		`
		_, err = tx.ExecContext(ctx, suspendQuery, reason, userID)
		if err != nil {
			return fmt.Errorf("failed to update suspension info: %w", err)
		}
	}

	return tx.Commit()
}

// GetUserSessions 获取用户会话
func (r *Repository) GetUserSessions(ctx context.Context, userID string) ([]SessionSummary, error) {
	query := `
		SELECT id, session_id, ip_address, user_agent, last_activity, created_at, expires_at
		FROM user_sessions 
		WHERE user_id = $1 AND is_active = true AND expires_at > NOW()
		ORDER BY last_activity DESC
	`

	rows, err := r.GetDB().QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []SessionSummary
	for rows.Next() {
		var session SessionSummary
		err := rows.Scan(
			&session.ID, &session.ID, &session.IPAddress, &session.UserAgent,
			&session.LastActivity, &session.CreatedAt, &session.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// DeactivateUserSessions 停用用户会话
func (r *Repository) DeactivateUserSessions(ctx context.Context, userID string, sessionID *string) error {
	var query string
	var args []interface{}

	if sessionID != nil {
		// 停用指定会话
		query = `
			UPDATE user_sessions 
			SET is_active = false, last_activity = NOW()
			WHERE user_id = $1 AND session_id = $2
		`
		args = []interface{}{userID, *sessionID}
	} else {
		// 停用所有会话
		query = `
			UPDATE user_sessions 
			SET is_active = false, last_activity = NOW()
			WHERE user_id = $1
		`
		args = []interface{}{userID}
	}

	result, err := r.GetDB().ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to deactivate sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no sessions found or deactivated")
	}

	return nil
}

// CreateManagementLog 创建管理操作日志
func (r *Repository) CreateManagementLog(ctx context.Context, log *UserManagementLog) error {
	log.ID = uuid.New().String()

	query := `
		INSERT INTO login_logs (id, user_id, email, user_type, login_status, failure_reason, ip_address, user_agent, session_id, created_at)
		VALUES ($1, $2, $3, 'admin_operation', $4, $5, $6, $7, $8, NOW())
	`

	// 将管理操作转换为日志格式
	reason := fmt.Sprintf("Admin operation: %s", log.Action.String())
	if log.Reason != nil {
		reason += " - " + *log.Reason
	}

	_, err := r.GetDB().ExecContext(ctx, query,
		log.ID, log.TargetUserID, log.TargetEmail, "admin_operation",
		reason, log.IPAddress, log.UserAgent, log.AdminID,
	)

	if err != nil {
		return fmt.Errorf("failed to create management log: %w", err)
	}

	return nil
}

// GetUserActivityLogs 获取用户活动日志
func (r *Repository) GetUserActivityLogs(ctx context.Context, userID string, limit int) ([]UserActivityLogSummary, error) {
	query := `
		SELECT id, login_status as action, ip_address, created_at, 
		       CASE WHEN login_status = 'success' THEN 'success' ELSE 'failed' END as status
		FROM login_logs 
		WHERE user_id = $1 AND user_type = 'user'
		ORDER BY created_at DESC 
		LIMIT $2
	`

	rows, err := r.GetDB().QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity logs: %w", err)
	}
	defer rows.Close()

	var logs []UserActivityLogSummary
	for rows.Next() {
		var log UserActivityLogSummary
		err := rows.Scan(&log.ID, &log.Action, &log.IPAddress, &log.CreatedAt, &log.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}
