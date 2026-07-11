package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sidleo/yh-olap-cli/internal/auth"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录管理",
}

var loginLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "登录 OLAP 系统",
	RunE: func(cmd *cobra.Command, args []string) error {
		username, _ := cmd.Flags().GetString("user")
		password, _ := cmd.Flags().GetString("password")
		otpKey, _ := cmd.Flags().GetString("otp")
		save, _ := cmd.Flags().GetBool("save")
		setDefault, _ := cmd.Flags().GetBool("default")

		config, err := auth.Load()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 如果未提供用户名，尝试使用默认用户
		if username == "" {
			defaultUser := config.GetDefaultUser()
			if defaultUser != "" {
				username, password, otpKey, err = config.GetUser(defaultUser)
				if err != nil {
					return fmt.Errorf("获取默认用户凭据失败: %w", err)
				}
				fmt.Printf("使用默认用户: %s\n", username)
			} else {
				return fmt.Errorf("请提供用户名 (-u)")
			}
		}

		// 执行登录
		result, err := auth.Login(username, password, otpKey)
		if err != nil {
			return fmt.Errorf("登录失败: %w", err)
		}

		fmt.Println("登录成功！")
		fmt.Printf("JSessionID: %s\n", result.JSessionID)

		// 保存凭据
		if save {
			config.SaveUser(username, password, otpKey, setDefault)
			if err := config.Save(); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}
			fmt.Println("凭据已保存")
			if setDefault {
				fmt.Printf("已设为默认用户: %s\n", username)
			}
		}

		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout [username]",
	Short: "删除保存的用户凭据",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		config, err := auth.Load()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		if all {
			config.ClearAll()
			fmt.Println("已清除所有保存的凭据")
		} else if len(args) > 0 {
			config.RemoveUser(args[0])
			fmt.Printf("已清除用户 %s 的凭据\n", args[0])
		} else {
			return fmt.Errorf("请指定用户名或使用 --all")
		}

		return config.Save()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有保存的用户",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := auth.Load()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		users := config.ListUsers()
		defaultUser := config.GetDefaultUser()

		if len(users) == 0 {
			fmt.Println("没有保存的用户")
			return nil
		}

		fmt.Println("已保存的用户:")
		for _, user := range users {
			if user == defaultUser {
				fmt.Printf("  - %s (默认)\n", user)
			} else {
				fmt.Printf("  - %s\n", user)
			}
		}

		return nil
	},
}

func init() {
	loginLoginCmd.Flags().StringP("user", "u", "", "用户名")
	loginLoginCmd.Flags().StringP("password", "p", "", "密码")
	loginLoginCmd.Flags().StringP("otp", "o", "", "OTP 密钥")
	loginLoginCmd.Flags().Bool("save", true, "是否保存凭据")
	loginLoginCmd.Flags().BoolP("default", "d", false, "设为默认用户")

	loginCmd.AddCommand(loginLoginCmd)

	logoutCmd.Flags().BoolP("all", "a", false, "清除所有用户")
}
