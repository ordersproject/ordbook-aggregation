package tool

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// 加盐哈希算法
func hashPassword(password string) (string, error) {
	// 生成随机的盐值
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	// 将密码和盐值拼接起来，得到原始数据
	raw := []byte(password + base64.StdEncoding.EncodeToString(salt))

	// 使用 SHA256 算法进行哈希计算
	hashed := sha256.Sum256(raw)

	// 将哈希值和盐值拼接起来，得到最终的加密密码
	encoded := base64.StdEncoding.EncodeToString(hashed[:])
	return encoded, nil
}

// 校验密码
func checkPassword(password, encoded string) bool {
	// 解码加密密码
	hashed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return false
	}

	// 获取盐值
	salt := hashed[:16]

	// 计算密码的哈希值
	raw := []byte(password + base64.StdEncoding.EncodeToString(salt))
	calculated := sha256.Sum256(raw)

	// 比较哈希值是否相同
	return base64.StdEncoding.EncodeToString(calculated[:]) == encoded
}