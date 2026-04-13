# Fix & Optimize Design Spec

**Date:** 2026-04-13  
**Scope:** Code quality fixes, test coverage, Godoc comments, three stub package implementations  
**Execution order:** A → B → D → C

---

## Sub-project A: Code Quality Fixes

### Goal

Fix real bugs and unify error handling patterns across all packages.

### A1: Common Error Interface

Add `utils/wechat_error.go` — a pure interface, no shared struct:

```go
package utils

// WechatAPIError is implemented by all package-level API error types.
type WechatAPIError interface {
    error
    Code() int
    Message() string
}
```

Each package retains its own `APIError` struct and implements this interface by adding `Code() int` and `Message() string` methods. No cross-package struct sharing.

**Packages requiring changes:**

| Package | Action |
|---------|--------|
| `utils/` | Add `wechat_error.go` with interface definition |
| `offiaccount/` | Add `APIError` struct implementing interface |
| `channels/` | Add `Code()` / `Message()` methods to existing error type |
| `mini-program/` | Add `Code()` / `Message()` methods to existing error type |
| `mini-game/` | Add `Code()` / `Message()` methods to existing error type |
| `merchant/developed/` | Add `APIError` struct implementing interface |
| `oplatform/` | Add `Code()` / `Message()` methods to existing `WeixinError` |
| `work-wechat/isv/` | Add `Code()` / `Message()` methods to existing `WeixinError` |

### A2: Fix JSON Marshal Error Ignoring in merchant/developed/types

All `ToString()` methods that silently discard marshal errors must be fixed:

```go
// Before (buggy):
func (t *Foo) ToString() string {
    marshal, _ := json.Marshal(t)
    return string(marshal)
}

// After (correct):
func (t *Foo) ToString() string {
    b, err := json.Marshal(t)
    if err != nil {
        return ""
    }
    return string(b)
}
```

Affects all types in `merchant/developed/types/` that implement `ToString()`.

### A3: Fix offiaccount errcode Checking

Approximately 70% of `offiaccount`'s 51 API files do not check WeChat errcode in responses. `offiaccount` already has `WeixinError` and `CheckResp()` in `client.go` — standardize all API files to use the existing `CheckResp()` function consistently.

- No new helper file needed; `CheckResp()` already exists in `client.go`
- Apply `CheckResp()` to all API files that currently ignore `ErrCode` after calling `doPost` or `doGet`
- Replace ad-hoc `fmt.Errorf("wechat api error: %d - %s", ...)` patterns with `CheckResp()`

---

## Sub-project B: Test Coverage

### Goal

Bring critical packages from near-zero test coverage to ≥ 70%.

### Testing Strategy

Follow the patterns established in `work-wechat/isv` and `oplatform`:
- **HTTP mock:** inject `*httptest.Server` via client options — no real WeChat API calls
- **Table-driven tests:** each API method covers three cases minimum:
  1. Successful response (2xx, errcode=0)
  2. WeChat error response (errcode ≠ 0)
  3. Network/HTTP error
- **File naming:** `xxx_test.go` matches source file name

### B1: offiaccount (51 API files, ~3% → ≥ 70%)

Group API files into test files by functional domain:

| Test file | Covers |
|-----------|--------|
| `api.menu_test.go` | Custom menus |
| `api.message_test.go` | Customer service messages, mass messaging, template messages |
| `api.user_test.go` | User management, tags, blacklist |
| `api.material_test.go` | Temporary/permanent materials, draft box |
| `api.qrcode_test.go` | QR codes, short URLs |
| `api.analysis_test.go` | Data analytics (user, article, message, interface) |
| `api.invoice_test.go` | Electronic invoices |
| `api.aiopen_test.go` | OCR, image cropping, speech recognition |
| `api.nontax_test.go` | Non-tax payment |
| `api.comment_test.go` | Comment management |
| `api.jssdk_test.go` | JS-SDK signature |
| `api.oauth_test.go` | Web OAuth |

### B2: merchant/developed/types (0% → ≥ 70%)

- JSON serialization/deserialization roundtrip tests for all request/response types
- `ToString()` correctness (including edge cases: empty struct, nil fields)
- `ToUrlValues()` correctness

### B3: utils (0% → ≥ 80%)

- `http.go`: `Post`, `Get`, `Put` methods
  - Correct request body serialization
  - Header injection (Content-Type, Authorization)
  - Timeout handling
  - Non-2xx HTTP status error propagation

### Acceptance Criteria

```
go test ./... — all pass
offiaccount coverage ≥ 70%
merchant/developed coverage ≥ 70%
utils coverage ≥ 80%
```

---

## Sub-project D: Godoc Comments

### Goal

All exported symbols in all packages have meaningful Godoc comments.

### Rules

- Comments start with the symbol name
- Describe **what it is** and **key behavior or constraints**
- Do NOT restate the signature (e.g., avoid `// GetFoo returns foo`)
- Do NOT describe obvious behavior

```go
// Client manages authentication and provides access to offiaccount APIs.
// It automatically refreshes the access token before expiry.
type Client struct { ... }

// GetUserInfo retrieves basic user profile by OpenID.
// The access token is refreshed automatically if expired.
func (c *Client) GetUserInfo(ctx context.Context, openid string) (*UserInfo, error) { ... }
```

### Priority

| Priority | Scope |
|----------|-------|
| P0 | All `Client` structs and `NewClient()` constructors |
| P0 | All public API methods |
| P1 | Request/response structs and their fields |
| P2 | Package-level `package` doc comments |
| P3 | Constants, error variables |

### Scope

All 8 currently implemented packages. Sub-project C packages are written with Godoc from the start (no separate pass needed).

---

## Sub-project C: Three Stub Package Implementations

### Architecture Pattern

Follow the same patterns as `channels` and `mini-game`:
- `Config` struct + `NewClient()` constructor
- Automatic AccessToken caching and refresh via `TokenSource`
- `helper.go` for unified errcode checking
- One file per functional domain
- Godoc comments written inline (merged with D)
- Test files created alongside source files (target ≥ 70% coverage)

### C1: aispeech (智能对话)

**Base URL:** `https://openai.weixin.qq.com`

| File | Methods |
|------|---------|
| `client.go` | `Config`, `NewClient()`, token management |
| `helper.go` | `checkResponse()`, internal HTTP helpers |
| `asr.go` | `ASRLong(ctx, audioURL)`, `ASRShort(ctx, audioData)` |
| `tts.go` | `TextToSpeech(ctx, text, voiceType)` |
| `nlu.go` | `NLUUnderstand(ctx, query)`, `NLUIntentRecognize(ctx, query, skills)` |
| `dialog.go` | `DialogQuery(ctx, sessionID, query)`, `DialogReset(ctx, sessionID)` |

**Total: 7 methods**

### C2: mini-store (微信小店)

**Base URL:** `https://api.weixin.qq.com/shop/`

| File | Methods |
|------|---------|
| `client.go` | `Config`, `NewClient()`, token management |
| `helper.go` | `checkResponse()`, internal HTTP helpers |
| `product.go` | `AddProduct`, `UpdateProduct`, `DeleteProduct`, `GetProduct`, `ListProduct`, `UpdateProductStatus` |
| `order.go` | `GetOrder`, `ListOrder`, `UpdateOrderPrice`, `UpdateOrderExpressInfo` |
| `delivery.go` | `AddDeliveryInfo`, `GetDeliveryInfo` |
| `settlement.go` | `GetSettlementAccount`, `BindSettlementAccount`, `GetSettlementResult` |
| `coupon.go` | `CreateCoupon`, `UpdateCoupon`, `GetCoupon`, `ListCoupon`, `DeleteCoupon` |
| `after_sale.go` | `GetAfterSale`, `ListAfterSale`, `UpdateAfterSale` |

**Total: 24 methods**

### C3: xiaowei (微信硬件/IoT)

**Base URL:** `https://api.weixin.qq.com/hardware/`

| File | Methods |
|------|---------|
| `client.go` | `Config`, `NewClient()`, token management |
| `helper.go` | `checkResponse()`, internal HTTP helpers |
| `device.go` | `RegisterDevice`, `AuthorizeDevice`, `GetDeviceInfo`, `ListDevice` |
| `binding.go` | `BindDevice`, `UnbindDevice`, `GetBindUser` |
| `message.go` | `SendDeviceMessage`, `GetDeviceMessage` |
| `firmware.go` | `GetFirmwareInfo`, `CreateFirmware`, `SetFirmwareVersion` |

**Total: 12 methods**

---

## Summary

| Sub-project | Deliverables | New files | Modified files |
|------------|-------------|-----------|---------------|
| A | Bug fixes + error interface | 1 (`utils/wechat_error.go`) | ~60 |
| B | Test files | ~15 new test files | 0 |
| D | Godoc comments | 0 | ~140 |
| C | Three full packages | ~25 new files | 0 |

**Total new API methods added by C:** 43 (7 + 24 + 12)
