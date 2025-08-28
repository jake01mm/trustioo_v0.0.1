package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"trusioo_api_v0.0.1/internal/config"
	"trusioo_api_v0.0.1/internal/infrastructure/database"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// JWTManager JWT管理器
type JWTManager struct {
	config *config.JWTConfig
	db     *database.Database
	logger *logrus.Logger
}

// Claims JWT声明结构
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	UserType string `json:"user_type"` // admin, user
	jwt.RegisteredClaims
}

// TokenPair 令牌对结构
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// RefreshToken 刷新令牌模型（集成到JWT管理器中）
type RefreshToken struct {
	ID         string                  `json:"id" db:"id"`
	TokenID    string                  `json:"token_id" db:"token_id"` // JWT jti claim
	UserID     string                  `json:"user_id" db:"user_id"`
	UserType   string                  `json:"user_type" db:"user_type"` // admin, user
	DeviceInfo *map[string]interface{} `json:"device_info" db:"device_info"`
	IPAddress  *string                 `json:"ip_address" db:"ip_address"`
	IsRevoked  bool                    `json:"is_revoked" db:"is_revoked"`
	ExpiresAt  time.Time               `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time               `json:"updated_at" db:"updated_at"`
}

// IsExpired 检查令牌是否过期
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid 检查令牌是否有效
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsRevoked && !rt.IsExpired()
}

// Revoke 撤销令牌
func (rt *RefreshToken) Revoke() {
	rt.IsRevoked = true
	rt.UpdatedAt = time.Now()
}

// NewJWTManager 创建新的JWT管理器
func NewJWTManager(cfg *config.JWTConfig, db *database.Database, logger *logrus.Logger) *JWTManager {
	return &JWTManager{
		config: cfg,
		db:     db,
		logger: logger,
	}
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func (j *JWTManager) GenerateTokenPair(userID, email, role, userType string) (*TokenPair, error) {
	return j.GenerateTokenPairWithContext(context.Background(), userID, email, role, userType, nil, nil)
}

// GenerateTokenPairWithContext 生成令牌对（带上下文和设备信息）
func (j *JWTManager) GenerateTokenPairWithContext(ctx context.Context, userID, email, role, userType string, deviceInfo *map[string]interface{}, ipAddress *string) (*TokenPair, error) {
	// 生成访问令牌
	accessToken, err := j.GenerateAccessToken(userID, email, role, userType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 生成刷新令牌并保存到数据库
	refreshTokenString, _, err := j.GenerateAndStoreRefreshToken(ctx, userID, userType, deviceInfo, ipAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate and store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(j.config.ExpireDuration.Seconds()),
	}, nil
}

// GenerateAccessToken 生成访问令牌
func (j *JWTManager) GenerateAccessToken(userID, email, role, userType string) (string, error) {
	now := time.Now()

	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   userID,
			Issuer:    "trusioo_api",
			Audience:  []string{"trusioo_app"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.ExpireDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.Secret))
}

// GenerateRefreshToken 生成刷新令牌
func (j *JWTManager) GenerateRefreshToken(userID, userType string) (string, error) {
	now := time.Now()

	claims := &Claims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   userID,
			Issuer:    "trusioo_api",
			Audience:  []string{"trusioo_refresh"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.RefreshExpireDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.Secret))
}

// ValidateToken 验证令牌
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ValidateRefreshToken 验证刷新令牌
func (j *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// 检查是否为刷新令牌
	for _, aud := range claims.Audience {
		if aud == "trusioo_refresh" {
			return claims, nil
		}
	}

	return nil, errors.New("not a refresh token")
}

// RefreshTokenPair 使用刷新令牌生成新的令牌对
func (j *JWTManager) RefreshTokenPair(refreshToken string, email, role string) (*TokenPair, error) {
	// 验证刷新令牌
	claims, err := j.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 生成新的令牌对
	return j.GenerateTokenPair(claims.UserID, email, role, claims.UserType)
}

// ExtractTokenFromHeader 从请求头中提取令牌
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	// 检查Bearer前缀
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("invalid authorization header format")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", errors.New("token is required")
	}

	return token, nil
}

// IsTokenExpired 检查令牌是否过期
func (j *JWTManager) IsTokenExpired(claims *Claims) bool {
	return time.Now().After(claims.ExpiresAt.Time)
}

// GetTokenRemainingTime 获取令牌剩余时间
func (j *JWTManager) GetTokenRemainingTime(claims *Claims) time.Duration {
	return time.Until(claims.ExpiresAt.Time)
}

// ========== RefreshToken 数据库管理方法 ==========

// GenerateAndStoreRefreshToken 生成并存储刷新令牌
func (j *JWTManager) GenerateAndStoreRefreshToken(ctx context.Context, userID, userType string, deviceInfo *map[string]interface{}, ipAddress *string) (string, string, error) {
	now := time.Now()
	tokenID := uuid.New().String()

	claims := &Claims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Subject:   userID,
			Issuer:    "trusioo_api",
			Audience:  []string{"trusioo_refresh"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.RefreshExpireDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.config.Secret))
	if err != nil {
		return "", "", err
	}

	// 存储到数据库
	refreshToken := &RefreshToken{
		ID:         uuid.New().String(),
		TokenID:    tokenID,
		UserID:     userID,
		UserType:   userType,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		IsRevoked:  false,
		ExpiresAt:  claims.ExpiresAt.Time,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := j.storeRefreshToken(ctx, refreshToken); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return tokenString, tokenID, nil
}

// ValidateAndGetRefreshToken 验证并获取刷新令牌
func (j *JWTManager) ValidateAndGetRefreshToken(ctx context.Context, tokenString string) (*RefreshToken, *Claims, error) {
	// 验证JWT的签名和结构
	claims, err := j.ValidateRefreshToken(tokenString)
	if err != nil {
		return nil, nil, err
	}

	// 从数据库获取刷新令牌记录
	refreshToken, err := j.getRefreshTokenByTokenID(ctx, claims.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("refresh token not found in database: %w", err)
	}

	// 检查令牌是否有效
	if !refreshToken.IsValid() {
		if refreshToken.IsRevoked {
			return nil, nil, errors.New("refresh token has been revoked")
		}
		if refreshToken.IsExpired() {
			return nil, nil, errors.New("refresh token has expired")
		}
	}

	return refreshToken, claims, nil
}

// RefreshTokenPairWithRotation 刷新令牌轮换（撤销旧令牌并创建新令牌）
func (j *JWTManager) RefreshTokenPairWithRotation(ctx context.Context, refreshTokenString, email, role string) (*TokenPair, error) {
	// 验证旧刷新令牌
	oldToken, claims, err := j.ValidateAndGetRefreshToken(ctx, refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 撤销旧令牌
	if err := j.revokeRefreshToken(ctx, oldToken.TokenID); err != nil {
		return nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	// 生成新的令牌对
	newTokenPair, err := j.GenerateTokenPairWithContext(ctx, claims.UserID, email, role, claims.UserType, oldToken.DeviceInfo, oldToken.IPAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token pair: %w", err)
	}

	j.logger.WithFields(logrus.Fields{
		"old_token_id": oldToken.TokenID,
		"user_id":      claims.UserID,
	}).Info("Refresh token rotation completed")

	return newTokenPair, nil
}

// RevokeRefreshToken 撤销刷新令牌
func (j *JWTManager) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	return j.revokeRefreshToken(ctx, tokenID)
}

// RevokeAllUserRefreshTokens 撤销用户所有刷新令牌
func (j *JWTManager) RevokeAllUserRefreshTokens(ctx context.Context, userID, userType string) error {
	query := `
		UPDATE refresh_tokens 
		SET is_revoked = true, updated_at = NOW() 
		WHERE user_id = $1 AND user_type = $2 AND is_revoked = false
	`

	_, err := j.db.ExecContext(ctx, query, userID, userType)
	if err != nil {
		j.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":   userID,
			"user_type": userType,
		}).Error("Failed to revoke all user refresh tokens")
		return fmt.Errorf("failed to revoke all user refresh tokens: %w", err)
	}

	j.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"user_type": userType,
	}).Info("All user refresh tokens revoked")

	return nil
}

// CleanupExpiredRefreshTokens 清理过期的刷新令牌
func (j *JWTManager) CleanupExpiredRefreshTokens(ctx context.Context) (int64, error) {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`

	result, err := j.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired refresh tokens: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	if deleted > 0 {
		j.logger.WithField("deleted_count", deleted).Info("Expired refresh tokens cleaned up")
	}

	return deleted, nil
}

// ========== 私有辅助方法 ==========

// storeRefreshToken 存储刷新令牌到数据库
func (j *JWTManager) storeRefreshToken(ctx context.Context, token *RefreshToken) error {
	// 序列化设备信息
	var deviceInfoJSON interface{}
	if token.DeviceInfo != nil {
		data, err := json.Marshal(token.DeviceInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal device info: %w", err)
		}
		deviceInfoJSON = string(data)
	} else {
		deviceInfoJSON = nil
	}

	query := `
		INSERT INTO refresh_tokens (
			id, token_id, user_id, user_type, device_info, ip_address,
			is_revoked, expires_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := j.db.ExecContext(ctx, query,
		token.ID, token.TokenID, token.UserID, token.UserType,
		deviceInfoJSON, token.IPAddress, token.IsRevoked,
		token.ExpiresAt, token.CreatedAt, token.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

// getRefreshTokenByTokenID 根据TokenID获取刷新令牌
func (j *JWTManager) getRefreshTokenByTokenID(ctx context.Context, tokenID string) (*RefreshToken, error) {
	query := `
		SELECT id, token_id, user_id, user_type, device_info, ip_address,
		       is_revoked, expires_at, created_at, updated_at
		FROM refresh_tokens
		WHERE token_id = $1
	`

	token := &RefreshToken{}
	var deviceInfoJSON sql.NullString

	err := j.db.QueryRowContext(ctx, query, tokenID).Scan(
		&token.ID, &token.TokenID, &token.UserID, &token.UserType,
		&deviceInfoJSON, &token.IPAddress, &token.IsRevoked,
		&token.ExpiresAt, &token.CreatedAt, &token.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("refresh token not found")
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	// 反序列化设备信息
	if deviceInfoJSON.Valid {
		var deviceInfo map[string]interface{}
		if err := json.Unmarshal([]byte(deviceInfoJSON.String), &deviceInfo); err == nil {
			token.DeviceInfo = &deviceInfo
		}
	}

	return token, nil
}

// revokeRefreshToken 撤销刷新令牌
func (j *JWTManager) revokeRefreshToken(ctx context.Context, tokenID string) error {
	query := `
		UPDATE refresh_tokens 
		SET is_revoked = true, updated_at = NOW() 
		WHERE token_id = $1 AND is_revoked = false
	`

	result, err := j.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("refresh token not found or already revoked")
	}

	j.logger.WithField("token_id", tokenID).Info("Refresh token revoked")
	return nil
}
