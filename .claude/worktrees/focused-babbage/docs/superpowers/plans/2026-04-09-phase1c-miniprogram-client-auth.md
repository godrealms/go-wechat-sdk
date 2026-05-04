# Phase 1C: mini-program Client, Auth, User, wxacode, scheme, urllink, shortlink

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bootstrap the mini-program package with client setup, token management via POST, and the first API groups: auth (jscode2session), user (phone/unionid), QR code generation, scheme/urllink/shortlink.

**Architecture:** mini-program.Client embeds *core.BaseClient with TokenMethod="POST" against /cgi-bin/stable_token. New PostBinary helper added to utils/http.go for binary image responses. All API files use c.Ctx and c.TokenQuery().

**Tech Stack:** Go 1.23.1, standard library only. Depends on Plan A (core/ package).

**PREREQUISITE:** `2026-04-09-phase1a-core-package.md` must be completed first.

---

### Task 1: utils/http.go — add PostBinary method

**Files:**
- Modify: `utils/http.go`

- [ ] **Step 1: Add `PostBinary` method to the `HTTP` struct**

  Open `utils/http.go`. After the existing `GetWithHeaders` method (currently ending around line 247), append the following method. The file already imports `bytes`, `io`, `encoding/json`, `net/http`, `context`, and `fmt` — no new imports are needed.

  ```go
  // PostBinary sends a POST request and returns raw response bytes (for binary responses like images).
  func (h *HTTP) PostBinary(ctx context.Context, path string, body interface{}) ([]byte, error) {
  	jsonBody, err := json.Marshal(body)
  	if err != nil {
  		return nil, fmt.Errorf("marshal request body failed: %w", err)
  	}
  	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.BaseURL+path, bytes.NewReader(jsonBody))
  	if err != nil {
  		return nil, fmt.Errorf("create request failed: %w", err)
  	}
  	req.Header.Set("Content-Type", "application/json")
  	resp, err := h.Client.Do(req)
  	if err != nil {
  		return nil, fmt.Errorf("send request failed: %w", err)
  	}
  	defer resp.Body.Close()
  	respBody, err := io.ReadAll(io.LimitReader(resp.Body, DefaultMaxResponseSize))
  	if err != nil {
  		return nil, fmt.Errorf("read response body failed: %w", err)
  	}
  	return respBody, nil
  }
  ```

  The complete updated `utils/http.go` after this addition will be:

  ```go
  package utils

  import (
  	"bytes"
  	"context"
  	"encoding/json"
  	"fmt"
  	"io"
  	"net/http"
  	"net/url"
  	"strings"
  	"time"
  )

  const (
  	// DefaultMaxResponseSize 默认最大响应体大小 (10MB)
  	DefaultMaxResponseSize = 10 * 1024 * 1024
  )

  // Logger 日志接口
  type Logger interface {
  	Printf(format string, v ...interface{})
  }

  // HTTP 客户端结构体
  type HTTP struct {
  	BaseURL string
  	Client  *http.Client
  	Headers map[string]string
  	Timeout time.Duration
  	Logger  Logger
  }

  // Option 定义配置选项的函数类型
  type Option func(*HTTP)

  // NewHTTP 创建新的 HTTP 客户端
  func NewHTTP(baseURL string, opts ...Option) *HTTP {
  	h := &HTTP{
  		BaseURL: baseURL,
  		Headers: make(map[string]string),
  		Timeout: 30 * time.Second,
  	}

  	// 应用所有选项
  	for _, opt := range opts {
  		opt(h)
  	}

  	// 初始化 http.Client
  	h.Client = &http.Client{
  		Timeout: h.Timeout,
  	}

  	return h
  }

  // WithTimeout 设置超时时间的选项
  func WithTimeout(timeout time.Duration) Option {
  	return func(h *HTTP) {
  		h.Timeout = timeout
  	}
  }

  // WithHeaders 设置请求头的选项
  func WithHeaders(headers map[string]string) Option {
  	return func(h *HTTP) {
  		for k, v := range headers {
  			h.Headers[k] = v
  		}
  	}
  }

  // WithLogger 设置日志记录器
  func WithLogger(logger Logger) Option {
  	return func(h *HTTP) {
  		h.Logger = logger
  	}
  }

  // SetBaseURL 设置基础URL
  func (h *HTTP) SetBaseURL(url string) {
  	h.BaseURL = url
  }

  // do 执行 HTTP 请求的通用方法
  func (h *HTTP) do(ctx context.Context, method, path string, body interface{}, query url.Values, result interface{}) error {
  	if query != nil {
  		path += "?" + query.Encode()
  	}
  	// 构建完整URL
  	fullURL := h.BaseURL + path

  	if h.Logger != nil {
  		h.Logger.Printf("method: %s", method)
  		h.Logger.Printf("url: %s", fullURL)
  	}

  	// 处理请求体
  	var bodyReader io.Reader
  	if body != nil {
  		jsonBody, err := json.Marshal(body)
  		if err != nil {
  			return fmt.Errorf("marshal request body failed: %w", err)
  		}
  		bodyReader = bytes.NewReader(jsonBody)
  		if h.Logger != nil {
  			h.Logger.Printf("body: %s", string(jsonBody))
  		}
  	}

  	// 创建请求
  	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
  	if err != nil {
  		return fmt.Errorf("create request failed: %w", err)
  	}

  	// 设置默认请求头
  	if body != nil {
  		req.Header.Set("Content-Type", "application/json")
  	}

  	// 添加自定义请求头
  	for k, v := range h.Headers {
  		req.Header.Set(k, v)
  	}

  	// 发送请求
  	resp, err := h.Client.Do(req)
  	if err != nil {
  		return fmt.Errorf("send request failed: %w", err)
  	}
  	defer resp.Body.Close()

  	// 读取响应体（限制最大大小防止OOM）
  	respBody, err := io.ReadAll(io.LimitReader(resp.Body, DefaultMaxResponseSize))
  	if err != nil {
  		return fmt.Errorf("read response body failed: %w", err)
  	}

  	if h.Logger != nil {
  		h.Logger.Printf("response: %s", string(respBody))
  	}
  	// 检查响应状态码
  	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
  		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBody))
  	}

  	// 如果需要解析响应结果
  	if result != nil && len(respBody) > 0 {
  		if err = json.Unmarshal(respBody, result); err != nil {
  			return fmt.Errorf("unmarshal response body failed: %w: %s", err, string(respBody))
  		}
  	}

  	return nil
  }

  // Get 发送 GET 请求
  func (h *HTTP) Get(ctx context.Context, path string, query url.Values, result interface{}) error {
  	return h.do(ctx, http.MethodGet, path, nil, query, result)
  }

  // Post 发送 POST 请求
  func (h *HTTP) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
  	return h.do(ctx, http.MethodPost, path, body, nil, result)
  }

  // Put 发送 PUT 请求
  func (h *HTTP) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
  	return h.do(ctx, http.MethodPut, path, body, nil, result)
  }

  // Patch 发送 PATCH 请求
  func (h *HTTP) Patch(ctx context.Context, path string, body interface{}, result interface{}) error {
  	return h.do(ctx, http.MethodPatch, path, body, nil, result)
  }

  // Delete 发送 DELETE 请求
  func (h *HTTP) Delete(ctx context.Context, path string, result interface{}) error {
  	return h.do(ctx, http.MethodDelete, path, nil, nil, result)
  }

  // doWithHeaders 执行 HTTP 请求，使用请求级别的 headers（线程安全）
  func (h *HTTP) doWithHeaders(ctx context.Context, method, path string, body interface{}, reqHeaders map[string]string, query url.Values, result interface{}) error {
  	if query != nil {
  		path += "?" + query.Encode()
  	}
  	fullURL := h.BaseURL + path

  	var bodyReader io.Reader
  	if body != nil {
  		jsonBody, err := json.Marshal(body)
  		if err != nil {
  			return fmt.Errorf("marshal request body failed: %w", err)
  		}
  		bodyReader = bytes.NewReader(jsonBody)
  	}

  	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
  	if err != nil {
  		return fmt.Errorf("create request failed: %w", err)
  	}

  	if body != nil {
  		req.Header.Set("Content-Type", "application/json")
  	}

  	for k, v := range h.Headers {
  		req.Header.Set(k, v)
  	}
  	for k, v := range reqHeaders {
  		req.Header.Set(k, v)
  	}

  	resp, err := h.Client.Do(req)
  	if err != nil {
  		return fmt.Errorf("send request failed: %w", err)
  	}
  	defer resp.Body.Close()

  	respBody, err := io.ReadAll(io.LimitReader(resp.Body, DefaultMaxResponseSize))
  	if err != nil {
  		return fmt.Errorf("read response body failed: %w", err)
  	}

  	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
  		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBody))
  	}

  	if result != nil && len(respBody) > 0 {
  		if err = json.Unmarshal(respBody, result); err != nil {
  			return fmt.Errorf("unmarshal response body failed: %w: %s", err, string(respBody))
  		}
  	}
  	return nil
  }

  // PostWithHeaders 发送 POST 请求，使用请求级别的 headers（线程安全）
  func (h *HTTP) PostWithHeaders(ctx context.Context, path string, body interface{}, headers map[string]string, result interface{}) error {
  	return h.doWithHeaders(ctx, http.MethodPost, path, body, headers, nil, result)
  }

  // GetWithHeaders 发送 GET 请求，使用请求级别的 headers（线程安全）
  func (h *HTTP) GetWithHeaders(ctx context.Context, path string, headers map[string]string, query url.Values, result interface{}) error {
  	return h.doWithHeaders(ctx, http.MethodGet, path, nil, headers, query, result)
  }

  // PostBinary sends a POST request and returns raw response bytes (for binary responses like images).
  func (h *HTTP) PostBinary(ctx context.Context, path string, body interface{}) ([]byte, error) {
  	jsonBody, err := json.Marshal(body)
  	if err != nil {
  		return nil, fmt.Errorf("marshal request body failed: %w", err)
  	}
  	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.BaseURL+path, bytes.NewReader(jsonBody))
  	if err != nil {
  		return nil, fmt.Errorf("create request failed: %w", err)
  	}
  	req.Header.Set("Content-Type", "application/json")
  	resp, err := h.Client.Do(req)
  	if err != nil {
  		return nil, fmt.Errorf("send request failed: %w", err)
  	}
  	defer resp.Body.Close()
  	respBody, err := io.ReadAll(io.LimitReader(resp.Body, DefaultMaxResponseSize))
  	if err != nil {
  		return nil, fmt.Errorf("read response body failed: %w", err)
  	}
  	return respBody, nil
  }

  // PostForm 发送 POST 表单请求
  func (h *HTTP) PostForm(ctx context.Context, path string, form url.Values, result interface{}) error {
  	fullURL := h.BaseURL + path

  	// 创建请求
  	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, strings.NewReader(form.Encode()))
  	if err != nil {
  		return fmt.Errorf("create request failed: %w", err)
  	}

  	// 设置表单请求头
  	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

  	// 添加自定义请求头
  	for k, v := range h.Headers {
  		req.Header.Set(k, v)
  	}

  	// 发送请求
  	resp, err := h.Client.Do(req)
  	if err != nil {
  		return fmt.Errorf("send request failed: %w", err)
  	}
  	defer resp.Body.Close()

  	// 读取响应体（限制最大大小防止OOM）
  	respBody, err := io.ReadAll(io.LimitReader(resp.Body, DefaultMaxResponseSize))
  	if err != nil {
  		return fmt.Errorf("read response body failed: %w", err)
  	}

  	// 检查响应状态码
  	if resp.StatusCode != http.StatusOK {
  		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBody))
  	}

  	// 解析响应结果
  	if result != nil && len(respBody) > 0 {
  		if err = json.Unmarshal(respBody, result); err != nil {
  			return fmt.Errorf("unmarshal response body failed: %w:%s", err, string(respBody))
  		}
  	}

  	return nil
  }
  ```

- [ ] **Step 2: Verify build**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  go build ./utils/...
  ```

  Expected: no errors.

- [ ] **Step 3: Commit**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  git add utils/http.go
  git commit -m "feat(utils): add PostBinary for raw binary responses"
  ```

---

### Task 2: mini-program/client.go + mini-program/struct.base.go

**Files:**
- Create: `mini-program/client.go`
- Create: `mini-program/struct.base.go`

- [ ] **Step 1: Create `mini-program/client.go`**

  Replace the existing stub content of `mini-program/client.go` (which currently only contains `package mini_program` and a comment) with the following:

  ```go
  package mini_program

  import (
  	"context"
  	"github.com/godrealms/go-wechat-sdk/core"
  )

  // Config holds mini-program configuration
  type Config struct {
  	core.BaseConfig
  }

  // Client is the WeChat Mini-Program client
  type Client struct {
  	*core.BaseClient
  }

  // NewClient creates a new mini-program client
  // Token refresh uses POST to /cgi-bin/stable_token (different from offiaccount's GET /cgi-bin/token)
  func NewClient(ctx context.Context, config *Config) *Client {
  	base := core.NewBaseClient(ctx, &config.BaseConfig, "https://api.weixin.qq.com", "/cgi-bin/stable_token", "POST")
  	return &Client{BaseClient: base}
  }
  ```

- [ ] **Step 2: Create `mini-program/struct.base.go`**

  ```go
  package mini_program

  import "github.com/godrealms/go-wechat-sdk/core"

  // Code2SessionResult is the result of jscode2session
  type Code2SessionResult struct {
  	core.Resp
  	OpenId     string `json:"openid"`
  	SessionKey string `json:"session_key"`
  	UnionId    string `json:"unionid"`
  }

  // Watermark contains appid and timestamp for phone number verification
  type Watermark struct {
  	AppId     string `json:"appid"`
  	Timestamp int64  `json:"timestamp"`
  }

  // PhoneInfo contains decrypted phone number details
  type PhoneInfo struct {
  	PhoneNumber     string    `json:"phoneNumber"`
  	PurePhoneNumber string    `json:"purePhoneNumber"`
  	CountryCode     string    `json:"countryCode"`
  	Watermark       Watermark `json:"watermark"`
  }

  // PhoneNumberResult is the result of GetPhoneNumber
  type PhoneNumberResult struct {
  	core.Resp
  	PhoneInfo PhoneInfo `json:"phone_info"`
  }

  // PaidUnionIdResult is the result of GetPaidUnionId
  type PaidUnionIdResult struct {
  	core.Resp
  	Unionid string `json:"unionid"`
  }
  ```

- [ ] **Step 3: Verify build**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  go build ./mini-program/...
  ```

  Expected: no errors.

- [ ] **Step 4: Commit**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  git add mini-program/client.go mini-program/struct.base.go
  git commit -m "feat(mini-program): add client, config, base structs"
  ```

---

### Task 3: mini-program/api.auth.go + mini-program/api.user.go

**Files:**
- Create: `mini-program/api.auth.go`
- Create: `mini-program/api.user.go`

- [ ] **Step 1: Create `mini-program/api.auth.go`**

  ```go
  package mini_program

  import (
  	"net/url"
  	"github.com/godrealms/go-wechat-sdk/core"
  )

  // Code2Session 登录凭证校验
  // Uses appid+secret directly, no access_token needed
  // GET /sns/jscode2session
  func (c *Client) Code2Session(jsCode string) (*Code2SessionResult, error) {
  	query := url.Values{
  		"appid":      {c.Config.AppId},
  		"secret":     {c.Config.AppSecret},
  		"js_code":    {jsCode},
  		"grant_type": {"authorization_code"},
  	}
  	result := &Code2SessionResult{}
  	err := c.Https.Get(c.Ctx, "/sns/jscode2session", query, result)
  	if err != nil {
  		return nil, err
  	}
  	return result, result.GetError()
  }

  // CheckSessionKey 检验登录态
  // GET /wxa/checksession
  func (c *Client) CheckSessionKey(openid, signature, sigMethod string) error {
  	query := c.TokenQuery(url.Values{
  		"openid":     {openid},
  		"signature":  {signature},
  		"sig_method": {sigMethod},
  	})
  	result := &core.Resp{}
  	err := c.Https.Get(c.Ctx, "/wxa/checksession", query, result)
  	if err != nil {
  		return err
  	}
  	return result.GetError()
  }

  // ResetSessionKey 重置登录态
  // GET /wxa/resetusersessionkey
  func (c *Client) ResetSessionKey(openid, signature, sigMethod string) error {
  	query := c.TokenQuery(url.Values{
  		"openid":     {openid},
  		"signature":  {signature},
  		"sig_method": {sigMethod},
  	})
  	result := &core.Resp{}
  	err := c.Https.Get(c.Ctx, "/wxa/resetusersessionkey", query, result)
  	if err != nil {
  		return err
  	}
  	return result.GetError()
  }
  ```

- [ ] **Step 2: Create `mini-program/api.user.go`**

  ```go
  package mini_program

  import (
  	"fmt"
  	"net/url"
  )

  // GetPhoneNumber 获取手机号
  // POST /wxa/business/getuserphonenumber (access_token in URL)
  func (c *Client) GetPhoneNumber(code string) (*PhoneNumberResult, error) {
  	path := fmt.Sprintf("/wxa/business/getuserphonenumber?access_token=%s", c.GetAccessToken())
  	body := map[string]string{"code": code}
  	result := &PhoneNumberResult{}
  	err := c.Https.Post(c.Ctx, path, body, result)
  	if err != nil {
  		return nil, err
  	}
  	return result, result.GetError()
  }

  // GetPaidUnionId 用户支付完成后获取 UnionId (需要先有支付记录)
  // GET /wxa/getpaidunionid
  func (c *Client) GetPaidUnionId(openid, transactionId string) (*PaidUnionIdResult, error) {
  	query := c.TokenQuery(url.Values{
  		"openid":         {openid},
  		"transaction_id": {transactionId},
  	})
  	result := &PaidUnionIdResult{}
  	err := c.Https.Get(c.Ctx, "/wxa/getpaidunionid", query, result)
  	if err != nil {
  		return nil, err
  	}
  	return result, result.GetError()
  }
  ```

- [ ] **Step 3: Verify build**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  go build ./mini-program/...
  ```

  Expected: no errors.

- [ ] **Step 4: Commit**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  git add mini-program/api.auth.go mini-program/api.user.go
  git commit -m "feat(mini-program): add auth and user APIs"
  ```

---

### Task 4: mini-program/struct.wxacode.go + api.wxacode.go + api.scheme.go + api.urllink.go + api.shortlink.go

**Files:**
- Create: `mini-program/struct.wxacode.go`
- Create: `mini-program/api.wxacode.go`
- Create: `mini-program/api.scheme.go`
- Create: `mini-program/api.urllink.go`
- Create: `mini-program/api.shortlink.go`

- [ ] **Step 1: Create `mini-program/struct.wxacode.go`**

  ```go
  package mini_program

  import "github.com/godrealms/go-wechat-sdk/core"

  // QRCodeRequest is the request for GetQRCode (limited to 100,000 codes)
  type QRCodeRequest struct {
  	Path      string         `json:"path"`
  	Width     int            `json:"width,omitempty"`
  	AutoColor bool           `json:"auto_color,omitempty"`
  	LineColor map[string]int `json:"line_color,omitempty"`
  	IsHyaline bool           `json:"is_hyaline,omitempty"`
  }

  // UnlimitedQRCodeRequest is the request for GetUnlimited (no count limit)
  type UnlimitedQRCodeRequest struct {
  	Scene       string         `json:"scene"`
  	Page        string         `json:"page,omitempty"`
  	CheckPath   bool           `json:"check_path,omitempty"`
  	EnvVersion  string         `json:"env_version,omitempty"`
  	Width       int            `json:"width,omitempty"`
  	AutoColor   bool           `json:"auto_color,omitempty"`
  	LineColor   map[string]int `json:"line_color,omitempty"`
  	IsHyaline   bool           `json:"is_hyaline,omitempty"`
  }

  // CreateQRCodeRequest is the request for CreateQRCode (QR code, not mini-program code)
  type CreateQRCodeRequest struct {
  	Path  string `json:"path"`
  	Width int    `json:"width,omitempty"`
  }

  // JumpWxa describes the mini-program page for scheme/urllink
  type JumpWxa struct {
  	Path       string `json:"path"`
  	Query      string `json:"query"`
  	EnvVersion string `json:"env_version,omitempty"`
  }

  // CloudBase describes cloud development config for urllink
  type CloudBase struct {
  	Env           string `json:"env"`
  	Domain        string `json:"domain,omitempty"`
  	Path          string `json:"path,omitempty"`
  	Query         string `json:"query,omitempty"`
  	ResourceAppid string `json:"resource_appid,omitempty"`
  }

  // GenerateSchemeRequest is the request for GenerateScheme
  type GenerateSchemeRequest struct {
  	JumpWxa        *JumpWxa `json:"jump_wxa,omitempty"`
  	IsExpire       bool     `json:"is_expire,omitempty"`
  	ExpireType     int      `json:"expire_type,omitempty"`
  	ExpireTime     int64    `json:"expire_time,omitempty"`
  	ExpireInterval int      `json:"expire_interval,omitempty"`
  	EnvVersion     string   `json:"env_version,omitempty"`
  }

  // GenerateSchemeResult is the result of GenerateScheme
  type GenerateSchemeResult struct {
  	core.Resp
  	OpenLink string `json:"openlink"`
  }

  // SchemeInfo contains scheme metadata
  type SchemeInfo struct {
  	AppId      string `json:"appid"`
  	Path       string `json:"path"`
  	Query      string `json:"query"`
  	CreateTime int64  `json:"create_time"`
  	ExpireTime int64  `json:"expire_time"`
  	EnvVersion string `json:"env_version"`
  	OpenLink   string `json:"openlink"`
  }

  // SchemeQuota contains scheme quota information
  type SchemeQuota struct {
  	LongTimeUsedQuota    int64 `json:"long_time_used_quota"`
  	LongTimeDefaultQuota int64 `json:"long_time_default_quota"`
  }

  // QuerySchemeResult is the result of QueryScheme
  type QuerySchemeResult struct {
  	core.Resp
  	SchemeInfo  *SchemeInfo  `json:"scheme_info"`
  	SchemeQuota *SchemeQuota `json:"scheme_quota"`
  }

  // GenerateUrlLinkRequest is the request for GenerateUrlLink
  type GenerateUrlLinkRequest struct {
  	Path           string     `json:"path,omitempty"`
  	Query          string     `json:"query,omitempty"`
  	IsExpire       bool       `json:"is_expire,omitempty"`
  	ExpireType     int        `json:"expire_type,omitempty"`
  	ExpireTime     int64      `json:"expire_time,omitempty"`
  	ExpireInterval int        `json:"expire_interval,omitempty"`
  	CloudBase      *CloudBase `json:"cloud_base,omitempty"`
  	EnvVersion     string     `json:"env_version,omitempty"`
  }

  // GenerateUrlLinkResult is the result of GenerateUrlLink
  type GenerateUrlLinkResult struct {
  	core.Resp
  	UrlLink string `json:"url_link"`
  }

  // UrlLinkInfo contains URL Link metadata
  type UrlLinkInfo struct {
  	AppId      string     `json:"appid"`
  	Path       string     `json:"path"`
  	Query      string     `json:"query"`
  	CreateTime int64      `json:"create_time"`
  	ExpireTime int64      `json:"expire_time"`
  	EnvVersion string     `json:"env_version"`
  	UrlLink    string     `json:"url_link"`
  	CloudBase  *CloudBase `json:"cloud_base"`
  }

  // UrlLinkQuota contains URL Link quota information
  type UrlLinkQuota struct {
  	LongTimeUsedQuota    int64 `json:"long_time_used_quota"`
  	LongTimeDefaultQuota int64 `json:"long_time_default_quota"`
  }

  // QueryUrlLinkResult is the result of QueryUrlLink
  type QueryUrlLinkResult struct {
  	core.Resp
  	UrlLinkInfo  *UrlLinkInfo  `json:"url_link_info"`
  	UrlLinkQuota *UrlLinkQuota `json:"url_link_quota"`
  }

  // GenerateShortLinkRequest is the request for GenerateShortLink
  type GenerateShortLinkRequest struct {
  	PageUrl   string `json:"page_url"`
  	PageTitle string `json:"page_title,omitempty"`
  	Permanent bool   `json:"permanent,omitempty"`
  }

  // GenerateShortLinkResult is the result of GenerateShortLink
  type GenerateShortLinkResult struct {
  	core.Resp
  	Link string `json:"link"`
  }
  ```

- [ ] **Step 2: Create `mini-program/api.wxacode.go`**

  ```go
  package mini_program

  import "fmt"

  // GetQRCode 获取小程序码 (适用于需要的码数量较少的业务场景，总共生成的码不超过10万个)
  // POST /wxa/getwxacode — returns raw PNG bytes
  func (c *Client) GetQRCode(req *QRCodeRequest) ([]byte, error) {
  	path := fmt.Sprintf("/wxa/getwxacode?access_token=%s", c.GetAccessToken())
  	return c.Https.PostBinary(c.Ctx, path, req)
  }

  // GetUnlimited 获取不限制的小程序码 (适用于需要的码数量极多或者无法预知总数量的业务场景)
  // POST /wxa/getwxacodeunlimit — returns raw PNG bytes
  func (c *Client) GetUnlimited(req *UnlimitedQRCodeRequest) ([]byte, error) {
  	path := fmt.Sprintf("/wxa/getwxacodeunlimit?access_token=%s", c.GetAccessToken())
  	return c.Https.PostBinary(c.Ctx, path, req)
  }

  // CreateQRCode 获取小程序二维码 (适用于需要的码数量较少的业务场景，生成的码永久有效)
  // POST /cgi-bin/wxaapp/createwxaqrcode — returns raw PNG bytes
  func (c *Client) CreateQRCode(req *CreateQRCodeRequest) ([]byte, error) {
  	path := fmt.Sprintf("/cgi-bin/wxaapp/createwxaqrcode?access_token=%s", c.GetAccessToken())
  	return c.Https.PostBinary(c.Ctx, path, req)
  }
  ```

- [ ] **Step 3: Create `mini-program/api.scheme.go`**

  ```go
  package mini_program

  import (
  	"fmt"
  	"net/url"
  )

  // GenerateScheme 获取小程序 scheme 码
  // POST /wxa/generatescheme
  func (c *Client) GenerateScheme(req *GenerateSchemeRequest) (*GenerateSchemeResult, error) {
  	path := fmt.Sprintf("/wxa/generatescheme?access_token=%s", c.GetAccessToken())
  	result := &GenerateSchemeResult{}
  	err := c.Https.Post(c.Ctx, path, req, result)
  	if err != nil {
  		return nil, err
  	}
  	return result, result.GetError()
  }

  // QueryScheme 查询小程序 scheme 码
  // GET /wxa/queryscheme
  func (c *Client) QueryScheme(scheme string) (*QuerySchemeResult, error) {
  	query := c.TokenQuery(url.Values{"scheme": {scheme}})
  	result := &QuerySchemeResult{}
  	err := c.Https.Get(c.Ctx, "/wxa/queryscheme", query, result)
  	if err != nil {
  		return nil, err
  	}
  	return result, result.GetError()
  }
  ```

- [ ] **Step 4: Create `mini-program/api.urllink.go`**

  ```go
  package mini_program

  import (
  	"fmt"
  	"net/url"
  )

  // GenerateUrlLink 获取小程序 URL Link
  // POST /wxa/generate_urllink
  func (c *Client) GenerateUrlLink(req *GenerateUrlLinkRequest) (*GenerateUrlLinkResult, error) {
  	path := fmt.Sprintf("/wxa/generate_urllink?access_token=%s", c.GetAccessToken())
  	result := &GenerateUrlLinkResult{}
  	err := c.Https.Post(c.Ctx, path, req, result)
  	if err != nil {
  		return nil, err
  	}
  	return result, result.GetError()
  }

  // QueryUrlLink 查询小程序 URL Link
  // GET /wxa/query_urllink
  func (c *Client) QueryUrlLink(urlLink string) (*QueryUrlLinkResult, error) {
  	query := c.TokenQuery(url.Values{"url_link": {urlLink}})
  	result := &QueryUrlLinkResult{}
  	err := c.Https.Get(c.Ctx, "/wxa/query_urllink", query, result)
  	if err != nil {
  		return nil, err
  	}
  	return result, result.GetError()
  }
  ```

- [ ] **Step 5: Create `mini-program/api.shortlink.go`**

  ```go
  package mini_program

  import "fmt"

  // GenerateShortLink 获取小程序 Short Link
  // POST /wxa/genwxashortlink
  func (c *Client) GenerateShortLink(req *GenerateShortLinkRequest) (*GenerateShortLinkResult, error) {
  	path := fmt.Sprintf("/wxa/genwxashortlink?access_token=%s", c.GetAccessToken())
  	result := &GenerateShortLinkResult{}
  	err := c.Https.Post(c.Ctx, path, req, result)
  	if err != nil {
  		return nil, err
  	}
  	return result, result.GetError()
  }
  ```

- [ ] **Step 6: Verify build**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  go build ./mini-program/...
  ```

  Expected: no errors.

- [ ] **Step 7: Commit**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  git add mini-program/struct.wxacode.go mini-program/api.wxacode.go mini-program/api.scheme.go mini-program/api.urllink.go mini-program/api.shortlink.go
  git commit -m "feat(mini-program): add wxacode, scheme, urllink, shortlink APIs"
  ```

---

### Task 5: Tests + build verification

**Files:**
- Create: `mini-program/client_test.go`

- [ ] **Step 1: Create `mini-program/client_test.go`**

  ```go
  package mini_program

  import (
  	"context"
  	"encoding/json"
  	"net/http"
  	"net/http/httptest"
  	"testing"

  	"github.com/godrealms/go-wechat-sdk/core"
  )

  func TestNewClient_UsesPostForToken(t *testing.T) {
  	var gotMethod string
  	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  		gotMethod = r.Method
  		json.NewEncoder(w).Encode(map[string]interface{}{
  			"access_token": "mini-token",
  			"expires_in":   7200,
  		})
  	}))
  	defer srv.Close()

  	cfg := &Config{BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"}}
  	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/stable_token", "POST")
  	c := &Client{BaseClient: base}

  	token := c.GetAccessToken()
  	if token != "mini-token" {
  		t.Errorf("expected mini-token, got %s", token)
  	}
  	if gotMethod != "POST" {
  		t.Errorf("expected POST token method, got %s", gotMethod)
  	}
  }

  func TestCode2Session_UsesAppidSecret(t *testing.T) {
  	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  		if r.URL.Path == "/sns/jscode2session" {
  			if r.URL.Query().Get("appid") == "" {
  				http.Error(w, "missing appid", 400)
  				return
  			}
  			json.NewEncoder(w).Encode(map[string]interface{}{
  				"openid":      "test-openid",
  				"session_key": "test-session-key",
  				"errcode":     0,
  				"errmsg":      "ok",
  			})
  		} else {
  			// token endpoint
  			json.NewEncoder(w).Encode(map[string]interface{}{
  				"access_token": "tok", "expires_in": 7200,
  			})
  		}
  	}))
  	defer srv.Close()

  	cfg := &Config{BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"}}
  	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/stable_token", "POST")
  	c := &Client{BaseClient: base}

  	result, err := c.Code2Session("test-code")
  	if err != nil {
  		t.Fatalf("unexpected error: %v", err)
  	}
  	if result.OpenId != "test-openid" {
  		t.Errorf("expected test-openid, got %s", result.OpenId)
  	}
  }
  ```

- [ ] **Step 2: Run full build and tests**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  go build ./...
  go test ./mini-program/ -v
  ```

  Expected output: both `TestNewClient_UsesPostForToken` and `TestCode2Session_UsesAppidSecret` pass with `PASS` status.

- [ ] **Step 3: Commit**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  git add mini-program/client_test.go
  git commit -m "test(mini-program): add client and auth tests"
  ```
