package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:   "yh-olap-cli",
	Short: "YH OLAP 命令行工具 - 用于 OLAP 数据查询与下载",
	Long: `YH OLAP CLI 是一个用于永辉 OLAP 系统的命令行工具。
支持 SQL 执行、结果查询、Excel 下载等功能。`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(resultCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(enginesCmd)
	rootCmd.AddCommand(versionCmd)
}
