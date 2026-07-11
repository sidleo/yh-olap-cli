package auth

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

const (
	OLAPServiceURL = "https://idaas-cas.yonghui.cn/cas/login?service=http://cas-prod.bigdata.yonghui.cn:7070/redirect?redirectUrl=https://bigdata.yonghui.cn/#/olap/main"
	SkipADService  = "https://oa"
	UserAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0"
)

type LoginResult struct {
	JSessionID string
	URL        string
}

func Login(username, password, otpKey string) (*LoginResult, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	jar, _ := cookiejar.New(nil)

	// 客户端1: 跟随所有重饮向（用于步骤1获取登录页URL）
	followClient := &http.Client{Transport: transport, Jar: jar}

	// 客户端2: 不跟随重饮向（用于POST认证）
	noRedirectClient := &http.Client{
		Transport: transport,
		Jar:       jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	headers := map[string]string{"User-Agent": UserAgent}
	formHeaders := map[string]string{
		"User-Agent":   UserAgent,
		"Content-Type": "application/x-www-form-urlencoded",
	}

	// 步骤1: GET 目标URL，跟随重定向获取实际登录页面
	resp, err := doReq(followClient, "GET", OLAPServiceURL, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("获取登录页面失败: %w", err)
	}
	actualLoginURL := resp.Request.URL.String()
	resp.Body.Close()

	// 提取登录基地址
	parsed, _ := url.Parse(actualLoginURL)
	loginBase := fmt.Sprintf("%s://%s%s", parsed.Scheme, parsed.Host, parsed.Path)

	// 步骤2: 带 skip_ad service 的登录URL
	loginWithService := fmt.Sprintf("%s?service=%s", loginBase, SkipADService)

	// 步骤3: GET 初始化session
	resp, err = doReq(noRedirectClient, "GET", loginWithService, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("初始化会话失败: %w", err)
	}
	resp.Body.Close()

	// 步骤4: POST 密码
	e1s1 := buildE1S1Data(username, password)
	resp, err = doReq(noRedirectClient, "POST", loginWithService, formHeaders, strings.NewReader(e1s1))
	if err != nil {
		return nil, fmt.Errorf("密码认证失败: %w", err)
	}
	resp.Body.Close()

	// 步骤5: 如果401则需要OTP
	if resp.StatusCode == 401 {
		if otpKey == "" {
			return nil, fmt.Errorf("需要 OTP 验证码但未提供 otp_key")
		}
		e1s2, err := buildE1S2Data(username, otpKey)
		if err != nil {
			return nil, fmt.Errorf("生成 TOTP 失败: %w", err)
		}
		resp, err = doReq(noRedirectClient, "POST", loginWithService, formHeaders, strings.NewReader(e1s2))
		if err != nil {
			return nil, fmt.Errorf("OTP 认证失败: %w", err)
		}
		resp.Body.Close()
	}

	// 步骤6: 检查302
	if resp.StatusCode != 302 {
		return nil, fmt.Errorf("登录失败，状态码: %d", resp.StatusCode)
	}

	// 步骤7: 关键！从302响应中提取ticket，拼回到原始OLAP service URL
	location := resp.Header.Get("Location")
	ticket := ""
	if idx := strings.Index(location, "ticket="); idx != -1 {
		ticket = location[idx+8:]
	}
	if ticket == "" {
		return nil, fmt.Errorf("未获取到 ticket")
	}

	// 用原始 service URL + ticket 发起请求
	redirectURL := fmt.Sprintf("%s&ticket=%s", OLAPServiceURL, ticket)
	resp, err = doReq(noRedirectClient, "GET", redirectURL, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("ticket 验证失败: %w", err)
	}
	resp.Body.Close()

	// 步骤8: 跟随后续重定向获取JSESSIONID
	result := &LoginResult{}
	nextURL := resp.Header.Get("Location")
	for nextURL != "" {
		resp, err = doReq(followClient, "GET", nextURL, headers, nil)
		if err != nil {
			break
		}
		// 检查URL参数中的JSESSIONID
		if u, e := url.Parse(resp.Request.URL.String()); e == nil {
			if jsid := u.Query().Get("JSESSIONID"); jsid != "" {
				result.JSessionID = jsid
			}
		}
		// 检查cookies
		for _, c := range resp.Cookies() {
			if c.Name == "JSESSIONID" {
				result.JSessionID = c.Value
			}
		}
		if resp.StatusCode == 200 {
			result.URL = resp.Request.URL.String()
			resp.Body.Close()
			break
		}
		nextURL = resp.Header.Get("Location")
		resp.Body.Close()
	}

	if result.JSessionID == "" {
		return nil, fmt.Errorf("登录失败，未获取到 JSESSIONID")
	}
	return result, nil
}

func buildE1S1Data(username, password string) string {
	d := url.Values{}
	d.Set("flag", "1")
	d.Set("username", username)
	d.Set("password", password)
	d.Set("sourceType", "1")
	d.Set("execution", "e2s1")
	d.Set("_eventId", "submit")
	return d.Encode()
}

func buildE1S2Data(username, otpKey string) (string, error) {
	token, err := GenerateTOTP(otpKey)
	if err != nil {
		return "", err
	}
	d := url.Values{}
	d.Set("flag", "1")
	d.Set("token", token)
	d.Set("username", username)
	d.Set("sourceType", "1")
	d.Set("execution", "e2s2")
	d.Set("_eventId", "submit")
	return d.Encode(), nil
}

func doReq(client *http.Client, method, url string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return client.Do(req)
}
