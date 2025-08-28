// Package crypto 提供密码加密和安全相关功能
package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// PasswordEncryptor 密码加密器
type PasswordEncryptor struct {
	encryptionKey []byte
	method        string
	saltLength    int
}

// PasswordConfig 密码配置
type PasswordConfig struct {
	EncryptionKey string `json:"encryption_key"`
	Method        string `json:"method"`
	SaltLength    int    `json:"salt_length"`
	BcryptCost    int    `json:"bcrypt_cost"`
}

// NewPasswordEncryptor 创建新的密码加密器
func NewPasswordEncryptor(key, method string) *PasswordEncryptor {
	return &PasswordEncryptor{
		encryptionKey: []byte(key),
		method:        method,
		saltLength:    32, // 默认盐长度
	}
}

// NewPasswordEncryptorWithConfig 使用配置创建密码加密器
func NewPasswordEncryptorWithConfig(config *PasswordConfig) *PasswordEncryptor {
	saltLength := config.SaltLength
	if saltLength == 0 {
		saltLength = 32
	}

	return &PasswordEncryptor{
		encryptionKey: []byte(config.EncryptionKey),
		method:        config.Method,
		saltLength:    saltLength,
	}
}

// HashPassword 加密密码
// 使用HMAC-SHA256对密码进行签名，然后再用bcrypt加密
func (pe *PasswordEncryptor) HashPassword(password string) (string, error) {
	switch pe.method {
	case "hmac-bcrypt":
		return pe.hashWithHMACBcrypt(password)
	case "bcrypt":
		return pe.hashWithBcrypt(password)
	default:
		return "", fmt.Errorf("unsupported encryption method: %s", pe.method)
	}
}

// VerifyPassword 验证密码
func (pe *PasswordEncryptor) VerifyPassword(password, hashedPassword string) error {
	switch pe.method {
	case "hmac-bcrypt":
		return pe.verifyHMACBcrypt(password, hashedPassword)
	case "bcrypt":
		return pe.verifyBcrypt(password, hashedPassword)
	default:
		return fmt.Errorf("unsupported encryption method: %s", pe.method)
	}
}

// hashWithHMACBcrypt 使用HMAC+bcrypt双重加密
func (pe *PasswordEncryptor) hashWithHMACBcrypt(password string) (string, error) {
	// 第一步：使用HMAC-SHA256对密码进行签名
	hmacSigned := pe.signWithHMAC(password)

	// 第二步：对HMAC签名结果使用bcrypt加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(hmacSigned), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to bcrypt hash password: %w", err)
	}

	return string(hashedPassword), nil
}

// verifyHMACBcrypt 验证HMAC+bcrypt双重加密的密码
func (pe *PasswordEncryptor) verifyHMACBcrypt(password, hashedPassword string) error {
	// 第一步：使用HMAC-SHA256对输入密码进行签名
	hmacSigned := pe.signWithHMAC(password)

	// 第二步：使用bcrypt验证HMAC签名结果
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(hmacSigned))
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}

	return nil
}

// hashWithBcrypt 仅使用bcrypt加密（向后兼容）
func (pe *PasswordEncryptor) hashWithBcrypt(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to bcrypt hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// verifyBcrypt 验证bcrypt加密的密码
func (pe *PasswordEncryptor) verifyBcrypt(password, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("password verification failed: %w", err)
	}
	return nil
}

// signWithHMAC 使用HMAC-SHA256对数据进行签名
func (pe *PasswordEncryptor) signWithHMAC(data string) string {
	h := hmac.New(sha256.New, pe.encryptionKey)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateSalt 生成随机盐
func (pe *PasswordEncryptor) GenerateSalt() (string, error) {
	salt := make([]byte, pe.saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}
	return hex.EncodeToString(salt), nil
}

// ValidatePasswordStrength 验证密码强度
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= 33 && char <= 47 || char >= 58 && char <= 64 || char >= 91 && char <= 96 || char >= 123 && char <= 126:
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// GenerateRandomPassword 生成随机密码
func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 12 // 默认长度
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)

	_, err := rand.Read(password)
	if err != nil {
		return "", fmt.Errorf("failed to generate random password: %w", err)
	}

	for i := range password {
		password[i] = charset[int(password[i])%len(charset)]
	}

	return string(password), nil
}

// PasswordManager 密码管理器
type PasswordManager struct {
	encryptor       *PasswordEncryptor
	maxAttempts     int
	lockoutDuration time.Duration
	attempts        map[string]int
	lockouts        map[string]time.Time
}

// NewPasswordManager 创建密码管理器
func NewPasswordManager(encryptor *PasswordEncryptor) *PasswordManager {
	return &PasswordManager{
		encryptor:       encryptor,
		maxAttempts:     5,                // 最大尝试次数
		lockoutDuration: 15 * time.Minute, // 锁定时间
		attempts:        make(map[string]int),
		lockouts:        make(map[string]time.Time),
	}
}

// VerifyPasswordWithLockout 验证密码（包含锁定机制）
func (pm *PasswordManager) VerifyPasswordWithLockout(identifier, password, hashedPassword string) error {
	// 检查是否被锁定
	if lockTime, exists := pm.lockouts[identifier]; exists {
		if time.Since(lockTime) < pm.lockoutDuration {
			return fmt.Errorf("account is locked due to too many failed attempts")
		}
		// 锁定时间已过，清除锁定记录
		delete(pm.lockouts, identifier)
		delete(pm.attempts, identifier)
	}

	// 验证密码
	err := pm.encryptor.VerifyPassword(password, hashedPassword)
	if err != nil {
		// 密码错误，增加尝试次数
		pm.attempts[identifier]++
		if pm.attempts[identifier] >= pm.maxAttempts {
			pm.lockouts[identifier] = time.Now()
			return fmt.Errorf("account locked due to too many failed attempts")
		}
		return fmt.Errorf("invalid password (%d/%d attempts)", pm.attempts[identifier], pm.maxAttempts)
	}

	// 密码正确，清除尝试记录
	delete(pm.attempts, identifier)
	return nil
}

// IsValidMethod 检查加密方法是否有效
func IsValidMethod(method string) bool {
	validMethods := map[string]bool{
		"hmac-bcrypt": true,
		"bcrypt":      true,
		// 可以在这里添加更多支持的方法
	}
	return validMethods[method]
}

// GetSupportedMethods 获取支持的加密方法列表
func GetSupportedMethods() []string {
	return []string{"hmac-bcrypt", "bcrypt"}
}

// DefaultPasswordConfig 返回默认密码配置
func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		Method:     "hmac-bcrypt",
		SaltLength: 32,
		BcryptCost: bcrypt.DefaultCost,
	}
}
