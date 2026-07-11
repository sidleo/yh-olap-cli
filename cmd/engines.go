package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/sidleo/yh-olap-cli/internal/engine"
)

var enginesCmd = &cobra.Command{
	Use:   "engines",
	Short: "引擎管理",
}

var enginesListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有可用的查询引擎",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("可用的查询引擎:")
		fmt.Println()
		for _, eng := range engine.Engines {
			fmt.Printf("  - %-12s (engine=%s, dsId=%d)\n", eng.Name, eng.Engine, eng.DsID)
		}
		fmt.Println()
		fmt.Println("默认引擎: impala")
	},
}

func init() {
	enginesCmd.AddCommand(enginesListCmd)
}
