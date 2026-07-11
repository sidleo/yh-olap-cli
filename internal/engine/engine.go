package engine

import "fmt"

// Engine 查询引擎定义
type Engine struct {
	Name   string
	Engine string
	DsID   int
}

var (
	// Hive Hive 引擎
	Hive = Engine{Name: "hive", Engine: "1", DsID: 1}
	// Impala Impala 引擎（默认）
	Impala = Engine{Name: "impala", Engine: "2", DsID: 2}
	// Clickhouse Clickhouse 引擎
	Clickhouse = Engine{Name: "clickhouse", Engine: "3", DsID: 14004}
)

// Engines 所有可用引擎
var Engines = []Engine{Hive, Impala, Clickhouse}

// GetEngine 根据名称获取引擎
func GetEngine(name string) (Engine, error) {
	switch name {
	case "hive":
		return Hive, nil
	case "impala", "":
		return Impala, nil
	case "clickhouse":
		return Clickhouse, nil
	default:
		return Impala, fmt.Errorf("未知引擎: %s，可用引擎: impala, hive, clickhouse", name)
	}
}
