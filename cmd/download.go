package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/sidleo/yh-olap-cli/internal/engine"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "下载查询结果为 Excel",
}

var downloadResultCmd = &cobra.Command{
	Use:   "result [REQUEST_ID]",
	Short: "下载已有 SQL 执行结果为 Excel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		requestID := args[0]
		outputPath, _ := cmd.Flags().GetString("output")
		engineName, _ := cmd.Flags().GetString("engine")
		username, _ := cmd.Flags().GetString("user")

		eng, err := engine.GetEngine(engineName)
		if err != nil {
			return err
		}

		client, err := getClient(username)
		if err != nil {
			return err
		}

		if outputPath == "" {
			outputPath = requestID + ".xlsx"
		}

		fmt.Println("下载结果...")
		savedPath, err := client.DownloadExcel(requestID, outputPath, eng)
		if err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}

		fmt.Printf("已保存到: %s\n", savedPath)
		return nil
	},
}

var downloadQueryCmd = &cobra.Command{
	Use:   "query [SQL]",
	Short: "执行 SQL 并下载结果",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sql := args[0]
		outputPath, _ := cmd.Flags().GetString("output")
		engineName, _ := cmd.Flags().GetString("engine")
		username, _ := cmd.Flags().GetString("user")

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
		fmt.Println("等待执行完成...")

		_, _, err = client.WaitForFinish(requestID, 500*time.Millisecond, 30*time.Minute)
		if err != nil {
			return fmt.Errorf("等待执行完成失败: %w", err)
		}

		if outputPath == "" {
			outputPath = requestID + ".xlsx"
		}

		fmt.Println("下载结果...")
		savedPath, err := client.DownloadExcel(requestID, outputPath, eng)
		if err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}

		fmt.Printf("已保存到: %s\n", savedPath)
		return nil
	},
}

var downloadFileCmd = &cobra.Command{
	Use:   "file [SQL_FILE]",
	Short: "从文件执行 SQL 并下载结果",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("读取文件失败: %w", err)
		}
		sql := string(data)

		outputPath, _ := cmd.Flags().GetString("output")
		engineName, _ := cmd.Flags().GetString("engine")
		username, _ := cmd.Flags().GetString("user")

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
		fmt.Println("等待执行完成...")

		_, _, err = client.WaitForFinish(requestID, 500*time.Millisecond, 30*time.Minute)
		if err != nil {
			return fmt.Errorf("等待执行完成失败: %w", err)
		}

		if outputPath == "" {
			outputPath = requestID + ".xlsx"
		}

		fmt.Println("下载结果...")
		savedPath, err := client.DownloadExcel(requestID, outputPath, eng)
		if err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}

		fmt.Printf("已保存到: %s\n", savedPath)
		return nil
	},
}

func init() {
	downloadResultCmd.Flags().StringP("output", "o", "", "输出文件路径")
	downloadResultCmd.Flags().StringP("engine", "e", "impala", "查询引擎 (impala/hive/clickhouse)")
	downloadResultCmd.Flags().StringP("user", "u", "", "用户名")

	downloadQueryCmd.Flags().StringP("output", "o", "", "输出文件路径")
	downloadQueryCmd.Flags().StringP("engine", "e", "impala", "查询引擎 (impala/hive/clickhouse)")
	downloadQueryCmd.Flags().StringP("user", "u", "", "用户名")

	downloadFileCmd.Flags().StringP("output", "o", "", "输出文件路径")
	downloadFileCmd.Flags().StringP("engine", "e", "impala", "查询引擎 (impala/hive/clickhouse)")
	downloadFileCmd.Flags().StringP("user", "u", "", "用户名")

	downloadCmd.AddCommand(downloadResultCmd)
	downloadCmd.AddCommand(downloadQueryCmd)
	downloadCmd.AddCommand(downloadFileCmd)
}
