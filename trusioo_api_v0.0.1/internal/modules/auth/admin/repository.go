package admin

import (
	"context"
	"database/sql"
	"fmt"

	"trusioo_api_v0.0.1/internal/infrastructure/database"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Repository 管理员仓储
type Repository struct {
	*database.BaseRepository
	logger *logrus.Logger
}

// NewRepository 创建新的管理员仓储
func NewRepository(db *database.Database, logger *logrus.Logger) *Repository {
	return &Repository{
		BaseRepository: database.NewBaseRepository(db, logger),
		logger:         logger,
	}
}

// Create 创建管理员
func (r *Repository) Create(ctx context.Context, admin *Admin) error {
	// 生成UUID
	admin.ID = uuid.New().String()

	query := `
		INSERT INTO admins (id, email, name, password, role, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`

	_, err := r.GetDB().ExecContext(ctx, query,
		admin.ID, admin.Email, admin.Name, admin.Password, admin.Role, admin.Active)

	if err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}

	return nil
}

// GetByID 根据ID获取管理员
func (r *Repository) GetByID(ctx context.Context, id string) (*Admin, error) {
	query := `
		SELECT id, email, name, password, role, active, created_at, updated_at
		FROM admins
		WHERE id = $1 AND deleted_at IS NULL
	`

	admin := &Admin{}
	err := r.GetDB().QueryRowContext(ctx, query, id).Scan(
		&admin.ID, &admin.Email, &admin.Name, &admin.Password, &admin.Role, &admin.Active, &admin.CreatedAt, &admin.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admin not found")
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}

	return admin, nil
}

// GetByEmail 根据邮箱获取管理员
func (r *Repository) GetByEmail(ctx context.Context, email string) (*Admin, error) {
	query := `
		SELECT id, email, name, password, role, active, created_at, updated_at
		FROM admins
		WHERE email = $1 AND deleted_at IS NULL
	`

	admin := &Admin{}
	err := r.GetDB().QueryRowContext(ctx, query, email).Scan(
		&admin.ID, &admin.Email, &admin.Name, &admin.Password, &admin.Role, &admin.Active, &admin.CreatedAt, &admin.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admin not found")
		}
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}

	return admin, nil
}

// UpdatePassword 更新管理员密码
func (r *Repository) UpdatePassword(ctx context.Context, adminID, hashedPassword string) error {
	query := `
		UPDATE admins
		SET password = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.GetDB().ExecContext(ctx, query, hashedPassword, adminID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("admin not found or no changes made")
	}

	return nil
}

// UpdateStatus 更新管理员状态
func (r *Repository) UpdateStatus(ctx context.Context, adminID string, active bool) error {
	query := `
		UPDATE admins
		SET active = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.GetDB().ExecContext(ctx, query, active, adminID)
	if err != nil {
		return fmt.Errorf("failed to update admin status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("admin not found or no changes made")
	}

	return nil
}

// List 获取管理员列表
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*Admin, error) {
	query := `
		SELECT id, email, name, role, active
		FROM admins
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.GetDB().QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list admins: %w", err)
	}
	defer rows.Close()

	var admins []*Admin
	for rows.Next() {
		admin := &Admin{}
		err := rows.Scan(&admin.ID, &admin.Email, &admin.Name, &admin.Role, &admin.Active)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin: %w", err)
		}
		admins = append(admins, admin)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return admins, nil
}

// Delete 软删除管理员
func (r *Repository) Delete(ctx context.Context, adminID string) error {
	query := `
		UPDATE admins
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.GetDB().ExecContext(ctx, query, adminID)
	if err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("admin not found or already deleted")
	}

	return nil
}

// Count 统计管理员数量
func (r *Repository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM admins WHERE deleted_at IS NULL`

	var count int64
	err := r.GetDB().QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count admins: %w", err)
	}

	return count, nil
}

// ExistsByEmail 检查邮箱是否已存在
func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM admins WHERE email = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.GetDB().QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// CreatePasswordReset 创建密码重置记录
func (r *Repository) CreatePasswordReset(ctx context.Context, reset *PasswordReset) error {
	// 生成UUID
	reset.ID = uuid.New().String()

	query := `
		INSERT INTO password_resets (id, email, user_type, token, ip_address, used, used_at, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`

	_, err := r.GetDB().ExecContext(ctx, query,
		reset.ID, reset.Email, reset.UserType, reset.Token, reset.IPAddress, reset.Used, reset.UsedAt, reset.ExpiresAt)

	if err != nil {
		return fmt.Errorf("failed to create password reset record: %w", err)
	}

	return nil
}
