# Mini-Game (小游戏) Package — Design Spec

**Package:** `github.com/godrealms/go-wechat-sdk/mini-game`
**Date:** 2026-04-12
**Depends on:** `utils.HTTP`

---

## 1. Scope

12 new public methods + 2 private helpers:

| # | Method | HTTP | Path | Returns |
|---|---|---|---|---|
| 1 | Code2Session | GET | `/sns/jscode2session` | JSON |
| 2 | AccessToken | GET | `/cgi-bin/token` (cached) | string |
| 3 | GetDailySummary | POST | `/datacube/getweanalysisappiddailysummarytrend` | JSON |
| 4 | GetDailyRetain | POST | `/datacube/getweanalysisappiddailyretaininfo` | JSON |
| 5 | MsgSecCheck | POST | `/wxa/msg_sec_check` | JSON |
| 6 | CreateOrder | POST | `/wxa/game/createorder` | JSON |
| 7 | QueryOrder | POST | `/wxa/game/queryorder` | JSON |
| 8 | GetGameAdData | POST | `/wxa/game/getgameaddata` | JSON |
| 9 | CreateGameRoom | POST | `/wxa/game/createroom` | JSON |
| 10 | GetRoomInfo | POST | `/wxa/game/getroominfo` | JSON |
| 11 | SetUserStorage | POST | `/wxa/set_user_storage` | JSON |
| 12 | GetUserStorage | POST | `/wxa/get_user_storage` | JSON |

### 1.1 Private Helpers

| Helper | Purpose |
|---|---|
| `doPost(ctx, path, body, out)` | POST JSON with auto access_token + errcode check (json.RawMessage pattern) |

---

## 2. File Layout

| File | Content |
|---|---|
| `client.go` (rewrite) | Config, Client, NewClient, AccessToken, Code2Session, TokenSource, Option |
| `helper.go` (new) | baseResp, doPost |
| `client_test.go` (new) | TestNewClient, TestAccessToken, TestCode2Session |
| `analysis.go` (new) | GetDailySummary, GetDailyRetain + DTOs |
| `analysis_test.go` (new) | 2 tests |
| `security.go` (new) | MsgSecCheck + DTOs |
| `security_test.go` (new) | 1 test |
| `payment.go` (new) | CreateOrder, QueryOrder + DTOs |
| `payment_test.go` (new) | 2 tests |
| `ad.go` (new) | GetGameAdData + DTOs |
| `ad_test.go` (new) | 1 test |
| `frame_sync.go` (new) | CreateGameRoom, GetRoomInfo + DTOs |
| `frame_sync_test.go` (new) | 2 tests |
| `storage.go` (new) | SetUserStorage, GetUserStorage + DTOs |
| `storage_test.go` (new) | 2 tests |

---

## 3. Foundation

### 3.1 Client (`client.go`)

Same pattern as mini-program and channels:
```go
type Config struct { AppId, AppSecret string }
type Client struct { cfg, http, mu, accessToken, expiresAt, tokenSource }
type TokenSource interface { AccessToken(ctx) (string, error) }
func NewClient(cfg, opts...) (*Client, error)
func (c *Client) AccessToken(ctx) (string, error)  // double-checked locking, /cgi-bin/token
func (c *Client) Code2Session(ctx, jsCode) (*Code2SessionResp, error)
```

Error prefix: `mini_game:`

### 3.2 Helper (`helper.go`)

```go
type baseResp struct { ErrCode int; ErrMsg string }

func (c *Client) doPost(ctx, path, body, out) error
// Uses json.RawMessage → check errcode → unmarshal into out
```

---

## 4. DTOs & Methods

### 4.1 Analysis (`analysis.go`)

```go
type AnalysisDateReq struct {
    BeginDate string `json:"begin_date"`
    EndDate   string `json:"end_date"`
}

type DailySummaryItem struct {
    RefDate    string `json:"ref_date"`
    VisitTotal int64  `json:"visit_total"`
    SharePV    int64  `json:"share_pv"`
    ShareUV    int64  `json:"share_uv"`
}
type GetDailySummaryResp struct {
    List []DailySummaryItem `json:"list"`
}

type DailyRetainItem struct {
    DateKey string `json:"date_key"`
    Value   int    `json:"value"`
}
type GetDailyRetainResp struct {
    RefDate    string            `json:"ref_date"`
    VisitUVNew []DailyRetainItem `json:"visit_uv_new"`
    VisitUV    []DailyRetainItem `json:"visit_uv"`
}
```

### 4.2 Security (`security.go`)

```go
type MsgSecCheckReq struct {
    Content string `json:"content"`
    Version int    `json:"version"`
    Scene   int    `json:"scene"`
    OpenID  string `json:"openid"`
}
type SecCheckResult struct {
    Suggest string `json:"suggest"`
    Label   int    `json:"label"`
}
type MsgSecCheckResp struct {
    Result SecCheckResult `json:"result"`
}
```

### 4.3 Payment (`payment.go`)

```go
type CreateOrderReq struct {
    OpenID    string `json:"openid"`
    Env       int    `json:"env"`
    Zone      string `json:"zone_id"`
    ProductID string `json:"product_id"`
    Quantity  int    `json:"quantity"`
}
type CreateOrderResp struct {
    OrderID string `json:"order_id"`
    Balance int64  `json:"balance"`
}

type QueryOrderReq struct {
    OrderID string `json:"order_id"`
    OpenID  string `json:"openid"`
}
type QueryOrderResp struct {
    OrderID    string `json:"order_id"`
    Status     int    `json:"status"`
    PayAmount  int64  `json:"pay_amount"`
    CreateTime int64  `json:"create_time"`
}
```

### 4.4 Ad (`ad.go`)

```go
type GetGameAdDataReq struct {
    StartDate string `json:"start_date"`
    EndDate   string `json:"end_date"`
    AdUnitID  string `json:"ad_unit_id,omitempty"`
}
type GameAdData struct {
    Date       string `json:"date"`
    AdUnitID   string `json:"ad_unit_id"`
    ReqCount   int64  `json:"req_count"`
    ShowCount  int64  `json:"show_count"`
    ClickCount int64  `json:"click_count"`
    Income     int64  `json:"income"`
}
type GetGameAdDataResp struct {
    Items []GameAdData `json:"items"`
}
```

### 4.5 Frame Sync (`frame_sync.go`)

```go
type CreateGameRoomReq struct {
    MaxNum      int    `json:"max_num"`
    AccessInfo  string `json:"access_info,omitempty"`
}
type CreateGameRoomResp struct {
    RoomID string `json:"room_id"`
}

type GetRoomInfoReq struct {
    RoomID string `json:"room_id"`
}
type RoomMember struct {
    OpenID string `json:"openid"`
    Role   int    `json:"role"`
}
type GetRoomInfoResp struct {
    RoomID  string       `json:"room_id"`
    Status  int          `json:"status"`
    Members []RoomMember `json:"members"`
}
```

### 4.6 Storage (`storage.go`)

```go
type KVData struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

type SetUserStorageReq struct {
    OpenID   string   `json:"openid"`
    KVList   []KVData `json:"kv_list"`
    SigMethod string  `json:"sig_method"`
    Signature string  `json:"signature"`
}

type GetUserStorageReq struct {
    OpenID    string   `json:"openid"`
    KeyList   []string `json:"key_list"`
    SigMethod string   `json:"sig_method"`
    Signature string   `json:"signature"`
}
type GetUserStorageResp struct {
    KVList []KVData `json:"kv_list"`
}
```

---

## 5. Testing

Same httptest pattern. Helper `newTestClient` in `client_test.go`. All other test files reuse it.

## 6. Error Prefix

All errors: `mini_game:` prefix.
