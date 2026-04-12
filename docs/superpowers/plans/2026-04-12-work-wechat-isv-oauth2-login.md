# work-wechat ISV Sub-Project 7 — OAuth2 Web Login Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Append 3 public methods to the existing `work-wechat/isv` package covering the WeChat Work third-party web OAuth2 trio: `OAuth2URL` (local URL builder), `GetUserInfo3rd`, and `GetUserDetail3rd`.

**Architecture:** Pure additive extension of sub-projects 1 and 2. `OAuth2URL` is a pure-local function that returns `https://open.weixin.qq.com/connect/oauth2/authorize?...#wechat_redirect`. The two remote methods use `provider_access_token` — `GetUserInfo3rd` is a GET (new private helper `providerDoGet`), `GetUserDetail3rd` is a POST (existing `providerDoPost`). Functional Option pattern (`OAuth2Option`) mirrors `WithStore`.

**Tech Stack:** Go 1.23+, `net/url`, `strconv`, `net/http/httptest`, existing `work-wechat/isv` package. Spec: `docs/superpowers/specs/2026-04-12-work-wechat-isv-oauth2-login-design.md`.

---

## File Layout

```
work-wechat/isv/
├── struct.oauth2.go              # NEW — DTO: UserInfo3rdResp, UserDetail3rdResp
├── oauth2.go                     # NEW — OAuth2URL + OAuth2Option + GetUserInfo3rd + GetUserDetail3rd
├── oauth2_test.go                # NEW — 5 tests
└── provider.id_convert.go        # MODIFY — append private providerDoGet helper
```

**Assumptions about existing helpers (verified before writing this plan):**
- `newTestISVClientWithProvider(t, baseURL)` exists in `provider.id_convert_test.go` — seeds suite_ticket, sets ProviderCorpID=`wxprov`, ProviderSecret=`PSEC`, SuiteID=`suite1`.
- `providerDoPost(ctx, path, body, out)` exists in `provider.id_convert.go` — injects `provider_access_token` query param.
- `getProviderAccessToken(ctx)` is the private lazy-fetch helper used by both `providerDoPost` (and will be used by `providerDoGet`).
- `doRequestRaw(ctx, method, path, query, body, out)` is the low-level HTTP helper (we use it directly — `doGet` auto-injects **suite** token, which is wrong for provider-scoped GETs).
- `c.cfg.SuiteID` is accessible for URL construction.

**Spec deviation:** The spec (§2.5) suggests `providerDoGet` wraps `c.doGet`. That is incorrect — `c.doGet` auto-injects `suite_access_token`, not `provider_access_token`. This plan wraps `c.doRequestRaw` directly instead. Same ~10-line footprint, correct semantics.

---

## Task 1: DTO scaffolding — `struct.oauth2.go`

**Files:**
- Create: `work-wechat/isv/struct.oauth2.go`

- [ ] **Step 1.1: Write the new DTO file**

Create `work-wechat/isv/struct.oauth2.go`:

```go
package isv

// UserInfo3rdResp 是 service/auth/getuserinfo3rd 的响应。
// 企业成员 / 非企业成员 / 外部联系人返回的字段不同,全部用 omitempty 填进同一个结构体。
// 注意:官方字段名大小写混用,必须原样映射。
type UserInfo3rdResp struct {
	CorpID         string `json:"CorpId"`
	UserID         string `json:"UserId"`          // 企业成员
	DeviceID       string `json:"DeviceId"`
	UserTicket     string `json:"user_ticket"`     // 企业成员才返回,用于后续换详情
	ExpiresIn      int    `json:"expires_in"`      // user_ticket 有效期(秒)
	OpenUserID     string `json:"open_userid"`     // 跨服务商匿名 id
	OpenID         string `json:"OpenId"`          // 非企业成员时的微信 openid
	ExternalUserID string `json:"external_userid"` // 外部联系人
}

// UserDetail3rdResp 是 service/auth/getuserdetail3rd 的响应。
// 注意:此接口对敏感字段有调用者备案要求,调用前请确认合规。
type UserDetail3rdResp struct {
	CorpID  string `json:"corpid"`
	UserID  string `json:"userid"`
	Gender  string `json:"gender"` // 1 男 / 2 女
	Avatar  string `json:"avatar"`
	QRCode  string `json:"qr_code"`
	Mobile  string `json:"mobile"`
	Email   string `json:"email"`
	BizMail string `json:"biz_mail"`
	Address string `json:"address"`
}
```

- [ ] **Step 1.2: Verify compilation**

Run: `go build ./work-wechat/isv/...`
Expected: clean build (no output).

- [ ] **Step 1.3: Commit**

```bash
git add work-wechat/isv/struct.oauth2.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add OAuth2 userinfo3rd/userdetail3rd DTOs

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: `providerDoGet` private helper

**Files:**
- Modify: `work-wechat/isv/provider.id_convert.go` — append `providerDoGet`

- [ ] **Step 2.1: Append the helper**

Append to `work-wechat/isv/provider.id_convert.go` (after the existing `providerDoPost`, before `CorpIDToOpenCorpID`):

```go
// providerDoGet 和 doGet 类似,只是注入的 token 是 provider_access_token。
// 不能复用 c.doGet —— 后者会自动注入 suite_access_token。
func (c *Client) providerDoGet(ctx context.Context, path string, extra url.Values, out interface{}) error {
	tok, err := c.getProviderAccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"provider_access_token": {tok}}
	for k, vs := range extra {
		q[k] = vs
	}
	return c.doRequestRaw(ctx, http.MethodGet, path, q, nil, out)
}
```

The imports `context`, `net/url`, and `net/http` are already in the file (verify before saving — `http` may need to be added). After editing, the import block at the top should include:

```go
import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)
```

- [ ] **Step 2.2: Verify compilation**

Run: `go build ./work-wechat/isv/...`
Expected: clean build. This proves the helper compiles; no test yet (Task 3 will exercise it end-to-end).

- [ ] **Step 2.3: Commit**

```bash
git add work-wechat/isv/provider.id_convert.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add providerDoGet helper for GET with provider token

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: `OAuth2URL` + `OAuth2Option`

**Files:**
- Create: `work-wechat/isv/oauth2.go`
- Create: `work-wechat/isv/oauth2_test.go`

- [ ] **Step 3.1: Write the failing tests**

Create `work-wechat/isv/oauth2_test.go`:

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
)

func TestOAuth2URL_Default(t *testing.T) {
	cfg := testConfig()
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	got := c.OAuth2URL("https://app.example.com/cb?x=1&y=2", "STATE1")

	// Split off the fragment before parsing.
	if !strings.HasSuffix(got, "#wechat_redirect") {
		t.Errorf("missing fragment: %q", got)
	}
	base := strings.TrimSuffix(got, "#wechat_redirect")

	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "open.weixin.qq.com" {
		t.Errorf("host: %q", u.Host)
	}
	if u.Path != "/connect/oauth2/authorize" {
		t.Errorf("path: %q", u.Path)
	}
	q := u.Query()
	if q.Get("appid") != cfg.SuiteID {
		t.Errorf("appid: %q", q.Get("appid"))
	}
	if q.Get("redirect_uri") != "https://app.example.com/cb?x=1&y=2" {
		t.Errorf("redirect_uri: %q", q.Get("redirect_uri"))
	}
	if q.Get("response_type") != "code" {
		t.Errorf("response_type: %q", q.Get("response_type"))
	}
	if q.Get("scope") != "snsapi_privateinfo" {
		t.Errorf("scope: %q", q.Get("scope"))
	}
	if q.Get("state") != "STATE1" {
		t.Errorf("state: %q", q.Get("state"))
	}
	if q.Get("agentid") != "" {
		t.Errorf("agentid should be absent by default, got %q", q.Get("agentid"))
	}
}

func TestOAuth2URL_WithOptions(t *testing.T) {
	cfg := testConfig()
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	got := c.OAuth2URL(
		"https://app.example.com/cb",
		"STATE2",
		WithOAuth2Scope("snsapi_base"),
		WithOAuth2AgentID(1000001),
	)
	base := strings.TrimSuffix(got, "#wechat_redirect")
	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if q.Get("scope") != "snsapi_base" {
		t.Errorf("scope: %q", q.Get("scope"))
	}
	if q.Get("agentid") != "1000001" {
		t.Errorf("agentid: %q", q.Get("agentid"))
	}
}
```

- [ ] **Step 3.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run "TestOAuth2URL" -v`
Expected: FAIL — `c.OAuth2URL undefined` / `WithOAuth2Scope undefined` / `WithOAuth2AgentID undefined`.

- [ ] **Step 3.3: Implement OAuth2URL + options**

Create `work-wechat/isv/oauth2.go`:

```go
package isv

import (
	"context"
	"net/url"
	"strconv"
)

// OAuth2Option 配置 OAuth2URL 的可选参数。
type OAuth2Option func(*oauth2Params)

type oauth2Params struct {
	scope    string
	agentID  int
	hasAgent bool
}

// WithOAuth2Scope 覆盖默认的 scope。默认 "snsapi_privateinfo"。
// 可选值:snsapi_base / snsapi_privateinfo。
func WithOAuth2Scope(scope string) OAuth2Option {
	return func(p *oauth2Params) { p.scope = scope }
}

// WithOAuth2AgentID 设置 agentid query 参数。
// 仅当 scope=snsapi_privateinfo 时必填,调用方负责正确性。
func WithOAuth2AgentID(agentID int) OAuth2Option {
	return func(p *oauth2Params) {
		p.agentID = agentID
		p.hasAgent = true
	}
}

// OAuth2URL 构造企业微信第三方网页授权 URL。
// 调用方把返回值塞到 302 Location header 即可。
// redirectURI:用户同意授权后企业微信回跳的 URL,必须在服务商后台白名单内。
// state:调用方自定义的防 CSRF 值,回跳时原样带回。
func (c *Client) OAuth2URL(redirectURI, state string, opts ...OAuth2Option) string {
	p := &oauth2Params{scope: "snsapi_privateinfo"}
	for _, opt := range opts {
		opt(p)
	}
	q := url.Values{}
	q.Set("appid", c.cfg.SuiteID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", p.scope)
	q.Set("state", state)
	if p.hasAgent {
		q.Set("agentid", strconv.Itoa(p.agentID))
	}
	return "https://open.weixin.qq.com/connect/oauth2/authorize?" + q.Encode() + "#wechat_redirect"
}

// Unused import guard — removed in Task 4 when GetUserInfo3rd lands.
var _ context.Context
```

Wait — the `context` import is unused in Task 3. Remove it until Task 4 adds the HTTP methods. Revised file **without** the guard:

```go
package isv

import (
	"net/url"
	"strconv"
)

// OAuth2Option 配置 OAuth2URL 的可选参数。
type OAuth2Option func(*oauth2Params)

type oauth2Params struct {
	scope    string
	agentID  int
	hasAgent bool
}

// WithOAuth2Scope 覆盖默认的 scope。默认 "snsapi_privateinfo"。
// 可选值:snsapi_base / snsapi_privateinfo。
func WithOAuth2Scope(scope string) OAuth2Option {
	return func(p *oauth2Params) { p.scope = scope }
}

// WithOAuth2AgentID 设置 agentid query 参数。
// 仅当 scope=snsapi_privateinfo 时必填,调用方负责正确性。
func WithOAuth2AgentID(agentID int) OAuth2Option {
	return func(p *oauth2Params) {
		p.agentID = agentID
		p.hasAgent = true
	}
}

// OAuth2URL 构造企业微信第三方网页授权 URL。
// 调用方把返回值塞到 302 Location header 即可。
// redirectURI:用户同意授权后企业微信回跳的 URL,必须在服务商后台白名单内。
// state:调用方自定义的防 CSRF 值,回跳时原样带回。
func (c *Client) OAuth2URL(redirectURI, state string, opts ...OAuth2Option) string {
	p := &oauth2Params{scope: "snsapi_privateinfo"}
	for _, opt := range opts {
		opt(p)
	}
	q := url.Values{}
	q.Set("appid", c.cfg.SuiteID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", p.scope)
	q.Set("state", state)
	if p.hasAgent {
		q.Set("agentid", strconv.Itoa(p.agentID))
	}
	return "https://open.weixin.qq.com/connect/oauth2/authorize?" + q.Encode() + "#wechat_redirect"
}
```

The test file's `context`, `net/http`, `net/http/httptest`, `encoding/json` imports are needed for Task 4's tests, but Go will complain about unused imports in Task 3. **Solution:** write the test file in Task 3 **without** those imports; Task 4 will rewrite the imports when it adds its tests. Revised Step 3.1 imports:

```go
import (
	"net/url"
	"strings"
	"testing"
)
```

(Remove `context`, `encoding/json`, `net/http`, `net/http/httptest` from Step 3.1.)

- [ ] **Step 3.4: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run "TestOAuth2URL" -v`
Expected: both PASS.

- [ ] **Step 3.5: Commit**

```bash
git add work-wechat/isv/oauth2.go work-wechat/isv/oauth2_test.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add OAuth2URL + OAuth2Option for web login

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: `GetUserInfo3rd` + `GetUserDetail3rd`

**Files:**
- Modify: `work-wechat/isv/oauth2.go` — append 2 methods, add `context` import
- Modify: `work-wechat/isv/oauth2_test.go` — append 3 tests, expand imports

- [ ] **Step 4.1: Write the failing tests**

Replace the import block at the top of `work-wechat/isv/oauth2_test.go` with:

```go
import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)
```

Append to the bottom of `work-wechat/isv/oauth2_test.go`:

```go
func TestGetUserInfo3rd_Member(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/auth/getuserinfo3rd":
			if r.Method != http.MethodGet {
				t.Errorf("method: %s", r.Method)
			}
			if got := r.URL.Query().Get("provider_access_token"); got != "PTOK" {
				t.Errorf("token query: %q", got)
			}
			if got := r.URL.Query().Get("code"); got != "AUTH1" {
				t.Errorf("code query: %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"CorpId":      "wxcorp1",
				"UserId":      "u1",
				"DeviceId":    "dev1",
				"user_ticket": "TICKET1",
				"expires_in":  1800,
				"open_userid": "ou1",
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetUserInfo3rd(context.Background(), "AUTH1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.CorpID != "wxcorp1" || resp.UserID != "u1" || resp.DeviceID != "dev1" {
		t.Errorf("member fields: %+v", resp)
	}
	if resp.UserTicket != "TICKET1" || resp.ExpiresIn != 1800 {
		t.Errorf("ticket: %+v", resp)
	}
	if resp.OpenUserID != "ou1" {
		t.Errorf("open_userid: %q", resp.OpenUserID)
	}
	if resp.OpenID != "" {
		t.Errorf("OpenID should be empty for member, got %q", resp.OpenID)
	}
}

func TestGetUserInfo3rd_NonMember(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/auth/getuserinfo3rd":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"CorpId": "wxcorp1",
				"OpenId": "oAbCdEf",
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetUserInfo3rd(context.Background(), "AUTH2")
	if err != nil {
		t.Fatal(err)
	}
	if resp.UserID != "" {
		t.Errorf("UserID should be empty for non-member, got %q", resp.UserID)
	}
	if resp.OpenID != "oAbCdEf" {
		t.Errorf("OpenID: %q", resp.OpenID)
	}
}

func TestGetUserDetail3rd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/auth/getuserdetail3rd":
			if r.Method != http.MethodPost {
				t.Errorf("method: %s", r.Method)
			}
			if got := r.URL.Query().Get("provider_access_token"); got != "PTOK" {
				t.Errorf("token query: %q", got)
			}
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["user_ticket"] != "TICKET1" {
				t.Errorf("body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"corpid":   "wxcorp1",
				"userid":   "u1",
				"gender":   "1",
				"avatar":   "http://img/a.png",
				"mobile":   "13800000000",
				"email":    "u1@example.com",
				"biz_mail": "u1@biz.example.com",
				"address":  "Beijing",
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetUserDetail3rd(context.Background(), "TICKET1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.CorpID != "wxcorp1" || resp.UserID != "u1" {
		t.Errorf("ids: %+v", resp)
	}
	if resp.Mobile != "13800000000" || resp.Email != "u1@example.com" || resp.BizMail != "u1@biz.example.com" {
		t.Errorf("contact: %+v", resp)
	}
	if resp.Gender != "1" || resp.Avatar != "http://img/a.png" || resp.Address != "Beijing" {
		t.Errorf("profile: %+v", resp)
	}
}
```

- [ ] **Step 4.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run "TestGetUserInfo3rd|TestGetUserDetail3rd" -v`
Expected: FAIL — `c.GetUserInfo3rd undefined` / `c.GetUserDetail3rd undefined`.

- [ ] **Step 4.3: Implement both methods**

Update `work-wechat/isv/oauth2.go` import block to:

```go
import (
	"context"
	"net/url"
	"strconv"
)
```

Append to the bottom of `work-wechat/isv/oauth2.go`:

```go
// GetUserInfo3rd 用回调返回的 auth_code 换取成员身份(UserId / user_ticket / open_userid)。
// 接口:GET /cgi-bin/service/auth/getuserinfo3rd?code=<authCode>
// 使用 provider_access_token。
// 返回的 UserTicket 可继续调用 GetUserDetail3rd 换取敏感详情。
func (c *Client) GetUserInfo3rd(ctx context.Context, authCode string) (*UserInfo3rdResp, error) {
	extra := url.Values{"code": {authCode}}
	var resp UserInfo3rdResp
	if err := c.providerDoGet(ctx, "/cgi-bin/service/auth/getuserinfo3rd", extra, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUserDetail3rd 用 user_ticket 换取成员的敏感详情(姓名/邮箱/头像/手机号)。
// 接口:POST /cgi-bin/service/auth/getuserdetail3rd,body 为 {"user_ticket": "..."}。
// 注意:此接口对敏感字段有调用者备案要求,调用前请确认合规。
func (c *Client) GetUserDetail3rd(ctx context.Context, userTicket string) (*UserDetail3rdResp, error) {
	body := map[string]string{"user_ticket": userTicket}
	var resp UserDetail3rdResp
	if err := c.providerDoPost(ctx, "/cgi-bin/service/auth/getuserdetail3rd", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 4.4: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run "TestGetUserInfo3rd|TestGetUserDetail3rd" -v`
Expected: all 3 PASS.

- [ ] **Step 4.5: Full isv suite + race + coverage + vet**

Run (serially):

```bash
go vet ./work-wechat/isv/...
go test -race ./work-wechat/isv/... -count=1
go test -cover ./work-wechat/isv/...
```

Expected:
- vet: clean
- race: all green
- coverage: ≥85% (sub-project 2 was 86.2%; this task adds ~75 prod LOC and ~220 test LOC, so coverage should stay ≥85%)

- [ ] **Step 4.6: Full repo regression**

Run: `go test ./... -count=1`
Expected: every package green (sub-projects 1 and 2 tests, oplatform, merchant, etc.).

- [ ] **Step 4.7: Commit**

```bash
git add work-wechat/isv/oauth2.go work-wechat/isv/oauth2_test.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add GetUserInfo3rd + GetUserDetail3rd

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 4.8: Tag final state and report**

Run the following and report the output:

```bash
git log --oneline -6
go test -cover ./work-wechat/isv/...
```

Expected: 4 new `feat(work-wechat/isv)` commits on top of the sub-project 7 spec commit (`6aaf774`), plus the coverage line.

---

## Self-Review

### Spec coverage

| Spec requirement | Task |
|---|---|
| §1.1 #1 `OAuth2URL` | Task 3 |
| §1.1 #2 `GetUserInfo3rd` | Task 4 |
| §1.1 #3 `GetUserDetail3rd` | Task 4 |
| §2.3 `OAuth2Option` / `WithOAuth2Scope` / `WithOAuth2AgentID` | Task 3 |
| §2.5 `providerDoGet` helper | Task 2 (with deviation: wraps `doRequestRaw` not `doGet`) |
| §3.1 `UserInfo3rdResp` + mixed-case JSON tags | Task 1 |
| §3.2 `UserDetail3rdResp` | Task 1 |
| §4.1 — §4.3 method bodies | Tasks 3, 4 |
| §5 5 test cases | Task 3 (2) + Task 4 (3) = **5** ✔ |
| §6 no new sentinel errors | ✔ — no task introduces new sentinels |
| §7 scale estimate ~117 prod / ~200 test | Task 1 ~30 + Task 2 ~12 + Task 3 ~50 + Task 4 ~25 = ~117 prod ✔ |
| §8 ~4 commits | 4 task commits ✔ |

### Placeholder scan

- No TBD / TODO / "implement later".
- Every code step contains complete Go code.
- Every test step contains complete Go test code.
- Every command step has the exact command + expected output.
- Step 3.3 contains a revision (the initial draft with an unused `context` import is explicitly replaced by the import-free version). This is a concrete rewrite, not a placeholder.

### Type / name consistency

- `UserInfo3rdResp` fields (`CorpID`, `UserID`, `DeviceID`, `UserTicket`, `ExpiresIn`, `OpenUserID`, `OpenID`, `ExternalUserID`): defined in Task 1, accessed in Task 4 tests. ✔
- `UserDetail3rdResp` fields (`CorpID`, `UserID`, `Gender`, `Avatar`, `Mobile`, `Email`, `BizMail`, `Address`): defined in Task 1, accessed in Task 4 tests. ✔
- `OAuth2Option` / `oauth2Params` / `WithOAuth2Scope` / `WithOAuth2AgentID`: defined in Task 3, no later task references them. ✔
- `providerDoGet`: defined in Task 2, used in Task 4 `GetUserInfo3rd`. ✔
- `providerDoPost`: pre-existing (sub-project 1), used in Task 4 `GetUserDetail3rd`. ✔
- `newTestISVClientWithProvider`: pre-existing (sub-project 1), used in Tasks 3 (via `testConfig`) and 4. ✔
- `testConfig`: pre-existing (sub-project 1 `client_test.go`), used in Task 3. ✔
- JSON tag casing (`CorpId` / `UserId` / `DeviceId` / `OpenId`): test payloads in Task 4 use the exact mixed-case keys the DTO expects. ✔

**Plan is self-consistent and fully covers the spec.**
