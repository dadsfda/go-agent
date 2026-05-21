package app

import "golang.org/x/crypto/bcrypt"

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func verifyPassword(stored, password string) bool {
	if bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil {
		return true
	}
	// 兼容旧演示数据：早期版本曾将明文密码写入 password_hash 字段。
	return stored == password
}
