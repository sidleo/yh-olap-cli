package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sidleo/yh-olap-cli/internal/engine"
)

// DownloadSimple 快速下载 SQL 结果（最多 1000 条）
func (c *Client) DownloadSimple(requestID, savePath string) (string, error) {
	if savePath == "" {
		savePath = requestID + ".xlsx"
	}

	data, err := c.DoGetBytes("/download/olapResultSimple/" + requestID)
	if err != nil {
		return "", fmt.Errorf("下载失败: %w", err)
	}

	dir := filepath.Dir(savePath)
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("创建目录失败: %w", err)
		}
	}

	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}

	return savePath, nil
}

// DownloadOlapResult 下载 SQL 结果为 Excel 文件（Hive 引擎）
func (c *Client) DownloadOlapResult(requestID, savePath string) (string, error) {
	if savePath == "" {
		savePath = requestID + ".xlsx"
	}

	data, err := c.DoGetBytes("/download/olapResult/" + requestID)
	if err != nil {
		return "", fmt.Errorf("下载失败: %w", err)
	}

	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}

	return savePath, nil
}

// DownloadToExcel 下载 SQL 结果为 Excel 文件（Impala 引擎）
func (c *Client) DownloadToExcel(requestID, fileName, savePath string) (string, error) {
	if savePath == "" {
		savePath = requestID + ".xlsx"
	}

	url := fmt.Sprintf("https://prokongbigdata.yonghui.cn/yh-magpie-bridge-manager/open/api/downloadToExcel?requestId=%s&fileName=%s", requestID, fileName)
	data, err := c.DoGetBytesFromURL(url)
	if err != nil {
		return "", fmt.Errorf("下载失败: %w", err)
	}

	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}

	return savePath, nil
}

// RefreshDownload 刷新下载任务
func (c *Client) RefreshDownload(downloadID int) error {
	url := fmt.Sprintf("/download/refresh?downloadId=%d", downloadID)
	data := map[string]int{"downloadId": downloadID}
	_, err := c.DoPost(url, data)
	return err
}

// ApprovalDetail 审批任务详情
type ApprovalDetail struct {
	TaskState         int    `json:"taskState"`
	TaskStateName     string `json:"taskStateName"`
	Engine            int    `json:"engine"`
	RequestID         string `json:"requestId"`
	DownLoadRequestId string `json:"downLoadRequestId"`
}

// CreateSkipDownloadOrder 创建全量下载订单（最多 500000 条）
func (c *Client) CreateSkipDownloadOrder(requestID string, eng engine.Engine) (int, error) {
	data := map[string]interface{}{
		"requestId": requestID,
		"engine":    eng.Engine,
	}

	resp, err := c.DoPut("/approval/createSkipDownloadOrder", data)
	if err != nil {
		return 0, fmt.Errorf("创建下载订单失败: %w", err)
	}

	rawData, err := resp.GetData()
	if err != nil {
		return 0, err
	}

	var result struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(rawData, &result); err != nil {
		return 0, err
	}

	return result.ID, nil
}

// CreateMiddleDownloadOrder 创建中等下载订单（最多 50000 条）
func (c *Client) CreateMiddleDownloadOrder(requestID string, eng engine.Engine) (int, error) {
	data := map[string]interface{}{
		"requestId": requestID,
		"engine":    eng.Engine,
	}

	resp, err := c.DoPut("/approval/createMiddleDownloadOrder", data)
	if err != nil {
		return 0, fmt.Errorf("创建下载订单失败: %w", err)
	}

	rawData, err := resp.GetData()
	if err != nil {
		return 0, err
	}

	var result struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(rawData, &result); err != nil {
		return 0, err
	}

	return result.ID, nil
}

// GetApprovalDetail 获取审批任务详情
func (c *Client) GetApprovalDetail(approvalID int) (*ApprovalDetail, error) {
	url := fmt.Sprintf("/approval/detail?approvalId=%d", approvalID)
	resp, err := c.DoGet(url)
	if err != nil {
		return nil, err
	}

	rawData, err := resp.GetData()
	if err != nil {
		return nil, err
	}

	var detail ApprovalDetail
	if err := json.Unmarshal(rawData, &detail); err != nil {
		return nil, err
	}

	return &detail, nil
}

// WaitForDownload 等待下载任务完成
func (c *Client) WaitForDownload(approvalID int, interval time.Duration) (*ApprovalDetail, error) {
	for {
		detail, err := c.GetApprovalDetail(approvalID)
		if err != nil {
			return nil, err
		}

		switch detail.TaskState {
		case 1: // 数据生成中
			time.Sleep(interval)
		case 2: // 数据已生成
			time.Sleep(500 * time.Millisecond) // 避免服务器文件生成延迟
			return detail, nil
		case 3: // 失败，刷新重试
			c.RefreshDownload(approvalID)
			time.Sleep(interval)
		default:
			time.Sleep(interval)
		}
	}
}

// DownloadExcel 智能下载 Excel 文件
func (c *Client) DownloadExcel(requestID, savePath string, eng engine.Engine) (string, error) {
	// 获取结果总数
	result, err := c.GetSqlResult(requestID, 1, 1)
	if err != nil {
		return "", fmt.Errorf("获取结果信息失败: %w", err)
	}

	total := result.Total
	if total <= 0 {
		total = 100000
	}

	if total <= 1000 {
		// 小结果集直接下载
		return c.DownloadSimple(requestID, savePath)
	}

	// 大结果集走审批流程
	approvalID, err := c.CreateSkipDownloadOrder(requestID, eng)
	if err != nil {
		return "", err
	}

	detail, err := c.WaitForDownload(approvalID, 5*time.Second)
	if err != nil {
		return "", err
	}

	// 根据引擎选择下载端点
	if detail.Engine == 1 {
		return c.DownloadOlapResult(detail.RequestID, savePath)
	} else if detail.Engine == 2 {
		return c.DownloadToExcel(detail.DownLoadRequestId, detail.DownLoadRequestId, savePath)
	}

	return "", fmt.Errorf("未知引擎类型: %d", detail.Engine)
}
