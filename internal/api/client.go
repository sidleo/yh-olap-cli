package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL   = "http://prokong.bigdata.yonghui.cn/yh-olap-web"
	OrgCode   = "bgzt000004"
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0"
)

// Client OLAP API 客户端
type Client struct {
	Token       string
	HTTPClient  *http.Client
	Timeout     time.Duration
}

// NewClient 创建新的 API 客户端
func NewClient(token string) *Client {
	return &Client{
		Token: token,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: 30 * time.Second,
		},
		Timeout: 30 * time.Second,
	}
}

// SetToken 设置 token
func (c *Client) SetToken(token string) {
	c.Token = token
}

// APIResponse API 统一响应格式
type APIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// GetData 获取响应数据，失败时返回错误
func (r *APIResponse) GetData() (json.RawMessage, error) {
	if !r.Success {
		return nil, fmt.Errorf("API 错误: %s", r.Message)
	}
	return r.Data, nil
}

// DoPost 发送 POST 请求
func (c *Client) DoPost(path string, body interface{}) (*APIResponse, error) {
	url := BaseURL + path
	return c.doRequest("POST", url, body)
}

// DoPut 发送 PUT 请求
func (c *Client) DoPut(path string, body interface{}) (*APIResponse, error) {
	url := BaseURL + path
	return c.doRequest("PUT", url, body)
}

// DoGet 发送 GET 请求
func (c *Client) DoGet(path string) (*APIResponse, error) {
	url := BaseURL + path
	return c.doRequest("GET", url, nil)
}

// DoGetBytes 发送 GET 请求并返回原始字节（用于文件下载）
func (c *Client) DoGetBytes(path string) ([]byte, error) {
	url := BaseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// DoGetBytesFromURL 从指定 URL 获取原始字节
func (c *Client) DoGetBytesFromURL(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) doRequest(method, url string, body interface{}) (*APIResponse, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w (响应: %s)", err, string(respBody))
	}

	return &apiResp, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("token", c.Token)
	req.Header.Set("orgCode", OrgCode)
}

// GetSqlResultData 获取 SQL 结果数据
type GetSqlResultData struct {
	RequestID  string `json:"requestId"`
	PageSize   int    `json:"pageSize"`
	PageNo     int    `json:"pageNo"`
}

// RunSqlData 执行 SQL 请求数据
type RunSqlData struct {
	Engine         string                 `json:"engine"`
	DsID           int                    `json:"dsId"`
	SQL            string                 `json:"sql"`
	Params         []interface{}          `json:"params"`
	ExecuteConfigs map[string]interface{} `json:"executeConfigs"`
}

// RunSqlResponse 执行 SQL 响应数据
type RunSqlResponse struct {
	ExecuteID string `json:"executeId"`
}

// GetSqlResultResponse 获取 SQL 结果响应
type GetSqlResultResponse struct {
	ColumnNameList []string                 `json:"columnNameList"`
	List           []map[string]interface{} `json:"list"`
	Total          int                      `json:"total"`
	IsReady        string                   `json:"isReady"`
}

// GetLogResultResponse 获取日志结果响应
type GetLogResultResponse struct {
	Finish string      `json:"finish"`
	Error  interface{} `json:"error"`
	Data   string      `json:"data"`
}

// DownloadSimpleResponse 快速下载响应（Excel 文件）
// 直接返回文件内容，不需要解析 JSON
