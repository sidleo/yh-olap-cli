package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// PrintTable 以表格形式打印结果
func PrintTable(data map[string]interface{}, limit int) {
	columns, ok := data["columnNameList"].([]interface{})
	if !ok {
		fmt.Println("无数据")
		return
	}

	list, ok := data["list"].([]interface{})
	if !ok {
		fmt.Println("无数据")
		return
	}

	// 转换列名
	colNames := make([]string, len(columns))
	for i, col := range columns {
		colNames[i] = fmt.Sprintf("%v", col)
	}

	// 限制行数
	rows := list
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}

	// 计算每列最大宽度
	widths := make([]int, len(colNames))
	for i, name := range colNames {
		widths[i] = len(name)
	}
	for _, row := range rows {
		if rowMap, ok := row.(map[string]interface{}); ok {
			for i, name := range colNames {
				val := fmt.Sprintf("%v", rowMap[name])
				if len(val) > widths[i] {
					widths[i] = len(val)
				}
			}
		}
	}

	// 打印表头
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	header := make([]string, len(colNames))
	for i, name := range colNames {
		header[i] = name
	}
	fmt.Fprintln(w, strings.Join(header, "\t"))
	fmt.Fprintln(w, strings.Repeat("-", 80))

	// 打印数据行
	for _, row := range rows {
		if rowMap, ok := row.(map[string]interface{}); ok {
			values := make([]string, len(colNames))
			for i, name := range colNames {
				val := rowMap[name]
				if val == nil {
					values[i] = "NULL"
				} else {
					values[i] = fmt.Sprintf("%v", val)
				}
			}
			fmt.Fprintln(w, strings.Join(values, "\t"))
		}
	}
	w.Flush()

	// 打印统计信息
	totalRows := len(list)
	fmt.Printf("\n共 %d 行", totalRows)
	if limit > 0 && totalRows > limit {
		fmt.Printf("（显示前 %d 行）", limit)
	}
	fmt.Println()
}

// PrintJSON 以 JSON 形式打印结果
func PrintJSON(data map[string]interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON 序列化失败: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// PrintCSV 以 CSV 形式打印结果
func PrintCSV(data map[string]interface{}) {
	columns, ok := data["columnNameList"].([]interface{})
	if !ok {
		return
	}

	list, ok := data["list"].([]interface{})
	if !ok {
		return
	}

	// 打印表头
	colNames := make([]string, len(columns))
	for i, col := range columns {
		colNames[i] = escapeCSV(fmt.Sprintf("%v", col))
	}
	fmt.Println(strings.Join(colNames, ","))

	// 打印数据行
	for _, row := range list {
		if rowMap, ok := row.(map[string]interface{}); ok {
			values := make([]string, len(colNames))
			for i, name := range columns {
				nameStr := fmt.Sprintf("%v", name)
				val := rowMap[nameStr]
				if val == nil {
					values[i] = ""
				} else {
					values[i] = escapeCSV(fmt.Sprintf("%v", val))
				}
			}
			fmt.Println(strings.Join(values, ","))
		}
	}
}

func escapeCSV(s string) string {
	if strings.ContainsAny(s, ",\"\n\r") {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(s, "\"", "\"\""))
	}
	return s
}
