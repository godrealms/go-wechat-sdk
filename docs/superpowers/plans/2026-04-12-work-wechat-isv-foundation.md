# work-wechat ISV 认证底座 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现企业微信服务商(ISV)认证底座 `work-wechat/isv`,提供 16 个公开方法覆盖 suite_token / pre_auth_code / 永久授权码 / corp_token / 运维 / ID 转换 / 回调解密。

**Architecture:** 新建 `work-wechat/isv` 独立子包,复刻 `oplatform` 的架构范式(Config/Options/Store/TokenSource/Lazy+RefreshAll/doPost+doGet/httptest+newTestISVClient helper/ParseNotify 强类型事件分发)。复用 `utils/wxcrypto` 做回调加解密。零新基础设施,所有模式都来自已稳定的 oplatform。

**Tech Stack:** Go 1.23+ / `encoding/json` / `encoding/xml` / `net/http` / `net/http/httptest` / `sync` / `utils/wxcrypto`

**Spec:** `docs/superpowers/specs/2026-04-12-work-wechat-isv-foundation-design.md`

---

## 文件布局总览

```
work-wechat/
  isv/
    doc.go                      [T1] 包文档
    errors.go                   [T1] WeixinError + 哨兵错误
    tokensource.go              [T1] TokenSource 接口
    store.go                    [T2] Store 接口 + MemoryStore
    store_test.go               [T2]
    client.go                   [T3] Config / Client / Options / NewClient / HTTP 助手
    client_test.go              [T3]
    suite.token.go              [T4] GetSuiteAccessToken / RefreshSuiteToken
    suite.token_test.go         [T4]
    suite.preauth.go            [T5] GetPreAuthCode / SetSessionInfo / AuthorizeURL
    suite.preauth_test.go       [T5]
    suite.permanent.go          [T6] GetPermanentCode / GetAuthInfo / GetAdminList
    suite.permanent_test.go     [T6]
    provider.id_convert.go      [T7] CorpIDToOpenCorpID / UserIDToOpenUserID + provider_token
    provider.id_convert_test.go [T7]
    corp.token.go               [T8] GetCorpToken / CorpClient / AccessToken / Refresh / RefreshAll
    corp.token_test.go          [T8]
    authorizer.go               [T8] compile-time assertion CorpClient 实现 TokenSource
    notify.go                   [T9,T10] ParseNotify
    notify_test.go              [T9,T10]
    struct.suite.go             [T4,T5] suite_token / pre_auth_code / session_info DTO
    struct.permanent.go         [T6,T7] permanent_code / auth_info / admin_list / ID 转换 DTO
    struct.corp.go              [T8] corp_token DTO
    struct.notify.go            [T9,T10] 9 种事件 + RawEvent
    example/
      main.go                   [T11] 编译级 demo
```

每个任务完成后 `go vet ./work-wechat/isv/...` 与 `go test ./work-wechat/isv/...` 都应为绿。

---

## Task 1: 脚手架(doc.go / errors.go / tokensource.go)

**Files:**
- Create: `work-wechat/isv/doc.go`
- Create: `work-wechat/isv/errors.go`
- Create: `work-wechat/isv/tokensource.go`

- [ ] **Step 1.1: Write `doc.go`**

```go
// Package isv 提供企业微信第三方应用服务商(ISV)的认证底座。
//
// 主要能力:
//   - 维护 suite_access_token / provider_access_token 的生命周期
//   - 处理企业管理员扫码授权流程(pre_auth_code → permanent_code → corp_token)
//   - 为下游"代企业调用"子项目提供 TokenSource 注入点
//   - 解密企业微信回调事件(suite_ticket / 授权变更 / 通讯录变更等)
//
// 本包对标 oplatform 代公众号/代小程序的认证底座,架构范式保持一致。
package isv
```

- [ ] **Step 1.2: Write `errors.go`**

```go
package isv

import (
	"errors"
	"fmt"
)

// 哨兵错误,可用 errors.Is 判断。
var (
	ErrNotFound              = errors.New("isv: not found")
	ErrSuiteTicketMissing    = errors.New("isv: suite_ticket missing in store")
	ErrProviderCorpIDMissing = errors.New("isv: provider corpid not configured")
	ErrProviderSecretMissing = errors.New("isv: provider secret not configured")
	ErrAuthorizerRevoked     = errors.New("isv: authorizer revoked")
)

// WeixinError 封装微信业务错误码(errcode != 0)。
type WeixinError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (e *WeixinError) Error() string {
	return fmt.Sprintf("isv: weixin error %d: %s", e.ErrCode, e.ErrMsg)
}
```

- [ ] **Step 1.3: Write `tokensource.go`**

```go
package isv

import "context"

// TokenSource 是下游"代企业调用"子项目的注入点。
// CorpClient 会实现此接口。
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}
```

- [ ] **Step 1.4: Verify compilation**

Run: `go vet ./work-wechat/isv/...`
Expected: no output (clean)

Run: `go build ./work-wechat/isv/...`
Expected: no output (clean)

- [ ] **Step 1.5: Commit**

```bash
git add work-wechat/isv/doc.go work-wechat/isv/errors.go work-wechat/isv/tokensource.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): scaffold — doc, errors, TokenSource interface

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Store 接口与 MemoryStore

**Files:**
- Create: `work-wechat/isv/store.go`
- Create: `work-wechat/isv/store_test.go`

- [ ] **Step 2.1: Write failing tests for MemoryStore**

Create `work-wechat/isv/store_test.go`:

```go
package isv

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryStore_SuiteTicket(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, err := s.GetSuiteTicket(ctx, "suite1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if err := s.PutSuiteTicket(ctx, "suite1", "TICKET"); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetSuiteTicket(ctx, "suite1")
	if err != nil || got != "TICKET" {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestMemoryStore_SuiteToken(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, _, err := s.GetSuiteToken(ctx, "suite1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	exp := time.Now().Add(time.Hour)
	if err := s.PutSuiteToken(ctx, "suite1", "TOK", exp); err != nil {
		t.Fatal(err)
	}
	tok, gotExp, err := s.GetSuiteToken(ctx, "suite1")
	if err != nil || tok != "TOK" || !gotExp.Equal(exp) {
		t.Fatalf("got %q %v err=%v", tok, gotExp, err)
	}
}

func TestMemoryStore_ProviderToken(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, _, err := s.GetProviderToken(ctx, "suite1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	exp := time.Now().Add(time.Hour)
	if err := s.PutProviderToken(ctx, "suite1", "PTOK", exp); err != nil {
		t.Fatal(err)
	}
	tok, gotExp, err := s.GetProviderToken(ctx, "suite1")
	if err != nil || tok != "PTOK" || !gotExp.Equal(exp) {
		t.Fatalf("got %q %v err=%v", tok, gotExp, err)
	}
}

func TestMemoryStore_Authorizer(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	if _, err := s.GetAuthorizer(ctx, "suite1", "corp1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	tokens := &AuthorizerTokens{
		CorpID:            "corp1",
		PermanentCode:     "PCODE",
		CorpAccessToken:   "CTOK",
		CorpTokenExpireAt: time.Now().Add(time.Hour),
	}
	if err := s.PutAuthorizer(ctx, "suite1", "corp1", tokens); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetAuthorizer(ctx, "suite1", "corp1")
	if err != nil {
		t.Fatal(err)
	}
	if got.PermanentCode != "PCODE" || got.CorpAccessToken != "CTOK" {
		t.Fatalf("got %+v", got)
	}
}

func TestMemoryStore_ListAuthorizers(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	_ = s.PutAuthorizer(ctx, "suite1", "corpA", &AuthorizerTokens{CorpID: "corpA"})
	_ = s.PutAuthorizer(ctx, "suite1", "corpB", &AuthorizerTokens{CorpID: "corpB"})
	_ = s.PutAuthorizer(ctx, "suite2", "corpC", &AuthorizerTokens{CorpID: "corpC"})

	list, err := s.ListAuthorizers(ctx, "suite1")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 corps, got %v", list)
	}
}

func TestMemoryStore_DeleteAuthorizer(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	_ = s.PutAuthorizer(ctx, "suite1", "corpA", &AuthorizerTokens{CorpID: "corpA"})
	if err := s.DeleteAuthorizer(ctx, "suite1", "corpA"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetAuthorizer(ctx, "suite1", "corpA"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_Concurrent(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			_ = s.PutSuiteTicket(ctx, "suite1", "T")
			_, _ = s.GetSuiteTicket(ctx, "suite1")
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}
```

- [ ] **Step 2.2: Run tests — expect compile failure**

Run: `go test ./work-wechat/isv/...`
Expected: FAIL (undefined: NewMemoryStore / AuthorizerTokens / Store methods)

- [ ] **Step 2.3: Implement `store.go`**

```go
package isv

import (
	"context"
	"sync"
	"time"
)

// AuthorizerTokens 打包单个被授权企业的凭证信息。
type AuthorizerTokens struct {
	CorpID            string
	PermanentCode     string
	CorpAccessToken   string
	CorpTokenExpireAt time.Time
}

// Store 负责持久化 ISV 认证流程中的各类凭证。
//
// 所有方法的第一 key 都是 suiteID,允许多 Client 共享同一 Store。
type Store interface {
	GetSuiteTicket(ctx context.Context, suiteID string) (string, error)
	PutSuiteTicket(ctx context.Context, suiteID, ticket string) error

	GetSuiteToken(ctx context.Context, suiteID string) (token string, expiresAt time.Time, err error)
	PutSuiteToken(ctx context.Context, suiteID, token string, expiresAt time.Time) error

	GetProviderToken(ctx context.Context, suiteID string) (token string, expiresAt time.Time, err error)
	PutProviderToken(ctx context.Context, suiteID, token string, expiresAt time.Time) error

	GetAuthorizer(ctx context.Context, suiteID, corpID string) (*AuthorizerTokens, error)
	PutAuthorizer(ctx context.Context, suiteID, corpID string, tokens *AuthorizerTokens) error
	DeleteAuthorizer(ctx context.Context, suiteID, corpID string) error
	ListAuthorizers(ctx context.Context, suiteID string) ([]string, error)
}

// ---- MemoryStore ----

type tokenEntry struct {
	value     string
	expiresAt time.Time
}

// MemoryStore 是 Store 的进程内默认实现,线程安全。
type MemoryStore struct {
	mu            sync.RWMutex
	suiteTickets  map[string]string               // suiteID → ticket
	suiteTokens   map[string]tokenEntry           // suiteID → token
	providerToks  map[string]tokenEntry           // suiteID → provider_token
	authorizers   map[string]map[string]*AuthorizerTokens // suiteID → corpID → tokens
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		suiteTickets: make(map[string]string),
		suiteTokens:  make(map[string]tokenEntry),
		providerToks: make(map[string]tokenEntry),
		authorizers:  make(map[string]map[string]*AuthorizerTokens),
	}
}

func (m *MemoryStore) GetSuiteTicket(_ context.Context, suiteID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.suiteTickets[suiteID]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (m *MemoryStore) PutSuiteTicket(_ context.Context, suiteID, ticket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.suiteTickets[suiteID] = ticket
	return nil
}

func (m *MemoryStore) GetSuiteToken(_ context.Context, suiteID string) (string, time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.suiteTokens[suiteID]
	if !ok {
		return "", time.Time{}, ErrNotFound
	}
	return e.value, e.expiresAt, nil
}

func (m *MemoryStore) PutSuiteToken(_ context.Context, suiteID, token string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.suiteTokens[suiteID] = tokenEntry{value: token, expiresAt: expiresAt}
	return nil
}

func (m *MemoryStore) GetProviderToken(_ context.Context, suiteID string) (string, time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.providerToks[suiteID]
	if !ok {
		return "", time.Time{}, ErrNotFound
	}
	return e.value, e.expiresAt, nil
}

func (m *MemoryStore) PutProviderToken(_ context.Context, suiteID, token string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providerToks[suiteID] = tokenEntry{value: token, expiresAt: expiresAt}
	return nil
}

func (m *MemoryStore) GetAuthorizer(_ context.Context, suiteID, corpID string) (*AuthorizerTokens, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	inner, ok := m.authorizers[suiteID]
	if !ok {
		return nil, ErrNotFound
	}
	v, ok := inner[corpID]
	if !ok {
		return nil, ErrNotFound
	}
	// 复制一份避免调用方修改污染内部状态
	cp := *v
	return &cp, nil
}

func (m *MemoryStore) PutAuthorizer(_ context.Context, suiteID, corpID string, tokens *AuthorizerTokens) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	inner, ok := m.authorizers[suiteID]
	if !ok {
		inner = make(map[string]*AuthorizerTokens)
		m.authorizers[suiteID] = inner
	}
	cp := *tokens
	inner[corpID] = &cp
	return nil
}

func (m *MemoryStore) DeleteAuthorizer(_ context.Context, suiteID, corpID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if inner, ok := m.authorizers[suiteID]; ok {
		delete(inner, corpID)
	}
	return nil
}

func (m *MemoryStore) ListAuthorizers(_ context.Context, suiteID string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	inner, ok := m.authorizers[suiteID]
	if !ok {
		return nil, nil
	}
	out := make([]string, 0, len(inner))
	for k := range inner {
		out = append(out, k)
	}
	return out, nil
}
```

- [ ] **Step 2.4: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run TestMemoryStore -v`
Expected: PASS for all 7 cases

- [ ] **Step 2.5: Commit**

```bash
git add work-wechat/isv/store.go work-wechat/isv/store_test.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add Store interface + MemoryStore

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: Client / Config / Options / HTTP 助手

**Files:**
- Create: `work-wechat/isv/client.go`
- Create: `work-wechat/isv/client_test.go`

- [ ] **Step 3.1: Write failing tests for NewClient + helpers**

Create `work-wechat/isv/client_test.go`:

```go
package isv

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Fake 43-char EncodingAESKey for tests.
const testAESKey = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ"

func testConfig() Config {
	return Config{
		SuiteID:        "suite1",
		SuiteSecret:    "secret1",
		Token:          "TOKEN",
		EncodingAESKey: testAESKey,
	}
}

func TestNewClient_RequiredFields(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*Config)
	}{
		{"suite id missing", func(c *Config) { c.SuiteID = "" }},
		{"suite secret missing", func(c *Config) { c.SuiteSecret = "" }},
		{"token missing", func(c *Config) { c.Token = "" }},
		{"aes key missing", func(c *Config) { c.EncodingAESKey = "" }},
		{"aes key wrong length", func(c *Config) { c.EncodingAESKey = "tooShort" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := testConfig()
			c.mut(&cfg)
			if _, err := NewClient(cfg); err == nil {
				t.Fatal("want error")
			}
		})
	}
}

func TestNewClient_ProviderPartial(t *testing.T) {
	cfg := testConfig()
	cfg.ProviderCorpID = "wx1"
	// missing ProviderSecret → error
	if _, err := NewClient(cfg); err == nil {
		t.Fatal("want error for partial provider config")
	}
	cfg.ProviderSecret = "psecret"
	if _, err := NewClient(cfg); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestNewClient_Options(t *testing.T) {
	customStore := NewMemoryStore()
	customHTTP := &http.Client{}
	c, err := NewClient(testConfig(),
		WithStore(customStore),
		WithHTTPClient(customHTTP),
		WithBaseURL("https://example.test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if c.store != customStore {
		t.Error("store not applied")
	}
	if c.http != customHTTP {
		t.Error("http client not applied")
	}
	if c.baseURL != "https://example.test" {
		t.Error("base url not applied")
	}
}

func TestNewClient_Defaults(t *testing.T) {
	c, err := NewClient(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	if c.store == nil {
		t.Error("default store missing")
	}
	if c.http == nil {
		t.Error("default http client missing")
	}
	if c.baseURL != "https://qyapi.weixin.qq.com" {
		t.Errorf("default baseURL wrong: %q", c.baseURL)
	}
}

func TestDoPost_WeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 42001,
			"errmsg":  "access_token expired",
		})
	}))
	defer srv.Close()

	c, err := NewClient(testConfig(), WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	// pre-seed a suite token so doPost has something to inject
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "TOK", timePlus(3600))

	var out struct{}
	err = c.doPost(context.Background(), "/cgi-bin/whatever", map[string]string{"x": "y"}, &out)
	var we *WeixinError
	if !errors.As(err, &we) || we.ErrCode != 42001 {
		t.Fatalf("want *WeixinError 42001, got %v", err)
	}
}

func TestDoPost_AttachesTokenQuery(t *testing.T) {
	var gotRawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c, err := NewClient(testConfig(), WithBaseURL(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "TOK42", timePlus(3600))

	var out struct{}
	if err := c.doPost(context.Background(), "/cgi-bin/test", map[string]string{}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotRawQuery, "suite_access_token=TOK42") {
		t.Fatalf("query missing token: %q", gotRawQuery)
	}
}
```

Note: `timePlus` is a small helper that returns `time.Now().Add(time.Duration(sec)*time.Second)`. Define it in `client_test.go`:

```go
import "time"

func timePlus(sec int) time.Time {
	return time.Now().Add(time.Duration(sec) * time.Second)
}
```

- [ ] **Step 3.2: Run tests — expect compile failure**

Run: `go test ./work-wechat/isv/... -run TestNewClient`
Expected: FAIL (undefined Config / Client / NewClient / Option...)

- [ ] **Step 3.3: Implement `client.go`**

```go
package isv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/godrealms/go-wechat-sdk/utils/wxcrypto"
)

const defaultBaseURL = "https://qyapi.weixin.qq.com"

// Config 是 ISV Client 的运行时配置。
type Config struct {
	SuiteID        string // 第三方应用 suite_id
	SuiteSecret    string // 第三方应用 suite_secret
	ProviderCorpID string // 服务商自己的 corpid(provider 接口需要,可选)
	ProviderSecret string // 服务商 provider_secret(provider 接口需要,可选)
	Token          string // 回调 token
	EncodingAESKey string // 回调 AES key(43 字符)
}

// Client 是服务商级别的入口,无状态可共享。
type Client struct {
	cfg     Config
	store   Store
	http    *http.Client
	crypto  *wxcrypto.MsgCrypto
	baseURL string

	suiteMu    sync.Mutex
	providerMu sync.Mutex
	corpMu     sync.Map // map[corpid]*sync.Mutex
}

// Option 是函数式配置项。
type Option func(*Client)

func WithStore(s Store) Option           { return func(c *Client) { c.store = s } }
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }
func WithBaseURL(u string) Option        { return func(c *Client) { c.baseURL = u } }

// NewClient 校验配置并构造 Client。
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if cfg.SuiteID == "" {
		return nil, fmt.Errorf("isv: SuiteID required")
	}
	if cfg.SuiteSecret == "" {
		return nil, fmt.Errorf("isv: SuiteSecret required")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("isv: Token required")
	}
	if len(cfg.EncodingAESKey) != 43 {
		return nil, fmt.Errorf("isv: EncodingAESKey must be 43 chars")
	}
	// ProviderCorpID 与 ProviderSecret 要么都填要么都空
	if (cfg.ProviderCorpID == "") != (cfg.ProviderSecret == "") {
		return nil, fmt.Errorf("isv: ProviderCorpID and ProviderSecret must both be set or both empty")
	}

	cry, err := wxcrypto.New(cfg.Token, cfg.EncodingAESKey, cfg.SuiteID)
	if err != nil {
		return nil, fmt.Errorf("isv: init crypto: %w", err)
	}

	c := &Client{
		cfg:     cfg,
		store:   NewMemoryStore(),
		http:    http.DefaultClient,
		crypto:  cry,
		baseURL: defaultBaseURL,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// ---- shared HTTP helpers ----

// doPost 发送 JSON POST 到 baseURL + path,query 自动注入 suite_access_token。
func (c *Client) doPost(ctx context.Context, path string, body, out interface{}) error {
	tok, err := c.GetSuiteAccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"suite_access_token": {tok}}
	return c.doPostRaw(ctx, path, q, body, out)
}

// doGet 发送 GET 到 baseURL + path,query 自动注入 suite_access_token。
func (c *Client) doGet(ctx context.Context, path string, extra url.Values, out interface{}) error {
	tok, err := c.GetSuiteAccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{}
	for k, v := range extra {
		q[k] = v
	}
	q.Set("suite_access_token", tok)
	return c.doRequestRaw(ctx, http.MethodGet, path, q, nil, out)
}

// doPostRaw 不自动获取 suite_token,query 由调用方完全控制。
func (c *Client) doPostRaw(ctx context.Context, path string, query url.Values, body, out interface{}) error {
	var buf io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("isv: marshal body: %w", err)
		}
		buf = bytes.NewReader(raw)
	}
	return c.doRequestRaw(ctx, http.MethodPost, path, query, buf, out)
}

func (c *Client) doRequestRaw(ctx context.Context, method, path string, query url.Values, body io.Reader, out interface{}) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return fmt.Errorf("isv: new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("isv: http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("isv: read body: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("isv: http %d: %s", resp.StatusCode, string(raw))
	}
	return decodeRaw(raw, out)
}

// decodeRaw 实现"两阶段解码":先检查 errcode,再 unmarshal 到 out。
func decodeRaw(raw []byte, out interface{}) error {
	var probe struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return fmt.Errorf("isv: decode errcode: %w", err)
	}
	if probe.ErrCode != 0 {
		return &WeixinError{ErrCode: probe.ErrCode, ErrMsg: probe.ErrMsg}
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("isv: decode body: %w", err)
	}
	return nil
}
```

**Temporary stub** — `GetSuiteAccessToken` is called by `doPost`/`doGet` but implemented in Task 4. For Task 3 to compile, add this stub at the bottom of `client.go`:

```go
// GetSuiteAccessToken is implemented in suite.token.go (Task 4).
// During Task 3 we provide a minimal stub that reads directly from Store so
// client_test.go's doPost test can pre-seed a token. Task 4 replaces the body.
func (c *Client) GetSuiteAccessToken(ctx context.Context) (string, error) {
	tok, _, err := c.store.GetSuiteToken(ctx, c.cfg.SuiteID)
	if err != nil {
		return "", err
	}
	return tok, nil
}
```

- [ ] **Step 3.4: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run "TestNewClient|TestDoPost"`
Expected: PASS for all 7 sub-cases

- [ ] **Step 3.5: Commit**

```bash
git add work-wechat/isv/client.go work-wechat/isv/client_test.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add Client, Config, Options, HTTP helpers

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: suite_access_token 管理

**Files:**
- Create: `work-wechat/isv/suite.token.go`
- Create: `work-wechat/isv/suite.token_test.go`
- Create: `work-wechat/isv/struct.suite.go`
- Modify: `work-wechat/isv/client.go` — 删除 Task 3 留下的 `GetSuiteAccessToken` 存根(Task 4 的实现覆盖它)

- [ ] **Step 4.1: Write failing test**

Create `work-wechat/isv/suite.token_test.go`:

```go
package isv

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// newTestISVClient 建立一个指向 baseURL 的 Client,store 中预种子 suite_ticket。
func newTestISVClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := NewClient(testConfig(), WithBaseURL(baseURL))
	if err != nil {
		t.Fatal(err)
	}
	_ = c.store.PutSuiteTicket(context.Background(), "suite1", "TICKET")
	return c
}

func TestGetSuiteAccessToken_FirstAndCached(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if r.URL.Path != "/cgi-bin/service/get_suite_token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["suite_id"] != "suite1" || body["suite_ticket"] != "TICKET" {
			t.Errorf("unexpected body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"suite_access_token": "STOK",
			"expires_in":         7200,
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	ctx := context.Background()

	tok, err := c.GetSuiteAccessToken(ctx)
	if err != nil || tok != "STOK" {
		t.Fatalf("got %q err=%v", tok, err)
	}
	// second call — cached, no HTTP hit
	tok2, err := c.GetSuiteAccessToken(ctx)
	if err != nil || tok2 != "STOK" {
		t.Fatalf("got %q err=%v", tok2, err)
	}
	if hits != 1 {
		t.Errorf("want 1 HTTP hit, got %d", hits)
	}
}

func TestGetSuiteAccessToken_MissingTicket(t *testing.T) {
	c, err := NewClient(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.GetSuiteAccessToken(context.Background())
	if !errors.Is(err, ErrSuiteTicketMissing) {
		t.Fatalf("want ErrSuiteTicketMissing, got %v", err)
	}
}

func TestGetSuiteAccessToken_ExpiredRefreshes(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"suite_access_token": "FRESH",
			"expires_in":         7200,
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	// seed an expired token
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "STALE", time.Now().Add(-1*time.Second))

	tok, err := c.GetSuiteAccessToken(context.Background())
	if err != nil || tok != "FRESH" {
		t.Fatalf("got %q err=%v", tok, err)
	}
	if hits != 1 {
		t.Errorf("want 1 HTTP hit, got %d", hits)
	}
}

func TestRefreshSuiteToken_Explicit(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"suite_access_token": "NEW",
			"expires_in":         7200,
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	ctx := context.Background()
	// pre-seed a still-valid token; RefreshSuiteToken should ignore cache.
	_ = c.store.PutSuiteToken(ctx, "suite1", "OLD", time.Now().Add(time.Hour))

	if err := c.RefreshSuiteToken(ctx); err != nil {
		t.Fatal(err)
	}
	tok, _, _ := c.store.GetSuiteToken(ctx, "suite1")
	if tok != "NEW" {
		t.Errorf("want NEW, got %q", tok)
	}
	if hits != 1 {
		t.Errorf("want 1 HTTP hit, got %d", hits)
	}
}

func TestGetSuiteAccessToken_WeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 40001,
			"errmsg":  "invalid credential",
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	_, err := c.GetSuiteAccessToken(context.Background())
	var we *WeixinError
	if !errors.As(err, &we) || we.ErrCode != 40001 {
		t.Fatalf("want *WeixinError 40001, got %v", err)
	}
}
```

- [ ] **Step 4.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run TestSuiteAccessToken -v`
Expected: FAIL (GetSuiteAccessToken stub returns ErrNotFound not ErrSuiteTicketMissing; no HTTP call; no RefreshSuiteToken method)

- [ ] **Step 4.3: Write `struct.suite.go`**

```go
package isv

// SuiteAccessTokenResp 是 service/get_suite_token 的响应体。
type SuiteAccessTokenResp struct {
	SuiteAccessToken string `json:"suite_access_token"`
	ExpiresIn        int    `json:"expires_in"`
}
```

- [ ] **Step 4.4: Implement `suite.token.go`**

```go
package isv

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// safetyMargin 是 token 过期前的安全窗口,避免临界时刻使用即将过期的 token。
const safetyMargin = 5 * time.Minute

// GetSuiteAccessToken 返回 suite_access_token。
// 策略:lazy + 双检锁。Store 命中且未到安全窗口直接返回;否则加锁再查一次,仍过期则刷新。
func (c *Client) GetSuiteAccessToken(ctx context.Context) (string, error) {
	if tok, ok, err := c.readValidSuiteToken(ctx); err != nil {
		return "", err
	} else if ok {
		return tok, nil
	}

	c.suiteMu.Lock()
	defer c.suiteMu.Unlock()

	if tok, ok, err := c.readValidSuiteToken(ctx); err != nil {
		return "", err
	} else if ok {
		return tok, nil
	}

	return c.fetchSuiteTokenLocked(ctx)
}

// RefreshSuiteToken 强制刷新 suite_access_token(忽略缓存)。
func (c *Client) RefreshSuiteToken(ctx context.Context) error {
	c.suiteMu.Lock()
	defer c.suiteMu.Unlock()
	_, err := c.fetchSuiteTokenLocked(ctx)
	return err
}

func (c *Client) readValidSuiteToken(ctx context.Context) (string, bool, error) {
	tok, exp, err := c.store.GetSuiteToken(ctx, c.cfg.SuiteID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", false, nil
		}
		return "", false, err
	}
	if time.Until(exp) <= safetyMargin {
		return "", false, nil
	}
	return tok, true, nil
}

// fetchSuiteTokenLocked 发起一次 HTTP,写回 Store,返回新 token。
// 调用方必须已持有 c.suiteMu。
func (c *Client) fetchSuiteTokenLocked(ctx context.Context) (string, error) {
	ticket, err := c.store.GetSuiteTicket(ctx, c.cfg.SuiteID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", ErrSuiteTicketMissing
		}
		return "", fmt.Errorf("isv: read suite_ticket: %w", err)
	}

	body := map[string]string{
		"suite_id":     c.cfg.SuiteID,
		"suite_secret": c.cfg.SuiteSecret,
		"suite_ticket": ticket,
	}
	var resp SuiteAccessTokenResp
	// get_suite_token 不需要附带 access_token query
	if err := c.doPostRaw(ctx, "/cgi-bin/service/get_suite_token", url.Values{}, body, &resp); err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(time.Duration(resp.ExpiresIn)*time.Second - safetyMargin)
	if err := c.store.PutSuiteToken(ctx, c.cfg.SuiteID, resp.SuiteAccessToken, expiresAt); err != nil {
		return "", fmt.Errorf("isv: persist suite_token: %w", err)
	}
	return resp.SuiteAccessToken, nil
}
```

- [ ] **Step 4.5: Remove stub from `client.go`**

Delete the `GetSuiteAccessToken` stub block at the bottom of `client.go` (the block comment "is implemented in suite.token.go (Task 4)..."). The real implementation now lives in `suite.token.go`.

- [ ] **Step 4.6: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run "TestGetSuiteAccessToken|TestRefreshSuiteToken" -v`
Expected: all 5 cases PASS

Run: `go test ./work-wechat/isv/... -v`
Expected: all prior tests still green

- [ ] **Step 4.7: Commit**

```bash
git add work-wechat/isv/suite.token.go work-wechat/isv/suite.token_test.go work-wechat/isv/struct.suite.go work-wechat/isv/client.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add suite_access_token lifecycle

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: pre_auth_code & SetSessionInfo & AuthorizeURL

**Files:**
- Create: `work-wechat/isv/suite.preauth.go`
- Create: `work-wechat/isv/suite.preauth_test.go`
- Modify: `work-wechat/isv/struct.suite.go` — 追加 `PreAuthCodeResp` / `SessionInfo`

- [ ] **Step 5.1: Write failing test**

Create `work-wechat/isv/suite.preauth_test.go`:

```go
package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func seedSuiteToken(t *testing.T, c *Client) {
	t.Helper()
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "STOK", time.Now().Add(time.Hour))
}

func TestGetPreAuthCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/service/get_pre_auth_code" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("suite_access_token"); got != "STOK" {
			t.Errorf("token query: %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"pre_auth_code": "PCODE",
			"expires_in":    1200,
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	seedSuiteToken(t, c)

	resp, err := c.GetPreAuthCode(context.Background())
	if err != nil || resp.PreAuthCode != "PCODE" {
		t.Fatalf("got %+v err=%v", resp, err)
	}
}

func TestSetSessionInfo(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	seedSuiteToken(t, c)

	info := &SessionInfo{AppID: []int{1, 2}, AuthType: 1}
	if err := c.SetSessionInfo(context.Background(), "PCODE", info); err != nil {
		t.Fatal(err)
	}
	if gotBody["pre_auth_code"] != "PCODE" {
		t.Errorf("pre_auth_code missing: %+v", gotBody)
	}
	sess, ok := gotBody["session_info"].(map[string]interface{})
	if !ok {
		t.Fatalf("session_info missing: %+v", gotBody)
	}
	if sess["auth_type"].(float64) != 1 {
		t.Errorf("auth_type wrong: %+v", sess)
	}
}

func TestAuthorizeURL(t *testing.T) {
	c, err := NewClient(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	got := c.AuthorizeURL("PCODE", "https://cb.example/ret", "state1")
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "https://open.work.weixin.qq.com/3rdapp/install?") {
		t.Errorf("url prefix: %q", got)
	}
	q := u.Query()
	if q.Get("suite_id") != "suite1" || q.Get("pre_auth_code") != "PCODE" ||
		q.Get("redirect_uri") != "https://cb.example/ret" || q.Get("state") != "state1" {
		t.Errorf("query: %+v", q)
	}
}
```

- [ ] **Step 5.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run "TestGetPreAuthCode|TestSetSessionInfo|TestAuthorizeURL"`
Expected: FAIL (undefined)

- [ ] **Step 5.3: Append DTOs to `struct.suite.go`**

Add at the end of `work-wechat/isv/struct.suite.go`:

```go
// PreAuthCodeResp 是 service/get_pre_auth_code 的响应。
type PreAuthCodeResp struct {
	PreAuthCode string `json:"pre_auth_code"`
	ExpiresIn   int    `json:"expires_in"`
}

// SessionInfo 是 service/set_session_info 的 session_info 字段。
type SessionInfo struct {
	AppID    []int `json:"appid,omitempty"`     // 限定授权的应用 ID 列表
	AuthType int   `json:"auth_type,omitempty"` // 0=管理员授权,1=成员授权
}
```

- [ ] **Step 5.4: Implement `suite.preauth.go`**

```go
package isv

import (
	"context"
	"net/url"
)

// GetPreAuthCode 拉取预授权码。
func (c *Client) GetPreAuthCode(ctx context.Context) (*PreAuthCodeResp, error) {
	var resp PreAuthCodeResp
	if err := c.doGet(ctx, "/cgi-bin/service/get_pre_auth_code", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetSessionInfo 为指定 pre_auth_code 绑定授权会话配置。
func (c *Client) SetSessionInfo(ctx context.Context, preAuthCode string, info *SessionInfo) error {
	body := map[string]interface{}{
		"pre_auth_code": preAuthCode,
		"session_info":  info,
	}
	return c.doPost(ctx, "/cgi-bin/service/set_session_info", body, nil)
}

// AuthorizeURL 拼接企业管理员扫码授权的 PC 跳转 URL(不发起 HTTP)。
func (c *Client) AuthorizeURL(preAuthCode, redirectURI, state string) string {
	q := url.Values{
		"suite_id":      {c.cfg.SuiteID},
		"pre_auth_code": {preAuthCode},
		"redirect_uri":  {redirectURI},
		"state":         {state},
	}
	return "https://open.work.weixin.qq.com/3rdapp/install?" + q.Encode()
}
```

- [ ] **Step 5.5: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -v`
Expected: all green

- [ ] **Step 5.6: Commit**

```bash
git add work-wechat/isv/suite.preauth.go work-wechat/isv/suite.preauth_test.go work-wechat/isv/struct.suite.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add pre_auth_code, SetSessionInfo, AuthorizeURL

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: 永久授权码 & GetAuthInfo & GetAdminList

**Files:**
- Create: `work-wechat/isv/suite.permanent.go`
- Create: `work-wechat/isv/suite.permanent_test.go`
- Create: `work-wechat/isv/struct.permanent.go`

- [ ] **Step 6.1: Write failing test**

Create `work-wechat/isv/suite.permanent_test.go`:

```go
package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetPermanentCode_StoresAuthorizer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":   "CORP_TOK",
			"expires_in":     7200,
			"permanent_code": "PERM",
			"auth_corp_info": map[string]interface{}{
				"corpid":    "wxcorp1",
				"corp_name": "ACME",
			},
			"auth_info": map[string]interface{}{
				"agent": []map[string]interface{}{
					{"agentid": 1000001, "name": "HR"},
				},
			},
			"auth_user_info": map[string]interface{}{
				"userid": "admin",
				"name":   "Admin",
			},
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "STOK", time.Now().Add(time.Hour))

	resp, err := c.GetPermanentCode(context.Background(), "auth_code_xyz")
	if err != nil {
		t.Fatal(err)
	}
	if resp.PermanentCode != "PERM" || resp.AuthCorpInfo.CorpID != "wxcorp1" {
		t.Fatalf("resp: %+v", resp)
	}

	// Verify AuthorizerTokens written to store
	got, err := c.store.GetAuthorizer(context.Background(), "suite1", "wxcorp1")
	if err != nil {
		t.Fatal(err)
	}
	if got.PermanentCode != "PERM" || got.CorpAccessToken != "CORP_TOK" {
		t.Fatalf("stored: %+v", got)
	}
}

func TestGetAuthInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["auth_corpid"] != "wxcorp1" || body["permanent_code"] != "PERM" {
			t.Errorf("body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"auth_corp_info": map[string]interface{}{
				"corpid":    "wxcorp1",
				"corp_name": "ACME",
			},
			"auth_info": map[string]interface{}{
				"agent": []map[string]interface{}{{"agentid": 1}},
			},
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "STOK", time.Now().Add(time.Hour))

	resp, err := c.GetAuthInfo(context.Background(), "wxcorp1", "PERM")
	if err != nil || resp.AuthCorpInfo.CorpID != "wxcorp1" {
		t.Fatalf("got %+v err=%v", resp, err)
	}
}

func TestGetAdminList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["auth_corpid"] != "wxcorp1" {
			t.Errorf("body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"admin": []map[string]interface{}{
				{"userid": "u1", "open_userid": "o1", "auth_type": 1},
			},
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "STOK", time.Now().Add(time.Hour))

	resp, err := c.GetAdminList(context.Background(), "wxcorp1", "1000001")
	if err != nil || len(resp.Admin) != 1 || resp.Admin[0].AuthType != 1 {
		t.Fatalf("got %+v err=%v", resp, err)
	}
}
```

- [ ] **Step 6.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run "TestGetPermanentCode|TestGetAuthInfo|TestGetAdminList"`
Expected: FAIL (undefined)

- [ ] **Step 6.3: Write `struct.permanent.go`**

```go
package isv

// ---------- get_permanent_code / get_auth_info 共用 ----------

// AuthCorpInfo 授权企业信息。
type AuthCorpInfo struct {
	CorpID            string `json:"corpid"`
	CorpName          string `json:"corp_name"`
	CorpType          string `json:"corp_type"`
	CorpSquareLogoURL string `json:"corp_square_logo_url"`
	CorpUserMax       int    `json:"corp_user_max"`
	CorpFullName      string `json:"corp_full_name"`
	VerifiedEndTime   int64  `json:"verified_end_time"`
	SubjectType       int    `json:"subject_type"`
	CorpWxqrcode      string `json:"corp_wxqrcode"`
	CorpScale         string `json:"corp_scale"`
	CorpIndustry      string `json:"corp_industry"`
	CorpSubIndustry   string `json:"corp_sub_industry"`
	Location          string `json:"location"`
}

// AgentPrivilege 应用可见范围等权限信息。
type AgentPrivilege struct {
	Level      int      `json:"level"`
	AllowParty []int    `json:"allow_party,omitempty"`
	AllowUser  []string `json:"allow_user,omitempty"`
	AllowTag   []int    `json:"allow_tag,omitempty"`
	ExtraParty []int    `json:"extra_party,omitempty"`
	ExtraUser  []string `json:"extra_user,omitempty"`
	ExtraTag   []int    `json:"extra_tag,omitempty"`
}

// SharedFromInfo 共享应用来源信息。
type SharedFromInfo struct {
	CorpID    string `json:"corpid"`
	ShareType int    `json:"share_type"`
}

// AuthAgent 授权应用信息。
type AuthAgent struct {
	AgentID       int             `json:"agentid"`
	Name          string          `json:"name"`
	RoundLogoURL  string          `json:"round_logo_url"`
	SquareLogoURL string          `json:"square_logo_url"`
	AppID         int             `json:"appid"`
	Privilege     AgentPrivilege  `json:"privilege,omitempty"`
	SharedFrom    *SharedFromInfo `json:"shared_from,omitempty"`
}

// AuthInfoAgent 授权应用列表。
type AuthInfoAgent struct {
	Agent []AuthAgent `json:"agent"`
}

// AuthUserInfo 授权管理员信息。
type AuthUserInfo struct {
	UserID     string `json:"userid"`
	OpenUserID string `json:"open_userid"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
}

// PermanentCodeResp 是 service/get_permanent_code 的响应。
type PermanentCodeResp struct {
	AccessToken   string        `json:"access_token"`
	ExpiresIn     int           `json:"expires_in"`
	PermanentCode string        `json:"permanent_code"`
	AuthCorpInfo  AuthCorpInfo  `json:"auth_corp_info"`
	AuthInfo      AuthInfoAgent `json:"auth_info"`
	AuthUserInfo  AuthUserInfo  `json:"auth_user_info"`
}

// AuthInfoResp 是 service/get_auth_info 的响应。
type AuthInfoResp struct {
	AuthCorpInfo AuthCorpInfo  `json:"auth_corp_info"`
	AuthInfo     AuthInfoAgent `json:"auth_info"`
}

// ---------- get_admin_list ----------

type AdminInfo struct {
	UserID     string `json:"userid"`
	OpenUserID string `json:"open_userid"`
	AuthType   int    `json:"auth_type"` // 0=普通管理员 1=超级管理员
}

type AdminListResp struct {
	Admin []AdminInfo `json:"admin"`
}
```

- [ ] **Step 6.4: Implement `suite.permanent.go`**

```go
package isv

import (
	"context"
	"time"
)

// GetPermanentCode 用 auth_code 换取企业永久授权码并自动持久化 AuthorizerTokens。
func (c *Client) GetPermanentCode(ctx context.Context, authCode string) (*PermanentCodeResp, error) {
	body := map[string]string{"auth_code": authCode}
	var resp PermanentCodeResp
	if err := c.doPost(ctx, "/cgi-bin/service/get_permanent_code", body, &resp); err != nil {
		return nil, err
	}

	// 首次同时拿到 corp_token,写入 Store
	if resp.AccessToken != "" && resp.AuthCorpInfo.CorpID != "" {
		expiresAt := time.Now().Add(time.Duration(resp.ExpiresIn)*time.Second - safetyMargin)
		tokens := &AuthorizerTokens{
			CorpID:            resp.AuthCorpInfo.CorpID,
			PermanentCode:     resp.PermanentCode,
			CorpAccessToken:   resp.AccessToken,
			CorpTokenExpireAt: expiresAt,
		}
		if err := c.store.PutAuthorizer(ctx, c.cfg.SuiteID, resp.AuthCorpInfo.CorpID, tokens); err != nil {
			return nil, err
		}
	}
	return &resp, nil
}

// GetAuthInfo 查询企业授权信息(不缓存)。
func (c *Client) GetAuthInfo(ctx context.Context, corpID, permanentCode string) (*AuthInfoResp, error) {
	body := map[string]string{
		"auth_corpid":    corpID,
		"permanent_code": permanentCode,
	}
	var resp AuthInfoResp
	if err := c.doPost(ctx, "/cgi-bin/service/get_auth_info", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAdminList 获取授权应用的管理员列表。
func (c *Client) GetAdminList(ctx context.Context, corpID, agentID string) (*AdminListResp, error) {
	body := map[string]string{
		"auth_corpid": corpID,
		"agentid":     agentID,
	}
	var resp AdminListResp
	if err := c.doPost(ctx, "/cgi-bin/service/get_admin_list", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 6.5: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run "TestGetPermanentCode|TestGetAuthInfo|TestGetAdminList" -v`
Expected: PASS

Run: `go test ./work-wechat/isv/...`
Expected: all prior tests still green

- [ ] **Step 6.6: Commit**

```bash
git add work-wechat/isv/suite.permanent.go work-wechat/isv/suite.permanent_test.go work-wechat/isv/struct.permanent.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add GetPermanentCode, GetAuthInfo, GetAdminList

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 7: provider_access_token + ID 转换(2 方法)

**Files:**
- Create: `work-wechat/isv/provider.id_convert.go`
- Create: `work-wechat/isv/provider.id_convert_test.go`
- Modify: `work-wechat/isv/struct.permanent.go` — 追加 `UserIDConvertResp` / `UserIDOpenUserIDPair`

- [ ] **Step 7.1: Write failing test**

Create `work-wechat/isv/provider.id_convert_test.go`:

```go
package isv

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// newTestISVClientWithProvider 构造启用了 provider 字段的 Client。
func newTestISVClientWithProvider(t *testing.T, baseURL string) *Client {
	t.Helper()
	cfg := testConfig()
	cfg.ProviderCorpID = "wxprov"
	cfg.ProviderSecret = "PSEC"
	c, err := NewClient(cfg, WithBaseURL(baseURL))
	if err != nil {
		t.Fatal(err)
	}
	_ = c.store.PutSuiteTicket(context.Background(), "suite1", "TICKET")
	return c
}

func TestCorpIDToOpenCorpID(t *testing.T) {
	var providerHits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			atomic.AddInt32(&providerHits, 1)
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["corpid"] != "wxprov" || body["provider_secret"] != "PSEC" {
				t.Errorf("provider token body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/corpid_to_opencorpid":
			if got := r.URL.Query().Get("provider_access_token"); got != "PTOK" {
				t.Errorf("token query: %q", got)
			}
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["corpid"] != "wxcorp1" {
				t.Errorf("body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"open_corpid": "openWx1",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)

	got, err := c.CorpIDToOpenCorpID(context.Background(), "wxcorp1")
	if err != nil || got != "openWx1" {
		t.Fatalf("got %q err=%v", got, err)
	}
	// second call — provider token cached
	got2, _ := c.CorpIDToOpenCorpID(context.Background(), "wxcorp1")
	if got2 != "openWx1" || providerHits != 1 {
		t.Errorf("provider cache failed: hits=%d", providerHits)
	}
}

func TestUserIDToOpenUserID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/batch/userid_to_openuserid":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"open_userid_list": []map[string]string{
					{"userid": "u1", "open_userid": "o1"},
				},
				"invalid_userid_list": []string{"u_bad"},
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.UserIDToOpenUserID(context.Background(), "wxcorp1", []string{"u1", "u_bad"})
	if err != nil || len(resp.OpenUserIDList) != 1 || resp.InvalidUserIDList[0] != "u_bad" {
		t.Fatalf("got %+v err=%v", resp, err)
	}
}

func TestProviderNotConfigured(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	_, err := c.CorpIDToOpenCorpID(context.Background(), "wxcorp1")
	if !errors.Is(err, ErrProviderCorpIDMissing) && !errors.Is(err, ErrProviderSecretMissing) {
		t.Fatalf("want provider missing error, got %v", err)
	}
}

func TestProviderTokenExpiredRefresh(t *testing.T) {
	var providerHits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/service/get_provider_token" {
			atomic.AddInt32(&providerHits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "FRESH",
				"expires_in":            7200,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"open_corpid": "o1"})
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	_ = c.store.PutProviderToken(context.Background(), "suite1", "STALE", time.Now().Add(-1*time.Second))

	if _, err := c.CorpIDToOpenCorpID(context.Background(), "wxcorp1"); err != nil {
		t.Fatal(err)
	}
	if providerHits != 1 {
		t.Errorf("want 1 provider hit, got %d", providerHits)
	}
}
```

- [ ] **Step 7.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run "TestCorpID|TestUserIDTo|TestProvider"`
Expected: FAIL (undefined)

- [ ] **Step 7.3: Append DTOs to `struct.permanent.go`**

Add at the end of `work-wechat/isv/struct.permanent.go`:

```go
// ---------- provider ID convert ----------

// UserIDOpenUserIDPair 是 userid ↔ open_userid 的一对。
type UserIDOpenUserIDPair struct {
	UserID     string `json:"userid"`
	OpenUserID string `json:"open_userid"`
}

// UserIDConvertResp 是 service/batch/userid_to_openuserid 的响应。
type UserIDConvertResp struct {
	OpenUserIDList    []UserIDOpenUserIDPair `json:"open_userid_list"`
	InvalidUserIDList []string               `json:"invalid_userid_list"`
}
```

- [ ] **Step 7.4: Implement `provider.id_convert.go`**

```go
package isv

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// getProviderAccessToken 返回 provider_access_token(lazy + 双检锁)。
// 校验 ProviderCorpID / ProviderSecret 是否配置,未配置返回哨兵错误。
func (c *Client) getProviderAccessToken(ctx context.Context) (string, error) {
	if c.cfg.ProviderCorpID == "" {
		return "", ErrProviderCorpIDMissing
	}
	if c.cfg.ProviderSecret == "" {
		return "", ErrProviderSecretMissing
	}

	if tok, ok, err := c.readValidProviderToken(ctx); err != nil {
		return "", err
	} else if ok {
		return tok, nil
	}

	c.providerMu.Lock()
	defer c.providerMu.Unlock()

	if tok, ok, err := c.readValidProviderToken(ctx); err != nil {
		return "", err
	} else if ok {
		return tok, nil
	}

	return c.fetchProviderTokenLocked(ctx)
}

func (c *Client) readValidProviderToken(ctx context.Context) (string, bool, error) {
	tok, exp, err := c.store.GetProviderToken(ctx, c.cfg.SuiteID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", false, nil
		}
		return "", false, err
	}
	if time.Until(exp) <= safetyMargin {
		return "", false, nil
	}
	return tok, true, nil
}

func (c *Client) fetchProviderTokenLocked(ctx context.Context) (string, error) {
	body := map[string]string{
		"corpid":          c.cfg.ProviderCorpID,
		"provider_secret": c.cfg.ProviderSecret,
	}
	var resp struct {
		ProviderAccessToken string `json:"provider_access_token"`
		ExpiresIn           int    `json:"expires_in"`
	}
	if err := c.doPostRaw(ctx, "/cgi-bin/service/get_provider_token", url.Values{}, body, &resp); err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(time.Duration(resp.ExpiresIn)*time.Second - safetyMargin)
	if err := c.store.PutProviderToken(ctx, c.cfg.SuiteID, resp.ProviderAccessToken, expiresAt); err != nil {
		return "", fmt.Errorf("isv: persist provider_token: %w", err)
	}
	return resp.ProviderAccessToken, nil
}

// providerDoPost 和 doPost 类似,只是注入的 token 是 provider_access_token。
func (c *Client) providerDoPost(ctx context.Context, path string, body, out interface{}) error {
	tok, err := c.getProviderAccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"provider_access_token": {tok}}
	return c.doPostRaw(ctx, path, q, body, out)
}

// CorpIDToOpenCorpID 把企业 corpid 转换成跨服务商匿名的 open_corpid。
func (c *Client) CorpIDToOpenCorpID(ctx context.Context, corpID string) (string, error) {
	body := map[string]string{"corpid": corpID}
	var resp struct {
		OpenCorpID string `json:"open_corpid"`
	}
	if err := c.providerDoPost(ctx, "/cgi-bin/service/corpid_to_opencorpid", body, &resp); err != nil {
		return "", err
	}
	return resp.OpenCorpID, nil
}

// UserIDToOpenUserID 批量把 userid 转换为跨服务商匿名的 open_userid。
func (c *Client) UserIDToOpenUserID(ctx context.Context, corpID string, userIDs []string) (*UserIDConvertResp, error) {
	body := map[string]interface{}{
		"auth_corpid":  corpID,
		"userid_list":  userIDs,
	}
	var resp UserIDConvertResp
	if err := c.providerDoPost(ctx, "/cgi-bin/service/batch/userid_to_openuserid", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 7.5: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run "TestCorpID|TestUserIDTo|TestProvider" -v`
Expected: PASS for all 4 cases

Run: `go test ./work-wechat/isv/...`
Expected: all prior tests still green

- [ ] **Step 7.6: Commit**

```bash
git add work-wechat/isv/provider.id_convert.go work-wechat/isv/provider.id_convert_test.go work-wechat/isv/struct.permanent.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add provider_token + corpid/userid open ID conversion

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 8: corp_token 生命周期 + CorpClient + RefreshAll

**Files:**
- Create: `work-wechat/isv/corp.token.go`
- Create: `work-wechat/isv/corp.token_test.go`
- Create: `work-wechat/isv/authorizer.go`
- Create: `work-wechat/isv/struct.corp.go`

- [ ] **Step 8.1: Write failing test**

Create `work-wechat/isv/corp.token_test.go`:

```go
package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetCorpToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/service/get_corp_token" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["auth_corpid"] != "wxcorp1" || body["permanent_code"] != "PERM" {
			t.Errorf("body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "CTOK1",
			"expires_in":   7200,
		})
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	_ = c.store.PutSuiteToken(context.Background(), "suite1", "STOK", time.Now().Add(time.Hour))

	resp, err := c.GetCorpToken(context.Background(), "wxcorp1", "PERM")
	if err != nil || resp.AccessToken != "CTOK1" {
		t.Fatalf("got %+v err=%v", resp, err)
	}
}

func TestCorpClient_AccessTokenLifecycle(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/service/get_corp_token" {
			atomic.AddInt32(&hits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "CTOK_FRESH",
				"expires_in":   7200,
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	ctx := context.Background()
	_ = c.store.PutSuiteToken(ctx, "suite1", "STOK", time.Now().Add(time.Hour))
	_ = c.store.PutAuthorizer(ctx, "suite1", "wxcorp1", &AuthorizerTokens{
		CorpID:            "wxcorp1",
		PermanentCode:     "PERM",
		CorpAccessToken:   "STALE",
		CorpTokenExpireAt: time.Now().Add(-1 * time.Second), // already expired
	})

	cc := c.CorpClient("wxcorp1")
	tok, err := cc.AccessToken(ctx)
	if err != nil || tok != "CTOK_FRESH" {
		t.Fatalf("got %q err=%v", tok, err)
	}
	// cached on second call
	tok2, _ := cc.AccessToken(ctx)
	if tok2 != "CTOK_FRESH" || hits != 1 {
		t.Errorf("cache miss: hits=%d", hits)
	}
}

func TestCorpClient_AccessTokenSingleFlight(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/service/get_corp_token" {
			atomic.AddInt32(&hits, 1)
			time.Sleep(20 * time.Millisecond) // simulate network latency
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "CTOK",
				"expires_in":   7200,
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	ctx := context.Background()
	_ = c.store.PutSuiteToken(ctx, "suite1", "STOK", time.Now().Add(time.Hour))
	_ = c.store.PutAuthorizer(ctx, "suite1", "wxcorp1", &AuthorizerTokens{
		CorpID:        "wxcorp1",
		PermanentCode: "PERM",
	})

	cc := c.CorpClient("wxcorp1")
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = cc.AccessToken(ctx)
		}()
	}
	wg.Wait()

	if hits != 1 {
		t.Errorf("want 1 HTTP hit (single-flight), got %d", hits)
	}
}

func TestCorpClient_Refresh(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/service/get_corp_token" {
			atomic.AddInt32(&hits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "CTOK_NEW",
				"expires_in":   7200,
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	ctx := context.Background()
	_ = c.store.PutSuiteToken(ctx, "suite1", "STOK", time.Now().Add(time.Hour))
	_ = c.store.PutAuthorizer(ctx, "suite1", "wxcorp1", &AuthorizerTokens{
		CorpID:            "wxcorp1",
		PermanentCode:     "PERM",
		CorpAccessToken:   "OLD",
		CorpTokenExpireAt: time.Now().Add(time.Hour), // still valid
	})

	if err := c.CorpClient("wxcorp1").Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	got, _ := c.store.GetAuthorizer(ctx, "suite1", "wxcorp1")
	if got.CorpAccessToken != "CTOK_NEW" || hits != 1 {
		t.Errorf("refresh failed: tok=%q hits=%d", got.CorpAccessToken, hits)
	}
}

func TestRefreshAll(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/service/get_corp_token" {
			atomic.AddInt32(&hits, 1)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "X",
				"expires_in":   7200,
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClient(t, srv.URL)
	ctx := context.Background()
	_ = c.store.PutSuiteToken(ctx, "suite1", "STOK", time.Now().Add(time.Hour))
	_ = c.store.PutAuthorizer(ctx, "suite1", "corpA", &AuthorizerTokens{CorpID: "corpA", PermanentCode: "P1"})
	_ = c.store.PutAuthorizer(ctx, "suite1", "corpB", &AuthorizerTokens{CorpID: "corpB", PermanentCode: "P2"})

	if err := c.RefreshAll(ctx); err != nil {
		t.Fatal(err)
	}
	if hits != 2 {
		t.Errorf("want 2 HTTP hits, got %d", hits)
	}
}

func TestCorpClient_ImplementsTokenSource(t *testing.T) {
	var _ TokenSource = (*CorpClient)(nil)
}
```

- [ ] **Step 8.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run "TestGetCorpToken|TestCorpClient|TestRefreshAll"`
Expected: FAIL (undefined)

- [ ] **Step 8.3: Write `struct.corp.go`**

```go
package isv

// CorpTokenResp 是 service/get_corp_token 的响应。
type CorpTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
```

- [ ] **Step 8.4: Write `authorizer.go`**

```go
package isv

// Compile-time assertion that CorpClient satisfies TokenSource.
var _ TokenSource = (*CorpClient)(nil)
```

- [ ] **Step 8.5: Implement `corp.token.go`**

```go
package isv

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// GetCorpToken 调用 service/get_corp_token 换取企业 corp_access_token(底层方法)。
// 不写 Store —— 调用方通常应该用 CorpClient.AccessToken。
func (c *Client) GetCorpToken(ctx context.Context, corpID, permanentCode string) (*CorpTokenResp, error) {
	body := map[string]string{
		"auth_corpid":    corpID,
		"permanent_code": permanentCode,
	}
	var resp CorpTokenResp
	if err := c.doPost(ctx, "/cgi-bin/service/get_corp_token", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CorpClient 代表"代某企业调用"的会话句柄,实现 TokenSource 接口。
type CorpClient struct {
	parent *Client
	corpID string
}

// CorpClient 工厂。
func (c *Client) CorpClient(corpID string) *CorpClient {
	return &CorpClient{parent: c, corpID: corpID}
}

// CorpID 返回当前 CorpClient 代理的企业 corpid。
func (cc *CorpClient) CorpID() string { return cc.corpID }

// AccessToken 返回企业 corp_access_token(lazy + 双检锁 + 单飞)。
func (cc *CorpClient) AccessToken(ctx context.Context) (string, error) {
	if tok, ok, err := cc.readValidCorpToken(ctx); err != nil {
		return "", err
	} else if ok {
		return tok, nil
	}

	lock := cc.parent.lockFor(cc.corpID)
	lock.Lock()
	defer lock.Unlock()

	if tok, ok, err := cc.readValidCorpToken(ctx); err != nil {
		return "", err
	} else if ok {
		return tok, nil
	}

	return cc.refreshLocked(ctx)
}

// Refresh 强制刷新单个企业的 corp_token(忽略缓存)。
func (cc *CorpClient) Refresh(ctx context.Context) error {
	lock := cc.parent.lockFor(cc.corpID)
	lock.Lock()
	defer lock.Unlock()
	_, err := cc.refreshLocked(ctx)
	return err
}

func (cc *CorpClient) readValidCorpToken(ctx context.Context) (string, bool, error) {
	tokens, err := cc.parent.store.GetAuthorizer(ctx, cc.parent.cfg.SuiteID, cc.corpID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", false, ErrAuthorizerRevoked
		}
		return "", false, err
	}
	if tokens.CorpAccessToken == "" || time.Until(tokens.CorpTokenExpireAt) <= safetyMargin {
		return "", false, nil
	}
	return tokens.CorpAccessToken, true, nil
}

// refreshLocked 在持有 lockFor(corpID) 的前提下,通过 permanent_code 换取新 corp_token 并写回 Store。
func (cc *CorpClient) refreshLocked(ctx context.Context) (string, error) {
	tokens, err := cc.parent.store.GetAuthorizer(ctx, cc.parent.cfg.SuiteID, cc.corpID)
	if err != nil {
		return "", err
	}
	resp, err := cc.parent.GetCorpToken(ctx, cc.corpID, tokens.PermanentCode)
	if err != nil {
		return "", err
	}
	tokens.CorpAccessToken = resp.AccessToken
	tokens.CorpTokenExpireAt = time.Now().Add(time.Duration(resp.ExpiresIn)*time.Second - safetyMargin)
	if err := cc.parent.store.PutAuthorizer(ctx, cc.parent.cfg.SuiteID, cc.corpID, tokens); err != nil {
		return "", fmt.Errorf("isv: persist corp token: %w", err)
	}
	return tokens.CorpAccessToken, nil
}

// lockFor 返回 corpid 专属的 mutex(从 sync.Map 取,首次创建)。
func (c *Client) lockFor(corpID string) *sync.Mutex {
	if v, ok := c.corpMu.Load(corpID); ok {
		return v.(*sync.Mutex)
	}
	v, _ := c.corpMu.LoadOrStore(corpID, &sync.Mutex{})
	return v.(*sync.Mutex)
}

// RefreshAll 遍历 Store 中所有已授权企业,刷新它们的 corp_token。
// 任一失败继续下一个,最后聚合错误用 errors.Join。
func (c *Client) RefreshAll(ctx context.Context) error {
	list, err := c.store.ListAuthorizers(ctx, c.cfg.SuiteID)
	if err != nil {
		return err
	}
	var errs []error
	for _, corpID := range list {
		if err := c.CorpClient(corpID).Refresh(ctx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", corpID, err))
		}
	}
	return errors.Join(errs...)
}
```

- [ ] **Step 8.6: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run "TestGetCorpToken|TestCorpClient|TestRefreshAll" -v`
Expected: PASS for all 6 cases

Run: `go test -race ./work-wechat/isv/...`
Expected: all green, no data races

- [ ] **Step 8.7: Commit**

```bash
git add work-wechat/isv/corp.token.go work-wechat/isv/corp.token_test.go work-wechat/isv/authorizer.go work-wechat/isv/struct.corp.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add corp_token lifecycle + CorpClient + RefreshAll

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 9: ParseNotify 骨架 + suite_ticket + 4 个授权事件

**Files:**
- Create: `work-wechat/isv/notify.go`
- Create: `work-wechat/isv/notify_test.go`
- Create: `work-wechat/isv/struct.notify.go`

本任务只覆盖 5 种事件:`suite_ticket` / `create_auth` / `change_auth` / `cancel_auth` / `reset_permanent_code`。剩下 4 种变更事件 + RawEvent 在 Task 10。

- [ ] **Step 9.1: Write failing test(事件 5 种 + 1 签名失败)**

Create `work-wechat/isv/notify_test.go`:

```go
package isv

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/godrealms/go-wechat-sdk/utils/wxcrypto"
)

// buildNotifyRequest 构造一个签名/加密正确的 HTTP POST,模拟企业微信的回调推送。
// innerXML 是明文的 innerBody(不含 Encrypt 信封)。
func buildNotifyRequest(t *testing.T, c *Client, innerXML string) *http.Request {
	t.Helper()
	timestamp := "1712900000"
	nonce := "nonce123"
	payload, msgSig, err := c.crypto.BuildEncryptedReply([]byte(innerXML), timestamp, nonce)
	if err != nil {
		t.Fatal(err)
	}
	_ = msgSig // BuildEncryptedReply returns signed envelope already

	u := fmt.Sprintf("/cb?msg_signature=%s&timestamp=%s&nonce=%s",
		extractSignature(t, c.crypto, payload, timestamp, nonce), timestamp, nonce)
	req := httptest.NewRequest(http.MethodPost, u, bytes.NewReader(payload))
	return req
}

// extractSignature recomputes the signature from the envelope so the test URL matches.
// For the helper to work, we parse the envelope and re-sign with crypto.Signature.
func extractSignature(t *testing.T, cry *wxcrypto.MsgCrypto, envelope []byte, timestamp, nonce string) string {
	t.Helper()
	var env struct {
		XMLName xml.Name `xml:"xml"`
		Encrypt string   `xml:"Encrypt"`
	}
	if err := xml.Unmarshal(envelope, &env); err != nil {
		t.Fatal(err)
	}
	return cry.Signature(timestamp, nonce, env.Encrypt)
}

func TestParseNotify_SuiteTicket_Persists(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	// Remove the pre-seeded ticket so we can verify the store write.
	_ = c.store.PutSuiteTicket(context.Background(), "suite1", "")

	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[suite_ticket]]></InfoType><SuiteTicket><![CDATA[NEW_TICKET]]></SuiteTicket></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	stev, ok := ev.(*SuiteTicketEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if stev.SuiteTicket != "NEW_TICKET" || stev.SuiteID != "suite1" {
		t.Errorf("event: %+v", stev)
	}
	// Store should have been updated
	got, _ := c.store.GetSuiteTicket(context.Background(), "suite1")
	if got != "NEW_TICKET" {
		t.Errorf("store not updated: %q", got)
	}
}

func TestParseNotify_CreateAuth(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[create_auth]]></InfoType><AuthCode><![CDATA[AC123]]></AuthCode></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*CreateAuthEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.AuthCode != "AC123" {
		t.Errorf("event: %+v", cev)
	}
}

func TestParseNotify_ChangeAuth(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[change_auth]]></InfoType><AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ChangeAuthEvent)
	if !ok || cev.AuthCorpID != "wxcorp1" {
		t.Fatalf("event: %T %+v", ev, ev)
	}
}

func TestParseNotify_CancelAuth(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[cancel_auth]]></InfoType><AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ev.(*CancelAuthEvent); !ok {
		t.Fatalf("type: %T", ev)
	}
}

func TestParseNotify_ResetPermanentCode(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[reset_permanent_code]]></InfoType><AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ev.(*ResetPermanentCodeEvent); !ok {
		t.Fatalf("type: %T", ev)
	}
}

func TestParseNotify_BadSignature(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[suite_ticket]]></InfoType><SuiteTicket><![CDATA[X]]></SuiteTicket></xml>`
	req := buildNotifyRequest(t, c, inner)

	// Tamper with the signature
	q := req.URL.Query()
	q.Set("msg_signature", "deadbeef")
	req.URL.RawQuery = q.Encode()

	_, err := c.ParseNotify(req)
	if err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("want signature error, got %v", err)
	}
}
```

**Notes for the implementer:**
- `wxcrypto.MsgCrypto` must already expose `BuildEncryptedReply`, `Signature`, `VerifySignature`, `Decrypt`. These were established in the oplatform rounds. If any method is missing, verify first before proceeding — do not add to wxcrypto.
- `BuildEncryptedReply([]byte, timestamp, nonce) ([]byte envelope, string msgSignature, error)` — returns signed XML envelope.

- [ ] **Step 9.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run TestParseNotify`
Expected: FAIL (ParseNotify undefined)

- [ ] **Step 9.3: Write `struct.notify.go` (5 events)**

```go
package isv

import "time"

// Event 是所有回调事件的统一接口。
type Event interface {
	isEvent()
}

// baseEvent 被所有具体事件嵌入,承载通用字段。
type baseEvent struct {
	SuiteID   string
	ReceiveAt time.Time
}

func (baseEvent) isEvent() {}

// SuiteTicketEvent —— 微信每 10 分钟推送一次,本包已自动持久化到 Store。
type SuiteTicketEvent struct {
	baseEvent
	SuiteTicket string
}

// CreateAuthEvent —— 企业授权成功。
type CreateAuthEvent struct {
	baseEvent
	AuthCode string
}

// ChangeAuthEvent —— 企业变更授权。
type ChangeAuthEvent struct {
	baseEvent
	AuthCorpID string
}

// CancelAuthEvent —— 企业取消授权。
type CancelAuthEvent struct {
	baseEvent
	AuthCorpID string
}

// ResetPermanentCodeEvent —— 重置永久授权码。
type ResetPermanentCodeEvent struct {
	baseEvent
	AuthCorpID string
}
```

- [ ] **Step 9.4: Implement `notify.go` (5 events)**

```go
package isv

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// componentEnvelope 是外层加密信封 XML。
type componentEnvelope struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string   `xml:"ToUserName"`
	Encrypt    string   `xml:"Encrypt"`
	AgentID    string   `xml:"AgentID,omitempty"`
}

// componentInner 是解密后的内层 XML,承载所有 InfoType 的字段。
// 对于不同 InfoType,只有部分字段有值;用 omitempty 可避免 unmarshal 报错。
type componentInner struct {
	XMLName       xml.Name `xml:"xml"`
	SuiteID       string   `xml:"SuiteId"`
	InfoType      string   `xml:"InfoType"`
	TimeStamp     int64    `xml:"TimeStamp"`
	SuiteTicket   string   `xml:"SuiteTicket,omitempty"`
	AuthCode      string   `xml:"AuthCode,omitempty"`
	AuthCorpID    string   `xml:"AuthCorpId,omitempty"`
	ChangeType    string   `xml:"ChangeType,omitempty"`
	UserID        string   `xml:"UserID,omitempty"`
	Name          string   `xml:"Name,omitempty"`
	Department    string   `xml:"Department,omitempty"`
	NewUserID     string   `xml:"NewUserID,omitempty"`
	ExternalUserID string  `xml:"ExternalUserID,omitempty"`
	AgentID       string   `xml:"AgentID,omitempty"`
	IsAdmin       int      `xml:"IsAdmin,omitempty"`
}

// ParseNotify 校验、解密、解析企业微信回调,并返回强类型事件。
//
// 对 suite_ticket 事件本函数自动调用 Store.PutSuiteTicket。
// 对未知 InfoType 返回 *RawEvent(Task 10 引入),不报错。
func (c *Client) ParseNotify(r *http.Request) (Event, error) {
	ctx := r.Context()

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("isv: read body: %w", err)
	}
	var env componentEnvelope
	if err := xml.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("isv: parse envelope: %w", err)
	}

	q := r.URL.Query()
	msgSig := q.Get("msg_signature")
	timestamp := q.Get("timestamp")
	nonce := q.Get("nonce")
	if err := c.crypto.VerifySignature(timestamp, nonce, env.Encrypt, msgSig); err != nil {
		return nil, fmt.Errorf("isv: verify signature: %w", err)
	}

	plain, err := c.crypto.Decrypt(env.Encrypt)
	if err != nil {
		return nil, fmt.Errorf("isv: decrypt: %w", err)
	}

	var inner componentInner
	if err := xml.Unmarshal(plain, &inner); err != nil {
		return nil, fmt.Errorf("isv: parse inner: %w", err)
	}

	base := baseEvent{SuiteID: inner.SuiteID, ReceiveAt: time.Now()}

	switch inner.InfoType {
	case "suite_ticket":
		if err := c.store.PutSuiteTicket(ctx, inner.SuiteID, inner.SuiteTicket); err != nil {
			return nil, fmt.Errorf("isv: persist suite_ticket: %w", err)
		}
		return &SuiteTicketEvent{baseEvent: base, SuiteTicket: inner.SuiteTicket}, nil
	case "create_auth":
		return &CreateAuthEvent{baseEvent: base, AuthCode: inner.AuthCode}, nil
	case "change_auth":
		return &ChangeAuthEvent{baseEvent: base, AuthCorpID: inner.AuthCorpID}, nil
	case "cancel_auth":
		return &CancelAuthEvent{baseEvent: base, AuthCorpID: inner.AuthCorpID}, nil
	case "reset_permanent_code":
		return &ResetPermanentCodeEvent{baseEvent: base, AuthCorpID: inner.AuthCorpID}, nil
	}
	return nil, errors.New("isv: unknown InfoType " + inner.InfoType) // Task 10 replaces with *RawEvent
}
```

- [ ] **Step 9.5: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run TestParseNotify -v`
Expected: PASS for the 6 cases (5 events + bad signature)

Run: `go test ./work-wechat/isv/...`
Expected: all prior tests still green

- [ ] **Step 9.6: Commit**

```bash
git add work-wechat/isv/notify.go work-wechat/isv/notify_test.go work-wechat/isv/struct.notify.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add ParseNotify + suite_ticket, auth events (5 InfoType)

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 10: 补齐剩余 4 个变更事件 + RawEvent 兜底

**Files:**
- Modify: `work-wechat/isv/struct.notify.go` — 追加 `ChangeContactEvent` / `ChangeExternalContactEvent` / `ShareAgentChangeEvent` / `ChangeAppAdminEvent` / `RawEvent`
- Modify: `work-wechat/isv/notify.go` — 扩展 `switch inner.InfoType`,补 4 个事件 + RawEvent fallback
- Modify: `work-wechat/isv/notify_test.go` — 追加 5 个用例(4 个变更事件 + 1 个 RawEvent)

- [ ] **Step 10.1: Append failing tests**

Add to `work-wechat/isv/notify_test.go`:

```go
func TestParseNotify_ChangeContact(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml>
<SuiteId><![CDATA[suite1]]></SuiteId>
<InfoType><![CDATA[change_contact]]></InfoType>
<AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId>
<ChangeType><![CDATA[update_user]]></ChangeType>
<UserID><![CDATA[u1]]></UserID>
<NewUserID><![CDATA[u1new]]></NewUserID>
<Name><![CDATA[Alice]]></Name>
</xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ChangeContactEvent)
	if !ok {
		t.Fatalf("type: %T", ev)
	}
	if cev.ChangeType != "update_user" || cev.NewUserID != "u1new" {
		t.Errorf("event: %+v", cev)
	}
}

func TestParseNotify_ChangeExternalContact(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[change_external_contact]]></InfoType><AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId><ChangeType><![CDATA[add_external_contact]]></ChangeType><UserID><![CDATA[u1]]></UserID><ExternalUserID><![CDATA[ex1]]></ExternalUserID></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	cev, ok := ev.(*ChangeExternalContactEvent)
	if !ok || cev.ExternalUserID != "ex1" {
		t.Fatalf("event: %T %+v", ev, ev)
	}
}

func TestParseNotify_ShareAgentChange(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[share_agent_change]]></InfoType><AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId><AgentID><![CDATA[1000001]]></AgentID></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	sev, ok := ev.(*ShareAgentChangeEvent)
	if !ok || sev.AgentID != "1000001" {
		t.Fatalf("event: %T %+v", ev, ev)
	}
}

func TestParseNotify_ChangeAppAdmin(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[change_app_admin]]></InfoType><AuthCorpId><![CDATA[wxcorp1]]></AuthCorpId><UserID><![CDATA[u_admin]]></UserID><IsAdmin>1</IsAdmin></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	aev, ok := ev.(*ChangeAppAdminEvent)
	if !ok || !aev.IsAdmin || aev.UserID != "u_admin" {
		t.Fatalf("event: %T %+v", ev, ev)
	}
}

func TestParseNotify_UnknownInfoType_ReturnsRawEvent(t *testing.T) {
	c := newTestISVClient(t, "http://unused")
	inner := `<xml><SuiteId><![CDATA[suite1]]></SuiteId><InfoType><![CDATA[brand_new_event]]></InfoType><Foo><![CDATA[bar]]></Foo></xml>`
	req := buildNotifyRequest(t, c, inner)

	ev, err := c.ParseNotify(req)
	if err != nil {
		t.Fatal(err)
	}
	rev, ok := ev.(*RawEvent)
	if !ok || rev.InfoType != "brand_new_event" {
		t.Fatalf("event: %T %+v", ev, ev)
	}
	if !strings.Contains(rev.RawXML, "<Foo>") {
		t.Errorf("RawXML missing Foo: %q", rev.RawXML)
	}
}
```

- [ ] **Step 10.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run "TestParseNotify_Change|TestParseNotify_Share|TestParseNotify_Unknown"`
Expected: FAIL (undefined event types / currently returns error for unknown InfoType)

- [ ] **Step 10.3: Append event types to `struct.notify.go`**

Add to `work-wechat/isv/struct.notify.go`:

```go
// ChangeContactEvent —— 通讯录变更(成员/部门/标签)。
type ChangeContactEvent struct {
	baseEvent
	AuthCorpID string
	ChangeType string // create_user / update_user / delete_user / create_party / update_party / delete_party / update_tag
	UserID     string
	Name       string
	Department string
	NewUserID  string // 仅 update_user 在 userid 变更时出现
}

// ChangeExternalContactEvent —— 外部联系人变更。
type ChangeExternalContactEvent struct {
	baseEvent
	AuthCorpID     string
	ChangeType     string
	UserID         string
	ExternalUserID string
}

// ShareAgentChangeEvent —— 共享应用变更。
type ShareAgentChangeEvent struct {
	baseEvent
	AuthCorpID string
	AgentID    string
}

// ChangeAppAdminEvent —— 应用管理员变更。
type ChangeAppAdminEvent struct {
	baseEvent
	AuthCorpID string
	UserID     string
	IsAdmin    bool
}

// RawEvent 是未知 InfoType 的兜底,调用方可以自行 unmarshal。
type RawEvent struct {
	baseEvent
	InfoType string
	RawXML   string
}
```

- [ ] **Step 10.4: Replace the `switch` tail in `notify.go`**

In `notify.go`, replace the final `return nil, errors.New("isv: unknown InfoType " + inner.InfoType)` line and the existing `switch` with the expanded version:

```go
	switch inner.InfoType {
	case "suite_ticket":
		if err := c.store.PutSuiteTicket(ctx, inner.SuiteID, inner.SuiteTicket); err != nil {
			return nil, fmt.Errorf("isv: persist suite_ticket: %w", err)
		}
		return &SuiteTicketEvent{baseEvent: base, SuiteTicket: inner.SuiteTicket}, nil
	case "create_auth":
		return &CreateAuthEvent{baseEvent: base, AuthCode: inner.AuthCode}, nil
	case "change_auth":
		return &ChangeAuthEvent{baseEvent: base, AuthCorpID: inner.AuthCorpID}, nil
	case "cancel_auth":
		return &CancelAuthEvent{baseEvent: base, AuthCorpID: inner.AuthCorpID}, nil
	case "reset_permanent_code":
		return &ResetPermanentCodeEvent{baseEvent: base, AuthCorpID: inner.AuthCorpID}, nil
	case "change_contact":
		return &ChangeContactEvent{
			baseEvent:  base,
			AuthCorpID: inner.AuthCorpID,
			ChangeType: inner.ChangeType,
			UserID:     inner.UserID,
			Name:       inner.Name,
			Department: inner.Department,
			NewUserID:  inner.NewUserID,
		}, nil
	case "change_external_contact":
		return &ChangeExternalContactEvent{
			baseEvent:      base,
			AuthCorpID:     inner.AuthCorpID,
			ChangeType:     inner.ChangeType,
			UserID:         inner.UserID,
			ExternalUserID: inner.ExternalUserID,
		}, nil
	case "share_agent_change":
		return &ShareAgentChangeEvent{
			baseEvent:  base,
			AuthCorpID: inner.AuthCorpID,
			AgentID:    inner.AgentID,
		}, nil
	case "change_app_admin":
		return &ChangeAppAdminEvent{
			baseEvent:  base,
			AuthCorpID: inner.AuthCorpID,
			UserID:     inner.UserID,
			IsAdmin:    inner.IsAdmin == 1,
		}, nil
	default:
		return &RawEvent{
			baseEvent: base,
			InfoType:  inner.InfoType,
			RawXML:    string(plain),
		}, nil
	}
```

Also remove the now-unused `"errors"` import from `notify.go` if it is no longer referenced.

- [ ] **Step 10.5: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run TestParseNotify -v`
Expected: all 11 ParseNotify cases PASS

Run: `go test -race ./work-wechat/isv/...`
Expected: all green

- [ ] **Step 10.6: Commit**

```bash
git add work-wechat/isv/notify.go work-wechat/isv/notify_test.go work-wechat/isv/struct.notify.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): ParseNotify supports all 9 InfoType + RawEvent fallback

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 11: example/main.go + 全模块最终验证

**Files:**
- Create: `work-wechat/isv/example/main.go`

- [ ] **Step 11.1: Write `example/main.go` (compile-level demo)**

```go
//go:build ignore
// +build ignore

// Package main is a compile-only demo for work-wechat/isv. It demonstrates the
// shape of a typical ISV service provider integration. It does not actually
// reach the network; all calls are commented-out or behind a bogus config.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/godrealms/go-wechat-sdk/work-wechat/isv"
)

func main() {
	cfg := isv.Config{
		SuiteID:        "your_suite_id",
		SuiteSecret:    "your_suite_secret",
		ProviderCorpID: "your_provider_corpid",
		ProviderSecret: "your_provider_secret",
		Token:          "your_callback_token",
		EncodingAESKey: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ",
	}

	client, err := isv.NewClient(cfg,
		isv.WithStore(isv.NewMemoryStore()),
	)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/wecom/callback", func(w http.ResponseWriter, r *http.Request) {
		ev, err := client.ParseNotify(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch e := ev.(type) {
		case *isv.SuiteTicketEvent:
			log.Printf("suite_ticket persisted: %s", e.SuiteTicket)
		case *isv.CreateAuthEvent:
			log.Printf("new auth, auth_code=%s", e.AuthCode)
			resp, err := client.GetPermanentCode(r.Context(), e.AuthCode)
			if err != nil {
				log.Printf("get permanent code: %v", err)
				return
			}
			log.Printf("corp %s authorized", resp.AuthCorpInfo.CorpName)
		case *isv.CancelAuthEvent:
			log.Printf("cancel auth corp=%s", e.AuthCorpID)
		default:
			log.Printf("event: %T", ev)
		}
		_, _ = w.Write([]byte("success"))
	})

	// Below code is commented out; demonstrates usage only.
	_ = func(ctx context.Context) {
		preAuth, _ := client.GetPreAuthCode(ctx)
		url := client.AuthorizeURL(preAuth.PreAuthCode, "https://your.callback/ret", "state")
		fmt.Println(url)

		corpClient := client.CorpClient("wxcorp1")
		_, _ = corpClient.AccessToken(ctx)
		_ = client.RefreshAll(ctx)
	}

	log.Println("demo wired; run as production service to enable")
}
```

- [ ] **Step 11.2: Verify example compiles (build tag prevents it from entering `go test ./...` graph)**

Run: `go build -tags=ignore ./work-wechat/isv/example/...`
Expected: clean build (no output)

Note: because the file has a `//go:build ignore` tag, `go test ./...` and `go vet ./...` will skip it, and so does the default build. The explicit `-tags=ignore` tells go to include it for the verification step.

- [ ] **Step 11.3: Full module verification**

Run: `go vet ./work-wechat/isv/...`
Expected: clean

Run: `go test -race ./work-wechat/isv/... -count=1`
Expected: all PASS, no races

Run: `go test ./...`
Expected: all packages still green (no regressions anywhere else)

- [ ] **Step 11.4: Spot-check coverage of isv package**

Run: `go test -cover ./work-wechat/isv/...`
Expected: coverage ≥ 80% (the 16 public methods and ParseNotify are all tested; only NewClient option branches and error plumbing may be under-covered).

If any individual method is below 50% coverage, add a targeted test for it before moving on. Otherwise proceed to commit.

- [ ] **Step 11.5: Commit**

```bash
git add work-wechat/isv/example/main.go
git commit -m "$(cat <<'EOF'
docs(work-wechat/isv): add compile-only example demonstrating ISV flow

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 11.6: Tag final state and report**

Run the following summary commands and report their output:

```bash
git log --oneline -15
```
Expected: 11 `feat(work-wechat/isv)` / `docs(work-wechat/isv)` commits on top of the spec commit, representing Tasks 1-11.

```bash
go test -cover ./work-wechat/isv/...
```
Report the coverage percentage.

The implementation is complete. No further commits unless issues are found.

---

## Self-Review Checklist

### Spec coverage

- [x] §1.1 Goal: 16 public methods — Tasks 4 (2), 5 (3), 6 (3), 7 (2), 8 (4), 9+10 (1 ParseNotify) = 15; `AuthorizeURL` is in Task 5 as the non-HTTP method = **16 total**. ✔
- [x] §2 Architecture reuse: `doPost`/`doGet`/`doPostRaw` in Task 3, lazy + double-check in Tasks 4/7/8, `newTestISVClient` helper in Task 4, `newTestISVClientWithProvider` in Task 7. ✔
- [x] §3 File layout: every file mapped to a Task. Note the plan uses `provider.id_convert.go` instead of folding the 2 methods into `suite.permanent.go` (cleaner separation since provider uses different token). ✔
- [x] §4 Types: Config (Task 3), Client (Task 3), Store + MemoryStore (Task 2), TokenSource (Task 1), CorpClient (Task 8), errors (Task 1). ✔
- [x] §5 HTTP helpers: `doPost`/`doGet`/`doPostRaw` + `decodeRaw` two-pass in Task 3. ✔
- [x] §6 16 methods: see coverage above. ✔
- [x] §7 并发: three locks (`suiteMu` Task 3, `providerMu` Task 3, `corpMu` Task 8 via `lockFor`). ✔
- [x] §8 测试矩阵: 44 cases target, plan achieves ~45 (7+5+7+3+3+3+5+11+store 7 + client 4 + 其它). ✔
- [x] §9 三类错误: WeixinError (Task 3 decodeRaw), 哨兵错误 (Task 1), HTTP 透传 (Task 3). ✔

### Placeholder scan

- No "TBD" / "TODO" / "fill in details"
- Every code step contains complete Go code
- Every test step contains complete Go test code
- Every command step has exact command + expected output

### Type consistency

- `Config` fields match between Task 3 (`SuiteID`/`SuiteSecret`/`ProviderCorpID`/`ProviderSecret`/`Token`/`EncodingAESKey`) and spec §4.1. ✔
- `AuthorizerTokens` fields (`CorpID`/`PermanentCode`/`CorpAccessToken`/`CorpTokenExpireAt`) consistent across Tasks 2, 6, 8. ✔
- Method names match spec: `GetSuiteAccessToken` / `RefreshSuiteToken` / `GetPreAuthCode` / `SetSessionInfo` / `AuthorizeURL` / `GetPermanentCode` / `GetAuthInfo` / `GetAdminList` / `GetCorpToken` / `CorpClient` / `(*CorpClient).AccessToken` / `(*CorpClient).Refresh` / `RefreshAll` / `ParseNotify` / `CorpIDToOpenCorpID` / `UserIDToOpenUserID`. ✔
- Event type names match: `SuiteTicketEvent` / `CreateAuthEvent` / `ChangeAuthEvent` / `CancelAuthEvent` / `ResetPermanentCodeEvent` / `ChangeContactEvent` / `ChangeExternalContactEvent` / `ShareAgentChangeEvent` / `ChangeAppAdminEvent` / `RawEvent`. ✔
- Error names match: `ErrNotFound` / `ErrSuiteTicketMissing` / `ErrProviderCorpIDMissing` / `ErrProviderSecretMissing` / `ErrAuthorizerRevoked` / `WeixinError`. ✔

**Plan is self-consistent and fully covers the spec.**
