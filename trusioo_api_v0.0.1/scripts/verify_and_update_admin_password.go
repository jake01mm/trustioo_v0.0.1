package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <password> [encryption_key]\n", os.Args[0])
		fmt.Println("Example: go run verify_and_update_admin_password.go admin123 your-ultra-secret-password-encryption-key-change-this")
		os.Exit(1)
	}

	password := os.Args[1]
	encryptionKey := "your-ultra-secret-password-encryption-key-change-this"
	if len(os.Args) > 2 {
		encryptionKey = os.Args[2]
	}

	// 连接数据库
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/trusioo_api?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 生成新的密码哈希
	hashedPassword, err := hashWithHMACBcrypt(password, encryptionKey)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// 更新数据库中的密码
	_, err = db.Exec("UPDATE admins SET password = $1 WHERE email = $2", hashedPassword, "admin@trusioo.com")
	if err != nil {
		log.Fatalf("Failed to update password: %v", err)
	}

	fmt.Printf("Successfully updated admin password in database\n")
	fmt.Printf("Email: admin@trusioo.com\n")
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("New hash: %s\n", hashedPassword)

	// 验证更新
	var storedHash string
	err = db.QueryRow("SELECT password FROM admins WHERE email = $1", "admin@trusioo.com").Scan(&storedHash)
	if err != nil {
		log.Fatalf("Failed to verify update: %v", err)
	}

	// 验证密码是否正确
	hmacSigned := signWithHMAC(password, encryptionKey)
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(hmacSigned))
	if err != nil {
		log.Fatalf("Password verification failed: %v", err)
	}

	fmt.Println("✅ Password verification successful!")
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