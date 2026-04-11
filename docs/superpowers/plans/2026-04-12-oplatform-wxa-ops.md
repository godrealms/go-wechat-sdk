# oplatform WxaOps Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add 16 代小程序运营 methods to existing `WxaAdminClient` across 3 sub-families (plugin / subscribe message / customer message).

**Architecture:** Pure additive. Reuse `WxaAdminClient.doPost` / `doGet` helpers from sub-project 3. One `.go` + one `*_test.go` per sub-family; DTOs appended to `wxa.struct.go`. No new types, no new test helpers.

**Tech Stack:** Go 1.23, stdlib only, existing `utils.HTTP`, `httptest`. Zero new deps.

**Spec:** `docs/superpowers/specs/2026-04-12-oplatform-wxa-ops-design.md`

**Module path:** `github.com/godrealms/go-wechat-sdk`

---

## Conventions

- Every task ends with a commit. Stage ONLY the 3 files the task touches.
- TDD: failing test first → verify → minimal implementation → verify pass → commit.
- Working tree has unrelated WIP — never `git add -A`. Verify with `git diff --cached --stat`.
- Paths are relative to `/Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk`.

### Shared context (already exists)

- `WxaAdminClient` with `doPost(ctx, path, body, out any) error` and `doGet(ctx, path, q url.Values, out any) error` helpers in `wxa.client.go`.
- `doPost` prepends path with `?access_token=...`, two-pass JSON decodes via package-level `decodeRaw`.
- `doGet` merges `access_token` into caller-provided query values.
- Test helper `newTestWxaAdmin(t *testing.T, baseURL string) *WxaAdminClient` in `wxa.client_test.go`. Pre-seeds Store with non-expired `AccessToken="ATOK"` so mocks don't need to handle authorizer token refresh.
- `WeixinError{ErrCode, ErrMsg}` in `errors.go`.
- `wxa.struct.go` is the shared DTO file for all WxaAdmin sub-families; append new sections at the bottom.

---

## Task 1: `wxa.plugin.go` — plugin management (7 methods)

**Files:**
- Create: `oplatform/wxa.plugin.go`
- Create: `oplatform/wxa.plugin_test.go`
- Modify: `oplatform/wxa.struct.go` (append plugin DTOs)

- [ ] **Step 1.1: Append plugin DTOs to `oplatform/wxa.struct.go`**

Append at the bottom of the file:

```go

// ----- plugin -----

type WxaPluginListItem struct {
	AppID      string `json:"appid"`
	Status     int    `json:"status"` // 1=申请中 2=申请通过 3=已拒绝 4=已超时
	Nickname   string `json:"nickname,omitempty"`
	HeadImgURL string `json:"headimgurl,omitempty"`
}

type WxaPluginList struct {
	PluginList []WxaPluginListItem `json:"plugin_list"`
}

type WxaPluginDevApplyItem struct {
	AppID      string `json:"appid"`
	Status     int    `json:"status"`
	Nickname   string `json:"nickname,omitempty"`
	HeadImgURL string `json:"headimgurl,omitempty"`
	Categories []struct {
		First  string `json:"first"`
		Second string `json:"second"`
	} `json:"categories,omitempty"`
	CreateTime string `json:"create_time,omitempty"`
	ApplyURL   string `json:"apply_url,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type WxaPluginDevApplyList struct {
	ApplyList []WxaPluginDevApplyItem `json:"apply_list"`
}
```

- [ ] **Step 1.2: Write failing tests `oplatform/wxa.plugin_test.go`**

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

func TestWxaAdmin_ApplyPlugin(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/plugin") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.ApplyPlugin(context.Background(), "wxPLUG"); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "apply" || body["plugin_appid"] != "wxPLUG" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_ApplyPlugin_Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":89236,"errmsg":"duplicate apply"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.ApplyPlugin(context.Background(), "wxPLUG")
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 89236 {
		t.Errorf("expected 89236, got %v", err)
	}
}

func TestWxaAdmin_ListPlugins(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"plugin_list":[{"appid":"wxA","status":2,"nickname":"N1"},{"appid":"wxB","status":1}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	list, err := w.ListPlugins(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if body["action"] != "list" {
		t.Errorf("body: %+v", body)
	}
	if len(list.PluginList) != 2 || list.PluginList[0].AppID != "wxA" {
		t.Errorf("list: %+v", list)
	}
}

func TestWxaAdmin_UnbindPlugin(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.UnbindPlugin(context.Background(), "wxPLUG"); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "unbind" || body["plugin_appid"] != "wxPLUG" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_GetPluginDevApplyList(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/devplugin") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"apply_list":[{"appid":"wxUSER","status":1,"nickname":"U1"}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	list, err := w.GetPluginDevApplyList(context.Background(), 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if body["action"] != "dev_apply_list" {
		t.Errorf("body: %+v", body)
	}
	if int(body["page"].(float64)) != 0 || int(body["num"].(float64)) != 10 {
		t.Errorf("page/num: %+v", body)
	}
	if len(list.ApplyList) != 1 || list.ApplyList[0].AppID != "wxUSER" {
		t.Errorf("list: %+v", list)
	}
}

func TestWxaAdmin_AgreeDevPlugin(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.AgreeDevPlugin(context.Background(), "wxUSER"); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "dev_agree" || body["appid"] != "wxUSER" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_RefuseDevPlugin(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.RefuseDevPlugin(context.Background(), "违反规范"); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "dev_refuse" || body["reason"] != "违反规范" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_DeleteDevPlugin(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.DeleteDevPlugin(context.Background(), "wxUSER"); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "dev_delete" || body["appid"] != "wxUSER" {
		t.Errorf("body: %+v", body)
	}
}
```

- [ ] **Step 1.3: Run failing tests — expect undefined methods.**

Run: `go test ./oplatform/ -run TestWxaAdmin_.*Plugin`
Expected: build errors.

- [ ] **Step 1.4: Create `oplatform/wxa.plugin.go`**

```go
package oplatform

import "context"

// ApplyPlugin 使用方申请使用插件。
// POST /wxa/plugin body: {"action":"apply","plugin_appid":"..."}
func (w *WxaAdminClient) ApplyPlugin(ctx context.Context, pluginAppID string) error {
	body := map[string]string{
		"action":       "apply",
		"plugin_appid": pluginAppID,
	}
	return w.doPost(ctx, "/wxa/plugin", body, nil)
}

// ListPlugins 使用方查询已添加的插件列表。
// POST /wxa/plugin body: {"action":"list"}
func (w *WxaAdminClient) ListPlugins(ctx context.Context) (*WxaPluginList, error) {
	body := map[string]string{"action": "list"}
	var resp WxaPluginList
	if err := w.doPost(ctx, "/wxa/plugin", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnbindPlugin 使用方解除插件。
// POST /wxa/plugin body: {"action":"unbind","plugin_appid":"..."}
func (w *WxaAdminClient) UnbindPlugin(ctx context.Context, pluginAppID string) error {
	body := map[string]string{
		"action":       "unbind",
		"plugin_appid": pluginAppID,
	}
	return w.doPost(ctx, "/wxa/plugin", body, nil)
}

// GetPluginDevApplyList 插件方：查询当前所有插件使用方申请列表。
// POST /wxa/devplugin body: {"action":"dev_apply_list","page":0,"num":10}
func (w *WxaAdminClient) GetPluginDevApplyList(ctx context.Context, page, num int) (*WxaPluginDevApplyList, error) {
	body := map[string]any{
		"action": "dev_apply_list",
		"page":   page,
		"num":    num,
	}
	var resp WxaPluginDevApplyList
	if err := w.doPost(ctx, "/wxa/devplugin", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AgreeDevPlugin 插件方：同意某个使用方的申请。
// POST /wxa/devplugin body: {"action":"dev_agree","appid":"..."}
func (w *WxaAdminClient) AgreeDevPlugin(ctx context.Context, userAppID string) error {
	body := map[string]string{
		"action": "dev_agree",
		"appid":  userAppID,
	}
	return w.doPost(ctx, "/wxa/devplugin", body, nil)
}

// RefuseDevPlugin 插件方：拒绝申请并给出原因。
// POST /wxa/devplugin body: {"action":"dev_refuse","reason":"..."}
func (w *WxaAdminClient) RefuseDevPlugin(ctx context.Context, reason string) error {
	body := map[string]string{
		"action": "dev_refuse",
		"reason": reason,
	}
	return w.doPost(ctx, "/wxa/devplugin", body, nil)
}

// DeleteDevPlugin 插件方：删除某个使用方的授权。
// POST /wxa/devplugin body: {"action":"dev_delete","appid":"..."}
func (w *WxaAdminClient) DeleteDevPlugin(ctx context.Context, userAppID string) error {
	body := map[string]string{
		"action": "dev_delete",
		"appid":  userAppID,
	}
	return w.doPost(ctx, "/wxa/devplugin", body, nil)
}
```

- [ ] **Step 1.5: Run tests + commit**

```bash
go test ./oplatform/ -run TestWxaAdmin_.*Plugin
go test ./oplatform/...
go build ./...

git add oplatform/wxa.plugin.go oplatform/wxa.plugin_test.go oplatform/wxa.struct.go
git diff --cached --stat   # verify exactly 3 files
git commit -m "feat(oplatform): add wxa plugin management (7 methods)

ApplyPlugin / ListPlugins / UnbindPlugin (user-side) and
GetPluginDevApplyList / AgreeDevPlugin / RefuseDevPlugin /
DeleteDevPlugin (plugin-developer-side). The action field lives
in the request body, not the URL query, so doPost needs no
modification — callers construct body maps directly.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: `wxa.submsg.go` — subscribe message & template library (7 methods)

**Files:**
- Create: `oplatform/wxa.submsg.go`
- Create: `oplatform/wxa.submsg_test.go`
- Modify: `oplatform/wxa.struct.go` (append submsg DTOs)

- [ ] **Step 2.1: Append submsg DTOs to `oplatform/wxa.struct.go`**

Append at the bottom:

```go

// ----- subscribe message -----

type WxaSubscribeCategoryItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type WxaSubscribeCategoryResp struct {
	Data []WxaSubscribeCategoryItem `json:"data"`
}

type WxaPubTemplateTitleItem struct {
	TID        int    `json:"tid"`
	Title      string `json:"title"`
	Type       int    `json:"type"` // 2=一次性 3=长期
	CategoryID string `json:"categoryId"`
}

type WxaPubTemplateTitles struct {
	Count int                       `json:"count"`
	Data  []WxaPubTemplateTitleItem `json:"data"`
}

type WxaPubTemplateKeywordItem struct {
	KID     int    `json:"kid"`
	Name    string `json:"name"`
	Example string `json:"example"`
	Rule    string `json:"rule"`
}

type WxaPubTemplateKeywords struct {
	Count int                         `json:"count"`
	Data  []WxaPubTemplateKeywordItem `json:"data"`
}

type WxaAddSubscribeTemplateReq struct {
	TID       string `json:"tid"`
	KidList   []int  `json:"kidList"`
	SceneDesc string `json:"sceneDesc,omitempty"`
}

type WxaAddSubscribeTemplateResp struct {
	PriTmplID string `json:"priTmplId"`
}

type WxaSubscribeTemplateItem struct {
	PriTmplID string `json:"priTmplId"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Example   string `json:"example"`
	Type      int    `json:"type"`
}

type WxaSubscribeTemplateList struct {
	Data []WxaSubscribeTemplateItem `json:"data"`
}

type WxaSubscribeTemplateDataField struct {
	Value string `json:"value"`
}

type WxaSendSubscribeReq struct {
	ToUser           string                                   `json:"touser"`
	TemplateID       string                                   `json:"template_id"`
	Page             string                                   `json:"page,omitempty"`
	MiniprogramState string                                   `json:"miniprogram_state,omitempty"`
	Lang             string                                   `json:"lang,omitempty"`
	Data             map[string]WxaSubscribeTemplateDataField `json:"data"`
}
```

- [ ] **Step 2.2: Write failing tests `oplatform/wxa.submsg_test.go`**

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

func TestWxaAdmin_GetSubscribeCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxaapi/newtmpl/getcategory") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"data":[{"id":1,"name":"工具"},{"id":2,"name":"教育"}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetSubscribeCategory(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 2 || resp.Data[0].Name != "工具" {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_GetPubTemplateTitles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxaapi/newtmpl/getpubtemplatetitles") {
			t.Errorf("path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("ids") != "2-3" {
			t.Errorf("ids: %q", q.Get("ids"))
		}
		if q.Get("start") != "0" || q.Get("limit") != "30" {
			t.Errorf("pagination: %v", q)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"count":1,"data":[{"tid":99,"title":"订单已发货","type":2,"categoryId":"2"}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetPubTemplateTitles(context.Background(), "2-3", 0, 30)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Count != 1 || len(resp.Data) != 1 || resp.Data[0].TID != 99 {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_GetPubTemplateKeywords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxaapi/newtmpl/getpubtemplatekeywords") {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("tid") != "99" {
			t.Errorf("tid: %q", r.URL.Query().Get("tid"))
		}
		_, _ = w.Write([]byte(`{"errcode":0,"count":2,"data":[{"kid":1,"name":"订单号","example":"1234","rule":"thing"},{"kid":2,"name":"金额","example":"10","rule":"amount"}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetPubTemplateKeywords(context.Background(), "99")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Count != 2 || len(resp.Data) != 2 || resp.Data[0].Name != "订单号" {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_AddSubscribeTemplate(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxaapi/newtmpl/addtemplate") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"priTmplId":"PTID_1"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.AddSubscribeTemplate(context.Background(), &WxaAddSubscribeTemplateReq{
		TID:       "99",
		KidList:   []int{1, 2},
		SceneDesc: "订单通知",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.PriTmplID != "PTID_1" {
		t.Errorf("priTmplId: %q", resp.PriTmplID)
	}
	if body["tid"] != "99" {
		t.Errorf("body tid: %+v", body)
	}
	kids, _ := body["kidList"].([]any)
	if len(kids) != 2 {
		t.Errorf("body kidList: %+v", body)
	}
}

func TestWxaAdmin_DeleteSubscribeTemplate(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxaapi/newtmpl/deltemplate") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.DeleteSubscribeTemplate(context.Background(), "PTID_1"); err != nil {
		t.Fatal(err)
	}
	if body["priTmplId"] != "PTID_1" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_ListSubscribeTemplates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxaapi/newtmpl/gettemplate") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"data":[{"priTmplId":"P1","title":"订单通知","content":"{{c1.DATA}}","example":"xxx","type":2}]}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	list, err := w.ListSubscribeTemplates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Data) != 1 || list.Data[0].PriTmplID != "P1" {
		t.Errorf("unexpected: %+v", list)
	}
}

func TestWxaAdmin_SendSubscribeMessage(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/message/subscribe/send") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.SendSubscribeMessage(context.Background(), &WxaSendSubscribeReq{
		ToUser:     "OPENID_1",
		TemplateID: "P1",
		Page:       "pages/order/detail",
		Data: map[string]WxaSubscribeTemplateDataField{
			"c1": {Value: "12345"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["touser"] != "OPENID_1" || body["template_id"] != "P1" {
		t.Errorf("body: %+v", body)
	}
	data, _ := body["data"].(map[string]any)
	c1, _ := data["c1"].(map[string]any)
	if c1["value"] != "12345" {
		t.Errorf("body.data: %+v", body)
	}
}
```

- [ ] **Step 2.3: Run failing test — expect undefined methods.**

- [ ] **Step 2.4: Create `oplatform/wxa.submsg.go`**

```go
package oplatform

import (
	"context"
	"fmt"
	"net/url"
)

// GetSubscribeCategory 获取小程序账号所属类目。
// GET /wxaapi/newtmpl/getcategory
func (w *WxaAdminClient) GetSubscribeCategory(ctx context.Context) (*WxaSubscribeCategoryResp, error) {
	var resp WxaSubscribeCategoryResp
	if err := w.doGet(ctx, "/wxaapi/newtmpl/getcategory", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPubTemplateTitles 获取模板库标题列表。
// ids 是用 "-" 连接的类目 ID 列表，例如 "2-3-5"。
// GET /wxaapi/newtmpl/getpubtemplatetitles
func (w *WxaAdminClient) GetPubTemplateTitles(ctx context.Context, ids string, start, limit int) (*WxaPubTemplateTitles, error) {
	q := url.Values{
		"ids":   {ids},
		"start": {fmt.Sprintf("%d", start)},
		"limit": {fmt.Sprintf("%d", limit)},
	}
	var resp WxaPubTemplateTitles
	if err := w.doGet(ctx, "/wxaapi/newtmpl/getpubtemplatetitles", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPubTemplateKeywords 获取模板库标题下的关键词列表。
// GET /wxaapi/newtmpl/getpubtemplatekeywords
func (w *WxaAdminClient) GetPubTemplateKeywords(ctx context.Context, tid string) (*WxaPubTemplateKeywords, error) {
	q := url.Values{"tid": {tid}}
	var resp WxaPubTemplateKeywords
	if err := w.doGet(ctx, "/wxaapi/newtmpl/getpubtemplatekeywords", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddSubscribeTemplate 组合关键词添加私有模板。
// POST /wxaapi/newtmpl/addtemplate
func (w *WxaAdminClient) AddSubscribeTemplate(ctx context.Context, req *WxaAddSubscribeTemplateReq) (*WxaAddSubscribeTemplateResp, error) {
	var resp WxaAddSubscribeTemplateResp
	if err := w.doPost(ctx, "/wxaapi/newtmpl/addtemplate", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteSubscribeTemplate 删除私有模板。
// POST /wxaapi/newtmpl/deltemplate
func (w *WxaAdminClient) DeleteSubscribeTemplate(ctx context.Context, priTmplID string) error {
	body := map[string]string{"priTmplId": priTmplID}
	return w.doPost(ctx, "/wxaapi/newtmpl/deltemplate", body, nil)
}

// ListSubscribeTemplates 获取账号下已添加的私有模板列表。
// GET /wxaapi/newtmpl/gettemplate
func (w *WxaAdminClient) ListSubscribeTemplates(ctx context.Context) (*WxaSubscribeTemplateList, error) {
	var resp WxaSubscribeTemplateList
	if err := w.doGet(ctx, "/wxaapi/newtmpl/gettemplate", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendSubscribeMessage 发送订阅消息。
// POST /cgi-bin/message/subscribe/send
func (w *WxaAdminClient) SendSubscribeMessage(ctx context.Context, req *WxaSendSubscribeReq) error {
	return w.doPost(ctx, "/cgi-bin/message/subscribe/send", req, nil)
}
```

- [ ] **Step 2.5: Run tests + commit**

```bash
go test ./oplatform/ -run TestWxaAdmin
go test ./oplatform/...
go build ./...

git add oplatform/wxa.submsg.go oplatform/wxa.submsg_test.go oplatform/wxa.struct.go
git commit -m "feat(oplatform): add wxa subscribe message & template library (7 methods)

GetSubscribeCategory / GetPubTemplateTitles / GetPubTemplateKeywords /
AddSubscribeTemplate / DeleteSubscribeTemplate / ListSubscribeTemplates /
SendSubscribeMessage on WxaAdminClient. ids parameter for
GetPubTemplateTitles is the WeChat-native \"-\" separated format
(e.g. \"2-3-5\"), passed through verbatim.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: `wxa.customer.go` — customer message (2 methods)

**Files:**
- Create: `oplatform/wxa.customer.go`
- Create: `oplatform/wxa.customer_test.go`
- Modify: `oplatform/wxa.struct.go` (append customer DTOs)

- [ ] **Step 3.1: Append customer DTOs to `oplatform/wxa.struct.go`**

Append at the bottom:

```go

// ----- customer message -----

type WxaCustomerTextPayload struct {
	Content string `json:"content"`
}

type WxaCustomerImagePayload struct {
	MediaID string `json:"media_id"`
}

type WxaCustomerLinkPayload struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	ThumbURL    string `json:"thumb_url"`
}

type WxaCustomerMiniProgramPagePayload struct {
	Title        string `json:"title"`
	Pagepath     string `json:"pagepath"`
	ThumbMediaID string `json:"thumb_media_id"`
}

type WxaSendCustomerMessageReq struct {
	ToUser          string                             `json:"touser"`
	MsgType         string                             `json:"msgtype"` // text/image/link/miniprogrampage
	Text            *WxaCustomerTextPayload            `json:"text,omitempty"`
	Image           *WxaCustomerImagePayload           `json:"image,omitempty"`
	Link            *WxaCustomerLinkPayload            `json:"link,omitempty"`
	MiniProgramPage *WxaCustomerMiniProgramPagePayload `json:"miniprogrampage,omitempty"`
}
```

- [ ] **Step 3.2: Write failing tests `oplatform/wxa.customer_test.go`**

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

func TestWxaAdmin_SendCustomerMessage_Text(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/message/custom/send") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.SendCustomerMessage(context.Background(), &WxaSendCustomerMessageReq{
		ToUser:  "OPENID_1",
		MsgType: "text",
		Text:    &WxaCustomerTextPayload{Content: "hello"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["touser"] != "OPENID_1" || body["msgtype"] != "text" {
		t.Errorf("body top: %+v", body)
	}
	text, _ := body["text"].(map[string]any)
	if text["content"] != "hello" {
		t.Errorf("body.text: %+v", body)
	}
	if _, ok := body["image"]; ok {
		t.Errorf("image should be omitted when nil")
	}
}

func TestWxaAdmin_SendCustomerMessage_Image(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.SendCustomerMessage(context.Background(), &WxaSendCustomerMessageReq{
		ToUser:  "OPENID_2",
		MsgType: "image",
		Image:   &WxaCustomerImagePayload{MediaID: "MEDIA_X"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["msgtype"] != "image" {
		t.Errorf("msgtype: %+v", body)
	}
	img, _ := body["image"].(map[string]any)
	if img["media_id"] != "MEDIA_X" {
		t.Errorf("body.image: %+v", body)
	}
}

func TestWxaAdmin_SendCustomerMessage_MiniProgramPage(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	err := w.SendCustomerMessage(context.Background(), &WxaSendCustomerMessageReq{
		ToUser:  "OPENID_3",
		MsgType: "miniprogrampage",
		MiniProgramPage: &WxaCustomerMiniProgramPagePayload{
			Title:        "订单详情",
			Pagepath:     "pages/order/detail?id=123",
			ThumbMediaID: "THUMB_1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if body["msgtype"] != "miniprogrampage" {
		t.Errorf("msgtype: %+v", body)
	}
	mpp, _ := body["miniprogrampage"].(map[string]any)
	if mpp["title"] != "订单详情" || mpp["pagepath"] != "pages/order/detail?id=123" {
		t.Errorf("body.miniprogrampage: %+v", body)
	}
}

func TestWxaAdmin_SendCustomerTyping(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/message/custom/typing") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.SendCustomerTyping(context.Background(), "OPENID_1", "Typing"); err != nil {
		t.Fatal(err)
	}
	if body["touser"] != "OPENID_1" || body["command"] != "Typing" {
		t.Errorf("body: %+v", body)
	}
}
```

- [ ] **Step 3.3: Run failing test — expect undefined methods.**

- [ ] **Step 3.4: Create `oplatform/wxa.customer.go`**

```go
package oplatform

import "context"

// SendCustomerMessage 发送客服消息。req.MsgType 合法值：
// "text" / "image" / "link" / "miniprogrampage"，与之对应的 payload
// 字段（Text/Image/Link/MiniProgramPage）填一个即可；其它字段会被
// omitempty 忽略。SDK 不做字段互斥校验，保持最薄封装。
// POST /cgi-bin/message/custom/send
func (w *WxaAdminClient) SendCustomerMessage(ctx context.Context, req *WxaSendCustomerMessageReq) error {
	return w.doPost(ctx, "/cgi-bin/message/custom/send", req, nil)
}

// SendCustomerTyping 下发"正在输入"状态。command 合法值：
// "Typing" 或 "CancelTyping"。
// POST /cgi-bin/message/custom/typing
func (w *WxaAdminClient) SendCustomerTyping(ctx context.Context, toUser, command string) error {
	body := map[string]string{
		"touser":  toUser,
		"command": command,
	}
	return w.doPost(ctx, "/cgi-bin/message/custom/typing", body, nil)
}
```

- [ ] **Step 3.5: Run tests + full sweep + commit**

```bash
go test ./oplatform/ -run TestWxaAdmin -v
go test -race ./...
go build ./...
go vet ./...

git add oplatform/wxa.customer.go oplatform/wxa.customer_test.go oplatform/wxa.struct.go
git commit -m "feat(oplatform): add wxa customer message (2 methods)

SendCustomerMessage supports text / image / link / miniprogrampage
payloads via a single request struct with mutually-optional pointer
fields; SDK does not enforce field exclusivity. SendCustomerTyping
wraps /cgi-bin/message/custom/typing for Typing / CancelTyping
commands.

Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Final verification sweep

- [ ] **Step 4.1: Full build + vet + test**

```bash
go build ./...
go vet ./...
go test -race ./...
```
All must be clean with every package showing `ok`.

- [ ] **Step 4.2: Method count verification**

```bash
grep -hE '^func \(w \*WxaAdminClient\)' oplatform/wxa.plugin.go oplatform/wxa.submsg.go oplatform/wxa.customer.go | wc -l
```
Expected: `16` (7 plugin + 7 submsg + 2 customer).

- [ ] **Step 4.3: Test count verification**

```bash
grep -hE '^func TestWxaAdmin_' oplatform/wxa.plugin_test.go oplatform/wxa.submsg_test.go oplatform/wxa.customer_test.go | wc -l
```
Expected: `19` (8 plugin + 7 submsg + 4 customer).

- [ ] **Step 4.4: Git log sanity**

Run: `git log --oneline 7acc77d^..HEAD`
Expected: 1 docs (spec) + 1 docs (plan) + 3 feat commits.

No commit at this step — verification only.

---

## Coverage Map (self-review)

| Spec section | Task |
|---|---|
| §2 WxaAdminClient reuse | Tasks 1-3 |
| §3.1 plugin 7 methods | Task 1 |
| §3.2 submsg 7 methods | Task 2 |
| §3.3 customer 2 methods | Task 3 |
| §4 DTOs | Tasks 1-3 (appended to wxa.struct.go) |
| §5 error handling | All tasks (doPost/doGet auto-fold errcode) |
| §6 concurrency/lifecycle | Trivial — stateless methods |
| §7 testing strategy | Tasks 1-3 use `newTestWxaAdmin` |
| §8 compatibility (additive) | All tasks |
| §9 delivery list | Tasks 1-3 |

**Total: 7 + 7 + 2 = 16 methods ✓**
**Tests: 8 + 7 + 4 = 19 cases ✓**

All method names consistent with spec:
- **Plugin:** `ApplyPlugin` / `ListPlugins` / `UnbindPlugin` / `GetPluginDevApplyList` / `AgreeDevPlugin` / `RefuseDevPlugin` / `DeleteDevPlugin`
- **Submsg:** `GetSubscribeCategory` / `GetPubTemplateTitles` / `GetPubTemplateKeywords` / `AddSubscribeTemplate` / `DeleteSubscribeTemplate` / `ListSubscribeTemplates` / `SendSubscribeMessage`
- **Customer:** `SendCustomerMessage` / `SendCustomerTyping`

No placeholders. All types (`WxaPluginList`, `WxaPluginDevApplyList`, `WxaSubscribeCategoryResp`, `WxaPubTemplateTitles`, `WxaPubTemplateKeywords`, `WxaAddSubscribeTemplateReq/Resp`, `WxaSubscribeTemplateList`, `WxaSendSubscribeReq`, `WxaSendCustomerMessageReq`) defined before use.
