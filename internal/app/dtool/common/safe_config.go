package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	// SafePasswordVersionSalt 密码版本计算盐值
	SafePasswordVersionSalt = "dtool_safe_login_v1"
)

// SafeConfig Safe 配置结构体
type SafeConfig struct {
	Password        string
	PasswordVersion string
}

// BuildSafePasswordVersion 根据密码生成版本标识
// 使用 md5(password + AppName + fixed_salt) 计算
func BuildSafePasswordVersion(password string, appName string) string {
	if appName == "" {
		appName = "dtool"
	}

	data := fmt.Sprintf("%s%s%s", password, appName, SafePasswordVersionSalt)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// getTodayEndOfDay 返回今天结束的时间戳（23:59:59）
func getTodayEndOfDay() int64 {
	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	return endOfDay.Unix()
}
