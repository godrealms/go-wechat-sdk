# Phase 1: Core Base + Mini Program + Payment Extension Design Spec

## Context

The go-wechat-sdk currently has 2 implemented modules (offiaccount: 199 methods, merchant/developed: 14 methods) and 8 empty stubs. To scale the SDK to cover the full WeChat ecosystem, we need a shared foundation that eliminates code duplication across modules. Phase 1 establishes this foundation and implements the highest-value missing modules.

## Goal

1. Extract a `core/` base package from offiaccount's patterns (BaseClient, token management, response types)
2. Refactor offiaccount to embed core.BaseClient
3. Implement mini-program module with ~60 API endpoints
4. Extend merchant module with profitsharing, transfer, certificate, and bill sub-packages (~19 endpoints)
5. Add missing offiaccount APIs (content security, card management)

## Architecture

### core/ Package

The `core/` package provides the shared foundation all modules embed.

```
core/
├── client.go      # BaseClient struct + NewBaseClient factory
├── config.go      # BaseConfig struct
├── token.go       # AccessToken struct + refresh/cache logic with RWMutex
└── response.go    # Resp struct + error checking helpers
```

#### BaseConfig

```go
package core

type BaseConfig struct {
    AppId     string `json:"appId"`
    AppSecret string `json:"appSecret"`
}
```

#### BaseClient

```go
package core

import (
    "context"
    "fmt"
    "log"
    "net/url"
    "sync"
    "time"

    "github.com/godrealms/go-wechat-sdk/utils"
)

type AccessToken struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int64  `json:"expires_in"`
}

type BaseClient struct {
    Ctx         context.Context
    Config      *BaseConfig
    Https       *utils.HTTP
    accessToken *AccessToken
    tokenMutex  sync.RWMutex
    TokenURL    string // e.g. "/cgi-bin/token" for offiaccount, "/cgi-bin/stable_token" for mini-program
}

func NewBaseClient(ctx context.Context, config *BaseConfig, baseURL string, tokenURL string) *BaseClient {
    return &BaseClient{
        Ctx:      ctx,
        Config:   config,
        Https:    utils.NewHTTP(baseURL, utils.WithTimeout(time.Second*30)),
        TokenURL: tokenURL,
    }
}
```

**Token management methods on BaseClient:**
- `GetAccessToken() string` — public, returns cached token or refreshes
- `GetAccessTokenWithError() (string, error)` — returns error if refresh fails
- `getAccessToken() string` — private, double-checked locking with RWMutex
- `refreshAccessToken() (*AccessToken, error)` — calls TokenURL with appid+secret
- `tokenQuery(extra ...url.Values) url.Values` — returns url.Values with access_token merged

All these methods are identical to the current offiaccount implementation, just moved to core.

#### Resp (Base Response)

```go
package core

import "fmt"

type Resp struct {
    ErrCode int    `json:"errcode"`
    ErrMsg  string `json:"errmsg"`
}

func (r *Resp) GetError() error {
    if r.ErrCode != 0 {
        return fmt.Errorf("wechat api error %d: %s", r.ErrCode, r.ErrMsg)
    }
    return nil
}
```

### offiaccount/ Refactor

**Current:**
```go
type Client struct {
    ctx         context.Context
    Config      *Config
    Https       *utils.HTTP
    accessToken *AccessToken
    tokenMutex  sync.RWMutex
}
```

**After refactor:**
```go
package offiaccount

import "github.com/godrealms/go-wechat-sdk/core"

type Config struct {
    core.BaseConfig
    Token          string `json:"token"`
    EncodingAESKey string `json:"encodingAESKey"`
}

type Client struct {
    *core.BaseClient
    Token          string
    EncodingAESKey string
}

func NewClient(ctx context.Context, config *Config) *Client {
    base := core.NewBaseClient(ctx, &config.BaseConfig, "https://api.weixin.qq.com", "/cgi-bin/token")
    return &Client{
        BaseClient:     base,
        Token:          config.Token,
        EncodingAESKey: config.EncodingAESKey,
    }
}
```

**API compatibility:** All existing methods (GetAccessToken, tokenQuery, GetAccessTokenWithError) are now inherited from BaseClient. Existing API methods (api.*.go files) continue to work because they call `c.GetAccessToken()`, `c.tokenQuery()`, `c.Https.Get()`, etc., which are all promoted from the embedded BaseClient.

**Critical change:** All api.*.go files currently use `c.ctx` (private field). After refactor, this becomes `c.Ctx` (public field on BaseClient). A global find-replace of `c.ctx` to `c.Ctx` is needed across all 39 api files.

### mini-program/ Module

**Base URL:** `https://api.weixin.qq.com`
**Token URL:** `/cgi-bin/stable_token` (uses POST with JSON body, unlike offiaccount's GET)

```go
package miniprogram

import "github.com/godrealms/go-wechat-sdk/core"

type Config struct {
    core.BaseConfig
}

type Client struct {
    *core.BaseClient
}

func NewClient(ctx context.Context, config *Config) *Client {
    base := core.NewBaseClient(ctx, &config.BaseConfig, "https://api.weixin.qq.com", "/cgi-bin/stable_token")
    return &Client{BaseClient: base}
}
```

**Note on token refresh:** The mini-program stable_token API uses POST with JSON body `{"grant_type":"client_credential","appid":"...","secret":"..."}` instead of GET with query params. The BaseClient.refreshAccessToken needs to support both modes. Solution: add a `TokenMethod string` field to BaseClient ("GET" or "POST"), and branch in refreshAccessToken accordingly.

#### API Files

| File | Methods | WeChat API Endpoints |
|------|---------|---------------------|
| api.auth.go | Code2Session, CheckSessionKey, ResetSessionKey | /sns/jscode2session, /wxa/checksession, /wxa/resetusersessionkey |
| api.user.go | GetPhoneNumber, GetPaidUnionId | /wxa/business/getuserphonenumber, /wxa/getpaidunionid |
| api.wxacode.go | GetQRCode, GetUnlimited, CreateQRCode | /wxa/getwxacode, /wxa/getwxacodeunlimit, /cgi-bin/wxaapp/createwxaqrcode |
| api.scheme.go | GenerateScheme, QueryScheme | /wxa/generatescheme, /wxa/queryscheme |
| api.urllink.go | GenerateUrlLink, QueryUrlLink | /wxa/generate_urllink, /wxa/query_urllink |
| api.shortlink.go | GenerateShortLink | /wxa/genwxashortlink |
| api.subscribe.go | Send, GetTemplateList, GetCategory, AddTemplate, DeleteTemplate, GetPubTemplateKeyWords, GetPubTemplateTitles | /cgi-bin/message/subscribe/send, etc. |
| api.customer.go | SendCustomMessage, SetTyping | /cgi-bin/message/custom/send, /cgi-bin/message/custom/typing |
| api.security.go | MsgSecCheck, MediaCheckAsync, ImgSecCheck | /wxa/msg_sec_check, /wxa/media_check_async, /wxa/img_sec_check |
| api.analysis.go | GetDailySummary, GetDailyVisitTrend, GetWeeklyVisitTrend, GetMonthlyVisitTrend, GetDailyRetain, GetWeeklyRetain, GetMonthlyRetain, GetVisitPage, GetUserPortrait | /datacube/getweanalysisapp* |
| api.live.go | CreateRoom, GetLiveInfo, DeleteRoom, EditRoom, GetPushUrl, AddGoods, AddAssistant, ModifyAssistant | /wxaapi/broadcast/* |
| api.logistics.go | AddOrder, CancelOrder, GetOrder, GetPath, GetPrinter, GetQuota, GetContact, GetDelivery, UpdatePrinter | /cgi-bin/express/business/* |

#### Struct Files

| File | Types |
|------|-------|
| struct.base.go | Code2SessionResult, PhoneNumberResult, etc. |
| struct.wxacode.go | QRCodeRequest, UnlimitedQRCodeRequest, etc. |
| struct.subscribe.go | SubscribeMessage, TemplateInfo, etc. |
| struct.security.go | MsgSecCheckRequest, MediaCheckResult, etc. |
| struct.analysis.go | DailySummary, VisitTrend, RetainInfo, VisitPage, UserPortrait |
| struct.live.go | LiveRoom, LiveGoods, LiveAssistant, etc. |
| struct.logistics.go | LogisticsOrder, LogisticsPath, Delivery, etc. |

### merchant/ Extension

All new merchant sub-packages share the existing merchant Client pattern (builder with RSA signing). Each sub-package gets its own file organization but reuses the existing `merchant/developed/client.go` Client.

**Approach:** New sub-packages import and use the existing merchant Client rather than creating their own. They are organized as method files that could either:
- (A) Live as separate packages with their own Client wrapping merchant Client
- (B) Add methods directly to the existing merchant Client in new files

**Decision: Option A — separate sub-packages.** Each gets its own types and a thin Client that wraps the merchant Client for signing.

#### merchant/profitsharing/

```go
package profitsharing

type Client struct {
    MerchantClient *developed.Client
}

func NewClient(mc *developed.Client) *Client {
    return &Client{MerchantClient: mc}
}
```

| File | Methods |
|------|---------|
| api.go | OrdersCreate, OrdersQuery, OrdersUnfreeze, ReturnCreate, ReturnQuery, AddReceiver, DeleteReceiver, QueryMaxRatio |
| types.go | ProfitSharingOrder, ProfitSharingReceiver, ProfitSharingReturn, etc. |

#### merchant/transfer/

| File | Methods |
|------|---------|
| api.go | BatchesCreate, BatchesQuery, BatchesDetailQuery, ReceiptApply, ReceiptQuery, DetailReceiptApply, DetailReceiptQuery |
| types.go | TransferBatch, TransferDetail, TransferReceipt, etc. |

#### merchant/certificate/

| File | Methods |
|------|---------|
| api.go | GetCertificates |
| types.go | PlatformCertificate, EncryptCertificate |

#### merchant/bill/

| File | Methods |
|------|---------|
| api.go | TradeBillApply, FundFlowBillApply, SubMerchantFundFlowBillApply |
| types.go | BillRequest, BillResult |

### offiaccount/ New APIs

| File | Methods |
|------|---------|
| api.security.go | MsgSecCheck, ImgSecCheck, MediaCheckAsync |
| struct.security.go | MsgSecCheckRequest, ImgSecCheckRequest, MediaCheckAsyncRequest, MediaCheckAsyncResult |
| api.card.go | CreateCard, UpdateCard, DeleteCard, SetWhiteList, GetCardInfo, ModifyStock, GetCardList, BatchGetCard, DecryptCode, GetUserCardList |
| struct.card.go | Card, CardBaseInfo, CardAdvancedInfo, MemberCard, GrouponCard, CashCard, DiscountCard, GiftCard, GeneralCoupon |

## Error Handling

All modules follow the same pattern:
1. Check HTTP error from `c.Https.Post/Get` first
2. Check `result.ErrCode != 0` second
3. Use `core.Resp.GetError()` helper for consistent error wrapping

For merchant sub-packages, the error pattern follows WeChat Pay v3 conventions (HTTP status codes + JSON error body) which is already handled by the existing HTTP client.

## Testing Strategy

Each new package gets a `*_test.go` file with:
1. Client initialization tests (verify config, base URL, token URL)
2. Unit tests for any data transformation logic (e.g., signature generation, encryption/decryption)
3. Table-driven tests using httptest.NewServer for API call verification

Priority test areas:
- `core/token_test.go` — token caching, expiration, concurrent access
- `core/response_test.go` — error checking helper
- `miniprogram/api.auth_test.go` — Code2Session parsing
- `merchant/profitsharing/api_test.go` — signing and request format

## Migration / Backward Compatibility

The offiaccount refactor changes:
1. `c.ctx` (private) becomes `c.Ctx` (public, from BaseClient) — requires find-replace across all api files
2. `Config` embeds `core.BaseConfig` — `AppId` and `AppSecret` access unchanged
3. `AccessToken` type moves to core — internal, no public API change
4. All public methods (GetAccessToken, tokenQuery, etc.) are promoted from embedded BaseClient — callers see no difference

**Import path unchanged:** `github.com/godrealms/go-wechat-sdk/offiaccount` stays the same.

**The go.mod remains dependency-free** — no external packages needed.

## File Count Estimate

| Package | New Files | Modified Files |
|---------|-----------|----------------|
| core/ | 4 | 0 |
| offiaccount/ | 4 (security + card) | 40 (client.go + 39 api files for ctx rename) |
| mini-program/ | ~20 (client + 12 api + 7 struct) | 0 |
| merchant/profitsharing/ | 2 | 0 |
| merchant/transfer/ | 2 | 0 |
| merchant/certificate/ | 2 | 0 |
| merchant/bill/ | 2 | 0 |
| **Total** | **~36 new** | **~40 modified** |

## Verification

After implementation:
```bash
go build ./...                    # Full project builds
go test ./core/ -v                # Core tests pass
go test ./offiaccount/ -v         # Existing + new tests pass
go test ./mini-program/ -v        # Mini program tests pass
go test ./merchant/... -v         # All merchant tests pass
go vet ./...                      # No issues
```
