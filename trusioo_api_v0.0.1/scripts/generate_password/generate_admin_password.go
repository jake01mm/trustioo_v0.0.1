package main

import (
	"fmt"
	"log"
	"os"
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
