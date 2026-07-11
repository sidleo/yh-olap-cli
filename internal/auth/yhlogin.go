package auth

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	// OLAPServiceURL CAS 登录的 OLAP 服务 URL
	OLAPServiceURL = "https://idaas-cas.yonghui.cn/cas/login?service=http://cas-prod.bigdata.yonghui.cn:7070/redirect?redirectUrl=https://bigdata.yonghui.cn/#/olap/main"
	// SkipADService 跳过域验证的 service 参数
	SkipADService = "https://oa"
	// UserAgent 模拟浏览器 User-Agent
	UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0"
)

// LoginResult 登录结果
type LoginResult struct {
	JSessionID string
	Cookies    []*http.Cookie
	URL        string
}

// Login 执行 CAS 登录流程
func Login(username, password, otpKey string) (*LoginResult, error) {
	// 创建禁用证书验证的 HTTP 客户端（与 Python 版本一致）
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		// 不自动跟随重饮向，手动处理
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   UserAgent,
	}

	// 步骤 1: GET 目标登录 URL，获取实际的登录页面 URL
	targetURL := OLAPServiceURL
	resp, err := doGet(client, targetURL, headers)
	if err != nil {
		return nil, fmt.Errorf("获取登录页面失败: %w", err)
	}
	targetLoginURL := resp.Header.Get("Location")
	if targetLoginURL == "" {
		targetLoginURL = targetURL
	}
	resp.Body.Close()

	// 提取登录 URL（去掉 query 参数）
	loginURLParts := strings.SplitN(targetLoginURL, "?", 2)
	loginURL := loginURLParts[0]

	// 步骤 2: 构建带 service 参数的登录 URL（跳过域验证）
	serviceURL := SkipADService
	fullLoginURL := fmt.Sprintf("%s?service=%s", loginURL, serviceURL)

	// 步骤 3: GET 登录页面（初始化 session）
	resp, err = doGet(client, fullLoginURL, headers)
	if err != nil {
		return nil, fmt.Errorf("初始化登录会话失败: %w", err)
	}
	resp.Body.Close()

	// 步骤 4: POST 密码认证 (e2s1)
	e1s1Data := buildE1S1Data(username, password)
	resp, err = doPost(client, fullLoginURL, headers, e1s1Data)
	if err != nil {
		return nil, fmt.Errorf("密码认证失败: %w", err)
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	// 步骤 5: 如果需要 OTP，POST OTP 认证 (e2s2)
	if resp.StatusCode == 401 {
		if strings.Contains(string(respBody), "dynamicPassword") {
			if otpKey == "" {
				return nil, fmt.Errorf("需要 OTP 验证码但未提供 otp_key")
			}
			e1s2Data, err := buildE1S2Data(username, otpKey)
			if err != nil {
				return nil, fmt.Errorf("生成 TOTP 验证码失败: %w", err)
			}
			resp, err = doPost(client, fullLoginURL, headers, e1s2Data)
			if err != nil {
				return nil, fmt.Errorf("OTP 认证失败: %w", err)
			}
			resp.Body.Close()
		}
	}

	// 步骤 6: 检查是否 302 重定向（登录成功）
	if resp.StatusCode != 302 {
		return nil, fmt.Errorf("登录失败，状态码: %d", resp.StatusCode)
	}

	// 步骤 7: 跟随重定向获取 JSESSIONID
	result := &LoginResult{}
	cookies := make([]*http.Cookie, 0)

	// 收集所有 cookies
	for _, cookie := range resp.Cookies() {
		cookies = append(cookies, cookie)
		if cookie.Name == "JSESSIONID" {
			result.JSessionID = cookie.Value
		}
	}

	// 跟随重定向链
	redirectURL := resp.Header.Get("Location")
	for redirectURL != "" {
		resp, err = doGet(client, redirectURL, headers)
		if err != nil {
			break
		}
		for _, cookie := range resp.Cookies() {
			cookies = append(cookies, cookie)
			if cookie.Name == "JSESSIONID" {
				result.JSessionID = cookie.Value
			}
		}
		if resp.StatusCode == 200 {
			result.URL = resp.Request.URL.String()
			resp.Body.Close()
			break
		}
		redirectURL = resp.Header.Get("Location")
		resp.Body.Close()
	}

	result.Cookies = cookies

	if result.JSessionID == "" {
		return nil, fmt.Errorf("登录失败，未获取到 JSESSIONID")
	}

	return result, nil
}

func buildE1S1Data(username, password string) string {
	data := url.Values{}
	data.Set("flag", "1")
	data.Set("username", username)
	data.Set("password", password)
	data.Set("phoneNum", "")
	data.Set("captcha", "")
	data.Set("sourceType", "1")
	data.Set("execution", "e2s1")
	data.Set("_eventId", "submit")
	data.Set("geolocation", "")
	return data.Encode()
}

func buildE1S2Data(username, otpKey string) (string, error) {
	token, err := GenerateTOTP(otpKey)
	if err != nil {
		return "", err
	}
	data := url.Values{}
	data.Set("flag", "1")
	data.Set("token", token)
	data.Set("username", username)
	data.Set("password", "")
	data.Set("phoneNum", "")
	data.Set("captcha", "")
	data.Set("sourceType", "1")
	data.Set("execution", "e2s2")
	data.Set("_eventId", "submit")
	data.Set("geolocation", "")
	return data.Encode(), nil
}

func doGet(client *http.Client, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return client.Do(req)
}

func doPost(client *http.Client, url string, headers map[string]string, body string) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return client.Do(req)
}
