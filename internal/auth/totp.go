package auth

import (
	"fmt"
	"time"

	"github.com/pquerna/otp/totp"
)

// GenerateTOTP 根据密钥生成当前时间的 TOTP 验证码
func GenerateTOTP(otpKey string) (string, error) {
	code, err := totp.GenerateCode(otpKey, timeNow())
	if err != nil {
		return "", fmt.Errorf("生成 TOTP 验证码失败: %w", err)
	}
	return code, nil
}

// 集中 time.Now() 方便测试时 mock
var timeNow = func() time.Time {
	return time.Now()
}
