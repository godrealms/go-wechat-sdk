# Sub-project A: Code Quality Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix real bugs and unify error handling patterns across all packages in the go-wechat-sdk project.

**Architecture:** Add a `WechatAPIError` interface to `utils/wechat_error.go`; each existing package adds `Code() int` and `Message() string` methods to its own error type to implement it; fix JSON marshal error-ignoring bugs in `merchant/developed/types`; apply consistent errcode checking via `CheckResp()` across all `offiaccount` API files.

**Tech Stack:** Go 1.23.1, standard library only

---

## Current State Audit

Before implementing, note these pre-existing build errors that some tasks below must resolve:

- `offiaccount` does not compile: `WeixinError` and `CheckResp` are used in `client.go` and tests but not yet defined in the package. `AccessToken` struct lacks `ErrCode`/`ErrMsg` fields. `client.go:GetAccessToken` duplicates `api.base.go:GetAccessToken`. Task 4 fixes these.
- `mini-program` does not compile: `utils.HTTP.DoRequestWithRawResponse` is missing. This is outside the scope of sub-project A; do not fix it here — skip `mini-program` package-level `go test` commands to avoid noise.
- `oplatform`, `channels`, `mini-game`, `merchant/developed` all compile cleanly at the start of this plan.
- `work-wechat/isv` compiles and all tests pass.

---

## Task 1: Add `WechatAPIError` interface to `utils`

**Purpose:** Provide a shared interface that all package-specific error types implement, enabling callers to inspect error codes generically.

**Files:**
- Create `utils/wechat_error.go`
- Create `utils/wechat_error_test.go`

### Steps

- [ ] **1.1** Write the failing test first.

  Create `utils/wechat_error_test.go`:

  ```go
  package utils_test

  import (
  	"testing"

  	"github.com/godrealms/go-wechat-sdk/utils"
  )

  // fakeAPIError is a local type used to verify the interface contract.
  type fakeAPIError struct {
  	code int
  	msg  string
  }

  func (f *fakeAPIError) Error() string   { return f.msg }
  func (f *fakeAPIError) Code() int       { return f.code }
  func (f *fakeAPIError) Message() string { return f.msg }

  // Compile-time assertion: fakeAPIError satisfies WechatAPIError.
  var _ utils.WechatAPIError = (*fakeAPIError)(nil)

  func TestWechatAPIError_Interface(t *testing.T) {
  	tests := []struct {
  		name    string
  		code    int
  		message string
  	}{
  		{"zero errcode", 0, "ok"},
  		{"common auth error", 40001, "invalid credential"},
  		{"rate limit", 45009, "api freq out of limit"},
  	}
  	for _, tt := range tests {
  		t.Run(tt.name, func(t *testing.T) {
  			var e utils.WechatAPIError = &fakeAPIError{code: tt.code, msg: tt.message}
  			if got := e.Code(); got != tt.code {
  				t.Errorf("Code() = %d, want %d", got, tt.code)
  			}
  			if got := e.Message(); got != tt.message {
  				t.Errorf("Message() = %q, want %q", got, tt.message)
  			}
  			// WechatAPIError embeds error; Error() must be callable.
  			if got := e.Error(); got == "" && tt.message != "" {
  				t.Errorf("Error() returned empty string, want %q", tt.message)
  			}
  		})
  	}
  }

  func TestWechatAPIError_NilSafety(t *testing.T) {
  	// A nil *fakeAPIError should not satisfy a non-nil interface value.
  	var f *fakeAPIError
  	var e utils.WechatAPIError = f
  	// Interface value is non-nil (has type), but underlying pointer is nil.
  	// Calling methods on a nil pointer panics unless the method handles it.
  	// This test verifies the interface itself is declared correctly.
  	if e == nil {
  		t.Error("interface value should not be nil when set to a typed nil")
  	}
  }
  ```

- [ ] **1.2** Run the test; confirm it fails to compile (interface not yet defined):

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go test ./utils/ 2>&1
  # Expected: build error "undefined: utils.WechatAPIError"
  ```

- [ ] **1.3** Implement the interface.

  Create `utils/wechat_error.go`:

  ```go
  package utils

  // WechatAPIError is implemented by every package-specific WeChat API error type
  // in this SDK. It allows callers to inspect the numeric error code and human-
  // readable message without importing a concrete package.
  //
  // Usage:
  //
  //	var apiErr utils.WechatAPIError
  //	if errors.As(err, &apiErr) {
  //	    log.Printf("WeChat API error %d: %s", apiErr.Code(), apiErr.Message())
  //	}
  type WechatAPIError interface {
  	error
  	// Code returns the numeric errcode returned by the WeChat API (e.g. 40001).
  	Code() int
  	// Message returns the human-readable errmsg returned by the WeChat API.
  	Message() string
  }
  ```

- [ ] **1.4** Run the test; confirm it passes:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go test ./utils/ -v -run TestWechatAPIError
  # Expected output:
  # === RUN   TestWechatAPIError_Interface
  # === RUN   TestWechatAPIError_Interface/zero_errcode
  # === RUN   TestWechatAPIError_Interface/common_auth_error
  # === RUN   TestWechatAPIError_Interface/rate_limit
  # --- PASS: TestWechatAPIError_Interface (0.00s)
  # === RUN   TestWechatAPIError_NilSafety
  # --- PASS: TestWechatAPIError_NilSafety (0.00s)
  # PASS
  ```

- [ ] **1.5** Commit:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  git add utils/wechat_error.go utils/wechat_error_test.go
  git commit -m "feat(utils): add WechatAPIError interface"
  ```

---

## Task 2: Implement `WechatAPIError` on `oplatform.WeixinError`

**Purpose:** The `oplatform` package has a complete `WeixinError` struct. Add `Code()` and `Message()` methods so it satisfies the new interface.

**Files:**
- Modify `oplatform/errors.go`
- Create `oplatform/errors_test.go`

### Steps

- [ ] **2.1** Write the failing test first.

  Create `oplatform/errors_test.go`:

  ```go
  package oplatform_test

  import (
  	"errors"
  	"testing"

  	"github.com/godrealms/go-wechat-sdk/oplatform"
  	"github.com/godrealms/go-wechat-sdk/utils"
  )

  // Compile-time assertion: *oplatform.WeixinError satisfies utils.WechatAPIError.
  var _ utils.WechatAPIError = (*oplatform.WeixinError)(nil)

  func TestWeixinError_Implements_WechatAPIError(t *testing.T) {
  	tests := []struct {
  		name    string
  		errcode int
  		errmsg  string
  	}{
  		{"auth failure", 40001, "invalid credential"},
  		{"access token expired", 42001, "access_token expired"},
  		{"api unauthorized", 48001, "api unauthorized"},
  	}
  	for _, tt := range tests {
  		t.Run(tt.name, func(t *testing.T) {
  			e := &oplatform.WeixinError{ErrCode: tt.errcode, ErrMsg: tt.errmsg}

  			// Interface satisfaction via errors.As.
  			var apiErr utils.WechatAPIError
  			wrapped := fmt.Errorf("wrapped: %w", e)
  			if !errors.As(wrapped, &apiErr) {
  				t.Fatal("errors.As: expected WechatAPIError")
  			}
  			if apiErr.Code() != tt.errcode {
  				t.Errorf("Code() = %d, want %d", apiErr.Code(), tt.errcode)
  			}
  			if apiErr.Message() != tt.errmsg {
  				t.Errorf("Message() = %q, want %q", apiErr.Message(), tt.errmsg)
  			}

  			// Direct method calls.
  			if e.Code() != tt.errcode {
  				t.Errorf("e.Code() = %d, want %d", e.Code(), tt.errcode)
  			}
  			if e.Message() != tt.errmsg {
  				t.Errorf("e.Message() = %q, want %q", e.Message(), tt.errmsg)
  			}
  		})
  	}
  }

  func TestWeixinError_ErrorString(t *testing.T) {
  	e := &oplatform.WeixinError{ErrCode: 40001, ErrMsg: "invalid credential"}
  	got := e.Error()
  	if got != "oplatform: errcode=40001 errmsg=invalid credential" {
  		t.Errorf("Error() = %q, unexpected format", got)
  	}
  }
  ```

  Note: the test file imports `fmt` — add it to the import block:

  ```go
  import (
  	"errors"
  	"fmt"
  	"testing"

  	"github.com/godrealms/go-wechat-sdk/oplatform"
  	"github.com/godrealms/go-wechat-sdk/utils"
  )
  ```

- [ ] **2.2** Run the test; confirm it fails to compile (`Code` and `Message` undefined):

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go build ./oplatform/ 2>&1
  # Will also show offiaccount build errors (pre-existing) but oplatform itself should build.
  # After adding the test file:
  # errors_test.go:X: e.Code undefined
  # errors_test.go:X: e.Message undefined
  ```

- [ ] **2.3** Add the two methods to `oplatform/errors.go`.

  Open `oplatform/errors.go` and append after the existing `Error()` method:

  ```go
  // Code returns the numeric errcode. Implements utils.WechatAPIError.
  func (e *WeixinError) Code() int { return e.ErrCode }

  // Message returns the human-readable errmsg. Implements utils.WechatAPIError.
  func (e *WeixinError) Message() string { return e.ErrMsg }
  ```

  The complete `oplatform/errors.go` after the edit:

  ```go
  package oplatform

  import (
  	"errors"
  	"fmt"
  )

  // WeixinError 微信业务错误 (errcode != 0).
  type WeixinError struct {
  	ErrCode int
  	ErrMsg  string
  }

  func (e *WeixinError) Error() string {
  	if e == nil {
  		return ""
  	}
  	return fmt.Sprintf("oplatform: errcode=%d errmsg=%s", e.ErrCode, e.ErrMsg)
  }

  // Code returns the numeric errcode. Implements utils.WechatAPIError.
  func (e *WeixinError) Code() int { return e.ErrCode }

  // Message returns the human-readable errmsg. Implements utils.WechatAPIError.
  func (e *WeixinError) Message() string { return e.ErrMsg }

  // 常见哨兵错误。
  var (
  	// ErrNotFound 由 Store 实现返回，表示 key 不存在（非 I/O 错误）。
  	ErrNotFound = errors.New("oplatform: not found")

  	// ErrAuthorizerRevoked 当 refresh_token 失效 (errcode=61023) 时返回；
  	// 调用方应删除 Store 中该 authorizer 的记录并重新引导授权。
  	ErrAuthorizerRevoked = errors.New("oplatform: authorizer refresh_token revoked")

  	// ErrVerifyTicketMissing 当 component_access_token 需要刷新但 Store
  	// 尚未收到微信推送的 component_verify_ticket 时返回。
  	ErrVerifyTicketMissing = errors.New("oplatform: component_verify_ticket not yet received")
  )
  ```

- [ ] **2.4** Run the tests (note: `oplatform` test build currently fails due to the broken `offiaccount` dep which is fixed in Task 4; for now build just the oplatform package itself):

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go vet ./oplatform/ 2>&1
  # Expected: only noise from offiaccount dep; oplatform source is clean.
  ```

  After Task 4 fixes `offiaccount`, run the full test:

  ```bash
  go test ./oplatform/ -v -run TestWeixinError 2>&1
  # Expected:
  # === RUN   TestWeixinError_Implements_WechatAPIError
  # --- PASS: TestWeixinError_Implements_WechatAPIError (0.00s)
  # === RUN   TestWeixinError_ErrorString
  # --- PASS: TestWeixinError_ErrorString (0.00s)
  # PASS
  ```

- [ ] **2.5** Commit:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  git add oplatform/errors.go oplatform/errors_test.go
  git commit -m "feat(oplatform): implement WechatAPIError on WeixinError"
  ```

---

## Task 3: Implement `WechatAPIError` on `work-wechat/isv.WeixinError`

**Purpose:** The `work-wechat/isv` package also has its own `WeixinError`. Add the same two methods.

**Files:**
- Modify `work-wechat/isv/errors.go`
- Create `work-wechat/isv/errors_test.go`

### Steps

- [ ] **3.1** Write the failing test first.

  Create `work-wechat/isv/errors_test.go`:

  ```go
  package isv_test

  import (
  	"errors"
  	"fmt"
  	"testing"

  	"github.com/godrealms/go-wechat-sdk/utils"
  	"github.com/godrealms/go-wechat-sdk/work-wechat/isv"
  )

  // Compile-time assertion: *isv.WeixinError satisfies utils.WechatAPIError.
  var _ utils.WechatAPIError = (*isv.WeixinError)(nil)

  func TestWeixinError_Implements_WechatAPIError(t *testing.T) {
  	tests := []struct {
  		name    string
  		errcode int
  		errmsg  string
  	}{
  		{"suite token expired", 42001, "suite_access_token expired"},
  		{"no permission", 60011, "no privilege to access/modify contact"},
  		{"user not exist", 60111, "userid not found"},
  	}
  	for _, tt := range tests {
  		t.Run(tt.name, func(t *testing.T) {
  			e := &isv.WeixinError{ErrCode: tt.errcode, ErrMsg: tt.errmsg}

  			var apiErr utils.WechatAPIError
  			wrapped := fmt.Errorf("wrapped: %w", e)
  			if !errors.As(wrapped, &apiErr) {
  				t.Fatal("errors.As: expected WechatAPIError")
  			}
  			if apiErr.Code() != tt.errcode {
  				t.Errorf("Code() = %d, want %d", apiErr.Code(), tt.errcode)
  			}
  			if apiErr.Message() != tt.errmsg {
  				t.Errorf("Message() = %q, want %q", apiErr.Message(), tt.errmsg)
  			}
  			if e.Code() != tt.errcode {
  				t.Errorf("e.Code() = %d, want %d", e.Code(), tt.errcode)
  			}
  			if e.Message() != tt.errmsg {
  				t.Errorf("e.Message() = %q, want %q", e.Message(), tt.errmsg)
  			}
  		})
  	}
  }

  func TestWeixinError_ErrorString(t *testing.T) {
  	e := &isv.WeixinError{ErrCode: 42001, ErrMsg: "suite_access_token expired"}
  	got := e.Error()
  	want := "isv: weixin error 42001: suite_access_token expired"
  	if got != want {
  		t.Errorf("Error() = %q, want %q", got, want)
  	}
  }
  ```

- [ ] **3.2** Run to confirm test does not compile yet:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go build ./work-wechat/isv/ 2>&1
  # Package builds. Add test, then:
  # go test ./work-wechat/isv/ 2>&1
  # Expected: e.Code undefined, e.Message undefined
  ```

- [ ] **3.3** Add the two methods to `work-wechat/isv/errors.go`.

  The complete `work-wechat/isv/errors.go` after the edit:

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

  // Code returns the numeric errcode. Implements utils.WechatAPIError.
  func (e *WeixinError) Code() int { return e.ErrCode }

  // Message returns the human-readable errmsg. Implements utils.WechatAPIError.
  func (e *WeixinError) Message() string { return e.ErrMsg }
  ```

- [ ] **3.4** Run the tests:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go test ./work-wechat/isv/ -v -run TestWeixinError 2>&1
  # Expected:
  # === RUN   TestWeixinError_Implements_WechatAPIError
  # --- PASS: TestWeixinError_Implements_WechatAPIError (0.00s)
  # === RUN   TestWeixinError_ErrorString
  # --- PASS: TestWeixinError_ErrorString (0.00s)
  # PASS
  ```

- [ ] **3.5** Run the full `work-wechat/isv` suite to confirm no regressions:

  ```bash
  go test ./work-wechat/isv/ 2>&1
  # Expected: ok  github.com/godrealms/go-wechat-sdk/work-wechat/isv
  ```

- [ ] **3.6** Commit:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  git add work-wechat/isv/errors.go work-wechat/isv/errors_test.go
  git commit -m "feat(work-wechat/isv): implement WechatAPIError on WeixinError"
  ```

---

## Task 4: Add `WeixinError`, `CheckResp`, and fix compile errors in `offiaccount`

**Purpose:** The `offiaccount` package currently does not compile because `WeixinError` and `CheckResp` are referenced but never defined, and `AccessToken` lacks `ErrCode`/`ErrMsg`. This task defines a dedicated `offiaccount/errors.go`, fixes `AccessToken`, and makes `WeixinError` implement `WechatAPIError`.

**Files:**
- Create `offiaccount/errors.go`
- Modify `offiaccount/struct.base.go` (add `ErrCode`/`ErrMsg` to `AccessToken`)

**Note:** Do NOT touch `offiaccount/client.go` or `offiaccount/api.base.go` in this task — the pre-existing compile errors there (duplicate `GetAccessToken`, wrong `AccessToken` usage) are caused by the missing types; they will resolve once `WeixinError` and the corrected `AccessToken` are in place. If residual compile errors remain in those files after this task, fix them minimally within this task's commit.

### Steps

- [ ] **4.1** Write a failing test.

  The existing `offiaccount/client_test.go` already contains `TestCheckResp` and `TestClient_AccessTokenE_ReturnsWeixinError`. These serve as the TDD spec. Running tests now:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go test ./offiaccount/ 2>&1
  # Expected: build failures (WeixinError undefined, AccessToken fields missing, etc.)
  ```

- [ ] **4.2** Inspect pre-existing errors in detail:

  ```bash
  go build ./offiaccount/ 2>&1
  # offiaccount/client.go:68:18: method Client.GetAccessToken already declared at offiaccount/api.base.go:13:18
  # offiaccount/client.go:129:12: result.ErrCode undefined (type *AccessToken has no field or method ErrCode)
  # offiaccount/client.go:130:16: undefined: WeixinError
  # offiaccount/client.go:130:44: result.ErrCode undefined (type *AccessToken has no field or method ErrCode)
  # offiaccount/client.go:130:68: result.ErrMsg undefined (type *AccessToken has no field or method ErrMsg)
  # offiaccount/api.base.go:40:18: cannot use result (variable of type *AccessToken) as string value in assignment
  ```

- [ ] **4.3** Create `offiaccount/errors.go` defining `WeixinError` and `CheckResp`:

  ```go
  package offiaccount

  import "fmt"

  // WeixinError represents a WeChat API business error (errcode != 0).
  type WeixinError struct {
  	ErrCode int
  	ErrMsg  string
  }

  func (e *WeixinError) Error() string {
  	return fmt.Sprintf("offiaccount: errcode=%d errmsg=%s", e.ErrCode, e.ErrMsg)
  }

  // Code returns the numeric errcode. Implements utils.WechatAPIError.
  func (e *WeixinError) Code() int { return e.ErrCode }

  // Message returns the human-readable errmsg. Implements utils.WechatAPIError.
  func (e *WeixinError) Message() string { return e.ErrMsg }

  // CheckResp returns a *WeixinError if r.ErrCode != 0, otherwise nil.
  // Use this after every WeChat API call to normalise error handling.
  func CheckResp(r *Resp) error {
  	if r.ErrCode == 0 {
  		return nil
  	}
  	return &WeixinError{ErrCode: r.ErrCode, ErrMsg: r.ErrMsg}
  }
  ```

- [ ] **4.4** Fix `AccessToken` in `offiaccount/struct.base.go` to include `ErrCode` and `ErrMsg` fields (required by `client.go:refreshAccessToken`):

  Locate the existing `AccessToken` struct (around line 929) in `offiaccount/struct.base.go`:

  ```go
  // BEFORE:
  type AccessToken struct {
  	AccessToken string `json:"access_token"` // access_token
  	ExpiresIn   int64  `json:"expires_in"`   // access_token的过期时间
  }

  // AFTER:
  type AccessToken struct {
  	AccessToken string `json:"access_token"` // access_token
  	ExpiresIn   int64  `json:"expires_in"`   // access_token的过期时间
  	ErrCode     int    `json:"errcode"`      // 错误码（正常时为0）
  	ErrMsg      string `json:"errmsg"`       // 错误描述
  }
  ```

- [ ] **4.5** Resolve the duplicate `GetAccessToken` and the `api.base.go:GetStableAccessToken` type mismatch.

  Read `offiaccount/api.base.go` fully to understand `GetStableAccessToken` — it assigns a `*AccessToken` to `c.accessToken` (which is typed `string`). This means `client.go` was refactored but `api.base.go` was not updated. Fix `api.base.go:GetStableAccessToken` to match the real `Client` fields:

  In `offiaccount/api.base.go`, the `GetStableAccessToken` method (lines 25–43) references `c.accessToken` as `*AccessToken`. But `Client.accessToken` is `string` and `Client.expiresAt` is `time.Time`. Rewrite only the storage lines:

  ```go
  // BEFORE (lines 37–43 of api.base.go):
  	// 提前10秒过期，避免临界点问题
  	result.ExpiresIn = result.ExpiresIn + time.Now().Unix() - 10
  	c.tokenMutex.Lock()
  	c.accessToken = result
  	c.tokenMutex.Unlock()
  	return result, nil
  }

  // AFTER:
  	c.tokenMutex.Lock()
  	c.accessToken = result.AccessToken
  	c.expiresAt = time.Now().Add(time.Duration(result.ExpiresIn-10) * time.Second)
  	c.tokenMutex.Unlock()
  	return result, nil
  }
  ```

  Also remove the duplicate `GetAccessToken` declaration from `client.go` (lines 67–70 in `client.go`) since `api.base.go:13` already declares it:

  Remove this block from `offiaccount/client.go`:

  ```go
  // GetAccessToken 获取接口调用凭据
  // 获取全局唯一后台接口调用凭据，token有效期为7200s，开发者需要进行妥善保存。
  func (c *Client) GetAccessToken() string {
  	return c.getAccessToken()
  }
  ```

- [ ] **4.6** Verify the package now builds:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go build ./offiaccount/ 2>&1
  # Expected: no output (clean build)
  ```

- [ ] **4.7** Run the offiaccount tests:

  ```bash
  go test ./offiaccount/ -v -run "TestCheckResp|TestClient_AccessTokenE" 2>&1
  # Expected:
  # === RUN   TestCheckResp
  # --- PASS: TestCheckResp (0.00s)
  # === RUN   TestClient_AccessTokenE_CachesAndRefreshes
  # --- PASS: TestClient_AccessTokenE_CachesAndRefreshes (0.00s)
  # === RUN   TestClient_AccessTokenE_ReturnsWeixinError
  # --- PASS: TestClient_AccessTokenE_ReturnsWeixinError (0.00s)
  # === RUN   TestClient_GetAccessToken_BackwardsCompatible
  # --- PASS: TestClient_GetAccessToken_BackwardsCompatible (0.00s)
  # === RUN   TestClient_AccessTokenE_UsesInjectedTokenSource
  # --- PASS: TestClient_AccessTokenE_UsesInjectedTokenSource (0.00s)
  # PASS
  ```

- [ ] **4.8** Run the full test suite (minus known-broken mini-program):

  ```bash
  go test ./offiaccount/ 2>&1
  # Expected: ok  github.com/godrealms/go-wechat-sdk/offiaccount
  ```

- [ ] **4.9** Add a dedicated `offiaccount/errors_test.go` to lock in the interface contract:

  Create `offiaccount/errors_test.go`:

  ```go
  package offiaccount_test

  import (
  	"errors"
  	"testing"

  	"github.com/godrealms/go-wechat-sdk/offiaccount"
  	"github.com/godrealms/go-wechat-sdk/utils"
  )

  // Compile-time assertion: *offiaccount.WeixinError satisfies utils.WechatAPIError.
  var _ utils.WechatAPIError = (*offiaccount.WeixinError)(nil)

  func TestWeixinError_Code_Message(t *testing.T) {
  	tests := []struct {
  		code int
  		msg  string
  	}{
  		{40001, "invalid credential"},
  		{42001, "access_token expired"},
  		{48001, "api unauthorized"},
  	}
  	for _, tt := range tests {
  		e := &offiaccount.WeixinError{ErrCode: tt.code, ErrMsg: tt.msg}
  		if e.Code() != tt.code {
  			t.Errorf("Code() = %d, want %d", e.Code(), tt.code)
  		}
  		if e.Message() != tt.msg {
  			t.Errorf("Message() = %q, want %q", e.Message(), tt.msg)
  		}
  	}
  }

  func TestCheckResp_ZeroErrcode(t *testing.T) {
  	if err := offiaccount.CheckResp(&offiaccount.Resp{ErrCode: 0}); err != nil {
  		t.Errorf("expected nil, got %v", err)
  	}
  }

  func TestCheckResp_NonZeroErrcode(t *testing.T) {
  	err := offiaccount.CheckResp(&offiaccount.Resp{ErrCode: 40001, ErrMsg: "invalid credential"})
  	if err == nil {
  		t.Fatal("expected error, got nil")
  	}
  	var werr *offiaccount.WeixinError
  	if !errors.As(err, &werr) {
  		t.Fatalf("expected *WeixinError, got %T", err)
  	}
  	if werr.Code() != 40001 {
  		t.Errorf("Code() = %d, want 40001", werr.Code())
  	}
  	if werr.Message() != "invalid credential" {
  		t.Errorf("Message() = %q, want \"invalid credential\"", werr.Message())
  	}

  	// Also satisfies utils.WechatAPIError via errors.As.
  	var apiErr utils.WechatAPIError
  	if !errors.As(err, &apiErr) {
  		t.Fatal("expected utils.WechatAPIError")
  	}
  }
  ```

- [ ] **4.10** Run the extended test:

  ```bash
  go test ./offiaccount/ -v -run "TestWeixinError|TestCheckResp" 2>&1
  # Expected: all PASS
  ```

- [ ] **4.11** Commit:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  git add offiaccount/errors.go offiaccount/errors_test.go offiaccount/struct.base.go offiaccount/api.base.go offiaccount/client.go
  git commit -m "feat(offiaccount): add WeixinError+CheckResp, implement WechatAPIError, fix compile errors"
  ```

---

## Task 5: Add `APIError` to `channels`

**Purpose:** The `channels` package uses `fmt.Errorf` for API errors. Create a proper `APIError` struct that implements `WechatAPIError`, and update `helper.go` to return it.

**Files:**
- Create `channels/errors.go`
- Create `channels/errors_test.go`
- Modify `channels/helper.go`

### Steps

- [ ] **5.1** Write the failing test first.

  Create `channels/errors_test.go`:

  ```go
  package channels_test

  import (
  	"errors"
  	"testing"

  	"github.com/godrealms/go-wechat-sdk/channels"
  	"github.com/godrealms/go-wechat-sdk/utils"
  )

  // Compile-time assertion.
  var _ utils.WechatAPIError = (*channels.APIError)(nil)

  func TestAPIError_Code_Message(t *testing.T) {
  	tests := []struct {
  		code int
  		msg  string
  		path string
  	}{
  		{40001, "invalid credential", "/channels/ec/order/list"},
  		{45009, "api freq out of limit", "/channels/ec/product/list"},
  	}
  	for _, tt := range tests {
  		e := &channels.APIError{ErrCode: tt.code, ErrMsg: tt.msg, Path: tt.path}
  		if e.Code() != tt.code {
  			t.Errorf("Code() = %d, want %d", e.Code(), tt.code)
  		}
  		if e.Message() != tt.msg {
  			t.Errorf("Message() = %q, want %q", e.Message(), tt.msg)
  		}

  		var apiErr utils.WechatAPIError
  		if !errors.As(e, &apiErr) {
  			t.Fatalf("errors.As: expected WechatAPIError")
  		}
  		if apiErr.Code() != tt.code {
  			t.Errorf("via interface: Code() = %d, want %d", apiErr.Code(), tt.code)
  		}
  	}
  }

  func TestAPIError_ErrorString(t *testing.T) {
  	e := &channels.APIError{
  		ErrCode: 40001,
  		ErrMsg:  "invalid credential",
  		Path:    "/channels/ec/order/list",
  	}
  	got := e.Error()
  	want := "channels: /channels/ec/order/list errcode=40001 errmsg=invalid credential"
  	if got != want {
  		t.Errorf("Error() = %q, want %q", got, want)
  	}
  }
  ```

- [ ] **5.2** Run to confirm compile failure (`channels.APIError` not defined):

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go build ./channels/ 2>&1
  # Package builds. Test will fail:
  # go test ./channels/ 2>&1
  # Expected: channels.APIError undefined
  ```

- [ ] **5.3** Create `channels/errors.go`:

  ```go
  package channels

  import "fmt"

  // APIError represents a WeChat Channels API business error (errcode != 0).
  type APIError struct {
  	ErrCode int
  	ErrMsg  string
  	Path    string // the API path that triggered the error
  }

  func (e *APIError) Error() string {
  	return fmt.Sprintf("channels: %s errcode=%d errmsg=%s", e.Path, e.ErrCode, e.ErrMsg)
  }

  // Code returns the numeric errcode. Implements utils.WechatAPIError.
  func (e *APIError) Code() int { return e.ErrCode }

  // Message returns the human-readable errmsg. Implements utils.WechatAPIError.
  func (e *APIError) Message() string { return e.ErrMsg }
  ```

- [ ] **5.4** Update `channels/helper.go` to return `*APIError` instead of `fmt.Errorf`:

  In `channels/helper.go`, change the error return in `doPost`:

  ```go
  // BEFORE:
  	if base.ErrCode != 0 {
  		return fmt.Errorf("channels: %s errcode=%d errmsg=%s", path, base.ErrCode, base.ErrMsg)
  	}

  // AFTER:
  	if base.ErrCode != 0 {
  		return &APIError{ErrCode: base.ErrCode, ErrMsg: base.ErrMsg, Path: path}
  	}
  ```

  Also remove `"fmt"` from the import block in `helper.go` if it is no longer used after this change (check for other `fmt` usages first — if none, remove it).

- [ ] **5.5** Run the tests:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go test ./channels/ -v -run "TestAPIError" 2>&1
  # Expected:
  # === RUN   TestAPIError_Code_Message
  # --- PASS: TestAPIError_Code_Message (0.00s)
  # === RUN   TestAPIError_ErrorString
  # --- PASS: TestAPIError_ErrorString (0.00s)
  # PASS
  ```

- [ ] **5.6** Run the full channels suite to confirm no regressions:

  ```bash
  go test ./channels/ 2>&1
  # Expected: ok  github.com/godrealms/go-wechat-sdk/channels
  ```

- [ ] **5.7** Commit:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  git add channels/errors.go channels/errors_test.go channels/helper.go
  git commit -m "feat(channels): add APIError implementing WechatAPIError"
  ```

---

## Task 6: Add `APIError` to `mini-program`

**Purpose:** The `mini-program` package uses `fmt.Errorf` for API errors in `helper.go`. Create a proper `APIError` struct that implements `WechatAPIError`.

**Files:**
- Create `mini-program/errors.go`
- Create `mini-program/errors_test.go`
- Modify `mini-program/helper.go`

**Note:** The `mini-program` package has a pre-existing build error (`DoRequestWithRawResponse` missing) that is outside the scope of this plan. Create the error type and tests, update `helper.go` where it builds, and verify by running `go vet` on the package. Full `go test` will not pass until the other bug is fixed separately.

### Steps

- [ ] **6.1** Write the failing test first.

  Create `mini-program/errors_test.go`:

  ```go
  package mini_program_test

  import (
  	"errors"
  	"testing"

  	mini_program "github.com/godrealms/go-wechat-sdk/mini-program"
  	"github.com/godrealms/go-wechat-sdk/utils"
  )

  // Compile-time assertion.
  var _ utils.WechatAPIError = (*mini_program.APIError)(nil)

  func TestAPIError_Code_Message(t *testing.T) {
  	tests := []struct {
  		code int
  		msg  string
  		path string
  	}{
  		{40001, "invalid credential", "/wxa/get_wxacode"},
  		{45009, "api freq out of limit", "/wxa/msg_sec_check"},
  	}
  	for _, tt := range tests {
  		e := &mini_program.APIError{ErrCode: tt.code, ErrMsg: tt.msg, Path: tt.path}
  		if e.Code() != tt.code {
  			t.Errorf("Code() = %d, want %d", e.Code(), tt.code)
  		}
  		if e.Message() != tt.msg {
  			t.Errorf("Message() = %q, want %q", e.Message(), tt.msg)
  		}

  		var apiErr utils.WechatAPIError
  		if !errors.As(e, &apiErr) {
  			t.Fatalf("errors.As: expected WechatAPIError")
  		}
  		if apiErr.Code() != tt.code {
  			t.Errorf("via interface: Code() = %d, want %d", apiErr.Code(), tt.code)
  		}
  	}
  }

  func TestAPIError_ErrorString(t *testing.T) {
  	e := &mini_program.APIError{
  		ErrCode: 40001,
  		ErrMsg:  "invalid credential",
  		Path:    "/wxa/get_wxacode",
  	}
  	got := e.Error()
  	want := "mini_program: /wxa/get_wxacode errcode=40001 errmsg=invalid credential"
  	if got != want {
  		t.Errorf("Error() = %q, want %q", got, want)
  	}
  }
  ```

- [ ] **6.2** Create `mini-program/errors.go`:

  ```go
  package mini_program

  import "fmt"

  // APIError represents a WeChat Mini Program API business error (errcode != 0).
  type APIError struct {
  	ErrCode int
  	ErrMsg  string
  	Path    string // the API path that triggered the error
  }

  func (e *APIError) Error() string {
  	return fmt.Sprintf("mini_program: %s errcode=%d errmsg=%s", e.Path, e.ErrCode, e.ErrMsg)
  }

  // Code returns the numeric errcode. Implements utils.WechatAPIError.
  func (e *APIError) Code() int { return e.ErrCode }

  // Message returns the human-readable errmsg. Implements utils.WechatAPIError.
  func (e *APIError) Message() string { return e.ErrMsg }
  ```

- [ ] **6.3** Update `mini-program/helper.go` to return `*APIError` instead of `fmt.Errorf` in `doPost` and `doPostRaw`:

  In `doPost`:
  ```go
  // BEFORE:
  	if base.ErrCode != 0 {
  		return fmt.Errorf("mini_program: %s errcode=%d errmsg=%s", path, base.ErrCode, base.ErrMsg)
  	}

  // AFTER:
  	if base.ErrCode != 0 {
  		return &APIError{ErrCode: base.ErrCode, ErrMsg: base.ErrMsg, Path: path}
  	}
  ```

  In `doPostRaw`:
  ```go
  // BEFORE:
  		if json.Unmarshal(respBody, &resp) == nil && resp.ErrCode != 0 {
  			return nil, fmt.Errorf("mini_program: %s errcode=%d errmsg=%s", path, resp.ErrCode, resp.ErrMsg)
  		}

  // AFTER:
  		if json.Unmarshal(respBody, &resp) == nil && resp.ErrCode != 0 {
  			return nil, &APIError{ErrCode: resp.ErrCode, ErrMsg: resp.ErrMsg, Path: path}
  		}
  ```

  Remove `"fmt"` from the import block in `helper.go` if no other usages remain.

- [ ] **6.4** Verify via vet (full test blocked by pre-existing `DoRequestWithRawResponse` bug):

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go vet ./mini-program/ 2>&1
  # Expected: only the DoRequestWithRawResponse errors — the new files are clean.
  ```

- [ ] **6.5** Commit:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  git add mini-program/errors.go mini-program/errors_test.go mini-program/helper.go
  git commit -m "feat(mini-program): add APIError implementing WechatAPIError"
  ```

---

## Task 7: Add `APIError` to `mini-game`

**Purpose:** The `mini-game` package uses `fmt.Errorf` for API errors. Create a proper `APIError` struct.

**Files:**
- Create `mini-game/errors.go`
- Create `mini-game/errors_test.go`
- Modify `mini-game/helper.go`

### Steps

- [ ] **7.1** Write the failing test first.

  Create `mini-game/errors_test.go`:

  ```go
  package mini_game_test

  import (
  	"errors"
  	"testing"

  	mini_game "github.com/godrealms/go-wechat-sdk/mini-game"
  	"github.com/godrealms/go-wechat-sdk/utils"
  )

  // Compile-time assertion.
  var _ utils.WechatAPIError = (*mini_game.APIError)(nil)

  func TestAPIError_Code_Message(t *testing.T) {
  	tests := []struct {
  		code int
  		msg  string
  		path string
  	}{
  		{40001, "invalid credential", "/wxa/game/getaccessinfo"},
  		{45009, "api freq out of limit", "/wxa/game/getframesync"},
  	}
  	for _, tt := range tests {
  		e := &mini_game.APIError{ErrCode: tt.code, ErrMsg: tt.msg, Path: tt.path}
  		if e.Code() != tt.code {
  			t.Errorf("Code() = %d, want %d", e.Code(), tt.code)
  		}
  		if e.Message() != tt.msg {
  			t.Errorf("Message() = %q, want %q", e.Message(), tt.msg)
  		}

  		var apiErr utils.WechatAPIError
  		if !errors.As(e, &apiErr) {
  			t.Fatalf("errors.As: expected WechatAPIError")
  		}
  		if apiErr.Code() != tt.code {
  			t.Errorf("via interface: Code() = %d, want %d", apiErr.Code(), tt.code)
  		}
  	}
  }

  func TestAPIError_ErrorString(t *testing.T) {
  	e := &mini_game.APIError{
  		ErrCode: 40001,
  		ErrMsg:  "invalid credential",
  		Path:    "/wxa/game/getaccessinfo",
  	}
  	got := e.Error()
  	want := "mini_game: /wxa/game/getaccessinfo errcode=40001 errmsg=invalid credential"
  	if got != want {
  		t.Errorf("Error() = %q, want %q", got, want)
  	}
  }
  ```

- [ ] **7.2** Create `mini-game/errors.go`:

  ```go
  package mini_game

  import "fmt"

  // APIError represents a WeChat Mini Game API business error (errcode != 0).
  type APIError struct {
  	ErrCode int
  	ErrMsg  string
  	Path    string // the API path that triggered the error
  }

  func (e *APIError) Error() string {
  	return fmt.Sprintf("mini_game: %s errcode=%d errmsg=%s", e.Path, e.ErrCode, e.ErrMsg)
  }

  // Code returns the numeric errcode. Implements utils.WechatAPIError.
  func (e *APIError) Code() int { return e.ErrCode }

  // Message returns the human-readable errmsg. Implements utils.WechatAPIError.
  func (e *APIError) Message() string { return e.ErrMsg }
  ```

- [ ] **7.3** Update `mini-game/helper.go` to return `*APIError`:

  ```go
  // BEFORE:
  	if base.ErrCode != 0 {
  		return fmt.Errorf("mini_game: %s errcode=%d errmsg=%s", path, base.ErrCode, base.ErrMsg)
  	}

  // AFTER:
  	if base.ErrCode != 0 {
  		return &APIError{ErrCode: base.ErrCode, ErrMsg: base.ErrMsg, Path: path}
  	}
  ```

  Remove `"fmt"` from `helper.go`'s import block if no other `fmt` usages remain.

- [ ] **7.4** Run the tests:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go test ./mini-game/ -v -run "TestAPIError" 2>&1
  # Expected:
  # === RUN   TestAPIError_Code_Message
  # --- PASS: TestAPIError_Code_Message (0.00s)
  # === RUN   TestAPIError_ErrorString
  # --- PASS: TestAPIError_ErrorString (0.00s)
  # PASS
  ```

- [ ] **7.5** Run the full mini-game suite:

  ```bash
  go test ./mini-game/ 2>&1
  # Expected: ok  github.com/godrealms/go-wechat-sdk/mini-game
  ```

- [ ] **7.6** Commit:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  git add mini-game/errors.go mini-game/errors_test.go mini-game/helper.go
  git commit -m "feat(mini-game): add APIError implementing WechatAPIError"
  ```

---

## Task 8: Add `APIError` to `merchant/developed`

**Purpose:** The `merchant/developed` package has no centralised error type. Create one for use by API methods.

**Files:**
- Create `merchant/developed/errors.go`
- Create `merchant/developed/errors_test.go`

### Steps

- [ ] **8.1** Write the failing test first.

  Create `merchant/developed/errors_test.go`:

  ```go
  package developed_test

  import (
  	"errors"
  	"testing"

  	developed "github.com/godrealms/go-wechat-sdk/merchant/developed"
  	"github.com/godrealms/go-wechat-sdk/utils"
  )

  // Compile-time assertion.
  var _ utils.WechatAPIError = (*developed.APIError)(nil)

  func TestAPIError_Code_Message(t *testing.T) {
  	tests := []struct {
  		code int
  		msg  string
  		path string
  	}{
  		{400, "PARAM_ERROR", "/v3/pay/transactions/app"},
  		{401, "SIGN_ERROR", "/v3/pay/transactions/jsapi"},
  		{500, "SYSTEM_ERROR", "/v3/pay/transactions/native"},
  	}
  	for _, tt := range tests {
  		e := &developed.APIError{ErrCode: tt.code, ErrMsg: tt.msg, Path: tt.path}
  		if e.Code() != tt.code {
  			t.Errorf("Code() = %d, want %d", e.Code(), tt.code)
  		}
  		if e.Message() != tt.msg {
  			t.Errorf("Message() = %q, want %q", e.Message(), tt.msg)
  		}

  		var apiErr utils.WechatAPIError
  		if !errors.As(e, &apiErr) {
  			t.Fatalf("errors.As: expected WechatAPIError")
  		}
  		if apiErr.Code() != tt.code {
  			t.Errorf("via interface: Code() = %d, want %d", apiErr.Code(), tt.code)
  		}
  	}
  }

  func TestAPIError_ErrorString(t *testing.T) {
  	e := &developed.APIError{
  		ErrCode: 400,
  		ErrMsg:  "PARAM_ERROR",
  		Path:    "/v3/pay/transactions/app",
  	}
  	got := e.Error()
  	want := "merchant/developed: /v3/pay/transactions/app errcode=400 errmsg=PARAM_ERROR"
  	if got != want {
  		t.Errorf("Error() = %q, want %q", got, want)
  	}
  }
  ```

- [ ] **8.2** Create `merchant/developed/errors.go`:

  ```go
  package developed

  import "fmt"

  // APIError represents a WeChat Pay merchant API business error.
  type APIError struct {
  	ErrCode int
  	ErrMsg  string
  	Path    string // the API path that triggered the error
  }

  func (e *APIError) Error() string {
  	return fmt.Sprintf("merchant/developed: %s errcode=%d errmsg=%s", e.Path, e.ErrCode, e.ErrMsg)
  }

  // Code returns the numeric errcode. Implements utils.WechatAPIError.
  func (e *APIError) Code() int { return e.ErrCode }

  // Message returns the human-readable errmsg. Implements utils.WechatAPIError.
  func (e *APIError) Message() string { return e.ErrMsg }
  ```

- [ ] **8.3** Run the tests:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go test ./merchant/developed/ -v -run "TestAPIError" 2>&1
  # Expected:
  # === RUN   TestAPIError_Code_Message
  # --- PASS: TestAPIError_Code_Message (0.00s)
  # === RUN   TestAPIError_ErrorString
  # --- PASS: TestAPIError_ErrorString (0.00s)
  # PASS
  ```

- [ ] **8.4** Run full merchant/developed suite:

  ```bash
  go test ./merchant/developed/ 2>&1
  # Expected: ok  github.com/godrealms/go-wechat-sdk/merchant/developed
  ```

- [ ] **8.5** Commit:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  git add merchant/developed/errors.go merchant/developed/errors_test.go
  git commit -m "feat(merchant/developed): add APIError implementing WechatAPIError"
  ```

---

## Task 9: Fix JSON marshal error-ignoring bugs in `merchant/developed/types`

**Purpose:** Four `ToString()` methods across three files silently discard `json.Marshal` errors, returning an empty string on failure. Fix each to return `"<marshal error>"` on failure and log the error, making the bug observable without changing the method signature (which is `ToString() string` and must stay that way for backwards compatibility).

**Bug pattern (in 4 locations across 3 files):**
```go
// BROKEN — error silently dropped:
func (a *Transactions) ToString() string {
    marshal, _ := json.Marshal(a)
    return string(marshal)
}
```

**Fix pattern:**
```go
func (a *Transactions) ToString() string {
    marshal, err := json.Marshal(a)
    if err != nil {
        return "<marshal error: " + err.Error() + ">"
    }
    return string(marshal)
}
```

**Affected locations (confirmed by grep):**

| File | Type | Line (approx.) |
|------|------|----------------|
| `merchant/developed/types/pay.transactions.go` | `MchID` | 9–12 |
| `merchant/developed/types/pay.transactions.app.go` | `Transactions` | 178–181 |
| `merchant/developed/types/pay.transactions.refunds.go` | `Refunds` | 53–56 |
| `merchant/developed/types/pay.transactions.refunds.go` | `AbnormalRefund` | 155–158 |

**Files:**
- Modify `merchant/developed/types/pay.transactions.go`
- Modify `merchant/developed/types/pay.transactions.app.go`
- Modify `merchant/developed/types/pay.transactions.refunds.go`
- Create `merchant/developed/types/tostring_test.go`

### Steps

- [ ] **9.1** Write the failing test first to confirm current behaviour (error dropped) and specify desired behaviour.

  Create `merchant/developed/types/tostring_test.go`:

  ```go
  package types_test

  import (
  	"strings"
  	"testing"

  	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
  )

  // TestMchID_ToString_ValidJSON verifies MchID.ToString returns valid JSON.
  func TestMchID_ToString_ValidJSON(t *testing.T) {
  	m := &types.MchID{Mchid: "1234567890"}
  	got := m.ToString()
  	if !strings.Contains(got, `"mchid"`) {
  		t.Errorf("ToString() = %q, want JSON containing mchid key", got)
  	}
  	if !strings.Contains(got, "1234567890") {
  		t.Errorf("ToString() = %q, want JSON containing 1234567890", got)
  	}
  }

  // TestTransactions_ToString_ValidJSON verifies Transactions.ToString returns valid JSON.
  func TestTransactions_ToString_ValidJSON(t *testing.T) {
  	tx := &types.Transactions{
  		Appid:       "wx123",
  		Mchid:       "1234567890",
  		Description: "test",
  		OutTradeNo:  "order-001",
  		NotifyUrl:   "https://example.com/notify",
  		Amount:      &types.Amount{Total: 100, Currency: "CNY"},
  	}
  	got := tx.ToString()
  	if !strings.Contains(got, `"appid"`) {
  		t.Errorf("ToString() = %q, want JSON with appid key", got)
  	}
  	if !strings.Contains(got, "wx123") {
  		t.Errorf("ToString() = %q, want JSON with wx123", got)
  	}
  	if strings.HasPrefix(got, "<marshal error") {
  		t.Errorf("ToString() returned error string: %s", got)
  	}
  }

  // TestRefunds_ToString_ValidJSON verifies Refunds.ToString returns valid JSON.
  func TestRefunds_ToString_ValidJSON(t *testing.T) {
  	r := &types.Refunds{
  		OutTradeNo:  "order-001",
  		OutRefundNo: "refund-001",
  		Reason:      "test refund",
  		NotifyUrl:   "https://example.com/notify",
  		FundsAccount: "AVAILABLE",
  		Amount:      &types.Amount{Total: 100, Refund: 50},
  	}
  	got := r.ToString()
  	if !strings.Contains(got, `"out_trade_no"`) {
  		t.Errorf("ToString() = %q, want JSON with out_trade_no key", got)
  	}
  	if strings.HasPrefix(got, "<marshal error") {
  		t.Errorf("ToString() returned error string: %s", got)
  	}
  }

  // TestAbnormalRefund_ToString_ValidJSON verifies AbnormalRefund.ToString returns valid JSON.
  func TestAbnormalRefund_ToString_ValidJSON(t *testing.T) {
  	r := &types.AbnormalRefund{
  		OutRefundNo: "refund-001",
  		Type:        "USER_BANK_CARD",
  		BankType:    "ICBC",
  		BankAccount: "encrypted-account",
  		RealName:    "encrypted-name",
  	}
  	got := r.ToString()
  	if !strings.Contains(got, `"out_refund_no"`) {
  		t.Errorf("ToString() = %q, want JSON with out_refund_no key", got)
  	}
  	if strings.HasPrefix(got, "<marshal error") {
  		t.Errorf("ToString() returned error string: %s", got)
  	}
  }

  // TestToString_ErrorHandling verifies that an error is surfaced rather than silently dropped.
  // We cannot easily trigger a json.Marshal error on a plain struct in standard Go,
  // so we test the negative case: a valid struct should never produce "<marshal error".
  func TestToString_NeverSilentlyEmpty(t *testing.T) {
  	m := &types.MchID{Mchid: "test123"}
  	got := m.ToString()
  	if got == "" {
  		t.Error("ToString() returned empty string; expected JSON or error marker")
  	}
  }
  ```

- [ ] **9.2** Run to confirm tests currently pass (current code works for valid structs; the bug is only visible when marshal fails, which we cannot easily force):

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go test ./merchant/developed/types/ -v -run "TestMchID|TestTransactions|TestRefunds|TestAbnormalRefund|TestToString" 2>&1
  # Expected: all PASS (the test verifies correct output; error-handling fix is still needed for correctness)
  ```

- [ ] **9.3** Fix `merchant/developed/types/pay.transactions.go` — `MchID.ToString()`:

  ```go
  // BEFORE:
  func (t *MchID) ToString() string {
  	marshal, _ := json.Marshal(t)
  	return string(marshal)
  }

  // AFTER:
  func (t *MchID) ToString() string {
  	marshal, err := json.Marshal(t)
  	if err != nil {
  		return "<marshal error: " + err.Error() + ">"
  	}
  	return string(marshal)
  }
  ```

- [ ] **9.4** Fix `merchant/developed/types/pay.transactions.app.go` — `Transactions.ToString()`:

  ```go
  // BEFORE:
  func (a *Transactions) ToString() string {
  	marshal, _ := json.Marshal(a)
  	return string(marshal)
  }

  // AFTER:
  func (a *Transactions) ToString() string {
  	marshal, err := json.Marshal(a)
  	if err != nil {
  		return "<marshal error: " + err.Error() + ">"
  	}
  	return string(marshal)
  }
  ```

- [ ] **9.5** Fix `merchant/developed/types/pay.transactions.refunds.go` — both `Refunds.ToString()` and `AbnormalRefund.ToString()`:

  Fix `Refunds.ToString()` (around line 53):
  ```go
  // BEFORE:
  func (r *Refunds) ToString() string {
  	marshal, _ := json.Marshal(r)
  	return string(marshal)
  }

  // AFTER:
  func (r *Refunds) ToString() string {
  	marshal, err := json.Marshal(r)
  	if err != nil {
  		return "<marshal error: " + err.Error() + ">"
  	}
  	return string(marshal)
  }
  ```

  Fix `AbnormalRefund.ToString()` (around line 155):
  ```go
  // BEFORE:
  func (r *AbnormalRefund) ToString() string {
  	marshal, _ := json.Marshal(r)
  	return string(marshal)
  }

  // AFTER:
  func (r *AbnormalRefund) ToString() string {
  	marshal, err := json.Marshal(r)
  	if err != nil {
  		return "<marshal error: " + err.Error() + ">"
  	}
  	return string(marshal)
  }
  ```

- [ ] **9.6** Verify no other `json.Marshal` error-ignoring patterns remain in the types directory:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  grep -n "marshal, _" merchant/developed/types/*.go
  # Expected: no output (all instances fixed)
  ```

- [ ] **9.7** Run the tests:

  ```bash
  go test ./merchant/developed/types/ -v 2>&1
  # Expected: all PASS
  ```

- [ ] **9.8** Run full merchant/developed suite:

  ```bash
  go test ./merchant/developed/... 2>&1
  # Expected: ok  github.com/godrealms/go-wechat-sdk/merchant/developed
  #           ok  github.com/godrealms/go-wechat-sdk/merchant/developed/types (if tests exist)
  #           ?   github.com/godrealms/go-wechat-sdk/merchant/developed/errorx [no test files]
  ```

- [ ] **9.9** Commit:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  git add \
    merchant/developed/types/pay.transactions.go \
    merchant/developed/types/pay.transactions.app.go \
    merchant/developed/types/pay.transactions.refunds.go \
    merchant/developed/types/tostring_test.go
  git commit -m "fix(merchant/developed/types): handle json.Marshal errors in ToString() methods"
  ```

---

## Task 10: Fix `offiaccount` errcode checking — replace `errors.New(result.ErrMsg)` with `&WeixinError{}`

**Purpose:** Throughout the `offiaccount` API files, errcode errors are returned as bare `errors.New(result.ErrMsg)` strings, losing the numeric code. Replace these with `&WeixinError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}` or calls to `CheckResp()` where a `Resp`-embedded struct is available, so that callers can use `errors.As` to retrieve the code.

**Scope:** All `api.*.go` files in `offiaccount/` that contain `errors.New(result.ErrMsg)`.

**Confirmed affected files (from grep output):**
- `offiaccount/api.api-manage.go`
- `offiaccount/api.base.go`
- `offiaccount/api.custom-menu.go`
- `offiaccount/api.notify.message.go`
- `offiaccount/api.notify.notify.go`
- `offiaccount/api.notify.subscribe.go`

**Pattern to replace:**

```go
// BEFORE (in every affected location):
} else if result.ErrCode != 0 {
    return ..., errors.New(result.ErrMsg)
}

// AFTER — preserves errcode so callers can inspect it:
} else if result.ErrCode != 0 {
    return ..., &WeixinError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
}
```

**Alternatively**, if the `result` type embeds `Resp`, use `CheckResp`:

```go
// When result embeds Resp directly (e.g. type FooResp struct { Resp; ... }):
if err := CheckResp(&result.Resp); err != nil {
    return ..., err
}
```

Use the `&WeixinError{}` form uniformly — it is safer and does not require knowing whether `Resp` is embedded.

**Files:**
- Modify `offiaccount/api.api-manage.go`
- Modify `offiaccount/api.base.go`
- Modify `offiaccount/api.custom-menu.go`
- Modify `offiaccount/api.notify.message.go`
- Modify `offiaccount/api.notify.notify.go`
- Modify `offiaccount/api.notify.subscribe.go`
- Create `offiaccount/errcode_test.go`

### Steps

- [ ] **10.1** Write the failing test that demonstrates the current broken behaviour (errcode is lost):

  Create `offiaccount/errcode_test.go`:

  ```go
  package offiaccount_test

  import (
  	"context"
  	"errors"
  	"net/http"
  	"net/http/httptest"
  	"testing"
  	"time"

  	"github.com/godrealms/go-wechat-sdk/offiaccount"
  	"github.com/godrealms/go-wechat-sdk/utils"
  )

  func newTestClient(srv *httptest.Server) *offiaccount.Client {
  	cfg := &offiaccount.Config{AppId: "appid", AppSecret: "secret"}
  	c := offiaccount.NewClient(context.Background(), cfg,
  		offiaccount.WithHTTPClient(utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))),
  		offiaccount.WithTokenSource(&staticToken{token: "FAKE_TOKEN"}),
  	)
  	return c
  }

  type staticToken struct{ token string }

  func (s *staticToken) AccessToken(_ context.Context) (string, error) { return s.token, nil }

  // TestGetCallbackIp_ReturnsWeixinError verifies that a non-zero errcode from the API
  // is surfaced as a *WeixinError (with code preserved), not a bare string error.
  func TestGetCallbackIp_ReturnsWeixinError(t *testing.T) {
  	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  		w.Header().Set("Content-Type", "application/json")
  		_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid credential","ip_list":null}`))
  	}))
  	defer srv.Close()

  	c := newTestClient(srv)
  	_, err := c.GetCallbackIp()
  	if err == nil {
  		t.Fatal("expected error, got nil")
  	}

  	var werr *offiaccount.WeixinError
  	if !errors.As(err, &werr) {
  		t.Fatalf("expected *WeixinError, got %T: %v", err, err)
  	}
  	if werr.Code() != 40001 {
  		t.Errorf("Code() = %d, want 40001", werr.Code())
  	}

  	// Must also satisfy the generic interface.
  	var apiErr utils.WechatAPIError
  	if !errors.As(err, &apiErr) {
  		t.Fatal("expected utils.WechatAPIError")
  	}
  }

  // TestGetApiDomainIP_ReturnsWeixinError verifies the same for GetApiDomainIP.
  func TestGetApiDomainIP_ReturnsWeixinError(t *testing.T) {
  	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  		w.Header().Set("Content-Type", "application/json")
  		_, _ = w.Write([]byte(`{"errcode":48001,"errmsg":"api unauthorized","ip_list":null}`))
  	}))
  	defer srv.Close()

  	c := newTestClient(srv)
  	_, err := c.GetApiDomainIP()
  	if err == nil {
  		t.Fatal("expected error, got nil")
  	}

  	var werr *offiaccount.WeixinError
  	if !errors.As(err, &werr) {
  		t.Fatalf("expected *WeixinError, got %T: %v", err, err)
  	}
  	if werr.Code() != 48001 {
  		t.Errorf("Code() = %d, want 48001", werr.Code())
  	}
  }
  ```

- [ ] **10.2** Run the test to confirm it fails (currently `errors.New` is returned, so `errors.As(*WeixinError)` returns false):

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go test ./offiaccount/ -v -run "TestGetCallbackIp_ReturnsWeixinError|TestGetApiDomainIP_ReturnsWeixinError" 2>&1
  # Expected: FAIL — errors.As(*WeixinError) fails because errors.New was returned
  ```

- [ ] **10.3** Fix `offiaccount/api.base.go` — `GetCallbackIp` and `GetApiDomainIP`:

  ```go
  // In GetCallbackIp — BEFORE:
  } else if result.ErrCode != 0 {
      return nil, errors.New(result.ErrMsg)
  }

  // AFTER:
  } else if result.ErrCode != 0 {
      return nil, &WeixinError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
  }
  ```

  Apply the same change to `GetApiDomainIP`. Remove the `"errors"` import from `api.base.go` if it is no longer used elsewhere in that file after the changes.

- [ ] **10.4** Fix `offiaccount/api.api-manage.go`:

  Grep the file to find all `errors.New(result.ErrMsg)` patterns:

  ```bash
  grep -n "errors.New(result.ErrMsg)" /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig/offiaccount/api.api-manage.go
  ```

  For each occurrence, apply the fix:
  ```go
  // BEFORE:
  return nil, errors.New(result.ErrMsg)
  // or:
  return errors.New(result.ErrMsg)

  // AFTER:
  return nil, &WeixinError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
  // or:
  return &WeixinError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
  ```

  Remove `"errors"` import if no longer needed.

- [ ] **10.5** Fix `offiaccount/api.custom-menu.go`:

  ```bash
  grep -n "errors.New(result.ErrMsg)" /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig/offiaccount/api.custom-menu.go
  ```

  Apply the fix pattern to every occurrence. Remove `"errors"` import if no longer needed.

- [ ] **10.6** Fix `offiaccount/api.notify.message.go`:

  ```bash
  grep -n "errors.New\|Errorf.*wechat api error" /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig/offiaccount/api.notify.message.go
  ```

  Note: this file has a mix of `fmt.Errorf("wechat api error: %d - %s", ...)` and `errors.New(result.ErrMsg)`. Fix both patterns:

  ```go
  // fmt.Errorf pattern — BEFORE:
  return &uploadResp, fmt.Errorf("wechat api error: %d - %s", uploadResp.ErrCode, uploadResp.ErrMsg)

  // AFTER:
  return &uploadResp, &WeixinError{ErrCode: uploadResp.ErrCode, ErrMsg: uploadResp.ErrMsg}
  ```

  ```go
  // errors.New pattern — BEFORE:
  return ..., errors.New(result.ErrMsg)

  // AFTER:
  return ..., &WeixinError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
  ```

  Remove unused `"errors"` and `"fmt"` imports.

- [ ] **10.7** Fix `offiaccount/api.notify.notify.go`:

  ```bash
  grep -n "errors.New(result.ErrMsg)" /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig/offiaccount/api.notify.notify.go
  ```

  Apply fix to all occurrences.

- [ ] **10.8** Fix `offiaccount/api.notify.subscribe.go`:

  ```bash
  grep -n "errors.New(result.ErrMsg)" /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig/offiaccount/api.notify.subscribe.go
  ```

  Apply fix to all occurrences.

- [ ] **10.9** Scan all remaining `offiaccount/api.*.go` files to ensure no occurrences were missed:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  grep -rn "errors\.New(result\.ErrMsg)\|errors\.New(uploadResp\.ErrMsg)" offiaccount/api.*.go
  # Expected: no output
  ```

- [ ] **10.10** Build and run tests:

  ```bash
  go build ./offiaccount/ 2>&1
  # Expected: no output (clean)

  go test ./offiaccount/ -v -run "TestGetCallbackIp|TestGetApiDomainIP|TestCheckResp|TestWeixinError" 2>&1
  # Expected: all PASS
  ```

- [ ] **10.11** Run the full offiaccount suite:

  ```bash
  go test ./offiaccount/ 2>&1
  # Expected: ok  github.com/godrealms/go-wechat-sdk/offiaccount
  ```

- [ ] **10.12** Commit:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  git add \
    offiaccount/errcode_test.go \
    offiaccount/api.api-manage.go \
    offiaccount/api.base.go \
    offiaccount/api.custom-menu.go \
    offiaccount/api.notify.message.go \
    offiaccount/api.notify.notify.go \
    offiaccount/api.notify.subscribe.go
  git commit -m "fix(offiaccount): replace errors.New with WeixinError to preserve errcode in all api files"
  ```

---

## Final Verification

After all tasks are complete, run a broad check across all packages touched in this plan:

- [ ] **V.1** Run all passing package suites:

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-wechat-sdk/.claude/worktrees/xenodochial-taussig
  go test \
    ./utils/... \
    ./oplatform/... \
    ./work-wechat/isv/... \
    ./offiaccount/... \
    ./channels/... \
    ./mini-game/... \
    ./merchant/developed/... \
    2>&1
  # Expected: all packages report "ok" or "?" (no test files)
  # Known exception: mini-program is excluded due to pre-existing DoRequestWithRawResponse build error
  ```

- [ ] **V.2** Confirm no `errors.New(result.ErrMsg)` patterns remain in `offiaccount/`:

  ```bash
  grep -rn "errors\.New(result\.ErrMsg)" offiaccount/ 2>&1
  # Expected: no output
  ```

- [ ] **V.3** Confirm no `json.Marshal` error-ignoring patterns remain in `merchant/developed/types/`:

  ```bash
  grep -n "marshal, _" merchant/developed/types/*.go 2>&1
  # Expected: no output
  ```

- [ ] **V.4** Confirm all five packages export a type satisfying `WechatAPIError`:

  ```bash
  grep -rn "func.*Code() int\|func.*Message() string" \
    oplatform/ \
    work-wechat/isv/ \
    offiaccount/ \
    channels/ \
    mini-game/ \
    merchant/developed/ \
    2>&1 | grep -v "_test.go"
  # Expected: two lines per package (Code and Message methods)
  ```

- [ ] **V.5** Confirm `utils.WechatAPIError` is the only new public interface added to `utils/`:

  ```bash
  grep -n "^type\|^func\|^var" utils/wechat_error.go
  # Expected: only "type WechatAPIError interface"
  ```

---

## Summary of All Changed Files

| File | Action | Task |
|------|--------|------|
| `utils/wechat_error.go` | Create — `WechatAPIError` interface | 1 |
| `utils/wechat_error_test.go` | Create — interface satisfaction tests | 1 |
| `oplatform/errors.go` | Modify — add `Code()`, `Message()` | 2 |
| `oplatform/errors_test.go` | Create — interface tests | 2 |
| `work-wechat/isv/errors.go` | Modify — add `Code()`, `Message()` | 3 |
| `work-wechat/isv/errors_test.go` | Create — interface tests | 3 |
| `offiaccount/errors.go` | Create — `WeixinError`, `CheckResp` | 4 |
| `offiaccount/errors_test.go` | Create — interface + CheckResp tests | 4 |
| `offiaccount/struct.base.go` | Modify — add `ErrCode`, `ErrMsg` to `AccessToken` | 4 |
| `offiaccount/api.base.go` | Modify — fix `GetStableAccessToken` storage | 4 |
| `offiaccount/client.go` | Modify — remove duplicate `GetAccessToken` | 4 |
| `channels/errors.go` | Create — `APIError` struct | 5 |
| `channels/errors_test.go` | Create — interface tests | 5 |
| `channels/helper.go` | Modify — return `*APIError` | 5 |
| `mini-program/errors.go` | Create — `APIError` struct | 6 |
| `mini-program/errors_test.go` | Create — interface tests | 6 |
| `mini-program/helper.go` | Modify — return `*APIError` | 6 |
| `mini-game/errors.go` | Create — `APIError` struct | 7 |
| `mini-game/errors_test.go` | Create — interface tests | 7 |
| `mini-game/helper.go` | Modify — return `*APIError` | 7 |
| `merchant/developed/errors.go` | Create — `APIError` struct | 8 |
| `merchant/developed/errors_test.go` | Create — interface tests | 8 |
| `merchant/developed/types/pay.transactions.go` | Modify — fix `MchID.ToString()` | 9 |
| `merchant/developed/types/pay.transactions.app.go` | Modify — fix `Transactions.ToString()` | 9 |
| `merchant/developed/types/pay.transactions.refunds.go` | Modify — fix `Refunds.ToString()` and `AbnormalRefund.ToString()` | 9 |
| `merchant/developed/types/tostring_test.go` | Create — ToString tests | 9 |
| `offiaccount/errcode_test.go` | Create — errcode preservation tests | 10 |
| `offiaccount/api.api-manage.go` | Modify — `&WeixinError{}` instead of `errors.New` | 10 |
| `offiaccount/api.base.go` | Modify — `&WeixinError{}` instead of `errors.New` | 10 |
| `offiaccount/api.custom-menu.go` | Modify — `&WeixinError{}` instead of `errors.New` | 10 |
| `offiaccount/api.notify.message.go` | Modify — `&WeixinError{}` instead of `errors.New`/`fmt.Errorf` | 10 |
| `offiaccount/api.notify.notify.go` | Modify — `&WeixinError{}` instead of `errors.New` | 10 |
| `offiaccount/api.notify.subscribe.go` | Modify — `&WeixinError{}` instead of `errors.New` | 10 |
