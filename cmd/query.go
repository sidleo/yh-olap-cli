package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/sidleo/yh-olap-cli/internal/engine"
	"github.com/sidleo/yh-olap-cli/internal/output"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "执行 SQL 并获取结果（同步）",
}

var queryRunCmd = &cobra.Command{
	Use:   "run [SQL]",
	Short: "执行 SQL 并获取结果",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sql := args[0]
		engineName, _ := cmd.Flags().GetString("engine")
		format, _ := cmd.Flags().GetString("format")
		limit, _ := cmd.Flags().GetInt("limit")
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

		fmt.Println("获取结果...")
		result, err := client.GetSqlResult(requestID, 200, 1)
		if err != nil {
			return fmt.Errorf("获取结果失败: %w", err)
		}

		data := map[string]interface{}{
			"columnNameList": toInterfaceSlice(result.ColumnNameList),
			"list":           toInterfaceSlice(result.List),
			"total":          result.Total,
		}

		switch format {
		case "json":
			output.PrintJSON(data)
		case "csv":
			output.PrintCSV(data)
		default:
			output.PrintTable(data, limit)
		}

		return nil
	},
}

var queryFileCmd = &cobra.Command{
	Use:   "file [SQL_FILE]",
	Short: "从文件执行 SQL 并获取结果",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("读取文件失败: %w", err)
		}
		sql := string(data)

		engineName, _ := cmd.Flags().GetString("engine")
		format, _ := cmd.Flags().GetString("format")
		limit, _ := cmd.Flags().GetInt("limit")
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

		fmt.Println("获取结果...")
		result, err := client.GetSqlResult(requestID, 200, 1)
		if err != nil {
			return fmt.Errorf("获取结果失败: %w", err)
		}

		resultData := map[string]interface{}{
			"columnNameList": toInterfaceSlice(result.ColumnNameList),
			"list":           toInterfaceSlice(result.List),
			"total":          result.Total,
		}

		switch format {
		case "json":
			output.PrintJSON(resultData)
		case "csv":
			output.PrintCSV(resultData)
		default:
			output.PrintTable(resultData, limit)
		}

		return nil
	},
}

func init() {
	queryRunCmd.Flags().StringP("engine", "e", "impala", "查询引擎 (impala/hive/clickhouse)")
	queryRunCmd.Flags().StringP("format", "f", "table", "输出格式 (table/json/csv)")
	queryRunCmd.Flags().IntP("limit", "l", 20, "显示行数限制")
	queryRunCmd.Flags().StringP("user", "u", "", "用户名")

	queryFileCmd.Flags().StringP("engine", "e", "impala", "查询引擎 (impala/hive/clickhouse)")
	queryFileCmd.Flags().StringP("format", "f", "table", "输出格式 (table/json/csv)")
	queryFileCmd.Flags().IntP("limit", "l", 20, "显示行数限制")
	queryFileCmd.Flags().StringP("user", "u", "", "用户名")

	queryCmd.AddCommand(queryRunCmd)
	queryCmd.AddCommand(queryFileCmd)
}

// toInterfaceSlice 将具体类型的 slice 转换为 []interface{}
func toInterfaceSlice(v interface{}) []interface{} {
	switch val := v.(type) {
	case []string:
		result := make([]interface{}, len(val))
		for i, s := range val {
			result[i] = s
		}
		return result
	case []map[string]interface{}:
		result := make([]interface{}, len(val))
		for i, m := range val {
			result[i] = m
		}
		return result
	default:
		return nil
	}
}
