# Phase 1A: core/ Package Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create the shared `core/` package that all WeChat SDK modules will embed, providing token management, base config, and shared response types.

**Architecture:** Extract the token management logic from offiaccount into a reusable core.BaseClient. Use TokenMethod ("GET"/"POST") to support both offiaccount-style GET token and mini-program-style POST token. No external dependencies.

**Tech Stack:** Go 1.23.1, standard library only (context, sync, net/url, time, net/http/httptest for tests)

---

### Task 1: Scaffolding

**Files:**
- Create: `core/config.go`
- Create: `core/response.go`
- Create: `core/token.go`
- Test: `core/response_test.go`

- [ ] **Step 1: Create `core/config.go`**

```go
package core

type BaseConfig struct {
	AppId     string `json:"appId"`
	AppSecret string `json:"appSecret"`
}
```

- [ ] **Step 2: Create `core/response.go`**

```go
package core

import "fmt"

type Resp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (r *Resp) GetError() error {
	if r.ErrCode != 0 {
		return fmt.Errorf("wechat api error %d: %s", r.ErrCode, r.ErrMsg)
	}
	return nil
}
```

- [ ] **Step 3: Create `core/token.go`**

```go
package core

type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}
```

- [ ] **Step 4: Create `core/response_test.go`**

```go
package core

import (
	"testing"
)

func TestResp_GetError_ReturnsNilWhenErrCodeIsZero(t *testing.T) {
	r := &Resp{ErrCode: 0, ErrMsg: "ok"}
	if err := r.GetError(); err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestResp_GetError_ReturnsErrorWhenErrCodeIsNonZero(t *testing.T) {
	r := &Resp{ErrCode: 40001, ErrMsg: "invalid credential"}
	err := r.GetError()
	if err == nil {
		t.Fatal("expected non-nil error, got nil")
	}
	expected := "wechat api error 40001: invalid credential"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func TestResp_GetError_FormatsMessageCorrectly(t *testing.T) {
	r := &Resp{ErrCode: 40013, ErrMsg: "invalid appid"}
	err := r.GetError()
	if err == nil {
		t.Fatal("expected non-nil error, got nil")
	}
	expected := "wechat api error 40013: invalid appid"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}
```

- [ ] **Step 5: Run tests**

```
go test ./core/ -v -run TestResp
```

Expected output: all 3 TestResp tests pass.

- [ ] **Step 6: Commit**

```
git add core/config.go core/response.go core/token.go core/response_test.go
git commit -m "feat(core): add BaseConfig, Resp, and AccessToken scaffolding"
```

---

### Task 2: BaseClient

**Files:**
- Create: `core/client.go`

- [ ] **Step 1: Create `core/client.go`**

```go
package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// BaseClient provides shared token management and HTTP functionality
// for all WeChat SDK sub-packages.
type BaseClient struct {
	Ctx          context.Context
	Config       *BaseConfig
	Https        *utils.HTTP
	accessToken  *AccessToken
	tokenMutex   sync.RWMutex
	TokenURL     string // e.g. "/cgi-bin/token"
	TokenMethod  string // "GET" or "POST"
}

// NewBaseClient creates and returns an initialised BaseClient.
// baseURL is the HTTP base URL (e.g. "https://api.weixin.qq.com").
// tokenURL is the path used to fetch an access token.
// tokenMethod is either "GET" or "POST".
func NewBaseClient(
	ctx context.Context,
	config *BaseConfig,
	baseURL string,
	tokenURL string,
	tokenMethod string,
) *BaseClient {
	return &BaseClient{
		Ctx:         ctx,
		Config:      config,
		Https:       utils.NewHTTP(baseURL),
		TokenURL:    tokenURL,
		TokenMethod: tokenMethod,
	}
}

// getAccessToken returns a valid access token, refreshing if necessary.
// It uses double-checked locking to avoid redundant refreshes.
func (c *BaseClient) getAccessToken() string {
	// Fast path: read lock
	c.tokenMutex.RLock()
	if c.accessToken != nil && c.accessToken.ExpiresIn > time.Now().Unix() {
		token := c.accessToken.AccessToken
		c.tokenMutex.RUnlock()
		return token
	}
	c.tokenMutex.RUnlock()

	// Slow path: write lock with double-check
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	if c.accessToken != nil && c.accessToken.ExpiresIn > time.Now().Unix() {
		return c.accessToken.AccessToken
	}
	token, err := c.refreshAccessToken()
	if err != nil {
		return ""
	}
	c.accessToken = token
	return c.accessToken.AccessToken
}

// refreshAccessToken fetches a new access token from the WeChat API.
// The request method (GET or POST) is determined by BaseClient.TokenMethod.
func (c *BaseClient) refreshAccessToken() (*AccessToken, error) {
	var result AccessToken

	if c.TokenMethod == "POST" {
		body := map[string]string{
			"grant_type": "client_credential",
			"appid":      c.Config.AppId,
			"secret":     c.Config.AppSecret,
		}
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("core: marshal token request: %w", err)
		}
		resp, err := c.Https.Post(c.TokenURL, bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("core: POST token request: %w", err)
		}
		if err = json.Unmarshal(resp, &result); err != nil {
			return nil, fmt.Errorf("core: unmarshal token response: %w", err)
		}
	} else {
		// Default: GET
		params := url.Values{
			"grant_type": {"client_credential"},
			"appid":      {c.Config.AppId},
			"secret":     {c.Config.AppSecret},
		}
		resp, err := c.Https.Get(c.TokenURL + "?" + params.Encode())
		if err != nil {
			return nil, fmt.Errorf("core: GET token request: %w", err)
		}
		if err = json.Unmarshal(resp, &result); err != nil {
			return nil, fmt.Errorf("core: unmarshal token response: %w", err)
		}
	}

	// Convert relative TTL to absolute expiry, subtract 10 s as safety margin
	result.ExpiresIn = time.Now().Unix() + result.ExpiresIn - 10
	return &result, nil
}

// GetAccessToken returns the current valid access token.
func (c *BaseClient) GetAccessToken() string {
	return c.getAccessToken()
}

// GetAccessTokenWithError returns the current valid access token or an error
// if the token cannot be obtained.
func (c *BaseClient) GetAccessTokenWithError() (string, error) {
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
		return "", err
	}
	c.accessToken = token
	return c.accessToken.AccessToken, nil
}

// SetAccessToken replaces the stored access token. This is needed by callers
// that obtain a token via a different mechanism (e.g. GetStableAccessToken).
func (c *BaseClient) SetAccessToken(token *AccessToken) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.accessToken = token
}

// TokenQuery returns url.Values containing the current access token merged
// with any additional values provided via extra.
func (c *BaseClient) TokenQuery(extra ...url.Values) url.Values {
	q := url.Values{
		"access_token": {c.GetAccessToken()},
	}
	for _, v := range extra {
		for key, vals := range v {
			q[key] = vals
		}
	}
	return q
}
```

- [ ] **Step 2: Verify the package compiles**

```
go build ./core/
```

Expected: no errors.

- [ ] **Step 3: Commit**

```
git add core/client.go
git commit -m "feat(core): add BaseClient with token management"
```

---

### Task 3: BaseClient Tests

**Files:**
- Test: `core/client_test.go`

- [ ] **Step 1: Create `core/client_test.go`**

```go
package core

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"
)

// newTestServer starts an httptest.Server that returns a JSON access token
// response. It increments callCount on each hit so tests can verify caching.
func newTestServer(t *testing.T, callCount *int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(callCount, 1)
		resp := map[string]interface{}{
			"access_token": "test-token-12345",
			"expires_in":   7200,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

// newMethodTrackingServer starts an httptest.Server that records the HTTP
// method used for each request in addition to counting calls.
func newMethodTrackingServer(t *testing.T, callCount *int32, methods *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(callCount, 1)
		*methods = append(*methods, r.Method)
		resp := map[string]interface{}{
			"access_token": "post-token-67890",
			"expires_in":   7200,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

// TestBaseClient_GetAccessToken_CachesToken verifies that a valid token is
// returned from cache on the second call without hitting the HTTP server again.
func TestBaseClient_GetAccessToken_CachesToken(t *testing.T) {
	var callCount int32
	srv := newTestServer(t, &callCount)
	defer srv.Close()

	cfg := &BaseConfig{AppId: "wx_app_id", AppSecret: "app_secret"}
	client := NewBaseClient(context.Background(), cfg, srv.URL, "/token", "GET")

	token1 := client.GetAccessToken()
	token2 := client.GetAccessToken()

	if token1 == "" {
		t.Fatal("expected non-empty token on first call")
	}
	if token1 != token2 {
		t.Errorf("expected same token on second call; got %q and %q", token1, token2)
	}
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 HTTP call, got %d", atomic.LoadInt32(&callCount))
	}
}

// TestBaseClient_GetAccessToken_RefreshesExpiredToken verifies that an expired
// token causes a new HTTP call to refresh it.
func TestBaseClient_GetAccessToken_RefreshesExpiredToken(t *testing.T) {
	var callCount int32
	srv := newTestServer(t, &callCount)
	defer srv.Close()

	cfg := &BaseConfig{AppId: "wx_app_id", AppSecret: "app_secret"}
	client := NewBaseClient(context.Background(), cfg, srv.URL, "/token", "GET")

	// Pre-set an expired token (ExpiresIn is already an absolute Unix timestamp)
	expiredToken := &AccessToken{
		AccessToken: "expired-token",
		ExpiresIn:   time.Now().Unix() - 60, // expired 60 seconds ago
	}
	client.SetAccessToken(expiredToken)

	token := client.GetAccessToken()
	if token == "expired-token" {
		t.Error("expected token to be refreshed, but got the expired token")
	}
	if token == "" {
		t.Fatal("expected a new non-empty token after refresh")
	}
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("expected 1 HTTP call for refresh, got %d", atomic.LoadInt32(&callCount))
	}
}

// TestBaseClient_PostTokenMethod verifies that when TokenMethod is "POST",
// the BaseClient sends a POST request (not GET) to obtain the token.
func TestBaseClient_PostTokenMethod(t *testing.T) {
	var callCount int32
	var methods []string
	srv := newMethodTrackingServer(t, &callCount, &methods)
	defer srv.Close()

	cfg := &BaseConfig{AppId: "wx_app_id", AppSecret: "app_secret"}
	client := NewBaseClient(context.Background(), cfg, srv.URL, "/token", "POST")

	token := client.GetAccessToken()
	if token == "" {
		t.Fatal("expected non-empty token from POST method client")
	}
	if len(methods) == 0 {
		t.Fatal("no HTTP requests were recorded")
	}
	if methods[0] != http.MethodPost {
		t.Errorf("expected HTTP method POST, got %s", methods[0])
	}
}

// TestBaseClient_TokenQuery_MergesExtra verifies that TokenQuery returns
// url.Values with both "access_token" and any extra parameters merged in.
func TestBaseClient_TokenQuery_MergesExtra(t *testing.T) {
	var callCount int32
	srv := newTestServer(t, &callCount)
	defer srv.Close()

	cfg := &BaseConfig{AppId: "wx_app_id", AppSecret: "app_secret"}
	client := NewBaseClient(context.Background(), cfg, srv.URL, "/token", "GET")

	extra := url.Values{
		"openid": {"user_open_id_123"},
		"lang":   {"zh_CN"},
	}

	q := client.TokenQuery(extra)

	if len(q["access_token"]) == 0 || q["access_token"][0] == "" {
		t.Error("expected non-empty access_token in query values")
	}
	if q.Get("openid") != "user_open_id_123" {
		t.Errorf("expected openid=user_open_id_123, got %q", q.Get("openid"))
	}
	if q.Get("lang") != "zh_CN" {
		t.Errorf("expected lang=zh_CN, got %q", q.Get("lang"))
	}
}

// TestBaseClient_SetAccessToken verifies that a token set via SetAccessToken
// is immediately returned by GetAccessToken without any HTTP call.
func TestBaseClient_SetAccessToken(t *testing.T) {
	var callCount int32
	srv := newTestServer(t, &callCount)
	defer srv.Close()

	cfg := &BaseConfig{AppId: "wx_app_id", AppSecret: "app_secret"}
	client := NewBaseClient(context.Background(), cfg, srv.URL, "/token", "GET")

	freshToken := &AccessToken{
		AccessToken: "manually-set-token",
		ExpiresIn:   time.Now().Unix() + 7000,
	}
	client.SetAccessToken(freshToken)

	token := client.GetAccessToken()
	if token != "manually-set-token" {
		t.Errorf("expected manually-set-token, got %q", token)
	}
	if atomic.LoadInt32(&callCount) != 0 {
		t.Errorf("expected 0 HTTP calls (token was pre-set), got %d", atomic.LoadInt32(&callCount))
	}
}
```

- [ ] **Step 2: Run all core tests**

```
go test ./core/ -v
```

Expected output: all 8 tests pass (3 TestResp + 5 TestBaseClient).

- [ ] **Step 3: Commit**

```
git add core/client_test.go
git commit -m "test(core): add BaseClient unit tests with httptest"
```

---

### Task 4: Verify Build

**Files:** (none new — verification only)

- [ ] **Step 1: Build entire module**

```
go build ./...
```

Expected: exits 0 with no errors or warnings.

- [ ] **Step 2: Run full test suite for core**

```
go test ./core/ -v -count=1
```

Expected: all tests pass, no data races (add `-race` flag for extra confidence):

```
go test -race ./core/ -v -count=1
```

- [ ] **Step 3: Confirm package is importable from offiaccount**

Add a temporary blank import in `offiaccount/client.go` to confirm there are no circular dependencies, then remove it:

```go
import _ "github.com/godrealms/go-wechat-sdk/core"
```

Run `go build ./offiaccount/` — expect no errors — then remove the temporary import.

- [ ] **Step 4: Final commit (if any fixups needed)**

If any fixups were needed during verification:

```
git add -p
git commit -m "fix(core): address build/test issues found during verification"
```

If everything was clean, no commit is needed for this task.
