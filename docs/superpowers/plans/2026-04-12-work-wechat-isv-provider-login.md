# work-wechat ISV Sub-Project 2 — Provider Login Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Append 4 public methods to the existing `work-wechat/isv` package covering WeChat Work provider-self APIs: `GetProviderAccessToken`, `GetLoginInfo`, `GetRegisterCode`, `GetRegistrationInfo`.

**Architecture:** Pure additive extension of sub-project 1. All 4 methods hang off the existing `*Client`, all 3 remote methods go through the existing `providerDoPost` helper (provider_access_token injection). No Config/Store/Client struct changes; no new HTTP helpers; no new sentinel errors. File layout mirrors sub-project 1: one family-scoped `.go` + test + dedicated `struct.provider.go` DTO file.

**Tech Stack:** Go 1.23+, `encoding/json`, `net/http/httptest`, existing `work-wechat/isv` package. Spec: `docs/superpowers/specs/2026-04-12-work-wechat-isv-provider-login-design.md`.

---

## File Layout

```
work-wechat/isv/
├── struct.provider.go           # NEW — DTO: LoginInfoResp, GetRegisterCodeReq, RegisterCodeResp, RegistrationInfoResp
├── provider.login.go            # NEW — GetLoginInfo + GetRegisterCode + GetRegistrationInfo
├── provider.login_test.go       # NEW — ~7 tests
└── provider.id_convert.go       # MODIFY — append GetProviderAccessToken public wrapper + 1 test file tweak
```

**Assumptions about existing helpers (verified before writing this plan):**
- `newTestISVClientWithProvider(t, baseURL)` exists in `provider.id_convert_test.go` — seeds suite_ticket, sets ProviderCorpID=`wxprov`, ProviderSecret=`PSEC`.
- `testConfig()` exists in `client_test.go` — returns a base valid Config with SuiteID=`suite1`.
- `providerDoPost(ctx, path, body, out)` exists in `provider.id_convert.go` — injects `provider_access_token` query param.
- `store.PutProviderToken(ctx, suiteID, token, expireAt)` is on the `Store` interface.
- `AuthInfoAgent` (from `struct.permanent.go`) wraps `Agent []AuthAgent` under JSON key `agent`. `RegistrationInfoResp` reuses it.

---

## Task 1: DTO scaffolding — `struct.provider.go`

**Files:**
- Create: `work-wechat/isv/struct.provider.go`

- [ ] **Step 1.1: Write the new DTO file**

```go
package isv

// ---------- service/get_login_info ----------

// LoginInfoResp 是 service/get_login_info 的响应。
// UserType: 1 = 企业管理员,2 = 企业成员,3 = 服务商(代开发)成员。
type LoginInfoResp struct {
	UserType int                 `json:"usertype"`
	UserInfo LoginInfoUser       `json:"user_info"`
	CorpInfo LoginInfoCorp       `json:"corp_info"`
	Agent    []LoginInfoAgent    `json:"agent"`
	AuthInfo LoginInfoPermission `json:"auth_info"`
}

// LoginInfoUser 登录者自身信息。
type LoginInfoUser struct {
	UserID     string `json:"userid"`
	OpenUserID string `json:"open_userid"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
}

// LoginInfoCorp 登录者所属企业。
type LoginInfoCorp struct {
	CorpID string `json:"corpid"`
}

// LoginInfoAgent 第三方应用在该企业下的 agent。
type LoginInfoAgent struct {
	AgentID  int `json:"agentid"`
	AuthType int `json:"auth_type"` // 0 = 只使用,1 = 管理
}

// LoginInfoPermission —— 仅当 UserType=1(管理员)时非空,列出被管理的部门。
type LoginInfoPermission struct {
	Department []LoginInfoDepartment `json:"department"`
}

// LoginInfoDepartment 被管理的部门。
type LoginInfoDepartment struct {
	ID       int  `json:"id"`
	Writable bool `json:"writable"`
}

// ---------- service/get_register_code ----------

// GetRegisterCodeReq 是 service/get_register_code 的请求体。
// 所有字段均为可选,服务端按缺省值处理。
type GetRegisterCodeReq struct {
	TemplateID  string `json:"template_id,omitempty"`
	CorpName    string `json:"corp_name,omitempty"`
	AdminName   string `json:"admin_name,omitempty"`
	AdminMobile string `json:"admin_mobile,omitempty"`
	State       string `json:"state,omitempty"`
}

// RegisterCodeResp 是 service/get_register_code 的响应。
type RegisterCodeResp struct {
	RegisterCode string `json:"register_code"`
	ExpiresIn    int    `json:"expires_in"`
}

// ---------- service/get_registration_info ----------

// RegistrationInfoResp 是 service/get_registration_info 的响应。
// AuthInfo 复用子项目 1 的 AuthInfoAgent(字段布局一致)。
type RegistrationInfoResp struct {
	CorpInfo      RegistrationCorpInfo  `json:"corp_info"`
	AuthUserInfo  RegistrationAdminInfo `json:"auth_user_info"`
	ContactSync   RegistrationContact   `json:"contact_sync"`
	AuthInfo      AuthInfoAgent         `json:"auth_info"`
	PermanentCode string                `json:"permanent_code"`
}

// RegistrationCorpInfo 已注册企业的信息快照。
type RegistrationCorpInfo struct {
	CorpID            string `json:"corpid"`
	CorpName          string `json:"corp_name"`
	CorpType          string `json:"corp_type"`
	CorpSquareLogoURL string `json:"corp_square_logo_url"`
	CorpUserMax       int    `json:"corp_user_max"`
	SubjectType       int    `json:"subject_type"`
	VerifiedEndTime   int    `json:"verified_end_time"`
	CorpWxqrcode      string `json:"corp_wxqrcode"`
	CorpScale         string `json:"corp_scale"`
	CorpIndustry      string `json:"corp_industry"`
	CorpSubIndustry   string `json:"corp_sub_industry"`
}

// RegistrationAdminInfo 注册企业的初始管理员。
type RegistrationAdminInfo struct {
	UserID string `json:"userid"`
	Name   string `json:"name"`
}

// RegistrationContact —— 注册完成后返回的通讯录同步 token(单次有效)。
type RegistrationContact struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
```

- [ ] **Step 1.2: Verify compilation**

Run: `go build ./work-wechat/isv/...`
Expected: clean build (no output). This proves the DTOs compile and `AuthInfoAgent` is found.

- [ ] **Step 1.3: Commit**

```bash
git add work-wechat/isv/struct.provider.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add provider login DTOs

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: `GetProviderAccessToken` public wrapper

**Files:**
- Modify: `work-wechat/isv/provider.id_convert.go` — add one public method at the bottom
- Modify: `work-wechat/isv/provider.id_convert_test.go` — add 2 tests

- [ ] **Step 2.1: Write the failing tests**

Append to `work-wechat/isv/provider.id_convert_test.go`:

```go
func TestGetProviderAccessToken_Exposed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/service/get_provider_token" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"provider_access_token": "PTOK_PUB",
			"expires_in":            7200,
		})
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	tok, err := c.GetProviderAccessToken(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tok != "PTOK_PUB" {
		t.Errorf("tok: %q", tok)
	}
}

func TestGetProviderAccessToken_MissingConfig(t *testing.T) {
	cfg := testConfig() // sub-project 1 helper — no ProviderCorpID / ProviderSecret
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.GetProviderAccessToken(context.Background())
	if !errors.Is(err, ErrProviderCorpIDMissing) {
		t.Fatalf("want ErrProviderCorpIDMissing, got %v", err)
	}
}
```

- [ ] **Step 2.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run "TestGetProviderAccessToken" -v`
Expected: FAIL — `c.GetProviderAccessToken undefined`.

- [ ] **Step 2.3: Add the public wrapper**

Append to `work-wechat/isv/provider.id_convert.go` (after `UserIDToOpenUserID`):

```go
// GetProviderAccessToken 返回当前 provider_access_token(lazy 获取 + 自动缓存)。
// 未在 Config 里配置 ProviderCorpID / ProviderSecret 时返回对应哨兵错误。
func (c *Client) GetProviderAccessToken(ctx context.Context) (string, error) {
	return c.getProviderAccessToken(ctx)
}
```

- [ ] **Step 2.4: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run "TestGetProviderAccessToken" -v`
Expected: both PASS.

- [ ] **Step 2.5: Commit**

```bash
git add work-wechat/isv/provider.id_convert.go work-wechat/isv/provider.id_convert_test.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): expose GetProviderAccessToken public wrapper

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: `GetLoginInfo`

**Files:**
- Create: `work-wechat/isv/provider.login.go`
- Create: `work-wechat/isv/provider.login_test.go`

- [ ] **Step 3.1: Write the failing tests**

Create `work-wechat/isv/provider.login_test.go`:

```go
package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetLoginInfo_Admin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_login_info":
			if got := r.URL.Query().Get("provider_access_token"); got != "PTOK" {
				t.Errorf("token query: %q", got)
			}
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["auth_code"] != "AUTH1" {
				t.Errorf("auth_code body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"usertype": 1,
				"user_info": map[string]interface{}{
					"userid":      "admin1",
					"open_userid": "oadmin1",
					"name":        "Admin",
					"avatar":      "http://img/a.png",
				},
				"corp_info": map[string]interface{}{"corpid": "wxcorp1"},
				"agent": []map[string]interface{}{
					{"agentid": 1000001, "auth_type": 1},
				},
				"auth_info": map[string]interface{}{
					"department": []map[string]interface{}{
						{"id": 1, "writable": true},
						{"id": 2, "writable": false},
					},
				},
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetLoginInfo(context.Background(), "AUTH1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.UserType != 1 {
		t.Errorf("usertype: %d", resp.UserType)
	}
	if resp.UserInfo.UserID != "admin1" || resp.UserInfo.OpenUserID != "oadmin1" {
		t.Errorf("user_info: %+v", resp.UserInfo)
	}
	if resp.CorpInfo.CorpID != "wxcorp1" {
		t.Errorf("corp_info: %+v", resp.CorpInfo)
	}
	if len(resp.Agent) != 1 || resp.Agent[0].AgentID != 1000001 || resp.Agent[0].AuthType != 1 {
		t.Errorf("agent: %+v", resp.Agent)
	}
	if len(resp.AuthInfo.Department) != 2 || !resp.AuthInfo.Department[0].Writable {
		t.Errorf("auth_info: %+v", resp.AuthInfo)
	}
}

func TestGetLoginInfo_Member(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_login_info":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"usertype": 2,
				"user_info": map[string]interface{}{
					"userid":      "u1",
					"open_userid": "ou1",
					"name":        "Alice",
				},
				"corp_info": map[string]interface{}{"corpid": "wxcorp1"},
				"agent": []map[string]interface{}{
					{"agentid": 1000001, "auth_type": 0},
				},
				// No auth_info.department for a non-admin login.
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetLoginInfo(context.Background(), "AUTH2")
	if err != nil {
		t.Fatal(err)
	}
	if resp.UserType != 2 {
		t.Errorf("usertype: %d", resp.UserType)
	}
	if len(resp.AuthInfo.Department) != 0 {
		t.Errorf("department should be empty for member, got: %+v", resp.AuthInfo.Department)
	}
}

func TestGetLoginInfo_WeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_login_info":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode": 40029,
				"errmsg":  "invalid code",
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	_, err := c.GetLoginInfo(context.Background(), "BAD")
	if err == nil {
		t.Fatal("want error, got nil")
	}
	var we *WeixinError
	if !errorAs(err, &we) || we.ErrCode != 40029 {
		t.Errorf("want *WeixinError errcode=40029, got %v", err)
	}
}

// errorAs is a tiny helper wrapping errors.As to keep test bodies shorter.
func errorAs(err error, target interface{}) bool {
	return errorsAs(err, target)
}
```

Note: the `errorsAs` helper is just `errors.As` — to avoid adding an `errors` import on a test-only helper, add this import group and replace `errorsAs` with `errors.As`:

**Final imports and helper block** (replace the bottom of the test file):

```go
// (Move the helper below the 3 tests and inline errors.As.)
```

Actually, simplify. Delete the `errorAs`/`errorsAs` indirection and use `errors.As` directly. Final test file imports:

```go
import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)
```

And in `TestGetLoginInfo_WeixinError` replace the check with:

```go
	var we *WeixinError
	if !errors.As(err, &we) || we.ErrCode != 40029 {
		t.Errorf("want *WeixinError errcode=40029, got %v", err)
	}
```

Remove the `errorAs` / `errorsAs` helper functions entirely.

- [ ] **Step 3.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run "TestGetLoginInfo" -v`
Expected: FAIL — `c.GetLoginInfo undefined`.

- [ ] **Step 3.3: Implement `GetLoginInfo`**

Create `work-wechat/isv/provider.login.go`:

```go
package isv

import "context"

// GetLoginInfo 用服务商管理端 OAuth 回跳返回的 auth_code 换取登录身份。
// 使用 provider_access_token,不使用 suite_access_token。
func (c *Client) GetLoginInfo(ctx context.Context, authCode string) (*LoginInfoResp, error) {
	body := map[string]string{"auth_code": authCode}
	var resp LoginInfoResp
	if err := c.providerDoPost(ctx, "/cgi-bin/service/get_login_info", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 3.4: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run "TestGetLoginInfo" -v`
Expected: all 3 PASS.

- [ ] **Step 3.5: Commit**

```bash
git add work-wechat/isv/provider.login.go work-wechat/isv/provider.login_test.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add GetLoginInfo (provider OAuth auth_code exchange)

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: `GetRegisterCode` + `GetRegistrationInfo`

**Files:**
- Modify: `work-wechat/isv/provider.login.go` — append 2 methods
- Modify: `work-wechat/isv/provider.login_test.go` — append 3 tests

- [ ] **Step 4.1: Write the failing tests**

Append to `work-wechat/isv/provider.login_test.go`:

```go
func TestGetRegisterCode_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_register_code":
			if got := r.URL.Query().Get("provider_access_token"); got != "PTOK" {
				t.Errorf("token query: %q", got)
			}
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["template_id"] != "tmpl1" || body["corp_name"] != "ACME" || body["state"] != "s1" {
				t.Errorf("body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"register_code": "REG123",
				"expires_in":    604800,
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetRegisterCode(context.Background(), &GetRegisterCodeReq{
		TemplateID: "tmpl1",
		CorpName:   "ACME",
		State:      "s1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.RegisterCode != "REG123" || resp.ExpiresIn != 604800 {
		t.Errorf("resp: %+v", resp)
	}
}

func TestGetRegisterCode_NilRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_register_code":
			// Empty body is acceptable — all fields are optional.
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"register_code": "REG_EMPTY",
				"expires_in":    3600,
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetRegisterCode(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.RegisterCode != "REG_EMPTY" {
		t.Errorf("resp: %+v", resp)
	}
}

func TestGetRegistrationInfo_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/service/get_provider_token":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"provider_access_token": "PTOK",
				"expires_in":            7200,
			})
		case "/cgi-bin/service/get_registration_info":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["register_code"] != "REG123" {
				t.Errorf("body: %+v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"corp_info": map[string]interface{}{
					"corpid":         "wxcorp1",
					"corp_name":      "ACME",
					"corp_user_max":  200,
					"subject_type":   1,
					"corp_industry":  "Tech",
				},
				"auth_user_info": map[string]interface{}{
					"userid": "admin1",
					"name":   "Root",
				},
				"contact_sync": map[string]interface{}{
					"access_token": "CTOK",
					"expires_in":   7200,
				},
				"auth_info": map[string]interface{}{
					"agent": []map[string]interface{}{
						{"agentid": 1000001, "name": "HR"},
					},
				},
				"permanent_code": "PERM_REG",
			})
		}
	}))
	defer srv.Close()

	c := newTestISVClientWithProvider(t, srv.URL)
	resp, err := c.GetRegistrationInfo(context.Background(), "REG123")
	if err != nil {
		t.Fatal(err)
	}
	if resp.CorpInfo.CorpID != "wxcorp1" || resp.CorpInfo.CorpName != "ACME" {
		t.Errorf("corp_info: %+v", resp.CorpInfo)
	}
	if resp.AuthUserInfo.UserID != "admin1" {
		t.Errorf("auth_user_info: %+v", resp.AuthUserInfo)
	}
	if resp.ContactSync.AccessToken != "CTOK" {
		t.Errorf("contact_sync: %+v", resp.ContactSync)
	}
	if len(resp.AuthInfo.Agent) != 1 || resp.AuthInfo.Agent[0].AgentID != 1000001 {
		t.Errorf("auth_info: %+v", resp.AuthInfo)
	}
	if resp.PermanentCode != "PERM_REG" {
		t.Errorf("permanent_code: %q", resp.PermanentCode)
	}
}
```

- [ ] **Step 4.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run "TestGetRegisterCode|TestGetRegistrationInfo" -v`
Expected: FAIL — `c.GetRegisterCode undefined` / `c.GetRegistrationInfo undefined`.

- [ ] **Step 4.3: Implement both methods**

Append to `work-wechat/isv/provider.login.go`:

```go
// GetRegisterCode 生成注册企业微信的 register_code(邀请链接的核心参数)。
// 所有字段都是可选 —— 服务端按缺省值处理。
func (c *Client) GetRegisterCode(ctx context.Context, req *GetRegisterCodeReq) (*RegisterCodeResp, error) {
	if req == nil {
		req = &GetRegisterCodeReq{}
	}
	var resp RegisterCodeResp
	if err := c.providerDoPost(ctx, "/cgi-bin/service/get_register_code", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRegistrationInfo 查询 register_code 对应的注册进度,
// 成功后返回已注册企业的 corpid / 管理员 / 永久授权码 / 通讯录同步 token。
func (c *Client) GetRegistrationInfo(ctx context.Context, registerCode string) (*RegistrationInfoResp, error) {
	body := map[string]string{"register_code": registerCode}
	var resp RegistrationInfoResp
	if err := c.providerDoPost(ctx, "/cgi-bin/service/get_registration_info", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 4.4: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run "TestGetRegisterCode|TestGetRegistrationInfo" -v`
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
- coverage: ≥85% (sub-project 1 was 86.1%; this task adds ~60 prod LOC and ~220 test LOC, so coverage should stay ≥85%)

- [ ] **Step 4.6: Full repo regression**

Run: `go test ./... -count=1`
Expected: every package green (including sub-project 1 tests, oplatform, merchant, etc.).

- [ ] **Step 4.7: Commit**

```bash
git add work-wechat/isv/provider.login.go work-wechat/isv/provider.login_test.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add GetRegisterCode + GetRegistrationInfo

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 4.8: Tag final state and report**

Run the following and report the output:

```bash
git log --oneline -6
go test -cover ./work-wechat/isv/...
```

Expected: 4 new `feat(work-wechat/isv)` commits on top of the sub-project 2 spec commit (`40df5cc`), and the coverage line.

---

## Self-Review

### Spec coverage

| Spec requirement | Task |
|---|---|
| §1.1 #1 `GetProviderAccessToken` | Task 2 |
| §1.1 #2 `GetLoginInfo` | Task 3 |
| §1.1 #3 `GetRegisterCode` | Task 4 |
| §1.1 #4 `GetRegistrationInfo` | Task 4 |
| §3.1 `LoginInfoResp` + nested types | Task 1 |
| §3.2 `GetRegisterCodeReq` / `RegisterCodeResp` | Task 1 |
| §3.3 `RegistrationInfoResp` + nested types, reuse of `AuthInfoAgent` | Task 1 |
| §4.1 — §4.4 method bodies | Tasks 2, 3, 4 |
| §5.1 7 test cases | Task 2 (2) + Task 3 (3) + Task 4 (3) = **8** ✔ (one extra — `TestGetRegisterCode_NilRequest`) |
| §6 no new errors | ✔ — no task introduces new sentinels |
| §8 scale estimate ~154 prod / ~220 test | Task 1 ~95 + Task 2 ~5 + Task 3 ~15 + Task 4 ~25 = ~140 prod ✔ |
| §9 ~4 commits | 4 task commits ✔ |
| §10 example/main.go (optional) | **Intentionally deferred** — the spec marks this optional; skipped to keep the plan focused. |

### Placeholder scan
- No TBD / TODO / "implement later".
- Every code step contains complete Go code.
- Every test step contains complete Go test code.
- Every command step has the exact command + expected output.
- Step 3.1 contains inline revision instructions (delete `errorAs`/`errorsAs`, use `errors.As` directly). This is not a placeholder — it's a concrete rewrite instruction.

### Type / name consistency
- `LoginInfoResp` / `LoginInfoUser` / `LoginInfoCorp` / `LoginInfoAgent` / `LoginInfoPermission` / `LoginInfoDepartment`: used in Task 1, referenced in Task 3's test assertions (`resp.UserInfo.UserID`, `resp.Agent[0].AuthType`, `resp.AuthInfo.Department`). ✔
- `GetRegisterCodeReq` fields (`TemplateID` / `CorpName` / `State`): used in Task 4 test body construction. ✔
- `RegistrationInfoResp.AuthInfo.Agent[0].AgentID`: correctly accesses `AuthInfoAgent.Agent[0].AgentID` from `struct.permanent.go`. ✔
- `ErrProviderCorpIDMissing`: defined in sub-project 1 `errors.go`, used in Task 2. ✔
- `WeixinError`: defined in sub-project 1, used in Task 3 error test. ✔
- `newTestISVClientWithProvider`: defined in sub-project 1 `provider.id_convert_test.go`, used in all remote-path tests. ✔
- `testConfig`: defined in sub-project 1 `client_test.go`, used in Task 2 `TestGetProviderAccessToken_MissingConfig`. ✔

**Plan is self-consistent and fully covers the spec.**
