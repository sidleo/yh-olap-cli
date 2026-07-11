package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Config 用户配置管理，兼容 Python 版本的 ~/.yh_olap/config.json 格式
type Config struct {
	Users      map[string]UserConfig `json:"users"`
	DefaultUser string                `json:"default_user"`
}

// UserConfig 单个用户的凭据
type UserConfig struct {
	Password string `json:"password"` // base64 编码
	OtpKey   string `json:"otp_key"`  // base64 编码
	SavedAt  string `json:"saved_at"`
}

// configDir 返回配置目录路径
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}
	return filepath.Join(home, ".yh_olap"), nil
}

// configFile 返回配置文件路径
func configFile() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// EnsureConfigDir 确保配置目录存在
func EnsureConfigDir() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	// Windows 上 chmod 会被忽略，不影响功能
	if runtime.GOOS != "windows" {
		os.Chmod(dir, 0700)
	}
	return nil
}

// Load 加载配置文件
func Load() (*Config, error) {
	path, err := configFile()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Users: make(map[string]UserConfig)}, nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	if config.Users == nil {
		config.Users = make(map[string]UserConfig)
	}
	return &config, nil
}

// Save 保存配置文件
func (c *Config) Save() error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	path, err := configFile()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}

// SaveUser 保存用户凭据
func (c *Config) SaveUser(username, password, otpKey string, setDefault bool) {
	if c.Users == nil {
		c.Users = make(map[string]UserConfig)
	}
	c.Users[username] = UserConfig{
		Password: base64Encode(password),
		OtpKey:   base64Encode(otpKey),
		SavedAt:  time.Now().Format(time.RFC3339),
	}
	if setDefault || c.DefaultUser == "" {
		c.DefaultUser = username
	}
}

// GetUser 获取用户凭据
func (c *Config) GetUser(username string) (string, string, string, error) {
	user, ok := c.Users[username]
	if !ok {
		return "", "", "", fmt.Errorf("用户 %s 的凭据不存在", username)
	}
	password, err := base64Decode(user.Password)
	if err != nil {
		return "", "", "", fmt.Errorf("解码密码失败: %w", err)
	}
	otpKey, err := base64Decode(user.OtpKey)
	if err != nil {
		return "", "", "", fmt.Errorf("解码 OTP 密钥失败: %w", err)
	}
	return username, password, otpKey, nil
}

// GetDefaultUser 获取默认用户名
func (c *Config) GetDefaultUser() string {
	return c.DefaultUser
}

// ListUsers 列出所有保存的用户
func (c *Config) ListUsers() []string {
	users := make([]string, 0, len(c.Users))
	for u := range c.Users {
		users = append(users, u)
	}
	return users
}

// RemoveUser 删除用户凭据
func (c *Config) RemoveUser(username string) {
	delete(c.Users, username)
	if c.DefaultUser == username {
		c.DefaultUser = ""
	}
}

// ClearAll 清除所有用户凭据
func (c *Config) ClearAll() {
	c.Users = make(map[string]UserConfig)
	c.DefaultUser = ""
}

func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func base64Decode(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
