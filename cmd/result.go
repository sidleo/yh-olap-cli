package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sidleo/yh-olap-cli/internal/output"
)

var resultCmd = &cobra.Command{
	Use:   "result",
	Short: "获取 SQL 执行结果",
}

var resultGetCmd = &cobra.Command{
	Use:   "get [REQUEST_ID]",
	Short: "获取 SQL 执行结果",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		requestID := args[0]
		format, _ := cmd.Flags().GetString("format")
		limit, _ := cmd.Flags().GetInt("limit")
		username, _ := cmd.Flags().GetString("user")

		client, err := getClient(username)
		if err != nil {
			return err
		}

		fmt.Println("获取结果...")
		result, err := client.GetSqlResult(requestID, 200, 1)
		if err != nil {
			return fmt.Errorf("获取结果失败: %w", err)
		}

		data := map[string]interface{}{
			"columnNameList": result.ColumnNameList,
			"list":           result.List,
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

func init() {
	resultGetCmd.Flags().StringP("format", "f", "table", "输出格式 (table/json/csv)")
	resultGetCmd.Flags().IntP("limit", "l", 20, "显示行数限制")
	resultGetCmd.Flags().StringP("user", "u", "", "用户名")

	resultCmd.AddCommand(resultGetCmd)
}
