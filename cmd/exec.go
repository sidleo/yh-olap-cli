package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/sidleo/yh-olap-cli/internal/api"
	"github.com/sidleo/yh-olap-cli/internal/auth"
	"github.com/sidleo/yh-olap-cli/internal/engine"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "SQL 执行（异步）",
}

var execRunCmd = &cobra.Command{
	Use:   "run [SQL]",
	Short: "执行 SQL 语句（异步）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sql := args[0]
		engineName, _ := cmd.Flags().GetString("engine")
		username, _ := cmd.Flags().GetString("user")
		wait, _ := cmd.Flags().GetBool("wait")

		eng, err := engine.GetEngine(engineName)
		if err != nil {
			return err
		}

		client, err := getClient(username)
		if err != nil {
			return err
		}

		fmt.Println("执行 SQL...")
		requestID, err := client.RunSql(sql, eng)
		if err != nil {
			return fmt.Errorf("执行 SQL 失败: %w", err)
		}

		fmt.Printf("Request ID: %s\n", requestID)

		if wait {
			fmt.Println("等待执行完成...")
			_, _, err = client.WaitForFinish(requestID, 500*time.Millisecond, 30*time.Minute)
			if err != nil {
				return fmt.Errorf("等待执行完成失败: %w", err)
			}
			fmt.Println("执行完成")
		}

		return nil
	},
}

var execFileCmd = &cobra.Command{
	Use:   "file [SQL_FILE]",
	Short: "从文件执行 SQL（异步）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("读取文件失败: %w", err)
		}
		sql := string(data)

		engineName, _ := cmd.Flags().GetString("engine")
		username, _ := cmd.Flags().GetString("user")
		wait, _ := cmd.Flags().GetBool("wait")

		eng, err := engine.GetEngine(engineName)
		if err != nil {
			return err
		}

		client, err := getClient(username)
		if err != nil {
			return err
		}

		fmt.Printf("读取文件: %s\n", filePath)
		fmt.Println("执行 SQL...")
		requestID, err := client.RunSql(sql, eng)
		if err != nil {
			return fmt.Errorf("执行 SQL 失败: %w", err)
		}

		fmt.Printf("Request ID: %s\n", requestID)

		if wait {
			fmt.Println("等待执行完成...")
			_, _, err = client.WaitForFinish(requestID, 500*time.Millisecond, 30*time.Minute)
			if err != nil {
				return fmt.Errorf("等待执行完成失败: %w", err)
			}
			fmt.Println("执行完成")
		}

		return nil
	},
}

func init() {
	execRunCmd.Flags().StringP("engine", "e", "impala", "查询引擎 (impala/hive/clickhouse)")
	execRunCmd.Flags().StringP("user", "u", "", "用户名")
	execRunCmd.Flags().Bool("wait", true, "等待执行完成")

	execFileCmd.Flags().StringP("engine", "e", "impala", "查询引擎 (impala/hive/clickhouse)")
	execFileCmd.Flags().StringP("user", "u", "", "用户名")
	execFileCmd.Flags().Bool("wait", true, "等待执行完成")

	execCmd.AddCommand(execRunCmd)
	execCmd.AddCommand(execFileCmd)
}

func getClient(username string) (*api.Client, error) {
	config, err := auth.Load()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	if username == "" {
		username = config.GetDefaultUser()
	}

	if username == "" {
		return nil, fmt.Errorf("未指定用户名且没有默认用户，请先登录: yh-olap-cli login login -u <用户名> -p <密码>")
	}

	_, password, otpKey, err := config.GetUser(username)
	if err != nil {
		return nil, fmt.Errorf("获取用户凭据失败: %w", err)
	}

	// 登录获取 token
	result, err := auth.Login(username, password, otpKey)
	if err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}

	client := api.NewClient("JSESSIONID=" + result.JSessionID)
	return client, nil
}
