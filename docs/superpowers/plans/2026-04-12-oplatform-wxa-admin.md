# oplatform WxaAdmin Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement 31 代小程序开发管理 methods across 6 sub-families (account, category, domain, tester, code, release) via a new `WxaAdminClient` wrapper returned by `AuthorizerClient.WxaAdmin()`.

**Architecture:** `WxaAdminClient` is a thin no-state wrapper around `*AuthorizerClient`. It shares three internal helpers (`doPost`, `doGet`, `doGetRaw`) that handle token-as-query, errcode folding, and two-pass JSON decoding. All 31 methods call one of these helpers. One `.go` + one `*_test.go` file per sub-family; DTOs consolidated in `wxa.struct.go`.

**Tech Stack:** Go 1.23, `encoding/json`, `net/url`, existing `utils.HTTP`, `httptest` for mocking. Zero new dependencies.

**Spec:** `docs/superpowers/specs/2026-04-12-oplatform-wxa-admin-design.md`

**Module path:** `github.com/godrealms/go-wechat-sdk`

---

## Conventions

- Every task ends with a commit. Stage ONLY the files the task touches. Never `git add -A`.
- TDD: failing test first, verify it fails, minimal code to pass, verify, commit.
- Verify nothing broke with `go test ./oplatform/...` before each commit.
- Paths are relative to `/Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk`.
- The working tree has unrelated WIP from the user — stay focused on the listed files only.

### Shared test helper (defined in Task 1, used by Tasks 2-7)

```go
func newTestWxaAdmin(t *testing.T, baseURL string) *WxaAdminClient {
	t.Helper()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxBiz", AuthorizerTokens{
		AccessToken:  "ATOK",
		RefreshToken: "R",
		ExpireAt:     time.Now().Add(time.Hour),
	})
	c := newTestClient(t, baseURL, WithStore(store))
	return c.Authorizer("wxBiz").WxaAdmin()
}
```

This helper pre-seeds the store with a non-expired authorizer token (`ATOK`), so `AuthorizerClient.AccessToken(ctx)` returns from cache and never hits `/cgi-bin/component/api_authorizer_token`. Mocks in Tasks 2-7 only need to handle their specific business endpoint.

---

## Task 1: `WxaAdminClient` + shared helpers + `wxa.struct.go` skeleton

**Goal:** Build the `WxaAdminClient` type, the `AuthorizerClient.WxaAdmin()` factory, and the three shared helpers (`doPost`, `doGet`, `doGetRaw`). Create an empty `wxa.struct.go` for later tasks to fill. Test the helpers directly via httptest with a fake endpoint.

**Files:**
- Create: `oplatform/wxa.client.go`
- Create: `oplatform/wxa.struct.go` (header only)
- Create: `oplatform/wxa.client_test.go`

- [ ] **Step 1.1: Write failing tests in `oplatform/wxa.client_test.go`**

```go
package oplatform

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// Shared test helper used by wxa.*_test.go.
func newTestWxaAdmin(t *testing.T, baseURL string) *WxaAdminClient {
	t.Helper()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	_ = store.SetAuthorizer(context.Background(), "wxBiz", AuthorizerTokens{
		AccessToken:  "ATOK",
		RefreshToken: "R",
		ExpireAt:     time.Now().Add(time.Hour),
	})
	c := newTestClient(t, baseURL, WithStore(store))
	return c.Authorizer("wxBiz").WxaAdmin()
}

func TestWxaAdmin_DoPost_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/fake_endpoint") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("access_token") != "ATOK" {
			t.Errorf("missing access_token, got %q", r.URL.Query().Get("access_token"))
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","value":42}`))
	}))
	defer srv.Close()

	w := newTestWxaAdmin(t, srv.URL)
	var out struct {
		Value int `json:"value"`
	}
	if err := w.doPost(context.Background(), "/wxa/fake_endpoint", map[string]string{"k": "v"}, &out); err != nil {
		t.Fatal(err)
	}
	if out.Value != 42 {
		t.Errorf("got %d, want 42", out.Value)
	}
}

func TestWxaAdmin_DoPost_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":85013,"errmsg":"version not exist"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.doPost(context.Background(), "/wxa/fake_endpoint", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 85013 {
		t.Errorf("expected WeixinError 85013, got %v", err)
	}
}

func TestWxaAdmin_DoPost_NilOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)
	if err := w.doPost(context.Background(), "/wxa/fake", nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_DoGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("access_token") != "ATOK" {
			t.Errorf("missing access_token")
		}
		if r.URL.Query().Get("foo") != "bar" {
			t.Errorf("missing foo=bar, got %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"name":"zzz"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	var out struct {
		Name string `json:"name"`
	}
	if err := w.doGet(context.Background(), "/wxa/fake_get", url.Values{"foo": {"bar"}}, &out); err != nil {
		t.Fatal(err)
	}
	if out.Name != "zzz" {
		t.Errorf("got %q", out.Name)
	}
}

func TestWxaAdmin_DoGetRaw_Binary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("access_token") != "ATOK" {
			t.Errorf("missing access_token")
		}
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0}) // JPEG magic
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	body, ct, err := w.doGetRaw(context.Background(), "/wxa/fake_binary", url.Values{"path": {"pages/index"}})
	if err != nil {
		t.Fatal(err)
	}
	if ct != "image/jpeg" {
		t.Errorf("content-type: %q", ct)
	}
	if len(body) != 4 || body[0] != 0xFF {
		t.Errorf("body mismatch: %v", body)
	}
}
```

- [ ] **Step 1.2: Run failing test**

Run: `go test ./oplatform/ -run TestWxaAdmin`
Expected: build error — `WxaAdminClient`, `WxaAdmin`, `doPost`, `doGet`, `doGetRaw` undefined.

- [ ] **Step 1.3: Create `oplatform/wxa.struct.go` header**

```go
package oplatform

// 本文件汇总代小程序开发管理 (WxaAdmin) 所有子族的请求/响应 DTO。
// 各子族（account/category/domain/tester/code/release）的结构体
// 按顺序追加到下面的分隔注释段。
```

- [ ] **Step 1.4: Create `oplatform/wxa.client.go`**

```go
package oplatform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// WxaAdminClient 代小程序开发管理客户端。
//
// 所有方法都以 AuthorizerClient.AccessToken() 作为 token 源，
// 通过共享的 doPost / doGet / doGetRaw 辅助统一处理 errcode 折叠和
// 两次 JSON 解码（先提取 errcode，再反序列化到 typed struct）。
//
// WxaAdminClient 本身无状态，线程安全；可在多 goroutine 共享。
type WxaAdminClient struct {
	auth *AuthorizerClient
}

// WxaAdmin 从 AuthorizerClient 构造开发管理客户端。
// 构造不做 I/O。
func (a *AuthorizerClient) WxaAdmin() *WxaAdminClient {
	return &WxaAdminClient{auth: a}
}

// doPost 通用 POST 辅助：
//   - 从 AuthorizerClient 取 access_token，拼到 query
//   - 先把响应解析到 json.RawMessage
//   - 从 raw 里提取 errcode/errmsg；!= 0 时返回 *WeixinError
//   - 若 caller 提供 out != nil，再把 raw 反序列化到 out
//
// 两次解码的原因：少数微信接口（例如 setnickname）会把业务字段和
// errcode 混在同一层 JSON；先提取 errcode 可避免字段冲突。
func (w *WxaAdminClient) doPost(ctx context.Context, path string, body, out any) error {
	ctx = touchContext(ctx)
	token, err := w.auth.AccessToken(ctx)
	if err != nil {
		return err
	}
	fullPath := path + "?access_token=" + url.QueryEscape(token)

	var raw json.RawMessage
	if err := w.auth.c.http.Post(ctx, fullPath, body, &raw); err != nil {
		return fmt.Errorf("oplatform: %s: %w", path, err)
	}
	return decodeRaw(path, raw, out)
}

// doGet 通用 GET 辅助：access_token 合并到 caller 传入的 query values。
func (w *WxaAdminClient) doGet(ctx context.Context, path string, q url.Values, out any) error {
	ctx = touchContext(ctx)
	token, err := w.auth.AccessToken(ctx)
	if err != nil {
		return err
	}
	if q == nil {
		q = url.Values{}
	}
	q.Set("access_token", token)

	var raw json.RawMessage
	if err := w.auth.c.http.Get(ctx, path, q, &raw); err != nil {
		return fmt.Errorf("oplatform: %s: %w", path, err)
	}
	return decodeRaw(path, raw, out)
}

// doGetRaw 用于响应体是二进制而非 JSON 的接口（目前只有 GetQrcode）。
// 返回原始字节 + Content-Type。
func (w *WxaAdminClient) doGetRaw(ctx context.Context, path string, q url.Values) ([]byte, string, error) {
	ctx = touchContext(ctx)
	token, err := w.auth.AccessToken(ctx)
	if err != nil {
		return nil, "", err
	}
	if q == nil {
		q = url.Values{}
	}
	q.Set("access_token", token)

	_, header, body, err := w.auth.c.http.DoRequestWithRawResponse(
		ctx, http.MethodGet, path, q, nil, nil,
	)
	if err != nil {
		return nil, "", fmt.Errorf("oplatform: %s: %w", path, err)
	}
	// 微信在出错时仍可能返回 JSON（errcode!=0），检测一下
	if len(body) > 0 && body[0] == '{' {
		var base struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		if json.Unmarshal(body, &base) == nil && base.ErrCode != 0 {
			return nil, "", &WeixinError{ErrCode: base.ErrCode, ErrMsg: base.ErrMsg}
		}
	}
	return body, header.Get("Content-Type"), nil
}

// decodeRaw 折叠 errcode 检查 + typed out 反序列化。
func decodeRaw(path string, raw json.RawMessage, out any) error {
	var base struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	_ = json.Unmarshal(raw, &base)
	if err := checkWeixinErr(base.ErrCode, base.ErrMsg); err != nil {
		return err
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("oplatform: %s decode: %w", path, err)
		}
	}
	return nil
}
```

- [ ] **Step 1.5: Run tests**

Run: `go test ./oplatform/ -run TestWxaAdmin`
Expected: 5 tests PASS.

Run: `go test ./oplatform/...`
Expected: all existing oplatform tests still pass (no regressions).

Run: `go build ./...`
Expected: clean.

- [ ] **Step 1.6: Commit**

```bash
git add oplatform/wxa.client.go oplatform/wxa.client_test.go oplatform/wxa.struct.go
git commit -m "feat(oplatform): add WxaAdminClient wrapper + shared helpers

Introduces WxaAdminClient returned by AuthorizerClient.WxaAdmin().
Three shared helpers (doPost / doGet / doGetRaw) handle token-as-query,
two-pass JSON decoding (errcode extraction then typed unmarshal), and
binary response fetching. wxa.struct.go is an empty header file that
subsequent sub-family tasks append DTOs to.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: `wxa.account.go` — account management (5 methods)

**Files:**
- Create: `oplatform/wxa.account.go`
- Create: `oplatform/wxa.account_test.go`
- Modify: `oplatform/wxa.struct.go` (append account DTOs)

- [ ] **Step 2.1: Append account DTOs to `oplatform/wxa.struct.go`**

Append these types to the file:

```go

// ----- account -----

type WxaSetNicknameReq struct {
	Nickname     string `json:"nick_name"`
	IDCard       string `json:"id_card,omitempty"`       // 身份证照片 mediaid
	License      string `json:"license,omitempty"`       // 组织机构代码证 mediaid
	NamingOther1 string `json:"naming_other_stuff_1,omitempty"`
	NamingOther2 string `json:"naming_other_stuff_2,omitempty"`
	NamingOther3 string `json:"naming_other_stuff_3,omitempty"`
	NamingOther4 string `json:"naming_other_stuff_4,omitempty"`
	NamingOther5 string `json:"naming_other_stuff_5,omitempty"`
}

type WxaSetNicknameResp struct {
	Wording string `json:"wording,omitempty"`
	AuditID string `json:"audit_id,omitempty"`
}

type WxaQueryNicknameResp struct {
	Nickname  string `json:"nickname"`
	AuditStat int    `json:"audit_stat"`
	FailReason string `json:"fail_reason,omitempty"`
	CreateTime int64  `json:"create_time"`
	AuditTime  int64  `json:"audit_time"`
}

type WxaCheckNicknameResp struct {
	HitCondition bool   `json:"hit_condition"`
	Wording      string `json:"wording,omitempty"`
}
```

- [ ] **Step 2.2: Write failing test `oplatform/wxa.account_test.go`**

```go
package oplatform

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWxaAdmin_SetNickname(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/setnickname") {
			t.Errorf("path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","wording":"","audit_id":"12345"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.SetNickname(context.Background(), &WxaSetNicknameReq{Nickname: "cool app"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.AuditID != "12345" {
		t.Errorf("audit_id: %q", resp.AuditID)
	}
	if gotBody["nick_name"] != "cool app" {
		t.Errorf("body: %+v", gotBody)
	}
}

func TestWxaAdmin_QueryNickname(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/api_wxa_querynickname") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"nickname":"cool","audit_stat":3,"create_time":1700,"audit_time":1800}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.QueryNickname(context.Background(), "12345")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Nickname != "cool" || resp.AuditStat != 3 {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_CheckNickname(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxverify/checkwxverifynickname") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"hit_condition":false,"wording":""}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.CheckNickname(context.Background(), "cool")
	if err != nil {
		t.Fatal(err)
	}
	if resp.HitCondition {
		t.Errorf("expected false")
	}
}

func TestWxaAdmin_ModifyHeadImage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/account/modifyheadimage") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.ModifyHeadImage(context.Background(), "MEDIA_ID"); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_ModifySignature(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/account/modifysignature") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.ModifySignature(context.Background(), "awesome sig"); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2.3: Run failing test**

Run: `go test ./oplatform/ -run TestWxaAdmin_SetNickname`
Expected: undefined `SetNickname`, `QueryNickname`, `CheckNickname`, `ModifyHeadImage`, `ModifySignature`.

- [ ] **Step 2.4: Create `oplatform/wxa.account.go`**

```go
package oplatform

import "context"

// SetNickname 设置小程序名称。
// /wxa/setnickname
func (w *WxaAdminClient) SetNickname(ctx context.Context, req *WxaSetNicknameReq) (*WxaSetNicknameResp, error) {
	var resp WxaSetNicknameResp
	if err := w.doPost(ctx, "/wxa/setnickname", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryNickname 查询改名审核状态。
// /wxa/api_wxa_querynickname
func (w *WxaAdminClient) QueryNickname(ctx context.Context, auditID string) (*WxaQueryNicknameResp, error) {
	body := map[string]string{"audit_id": auditID}
	var resp WxaQueryNicknameResp
	if err := w.doPost(ctx, "/wxa/api_wxa_querynickname", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CheckNickname 名称合法性预检。
// /cgi-bin/wxverify/checkwxverifynickname
func (w *WxaAdminClient) CheckNickname(ctx context.Context, nickname string) (*WxaCheckNicknameResp, error) {
	body := map[string]string{"nick_name": nickname}
	var resp WxaCheckNicknameResp
	if err := w.doPost(ctx, "/cgi-bin/wxverify/checkwxverifynickname", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ModifyHeadImage 修改头像。头像区域固定为整张图 (0,0)-(1,1)。
// /cgi-bin/account/modifyheadimage
func (w *WxaAdminClient) ModifyHeadImage(ctx context.Context, mediaID string) error {
	body := map[string]any{
		"head_img_media_id": mediaID,
		"x1":                0.0,
		"y1":                0.0,
		"x2":                1.0,
		"y2":                1.0,
	}
	return w.doPost(ctx, "/cgi-bin/account/modifyheadimage", body, nil)
}

// ModifySignature 修改功能介绍。
// /cgi-bin/account/modifysignature
func (w *WxaAdminClient) ModifySignature(ctx context.Context, signature string) error {
	body := map[string]string{"signature": signature}
	return w.doPost(ctx, "/cgi-bin/account/modifysignature", body, nil)
}
```

- [ ] **Step 2.5: Run tests + commit**

```bash
go test ./oplatform/ -run TestWxaAdmin_ -v
# Expected: 5 new + 5 client tests PASS

go test ./oplatform/...
# Expected: all PASS

go build ./...
# Expected: clean

git add oplatform/wxa.account.go oplatform/wxa.account_test.go oplatform/wxa.struct.go
git commit -m "feat(oplatform): add wxa account management (SetNickname et al.)

Implements 5 account methods on WxaAdminClient: SetNickname,
QueryNickname, CheckNickname, ModifyHeadImage, ModifySignature.
All use the shared doPost helper; DTOs appended to wxa.struct.go.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: `wxa.category.go` — category management (5 methods)

**Files:**
- Create: `oplatform/wxa.category.go`
- Create: `oplatform/wxa.category_test.go`
- Modify: `oplatform/wxa.struct.go` (append category DTOs)

- [ ] **Step 3.1: Append category DTOs to `oplatform/wxa.struct.go`**

```go

// ----- category -----

type WxaCategoryItem struct {
	First        int    `json:"first"`
	Second       int    `json:"second"`
	FirstName    string `json:"first_name,omitempty"`
	SecondName   string `json:"second_name,omitempty"`
	AuditStatus  int    `json:"audit_status,omitempty"`
	AuditReason  string `json:"audit_reason,omitempty"`
}

type WxaGetCategoryResp struct {
	CategoriesList []WxaCategoryItem `json:"categories_list"`
}

type WxaGetAllCategoriesResp struct {
	CategoriesList []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Level    int    `json:"level"`
		Father   int    `json:"father"`
		Children []int  `json:"children,omitempty"`
	} `json:"categories_list"`
}

type WxaAddCategoryReq struct {
	Categories []WxaCategoryItem `json:"categories"`
}

type WxaModifyCategoryReq struct {
	First       int    `json:"first"`
	Second      int    `json:"second"`
	Certicates  []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"certicates,omitempty"`
}
```

- [ ] **Step 3.2: Write failing test `oplatform/wxa.category_test.go`**

```go
package oplatform

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWxaAdmin_GetCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/getcategory") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"categories_list":[{"first":1,"second":2,"first_name":"工具","second_name":"办公"}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetCategory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.CategoriesList) != 1 || resp.CategoriesList[0].First != 1 {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_GetAllCategories(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/getallcategories") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"categories_list":[{"id":1,"name":"root","level":1,"father":0,"children":[2,3]}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetAllCategories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.CategoriesList) != 1 || resp.CategoriesList[0].Name != "root" {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_AddCategory(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/addcategory") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.AddCategory(context.Background(), &WxaAddCategoryReq{
		Categories: []WxaCategoryItem{{First: 1, Second: 2}},
	})
	if err != nil {
		t.Fatal(err)
	}
	cats, _ := body["categories"].([]any)
	if len(cats) != 1 {
		t.Errorf("body categories: %+v", body)
	}
}

func TestWxaAdmin_DeleteCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/deletecategory") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.DeleteCategory(context.Background(), 1, 2); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_ModifyCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/modifycategory") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.ModifyCategory(context.Background(), &WxaModifyCategoryReq{First: 1, Second: 2})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_ModifyCategory_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":85003,"errmsg":"too frequent"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.ModifyCategory(context.Background(), &WxaModifyCategoryReq{First: 1, Second: 2})
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 85003 {
		t.Errorf("expected WeixinError 85003, got %v", err)
	}
}
```

- [ ] **Step 3.3: Run failing test**

Run: `go test ./oplatform/ -run TestWxaAdmin_GetCategory`
Expected: undefined symbols.

- [ ] **Step 3.4: Create `oplatform/wxa.category.go`**

```go
package oplatform

import "context"

// GetCategory 获取当前账号的类目。
// /cgi-bin/wxopen/getcategory
func (w *WxaAdminClient) GetCategory(ctx context.Context) (*WxaGetCategoryResp, error) {
	var resp WxaGetCategoryResp
	if err := w.doPost(ctx, "/cgi-bin/wxopen/getcategory", struct{}{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAllCategories 获取所有类目。
// /cgi-bin/wxopen/getallcategories
func (w *WxaAdminClient) GetAllCategories(ctx context.Context) (*WxaGetAllCategoriesResp, error) {
	var resp WxaGetAllCategoriesResp
	if err := w.doPost(ctx, "/cgi-bin/wxopen/getallcategories", struct{}{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddCategory 添加类目。
// /cgi-bin/wxopen/addcategory
func (w *WxaAdminClient) AddCategory(ctx context.Context, req *WxaAddCategoryReq) error {
	return w.doPost(ctx, "/cgi-bin/wxopen/addcategory", req, nil)
}

// DeleteCategory 删除类目。
// /cgi-bin/wxopen/deletecategory
func (w *WxaAdminClient) DeleteCategory(ctx context.Context, first, second int) error {
	body := map[string]int{"first": first, "second": second}
	return w.doPost(ctx, "/cgi-bin/wxopen/deletecategory", body, nil)
}

// ModifyCategory 修改类目资质。
// /cgi-bin/wxopen/modifycategory
func (w *WxaAdminClient) ModifyCategory(ctx context.Context, req *WxaModifyCategoryReq) error {
	return w.doPost(ctx, "/cgi-bin/wxopen/modifycategory", req, nil)
}
```

- [ ] **Step 3.5: Run tests + commit**

```bash
go test ./oplatform/ -run TestWxaAdmin
# Expected: all PASS

go build ./...

git add oplatform/wxa.category.go oplatform/wxa.category_test.go oplatform/wxa.struct.go
git commit -m "feat(oplatform): add wxa category management (5 methods)

GetCategory / GetAllCategories / AddCategory / DeleteCategory /
ModifyCategory on WxaAdminClient, all via doPost.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: `wxa.domain.go` — domain management (4 methods)

**Files:**
- Create: `oplatform/wxa.domain.go`
- Create: `oplatform/wxa.domain_test.go`
- Modify: `oplatform/wxa.struct.go`

- [ ] **Step 4.1: Append domain DTOs to `oplatform/wxa.struct.go`**

```go

// ----- domain -----

type WxaModifyServerDomainReq struct {
	Action           string   `json:"action"` // add/delete/set/get/delete_legal_domain
	Requestdomain    []string `json:"requestdomain,omitempty"`
	Wsrequestdomain  []string `json:"wsrequestdomain,omitempty"`
	Uploaddomain     []string `json:"uploaddomain,omitempty"`
	Downloaddomain   []string `json:"downloaddomain,omitempty"`
	Udpdomain        []string `json:"udpdomain,omitempty"`
	Tcpdomain        []string `json:"tcpdomain,omitempty"`
}

type WxaServerDomainResp struct {
	Requestdomain   []string `json:"requestdomain,omitempty"`
	Wsrequestdomain []string `json:"wsrequestdomain,omitempty"`
	Uploaddomain    []string `json:"uploaddomain,omitempty"`
	Downloaddomain  []string `json:"downloaddomain,omitempty"`
	Udpdomain       []string `json:"udpdomain,omitempty"`
	Tcpdomain       []string `json:"tcpdomain,omitempty"`
	InvalidRequestdomain   []string `json:"invalid_requestdomain,omitempty"`
	InvalidWsrequestdomain []string `json:"invalid_wsrequestdomain,omitempty"`
	InvalidUploaddomain    []string `json:"invalid_uploaddomain,omitempty"`
	InvalidDownloaddomain  []string `json:"invalid_downloaddomain,omitempty"`
}

type WxaSetWebviewDomainReq struct {
	Action        string   `json:"action"` // add/delete/set/get
	Webviewdomain []string `json:"webviewdomain,omitempty"`
}

type WxaDomainConfirmFile struct {
	FileName    string `json:"file_name"`
	FileContent string `json:"file_content"`
}

type WxaModifyDomainDirectlyReq struct {
	Action          string   `json:"action"`
	Requestdomain   []string `json:"requestdomain,omitempty"`
	Wsrequestdomain []string `json:"wsrequestdomain,omitempty"`
	Uploaddomain    []string `json:"uploaddomain,omitempty"`
	Downloaddomain  []string `json:"downloaddomain,omitempty"`
}
```

- [ ] **Step 4.2: Write failing test `oplatform/wxa.domain_test.go`**

```go
package oplatform

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWxaAdmin_ModifyServerDomain(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/modify_domain") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"requestdomain":["https://a.example.com"]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.ModifyServerDomain(context.Background(), &WxaModifyServerDomainReq{
		Action:        "add",
		Requestdomain: []string{"https://a.example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["action"] != "add" {
		t.Errorf("body action: %+v", body)
	}
	if len(resp.Requestdomain) != 1 {
		t.Errorf("resp: %+v", resp)
	}
}

func TestWxaAdmin_SetWebviewDomain(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/setwebviewdomain") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.SetWebviewDomain(context.Background(), &WxaSetWebviewDomainReq{
		Action:        "set",
		Webviewdomain: []string{"https://webview.example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_GetDomainConfirmFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/get_webview_confirmfile") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"file_name":"abc.txt","file_content":"xyz"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	file, err := w.GetDomainConfirmFile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if file.FileName != "abc.txt" || file.FileContent != "xyz" {
		t.Errorf("unexpected: %+v", file)
	}
}

func TestWxaAdmin_ModifyDomainDirectly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/modify_domain_directly") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"requestdomain":["https://a.example.com"]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.ModifyDomainDirectly(context.Background(), &WxaModifyDomainDirectlyReq{
		Action:        "set",
		Requestdomain: []string{"https://a.example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Requestdomain) != 1 {
		t.Errorf("resp: %+v", resp)
	}
}
```

- [ ] **Step 4.3: Run failing test, then create `oplatform/wxa.domain.go`**

```go
package oplatform

import "context"

// ModifyServerDomain 设置/增加/删除服务器域名。
// /wxa/modify_domain
func (w *WxaAdminClient) ModifyServerDomain(ctx context.Context, req *WxaModifyServerDomainReq) (*WxaServerDomainResp, error) {
	var resp WxaServerDomainResp
	if err := w.doPost(ctx, "/wxa/modify_domain", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetWebviewDomain 设置业务域名。
// /wxa/setwebviewdomain
func (w *WxaAdminClient) SetWebviewDomain(ctx context.Context, req *WxaSetWebviewDomainReq) error {
	return w.doPost(ctx, "/wxa/setwebviewdomain", req, nil)
}

// GetDomainConfirmFile 获取业务域名校验文件。
// /wxa/get_webview_confirmfile
func (w *WxaAdminClient) GetDomainConfirmFile(ctx context.Context) (*WxaDomainConfirmFile, error) {
	var resp WxaDomainConfirmFile
	if err := w.doPost(ctx, "/wxa/get_webview_confirmfile", struct{}{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ModifyDomainDirectly 快速配置小程序服务器域名。
// /wxa/modify_domain_directly
func (w *WxaAdminClient) ModifyDomainDirectly(ctx context.Context, req *WxaModifyDomainDirectlyReq) (*WxaServerDomainResp, error) {
	var resp WxaServerDomainResp
	if err := w.doPost(ctx, "/wxa/modify_domain_directly", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 4.4: Run tests + commit**

```bash
go test ./oplatform/ -run TestWxaAdmin
go build ./...

git add oplatform/wxa.domain.go oplatform/wxa.domain_test.go oplatform/wxa.struct.go
git commit -m "feat(oplatform): add wxa domain management (4 methods)

ModifyServerDomain / SetWebviewDomain / GetDomainConfirmFile /
ModifyDomainDirectly on WxaAdminClient.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: `wxa.tester.go` — tester management (3 methods)

**Files:**
- Create: `oplatform/wxa.tester.go`
- Create: `oplatform/wxa.tester_test.go`
- Modify: `oplatform/wxa.struct.go`

- [ ] **Step 5.1: Append tester DTOs**

```go

// ----- tester -----

type WxaBindTesterResp struct {
	UserStr string `json:"userstr"`
}

type WxaListTestersResp struct {
	Members []struct {
		UserStr string `json:"userstr"`
	} `json:"members"`
}
```

- [ ] **Step 5.2: Write failing test `oplatform/wxa.tester_test.go`**

```go
package oplatform

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWxaAdmin_BindTester(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/bind_tester") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"userstr":"USER_STR_1"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.BindTester(context.Background(), "cool_wechat")
	if err != nil {
		t.Fatal(err)
	}
	if resp.UserStr != "USER_STR_1" {
		t.Errorf("userstr: %q", resp.UserStr)
	}
	if body["wechatid"] != "cool_wechat" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_UnbindTester_ByWechatID(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/unbind_tester") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.UnbindTester(context.Background(), "cool_wechat", ""); err != nil {
		t.Fatal(err)
	}
	if body["wechatid"] != "cool_wechat" {
		t.Errorf("body: %+v", body)
	}
	if _, ok := body["userstr"]; ok {
		t.Errorf("userstr should be omitted when empty")
	}
}

func TestWxaAdmin_UnbindTester_ByUserStr(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.UnbindTester(context.Background(), "", "USER_STR_1"); err != nil {
		t.Fatal(err)
	}
	if body["userstr"] != "USER_STR_1" {
		t.Errorf("body: %+v", body)
	}
	if _, ok := body["wechatid"]; ok {
		t.Errorf("wechatid should be omitted when empty")
	}
}

func TestWxaAdmin_ListTesters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/memberauth") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"members":[{"userstr":"U1"},{"userstr":"U2"}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.ListTesters(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Members) != 2 || resp.Members[0].UserStr != "U1" {
		t.Errorf("unexpected: %+v", resp)
	}
}
```

- [ ] **Step 5.3: Create `oplatform/wxa.tester.go`**

```go
package oplatform

import "context"

// BindTester 绑定体验者。
// /wxa/bind_tester
func (w *WxaAdminClient) BindTester(ctx context.Context, wechatID string) (*WxaBindTesterResp, error) {
	body := map[string]string{"wechatid": wechatID}
	var resp WxaBindTesterResp
	if err := w.doPost(ctx, "/wxa/bind_tester", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnbindTester 解绑体验者。
// 提供 wechatID 或 userStr 任意一个（二选一）；两者都空时微信会报错。
// /wxa/unbind_tester
func (w *WxaAdminClient) UnbindTester(ctx context.Context, wechatID, userStr string) error {
	body := map[string]string{}
	if wechatID != "" {
		body["wechatid"] = wechatID
	}
	if userStr != "" {
		body["userstr"] = userStr
	}
	return w.doPost(ctx, "/wxa/unbind_tester", body, nil)
}

// ListTesters 获取体验者列表。
// /wxa/memberauth
func (w *WxaAdminClient) ListTesters(ctx context.Context) (*WxaListTestersResp, error) {
	body := map[string]string{"action": "get_experiencer"}
	var resp WxaListTestersResp
	if err := w.doPost(ctx, "/wxa/memberauth", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 5.4: Run tests + commit**

```bash
go test ./oplatform/ -run TestWxaAdmin
go build ./...

git add oplatform/wxa.tester.go oplatform/wxa.tester_test.go oplatform/wxa.struct.go
git commit -m "feat(oplatform): add wxa tester management (3 methods)

BindTester / UnbindTester / ListTesters on WxaAdminClient.
UnbindTester accepts either wechatID or userStr and omits empty
fields from the request body.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: `wxa.code.go` — code management (4 methods, includes binary)

**Files:**
- Create: `oplatform/wxa.code.go`
- Create: `oplatform/wxa.code_test.go`
- Modify: `oplatform/wxa.struct.go`

- [ ] **Step 6.1: Append code DTOs**

```go

// ----- code -----

type WxaCommitReq struct {
	TemplateID  int    `json:"template_id"`
	UserVersion string `json:"user_version"`
	UserDesc    string `json:"user_desc"`
	ExtJSON     string `json:"ext_json"`
}

type WxaGetPageResp struct {
	PageList []string `json:"page_list"`
}

type WxaGetCodeCategoryResp struct {
	CategoryList []struct {
		FirstClass  string `json:"first_class"`
		SecondClass string `json:"second_class"`
		ThirdClass  string `json:"third_class,omitempty"`
		FirstID     int    `json:"first_id"`
		SecondID    int    `json:"second_id"`
		ThirdID     int    `json:"third_id,omitempty"`
	} `json:"category_list"`
}
```

- [ ] **Step 6.2: Write failing test `oplatform/wxa.code_test.go`**

```go
package oplatform

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWxaAdmin_Commit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/commit") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.Commit(context.Background(), &WxaCommitReq{
		TemplateID:  1,
		UserVersion: "1.0.0",
		UserDesc:    "initial",
		ExtJSON:     `{"extEnable":true}`,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_GetPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/get_page") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"page_list":["pages/index/index","pages/detail/detail"]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetPage(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.PageList) != 2 {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_GetQrcode_Binary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/get_qrcode") {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("path") != "pages/index" {
			t.Errorf("path query: %q", r.URL.Query().Get("path"))
		}
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10})
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	body, ct, err := w.GetQrcode(context.Background(), "pages/index")
	if err != nil {
		t.Fatal(err)
	}
	if ct != "image/jpeg" {
		t.Errorf("content-type: %q", ct)
	}
	if len(body) != 6 || body[0] != 0xFF {
		t.Errorf("body: %v", body)
	}
}

func TestWxaAdmin_GetQrcode_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":85024,"errmsg":"no test version"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	_, _, err := w.GetQrcode(context.Background(), "pages/index")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWxaAdmin_GetCodeCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/get_category") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"category_list":[{"first_class":"A","second_class":"B","first_id":1,"second_id":2}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetCodeCategory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.CategoryList) != 1 || resp.CategoryList[0].FirstClass != "A" {
		t.Errorf("unexpected: %+v", resp)
	}
}
```

- [ ] **Step 6.3: Create `oplatform/wxa.code.go`**

```go
package oplatform

import (
	"context"
	"net/url"
)

// Commit 上传代码。
// /wxa/commit
func (w *WxaAdminClient) Commit(ctx context.Context, req *WxaCommitReq) error {
	return w.doPost(ctx, "/wxa/commit", req, nil)
}

// GetPage 获取已上传代码的页面列表。
// /wxa/get_page
func (w *WxaAdminClient) GetPage(ctx context.Context) (*WxaGetPageResp, error) {
	var resp WxaGetPageResp
	if err := w.doGet(ctx, "/wxa/get_page", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetQrcode 获取体验版二维码（返回二进制图片）。
// /wxa/get_qrcode
func (w *WxaAdminClient) GetQrcode(ctx context.Context, path string) ([]byte, string, error) {
	q := url.Values{}
	if path != "" {
		q.Set("path", path)
	}
	return w.doGetRaw(ctx, "/wxa/get_qrcode", q)
}

// GetCodeCategory 获取代码草稿可选类目。
// /wxa/get_category
func (w *WxaAdminClient) GetCodeCategory(ctx context.Context) (*WxaGetCodeCategoryResp, error) {
	var resp WxaGetCodeCategoryResp
	if err := w.doGet(ctx, "/wxa/get_category", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 6.4: Run tests + commit**

```bash
go test ./oplatform/ -run TestWxaAdmin
go build ./...

git add oplatform/wxa.code.go oplatform/wxa.code_test.go oplatform/wxa.struct.go
git commit -m "feat(oplatform): add wxa code management (4 methods, incl. binary GetQrcode)

Commit / GetPage / GetQrcode / GetCodeCategory on WxaAdminClient.
GetQrcode uses doGetRaw for binary response; doGetRaw detects JSON
errcode bodies to surface WeixinError on failure.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: `wxa.release.go` — audit & release (10 methods)

**Files:**
- Create: `oplatform/wxa.release.go`
- Create: `oplatform/wxa.release_test.go`
- Modify: `oplatform/wxa.struct.go`

- [ ] **Step 7.1: Append release DTOs**

```go

// ----- release -----

type WxaSubmitAuditReq struct {
	ItemList []struct {
		Address     string `json:"address"`
		Tag         string `json:"tag"`
		FirstClass  string `json:"first_class"`
		SecondClass string `json:"second_class"`
		ThirdClass  string `json:"third_class,omitempty"`
		FirstID     int    `json:"first_id"`
		SecondID    int    `json:"second_id"`
		ThirdID     int    `json:"third_id,omitempty"`
		Title       string `json:"title,omitempty"`
	} `json:"item_list,omitempty"`
	PreviewInfo *struct {
		VideoIDList []string `json:"video_id_list,omitempty"`
		PicIDList   []string `json:"pic_id_list,omitempty"`
	} `json:"preview_info,omitempty"`
	VersionDesc string `json:"version_desc,omitempty"`
	FeedbackInfo string `json:"feedback_info,omitempty"`
	FeedbackStuff string `json:"feedback_stuff,omitempty"`
}

type WxaSubmitAuditResp struct {
	AuditID int64 `json:"auditid"`
}

type WxaAuditStatus struct {
	AuditID    int64  `json:"auditid,omitempty"`
	Status     int    `json:"status"`
	Reason     string `json:"reason,omitempty"`
	ScreenShot string `json:"screenshot,omitempty"`
	UserVersion string `json:"user_version,omitempty"`
	UserDesc    string `json:"user_desc,omitempty"`
	SubmitAuditTime int64 `json:"submit_audit_time,omitempty"`
}

type WxaSupportVersionResp struct {
	NowVersion  string `json:"now_version"`
	UVInfo struct {
		Items []struct {
			Percentage float64 `json:"percentage"`
			Version    string  `json:"version"`
		} `json:"items"`
	} `json:"uv_info"`
}
```

- [ ] **Step 7.2: Write failing test `oplatform/wxa.release_test.go`**

```go
package oplatform

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWxaAdmin_SubmitAudit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/submit_audit") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"auditid":1234567890}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.SubmitAudit(context.Background(), &WxaSubmitAuditReq{
		VersionDesc: "bugfix",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.AuditID != 1234567890 {
		t.Errorf("audit_id: %d", resp.AuditID)
	}
}

func TestWxaAdmin_SubmitAudit_InProgress(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":85013,"errmsg":"invalid version"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	_, err := w.SubmitAudit(context.Background(), &WxaSubmitAuditReq{})
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 85013 {
		t.Errorf("expected 85013, got %v", err)
	}
}

func TestWxaAdmin_GetAuditStatus(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/get_auditstatus") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"status":2,"reason":"违反规则"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	status, err := w.GetAuditStatus(context.Background(), 9999)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != 2 || status.Reason != "违反规则" {
		t.Errorf("unexpected: %+v", status)
	}
	if int(body["auditid"].(float64)) != 9999 {
		t.Errorf("body auditid: %+v", body)
	}
}

func TestWxaAdmin_GetLatestAuditStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/get_latest_auditstatus") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"auditid":111,"status":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	status, err := w.GetLatestAuditStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.AuditID != 111 || status.Status != 0 {
		t.Errorf("unexpected: %+v", status)
	}
}

func TestWxaAdmin_UndoCodeAudit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/undocodeaudit") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.UndoCodeAudit(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_SpeedupAudit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/speedupaudit") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.SpeedupAudit(context.Background(), 1234); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_Release(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/release") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.Release(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_RevertCodeRelease(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/revertcoderelease") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.RevertCodeRelease(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_ChangeVisitStatus(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/change_visitstatus") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.ChangeVisitStatus(context.Background(), "close"); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "close" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_GetSupportVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/getweappsupportversion") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"now_version":"2.10.0","uv_info":{"items":[{"percentage":95.5,"version":"2.10.0"}]}}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetSupportVersion(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.NowVersion != "2.10.0" || len(resp.UVInfo.Items) != 1 {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_SetSupportVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/setweappsupportversion") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.SetSupportVersion(context.Background(), "2.10.0"); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 7.3: Create `oplatform/wxa.release.go`**

```go
package oplatform

import "context"

// SubmitAudit 提交审核。
// /wxa/submit_audit
func (w *WxaAdminClient) SubmitAudit(ctx context.Context, req *WxaSubmitAuditReq) (*WxaSubmitAuditResp, error) {
	var resp WxaSubmitAuditResp
	if err := w.doPost(ctx, "/wxa/submit_audit", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAuditStatus 查询指定版本审核状态。
// /wxa/get_auditstatus
func (w *WxaAdminClient) GetAuditStatus(ctx context.Context, auditID int64) (*WxaAuditStatus, error) {
	body := map[string]int64{"auditid": auditID}
	var resp WxaAuditStatus
	if err := w.doPost(ctx, "/wxa/get_auditstatus", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLatestAuditStatus 查询最新一次审核状态。
// /wxa/get_latest_auditstatus
func (w *WxaAdminClient) GetLatestAuditStatus(ctx context.Context) (*WxaAuditStatus, error) {
	var resp WxaAuditStatus
	if err := w.doGet(ctx, "/wxa/get_latest_auditstatus", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UndoCodeAudit 撤回代码审核。
// /wxa/undocodeaudit
func (w *WxaAdminClient) UndoCodeAudit(ctx context.Context) error {
	return w.doGet(ctx, "/wxa/undocodeaudit", nil, nil)
}

// SpeedupAudit 加急审核。
// /wxa/speedupaudit
func (w *WxaAdminClient) SpeedupAudit(ctx context.Context, auditID int64) error {
	body := map[string]int64{"auditid": auditID}
	return w.doPost(ctx, "/wxa/speedupaudit", body, nil)
}

// Release 发布已通过审核的版本。
// /wxa/release
func (w *WxaAdminClient) Release(ctx context.Context) error {
	return w.doPost(ctx, "/wxa/release", struct{}{}, nil)
}

// RevertCodeRelease 版本回退。
// /wxa/revertcoderelease
func (w *WxaAdminClient) RevertCodeRelease(ctx context.Context) error {
	return w.doGet(ctx, "/wxa/revertcoderelease", nil, nil)
}

// ChangeVisitStatus 修改可见状态。action = "open" | "close"
// /wxa/change_visitstatus
func (w *WxaAdminClient) ChangeVisitStatus(ctx context.Context, action string) error {
	body := map[string]string{"action": action}
	return w.doPost(ctx, "/wxa/change_visitstatus", body, nil)
}

// GetSupportVersion 查询小程序支持版本信息。
// /cgi-bin/wxopen/getweappsupportversion
func (w *WxaAdminClient) GetSupportVersion(ctx context.Context) (*WxaSupportVersionResp, error) {
	var resp WxaSupportVersionResp
	if err := w.doPost(ctx, "/cgi-bin/wxopen/getweappsupportversion", struct{}{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetSupportVersion 设置小程序最低支持版本。
// /cgi-bin/wxopen/setweappsupportversion
func (w *WxaAdminClient) SetSupportVersion(ctx context.Context, version string) error {
	body := map[string]string{"version": version}
	return w.doPost(ctx, "/cgi-bin/wxopen/setweappsupportversion", body, nil)
}
```

- [ ] **Step 7.4: Run tests + commit**

```bash
go test ./oplatform/ -run TestWxaAdmin -v
# Expected: all ~40 wxa tests pass

go test -race ./oplatform/...
# Expected: no races

go build ./...

git add oplatform/wxa.release.go oplatform/wxa.release_test.go oplatform/wxa.struct.go
git commit -m "feat(oplatform): add wxa audit & release (10 methods)

SubmitAudit / GetAuditStatus / GetLatestAuditStatus / UndoCodeAudit /
SpeedupAudit / Release / RevertCodeRelease / ChangeVisitStatus /
GetSupportVersion / SetSupportVersion — completes the 31-method
WxaAdmin surface.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 8: Final verification sweep

**Goal:** One last end-to-end check that everything is green. No code changes expected.

- [ ] **Step 8.1: Full build**

Run: `go build ./...`
Expected: clean.

- [ ] **Step 8.2: Full test suite with race detector**

Run: `go test -race ./...`
Expected: all PASS across all packages (`merchant/developed`, `mini-program`, `offiaccount`, `oplatform`, `utils`, `utils/wxcrypto`).

- [ ] **Step 8.3: Go vet**

Run: `go vet ./...`
Expected: no output.

- [ ] **Step 8.4: Count WxaAdmin methods**

Run:
```bash
grep -hE '^func \(w \*WxaAdminClient\)' oplatform/wxa.*.go | grep -v "doPost\|doGet\|doGetRaw" | wc -l
```
Expected: `31`.

- [ ] **Step 8.5: Count WxaAdmin tests**

Run:
```bash
grep -hE '^func TestWxaAdmin_' oplatform/wxa.*_test.go | wc -l
```
Expected: `>= 35`.

- [ ] **Step 8.6: Git log sanity check**

Run: `git log --oneline -10`
Expected: commits from Task 1 through Task 7 in order, 7 atomic commits plus the spec commit from the earlier phase.

No commit at this step — verification only.

---

## Coverage Map (self-review)

| Spec section | Task |
|---|---|
| §2.1 WxaAdminClient wrapper | Task 1 |
| §2.2 doPost/doGet/doGetRaw helpers | Task 1 |
| §2.3 file layout | All tasks |
| §3.1 account management (5 methods) | Task 2 |
| §3.2 category management (5 methods) | Task 3 |
| §3.3 domain management (4 methods) | Task 4 |
| §3.4 tester management (3 methods) | Task 5 |
| §3.5 code management (4 methods incl. binary) | Task 6 |
| §3.6 audit & release (10 methods) | Task 7 |
| §4 error handling (*WeixinError) | Task 1 (in helpers) |
| §5 concurrency / lifecycle | Task 1 (WxaAdminClient is stateless) |
| §6 testing strategy (httptest + newTestWxaAdmin) | Tasks 1-7 |
| §7 compatibility (zero break) | All tasks (additive only) |
| §8 delivery list | Tasks 1-7 |

**All spec requirements covered. No placeholders. Method name consistency verified:**

- `SetNickname`, `QueryNickname`, `CheckNickname`, `ModifyHeadImage`, `ModifySignature` (Task 2)
- `GetCategory`, `GetAllCategories`, `AddCategory`, `DeleteCategory`, `ModifyCategory` (Task 3)
- `ModifyServerDomain`, `SetWebviewDomain`, `GetDomainConfirmFile`, `ModifyDomainDirectly` (Task 4)
- `BindTester`, `UnbindTester`, `ListTesters` (Task 5)
- `Commit`, `GetPage`, `GetQrcode`, `GetCodeCategory` (Task 6)
- `SubmitAudit`, `GetAuditStatus`, `GetLatestAuditStatus`, `UndoCodeAudit`, `SpeedupAudit`, `Release`, `RevertCodeRelease`, `ChangeVisitStatus`, `GetSupportVersion`, `SetSupportVersion` (Task 7)

**Total: 5 + 5 + 4 + 3 + 4 + 10 = 31 methods ✓**

**Note on `GetCategory` vs `GetCodeCategory`:** intentionally two different methods at two different endpoints (`/cgi-bin/wxopen/getcategory` vs `/wxa/get_category`). Both defined; no collision.
