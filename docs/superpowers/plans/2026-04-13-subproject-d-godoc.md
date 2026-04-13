# Sub-project D: Godoc Comments Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add meaningful Godoc comments to all exported symbols across all 8 implemented packages.

**Architecture:** Per-package tasks; comments follow Go conventions (start with symbol name, describe what and key behavior, not the obvious); P0=Client structs + NewClient + API methods, P1=request/response types, P2=package doc, P3=constants.

**Tech Stack:** Go 1.23.1, godoc conventions

---

## Overview

Each task targets one package and follows the same shape:

1. Add/update package-level `doc.go` (or the existing package comment) — one sentence stating what the package provides.
2. Comment all exported types and functions — sentences begin with the symbol name and describe behaviour, not implementation.
3. Run `go doc` to verify comments are accepted and rendered.
4. Commit.

Priority legend used in each task:
- **P0** — Client, NewClient, primary API methods (must be done first)
- **P1** — Request/response struct types
- **P2** — Package-level doc comment
- **P3** — Constants, type aliases, minor helpers

---

## Task 1: `utils/` — HTTP client

**File:** `utils/http.go`

**Exported symbols requiring comments:**

| Symbol | Priority | Current state |
|--------|----------|---------------|
| `HTTP` struct | P0 | Chinese comment only: "HTTP 客户端结构体" |
| `Option` type | P1 | Chinese comment only |
| `NewHTTP` | P0 | Chinese comment only |
| `WithTimeout` | P1 | Chinese comment only |
| `WithHeaders` | P1 | Chinese comment only |
| `SetBaseURL` | P1 | Chinese comment only |
| `Get` | P0 | Chinese comment only |
| `Post` | P0 | Chinese comment only |
| `Put` | P0 | Chinese comment only |
| `Patch` | P0 | No comment |
| `Delete` | P0 | No comment |
| `PostForm` | P1 | No comment |

**Steps:**

- [ ] Open `utils/http.go`.

- [ ] Replace the current `HTTP` struct comment with an English Godoc comment:

```go
// HTTP is a thin JSON-over-HTTP client used by all SDK packages.
// It stores a base URL and optional default headers so callers never
// repeat boilerplate. Safe for concurrent use after construction.
type HTTP struct { ... }
```

- [ ] Replace the `Option` type comment:

```go
// Option is a functional configuration applied to HTTP during NewHTTP.
type Option func(*HTTP)
```

- [ ] Replace the `NewHTTP` comment:

```go
// NewHTTP constructs an HTTP client rooted at baseURL.
// All relative paths passed to Get/Post/etc. are joined to baseURL.
// Default timeout is 30 s; override with WithTimeout.
func NewHTTP(baseURL string, opts ...Option) *HTTP { ... }
```

- [ ] Replace `WithTimeout` comment:

```go
// WithTimeout overrides the default 30-second per-request deadline.
func WithTimeout(timeout time.Duration) Option { ... }
```

- [ ] Replace `WithHeaders` comment:

```go
// WithHeaders merges the provided key-value pairs into the client's
// default header set. Existing keys are overwritten.
func WithHeaders(headers map[string]string) Option { ... }
```

- [ ] Replace `SetBaseURL` comment:

```go
// SetBaseURL replaces the base URL after construction. Not goroutine-safe;
// call only before the client is shared between goroutines.
func (h *HTTP) SetBaseURL(url string) { ... }
```

- [ ] Add comment to `Get`:

```go
// Get sends a GET request to baseURL+path with the given query string
// and JSON-decodes the response body into result (if non-nil).
func (h *HTTP) Get(ctx context.Context, path string, query url.Values, result interface{}) error { ... }
```

- [ ] Add comment to `Post`:

```go
// Post sends a POST request with a JSON-encoded body to baseURL+path
// and JSON-decodes the response into result (if non-nil).
func (h *HTTP) Post(ctx context.Context, path string, body interface{}, result interface{}) error { ... }
```

- [ ] Add comment to `Put`:

```go
// Put sends a PUT request with a JSON-encoded body to baseURL+path
// and JSON-decodes the response into result (if non-nil).
func (h *HTTP) Put(ctx context.Context, path string, body interface{}, result interface{}) error { ... }
```

- [ ] Add comment to `Patch`:

```go
// Patch sends a PATCH request with a JSON-encoded body to baseURL+path
// and JSON-decodes the response into result (if non-nil).
func (h *HTTP) Patch(ctx context.Context, path string, body interface{}, result interface{}) error { ... }
```

- [ ] Add comment to `Delete`:

```go
// Delete sends a DELETE request to baseURL+path and JSON-decodes
// the response into result (if non-nil).
func (h *HTTP) Delete(ctx context.Context, path string, result interface{}) error { ... }
```

- [ ] Add comment to `PostForm`:

```go
// PostForm sends an application/x-www-form-urlencoded POST to baseURL+path
// and JSON-decodes the response into result (if non-nil).
func (h *HTTP) PostForm(ctx context.Context, path string, form url.Values, result interface{}) error { ... }
```

- [ ] Verify with:

```bash
go doc github.com/godrealms/go-wechat-sdk/utils
go doc github.com/godrealms/go-wechat-sdk/utils HTTP
go doc github.com/godrealms/go-wechat-sdk/utils NewHTTP
```

Expected: each symbol shows a rendered English one-liner or paragraph, no "missing comment" linter warnings.

- [ ] Commit:

```bash
git add utils/http.go
git commit -m "docs(utils): add English Godoc comments to HTTP client"
```

---

## Task 2: `offiaccount/` — WeChat Official Account

**Files:** `offiaccount/client.go`, `offiaccount/tokensource.go`, `offiaccount/api.base.go`, plus representative API files.

**Exported symbols requiring comments:**

| Symbol | File | Priority | Current state |
|--------|------|----------|---------------|
| `Config` struct | client.go | P1 | No comment |
| `Client` struct | client.go | P0 | Brief Chinese comment |
| `Option` type | client.go | P1 | Chinese comment |
| `WithTokenSource` | client.go | P1 | Chinese comment |
| `WithHTTPClient` | client.go | P1 | Chinese comment |
| `NewClient` | client.go | P0 | Chinese comment |
| `GetAccessToken` | client.go | P0 | Chinese comment |
| `AccessTokenE` | client.go | P0 | Chinese comment |
| `TokenSource` interface | tokensource.go | P1 | Chinese comment |
| `CheckResp` / `WeixinError` | struct.base.go | P1 | Missing |
| `GetStableAccessToken` | api.base.go | P0 | Chinese only |
| `CallbackCheck` | api.base.go | P0 | Chinese only |
| `GetCallbackIp` | api.base.go | P0 | Chinese only |
| `GetApiDomainIP` | api.base.go | P0 | Chinese only |
| `AccessToken` struct | struct.base.go | P1 | Missing |

**Before/after example for the three highest-priority symbols:**

```go
// BEFORE (no English comment):
type Client struct {
    ctx    context.Context
    Config *Config
    Https  *utils.HTTP
    ...
}

// AFTER:
// Client manages access token lifecycle and provides methods for the WeChat
// Official Account (公众号) server-side API. It is safe for concurrent use.
type Client struct { ... }
```

```go
// BEFORE:
// NewClient 创建客户端
func NewClient(ctx context.Context, config *Config, opts ...Option) *Client { ... }

// AFTER:
// NewClient creates an Official Account client. A background Context is used
// when ctx is nil. Call WithTokenSource to delegate token management to an
// open-platform authorizer.
func NewClient(ctx context.Context, config *Config, opts ...Option) *Client { ... }
```

```go
// BEFORE:
// AccessTokenE 显式获取 access_token，错误会被传递给调用方。
func (c *Client) AccessTokenE(ctx context.Context) (string, error) { ... }

// AFTER:
// AccessTokenE returns a valid access_token, propagating any fetch error to
// the caller. It uses an in-process read-write-locked cache; the token is
// refreshed 60 s before expiry. When a TokenSource is configured the call
// is delegated to it without touching /cgi-bin/token.
func (c *Client) AccessTokenE(ctx context.Context) (string, error) { ... }
```

**Steps:**

- [ ] Open `offiaccount/client.go`. Replace existing Chinese-only comments on `Config`, `Client`, `Option`, `WithTokenSource`, `WithHTTPClient`, `NewClient`, `GetAccessToken`, `AccessTokenE`, `getAccessToken` (mark deprecated), `refreshAccessToken` with English Godoc comments as shown above.

- [ ] Open `offiaccount/tokensource.go`. Update `TokenSource` interface comment:

```go
// TokenSource is an injectable access_token provider. When set on a Client
// via WithTokenSource, AccessTokenE delegates to it instead of calling
// /cgi-bin/token directly. Implement this interface to support open-platform
// component-on-behalf-of flows.
type TokenSource interface {
    AccessToken(ctx context.Context) (string, error)
}
```

- [ ] Open `offiaccount/struct.base.go`. Add comments to `AccessToken`, `WeixinError` (or `CheckResp`), and other exported structs that lack them. Example for `WeixinError`:

```go
// WeixinError wraps a non-zero WeChat errcode returned by any API endpoint.
// It implements the error interface and can be inspected with errors.As.
type WeixinError struct {
    ErrCode int
    ErrMsg  string
}
```

- [ ] Open `offiaccount/api.base.go`. Replace Chinese-only comments on `GetAccessToken`, `GetStableAccessToken`, `CallbackCheck`, `GetCallbackIp`, `GetApiDomainIP` with English Godoc. Keep Chinese text as an additional paragraph if it contains detail not present in the English summary:

```go
// GetStableAccessToken returns a stable access_token that is not invalidated
// by concurrent calls to the regular token endpoint. Pass forceRefresh=true
// to rotate the token immediately. Rate-limited to 10 000 calls/min, 500 000/day.
func (c *Client) GetStableAccessToken(forceRefresh bool) (*AccessToken, error) { ... }
```

- [ ] Add English summary comments to at least three additional API methods representative of different domains (e.g. one from `api.user.manage.userinfo.go`, one from `api.notify.template.go`, one from `api.qr-code.qr-codes.go`).

- [ ] Add a package-level doc comment to `offiaccount/client.go` (or a `doc.go`):

```go
// Package offiaccount provides a client for the WeChat Official Account
// (公众号) server-side API. Create a Client with NewClient, then call any
// of the API methods. Token refresh is automatic and concurrency-safe.
package offiaccount
```

- [ ] Verify:

```bash
go doc github.com/godrealms/go-wechat-sdk/offiaccount
go doc github.com/godrealms/go-wechat-sdk/offiaccount Client
go doc github.com/godrealms/go-wechat-sdk/offiaccount NewClient
go doc github.com/godrealms/go-wechat-sdk/offiaccount.Client AccessTokenE
```

- [ ] Commit:

```bash
git add offiaccount/
git commit -m "docs(offiaccount): add English Godoc comments to exported symbols"
```

---

## Task 3: `mini-program/` — WeChat Mini Program

**File:** `mini-program/client.go` (package `mini_program`), `mini-program/helper.go`, and all API files.

**Exported symbols requiring comments:**

| Symbol | Priority | Current state |
|--------|----------|---------------|
| Package doc | P2 | Present but Chinese mix |
| `Config` struct | P1 | Chinese only |
| `Client` struct | P0 | Chinese only |
| `Option` type | P1 | Chinese only |
| `WithHTTP` | P1 | Chinese only |
| `WithTokenSource` | P1 | Chinese only |
| `NewClient` | P0 | Chinese only |
| `HTTP` method | P1 | Chinese only |
| `Code2SessionResp` | P1 | Chinese only |
| `Code2Session` | P0 | Chinese only |
| `AccessToken` | P0 | Chinese only |
| `SendSubscribeMessage` | P0 | Chinese only |
| `TokenSource` | P1 | Present in mini-program/client.go |

**Steps:**

- [ ] Open `mini-program/client.go`. Update the package doc comment to be a clean English-first bilingual block:

```go
// Package mini_program provides a client for the WeChat Mini Program
// (小程序) server-side API. The primary entry point is NewClient.
//
// Implemented APIs:
//   - AccessToken: fetch and cache the global access_token
//   - Code2Session: exchange a wx.login js_code for openid and session_key
//   - SendSubscribeMessage: push a subscription message to a user
//
// For APIs not yet wrapped, use Client.HTTP().Get or Client.HTTP().Post directly.
package mini_program
```

- [ ] Add/update English Godoc to `Config`:

```go
// Config holds the Mini Program credentials obtained from the WeChat developer console.
type Config struct {
    AppId     string
    AppSecret string
}
```

- [ ] Add/update English Godoc to `Client`:

```go
// Client is the Mini Program server-side client. It caches the access_token
// in-process and refreshes it automatically 60 s before expiry.
// Safe for concurrent use.
type Client struct { ... }
```

- [ ] Add/update English Godoc to `NewClient`:

```go
// NewClient constructs a Mini Program client. Returns an error if AppId is
// empty or if AppSecret is empty and no TokenSource is provided.
func NewClient(cfg Config, opts ...Option) (*Client, error) { ... }
```

- [ ] Add/update English Godoc to `AccessToken`:

```go
// AccessToken returns a valid global access_token, refreshing it when fewer
// than 60 seconds remain before expiry. When a TokenSource is configured, the
// call is forwarded to it without contacting /cgi-bin/token.
func (c *Client) AccessToken(ctx context.Context) (string, error) { ... }
```

- [ ] Add/update English Godoc to `Code2SessionResp` and `Code2Session`:

```go
// Code2SessionResp is the response from the wx.login code exchange endpoint.
type Code2SessionResp struct { ... }

// Code2Session exchanges the js_code obtained from wx.login for the user's
// openid, session_key, and (if applicable) unionid.
// Returns an error if the WeChat server rejects the code.
func (c *Client) Code2Session(ctx context.Context, jsCode string) (*Code2SessionResp, error) { ... }
```

- [ ] Add/update English Godoc to `SendSubscribeMessage`:

```go
// SendSubscribeMessage sends a subscription template message to the user
// identified in body. The body must conform to the WeChat subscribe-message
// schema (touser, template_id, page, miniprogram_state, lang, data).
func (c *Client) SendSubscribeMessage(ctx context.Context, body any) error { ... }
```

- [ ] Update `WithHTTP` and `WithTokenSource` comments to English.

- [ ] Verify:

```bash
go doc github.com/godrealms/go-wechat-sdk/mini-program
go doc github.com/godrealms/go-wechat-sdk/mini-program Client
go doc github.com/godrealms/go-wechat-sdk/mini-program.Client Code2Session
go doc github.com/godrealms/go-wechat-sdk/mini-program.Client AccessToken
```

- [ ] Commit:

```bash
git add mini-program/
git commit -m "docs(mini-program): add English Godoc comments to exported symbols"
```

---

## Task 4: `mini-game/` — WeChat Mini Game

**Files:** `mini-game/client.go` (package `mini_game`), `mini-game/helper.go`, and all domain API files (`ad.go`, `analysis.go`, `frame_sync.go`, `payment.go`, `security.go`, `storage.go`).

**Exported symbols requiring comments:**

| Symbol | File | Priority | Current state |
|--------|------|----------|---------------|
| Package doc | client.go | P2 | Present, Chinese mix |
| `Config` | client.go | P1 | Chinese only |
| `Client` | client.go | P0 | Chinese only |
| `TokenSource` | client.go | P1 | Chinese only |
| `Option` | client.go | P1 | Chinese only |
| `WithHTTP` | client.go | P1 | Chinese only |
| `WithTokenSource` | client.go | P1 | Chinese only |
| `NewClient` | client.go | P0 | Chinese only |
| `HTTP` | client.go | P1 | Chinese only |
| `Code2SessionResp` | client.go | P1 | Chinese only |
| `Code2Session` | client.go | P0 | Chinese only |
| `AccessToken` | client.go | P0 | Chinese only |
| `GetGameAdDataReq/Resp` | ad.go | P1 | No comment |
| `GetGameAdData` | ad.go | P0 | No comment |
| All structs in analysis.go | analysis.go | P1 | No comment |
| All methods in analysis.go | analysis.go | P0 | No comment |
| All structs/methods in frame_sync.go | frame_sync.go | P0/P1 | No comment |
| All structs/methods in payment.go | payment.go | P0/P1 | No comment |
| All structs/methods in security.go | security.go | P0/P1 | No comment |
| All structs/methods in storage.go | storage.go | P0/P1 | No comment |

**Steps:**

- [ ] Open `mini-game/client.go`. Update package doc to English-first:

```go
// Package mini_game provides a client for the WeChat Mini Game (小游戏)
// server-side API. The primary entry point is NewClient.
//
// Implemented APIs:
//   - AccessToken: fetch and cache the global access_token
//   - Code2Session: exchange a wx.login js_code for openid/session_key
//   - GetGameAdData: ad performance data
//   - GetAnalysisSummary / GetAnalysisTrend: data analytics
//   - FrameSync APIs: real-time frame sync room management
//   - Payment APIs: virtual payment queries
//   - Security: content risk detection
//   - CloudStorage: key-value cloud storage
package mini_game
```

- [ ] Add English Godoc to `Config`, `Client`, `TokenSource`, `Option`, `WithHTTP`, `WithTokenSource`, `NewClient`, `HTTP`, `Code2SessionResp`, `Code2Session`, `AccessToken` using the same pattern as mini-program (see Task 3 for examples — substitute package prefix `mini_game`).

- [ ] Open `mini-game/ad.go`. Add comments:

```go
// GetGameAdDataReq is the request for fetching ad performance data.
// StartDate and EndDate are formatted as "YYYYMMDD".
type GetGameAdDataReq struct { ... }

// GameAdData holds ad metrics for a single (date, ad unit) pair.
type GameAdData struct { ... }

// GetGameAdDataResp is the response from GetGameAdData.
type GetGameAdDataResp struct { ... }

// GetGameAdData retrieves ad performance statistics for the given date range.
// AdUnitID is optional; omit to fetch across all ad units.
func (c *Client) GetGameAdData(ctx context.Context, req *GetGameAdDataReq) (*GetGameAdDataResp, error) { ... }
```

- [ ] For each remaining domain file (`analysis.go`, `frame_sync.go`, `payment.go`, `security.go`, `storage.go`): add a one-sentence English Godoc to every exported struct and method. Each method comment must start with the method name and state what the WeChat API does, e.g.:

```go
// GetAnalysisSummary returns the summary data analytics for the mini game
// over the given date range (at most 30 days).
func (c *Client) GetAnalysisSummary(ctx context.Context, req *GetAnalysisSummaryReq) (*GetAnalysisSummaryResp, error) { ... }
```

- [ ] Verify:

```bash
go doc github.com/godrealms/go-wechat-sdk/mini-game
go doc github.com/godrealms/go-wechat-sdk/mini-game Client
go doc github.com/godrealms/go-wechat-sdk/mini-game.Client GetGameAdData
go doc github.com/godrealms/go-wechat-sdk/mini-game.Client AccessToken
```

- [ ] Commit:

```bash
git add mini-game/
git commit -m "docs(mini-game): add English Godoc comments to all exported symbols"
```

---

## Task 5: `channels/` — WeChat Channels (视频号)

**Files:** `channels/client.go`, `channels/helper.go`, `channels/product.go`, `channels/order.go`, `channels/live.go`, `channels/data.go`.

**Exported symbols requiring comments:**

| Symbol | File | Priority | Current state |
|--------|------|----------|---------------|
| Package doc | client.go | P2 | Chinese only |
| `Config` | client.go | P1 | Chinese only |
| `Client` | client.go | P0 | Chinese only |
| `TokenSource` | client.go | P1 | Chinese comment, no English |
| `Option` | client.go | P1 | Chinese only |
| `WithHTTP` | client.go | P1 | Chinese only |
| `WithTokenSource` | client.go | P1 | Chinese only |
| `NewClient` | client.go | P0 | Chinese only |
| `HTTP` method | client.go | P1 | Chinese only |
| `AccessToken` | client.go | P0 | Chinese only |
| All structs in product.go | product.go | P1 | No comments |
| `AddProduct` | product.go | P0 | Chinese only |
| `UpdateProduct` | product.go | P0 | Chinese only |
| `GetProduct` | product.go | P0 | Chinese only |
| `ListProduct` | product.go | P0 | Chinese only |
| `DeleteProduct` | product.go | P0 | Chinese only |
| `OrderInfo` et al. | order.go | P1 | No comments |
| `GetOrder` | order.go | P0 | Chinese only |
| `ListOrder` | order.go | P0 | Chinese only |
| All symbols in live.go | live.go | P0/P1 | No comments |
| All symbols in data.go | data.go | P0/P1 | No comments |

**Steps:**

- [ ] Open `channels/client.go`. Update package comment to English-first:

```go
// Package channels provides a client for the WeChat Channels (视频号)
// e-commerce and live-streaming server-side API.
// Create a Client with NewClient; token refresh is automatic.
package channels
```

- [ ] Update `Config` comment:

```go
// Config holds the Channels app credentials from the WeChat developer console.
type Config struct { ... }
```

- [ ] Update `Client` comment:

```go
// Client is the Channels server-side client. It caches the access_token
// in-process with a read-write lock and refreshes 60 s before expiry.
// Safe for concurrent use.
type Client struct { ... }
```

- [ ] Update `TokenSource` comment:

```go
// TokenSource is an injectable access_token provider. Configure it via
// WithTokenSource to use an open-platform component-on-behalf-of token
// instead of fetching one from /cgi-bin/token.
type TokenSource interface { ... }
```

- [ ] Update `NewClient`, `WithHTTP`, `WithTokenSource`, `HTTP`, `AccessToken` with the same English-first pattern.

- [ ] Open `channels/product.go`. Add one-sentence Godoc to every struct (`ProductInfo`, `AddProductReq`, `AddProductResp`, `UpdateProductReq`, `GetProductReq`, `GetProductResp`, `ListProductReq`, `ListProductResp`, `DeleteProductReq`) and expand method comments:

```go
// ProductInfo describes a Channels e-commerce product record.
type ProductInfo struct { ... }

// AddProductReq is the request body for AddProduct.
type AddProductReq struct { ... }

// AddProductResp is the response from AddProduct, containing the assigned product ID.
type AddProductResp struct { ... }

// AddProduct creates a new product in the Channels e-commerce catalog.
// Returns the WeChat-assigned product_id on success.
func (c *Client) AddProduct(ctx context.Context, req *AddProductReq) (*AddProductResp, error) { ... }

// UpdateProduct updates fields on an existing product. Only non-zero fields in
// req.Product are applied; ProductID must be set.
func (c *Client) UpdateProduct(ctx context.Context, req *UpdateProductReq) error { ... }

// GetProduct retrieves the full product details for the given product_id.
func (c *Client) GetProduct(ctx context.Context, req *GetProductReq) (*GetProductResp, error) { ... }

// ListProduct returns a paginated list of products, optionally filtered by Status.
func (c *Client) ListProduct(ctx context.Context, req *ListProductReq) (*ListProductResp, error) { ... }

// DeleteProduct permanently removes the product with the given product_id.
func (c *Client) DeleteProduct(ctx context.Context, req *DeleteProductReq) error { ... }
```

- [ ] Open `channels/order.go`. Add Godoc to `OrderInfo`, `GetOrderReq`, `GetOrderResp`, `ListOrderReq`, `ListOrderResp`, `GetOrder`, `ListOrder` using the same approach.

- [ ] For each remaining file (`live.go`, `data.go`): add a one-sentence Godoc to every exported struct and method.

- [ ] Verify:

```bash
go doc github.com/godrealms/go-wechat-sdk/channels
go doc github.com/godrealms/go-wechat-sdk/channels Client
go doc github.com/godrealms/go-wechat-sdk/channels.Client AddProduct
go doc github.com/godrealms/go-wechat-sdk/channels.Client GetOrder
go doc github.com/godrealms/go-wechat-sdk/channels.Client AccessToken
```

- [ ] Commit:

```bash
git add channels/
git commit -m "docs(channels): add English Godoc comments to all exported symbols"
```

---

## Task 6: `oplatform/` — Open Platform (开放平台)

**Files:** `oplatform/client.go`, `oplatform/store.go`, `oplatform/errors.go`, `oplatform/component.token.go`, `oplatform/component.authorize.go`, `oplatform/component.preauth.go`, `oplatform/authorizer.go`, `oplatform/component.authorizer.token.go`, `oplatform/notify.go`, `oplatform/fastregister.go`, `oplatform/qrlogin.go`, and the `wxa.*` files.

**Exported symbols requiring comments (selected):**

| Symbol | File | Priority | Current state |
|--------|------|----------|---------------|
| Package doc | client.go | P2 | Missing |
| `Config` | client.go | P1 | Chinese only |
| `Client` | client.go | P0 | Chinese only |
| `Option` | client.go | P1 | Chinese only |
| `WithStore` | client.go | P1 | Chinese only |
| `WithHTTP` | client.go | P1 | Chinese only |
| `NewClient` | client.go | P0 | Chinese only |
| `Store` accessor | client.go | P1 | Chinese only |
| `HTTP` accessor | client.go | P1 | Chinese only |
| `ComponentAppID` | client.go | P1 | Chinese only |
| `Store` interface | store.go | P1 | Likely Chinese only |
| `MemoryStore` | store.go | P1 | Likely Chinese only |
| `WeixinError` | errors.go | P1 | Likely missing |
| `ComponentAccessToken` method | component.token.go | P0 | Chinese only |
| `GetPreAuthCode` | component.preauth.go | P0 | Chinese only |
| `HandleCallback` / `ParseNotify` | notify.go | P0 | Chinese only |
| `AuthorizerClient` | authorizer.go | P0 | Chinese only |
| `WxaClient` | wxa.client.go | P0 | Chinese only |

**Steps:**

- [ ] Add package-level doc to `oplatform/client.go` (or create `oplatform/doc.go`):

```go
// Package oplatform provides a client for the WeChat Open Platform
// (开放平台) third-party component API. Create a Client with NewClient,
// handle push callbacks via ParseNotify/HandleCallback, and obtain
// per-authorizer clients via AuthorizerClient or WxaClient.
package oplatform
```

- [ ] Update `Config` comment:

```go
// Config holds the credentials for a third-party platform component registered
// on the WeChat Open Platform.
type Config struct {
    ComponentAppID     string // third-party platform appid
    ComponentAppSecret string // third-party platform secret
    Token              string // callback signature token
    EncodingAESKey     string // 43-character AES key for callback decryption
}
```

- [ ] Update `Client` comment:

```go
// Client is the Open Platform third-party component client. It manages the
// component access_token, handles encrypted push notifications, and provides
// per-authorizer sub-clients. Safe for concurrent use.
type Client struct { ... }
```

- [ ] Update `NewClient`:

```go
// NewClient validates cfg and constructs a Client. No network requests are
// made during construction. Returns an error if any required field is empty
// or if EncodingAESKey is invalid.
func NewClient(cfg Config, opts ...Option) (*Client, error) { ... }
```

- [ ] For `store.go`: add Godoc to `Store` interface describing the contract (what callers must implement), and to `MemoryStore` (thread-safe in-memory default).

- [ ] For `errors.go`: add Godoc to `WeixinError`:

```go
// WeixinError represents a non-zero errcode returned by any WeChat API call.
// Use errors.As to inspect it after a failed API call.
type WeixinError struct { ... }
```

- [ ] For `component.token.go`: add English Godoc to the component-token fetch method, explaining that it requires a component_verify_ticket pushed by WeChat.

- [ ] For `authorizer.go`: add English Godoc to `AuthorizerClient` struct and its constructor, explaining that it returns a sub-client scoped to one authorizer appid.

- [ ] For `wxa.client.go`: add English Godoc to `WxaClient` and its constructor.

- [ ] For `notify.go`: add English Godoc to `ParseNotify` / `HandleCallback` explaining that they decrypt the XML envelope WeChat posts to the callback URL.

- [ ] For `fastregister.go`, `qrlogin.go`, and all remaining exported symbols in `component.*.go` and `wxa.*.go`: add at least a one-sentence English comment per exported function.

- [ ] Verify:

```bash
go doc github.com/godrealms/go-wechat-sdk/oplatform
go doc github.com/godrealms/go-wechat-sdk/oplatform Client
go doc github.com/godrealms/go-wechat-sdk/oplatform NewClient
go doc github.com/godrealms/go-wechat-sdk/oplatform.Client ComponentAccessToken
```

- [ ] Commit:

```bash
git add oplatform/
git commit -m "docs(oplatform): add English Godoc comments to all exported symbols"
```

---

## Task 7: `work-wechat/isv/` — Work WeChat ISV (企业微信服务商)

**Files:** `work-wechat/isv/client.go`, `work-wechat/isv/store.go`, `work-wechat/isv/errors.go`, `work-wechat/isv/tokensource.go`, `work-wechat/isv/notify.go`, `work-wechat/isv/oauth2.go`, `work-wechat/isv/suite.token.go`, `work-wechat/isv/suite.preauth.go`, `work-wechat/isv/suite.permanent.go`, `work-wechat/isv/provider.login.go`, `work-wechat/isv/provider.id_convert.go`, `work-wechat/isv/corp.*.go`, `work-wechat/isv/data_notify.go`.

**Exported symbols requiring comments (selected):**

| Symbol | File | Priority | Current state |
|--------|------|----------|---------------|
| Package doc | doc.go | P2 | Present or missing |
| `Config` | client.go | P1 | Chinese comment |
| `Client` | client.go | P0 | Chinese comment |
| `Option` | client.go | P1 | No comment |
| `WithStore` | client.go | P1 | No comment |
| `WithHTTPClient` | client.go | P1 | No comment |
| `WithBaseURL` | client.go | P1 | No comment |
| `NewClient` | client.go | P0 | Chinese comment |
| `GetSuiteAccessToken` | suite.token.go | P0 | Likely Chinese only |
| `GetCorpAccessToken` | corp.token.go | P0 | Likely Chinese only |
| `ParseNotify` | notify.go | P0 | Likely Chinese only |
| `OAuth2AuthURL` | oauth2.go | P0 | Likely Chinese only |
| `OAuth2GetUserInfo` | oauth2.go | P0 | Likely Chinese only |
| `ProviderLogin` | provider.login.go | P0 | Likely Chinese only |
| `Store` interface | store.go | P1 | Likely Chinese only |
| `WeixinError` | errors.go | P1 | Likely missing |

**Steps:**

- [ ] Ensure `work-wechat/isv/doc.go` exists with:

```go
// Package isv provides a Work WeChat (企业微信) ISV (independent software vendor)
// client that manages suite tokens, permanent auth codes, per-corp access tokens,
// and push-notification verification. Create a Client with NewClient.
package isv
```

- [ ] Open `work-wechat/isv/client.go`. Update English Godoc on `Config`:

```go
// Config is the runtime configuration for the ISV Client.
// SuiteID and SuiteSecret identify the Work WeChat third-party application.
// Token and EncodingAESKey are used to verify and decrypt callback payloads.
// ProviderCorpID and ProviderSecret must both be set (or both empty) for
// provider-level APIs.
type Config struct { ... }
```

- [ ] Update `Client`:

```go
// Client is the ISV service-provider entry point. It holds no per-request state
// and is safe to share across goroutines. Use WithStore to plug in a persistent
// token store; the default is an in-memory store suitable for single-process use.
type Client struct { ... }
```

- [ ] Update `Option`, `WithStore`, `WithHTTPClient`, `WithBaseURL`, `NewClient` with English Godoc.

- [ ] Open `work-wechat/isv/suite.token.go`. Add English Godoc to `GetSuiteAccessToken`:

```go
// GetSuiteAccessToken returns a valid suite_access_token, using the cached
// value when possible. A suite_access_token requires a component_verify_ticket
// stored via the push-notification handler; returns an error if no ticket is
// available.
func (c *Client) GetSuiteAccessToken(ctx context.Context) (string, error) { ... }
```

- [ ] Open `work-wechat/isv/corp.token.go`. Add English Godoc to `GetCorpAccessToken`:

```go
// GetCorpAccessToken returns a valid corp access_token for the authorizing
// corporation identified by corpID. The permanent auth code must have been
// stored previously via the PermanentAuth flow.
func (c *Client) GetCorpAccessToken(ctx context.Context, corpID string) (string, error) { ... }
```

- [ ] For `notify.go`: add English Godoc to `ParseNotify` / `HandleCallback` explaining XML decryption and signature verification.

- [ ] For `oauth2.go`, `provider.login.go`, `provider.id_convert.go`, `suite.preauth.go`, `suite.permanent.go`, `data_notify.go`: add at minimum a one-sentence English Godoc to every exported function.

- [ ] For all `corp.*.go` files: add a one-sentence English Godoc to every exported struct type and method. Focus on what the WeChat API does, not the implementation.

- [ ] Verify:

```bash
go doc github.com/godrealms/go-wechat-sdk/work-wechat/isv
go doc github.com/godrealms/go-wechat-sdk/work-wechat/isv Client
go doc github.com/godrealms/go-wechat-sdk/work-wechat/isv NewClient
go doc github.com/godrealms/go-wechat-sdk/work-wechat/isv.Client GetSuiteAccessToken
go doc github.com/godrealms/go-wechat-sdk/work-wechat/isv.Client GetCorpAccessToken
```

- [ ] Commit:

```bash
git add work-wechat/isv/
git commit -m "docs(work-wechat/isv): add English Godoc comments to all exported symbols"
```

---

## Task 8: `merchant/developed/` — WeChat Pay Merchant (微信支付)

**Files:** `merchant/developed/client.go`, `merchant/developed/pay.transactions.jsapi.go`, `merchant/developed/pay.transactions.app.go`, `merchant/developed/pay.transactions.h5.go`, `merchant/developed/pay.transactions.native.go`, `merchant/developed/pay.transactions.query.go`, `merchant/developed/pay.transactions.close.go`, `merchant/developed/pay.transactions.refunds.go`, `merchant/developed/pay.transactions.bill.go`, `merchant/developed/complaint.go`, `merchant/developed/transfer.go`, `merchant/developed/notify.go`.

**Package name:** `wechat` (package `wechat` inside `merchant/developed/`).

**Exported symbols requiring comments:**

| Symbol | File | Priority | Current state |
|--------|------|----------|---------------|
| Package doc | client.go | P2 | Missing |
| `Client` struct | client.go | P0 | Brief Chinese comment |
| `NewWechatClient` | client.go | P0 | No comment |
| `WithAppid` … `WithHttp` builder methods | client.go | P1 | No comment |
| `TransactionsJsapi` | pay.transactions.jsapi.go | P0 | No comment |
| `ModifyTransactionsJsapi` | pay.transactions.jsapi.go | P0 | Chinese comment |
| `TransactionsApp` | pay.transactions.app.go | P0 | No/brief comment |
| `TransactionsH5` | pay.transactions.h5.go | P0 | No/brief comment |
| `TransactionsNative` | pay.transactions.native.go | P0 | No/brief comment |
| `QueryTransactionByID` | pay.transactions.query.go | P0 | No/brief comment |
| `QueryTransactionByOutTradeNo` | pay.transactions.query.go | P0 | No/brief comment |
| `CloseTransaction` | pay.transactions.close.go | P0 | No/brief comment |
| `Refund` | pay.transactions.refunds.go | P0 | No/brief comment |
| `TradeBill` / `FundFlowBill` | pay.transactions.bill.go | P0 | No/brief comment |
| `GetComplaintList` | complaint.go | P0 | No/brief comment |
| `Transfer` | transfer.go | P0 | No/brief comment |
| `ParseNotify` | notify.go | P0 | No/brief comment |

**Steps:**

- [ ] Add package doc to `merchant/developed/client.go` (or create `merchant/developed/doc.go`):

```go
// Package wechat provides a client for the WeChat Pay v3 merchant API.
// Construct a Client using the builder pattern starting with NewWechatClient,
// then chaining With* methods for credentials and keys. All payment endpoints
// use WECHATPAY2-SHA256-RSA2048 request signing.
package wechat
```

- [ ] Update `Client` comment:

```go
// Client is the WeChat Pay v3 merchant client. It holds the merchant
// credentials and RSA keys required for request signing and response
// verification. Construct it with NewWechatClient and the With* builder
// methods. Safe for concurrent use once fully configured.
type Client struct { ... }
```

- [ ] Add Godoc to `NewWechatClient`:

```go
// NewWechatClient constructs a Client with the default HTTPS base URL
// (https://api.mch.weixin.qq.com). Call the With* methods to supply
// credentials before making any API calls.
func NewWechatClient() *Client { ... }
```

- [ ] Add one-sentence Godoc to each builder method (`WithAppid`, `WithMchid`, `WithCertificateNumber`, `WithAPIv3Key`, `WithCertificate`, `WithPrivateKey`, `WithPublicKey`, `WithHttp`). Each returns `*Client` for chaining.

- [ ] Open `pay.transactions.jsapi.go`. Add Godoc to `TransactionsJsapi` and `ModifyTransactionsJsapi`:

```go
// TransactionsJsapi initiates a JSAPI (公众号/小程序) payment order and
// returns the WeChat prepay_id. The caller must sign the resulting
// parameters with the merchant private key before returning them to the client.
func (c *Client) TransactionsJsapi(order *types.Transactions) (*types.TransactionsJsapiResp, error) { ... }

// ModifyTransactionsJsapi is a convenience wrapper around TransactionsJsapi
// that returns the signed JSAPI parameters ready to pass to wx.requestPayment.
func (c *Client) ModifyTransactionsJsapi(order *types.Transactions) (*types.TransactionsJsapi, error) { ... }
```

- [ ] Open `pay.transactions.app.go`. Add Godoc to `TransactionsApp`:

```go
// TransactionsApp initiates an App payment order and returns the signed
// parameters required by the WeChat Pay SDK for mobile apps.
func (c *Client) TransactionsApp(order *types.Transactions) (*types.TransactionsAppResp, error) { ... }
```

- [ ] For each remaining payment file, add a one-sentence English Godoc to every exported function describing which WeChat Pay v3 endpoint it calls and what it returns.

- [ ] Open `complaint.go`. Add Godoc to complaint-related methods:

```go
// GetComplaintList returns a paginated list of merchant complaints filed
// within the given date range. Use the returned next_offset for pagination.
func (c *Client) GetComplaintList(...) { ... }
```

- [ ] Open `transfer.go` and `notify.go`. Add similar English Godoc.

- [ ] Verify:

```bash
go doc github.com/godrealms/go-wechat-sdk/merchant/developed
go doc github.com/godrealms/go-wechat-sdk/merchant/developed Client
go doc github.com/godrealms/go-wechat-sdk/merchant/developed NewWechatClient
go doc github.com/godrealms/go-wechat-sdk/merchant/developed.Client TransactionsJsapi
go doc github.com/godrealms/go-wechat-sdk/merchant/developed.Client ModifyTransactionsJsapi
```

- [ ] Commit:

```bash
git add merchant/developed/
git commit -m "docs(merchant/developed): add English Godoc comments to all exported symbols"
```

---

## Final Verification

After all 8 tasks are complete, run the full suite to confirm nothing broke:

- [ ] `go build ./...` — zero errors.
- [ ] `go test ./...` — all existing tests pass.
- [ ] `go vet ./...` — zero warnings.
- [ ] Spot-check rendered docs: `go doc -all github.com/godrealms/go-wechat-sdk/channels | grep -c "^func"` — count should match the number of exported methods.

- [ ] Final commit (if any stray files remain):

```bash
git add -p   # review each hunk
git commit -m "docs: final Godoc cleanup across all packages"
```
