package main

import (
	"fmt"

	"trusioo_api_v0.0.1/pkg/cryptoutil"
)

func main() {
	// 使用与应用程序相同的配置
	encryptor := cryptoutil.NewPasswordEncryptor("your-ultra-secret-password-encryption-key-change-this", "hmac-bcrypt")

	// 数据库中的密码哈希
	storedHash := "$2a$10$w.doGJZnwj1jF4frXUnh7.2UNYtF3kQ241O6uHl3YQWs9z57UqzEK"

	// 测试密码
	password := "admin123"

	fmt.Printf("Testing password verification...\n")
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("Stored hash: %s\n", storedHash)

	err := encryptor.VerifyPassword(password, storedHash)
	if err != nil {
		fmt.Printf("Password verification FAILED: %v\n", err)
	} else {
		fmt.Printf("Password verification SUCCESS!\n")
	}

	// 也测试用相同密码生成新的哈希
	fmt.Printf("\nGenerating new hash for comparison...\n")
	newHash, err := encryptor.HashPassword(password)
	if err != nil {
		fmt.Printf("Hash generation failed: %v\n", err)
	} else {
		fmt.Printf("New hash: %s\n", newHash)

		// 验证新生成的哈希
		err = encryptor.VerifyPassword(password, newHash)
		if err != nil {
			fmt.Printf("New hash verification FAILED: %v\n", err)
		} else {
			fmt.Printf("New hash verification SUCCESS!\n")
		}
	}
}
