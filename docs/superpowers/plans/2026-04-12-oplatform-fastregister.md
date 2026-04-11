# oplatform FastRegister Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement 8 代注册小程序 methods (7 component-level on a new `FastRegisterClient` + 1 authorizer-level on existing `WxaAdminClient`).

**Architecture:** `FastRegisterClient` is a stateless wrapper around `*Client` returned by `c.FastRegister()`. It has its own `doPost` helper that uses `component_access_token` (rather than WxaAdmin's `access_token`) and handles paths that already contain `?action=create` / `?action=search`. The single authorizer-level method (`GetAccountBasicInfo`) is appended to the existing `wxa.account.go`.

**Tech Stack:** Go 1.23, stdlib only (`encoding/json`, `net/url`, `strings`), existing `utils.HTTP`, `httptest`. Zero new deps.

**Spec:** `docs/superpowers/specs/2026-04-12-oplatform-fastregister-design.md`

**Module path:** `github.com/godrealms/go-wechat-sdk`

---

## Conventions

- Every task ends with a commit. Stage ONLY the files the task touches. Never `git add -A`.
- TDD: failing test first → verify failure → minimal implementation → verify pass → commit.
- `go test ./oplatform/...` must be green before each commit.
- Paths are relative to `/Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk`.
- Working tree has unrelated WIP — stay focused on the listed files only.

### Shared context (already exists)

- `oplatform.Client` with private `http *utils.HTTP`, `store Store`, and method `ComponentAccessToken(ctx) (string, error)`.
- Package-level `decodeRaw(path string, raw json.RawMessage, out any) error` in `wxa.client.go` — two-pass JSON decoding (errcode folding + typed unmarshal).
- Package-level `touchContext(ctx) ctx` and `checkWeixinErr(errcode, errmsg) error` in `client.go`.
- `WeixinError{ErrCode int, ErrMsg string}` in `errors.go`.
- Test helper `newTestClient(t *testing.T, baseURL string, opts ...Option) *Client` in `component_test.go` — already exists.
- `WithStore(Store)` option.
- `NewMemoryStore()` with `SetVerifyTicket`, `SetComponentToken(ctx, token, expireAt time.Time)`, `SetAuthorizer(ctx, appid, tokens)`.
- `newTestWxaAdmin(t, baseURL)` helper in `wxa.client_test.go` for authorizer-scoped tests.

---

## Task 1: `FastRegisterClient` skeleton + shared `doPost` + bootstrap test

**Goal:** Introduce `FastRegisterClient` type, the `Client.FastRegister()` factory, the shared `doPost` helper (with `?action=xxx` handling), an empty `fastregister.struct.go`, and a single test that exercises the helper against a fake endpoint. No business methods yet.

**Files:**
- Create: `oplatform/fastregister.go`
- Create: `oplatform/fastregister.struct.go`
- Create: `oplatform/fastregister_test.go`

- [ ] **Step 1.1: Write failing test `oplatform/fastregister_test.go`**

```go
package oplatform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestFastRegister seeds the store with a non-expired component token
// so ComponentAccessToken() returns from cache and test mocks don't need
// to handle /cgi-bin/component/api_component_token.
func newTestFastRegister(t *testing.T, baseURL string) *FastRegisterClient {
	t.Helper()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetComponentToken(context.Background(), "CTOK", time.Now().Add(time.Hour))
	c := newTestClient(t, baseURL, WithStore(store))
	return c.FastRegister()
}

func TestFastRegister_DoPost_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/fake_endpoint") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("component_access_token") != "CTOK" {
			t.Errorf("missing component_access_token, got %q", r.URL.Query().Get("component_access_token"))
		}
		_, _ = w.Write([]byte(`{"errcode":0,"value":42}`))
	}))
	defer srv.Close()

	f := newTestFastRegister(t, srv.URL)
	var out struct {
		Value int `json:"value"`
	}
	if err := f.doPost(context.Background(), "/fake_endpoint", map[string]string{"k": "v"}, &out); err != nil {
		t.Fatal(err)
	}
	if out.Value != 42 {
		t.Errorf("got %d", out.Value)
	}
}

func TestFastRegister_DoPost_PathWithExistingQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("action") != "create" {
			t.Errorf("missing action=create: %s", r.URL.RawQuery)
		}
		if q.Get("component_access_token") != "CTOK" {
			t.Errorf("missing component_access_token: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()

	f := newTestFastRegister(t, srv.URL)
	if err := f.doPost(context.Background(), "/fake_endpoint?action=create", nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestFastRegister_DoPost_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":89249,"errmsg":"still creating"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	err := f.doPost(context.Background(), "/fake_endpoint", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 89249 {
		t.Errorf("expected WeixinError 89249, got %v", err)
	}
}
```

- [ ] **Step 1.2: Run failing test**

Run: `go test ./oplatform/ -run TestFastRegister_DoPost`
Expected: build error — `FastRegisterClient`, `FastRegister`, `doPost` undefined.

- [ ] **Step 1.3: Create empty `oplatform/fastregister.struct.go`**

```go
package oplatform

// 本文件汇总快速注册 (FastRegister) 子项目的请求/响应 DTO。
// 和 WxaAdmin 的 DTO 分开，因为它们不属于 WxaAdmin 家族。
```

- [ ] **Step 1.4: Create `oplatform/fastregister.go`**

```go
package oplatform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// FastRegisterClient 提供开放平台代注册小程序相关的 component 级接口。
//
// 所有方法都以 Client.ComponentAccessToken() 作为 token 源（而不是 authorizer
// 级别的 token），因为快速注册流程发生在 authorizer 关系建立之前。
//
// FastRegisterClient 无状态，线程安全，可在多 goroutine 共享。
type FastRegisterClient struct {
	c *Client
}

// FastRegister 从 Client 构造 FastRegisterClient。构造不做 I/O。
func (c *Client) FastRegister() *FastRegisterClient {
	return &FastRegisterClient{c: c}
}

// doPost 通用 POST 辅助：
//   - 从 Client 取 component_access_token 并以 query 参数形式拼接
//   - 正确处理 path 中已经带有 ?action=xxx 的情况
//   - 复用包级 decodeRaw 进行两段式 JSON 解码（errcode 折叠 + typed unmarshal）
func (f *FastRegisterClient) doPost(ctx context.Context, path string, body, out any) error {
	ctx = touchContext(ctx)
	token, err := f.c.ComponentAccessToken(ctx)
	if err != nil {
		return err
	}
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	fullPath := path + sep + "component_access_token=" + url.QueryEscape(token)

	var raw json.RawMessage
	if err := f.c.http.Post(ctx, fullPath, body, &raw); err != nil {
		return fmt.Errorf("oplatform: %s: %w", path, err)
	}
	return decodeRaw(path, raw, out)
}
```

- [ ] **Step 1.5: Run tests**

Run: `go test ./oplatform/ -run TestFastRegister_DoPost`
Expected: 3 tests PASS.

Run: `go test ./oplatform/...`
Expected: all existing oplatform tests still pass.

Run: `go build ./...`
Expected: clean.

- [ ] **Step 1.6: Commit**

Stage ONLY the 3 files; verify with `git diff --cached --stat`.

```bash
git add oplatform/fastregister.go oplatform/fastregister.struct.go oplatform/fastregister_test.go
git commit -m "feat(oplatform): add FastRegisterClient skeleton + doPost helper

FastRegisterClient wraps *Client and uses component_access_token
as the query key (not access_token), mirroring the official
代注册 接口 convention. doPost correctly handles endpoints that
already contain ?action=create/search queries by switching the
separator to &. Reuses the package-level decodeRaw helper for
two-pass JSON decoding.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: FastRegister enterprise + personal + beta + admin rebind (7 methods)

**Goal:** Implement the 7 component-level business methods and their tests. Append DTOs to `fastregister.struct.go` along the way.

**Files:**
- Modify: `oplatform/fastregister.go` (append 7 methods)
- Modify: `oplatform/fastregister.struct.go` (append DTOs)
- Modify: `oplatform/fastregister_test.go` (append 8 new tests)

- [ ] **Step 2.1: Append DTOs to `oplatform/fastregister.struct.go`**

Append below the existing header comment:

```go

// ----- enterprise -----

type FastRegEnterpriseReq struct {
	Name               string `json:"name"`                 // 企业名（和工商注册一致）
	Code               string `json:"code"`                 // 企业代码
	CodeType           int    `json:"code_type"`            // 1=统一社会信用代码 2=组织机构代码 3=营业执照注册号
	LegalPersonaWechat string `json:"legal_persona_wechat"` // 法人微信号
	LegalPersonaName   string `json:"legal_persona_name"`   // 法人姓名
	ComponentPhone     string `json:"component_phone"`      // 第三方联系电话
}

type FastRegEnterpriseResp struct{}

type FastRegEnterpriseStatus struct {
	Status          int    `json:"status"`
	AuthCode        string `json:"auth_code,omitempty"`
	AuthorizerAppid string `json:"authorizer_appid,omitempty"`
	IsWxVerify      bool   `json:"is_wx_verify,omitempty"`
	IsLinkMp        bool   `json:"is_link_mp,omitempty"`
}

// ----- personal -----

type FastRegPersonalReq struct {
	IDName         string `json:"idname"`
	WxUser         string `json:"wxuser"`
	ComponentPhone string `json:"component_phone,omitempty"`
}

type FastRegPersonalResp struct {
	TaskID string `json:"taskid"`
}

type FastRegPersonalStatus struct {
	Status            int    `json:"status"`
	AppID             string `json:"appid,omitempty"`
	AuthorizationCode string `json:"authorization_code,omitempty"`
}

// ----- beta (复用主体试用版) -----

type FastRegBetaReq struct {
	Name               string `json:"name"`
	Code               string `json:"code"`
	CodeType           int    `json:"code_type"`
	LegalPersonaWechat string `json:"legal_persona_wechat"`
	LegalPersonaName   string `json:"legal_persona_name"`
	ComponentPhone     string `json:"component_phone"`
}

type FastRegBetaResp struct {
	UniqueID string `json:"unique_id"`
}

type FastRegBetaStatus struct {
	Status            int    `json:"status"`
	AppID             string `json:"appid,omitempty"`
	AuthorizationCode string `json:"authorization_code,omitempty"`
}

// ----- admin rebind -----

type RebindAdminQrcode struct {
	TaskID    string `json:"taskid"`
	QrcodeURL string `json:"qrcode_url"`
}
```

- [ ] **Step 2.2: Append failing tests to `oplatform/fastregister_test.go`**

Add these test functions at the end of the file (imports — `encoding/json` and `io` — may need to be added to the existing import block):

```go
func TestFastRegister_CreateEnterpriseAccount(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/component/fastregisterweapp") {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("action") != "create" {
			t.Errorf("action: %s", r.URL.Query().Get("action"))
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	_, err := f.CreateEnterpriseAccount(context.Background(), &FastRegEnterpriseReq{
		Name:               "Acme Corp",
		Code:               "91310000123456789X",
		CodeType:           1,
		LegalPersonaWechat: "legal_wx",
		LegalPersonaName:   "张三",
		ComponentPhone:     "13800000000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["name"] != "Acme Corp" {
		t.Errorf("name: %+v", body)
	}
}

func TestFastRegister_QueryEnterpriseAccount(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("action") != "search" {
			t.Errorf("action: %s", r.URL.Query().Get("action"))
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"status":1,"auth_code":"AC","authorizer_appid":"wxNEW"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	status, err := f.QueryEnterpriseAccount(context.Background(), "legal_wx", "张三")
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != 1 || status.AuthorizerAppid != "wxNEW" {
		t.Errorf("unexpected: %+v", status)
	}
	if body["legal_persona_wechat"] != "legal_wx" || body["legal_persona_name"] != "张三" {
		t.Errorf("body: %+v", body)
	}
}

func TestFastRegister_CreatePersonalAccount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/component/fastregisterpersonalweapp") {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("action") != "create" {
			t.Errorf("action: %s", r.URL.Query().Get("action"))
		}
		_, _ = w.Write([]byte(`{"errcode":0,"taskid":"T123"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	resp, err := f.CreatePersonalAccount(context.Background(), &FastRegPersonalReq{
		IDName: "李四",
		WxUser: "user_wx",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.TaskID != "T123" {
		t.Errorf("taskid: %q", resp.TaskID)
	}
}

func TestFastRegister_QueryPersonalAccount(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("action") != "query" {
			t.Errorf("action: %s", r.URL.Query().Get("action"))
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"status":1,"appid":"wxPER","authorization_code":"AC"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	status, err := f.QueryPersonalAccount(context.Background(), "T123")
	if err != nil {
		t.Fatal(err)
	}
	if status.AppID != "wxPER" {
		t.Errorf("appid: %q", status.AppID)
	}
	if body["taskid"] != "T123" {
		t.Errorf("body: %+v", body)
	}
}

func TestFastRegister_CreateBetaAccount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/component/fastregisterbetaweapp") {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("action") != "create" {
			t.Errorf("action: %s", r.URL.Query().Get("action"))
		}
		_, _ = w.Write([]byte(`{"errcode":0,"unique_id":"U1"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	resp, err := f.CreateBetaAccount(context.Background(), &FastRegBetaReq{
		Name:               "Acme",
		Code:               "91310000123456789X",
		CodeType:           1,
		LegalPersonaWechat: "legal_wx",
		LegalPersonaName:   "张三",
		ComponentPhone:     "13800000000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.UniqueID != "U1" {
		t.Errorf("unique_id: %q", resp.UniqueID)
	}
}

func TestFastRegister_QueryBetaAccount(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("action") != "search" {
			t.Errorf("action: %s", r.URL.Query().Get("action"))
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"status":1,"appid":"wxBETA"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	status, err := f.QueryBetaAccount(context.Background(), "U1")
	if err != nil {
		t.Fatal(err)
	}
	if status.AppID != "wxBETA" {
		t.Errorf("appid: %q", status.AppID)
	}
	if body["unique_id"] != "U1" {
		t.Errorf("body: %+v", body)
	}
}

func TestFastRegister_GenerateAdminRebindQrcode(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/account/componentrebindadmin") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"taskid":"T1","qrcode_url":"https://mp.weixin.qq.com/qr/ABC"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	qr, err := f.GenerateAdminRebindQrcode(context.Background(), "https://example.com/rebind/cb")
	if err != nil {
		t.Fatal(err)
	}
	if qr.TaskID != "T1" || qr.QrcodeURL == "" {
		t.Errorf("unexpected: %+v", qr)
	}
	if body["redirect_uri"] != "https://example.com/rebind/cb" {
		t.Errorf("body: %+v", body)
	}
}

func TestFastRegister_CreateEnterpriseAccount_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":89247,"errmsg":"duplicate account"}`))
	}))
	defer srv.Close()
	f := newTestFastRegister(t, srv.URL)

	_, err := f.CreateEnterpriseAccount(context.Background(), &FastRegEnterpriseReq{
		Name: "Acme",
	})
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 89247 {
		t.Errorf("expected 89247, got %v", err)
	}
}
```

Make sure the test file's import block contains:
`context`, `encoding/json`, `errors`, `io`, `net/http`, `net/http/httptest`, `strings`, `testing`, `time`.

- [ ] **Step 2.3: Run failing tests — expect undefined methods.**

Run: `go test ./oplatform/ -run TestFastRegister_`
Expected: undefined `CreateEnterpriseAccount`, `QueryEnterpriseAccount`, `CreatePersonalAccount`, `QueryPersonalAccount`, `CreateBetaAccount`, `QueryBetaAccount`, `GenerateAdminRebindQrcode`.

- [ ] **Step 2.4: Append 7 business methods to `oplatform/fastregister.go`**

Append below the existing `doPost` method:

```go

// CreateEnterpriseAccount 企业快速注册小程序。
// /cgi-bin/component/fastregisterweapp?action=create
func (f *FastRegisterClient) CreateEnterpriseAccount(ctx context.Context, req *FastRegEnterpriseReq) (*FastRegEnterpriseResp, error) {
	var resp FastRegEnterpriseResp
	if err := f.doPost(ctx, "/cgi-bin/component/fastregisterweapp?action=create", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryEnterpriseAccount 查询企业快速注册任务状态。
// /cgi-bin/component/fastregisterweapp?action=search
func (f *FastRegisterClient) QueryEnterpriseAccount(ctx context.Context, legalPersonaWechat, legalPersonaName string) (*FastRegEnterpriseStatus, error) {
	body := map[string]string{
		"legal_persona_wechat": legalPersonaWechat,
		"legal_persona_name":   legalPersonaName,
	}
	var resp FastRegEnterpriseStatus
	if err := f.doPost(ctx, "/cgi-bin/component/fastregisterweapp?action=search", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreatePersonalAccount 个人类型小程序快速注册。
// /cgi-bin/component/fastregisterpersonalweapp?action=create
func (f *FastRegisterClient) CreatePersonalAccount(ctx context.Context, req *FastRegPersonalReq) (*FastRegPersonalResp, error) {
	var resp FastRegPersonalResp
	if err := f.doPost(ctx, "/cgi-bin/component/fastregisterpersonalweapp?action=create", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryPersonalAccount 查询个人注册任务状态。
// /cgi-bin/component/fastregisterpersonalweapp?action=query
func (f *FastRegisterClient) QueryPersonalAccount(ctx context.Context, taskID string) (*FastRegPersonalStatus, error) {
	body := map[string]string{"taskid": taskID}
	var resp FastRegPersonalStatus
	if err := f.doPost(ctx, "/cgi-bin/component/fastregisterpersonalweapp?action=query", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateBetaAccount 复用主体创建试用版小程序。
// /cgi-bin/component/fastregisterbetaweapp?action=create
func (f *FastRegisterClient) CreateBetaAccount(ctx context.Context, req *FastRegBetaReq) (*FastRegBetaResp, error) {
	var resp FastRegBetaResp
	if err := f.doPost(ctx, "/cgi-bin/component/fastregisterbetaweapp?action=create", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryBetaAccount 查询试用版创建任务状态。
// /cgi-bin/component/fastregisterbetaweapp?action=search
func (f *FastRegisterClient) QueryBetaAccount(ctx context.Context, uniqueID string) (*FastRegBetaStatus, error) {
	body := map[string]string{"unique_id": uniqueID}
	var resp FastRegBetaStatus
	if err := f.doPost(ctx, "/cgi-bin/component/fastregisterbetaweapp?action=search", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GenerateAdminRebindQrcode 生成小程序管理员变更二维码。
// /cgi-bin/account/componentrebindadmin
func (f *FastRegisterClient) GenerateAdminRebindQrcode(ctx context.Context, redirectURI string) (*RebindAdminQrcode, error) {
	body := map[string]string{"redirect_uri": redirectURI}
	var resp RebindAdminQrcode
	if err := f.doPost(ctx, "/cgi-bin/account/componentrebindadmin", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 2.5: Run tests**

Run: `go test ./oplatform/ -run TestFastRegister -v`
Expected: 11 tests PASS (3 helper tests from Task 1 + 8 new business tests).

Run: `go test ./oplatform/...`
Expected: entire oplatform suite green.

Run: `go build ./...`
Expected: clean.

- [ ] **Step 2.6: Commit**

```bash
git add oplatform/fastregister.go oplatform/fastregister.struct.go oplatform/fastregister_test.go
git commit -m "feat(oplatform): add 7 FastRegister component-level methods

CreateEnterpriseAccount / QueryEnterpriseAccount /
CreatePersonalAccount / QueryPersonalAccount /
CreateBetaAccount / QueryBetaAccount /
GenerateAdminRebindQrcode — covers enterprise, personal, and
reused-principal beta registration plus the admin rebind
qrcode generator. DTOs appended to fastregister.struct.go.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: WxaAdmin `GetAccountBasicInfo` (authorizer-level)

**Goal:** Add the one authorizer-level method that doesn't fit FastRegister's component-only model. Appends to existing `wxa.*` files.

**Files:**
- Modify: `oplatform/wxa.struct.go` (append `WxaAccountBasicInfo` DTO)
- Modify: `oplatform/wxa.account.go` (append `GetAccountBasicInfo` method)
- Modify: `oplatform/wxa.account_test.go` (append 1 test)

- [ ] **Step 3.1: Append DTO to `oplatform/wxa.struct.go`**

Append at the end:

```go

// ----- account basic info -----

type WxaAccountBasicInfo struct {
	AppID             string `json:"appid"`
	AccountType       int    `json:"account_type"`
	PrincipalType     int    `json:"principal_type"`
	PrincipalName     string `json:"principal_name"`
	RealnameStatus    int    `json:"realname_status"`
	Nickname          string `json:"nickname,omitempty"`
	HeadImg           string `json:"head_img,omitempty"`
	Signature         string `json:"signature,omitempty"`
	RegisteredCountry int    `json:"registered_country,omitempty"`
	WxVerifyInfo      struct {
		QualificationVerify bool `json:"qualification_verify"`
		NamingVerify        bool `json:"naming_verify"`
	} `json:"wx_verify_info,omitempty"`
	SignatureInfo struct {
		Signature       string `json:"signature"`
		ModifyUsedCount int    `json:"modify_used_count"`
		ModifyQuota     int    `json:"modify_quota"`
	} `json:"signature_info,omitempty"`
	HeadImageInfo struct {
		HeadImageURL    string `json:"head_image_url"`
		ModifyUsedCount int    `json:"modify_used_count"`
		ModifyQuota     int    `json:"modify_quota"`
	} `json:"head_image_info,omitempty"`
}
```

- [ ] **Step 3.2: Append failing test to `oplatform/wxa.account_test.go`**

Append at the end of the file:

```go

func TestWxaAdmin_GetAccountBasicInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/account/getaccountbasicinfo") {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("access_token") != "ATOK" {
			t.Errorf("access_token: %s", r.URL.Query().Get("access_token"))
		}
		_, _ = w.Write([]byte(`{
  "errcode": 0,
  "appid": "wxBIZ",
  "account_type": 0,
  "principal_type": 1,
  "principal_name": "Acme Corp",
  "realname_status": 1,
  "nickname": "Cool App",
  "head_img": "https://example.com/h.png",
  "signature": "hello",
  "registered_country": 1,
  "wx_verify_info": {"qualification_verify": true, "naming_verify": true},
  "signature_info": {"signature": "hello", "modify_used_count": 1, "modify_quota": 5},
  "head_image_info": {"head_image_url": "https://example.com/h.png", "modify_used_count": 0, "modify_quota": 5}
}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	info, err := w.GetAccountBasicInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.AppID != "wxBIZ" || info.PrincipalName != "Acme Corp" {
		t.Errorf("top-level: %+v", info)
	}
	if !info.WxVerifyInfo.QualificationVerify {
		t.Errorf("wx_verify_info not parsed")
	}
	if info.SignatureInfo.ModifyQuota != 5 {
		t.Errorf("signature_info: %+v", info.SignatureInfo)
	}
	if info.HeadImageInfo.HeadImageURL == "" {
		t.Errorf("head_image_info: %+v", info.HeadImageInfo)
	}
}
```

- [ ] **Step 3.3: Run failing test**

Run: `go test ./oplatform/ -run TestWxaAdmin_GetAccountBasicInfo`
Expected: undefined `GetAccountBasicInfo`.

- [ ] **Step 3.4: Append method to `oplatform/wxa.account.go`**

Append at the end of the file (after `ModifySignature`):

```go

// GetAccountBasicInfo 获取小程序账号基本信息。
// /cgi-bin/account/getaccountbasicinfo
// 这是 authorizer 级别接口（需要 authorizer_access_token）；使用 GET。
func (w *WxaAdminClient) GetAccountBasicInfo(ctx context.Context) (*WxaAccountBasicInfo, error) {
	var resp WxaAccountBasicInfo
	if err := w.doGet(ctx, "/cgi-bin/account/getaccountbasicinfo", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 3.5: Run tests**

Run: `go test ./oplatform/ -run TestWxaAdmin_GetAccountBasicInfo`
Expected: PASS.

Run: `go test ./oplatform/...`
Expected: full suite green.

Run: `go test -race ./...`
Expected: all packages green, no races.

Run: `go build ./...`
Expected: clean.

- [ ] **Step 3.6: Commit**

```bash
git add oplatform/wxa.struct.go oplatform/wxa.account.go oplatform/wxa.account_test.go
git commit -m "feat(oplatform): add WxaAdmin.GetAccountBasicInfo

Authorizer-level counterpart to FastRegister — retrieves the
small-program's basic info (principal name, nickname, verify
status, signature/headimg modify quotas). Uses doGet because
the endpoint is a GET and WxaAdminClient's doGet helper already
handles authorizer access_token query wiring.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Final verification sweep

**Goal:** Confirm everything green. No code changes.

- [ ] **Step 4.1: Full build + vet + test**

Run:
```bash
go build ./...
go vet ./...
go test -race ./...
```
All three must be clean, with every package showing `ok`.

- [ ] **Step 4.2: Method count verification**

```bash
grep -hE '^func \(f \*FastRegisterClient\)' oplatform/fastregister.go | grep -v doPost | wc -l
```
Expected: `7`.

```bash
grep -hE '^func \(w \*WxaAdminClient\) GetAccountBasicInfo' oplatform/wxa.account.go | wc -l
```
Expected: `1`.

- [ ] **Step 4.3: Test count verification**

```bash
grep -hE '^func TestFastRegister_' oplatform/fastregister_test.go | wc -l
```
Expected: `11` (3 doPost helper tests + 8 business tests, including 1 errcode).

```bash
grep -hE '^func TestWxaAdmin_GetAccountBasicInfo' oplatform/wxa.account_test.go | wc -l
```
Expected: `1`.

- [ ] **Step 4.4: Git log sanity**

Run: `git log --oneline -5`
Expected: commits from Task 1, Task 2, Task 3 plus the spec commit. Three feat commits plus one docs commit.

No commit at this step — verification only.

---

## Coverage Map (self-review)

| Spec section | Task |
|---|---|
| §2.1 FastRegisterClient + factory | Task 1 |
| §2.2 doPost helper | Task 1 |
| §2.3 query key `component_access_token` | Task 1 (verified in `TestFastRegister_DoPost_Success`) |
| §2.4 `?action=...` path handling | Task 1 (verified in `TestFastRegister_DoPost_PathWithExistingQuery`) |
| §2.5 file layout | Tasks 1-3 |
| §3.1 7 FastRegister methods | Task 2 |
| §3.2 WxaAdmin GetAccountBasicInfo | Task 3 |
| §4.1 FastRegister DTOs | Task 2 |
| §4.2 WxaAccountBasicInfo DTO | Task 3 |
| §5 error handling (*WeixinError) | Task 1 (in `doPost` via `decodeRaw`) + Task 2 errcode test |
| §6 concurrency / lifecycle | Task 1 (stateless wrapper) |
| §7 testing strategy (httptest + newTestFastRegister helper) | Tasks 1-3 |
| §8 compatibility (zero break) | All tasks (additive only) |
| §9 delivery list | Tasks 1-3 |

**All spec requirements covered. No placeholders. Method names consistent:**

- FastRegisterClient methods: `CreateEnterpriseAccount`, `QueryEnterpriseAccount`, `CreatePersonalAccount`, `QueryPersonalAccount`, `CreateBetaAccount`, `QueryBetaAccount`, `GenerateAdminRebindQrcode` (7)
- WxaAdminClient additions: `GetAccountBasicInfo` (1)
- **Total: 8 new methods ✓**

Test count target: 11 FastRegister tests + 1 WxaAdmin test = **12 new tests**.
