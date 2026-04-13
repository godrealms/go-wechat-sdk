# Sub-project B: Test Coverage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bring offiaccount (3%→70%), merchant/developed/types (0%→70%), and utils (0%→80%) to meaningful test coverage using httptest mocking.

**Architecture:** All tests inject a *httptest.Server via WithHTTPClient/WithHTTP option; table-driven tests cover success, WeChat errcode error, and network error cases for every API method; no real WeChat API calls.

**Tech Stack:** Go 1.23.1, net/http/httptest, standard testing package

---

## Prerequisites

Before executing tasks, verify the following invariants hold. If any check fails, address it before proceeding.

- [ ] **P.1** Confirm `offiaccount.WeixinError` and `offiaccount.CheckResp` exist. Per sub-project A (plan `2026-04-13-subproject-a-code-quality.md`, tasks 4.3–4.4), these are added to `offiaccount/errors.go`. If that file is absent, create it now with:

  ```go
  package offiaccount

  import "fmt"

  // WeixinError wraps a WeChat API errcode/errmsg pair.
  type WeixinError struct {
      ErrCode int
      ErrMsg  string
  }

  func (e *WeixinError) Error() string {
      return fmt.Sprintf("offiaccount: errcode=%d errmsg=%s", e.ErrCode, e.ErrMsg)
  }

  func (e *WeixinError) Code() int    { return e.ErrCode }
  func (e *WeixinError) Message() string { return e.ErrMsg }

  // CheckResp returns a *WeixinError if r.ErrCode != 0, otherwise nil.
  func CheckResp(r *Resp) error {
      if r.ErrCode == 0 {
          return nil
      }
      return &WeixinError{ErrCode: r.ErrCode, ErrMsg: r.ErrMsg}
  }
  ```

- [ ] **P.2** Confirm the module path is `github.com/godrealms/go-wechat-sdk` (check `go.mod`).

- [ ] **P.3** Run `go build ./...` from the worktree root to confirm zero compile errors before starting.

---

## Shared test helpers (used across tasks)

The following `fixedToken` helper is referenced in every offiaccount test file. Each file declares it locally since Go test packages are per-directory. Copy it verbatim into each `_test.go` file that needs it.

```go
// fixedToken is a TokenSource that always returns a constant token.
type fixedToken struct{ tok string }

func (f fixedToken) AccessToken(_ context.Context) (string, error) { return f.tok, nil }
```

Helper `newTestClient` used by offiaccount tests:

```go
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
    t.Helper()
    h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
    return NewClient(context.Background(), &Config{AppId: "test", AppSecret: "secret"},
        WithHTTPClient(h),
        WithTokenSource(fixedToken{"FAKE_TOKEN"}),
    )
}
```

---

## Task 1: Test utils/http.go

**File:** `utils/http_test.go`

**Coverage target:** 80% of `utils/http.go`

- [ ] **1.1** Create `utils/http_test.go` with the following complete content:

  ```go
  package utils

  import (
      "context"
      "net/http"
      "net/http/httptest"
      "net/url"
      "strings"
      "testing"
      "time"
  )

  // helper: start a server that returns the given status and body
  func startServer(t *testing.T, status int, body string) *httptest.Server {
      t.Helper()
      return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.Header().Set("Content-Type", "application/json")
          w.WriteHeader(status)
          _, _ = w.Write([]byte(body))
      }))
  }

  // helper: new HTTP client pointed at srv
  func newHTTP(t *testing.T, srv *httptest.Server) *HTTP {
      t.Helper()
      return NewHTTP(srv.URL, WithTimeout(3*time.Second))
  }

  // --- Get ---

  func TestHTTP_Get_Success(t *testing.T) {
      type Resp struct {
          ErrCode int    `json:"errcode"`
          ErrMsg  string `json:"errmsg"`
      }
      srv := startServer(t, 200, `{"errcode":0,"errmsg":"ok"}`)
      defer srv.Close()

      h := newHTTP(t, srv)
      var result Resp
      if err := h.Get(context.Background(), "/test", nil, &result); err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result.ErrCode != 0 {
          t.Errorf("expected errcode 0, got %d", result.ErrCode)
      }
  }

  func TestHTTP_Get_WithQueryParams(t *testing.T) {
      var captured url.Values
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          captured = r.URL.Query()
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{}`))
      }))
      defer srv.Close()

      h := newHTTP(t, srv)
      q := url.Values{"access_token": {"TOKEN"}, "openid": {"oUser123"}}
      if err := h.Get(context.Background(), "/user/info", q, &struct{}{}); err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if captured.Get("access_token") != "TOKEN" {
          t.Errorf("expected access_token=TOKEN, got %q", captured.Get("access_token"))
      }
  }

  func TestHTTP_Get_Non2xxStatus(t *testing.T) {
      srv := startServer(t, 500, `internal server error`)
      defer srv.Close()

      h := newHTTP(t, srv)
      err := h.Get(context.Background(), "/fail", nil, &struct{}{})
      if err == nil {
          t.Fatal("expected error for 500 status")
      }
      if !strings.Contains(err.Error(), "500") {
          t.Errorf("expected 500 in error, got: %v", err)
      }
  }

  func TestHTTP_Get_NetworkError(t *testing.T) {
      // Point at a closed server to provoke a network error
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
      url := srv.URL
      srv.Close() // close immediately

      h := NewHTTP(url, WithTimeout(time.Second))
      err := h.Get(context.Background(), "/path", nil, &struct{}{})
      if err == nil {
          t.Fatal("expected network error")
      }
  }

  func TestHTTP_Get_InvalidJSON(t *testing.T) {
      srv := startServer(t, 200, `not-json`)
      defer srv.Close()

      h := newHTTP(t, srv)
      type Resp struct{ ErrCode int `json:"errcode"` }
      var result Resp
      err := h.Get(context.Background(), "/bad-json", nil, &result)
      if err == nil {
          t.Fatal("expected unmarshal error")
      }
      if !strings.Contains(err.Error(), "unmarshal") {
          t.Errorf("expected unmarshal in error, got: %v", err)
      }
  }

  // --- Post ---

  func TestHTTP_Post_Success(t *testing.T) {
      type ReqBody struct{ Foo string `json:"foo"` }
      type Resp struct{ Bar int `json:"bar"` }

      var gotBody []byte
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          if r.Method != http.MethodPost {
              t.Errorf("expected POST, got %s", r.Method)
          }
          if ct := r.Header.Get("Content-Type"); ct != "application/json" {
              t.Errorf("expected Content-Type application/json, got %q", ct)
          }
          gotBody = make([]byte, r.ContentLength)
          _, _ = r.Body.Read(gotBody)
          w.Header().Set("Content-Type", "application/json")
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{"bar":42}`))
      }))
      defer srv.Close()

      h := newHTTP(t, srv)
      var result Resp
      if err := h.Post(context.Background(), "/post", ReqBody{Foo: "hello"}, &result); err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result.Bar != 42 {
          t.Errorf("expected bar=42, got %d", result.Bar)
      }
  }

  func TestHTTP_Post_Non2xxStatus(t *testing.T) {
      srv := startServer(t, 400, `{"errcode":1,"errmsg":"bad request"}`)
      defer srv.Close()

      h := newHTTP(t, srv)
      err := h.Post(context.Background(), "/fail", map[string]string{"k": "v"}, &struct{}{})
      if err == nil {
          t.Fatal("expected error for 400 status")
      }
  }

  func TestHTTP_Post_NetworkError(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
      u := srv.URL
      srv.Close()

      h := NewHTTP(u, WithTimeout(time.Second))
      err := h.Post(context.Background(), "/path", map[string]string{}, &struct{}{})
      if err == nil {
          t.Fatal("expected network error")
      }
  }

  func TestHTTP_Post_NilBody(t *testing.T) {
      srv := startServer(t, 200, `{"ok":true}`)
      defer srv.Close()

      h := newHTTP(t, srv)
      // nil body should be accepted
      if err := h.Post(context.Background(), "/nil-body", nil, &struct{}{}); err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
  }

  // --- Put ---

  func TestHTTP_Put_Success(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          if r.Method != http.MethodPut {
              t.Errorf("expected PUT, got %s", r.Method)
          }
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{}`))
      }))
      defer srv.Close()

      h := newHTTP(t, srv)
      if err := h.Put(context.Background(), "/resource/1", map[string]string{"name": "v"}, &struct{}{}); err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
  }

  func TestHTTP_Put_Non2xxStatus(t *testing.T) {
      srv := startServer(t, 404, `not found`)
      defer srv.Close()

      h := newHTTP(t, srv)
      err := h.Put(context.Background(), "/notfound", map[string]string{}, &struct{}{})
      if err == nil {
          t.Fatal("expected error for 404 status")
      }
  }

  // --- WithHeaders option ---

  func TestHTTP_WithHeaders(t *testing.T) {
      var gotHeader string
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          gotHeader = r.Header.Get("X-Custom")
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{}`))
      }))
      defer srv.Close()

      h := NewHTTP(srv.URL,
          WithTimeout(3*time.Second),
          WithHeaders(map[string]string{"X-Custom": "myvalue"}),
      )
      _ = h.Get(context.Background(), "/", nil, &struct{}{})
      if gotHeader != "myvalue" {
          t.Errorf("expected X-Custom=myvalue, got %q", gotHeader)
      }
  }

  // --- SetBaseURL ---

  func TestHTTP_SetBaseURL(t *testing.T) {
      srv := startServer(t, 200, `{}`)
      defer srv.Close()

      h := NewHTTP("http://old.example.com")
      h.SetBaseURL(srv.URL)
      if h.BaseURL != srv.URL {
          t.Errorf("SetBaseURL did not update BaseURL")
      }
  }
  ```

- [ ] **1.2** Run and verify:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go test ./utils/... -v -count=1 -run TestHTTP
  ```

  Expected output (all tests PASS):
  ```
  --- PASS: TestHTTP_Get_Success (0.00s)
  --- PASS: TestHTTP_Get_WithQueryParams (0.00s)
  --- PASS: TestHTTP_Get_Non2xxStatus (0.00s)
  --- PASS: TestHTTP_Get_NetworkError (0.00s)
  --- PASS: TestHTTP_Get_InvalidJSON (0.00s)
  --- PASS: TestHTTP_Post_Success (0.00s)
  --- PASS: TestHTTP_Post_Non2xxStatus (0.00s)
  --- PASS: TestHTTP_Post_NetworkError (0.00s)
  --- PASS: TestHTTP_Post_NilBody (0.00s)
  --- PASS: TestHTTP_Put_Success (0.00s)
  --- PASS: TestHTTP_Put_Non2xxStatus (0.00s)
  --- PASS: TestHTTP_WithHeaders (0.00s)
  --- PASS: TestHTTP_SetBaseURL (0.00s)
  PASS
  ok      github.com/godrealms/go-wechat-sdk/utils
  ```

- [ ] **1.3** Verify coverage is at least 80%:

  ```bash
  go test ./utils/... -cover -count=1
  ```

  Expected: `coverage: 80%+` on the `utils` line.

- [ ] **1.4** Commit:

  ```bash
  git add utils/http_test.go
  git commit -m "test(utils): add HTTP client table-driven tests (Get/Post/Put/headers)"
  ```

---

## Task 2: Test offiaccount/client.go and token management

**File:** `offiaccount/client_test.go` (already exists — extend it)

The existing `offiaccount/client_test.go` already covers: token caching, WeixinError return, backwards-compatible `GetAccessToken`, `CheckResp`, and `WithTokenSource`. Verify these pass and add the missing `AccessToken` struct field test.

- [ ] **2.1** Run the existing client tests to confirm they pass:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go test ./offiaccount/... -v -run "TestClient|TestCheckResp" -count=1
  ```

  Expected: all 5 existing tests PASS (TestClient_AccessTokenE_CachesAndRefreshes, TestClient_AccessTokenE_ReturnsWeixinError, TestClient_GetAccessToken_BackwardsCompatible, TestCheckResp, TestClient_AccessTokenE_UsesInjectedTokenSource).

- [ ] **2.2** Append the following additional tests to `offiaccount/client_test.go` (do not replace the file, append after the last function):

  ```go
  func TestClient_NewClient_NilConfig(t *testing.T) {
      // NewClient must not panic on nil config; refreshAccessToken will error lazily
      c := NewClient(nil, nil)
      if c == nil {
          t.Fatal("expected non-nil client")
      }
  }

  func TestClient_NewClient_WithHTTPClient_IgnoresNilHTTP(t *testing.T) {
      c := NewClient(context.Background(), &Config{AppId: "a"},
          WithHTTPClient(nil), // nil must be a no-op
      )
      if c.Https == nil {
          t.Error("expected Https to remain non-nil after WithHTTPClient(nil)")
      }
  }

  func TestClient_refreshAccessToken_EmptyToken(t *testing.T) {
      // Server returns 200 but empty access_token — must error
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          _, _ = w.Write([]byte(`{"access_token":"","expires_in":7200}`))
      }))
      defer srv.Close()

      c := newClientWithBaseURL(srv.URL, &Config{AppId: "a", AppSecret: "b"})
      _, err := c.AccessTokenE(context.Background())
      if err == nil {
          t.Fatal("expected error for empty access_token")
      }
  }

  func TestClient_refreshAccessToken_MissingCredentials(t *testing.T) {
      c := NewClient(context.Background(), &Config{})
      _, err := c.AccessTokenE(context.Background())
      if err == nil {
          t.Fatal("expected error when AppId and AppSecret are empty")
      }
  }
  ```

- [ ] **2.3** Run all offiaccount client tests:

  ```bash
  go test ./offiaccount/... -v -run "TestClient|TestCheckResp" -count=1
  ```

  Expected: all 9 tests PASS.

- [ ] **2.4** Commit:

  ```bash
  git add offiaccount/client_test.go
  git commit -m "test(offiaccount): extend client_test with nil-safety and edge-case token tests"
  ```

---

## Task 3: Test offiaccount menu APIs

**File:** `offiaccount/api.menu_test.go` (new file)

Covers: `CreateCustomMenu`, `GetCurrentSelfMenuInfo`, `GetMenu`, `DeleteMenu`, `AddConditionalMenu`, `DeleteConditionalMenu`, `TryMatchMenu` from `api.custom-menu.go`.

- [ ] **3.1** Create `offiaccount/api.menu_test.go`:

  ```go
  package offiaccount

  import (
      "context"
      "net/http"
      "net/http/httptest"
      "strings"
      "testing"
      "time"

      "github.com/godrealms/go-wechat-sdk/utils"
  )

  // fixedToken is a TokenSource that always returns a constant token.
  type fixedToken struct{ tok string }

  func (f fixedToken) AccessToken(_ context.Context) (string, error) { return f.tok, nil }

  func newMenuTestClient(t *testing.T, srv *httptest.Server) *Client {
      t.Helper()
      h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
      return NewClient(context.Background(), &Config{AppId: "test", AppSecret: "secret"},
          WithHTTPClient(h),
          WithTokenSource(fixedToken{"FAKE_TOKEN"}),
      )
  }

  func TestCreateCustomMenu(t *testing.T) {
      tests := []struct {
          name    string
          body    string
          status  int
          wantErr bool
          errMsg  string
      }{
          {
              name:    "success",
              body:    `{"errcode":0,"errmsg":"ok"}`,
              status:  200,
              wantErr: false,
          },
          {
              name:    "wechat errcode error",
              body:    `{"errcode":65001,"errmsg":"invalid menu type"}`,
              status:  200,
              wantErr: true,
              errMsg:  "invalid menu type",
          },
          {
              name:    "network error",
              wantErr: true,
          },
      }

      for _, tc := range tests {
          t.Run(tc.name, func(t *testing.T) {
              var srv *httptest.Server
              if tc.name == "network error" {
                  srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
                  srv.Close()
              } else {
                  srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                      w.WriteHeader(tc.status)
                      _, _ = w.Write([]byte(tc.body))
                  }))
                  defer srv.Close()
              }
              c := newMenuTestClient(t, srv)
              err := c.CreateCustomMenu(&CreateMenuButton{})
              if tc.wantErr && err == nil {
                  t.Error("expected error, got nil")
              }
              if !tc.wantErr && err != nil {
                  t.Errorf("unexpected error: %v", err)
              }
              if tc.errMsg != "" && err != nil && !strings.Contains(err.Error(), tc.errMsg) {
                  t.Errorf("expected error containing %q, got %v", tc.errMsg, err)
              }
          })
      }
  }

  func TestGetMenu(t *testing.T) {
      tests := []struct {
          name    string
          body    string
          wantErr bool
      }{
          {"success", `{"menu":{"button":[]}}`, false},
          {"network error", "", true},
      }

      for _, tc := range tests {
          t.Run(tc.name, func(t *testing.T) {
              var srv *httptest.Server
              if tc.name == "network error" {
                  srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
                  srv.Close()
              } else {
                  srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                      w.WriteHeader(200)
                      _, _ = w.Write([]byte(tc.body))
                  }))
                  defer srv.Close()
              }
              c := newMenuTestClient(t, srv)
              result, err := c.GetMenu()
              if tc.wantErr {
                  if err == nil {
                      t.Error("expected error, got nil")
                  }
                  return
              }
              if err != nil {
                  t.Fatalf("unexpected error: %v", err)
              }
              if result == nil {
                  t.Error("expected non-nil result")
              }
          })
      }
  }

  func TestDeleteMenu(t *testing.T) {
      tests := []struct {
          name    string
          body    string
          wantErr bool
      }{
          {"success", `{"errcode":0,"errmsg":"ok"}`, false},
          {"errcode error", `{"errcode":40001,"errmsg":"invalid credential"}`, true},
          {"network error", "", true},
      }

      for _, tc := range tests {
          t.Run(tc.name, func(t *testing.T) {
              var srv *httptest.Server
              if tc.name == "network error" {
                  srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
                  srv.Close()
              } else {
                  srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                      w.WriteHeader(200)
                      _, _ = w.Write([]byte(tc.body))
                  }))
                  defer srv.Close()
              }
              c := newMenuTestClient(t, srv)
              err := c.DeleteMenu()
              if tc.wantErr && err == nil {
                  t.Error("expected error, got nil")
              }
              if !tc.wantErr && err != nil {
                  t.Errorf("unexpected error: %v", err)
              }
          })
      }
  }

  func TestGetCurrentSelfMenuInfo(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{"is_menu_open":1,"selfmenu_info":{"button":[]}}`))
      }))
      defer srv.Close()

      c := newMenuTestClient(t, srv)
      result, err := c.GetCurrentSelfMenuInfo()
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result == nil {
          t.Error("expected non-nil result")
      }
  }

  func TestAddConditionalMenu(t *testing.T) {
      tests := []struct {
          name    string
          body    string
          wantErr bool
      }{
          {"success", `{"errcode":0,"errmsg":"ok","menuid":"502394"}`, false},
          {"errcode error", `{"errcode":65301,"errmsg":"menu count limit"}`, true},
          {"network error", "", true},
      }

      for _, tc := range tests {
          t.Run(tc.name, func(t *testing.T) {
              var srv *httptest.Server
              if tc.name == "network error" {
                  srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
                  srv.Close()
              } else {
                  srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                      w.WriteHeader(200)
                      _, _ = w.Write([]byte(tc.body))
                  }))
                  defer srv.Close()
              }
              c := newMenuTestClient(t, srv)
              _, err := c.AddConditionalMenu(&ConditionalMenu{})
              if tc.wantErr && err == nil {
                  t.Error("expected error, got nil")
              }
              if !tc.wantErr && err != nil {
                  t.Errorf("unexpected error: %v", err)
              }
          })
      }
  }

  func TestDeleteConditionalMenu(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
      }))
      defer srv.Close()

      c := newMenuTestClient(t, srv)
      _, err := c.DeleteConditionalMenu("502394")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
  }

  func TestTryMatchMenu(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{"menu":{"button":[]}}`))
      }))
      defer srv.Close()

      c := newMenuTestClient(t, srv)
      result, err := c.TryMatchMenu("oUser123")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result == nil {
          t.Error("expected non-nil result")
      }
  }
  ```

  **Note:** `ConditionalMenu` and `AddConditionalMenuResponse` must exist in `struct.menu.go`. Verify with `grep -n "ConditionalMenu" offiaccount/struct.menu.go` — if absent they will be defined in that struct file as part of the existing codebase.

- [ ] **3.2** Run:

  ```bash
  go test ./offiaccount/... -v -run "TestCreateCustomMenu|TestGetMenu|TestDeleteMenu|TestGetCurrentSelfMenuInfo|TestAddConditionalMenu|TestDeleteConditionalMenu|TestTryMatchMenu" -count=1
  ```

  Expected: all tests PASS.

- [ ] **3.3** Commit:

  ```bash
  git add offiaccount/api.menu_test.go
  git commit -m "test(offiaccount): add menu API tests (CreateCustomMenu, GetMenu, DeleteMenu, conditional menus)"
  ```

---

## Task 4: Test offiaccount messaging APIs

**File:** `offiaccount/api.message_test.go` (new file)

Covers: `SendTemplateMessage`, `AddTemplate`, `DeleteTemplate`, `GetAllTemplates`, `SetIndustry`, `TemplateSubscribe`, `DeleteMassMsg`, `MassSend`, `Preview`, `SendAll`, `SetSpeed` from `api.notify.template.go`, `api.notify.subscribe.go`, `api.notify.message.go`.

- [ ] **4.1** Create `offiaccount/api.message_test.go`:

  ```go
  package offiaccount

  import (
      "context"
      "net/http"
      "net/http/httptest"
      "testing"
      "time"

      "github.com/godrealms/go-wechat-sdk/utils"
  )

  func newMsgTestClient(t *testing.T, srv *httptest.Server) *Client {
      t.Helper()
      h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
      return NewClient(context.Background(), &Config{AppId: "test", AppSecret: "secret"},
          WithHTTPClient(h),
          WithTokenSource(fixedToken{"FAKE_TOKEN"}),
      )
  }

  // jsonServer returns a test server that always replies with the given JSON body and status.
  func jsonServer(t *testing.T, status int, body string) *httptest.Server {
      t.Helper()
      return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.Header().Set("Content-Type", "application/json")
          w.WriteHeader(status)
          _, _ = w.Write([]byte(body))
      }))
  }

  func TestSendTemplateMessage(t *testing.T) {
      tests := []struct {
          name    string
          resp    string
          wantErr bool
      }{
          {"success", `{"errcode":0,"errmsg":"ok","msgid":123456}`, false},
          {"errcode error", `{"errcode":40003,"errmsg":"invalid openid"}`, true},
      }
      for _, tc := range tests {
          t.Run(tc.name, func(t *testing.T) {
              srv := jsonServer(t, 200, tc.resp)
              defer srv.Close()
              c := newMsgTestClient(t, srv)
              req := &SubscribeMessageRequest{
                  Touser:     "oUser123",
                  TemplateId: "tplId",
              }
              result, err := c.SendTemplateMessage(req)
              if tc.wantErr {
                  if err == nil {
                      t.Error("expected error, got nil")
                  }
                  return
              }
              if err != nil {
                  t.Fatalf("unexpected error: %v", err)
              }
              if result == nil {
                  t.Error("expected non-nil result")
              }
          })
      }
  }

  func TestSendTemplateMessage_NetworkError(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
      srv.Close()
      c := newMsgTestClient(t, srv)
      _, err := c.SendTemplateMessage(&SubscribeMessageRequest{})
      if err == nil {
          t.Error("expected network error")
      }
  }

  func TestAddTemplate(t *testing.T) {
      srv := jsonServer(t, 200, `{"errcode":0,"errmsg":"ok","template_id":"tpl123"}`)
      defer srv.Close()
      c := newMsgTestClient(t, srv)
      result, err := c.AddTemplate("shortId", []string{"keyword1", "keyword2"})
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result == nil {
          t.Error("expected non-nil result")
      }
  }

  func TestAddTemplate_ErrCode(t *testing.T) {
      srv := jsonServer(t, 200, `{"errcode":40015,"errmsg":"invalid template id"}`)
      defer srv.Close()
      c := newMsgTestClient(t, srv)
      _, err := c.AddTemplate("bad", nil)
      if err == nil {
          t.Error("expected error, got nil")
      }
  }

  func TestDeleteTemplate(t *testing.T) {
      srv := jsonServer(t, 200, `{"errcode":0,"errmsg":"ok"}`)
      defer srv.Close()
      c := newMsgTestClient(t, srv)
      _, err := c.DeleteTemplate("tplId")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
  }

  func TestGetAllTemplates(t *testing.T) {
      srv := jsonServer(t, 200, `{"errcode":0,"errmsg":"ok","template_list":[]}`)
      defer srv.Close()
      c := newMsgTestClient(t, srv)
      result, err := c.GetAllTemplates()
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result == nil {
          t.Error("expected non-nil result")
      }
  }

  func TestSetIndustry(t *testing.T) {
      srv := jsonServer(t, 200, `{"errcode":0,"errmsg":"ok"}`)
      defer srv.Close()
      c := newMsgTestClient(t, srv)
      if err := c.SetIndustry("1", "2"); err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
  }

  func TestTemplateSubscribe(t *testing.T) {
      tests := []struct {
          name    string
          resp    string
          wantErr bool
      }{
          {"success", `{"errcode":0,"errmsg":"ok"}`, false},
          {"errcode error", `{"errcode":43101,"errmsg":"user refuse to accept the msg"}`, true},
      }
      for _, tc := range tests {
          t.Run(tc.name, func(t *testing.T) {
              srv := jsonServer(t, 200, tc.resp)
              defer srv.Close()
              c := newMsgTestClient(t, srv)
              err := c.TemplateSubscribe(&TemplateSubscribeReq{})
              if tc.wantErr && err == nil {
                  t.Error("expected error, got nil")
              }
              if !tc.wantErr && err != nil {
                  t.Errorf("unexpected error: %v", err)
              }
          })
      }
  }

  func TestDeleteMassMsg(t *testing.T) {
      srv := jsonServer(t, 200, `{"errcode":0,"errmsg":"ok"}`)
      defer srv.Close()
      c := newMsgTestClient(t, srv)
      msgId := "1234567"
      if err := c.DeleteMassMsg(&DeleteMassMsgRequest{MsgId: &msgId}); err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
  }

  func TestMassSend(t *testing.T) {
      srv := jsonServer(t, 200, `{"errcode":0,"errmsg":"send job submission success","msg_id":34182}`)
      defer srv.Close()
      c := newMsgTestClient(t, srv)
      result, err := c.MassSend(&MassSendRequest{})
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result == nil {
          t.Error("expected non-nil result")
      }
  }

  func TestPreview(t *testing.T) {
      srv := jsonServer(t, 200, `{"errcode":0,"errmsg":"ok"}`)
      defer srv.Close()
      c := newMsgTestClient(t, srv)
      result, err := c.Preview(&MassSendRequest{})
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result == nil {
          t.Error("expected non-nil result")
      }
  }

  func TestSendAll(t *testing.T) {
      srv := jsonServer(t, 200, `{"errcode":0,"errmsg":"send job submission success","msg_id":34183}`)
      defer srv.Close()
      c := newMsgTestClient(t, srv)
      result, err := c.SendAll(&MassSendByTagRequest{})
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result == nil {
          t.Error("expected non-nil result")
      }
  }

  func TestSetSpeed(t *testing.T) {
      srv := jsonServer(t, 200, `{"errcode":0,"errmsg":"ok"}`)
      defer srv.Close()
      c := newMsgTestClient(t, srv)
      result, err := c.SetSpeed(4)
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result == nil {
          t.Error("expected non-nil result")
      }
  }
  ```

  **Note on `DeleteMassMsgRequest`:** The struct in `struct.massmsg.go` uses `MsgId *string`. The test reflects that. If the field type differs, adjust accordingly.

- [ ] **4.2** Verify `DeleteMassMsgRequest.MsgId` field type:

  ```bash
  grep -A5 "type DeleteMassMsgRequest" offiaccount/struct.massmsg.go
  ```

  Adjust the test if `MsgId` is `string` (not `*string`) — replace `msgId := "1234567"; &DeleteMassMsgRequest{MsgId: &msgId}` with `&DeleteMassMsgRequest{MsgId: "1234567"}`.

- [ ] **4.3** Run:

  ```bash
  go test ./offiaccount/... -v -run "TestSendTemplate|TestAddTemplate|TestDeleteTemplate|TestGetAllTemplates|TestSetIndustry|TestTemplateSubscribe|TestDeleteMassMsg|TestMassSend|TestPreview|TestSendAll|TestSetSpeed" -count=1
  ```

  Expected: all tests PASS.

- [ ] **4.4** Commit:

  ```bash
  git add offiaccount/api.message_test.go
  git commit -m "test(offiaccount): add messaging API tests (template, subscribe, mass send)"
  ```

---

## Task 5: Test offiaccount user management APIs

**File:** `offiaccount/api.user_test.go` (new file)

Covers: `GetUserInfo`, `BatchGetUserInfo`, `UpdateRemark`, `GetFans`, `GetBlacklist`, `BatchBlacklist`, `BatchUnblacklist` from `api.user.manage.userinfo.go`; `GetTags`, `CreateTag`, `UpdateTag`, `DeleteTag`, `GetTagFans`, `BatchTagging`, `BatchUntagging`, `GetTagidList` from `api.user.manage.tag.go`.

- [ ] **5.1** Create `offiaccount/api.user_test.go`:

  ```go
  package offiaccount

  import (
      "context"
      "net/http"
      "net/http/httptest"
      "testing"
      "time"

      "github.com/godrealms/go-wechat-sdk/utils"
  )

  func newUserTestClient(t *testing.T, srv *httptest.Server) *Client {
      t.Helper()
      h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
      return NewClient(context.Background(), &Config{AppId: "test", AppSecret: "secret"},
          WithHTTPClient(h),
          WithTokenSource(fixedToken{"FAKE_TOKEN"}),
      )
  }

  func okServer(t *testing.T, body string) *httptest.Server {
      t.Helper()
      return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.Header().Set("Content-Type", "application/json")
          w.WriteHeader(200)
          _, _ = w.Write([]byte(body))
      }))
  }

  func closedServer(t *testing.T) *httptest.Server {
      t.Helper()
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
      srv.Close()
      return srv
  }

  // --- GetUserInfo ---

  func TestGetUserInfo_Success(t *testing.T) {
      body := `{"subscribe":1,"openid":"oUser123","language":"zh_CN"}`
      srv := okServer(t, body)
      defer srv.Close()

      c := newUserTestClient(t, srv)
      result, err := c.GetUserInfo("oUser123", "zh_CN")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result.Openid != "oUser123" {
          t.Errorf("expected openid oUser123, got %q", result.Openid)
      }
  }

  func TestGetUserInfo_NetworkError(t *testing.T) {
      c := newUserTestClient(t, closedServer(t))
      _, err := c.GetUserInfo("oUser123", "")
      if err == nil {
          t.Error("expected network error")
      }
  }

  // --- GetFans ---

  func TestGetFans_Success(t *testing.T) {
      body := `{"total":2,"count":2,"data":{"openid":["oUser1","oUser2"]},"next_openid":""}`
      srv := okServer(t, body)
      defer srv.Close()

      c := newUserTestClient(t, srv)
      result, err := c.GetFans("")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result.Total != 2 {
          t.Errorf("expected total=2, got %d", result.Total)
      }
  }

  func TestGetFans_NetworkError(t *testing.T) {
      c := newUserTestClient(t, closedServer(t))
      _, err := c.GetFans("")
      if err == nil {
          t.Error("expected network error")
      }
  }

  // --- GetBlacklist ---

  func TestGetBlacklist_Success(t *testing.T) {
      body := `{"total":1,"count":1,"data":{"openid":["oBlack1"]},"next_openid":""}`
      srv := okServer(t, body)
      defer srv.Close()

      c := newUserTestClient(t, srv)
      result, err := c.GetBlacklist("")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result.Count != 1 {
          t.Errorf("expected count=1, got %d", result.Count)
      }
  }

  // --- BatchBlacklist / BatchUnblacklist ---

  func TestBatchBlacklist_Success(t *testing.T) {
      srv := okServer(t, `{"errcode":0,"errmsg":"ok"}`)
      defer srv.Close()
      c := newUserTestClient(t, srv)
      _, err := c.BatchBlacklist([]string{"oUser1"})
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
  }

  func TestBatchUnblacklist_Success(t *testing.T) {
      srv := okServer(t, `{"errcode":0,"errmsg":"ok"}`)
      defer srv.Close()
      c := newUserTestClient(t, srv)
      _, err := c.BatchUnblacklist([]string{"oUser1"})
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
  }

  // --- UpdateRemark ---

  func TestUpdateRemark_Success(t *testing.T) {
      srv := okServer(t, `{"errcode":0,"errmsg":"ok"}`)
      defer srv.Close()
      c := newUserTestClient(t, srv)
      _, err := c.UpdateRemark("oUser1", "my remark")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
  }

  // --- BatchGetUserInfo ---

  func TestBatchGetUserInfo_Success(t *testing.T) {
      body := `{"user_info_list":[{"openid":"oUser1"},{"openid":"oUser2"}]}`
      srv := okServer(t, body)
      defer srv.Close()
      c := newUserTestClient(t, srv)
      req := &BatchGetUserInfoRequest{
          UserList: []*UserListItem{
              {Openid: "oUser1", Language: "zh_CN"},
              {Openid: "oUser2", Language: "zh_CN"},
          },
      }
      result, err := c.BatchGetUserInfo(req)
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if len(result.UserInfoList) != 2 {
          t.Errorf("expected 2 users, got %d", len(result.UserInfoList))
      }
  }

  // --- Tags ---

  func TestGetTags_Success(t *testing.T) {
      srv := okServer(t, `{"tags":[{"id":1,"name":"VIP","count":100}]}`)
      defer srv.Close()
      c := newUserTestClient(t, srv)
      result, err := c.GetTags()
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if len(result.Tags) != 1 {
          t.Errorf("expected 1 tag, got %d", len(result.Tags))
      }
  }

  func TestCreateTag_Success(t *testing.T) {
      srv := okServer(t, `{"tag":{"id":100,"name":"VIP","count":0}}`)
      defer srv.Close()
      c := newUserTestClient(t, srv)
      result, err := c.CreateTag("VIP")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result.Tag.Name != "VIP" {
          t.Errorf("expected tag name VIP, got %q", result.Tag.Name)
      }
  }

  func TestCreateTag_NetworkError(t *testing.T) {
      c := newUserTestClient(t, closedServer(t))
      _, err := c.CreateTag("VIP")
      if err == nil {
          t.Error("expected network error")
      }
  }

  func TestUpdateTag_Success(t *testing.T) {
      srv := okServer(t, `{"errcode":0,"errmsg":"ok"}`)
      defer srv.Close()
      c := newUserTestClient(t, srv)
      _, err := c.UpdateTag(100, "Premium")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
  }

  func TestDeleteTag_Success(t *testing.T) {
      srv := okServer(t, `{"errcode":0,"errmsg":"ok"}`)
      defer srv.Close()
      c := newUserTestClient(t, srv)
      _, err := c.DeleteTag(100)
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
  }

  func TestGetTagFans_Success(t *testing.T) {
      body := `{"count":1,"data":{"openid":["oUser1"]},"next_openid":""}`
      srv := okServer(t, body)
      defer srv.Close()
      c := newUserTestClient(t, srv)
      result, err := c.GetTagFans(&GetTagFansRequest{TagId: 100})
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result.Count != 1 {
          t.Errorf("expected count=1, got %d", result.Count)
      }
  }

  func TestBatchTagging_Success(t *testing.T) {
      srv := okServer(t, `{"errcode":0,"errmsg":"ok"}`)
      defer srv.Close()
      c := newUserTestClient(t, srv)
      _, err := c.BatchTagging([]string{"oUser1", "oUser2"}, 100)
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
  }

  func TestBatchUntagging_Success(t *testing.T) {
      srv := okServer(t, `{"errcode":0,"errmsg":"ok"}`)
      defer srv.Close()
      c := newUserTestClient(t, srv)
      _, err := c.BatchUntagging([]string{"oUser1"}, 100)
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
  }

  func TestGetTagidList_Success(t *testing.T) {
      srv := okServer(t, `{"tagid_list":[100,200]}`)
      defer srv.Close()
      c := newUserTestClient(t, srv)
      result, err := c.GetTagidList("oUser1")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if len(result.TagidList) != 2 {
          t.Errorf("expected 2 tagids, got %d", len(result.TagidList))
      }
  }
  ```

- [ ] **5.2** Run:

  ```bash
  go test ./offiaccount/... -v -run "TestGetUserInfo|TestGetFans|TestGetBlacklist|TestBatchBlacklist|TestBatchUnblacklist|TestUpdateRemark|TestBatchGetUserInfo|TestGetTags|TestCreateTag|TestUpdateTag|TestDeleteTag|TestGetTagFans|TestBatchTagging|TestBatchUntagging|TestGetTagidList" -count=1
  ```

  Expected: all tests PASS.

- [ ] **5.3** Commit:

  ```bash
  git add offiaccount/api.user_test.go
  git commit -m "test(offiaccount): add user management and tag API tests"
  ```

---

## Task 6: Test offiaccount material APIs

**File:** `offiaccount/api.material_test.go` (new file)

Covers: `GetTempMedia` (binary and JSON paths), `GetMaterial`, `AddDraft`, `GetDraft` from `api.material.temporary.go`, `api.material.permanent.go`, `api.draft-box.draft.manage.go`.

- [ ] **6.1** Create `offiaccount/api.material_test.go`:

  ```go
  package offiaccount

  import (
      "context"
      "net/http"
      "net/http/httptest"
      "testing"
      "time"

      "github.com/godrealms/go-wechat-sdk/utils"
  )

  func newMaterialTestClient(t *testing.T, srv *httptest.Server) *Client {
      t.Helper()
      h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
      return NewClient(context.Background(), &Config{AppId: "test", AppSecret: "secret"},
          WithHTTPClient(h),
          WithTokenSource(fixedToken{"FAKE_TOKEN"}),
      )
  }

  // TestGetTempMedia_BinaryResponse: server returns binary (non-JSON content-type).
  func TestGetTempMedia_BinaryResponse(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.Header().Set("Content-Type", "image/jpeg")
          w.WriteHeader(200)
          _, _ = w.Write([]byte{0xFF, 0xD8, 0xFF}) // fake JPEG bytes
      }))
      defer srv.Close()

      c := newMaterialTestClient(t, srv)
      data, videoResult, err := c.GetTempMedia("media123")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if videoResult != nil {
          t.Error("expected nil videoResult for binary response")
      }
      if len(data) == 0 {
          t.Error("expected non-empty binary data")
      }
  }

  // TestGetTempMedia_JSONResponse: server returns JSON (e.g. video URL or error).
  func TestGetTempMedia_JSONResponse(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.Header().Set("Content-Type", "application/json")
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{"video_url":"https://video.weixin.qq.com/xxx","down_url":"https://video.weixin.qq.com/yyy"}`))
      }))
      defer srv.Close()

      c := newMaterialTestClient(t, srv)
      data, videoResult, err := c.GetTempMedia("media123")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if data != nil {
          t.Error("expected nil data for JSON response")
      }
      if videoResult == nil {
          t.Error("expected non-nil videoResult")
      }
  }

  // TestGetTempMedia_NetworkError
  func TestGetTempMedia_NetworkError(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
      srv.Close()
      c := newMaterialTestClient(t, srv)
      _, _, err := c.GetTempMedia("media123")
      if err == nil {
          t.Error("expected network error")
      }
  }

  // TestGetMaterial_News: permanent material returns news JSON
  func TestGetMaterial_News(t *testing.T) {
      body := `{"news_item":[{"title":"Article 1","author":"Author","content":"Hello"}]}`
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.Header().Set("Content-Type", "application/json")
          w.WriteHeader(200)
          _, _ = w.Write([]byte(body))
      }))
      defer srv.Close()

      c := newMaterialTestClient(t, srv)
      newsResult, videoResult, rawData, err := c.GetMaterial("media_news_123")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      _ = videoResult
      _ = rawData
      if newsResult == nil {
          t.Error("expected non-nil newsResult")
      }
  }

  // TestGetMaterial_NetworkError
  func TestGetMaterial_NetworkError(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
      srv.Close()
      c := newMaterialTestClient(t, srv)
      _, _, _, err := c.GetMaterial("media123")
      if err == nil {
          t.Error("expected network error")
      }
  }

  // TestAddDraft_Success
  func TestAddDraft_Success(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.Header().Set("Content-Type", "application/json")
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{"media_id":"draft123"}`))
      }))
      defer srv.Close()

      c := newMaterialTestClient(t, srv)
      result, err := c.AddDraft([]*DraftArticle{{Title: "Hello"}})
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result == nil {
          t.Error("expected non-nil result")
      }
  }

  // TestAddDraft_NetworkError
  func TestAddDraft_NetworkError(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
      srv.Close()
      c := newMaterialTestClient(t, srv)
      _, err := c.AddDraft([]*DraftArticle{})
      if err == nil {
          t.Error("expected network error")
      }
  }

  // TestGetDraft_Success
  func TestGetDraft_Success(t *testing.T) {
      body := `{"news_item":[{"title":"Draft 1","author":"Writer"}],"update_time":1700000000}`
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.Header().Set("Content-Type", "application/json")
          w.WriteHeader(200)
          _, _ = w.Write([]byte(body))
      }))
      defer srv.Close()

      c := newMaterialTestClient(t, srv)
      result, err := c.GetDraft("draft123")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result == nil {
          t.Error("expected non-nil result")
      }
  }
  ```

  **Note on types:** `GetTempMediaVideoResult` is defined in `struct.material.go`; `DraftArticle`, `AddDraftResult`, `GetDraftResult` are in `struct.draft.go`; `GetMaterialNewsResult`, `GetMaterialVideoResult` are in `struct.material.go`. Verify these exist with:
  ```bash
  grep -n "type.*Result\|type DraftArticle" offiaccount/struct.material.go offiaccount/struct.draft.go
  ```

- [ ] **6.2** Run:

  ```bash
  go test ./offiaccount/... -v -run "TestGetTempMedia|TestGetMaterial|TestAddDraft|TestGetDraft" -count=1
  ```

  Expected: all tests PASS.

- [ ] **6.3** Commit:

  ```bash
  git add offiaccount/api.material_test.go
  git commit -m "test(offiaccount): add material and draft API tests"
  ```

---

## Task 7: Test offiaccount QR code and base APIs

**File:** `offiaccount/api.base_test.go` (new file)

Covers: `GetCallbackIp`, `GetApiDomainIP` from `api.base.go`; `CreateQRCode`, `GetQRCodeURL` from `api.qr-code.qr-codes.go`.

- [ ] **7.1** Create `offiaccount/api.base_test.go`:

  ```go
  package offiaccount

  import (
      "context"
      "net/http"
      "net/http/httptest"
      "strings"
      "testing"
      "time"

      "github.com/godrealms/go-wechat-sdk/utils"
  )

  func newBaseTestClient(t *testing.T, srv *httptest.Server) *Client {
      t.Helper()
      h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
      return NewClient(context.Background(), &Config{AppId: "test", AppSecret: "secret"},
          WithHTTPClient(h),
          WithTokenSource(fixedToken{"FAKE_TOKEN"}),
      )
  }

  func TestGetCallbackIp_Success(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{"ip_list":["101.226.103.0/25","101.226.62.0/26"]}`))
      }))
      defer srv.Close()

      c := newBaseTestClient(t, srv)
      ips, err := c.GetCallbackIp()
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if len(ips) != 2 {
          t.Errorf("expected 2 IPs, got %d", len(ips))
      }
  }

  func TestGetCallbackIp_ErrCode(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid credential"}`))
      }))
      defer srv.Close()

      c := newBaseTestClient(t, srv)
      _, err := c.GetCallbackIp()
      if err == nil {
          t.Error("expected error for errcode 40001")
      }
  }

  func TestGetCallbackIp_NetworkError(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
      srv.Close()
      c := newBaseTestClient(t, srv)
      _, err := c.GetCallbackIp()
      if err == nil {
          t.Error("expected network error")
      }
  }

  func TestGetApiDomainIP_Success(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{"ip_list":["203.205.254.0/24","203.205.226.0/24"]}`))
      }))
      defer srv.Close()

      c := newBaseTestClient(t, srv)
      ips, err := c.GetApiDomainIP()
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if len(ips) != 2 {
          t.Errorf("expected 2 IPs, got %d", len(ips))
      }
  }

  func TestGetApiDomainIP_ErrCode(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{"errcode":40013,"errmsg":"invalid appid"}`))
      }))
      defer srv.Close()

      c := newBaseTestClient(t, srv)
      _, err := c.GetApiDomainIP()
      if err == nil {
          t.Error("expected error for errcode 40013")
      }
  }

  func TestCreateQRCode_Success(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{"ticket":"gQH47joAAAAAAAAAASxodHRwOi8vd2VpeGlu","expire_seconds":604800,"url":"http://weixin.qq.com/q/kZgfwMTm72WWPkovabbI"}`))
      }))
      defer srv.Close()

      c := newBaseTestClient(t, srv)
      req := &CreateQRCodeRequest{
          ActionName: "QR_SCENE",
          ActionInfo: ActionInfo{Scene: Scene{SceneID: 123}},
      }
      result, err := c.CreateQRCode(req)
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result.Ticket == "" {
          t.Error("expected non-empty ticket")
      }
      if result.ExpireSeconds != 604800 {
          t.Errorf("expected 604800, got %d", result.ExpireSeconds)
      }
  }

  func TestCreateQRCode_NetworkError(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
      srv.Close()
      c := newBaseTestClient(t, srv)
      _, err := c.CreateQRCode(&CreateQRCodeRequest{})
      if err == nil {
          t.Error("expected network error")
      }
  }

  func TestGetQRCodeURL(t *testing.T) {
      // GetQRCodeURL is a pure function — no HTTP needed.
      c := &Client{}
      ticket := "gQH47joAAA+special chars&="
      got := c.GetQRCodeURL(ticket)
      if !strings.HasPrefix(got, "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=") {
          t.Errorf("unexpected URL prefix: %s", got)
      }
      if strings.Contains(got, " ") {
          t.Error("URL must not contain raw spaces")
      }
      if strings.Contains(got, "&=") {
          t.Error("URL must have ticket encoded")
      }
  }
  ```

- [ ] **7.2** Run:

  ```bash
  go test ./offiaccount/... -v -run "TestGetCallbackIp|TestGetApiDomainIP|TestCreateQRCode|TestGetQRCodeURL" -count=1
  ```

  Expected: all tests PASS.

- [ ] **7.3** Commit:

  ```bash
  git add offiaccount/api.base_test.go
  git commit -m "test(offiaccount): add base IP list and QR code API tests"
  ```

---

## Task 8: Test offiaccount analytics APIs

**File:** `offiaccount/api.analysis_test.go` (new file)

Covers: `GetUserSummary`, `GetUserCumulate` from `api.we-data.user.go`.

- [ ] **8.1** Create `offiaccount/api.analysis_test.go`:

  ```go
  package offiaccount

  import (
      "context"
      "net/http"
      "net/http/httptest"
      "testing"
      "time"

      "github.com/godrealms/go-wechat-sdk/utils"
  )

  func newAnalysisTestClient(t *testing.T, srv *httptest.Server) *Client {
      t.Helper()
      h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
      return NewClient(context.Background(), &Config{AppId: "test", AppSecret: "secret"},
          WithHTTPClient(h),
          WithTokenSource(fixedToken{"FAKE_TOKEN"}),
      )
  }

  func TestGetUserSummary_Success(t *testing.T) {
      body := `{"list":[{"ref_date":"2024-01-01","user_source":0,"new_user":100,"cancel_user":5}]}`
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          if r.Method != "POST" {
              t.Errorf("expected POST, got %s", r.Method)
          }
          w.Header().Set("Content-Type", "application/json")
          w.WriteHeader(200)
          _, _ = w.Write([]byte(body))
      }))
      defer srv.Close()

      c := newAnalysisTestClient(t, srv)
      result, err := c.GetUserSummary("2024-01-01", "2024-01-07")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result == nil {
          t.Error("expected non-nil result")
      }
  }

  func TestGetUserSummary_NetworkError(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
      srv.Close()
      c := newAnalysisTestClient(t, srv)
      _, err := c.GetUserSummary("2024-01-01", "2024-01-07")
      if err == nil {
          t.Error("expected network error")
      }
  }

  func TestGetUserSummary_InvalidJSON(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`not-json`))
      }))
      defer srv.Close()

      c := newAnalysisTestClient(t, srv)
      // The method currently does not check errcode and returns the raw struct;
      // but unmarshal failure from HTTP layer should propagate.
      // With GetUserSummaryResult being a struct, invalid JSON causes unmarshal error.
      _, err := c.GetUserSummary("2024-01-01", "2024-01-07")
      if err == nil {
          t.Error("expected unmarshal error for invalid JSON")
      }
  }

  func TestGetUserCumulate_Success(t *testing.T) {
      body := `{"list":[{"ref_date":"2024-01-01","cumulate_user":5000}]}`
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          w.Header().Set("Content-Type", "application/json")
          w.WriteHeader(200)
          _, _ = w.Write([]byte(body))
      }))
      defer srv.Close()

      c := newAnalysisTestClient(t, srv)
      result, err := c.GetUserCumulate("2024-01-01", "2024-01-07")
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result == nil {
          t.Error("expected non-nil result")
      }
  }

  func TestGetUserCumulate_NetworkError(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
      srv.Close()
      c := newAnalysisTestClient(t, srv)
      _, err := c.GetUserCumulate("2024-01-01", "2024-01-07")
      if err == nil {
          t.Error("expected network error")
      }
  }
  ```

  **Note on `GetUserSummaryResult`:** Verify the struct fields in `struct.base.go` around line 242. The `List` field name may be `List` or `UserSummaryList` — adjust the body JSON key if the struct uses a different tag.

  ```bash
  grep -A10 "type GetUserSummaryResult" offiaccount/struct.base.go
  grep -A5 "type GetUserCumulateResult" offiaccount/struct.base.go
  ```

- [ ] **8.2** Run:

  ```bash
  go test ./offiaccount/... -v -run "TestGetUserSummary|TestGetUserCumulate" -count=1
  ```

  Expected: all tests PASS.

- [ ] **8.3** Commit:

  ```bash
  git add offiaccount/api.analysis_test.go
  git commit -m "test(offiaccount): add analytics API tests (GetUserSummary, GetUserCumulate)"
  ```

---

## Task 9: Test merchant/developed/types serialization

**File:** `merchant/developed/types/types_test.go` (new file)

Covers: `MchID.ToString()`, `Transactions.ToString()`, `Refunds.ToString()`, `AbnormalRefund.ToString()`, `TradeBillQuest.ToUrlValues()`, `FundsBillQuest.ToUrlValues()`, `Notify.IsPaymentSuccess()`.

- [ ] **9.1** Create `merchant/developed/types/types_test.go`:

  ```go
  package types

  import (
      "encoding/json"
      "strings"
      "testing"
  )

  // --- MchID ---

  func TestMchID_ToString_WithValue(t *testing.T) {
      m := &MchID{Mchid: "1234567890"}
      s := m.ToString()
      if !strings.Contains(s, "1234567890") {
          t.Errorf("expected mchid in output, got: %s", s)
      }
      var roundtrip MchID
      if err := json.Unmarshal([]byte(s), &roundtrip); err != nil {
          t.Errorf("ToString() produced invalid JSON: %v", err)
      }
      if roundtrip.Mchid != "1234567890" {
          t.Errorf("roundtrip mchid mismatch: %q", roundtrip.Mchid)
      }
  }

  func TestMchID_ToString_Empty(t *testing.T) {
      m := &MchID{}
      s := m.ToString()
      if s == "" {
          t.Error("ToString() on empty struct should not return empty string")
      }
      var roundtrip MchID
      if err := json.Unmarshal([]byte(s), &roundtrip); err != nil {
          t.Errorf("ToString() produced invalid JSON: %v", err)
      }
  }

  // --- Transactions ---

  func TestTransactions_ToString_WithFields(t *testing.T) {
      tx := &Transactions{
          Appid:       "wx1234567890",
          Mchid:       "1234567890",
          Description: "Test Order",
          OutTradeNo:  "ORD20240101001",
          NotifyUrl:   "https://example.com/notify",
          Amount: &Amount{
              Total:    100,
              Currency: "CNY",
          },
      }
      s := tx.ToString()
      if !strings.Contains(s, "wx1234567890") {
          t.Errorf("expected appid in output, got: %s", s)
      }
      if !strings.Contains(s, "Test Order") {
          t.Errorf("expected description in output, got: %s", s)
      }
      var roundtrip Transactions
      if err := json.Unmarshal([]byte(s), &roundtrip); err != nil {
          t.Errorf("ToString() produced invalid JSON: %v", err)
      }
      if roundtrip.OutTradeNo != "ORD20240101001" {
          t.Errorf("roundtrip OutTradeNo mismatch: %q", roundtrip.OutTradeNo)
      }
      if roundtrip.Amount == nil || roundtrip.Amount.Total != 100 {
          t.Error("roundtrip Amount mismatch")
      }
  }

  func TestTransactions_ToString_Empty(t *testing.T) {
      tx := &Transactions{}
      s := tx.ToString()
      var roundtrip Transactions
      if err := json.Unmarshal([]byte(s), &roundtrip); err != nil {
          t.Errorf("ToString() produced invalid JSON: %v", err)
      }
  }

  func TestTransactions_ToString_WithOptionalFields(t *testing.T) {
      tx := &Transactions{
          Appid:       "wx9999",
          Mchid:       "9999",
          Description: "Optional fields test",
          OutTradeNo:  "ORD002",
          NotifyUrl:   "https://example.com/cb",
          TimeExpire:  "2026-12-31T23:59:59+08:00",
          GoodsTag:    "WXG",
          Amount:      &Amount{Total: 500, Currency: "CNY"},
          Detail: &Detail{
              CostPrice: 500,
              GoodsDetail: []*GoodsDetail{
                  {MerchantGoodsId: "G001", Quantity: 1, UnitPrice: 500},
              },
          },
          SceneInfo: &SceneInfo{PayerClientIp: "1.2.3.4"},
          SettleInfo: &SettleInfo{ProfitSharing: true},
      }
      s := tx.ToString()
      var roundtrip Transactions
      if err := json.Unmarshal([]byte(s), &roundtrip); err != nil {
          t.Errorf("ToString() produced invalid JSON: %v", err)
      }
      if roundtrip.GoodsTag != "WXG" {
          t.Errorf("roundtrip GoodsTag mismatch: %q", roundtrip.GoodsTag)
      }
  }

  // --- Refunds ---

  func TestRefunds_ToString_WithFields(t *testing.T) {
      r := &Refunds{
          TransactionId: "txn001",
          OutTradeNo:    "ORD001",
          OutRefundNo:   "REF001",
          Reason:        "Customer request",
          NotifyUrl:     "https://example.com/refund",
          FundsAccount:  "AVAILABLE",
          Amount:        &Amount{Refund: 100, Currency: "CNY"},
      }
      s := r.ToString()
      if !strings.Contains(s, "REF001") {
          t.Errorf("expected OutRefundNo in output, got: %s", s)
      }
      var roundtrip Refunds
      if err := json.Unmarshal([]byte(s), &roundtrip); err != nil {
          t.Errorf("ToString() produced invalid JSON: %v", err)
      }
      if roundtrip.OutRefundNo != "REF001" {
          t.Errorf("roundtrip OutRefundNo mismatch")
      }
  }

  func TestRefunds_ToString_Empty(t *testing.T) {
      r := &Refunds{}
      s := r.ToString()
      var roundtrip Refunds
      if err := json.Unmarshal([]byte(s), &roundtrip); err != nil {
          t.Errorf("ToString() produced invalid JSON: %v", err)
      }
  }

  // --- AbnormalRefund ---

  func TestAbnormalRefund_ToString(t *testing.T) {
      a := &AbnormalRefund{
          OutRefundNo: "REF002",
          Type:        "USER_BANK_CARD",
          BankType:    "CMB",
          BankAccount: "6225xxxxxxxxxxxx",
          RealName:    "Zhang San",
      }
      s := a.ToString()
      if !strings.Contains(s, "REF002") {
          t.Errorf("expected OutRefundNo in output, got: %s", s)
      }
      var roundtrip AbnormalRefund
      if err := json.Unmarshal([]byte(s), &roundtrip); err != nil {
          t.Errorf("ToString() produced invalid JSON: %v", err)
      }
      if roundtrip.Type != "USER_BANK_CARD" {
          t.Errorf("roundtrip Type mismatch")
      }
  }

  func TestAbnormalRefund_ToString_Empty(t *testing.T) {
      a := &AbnormalRefund{}
      s := a.ToString()
      var out AbnormalRefund
      if err := json.Unmarshal([]byte(s), &out); err != nil {
          t.Errorf("invalid JSON: %v", err)
      }
  }

  // --- TradeBillQuest.ToUrlValues ---

  func TestTradeBillQuest_ToUrlValues_AllFields(t *testing.T) {
      q := &TradeBillQuest{
          BillDate: "2024-01-15",
          BillType: "SUCCESS",
          TarType:  "GZIP",
      }
      vals := q.ToUrlValues()
      if vals.Get("bill_date") != "2024-01-15" {
          t.Errorf("expected bill_date=2024-01-15, got %q", vals.Get("bill_date"))
      }
      if vals.Get("bill_type") != "SUCCESS" {
          t.Errorf("expected bill_type=SUCCESS, got %q", vals.Get("bill_type"))
      }
      if vals.Get("tar_type") != "GZIP" {
          t.Errorf("expected tar_type=GZIP, got %q", vals.Get("tar_type"))
      }
  }

  func TestTradeBillQuest_ToUrlValues_OptionalFieldsOmitted(t *testing.T) {
      q := &TradeBillQuest{BillDate: "2024-01-15"}
      vals := q.ToUrlValues()
      if vals.Get("bill_type") != "" {
          t.Errorf("expected empty bill_type, got %q", vals.Get("bill_type"))
      }
      if vals.Get("tar_type") != "" {
          t.Errorf("expected empty tar_type, got %q", vals.Get("tar_type"))
      }
  }

  func TestTradeBillQuest_ToUrlValues_EmptyBillDateUsesToday(t *testing.T) {
      q := &TradeBillQuest{}
      vals := q.ToUrlValues()
      if vals.Get("bill_date") == "" {
          t.Error("expected bill_date to default to today when empty")
      }
  }

  // --- FundsBillQuest.ToUrlValues ---

  func TestFundsBillQuest_ToUrlValues_AllFields(t *testing.T) {
      q := &FundsBillQuest{
          BillDate:    "2024-01-15",
          AccountType: "BASIC",
          TarType:     "GZIP",
      }
      vals := q.ToUrlValues()
      if vals.Get("bill_date") != "2024-01-15" {
          t.Errorf("expected bill_date=2024-01-15, got %q", vals.Get("bill_date"))
      }
      if vals.Get("account_type") != "BASIC" {
          t.Errorf("expected account_type=BASIC, got %q", vals.Get("account_type"))
      }
  }

  func TestFundsBillQuest_ToUrlValues_EmptyDateUsesToday(t *testing.T) {
      q := &FundsBillQuest{}
      vals := q.ToUrlValues()
      if vals.Get("bill_date") == "" {
          t.Error("expected bill_date to default to today when empty")
      }
  }

  // --- Notify.IsPaymentSuccess ---

  func TestNotify_IsPaymentSuccess_True(t *testing.T) {
      n := &Notify{EventType: "TRANSACTION.SUCCESS"}
      if !n.IsPaymentSuccess() {
          t.Error("expected IsPaymentSuccess to return true")
      }
  }

  func TestNotify_IsPaymentSuccess_False(t *testing.T) {
      cases := []string{"", "TRANSACTION.REFUND", "REFUND.SUCCESS"}
      for _, et := range cases {
          n := &Notify{EventType: et}
          if n.IsPaymentSuccess() {
              t.Errorf("expected IsPaymentSuccess=false for EventType=%q", et)
          }
      }
  }
  ```

- [ ] **9.2** Run:

  ```bash
  go test ./merchant/developed/types/... -v -count=1
  ```

  Expected output:
  ```
  --- PASS: TestMchID_ToString_WithValue (0.00s)
  --- PASS: TestMchID_ToString_Empty (0.00s)
  --- PASS: TestTransactions_ToString_WithFields (0.00s)
  --- PASS: TestTransactions_ToString_Empty (0.00s)
  --- PASS: TestTransactions_ToString_WithOptionalFields (0.00s)
  --- PASS: TestRefunds_ToString_WithFields (0.00s)
  --- PASS: TestRefunds_ToString_Empty (0.00s)
  --- PASS: TestAbnormalRefund_ToString (0.00s)
  --- PASS: TestAbnormalRefund_ToString_Empty (0.00s)
  --- PASS: TestTradeBillQuest_ToUrlValues_AllFields (0.00s)
  --- PASS: TestTradeBillQuest_ToUrlValues_OptionalFieldsOmitted (0.00s)
  --- PASS: TestTradeBillQuest_ToUrlValues_EmptyBillDateUsesToday (0.00s)
  --- PASS: TestFundsBillQuest_ToUrlValues_AllFields (0.00s)
  --- PASS: TestFundsBillQuest_ToUrlValues_EmptyDateUsesToday (0.00s)
  --- PASS: TestNotify_IsPaymentSuccess_True (0.00s)
  --- PASS: TestNotify_IsPaymentSuccess_False (0.00s)
  PASS
  ok      github.com/godrealms/go-wechat-sdk/merchant/developed/types
  ```

- [ ] **9.3** Verify coverage:

  ```bash
  go test ./merchant/developed/types/... -cover -count=1
  ```

  Expected: `coverage: 70%+`.

- [ ] **9.4** Commit:

  ```bash
  git add merchant/developed/types/types_test.go
  git commit -m "test(merchant/types): add serialization and utility method tests for all types"
  ```

---

## Task 10: Test merchant/developed payment methods

**File:** `merchant/developed/merchant_test.go` (new file)

Covers: `TransactionsApp`, `TransactionsJsapi`, `TransactionsNative`, `TransactionsH5` from the merchant `wechat` package. These methods use RSA signing — tests use a generated RSA key so no real credentials are needed.

- [ ] **10.1** Create `merchant/developed/merchant_test.go`:

  ```go
  package wechat

  import (
      "context"
      "crypto/rand"
      "crypto/rsa"
      "encoding/json"
      "io"
      "net/http"
      "net/http/httptest"
      "strings"
      "testing"
      "time"

      "github.com/godrealms/go-wechat-sdk/merchant/developed/types"
      "github.com/godrealms/go-wechat-sdk/utils"
  )

  // generateTestKey creates a 2048-bit RSA key for test signing.
  func generateTestKey(t *testing.T) *rsa.PrivateKey {
      t.Helper()
      key, err := rsa.GenerateKey(rand.Reader, 2048)
      if err != nil {
          t.Fatalf("failed to generate RSA key: %v", err)
      }
      return key
  }

  // newTestMerchantClient creates a Client with a test httptest server and generated keys.
  func newTestMerchantClient(t *testing.T, srv *httptest.Server) *Client {
      t.Helper()
      key := generateTestKey(t)
      h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
      return NewWechatClient().
          WithAppid("wx_test_appid").
          WithMchid("1234567890").
          WithCertificateNumber("CERT_SERIAL_NO").
          WithAPIv3Key("abcdefgh12345678abcdefgh12345678").
          WithPrivateKey(key).
          WithHttp(h)
  }

  // testOrder builds a minimal Transactions order for test requests.
  func testOrder() *types.Transactions {
      return &types.Transactions{
          Appid:       "wx_test_appid",
          Mchid:       "1234567890",
          Description: "Test product",
          OutTradeNo:  "ORD20240101001",
          NotifyUrl:   "https://example.com/notify",
          Amount:      &types.Amount{Total: 100, Currency: "CNY"},
      }
  }

  // captureServer returns a test server that captures the request body and replies with the given JSON.
  func captureServer(t *testing.T, reply string) (*httptest.Server, *[]byte) {
      t.Helper()
      var captured []byte
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          captured, _ = io.ReadAll(r.Body)
          w.Header().Set("Content-Type", "application/json")
          w.WriteHeader(200)
          _, _ = w.Write([]byte(reply))
      }))
      return srv, &captured
  }

  // --- TransactionsApp ---

  func TestTransactionsApp_Success(t *testing.T) {
      srv, _ := captureServer(t, `{"prepay_id":"wx20240101prepayid"}`)
      defer srv.Close()

      c := newTestMerchantClient(t, srv)
      result, err := c.TransactionsApp(testOrder())
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result.PrepayId != "wx20240101prepayid" {
          t.Errorf("expected prepay_id wx20240101prepayid, got %q", result.PrepayId)
      }
  }

  func TestTransactionsApp_SetsAuthorizationHeader(t *testing.T) {
      var gotAuth string
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
          gotAuth = r.Header.Get("Authorization")
          w.WriteHeader(200)
          _, _ = w.Write([]byte(`{"prepay_id":"pid"}`))
      }))
      defer srv.Close()

      c := newTestMerchantClient(t, srv)
      _, err := c.TransactionsApp(testOrder())
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048") {
          t.Errorf("expected WECHATPAY2-SHA256-RSA2048 Authorization header, got: %q", gotAuth)
      }
      if !strings.Contains(gotAuth, "mchid=\"1234567890\"") {
          t.Errorf("expected mchid in Authorization header, got: %q", gotAuth)
      }
  }

  func TestTransactionsApp_SendsCorrectJSON(t *testing.T) {
      srv, captured := captureServer(t, `{"prepay_id":"pid"}`)
      defer srv.Close()

      c := newTestMerchantClient(t, srv)
      _, err := c.TransactionsApp(testOrder())
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      var body map[string]interface{}
      if err := json.Unmarshal(*captured, &body); err != nil {
          t.Fatalf("server received invalid JSON: %v, body: %s", err, string(*captured))
      }
      if body["out_trade_no"] != "ORD20240101001" {
          t.Errorf("expected out_trade_no, got: %v", body["out_trade_no"])
      }
  }

  func TestTransactionsApp_NetworkError(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
      u := srv.URL
      srv.Close()

      key := generateTestKey(t)
      h := utils.NewHTTP(u, utils.WithTimeout(time.Second))
      c := NewWechatClient().WithPrivateKey(key).WithHttp(h)
      _, err := c.TransactionsApp(testOrder())
      if err == nil {
          t.Error("expected network error")
      }
  }

  // --- ModifyTransactionsApp ---

  func TestModifyTransactionsApp_Success(t *testing.T) {
      srv, _ := captureServer(t, `{"prepay_id":"wx_prepay_001"}`)
      defer srv.Close()

      c := newTestMerchantClient(t, srv)
      result, err := c.ModifyTransactionsApp(testOrder())
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result.PrepayId != "wx_prepay_001" {
          t.Errorf("expected prepay_id wx_prepay_001, got %q", result.PrepayId)
      }
      if result.PackageValue != "Sign=WXPay" {
          t.Errorf("expected package=Sign=WXPay, got %q", result.PackageValue)
      }
      if result.Sign == "" {
          t.Error("expected non-empty Sign from ModifyTransactionsApp")
      }
  }

  // --- TransactionsJsapi ---

  func TestTransactionsJsapi_Success(t *testing.T) {
      srv, _ := captureServer(t, `{"prepay_id":"jsapi_prepay_id_001"}`)
      defer srv.Close()

      c := newTestMerchantClient(t, srv)
      result, err := c.TransactionsJsapi(testOrder())
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result.PrepayId != "jsapi_prepay_id_001" {
          t.Errorf("expected jsapi prepay_id, got %q", result.PrepayId)
      }
  }

  func TestTransactionsJsapi_NetworkError(t *testing.T) {
      srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
      u := srv.URL
      srv.Close()

      key := generateTestKey(t)
      h := utils.NewHTTP(u, utils.WithTimeout(time.Second))
      c := NewWechatClient().WithPrivateKey(key).WithHttp(h)
      _, err := c.TransactionsJsapi(testOrder())
      if err == nil {
          t.Error("expected network error")
      }
  }

  func TestModifyTransactionsJsapi_Success(t *testing.T) {
      srv, _ := captureServer(t, `{"prepay_id":"jsapi_pid_002"}`)
      defer srv.Close()

      c := newTestMerchantClient(t, srv)
      result, err := c.ModifyTransactionsJsapi(testOrder())
      if err != nil {
          t.Fatalf("unexpected error: %v", err)
      }
      if result.SignType != "RSA" {
          t.Errorf("expected SignType=RSA, got %q", result.SignType)
      }
      if !strings.HasPrefix(result.Package, "prepay_id=") {
          t.Errorf("expected Package to start with prepay_id=, got %q", result.Package)
      }
      if result.PaySign == "" {
          t.Error("expected non-empty PaySign from ModifyTransactionsJsapi")
      }
  }

  // --- HTTP injection verifications ---

  func TestNewWechatClient_DefaultBaseURL(t *testing.T) {
      c := NewWechatClient()
      if c.Http == nil {
          t.Fatal("expected non-nil Http")
      }
      if !strings.Contains(c.Http.BaseURL, "api.mch.weixin.qq.com") {
          t.Errorf("expected default base URL to contain api.mch.weixin.qq.com, got %q", c.Http.BaseURL)
      }
  }

  func TestWechatClient_WithHttp_ReplacesHTTP(t *testing.T) {
      c := NewWechatClient()
      h := utils.NewHTTP("http://custom.example.com")
      c = c.WithHttp(h)
      if c.Http.BaseURL != "http://custom.example.com" {
          t.Errorf("expected custom base URL, got %q", c.Http.BaseURL)
      }
  }

  func TestWechatClient_BuilderChain(t *testing.T) {
      key := generateTestKey(t)
      c := NewWechatClient().
          WithAppid("appid").
          WithMchid("mchid").
          WithCertificateNumber("cert_no").
          WithAPIv3Key("key32byteslongkey32byteslongkey!").
          WithPrivateKey(key)
      if c.Appid != "appid" {
          t.Errorf("expected appid, got %q", c.Appid)
      }
      if c.Mchid != "mchid" {
          t.Errorf("expected mchid, got %q", c.Mchid)
      }
      if c.CertificateNumber != "cert_no" {
          t.Errorf("expected cert_no, got %q", c.CertificateNumber)
      }
  }

  // --- Context usage ---

  func TestTransactionsApp_UsesContextBackground(t *testing.T) {
      // Verifies the call completes in reasonable time (context.Background, no cancel).
      srv, _ := captureServer(t, `{"prepay_id":"ctx_test"}`)
      defer srv.Close()

      done := make(chan error, 1)
      go func() {
          c := newTestMerchantClient(t, srv)
          _, err := c.TransactionsApp(testOrder())
          done <- err
      }()

      select {
      case err := <-done:
          if err != nil {
              t.Fatalf("unexpected error: %v", err)
          }
      case <-time.After(5 * time.Second):
          t.Fatal("TransactionsApp did not complete within 5 seconds")
      }
  }
  ```

  **Note on `TransactionsNative` and `TransactionsH5`:** Check if these methods exist in the current `merchant/developed` package:

  ```bash
  grep -rn "func.*TransactionsNative\|func.*TransactionsH5" merchant/developed/
  ```

  If they exist, add corresponding tests following the same pattern as `TransactionsApp`. If they are absent (not yet implemented), skip them and note this as a coverage gap — the task is complete without them.

- [ ] **10.2** Check for TransactionsNative and TransactionsH5, add tests if they exist:

  ```bash
  grep -rn "func (c \*Client) Transactions" /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig/merchant/developed/
  ```

  For each method found (e.g. `TransactionsNative` returning `*types.TransactionsNativeResp`, `TransactionsH5` returning `*types.TransactionsH5Resp`), add a test block of the same shape as `TestTransactionsApp_Success`, using the appropriate response JSON:
  - Native: `{"code_url":"weixin://wxpay/bizpayurl?pr=xxxx"}`
  - H5: `{"h5_url":"https://wx.tenpay.com/cgi-bin/mmpayweb-bin/checkmweb?prepay_id=wx&package=1298847324"}`

- [ ] **10.3** Run:

  ```bash
  go test ./merchant/developed/... -v -count=1
  ```

  Expected: all tests PASS. If `TransactionsNative`/`TransactionsH5` are absent, all the tests added in step 10.1 still PASS.

- [ ] **10.4** Verify coverage:

  ```bash
  go test ./merchant/developed/... -cover -count=1
  ```

  Expected: `coverage: 70%+` on the `merchant/developed` package.

- [ ] **10.5** Commit:

  ```bash
  git add merchant/developed/merchant_test.go
  git commit -m "test(merchant): add payment API tests with httptest mock and generated RSA keys"
  ```

---

## Final verification

- [ ] **F.1** Run the full test suite for all three sub-packages and confirm zero failures:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go test ./utils/... ./offiaccount/... ./merchant/developed/... ./merchant/developed/types/... -count=1
  ```

  Expected:
  ```
  ok      github.com/godrealms/go-wechat-sdk/utils
  ok      github.com/godrealms/go-wechat-sdk/offiaccount
  ok      github.com/godrealms/go-wechat-sdk/merchant/developed
  ok      github.com/godrealms/go-wechat-sdk/merchant/developed/types
  ```

- [ ] **F.2** Run with `-cover` to confirm targets are met:

  ```bash
  go test ./utils/... -cover -count=1
  go test ./offiaccount/... -cover -count=1
  go test ./merchant/developed/types/... -cover -count=1
  go test ./merchant/developed/... -cover -count=1
  ```

  Expected coverage:
  | Package | Target | Expected |
  |---------|--------|----------|
  | `utils` | 80% | 80%+ |
  | `offiaccount` | 70% | 70%+ |
  | `merchant/developed/types` | 70% | 70%+ |
  | `merchant/developed` | 70% | 70%+ |

- [ ] **F.3** Confirm no real WeChat API calls were made (all tests use `httptest.Server`):

  ```bash
  grep -rn "api.weixin.qq.com\|api.mch.weixin.qq.com" \
    utils/http_test.go \
    offiaccount/api.menu_test.go \
    offiaccount/api.message_test.go \
    offiaccount/api.user_test.go \
    offiaccount/api.material_test.go \
    offiaccount/api.base_test.go \
    offiaccount/api.analysis_test.go \
    merchant/developed/types/types_test.go \
    merchant/developed/merchant_test.go
  ```

  Expected: no output (zero matches).

- [ ] **F.4** Final commit (if any loose files remain unstaged):

  ```bash
  git status
  # stage and commit any remaining test files not yet committed
  ```

---

## Troubleshooting guide

**Compile error: `fixedToken redeclared`**
Each `_test.go` in the `offiaccount` package that declares `fixedToken` will conflict. Since they share the same `package offiaccount` test package, declare `fixedToken` only once. Consolidate it into a single file named `offiaccount/testhelpers_test.go`:

```go
package offiaccount

import (
    "context"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/godrealms/go-wechat-sdk/utils"
)

type fixedToken struct{ tok string }

func (f fixedToken) AccessToken(_ context.Context) (string, error) { return f.tok, nil }

func newOffiaccountTestClient(t *testing.T, srv *httptest.Server) *Client {
    t.Helper()
    h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
    return NewClient(context.Background(), &Config{AppId: "test", AppSecret: "secret"},
        WithHTTPClient(h),
        WithTokenSource(fixedToken{"FAKE_TOKEN"}),
    )
}
```

Then replace all `newMenuTestClient`, `newMsgTestClient`, `newUserTestClient`, `newMaterialTestClient`, `newBaseTestClient`, `newAnalysisTestClient` with calls to `newOffiaccountTestClient` and remove duplicate `fixedToken` declarations from each test file.

**`ConditionalMenu` undefined**
Check `struct.menu.go` for the struct name:
```bash
grep -n "type.*Menu\|type.*Conditional" offiaccount/struct.menu.go
```
Use the exact type name found.

**`DeleteMassMsgRequest.MsgId` wrong type**
Run `grep -A5 "type DeleteMassMsgRequest" offiaccount/struct.massmsg.go` and adjust the test to match the actual field type.

**RSA signing failure in merchant tests**
The `utils.SignSHA256WithRSA` function requires a valid `*rsa.PrivateKey`. The `generateTestKey` helper in Task 10 always generates a fresh 2048-bit key. If you see crypto errors, ensure `generateTestKey` is called before the client is used.

**Coverage below target**
Run `go test -coverprofile=cov.out ./...` then `go tool cover -func=cov.out` to see per-function coverage. Focus on untested methods — each uncovered function is a missed test case. Add additional table rows for edge cases (empty inputs, errcode != 0, non-2xx HTTP status).
