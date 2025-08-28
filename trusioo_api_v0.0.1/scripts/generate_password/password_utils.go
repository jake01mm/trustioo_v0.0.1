package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

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
