package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 检查命令行参数
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <password> <encryption_key>\n", os.Args[0])
		fmt.Println("Example: go run generate_admin_password.go admin123 your-ultra-secret-password-encryption-key-change-this")
		os.Exit(1)
	}

	password := os.Args[1]
	encryptionKey := os.Args[2]

	// 使用 hmac-bcrypt 方法生成密码哈希
	hashedPassword, err := hashWithHMACBcrypt(password, encryptionKey)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	fmt.Printf("Original password: %s\n", password)
	fmt.Printf("Encryption key: %s\n", encryptionKey)
	fmt.Printf("HMAC signature: %s\n", signWithHMAC(password, encryptionKey))
	fmt.Printf("Final hashed password: %s\n", hashedPassword)
	fmt.Println()
	fmt.Printf("SQL for migration file:\n")
	fmt.Printf("'%s'\n", hashedPassword)
}

// hashWithHMACBcrypt 使用HMAC+bcrypt双重加密
func hashWithHMACBcrypt(password, encryptionKey string) (string, error) {
	// 第一步：使用HMAC-SHA256对密码进行签名
	hmacSigned := signWithHMAC(password, encryptionKey)

	// 第二步：对HMAC签名结果使用bcrypt加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(hmacSigned), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to bcrypt hash password: %w", err)
	}

	return string(hashedPassword), nil
}

// signWithHMAC 使用HMAC-SHA256对数据进行签名
func signWithHMAC(data, encryptionKey string) string {
	h := hmac.New(sha256.New, []byte(encryptionKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}