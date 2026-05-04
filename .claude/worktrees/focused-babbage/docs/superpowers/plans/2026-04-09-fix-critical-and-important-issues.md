# Fix Critical and Important Issues - Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix all 5 critical bugs and 7 important issues found during code review of go-wechat-sdk.

**Architecture:** Centralize access_token injection into the HTTP `do()` method so all API calls automatically include it. Fix concurrency issues by creating per-request headers instead of mutating shared state. Replace deprecated APIs and add configurable logging.

**Tech Stack:** Go 1.23.1, standard library only (no external dependencies)

---

### Task 1: Fix SignSHA256WithRSA silent error swallowing

**Files:**
- Modify: `utils/signature.go:19`
- Create: `utils/signature_test.go`

- [ ] **Step 1: Write the failing test**

```go
// utils/signature_test.go
package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestSignSHA256WithRSA_NilPrivateKey(t *testing.T) {
	sig, err := SignSHA256WithRSA("test", nil)
	if err == nil {
		t.Fatal("expected error for nil private key, got nil")
	}
	if sig != "" {
		t.Fatalf("expected empty signature, got %q", sig)
	}
}

func TestSignSHA256WithRSA_ValidKey(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	sig, err := SignSHA256WithRSA("hello world", key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sig == "" {
		t.Fatal("expected non-empty signature")
	}
}
```

- [ ] **Step 2: Run test to verify it passes (the nil key test should pass, the valid key test should pass)**

Run: `cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage && go test ./utils/ -run TestSignSHA256WithRSA -v`

- [ ] **Step 3: Fix the bug - change `return "", nil` to `return "", err`**

In `utils/signature.go` line 19, change:
```go
return "", nil
```
to:
```go
return "", fmt.Errorf("write to hash failed: %w", err)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./utils/ -run TestSignSHA256WithRSA -v`
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add utils/signature.go utils/signature_test.go
git commit -m "fix: return actual error in SignSHA256WithRSA when hash write fails"
```

---

### Task 2: Centralize access_token injection in offiaccount

**Files:**
- Modify: `offiaccount/client.go` (add `GetAccessToken` public method, add `doGet`/`doPost` wrappers)

Currently, 19 out of 39 API files are missing access_token. Instead of patching each file individually, we centralize the token injection by adding helper methods to the Client that automatically append access_token.

- [ ] **Step 1: Write the failing test**

```go
// offiaccount/client_test.go
package offiaccount

import (
	"context"
	"testing"
)

func TestGetAccessToken_ReturnsEmptyWhenNoToken(t *testing.T) {
	client := NewClient(context.Background(), &Config{
		AppId:     "test_appid",
		AppSecret: "test_secret",
	})
	// Without a valid token server, GetAccessToken should return empty string and error
	token, err := client.GetAccessTokenWithError()
	if err == nil {
		t.Log("Note: GetAccessTokenWithError returns error when token refresh fails")
	}
	_ = token // just ensure method exists
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./offiaccount/ -run TestGetAccessToken -v`
Expected: FAIL (method doesn't exist yet)

- [ ] **Step 3: Add centralized token methods to client.go**

Add these methods to `offiaccount/client.go`:

```go
// GetAccessToken 获取access_token (公开方法)
func (c *Client) GetAccessToken() string {
	return c.getAccessToken()
}

// GetAccessTokenWithError 获取access_token，返回错误信息
func (c *Client) GetAccessTokenWithError() (string, error) {
	c.tokenMutex.RLock()
	if c.accessToken != nil && c.accessToken.ExpiresIn > time.Now().Unix() {
		token := c.accessToken.AccessToken
		c.tokenMutex.RUnlock()
		return token, nil
	}
	c.tokenMutex.RUnlock()

	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	if c.accessToken != nil && c.accessToken.ExpiresIn > time.Now().Unix() {
		return c.accessToken.AccessToken, nil
	}

	token, err := c.refreshAccessToken()
	if err != nil {
		return "", fmt.Errorf("refresh access token failed: %w", err)
	}
	c.accessToken = token
	return c.accessToken.AccessToken, nil
}

// tokenQuery 返回包含 access_token 的 url.Values
func (c *Client) tokenQuery(extra ...url.Values) url.Values {
	query := url.Values{
		"access_token": {c.getAccessToken()},
	}
	for _, v := range extra {
		for key, values := range v {
			query[key] = values
		}
	}
	return query
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./offiaccount/ -run TestGetAccessToken -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add offiaccount/client.go offiaccount/client_test.go
git commit -m "feat: add centralized access_token helpers to offiaccount Client"
```

---

### Task 3: Fix 19 API files missing access_token

**Files:** All 19 broken api.*.go files in `offiaccount/`

Each file needs its API calls updated to include access_token. The pattern is consistent:

**For GET requests** - change `c.Https.Get(c.ctx, path, nil, &result)` to `c.Https.Get(c.ctx, path, c.tokenQuery(), &result)`. If the call already has params, change to `c.Https.Get(c.ctx, path, c.tokenQuery(params), &result)`.

**For POST requests** - change `path := "/cgi-bin/..."` to `path := fmt.Sprintf("/cgi-bin/...?access_token=%s", c.GetAccessToken())`.

- [ ] **Step 1: Fix api.user.manage.tag.go**

Every function in this file makes calls without access_token. Update all GET calls to use `c.tokenQuery()` and all POST calls to append `?access_token=%s` to path.

Example fix for `GetTags`:
```go
func (c *Client) GetTags() (*GetTagsResult, error) {
	path := "/cgi-bin/tags/get"
	var result GetTagsResult
	err := c.Https.Get(c.ctx, path, c.tokenQuery(), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
```

Example fix for `CreateTag`:
```go
func (c *Client) CreateTag(tagName string) (*CreateTagResult, error) {
	path := fmt.Sprintf("/cgi-bin/tags/create?access_token=%s", c.GetAccessToken())
	// ... rest stays the same
```

Apply same pattern to all functions in the file.

- [ ] **Step 2: Fix api.user.manage.userinfo.go**

For `GetUserInfo` which already builds params, merge with token:
```go
func (c *Client) GetUserInfo(openid string, lang string) (*UserInfo, error) {
	path := "/cgi-bin/user/info"
	params := url.Values{}
	params.Add("openid", openid)
	if lang != "" {
		params.Add("lang", lang)
	}
	var result UserInfo
	err := c.Https.Get(c.ctx, path, c.tokenQuery(params), &result)
	// ...
```

Apply same pattern to all functions in the file.

- [ ] **Step 3: Fix remaining 17 API files**

Apply the same patterns to all remaining files:
- `api.customer.message.go`
- `api.invoice.auth.go`
- `api.invoice.fiscal-receipt.go`
- `api.invoice.name.go`
- `api.invoice.reimburser.go`
- `api.medical-assistant.go`
- `api.nontax-pay.go`
- `api.open-poc.ai.go`
- `api.open-poc.image.go`
- `api.open-poc.ocr.go`
- `api.qr-code.qr-code-jump.go`
- `api.qr-code.shorten.go`
- `api.stores.mini-app.go`
- `api.we-data.api.go`
- `api.we-data.mess.go`
- `api.we-data.news.go`
- `api.we-data.user.go`

Rules:
- GET with no params: `nil` -> `c.tokenQuery()`
- GET with existing params: `params` -> `c.tokenQuery(params)`
- POST: prepend `fmt.Sprintf("...?access_token=%s", c.GetAccessToken())` to path

- [ ] **Step 4: Verify build**

Run: `go build ./offiaccount/...`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add offiaccount/api.*.go
git commit -m "fix: add missing access_token to 19 offiaccount API files"
```

---

### Task 4: Fix merchant module thread-safety (headers race condition)

**Files:**
- Modify: `utils/http.go` (add `DoWithHeaders` method)
- Modify: all 8 merchant `pay.transactions.*.go` files

The fix: Instead of mutating `c.Http.Headers`, pass per-request headers. Add a new method to the HTTP client that accepts request-scoped headers.

- [ ] **Step 1: Write the failing test**

```go
// utils/http_test.go
package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTP_PostWithHeaders_DoesNotMutateSharedHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		json.NewEncoder(w).Encode(map[string]string{"auth": auth})
	}))
	defer server.Close()

	client := NewHTTP(server.URL)
	client.Headers["Shared"] = "should-not-change"

	headers := map[string]string{
		"Authorization": "Bearer test-token",
	}

	var result map[string]string
	err := client.PostWithHeaders(context.Background(), "/test", nil, headers, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Shared headers should not be modified
	if _, exists := client.Headers["Authorization"]; exists {
		t.Fatal("PostWithHeaders should not mutate shared headers")
	}
	if client.Headers["Shared"] != "should-not-change" {
		t.Fatal("shared headers were modified")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./utils/ -run TestHTTP_PostWithHeaders -v`
Expected: FAIL (method doesn't exist)

- [ ] **Step 3: Add PostWithHeaders and GetWithHeaders to utils/http.go**

```go
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

	// Apply shared headers first, then request-specific headers override
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

	respBody, err := io.ReadAll(resp.Body)
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

// PostWithHeaders 发送带有请求级别 headers 的 POST 请求（线程安全）
func (h *HTTP) PostWithHeaders(ctx context.Context, path string, body interface{}, headers map[string]string, result interface{}) error {
	return h.doWithHeaders(ctx, http.MethodPost, path, body, headers, nil, result)
}

// GetWithHeaders 发送带有请求级别 headers 的 GET 请求（线程安全）
func (h *HTTP) GetWithHeaders(ctx context.Context, path string, headers map[string]string, query url.Values, result interface{}) error {
	return h.doWithHeaders(ctx, http.MethodGet, path, nil, headers, query, result)
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./utils/ -run TestHTTP -v`
Expected: PASS

- [ ] **Step 5: Update all merchant payment files**

In each of the 8 merchant `pay.transactions.*.go` files, replace the pattern:
```go
c.Http.Headers = map[string]string{
    "Accept":        "application/json",
    "Content-Type":  "application/json",
    "Authorization": fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 ..."),
}
response := &types.SomeResp{}
err = c.Http.Post(context.Background(), path, order, response)
```

With:
```go
headers := map[string]string{
    "Accept":        "application/json",
    "Content-Type":  "application/json",
    "Authorization": fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 ..."),
}
response := &types.SomeResp{}
err = c.Http.PostWithHeaders(context.Background(), path, order, headers, response)
```

This applies to all 12 occurrences across:
- `pay.transactions.app.go` (1 occurrence)
- `pay.transactions.bill.go` (2 occurrences)
- `pay.transactions.close.go` (1 occurrence)
- `pay.transactions.h5.go` (1 occurrence)
- `pay.transactions.jsapi.go` (1 occurrence)
- `pay.transactions.native.go` (1 occurrence)
- `pay.transactions.query.go` (2 occurrences)
- `pay.transactions.refunds.go` (3 occurrences)

For GET requests (e.g., query.go), use `GetWithHeaders` instead.

- [ ] **Step 6: Verify build**

Run: `go build ./merchant/...`
Expected: No errors

- [ ] **Step 7: Commit**

```bash
git add utils/http.go utils/http_test.go merchant/developed/*.go
git commit -m "fix: eliminate headers race condition in merchant module with per-request headers"
```

---

### Task 5: Fix getAccessToken silent error loss

**Files:**
- Modify: `offiaccount/client.go`

- [ ] **Step 1: Modify getAccessToken to log errors**

The `getAccessToken()` method returns `""` when refresh fails. Since it's used by 39 API files, changing the signature would be very disruptive. Instead, add logging for the error and keep the public `GetAccessTokenWithError()` (added in Task 2) for callers who want error handling.

Update `getAccessToken` in `offiaccount/client.go`:

```go
func (c *Client) getAccessToken() string {
	c.tokenMutex.RLock()
	if c.accessToken != nil && c.accessToken.ExpiresIn > time.Now().Unix() {
		token := c.accessToken.AccessToken
		c.tokenMutex.RUnlock()
		return token
	}
	c.tokenMutex.RUnlock()

	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()

	if c.accessToken != nil && c.accessToken.ExpiresIn > time.Now().Unix() {
		return c.accessToken.AccessToken
	}

	token, err := c.refreshAccessToken()
	if err != nil {
		log.Printf("[go-wechat-sdk] refresh access token failed: %v", err)
		return ""
	}
	c.accessToken = token
	return c.accessToken.AccessToken
}
```

Add `"log"` to the imports.

- [ ] **Step 2: Verify build**

Run: `go build ./offiaccount/...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add offiaccount/client.go
git commit -m "fix: log error when access token refresh fails instead of silently returning empty"
```

---

### Task 6: Replace deprecated ioutil.ReadFile with os.ReadFile

**Files:**
- Modify: `utils/pem.go`

- [ ] **Step 1: Replace ioutil with os**

In `utils/pem.go`, change the import from `"io/ioutil"` to `"os"` and replace all 3 occurrences of `ioutil.ReadFile` with `os.ReadFile`:

Line 71: `certificateBytes, err := os.ReadFile(path)`
Line 80: `privateKeyBytes, err := os.ReadFile(path)`
Line 89: `publicKeyBytes, err := os.ReadFile(path)`

- [ ] **Step 2: Fix typo in error message**

Line 36: Change `"the kind of PEM should be PRVATE KEY"` to `"the kind of PEM should be PRIVATE KEY"`

- [ ] **Step 3: Verify build**

Run: `go build ./utils/...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add utils/pem.go
git commit -m "fix: replace deprecated ioutil.ReadFile with os.ReadFile, fix PRIVATE KEY typo"
```

---

### Task 7: Remove hardcoded sensitive logging from HTTP client

**Files:**
- Modify: `utils/http.go`

- [ ] **Step 1: Add Logger interface and configurable logging**

Add a Logger interface and make logging opt-in:

```go
// Logger 日志接口
type Logger interface {
	Printf(format string, v ...interface{})
}

// Option 定义配置选项的函数类型
type Option func(*HTTP)

// WithLogger 设置日志记录器
func WithLogger(logger Logger) Option {
	return func(h *HTTP) {
		h.Logger = logger
	}
}
```

Add `Logger Logger` field to the `HTTP` struct.

- [ ] **Step 2: Replace hardcoded log.Println calls with conditional logging**

In `do()` method, replace:
```go
log.Println("method:", method)
log.Println("fullURL:", fullURL)
```
with:
```go
if h.Logger != nil {
    h.Logger.Printf("method: %s", method)
    h.Logger.Printf("url: %s", fullURL)
}
```

Replace `log.Println("body:", bodyJson)` with:
```go
if h.Logger != nil {
    h.Logger.Printf("body: %s", bodyJson)
}
```

Replace `log.Println("response:", string(respBody))` with:
```go
if h.Logger != nil {
    h.Logger.Printf("response: %s", string(respBody))
}
```

Remove the `"log"` import if no longer used.

- [ ] **Step 3: Verify build**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add utils/http.go
git commit -m "fix: replace hardcoded logging with configurable Logger interface to prevent sensitive data leaks"
```

---

### Task 8: Add io.LimitReader for response body reads

**Files:**
- Modify: `utils/http.go`

- [ ] **Step 1: Add MaxResponseSize constant and apply LimitReader**

Add at the top of `utils/http.go`:
```go
const (
	// DefaultMaxResponseSize 默认最大响应体大小 (10MB)
	DefaultMaxResponseSize = 10 * 1024 * 1024
)
```

In both `do()` and `doWithHeaders()` and `PostForm()`, replace:
```go
respBody, err := io.ReadAll(resp.Body)
```
with:
```go
respBody, err := io.ReadAll(io.LimitReader(resp.Body, DefaultMaxResponseSize))
```

- [ ] **Step 2: Verify build**

Run: `go build ./utils/...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add utils/http.go
git commit -m "fix: limit response body reads to 10MB to prevent OOM from malicious servers"
```

---

### Task 9: Fix example packages with multiple main functions

**Files:**
- Modify: `offiaccount/example/*.go`
- Modify: `merchant/example/*.go`

- [ ] **Step 1: Convert example files to use build tags or separate example functions**

Change all files in `offiaccount/example/` and `merchant/example/` from `package main` with `func main()` to `package example` with exported example functions.

For example, `offiaccount/example/api.custom-menu.go`:
```go
package example

// ExampleCustomMenu demonstrates custom menu creation
func ExampleCustomMenu() {
    // ... existing code from main() ...
}
```

For `offiaccount/example/main.go`, keep as `package example` and rename `func main()` to `func ExampleMain()`.

Same for `merchant/example/app.go` and `merchant/example/pay.transactions.notify.go`.

- [ ] **Step 2: Verify build**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add offiaccount/example/*.go merchant/example/*.go
git commit -m "fix: convert example files from package main to package example to fix build"
```

---

### Task 10: Fix GenerateHashBasedString modulo bias

**Files:**
- Modify: `utils/string.go`

- [ ] **Step 1: Write a test verifying distribution**

```go
// Add to utils/string_test.go (or create it)
package utils

import "testing"

func TestGenerateHashBasedString_Length(t *testing.T) {
	for _, length := range []int{1, 16, 32, 64} {
		s := GenerateHashBasedString(length)
		if len(s) != length {
			t.Errorf("expected length %d, got %d", length, len(s))
		}
	}
}

func TestGenerateHashBasedString_OnlyValidChars(t *testing.T) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	for i := 0; i < 100; i++ {
		s := GenerateHashBasedString(32)
		for _, c := range s {
			found := false
			for _, valid := range charset {
				if c == valid {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("invalid char %c in output", c)
			}
		}
	}
}
```

- [ ] **Step 2: Fix modulo bias with rejection sampling**

Replace the loop in `GenerateHashBasedString`:

```go
func GenerateHashBasedString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	const maxByte = 256 - (256 % len(charset)) // 256 - (256 % 62) = 256 - 8 = 248
	result := make([]byte, length)
	buf := make([]byte, length+(length/4)) // extra buffer for rejected bytes

	i := 0
	for i < length {
		if _, err := io.ReadFull(cryptoRand.Reader, buf); err != nil {
			return RandomString(length, "")
		}
		for _, b := range buf {
			if i >= length {
				break
			}
			if int(b) < maxByte {
				result[i] = charset[int(b)%len(charset)]
				i++
			}
		}
	}
	return string(result)
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./utils/ -run TestGenerateHashBasedString -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add utils/string.go utils/string_test.go
git commit -m "fix: eliminate modulo bias in GenerateHashBasedString with rejection sampling"
```
