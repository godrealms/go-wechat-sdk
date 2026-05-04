# Phase 1B: offiaccount Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor offiaccount/client.go to embed core.BaseClient, eliminating duplicated token management code, while keeping all existing API methods working unchanged.

**Architecture:** offiaccount.Config embeds core.BaseConfig; offiaccount.Client embeds *core.BaseClient. All token methods (GetAccessToken, TokenQuery, etc.) are promoted from BaseClient. Existing api.*.go files get two mechanical renames: c.ctx→c.Ctx, c.tokenQuery→c.TokenQuery.

**Tech Stack:** Go 1.23.1, depends on Plan A (core/ package) being complete first.

**PREREQUISITE:** Plan A (2026-04-09-phase1a-core-package.md) must be completed before starting this plan.

---

### Task 1: Update `offiaccount/client.go`

**Files:**
- Modify: `offiaccount/client.go`

- [ ] **Step 1: Replace the entire file content**

  The current `client.go` contains `Config` (with AppId/AppSecret/Token/EncodingAESKey), `Client` struct (with ctx, Config, Https, accessToken, tokenMutex fields), `NewClient`, `getAccessToken`, `refreshAccessToken`, `GetAccessTokenWithError`, and `tokenQuery`. All of these must be replaced with the following content that delegates token management to `core.BaseClient`:

  ```go
  package offiaccount

  import (
  	"context"
  	"github.com/godrealms/go-wechat-sdk/core"
  )

  // AccessToken type alias for backward compatibility
  type AccessToken = core.AccessToken

  // Config holds offiaccount-specific configuration
  type Config struct {
  	core.BaseConfig
  	Token          string `json:"token"`
  	EncodingAESKey string `json:"encodingAESKey"`
  }

  // Client is the WeChat Official Account client
  type Client struct {
  	*core.BaseClient
  	Token          string
  	EncodingAESKey string
  }

  // NewClient creates a new offiaccount client
  func NewClient(ctx context.Context, config *Config) *Client {
  	base := core.NewBaseClient(ctx, &config.BaseConfig, "https://api.weixin.qq.com", "/cgi-bin/token", "GET")
  	return &Client{
  		BaseClient:     base,
  		Token:          config.Token,
  		EncodingAESKey: config.EncodingAESKey,
  	}
  }
  ```

  Key changes:
  - `Config` now embeds `core.BaseConfig` (which carries AppId and AppSecret) instead of declaring them directly
  - `Client` now embeds `*core.BaseClient` instead of holding ctx/Config/Https/accessToken/tokenMutex
  - `NewClient` delegates to `core.NewBaseClient`
  - The four methods `getAccessToken`, `refreshAccessToken`, `GetAccessTokenWithError`, `tokenQuery` are removed entirely — they are now promoted from `*core.BaseClient`

- [ ] **Step 2: Commit**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  git add offiaccount/client.go
  git commit -m "refactor(offiaccount): replace client.go to embed core.BaseClient

  - Config embeds core.BaseConfig (AppId, AppSecret promoted)
  - Client embeds *core.BaseClient (token management promoted)
  - NewClient delegates to core.NewBaseClient
  - Remove getAccessToken, refreshAccessToken, GetAccessTokenWithError, tokenQuery (now promoted from BaseClient)"
  ```

---

### Task 2: Update `offiaccount/api.base.go`

**Files:**
- Modify: `offiaccount/api.base.go`

- [ ] **Step 1: Remove `GetAccessToken()` wrapper**

  The current file has:
  ```go
  // GetAccessToken 获取接口调用凭据
  func (c *Client) GetAccessToken() string {
      return c.getAccessToken()
  }
  ```
  Delete this entire function (lines 11–15). It is now promoted from `*core.BaseClient` via embedding and the explicit wrapper shadows the promoted method.

- [ ] **Step 2: Update `GetStableAccessToken` — replace direct mutex/field access with `c.SetAccessToken` and `c.Ctx`**

  Current code accesses `c.tokenMutex` and `c.accessToken` directly, and uses `context.Background()`. Replace the entire `GetStableAccessToken` function body with:

  ```go
  // GetStableAccessToken 获取稳定 AccessToken
  // 获取全局后台接口调用凭据，有效期最长为7200s，开发者需要进行妥善保存；
  // 有两种调用模式:
  //  1. 普通模式，access_token 有效期内重复调用该接口不会更新 access_token，绝大部分场景下使用该模式；
  //  2. 强制刷新模式，会导致上次获取的 access_token 失效，并返回新的 access_token；
  //
  // 该接口调用频率限制为 1万次 每分钟，每天限制调用 50万 次；
  // 与getAccessToken获取的调用凭证完全隔离，互不影响。该接口仅支持 POST JSON 形式的调用；
  func (c *Client) GetStableAccessToken(forceRefresh bool) (*AccessToken, error) {
  	body := map[string]interface{}{
  		"grant_type":    "client_credential",
  		"appid":         c.Config.AppId,
  		"secret":        c.Config.AppSecret,
  		"force_refresh": forceRefresh,
  	}
  	result := &AccessToken{}
  	err := c.Https.Post(c.Ctx, "/cgi-bin/stable_token", body, result)
  	if err != nil {
  		return nil, err
  	}
  	// 提前10秒过期，避免临界点问题
  	result.ExpiresIn = result.ExpiresIn + time.Now().Unix() - 10
  	c.SetAccessToken(result)
  	return result, nil
  }
  ```

  Changes from original:
  - `context.Background()` → `c.Ctx`
  - `c.tokenMutex.Lock() / c.accessToken = result / c.tokenMutex.Unlock()` → `c.SetAccessToken(result)`

- [ ] **Step 3: Update `CallbackCheck`, `GetCallbackIp`, `GetApiDomainIP` — replace `c.getAccessToken()` and `context.Background()`**

  In each of these three functions:
  - Replace `c.getAccessToken()` with `c.GetAccessToken()`
  - Replace `context.Background()` with `c.Ctx`

  Full updated `CallbackCheck`:
  ```go
  func (c *Client) CallbackCheck(action, checkOperator string) (*CallbackCheckResponse, error) {
  	query := url.Values{
  		"access_token": {c.GetAccessToken()},
  	}
  	body := map[string]interface{}{
  		"action":         action,
  		"check_operator": checkOperator,
  	}
  	result := &CallbackCheckResponse{}
  	path := fmt.Sprintf("/cgi-bin/callback/check?%s", query.Encode())
  	err := c.Https.Post(c.Ctx, path, body, result)
  	if err != nil {
  		return nil, err
  	}
  	return result, nil
  }
  ```

  Full updated `GetCallbackIp`:
  ```go
  func (c *Client) GetCallbackIp() ([]string, error) {
  	query := url.Values{
  		"access_token": {c.GetAccessToken()},
  	}
  	var result = &IpList{}
  	err := c.Https.Get(c.Ctx, "/cgi-bin/getcallbackip", query, result)
  	if err != nil {
  		return nil, err
  	} else if result.ErrCode != 0 {
  		return nil, errors.New(result.ErrMsg)
  	}
  	return result.IpList, nil
  }
  ```

  Full updated `GetApiDomainIP`:
  ```go
  func (c *Client) GetApiDomainIP() ([]string, error) {
  	query := url.Values{
  		"access_token": {c.GetAccessToken()},
  	}
  	var result = &IpList{}
  	err := c.Https.Get(c.Ctx, "/cgi-bin/get_api_domain_ip", query, result)
  	if err != nil {
  		return nil, err
  	} else if result.ErrCode != 0 {
  		return nil, errors.New(result.ErrMsg)
  	}
  	return result.IpList, nil
  }
  ```

- [ ] **Step 4: Clean up imports in api.base.go**

  After the changes above the import block should be:
  ```go
  import (
  	"errors"
  	"fmt"
  	"net/url"
  	"time"
  )
  ```

  Remove `"context"` (no longer used directly — `c.Ctx` comes from the embedded struct).

- [ ] **Step 5: Commit**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  git add offiaccount/api.base.go
  git commit -m "refactor(offiaccount): update api.base.go to use promoted BaseClient methods

  - Remove GetAccessToken() wrapper (promoted from core.BaseClient)
  - GetStableAccessToken: use c.Ctx and c.SetAccessToken instead of direct mutex/field access
  - CallbackCheck, GetCallbackIp, GetApiDomainIP: c.getAccessToken() -> c.GetAccessToken(), context.Background() -> c.Ctx"
  ```

---

### Task 3: Global find-replace across all `api.*.go` files

**Files:**
- Modify: all `offiaccount/api.*.go` files (191 occurrences of `c.ctx`, 8 occurrences of `c.tokenQuery`, 4 occurrences of `c.getAccessToken()` spread across ~35 files)

- [ ] **Step 1: Rename `c.ctx` to `c.Ctx` in all api files**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  find offiaccount -name "api.*.go" -exec sed -i '' 's/c\.ctx\b/c.Ctx/g' {} +
  ```

  Verify the rename worked and no old references remain:
  ```bash
  grep -rn "c\.ctx\b" offiaccount/api.*.go
  # Expected output: (empty — no matches)
  ```

- [ ] **Step 2: Rename `c.tokenQuery(` to `c.TokenQuery(` in all api files**

  ```bash
  find offiaccount -name "api.*.go" -exec sed -i '' 's/c\.tokenQuery(/c.TokenQuery(/g' {} +
  ```

  Verify:
  ```bash
  grep -rn "c\.tokenQuery(" offiaccount/api.*.go
  # Expected output: (empty — no matches)
  ```

- [ ] **Step 3: Rename `c.getAccessToken()` to `c.GetAccessToken()` in all api files**

  ```bash
  find offiaccount -name "api.*.go" -exec sed -i '' 's/c\.getAccessToken()/c.GetAccessToken()/g' {} +
  ```

  Verify:
  ```bash
  grep -rn "c\.getAccessToken()" offiaccount/api.*.go
  # Expected output: (empty — no matches)
  ```

- [ ] **Step 4: Verify totals — confirm all renames are complete**

  ```bash
  # All three should return 0
  echo "c.ctx remaining:"; grep -rn "c\.ctx\b" offiaccount/api.*.go | wc -l
  echo "c.tokenQuery remaining:"; grep -rn "c\.tokenQuery(" offiaccount/api.*.go | wc -l
  echo "c.getAccessToken remaining:"; grep -rn "c\.getAccessToken()" offiaccount/api.*.go | wc -l
  ```

  Expected output:
  ```
  c.ctx remaining:
         0
  c.tokenQuery remaining:
         0
  c.getAccessToken remaining:
         0
  ```

- [ ] **Step 5: Commit**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  git add offiaccount/api.*.go
  git commit -m "refactor(offiaccount): mechanical rename in all api.*.go files

  - c.ctx -> c.Ctx (191 occurrences, exported field from core.BaseClient)
  - c.tokenQuery( -> c.TokenQuery( (8 occurrences, exported method from core.BaseClient)
  - c.getAccessToken() -> c.GetAccessToken() (4 occurrences, exported method from core.BaseClient)"
  ```

---

### Task 4: Add `offiaccount/client_test.go`

**Files:**
- Modify: `offiaccount/client_test.go` (new file)

- [ ] **Step 1: Create the test file**

  ```go
  package offiaccount

  import (
  	"context"
  	"encoding/json"
  	"net/http"
  	"net/http/httptest"
  	"testing"

  	"github.com/godrealms/go-wechat-sdk/core"
  )

  func TestNewClient_SetsFields(t *testing.T) {
  	cfg := &Config{
  		BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"},
  		Token:          "tok",
  		EncodingAESKey: "key",
  	}
  	c := NewClient(context.Background(), cfg)
  	if c.Token != "tok" {
  		t.Errorf("expected tok, got %s", c.Token)
  	}
  	if c.EncodingAESKey != "key" {
  		t.Errorf("expected key, got %s", c.EncodingAESKey)
  	}
  	if c.Config.AppId != "app1" {
  		t.Errorf("expected app1, got %s", c.Config.AppId)
  	}
  }

  func TestClient_GetAccessToken(t *testing.T) {
  	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  		json.NewEncoder(w).Encode(map[string]interface{}{
  			"access_token": "test-token-123",
  			"expires_in":   7200,
  		})
  	}))
  	defer srv.Close()

  	cfg := &Config{
  		BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"},
  	}
  	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "GET")
  	c := &Client{BaseClient: base}

  	token := c.GetAccessToken()
  	if token != "test-token-123" {
  		t.Errorf("expected test-token-123, got %s", token)
  	}
  }
  ```

- [ ] **Step 2: Commit**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  git add offiaccount/client_test.go
  git commit -m "test(offiaccount): add client_test.go for NewClient and GetAccessToken"
  ```

---

### Task 5: Build and test verification

**Files:**
- No file changes — verification only

- [ ] **Step 1: Build the entire module**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  go build ./...
  ```

  Expected output: no errors, no output (silent success).

  If there are compilation errors, they will be one of these categories:
  - `c.ctx` still referenced in a file missed by the sed — run the grep verification from Task 3 Step 4 to find stragglers
  - `c.tokenQuery` still referenced — same approach
  - `c.getAccessToken()` still referenced — same approach
  - Missing import in `api.base.go` — fix the import block per Task 2 Step 4
  - `c.accessToken` or `c.tokenMutex` still referenced somewhere — search with `grep -rn "c\.accessToken\|c\.tokenMutex" offiaccount/`

- [ ] **Step 2: Run TestNewClient_SetsFields**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  go test ./offiaccount/ -run TestNewClient -v
  ```

  Expected output:
  ```
  === RUN   TestNewClient_SetsFields
  --- PASS: TestNewClient_SetsFields (0.00s)
  PASS
  ok  	github.com/godrealms/go-wechat-sdk/offiaccount	0.xxxs
  ```

- [ ] **Step 3: Run TestClient_GetAccessToken**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  go test ./offiaccount/ -run TestClient_GetAccessToken -v
  ```

  Expected output:
  ```
  === RUN   TestClient_GetAccessToken
  --- PASS: TestClient_GetAccessToken (0.00s)
  PASS
  ok  	github.com/godrealms/go-wechat-sdk/offiaccount	0.xxxs
  ```

- [ ] **Step 4: Commit verification results (if any fixup commits were needed)**

  If the build or tests required additional fixes beyond what the plan specified, commit those fixes now:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
  git add -p   # stage only the fixup changes
  git commit -m "fix(offiaccount): address compilation issues found during Phase 1B verification"
  ```

  If build and tests passed cleanly with no extra fixes, skip this step.
