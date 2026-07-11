package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sidleo/yh-olap-cli/internal/engine"
)

// RunSql 执行 SQL，返回 requestId
func (c *Client) RunSql(sql string, eng engine.Engine) (string, error) {
	data := RunSqlData{
		Engine:         eng.Engine,
		DsID:           eng.DsID,
		SQL:            sql,
		Params:         []interface{}{},
		ExecuteConfigs: map[string]interface{}{},
	}

	resp, err := c.DoPost("/sql/manager/runSql", data)
	if err != nil {
		return "", fmt.Errorf("执行 SQL 失败: %w", err)
	}

	rawData, err := resp.GetData()
	if err != nil {
		return "", err
	}

	var result RunSqlResponse
	if err := json.Unmarshal(rawData, &result); err != nil {
		return "", fmt.Errorf("解析执行结果失败: %w", err)
	}

	return result.ExecuteID, nil
}

// GetLogResult 获取运行日志
func (c *Client) GetLogResult(requestID string) (*GetLogResultResponse, error) {
	data := map[string]string{"requestId": requestID}
	resp, err := c.DoPost("/sql/manager/getLogResult", data)
	if err != nil {
		return nil, err
	}

	rawData, err := resp.GetData()
	if err != nil {
		return nil, err
	}

	var result GetLogResultResponse
	if err := json.Unmarshal(rawData, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSqlResult 获取 SQL 执行结果
func (c *Client) GetSqlResult(requestID string, pageSize, pageNo int) (*GetSqlResultResponse, error) {
	data := map[string]interface{}{
		"requestId": requestID,
		"pageSize":  pageSize,
		"pageNo":    pageNo,
	}

	resp, err := c.DoPost("/sql/manager/getSqlResult", data)
	if err != nil {
		return nil, err
	}

	rawData, err := resp.GetData()
	if err != nil {
		return nil, err
	}

	var result GetSqlResultResponse
	if err := json.Unmarshal(rawData, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CheckState 检查运行状态
func (c *Client) CheckState(requestID string) (string, error) {
	data := map[string]string{"requestId": requestID}
	resp, err := c.DoPost("/sql/manager/checkState", data)
	if err != nil {
		return "", err
	}

	rawData, err := resp.GetData()
	if err != nil {
		return "", err
	}

	var result struct {
		Finish string `json:"finish"`
	}
	if err := json.Unmarshal(rawData, &result); err != nil {
		return "", err
	}

	return result.Finish, nil
}

// WaitForFinish 等待 SQL 执行完成
func (c *Client) WaitForFinish(requestID string, interval, timeout time.Duration) (*GetLogResultResponse, *GetSqlResultResponse, error) {
	var logResult *GetLogResultResponse
	var sqlResult *GetSqlResultResponse
	start := time.Now()

	// 等待日志完成
	for {
		if time.Since(start) > timeout {
			return nil, nil, fmt.Errorf("等待执行超时（%v）", timeout)
		}

		lr, err := c.GetLogResult(requestID)
		if err != nil {
			return nil, nil, err
		}
		logResult = lr

		if lr.Error {
			return nil, nil, fmt.Errorf("执行错误: %s", lr.Data)
		}

		if lr.Finish == "ok" {
			break
		}
		time.Sleep(interval)
	}

	// 等待结果就绪
	for {
		if time.Since(start) > timeout {
			return nil, nil, fmt.Errorf("等待结果超时（%v）", timeout)
		}

		sr, err := c.GetSqlResult(requestID, 200, 1)
		if err != nil {
			return nil, nil, err
		}
		sqlResult = sr

		if sr.IsReady == "ok" {
			break
		}
		time.Sleep(interval)
	}

	return logResult, sqlResult, nil
}
