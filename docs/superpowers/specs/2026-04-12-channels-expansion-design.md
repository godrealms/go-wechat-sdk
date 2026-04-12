# Channels (视频号) Package — Design Spec

**Package:** `github.com/godrealms/go-wechat-sdk/channels`
**Date:** 2026-04-12
**Depends on:** `utils.HTTP`

---

## 1. Scope

13 new public methods + 2 private helpers across 4 areas:

| # | Method | HTTP | Path | Returns |
|---|---|---|---|---|
| 1 | CreateRoom | POST | `/channels/ec/basics/live/createroom` | JSON |
| 2 | DeleteRoom | POST | `/channels/ec/basics/live/deleteroom` | JSON |
| 3 | GetLiveInfo | POST | `/channels/ec/basics/live/getliveinfo` | JSON |
| 4 | GetLiveReplayList | POST | `/channels/ec/basics/live/getlivereplaylist` | JSON |
| 5 | AddProduct | POST | `/channels/ec/product/add` | JSON |
| 6 | UpdateProduct | POST | `/channels/ec/product/update` | JSON |
| 7 | GetProduct | POST | `/channels/ec/product/get` | JSON |
| 8 | ListProduct | POST | `/channels/ec/product/list` | JSON |
| 9 | DeleteProduct | POST | `/channels/ec/product/delete` | JSON |
| 10 | GetOrder | POST | `/channels/ec/order/get` | JSON |
| 11 | ListOrder | POST | `/channels/ec/order/list` | JSON |
| 12 | GetFinderLiveDataList | POST | `/channels/ec/basics/getfinderlivedata` | JSON |
| 13 | GetFinderList | POST | `/channels/ec/basics/getfinderlist` | JSON |

### 1.1 Private Helpers

| Helper | Purpose |
|---|---|
| `doGet(ctx, path, extra, out)` | GET with auto access_token |
| `doPost(ctx, path, body, out)` | POST JSON with auto access_token + errcode check |

All 13 APIs are POST-only. `doGet` is included for future extensibility.

---

## 2. File Layout

| File | Content |
|---|---|
| `client.go` (rewrite) | Config, Client, NewClient, AccessToken, Option, WithHTTP, WithTokenSource |
| `helper.go` (new) | baseResp, doGet, doPost |
| `client_test.go` (new) | TestNewClient, TestAccessToken |
| `live.go` (new) | CreateRoomReq/Resp, DeleteRoomReq, GetLiveInfoReq/Resp, GetLiveReplayListReq/Resp + DTOs; 4 methods |
| `live_test.go` (new) | 4 tests |
| `product.go` (new) | AddProductReq/Resp, UpdateProductReq, GetProductReq/Resp, ListProductReq/Resp, DeleteProductReq + DTOs; 5 methods |
| `product_test.go` (new) | 5 tests |
| `order.go` (new) | GetOrderReq/Resp, ListOrderReq/Resp + DTOs; 2 methods |
| `order_test.go` (new) | 2 tests |
| `data.go` (new) | GetFinderLiveDataListReq/Resp, GetFinderListReq/Resp + DTOs; 2 methods |
| `data_test.go` (new) | 2 tests |

---

## 3. Foundation

### 3.1 Client (`client.go`)

```go
package channels

type Config struct {
    AppId     string
    AppSecret string
}

type Client struct {
    cfg  Config
    http *utils.HTTP
    mu          sync.RWMutex
    accessToken string
    expiresAt   time.Time
    tokenSource TokenSource
}

type TokenSource interface {
    AccessToken(ctx context.Context) (string, error)
}

type Option func(*Client)

func WithHTTP(h *utils.HTTP) Option
func WithTokenSource(ts TokenSource) Option
func NewClient(cfg Config, opts ...Option) (*Client, error)
func (c *Client) AccessToken(ctx context.Context) (string, error)
func (c *Client) HTTP() *utils.HTTP
```

Same pattern as mini-program Client.

### 3.2 Helper (`helper.go`)

```go
type baseResp struct {
    ErrCode int    `json:"errcode"`
    ErrMsg  string `json:"errmsg"`
}

func (c *Client) doGet(ctx, path, extra, out) error  // GET + access_token
func (c *Client) doPost(ctx, path, body, out) error   // POST JSON + access_token + errcode
```

---

## 4. DTOs & Methods

### 4.1 Live (`live.go`)

```go
type CreateRoomReq struct {
    Name      string `json:"name"`
    CoverImg  string `json:"cover_img,omitempty"`
    StartTime int64  `json:"start_time"`
    EndTime   int64  `json:"end_time"`
}
type CreateRoomResp struct {
    RoomID string `json:"room_id"`
}

type DeleteRoomReq struct {
    RoomID string `json:"room_id"`
}

type GetLiveInfoReq struct {
    RoomID string `json:"room_id"`
}
type LiveInfo struct {
    RoomID    string `json:"room_id"`
    Name      string `json:"name"`
    Status    int    `json:"status"`
    StartTime int64  `json:"start_time"`
    EndTime   int64  `json:"end_time"`
}
type GetLiveInfoResp struct {
    LiveInfo LiveInfo `json:"live_info"`
}

type GetLiveReplayListReq struct {
    RoomID string `json:"room_id"`
    Offset int    `json:"offset,omitempty"`
    Limit  int    `json:"limit,omitempty"`
}
type LiveReplay struct {
    MediaURL   string `json:"media_url"`
    ExpireTime int64  `json:"expire_time"`
    CreateTime int64  `json:"create_time"`
}
type GetLiveReplayListResp struct {
    LiveReplayList []LiveReplay `json:"live_replay_list"`
    Total          int          `json:"total"`
}
```

Methods:
- `CreateRoom(ctx, *CreateRoomReq) (*CreateRoomResp, error)`
- `DeleteRoom(ctx, *DeleteRoomReq) error`
- `GetLiveInfo(ctx, *GetLiveInfoReq) (*GetLiveInfoResp, error)`
- `GetLiveReplayList(ctx, *GetLiveReplayListReq) (*GetLiveReplayListResp, error)`

### 4.2 Product (`product.go`)

```go
type ProductInfo struct {
    ProductID  string `json:"product_id,omitempty"`
    Title      string `json:"title"`
    SubTitle   string `json:"sub_title,omitempty"`
    HeadImgs   []string `json:"head_imgs,omitempty"`
    Status     int    `json:"status,omitempty"`
    CreateTime int64  `json:"create_time,omitempty"`
}

type AddProductReq struct {
    Product ProductInfo `json:"product"`
}
type AddProductResp struct {
    ProductID string `json:"product_id"`
}

type UpdateProductReq struct {
    Product ProductInfo `json:"product"`
}

type GetProductReq struct {
    ProductID string `json:"product_id"`
}
type GetProductResp struct {
    Product ProductInfo `json:"product"`
}

type ListProductReq struct {
    Status *int `json:"status,omitempty"`
    Offset int  `json:"offset,omitempty"`
    Limit  int  `json:"limit,omitempty"`
}
type ListProductResp struct {
    Products []ProductInfo `json:"products"`
    Total    int           `json:"total"`
}

type DeleteProductReq struct {
    ProductID string `json:"product_id"`
}
```

Methods:
- `AddProduct(ctx, *AddProductReq) (*AddProductResp, error)`
- `UpdateProduct(ctx, *UpdateProductReq) error`
- `GetProduct(ctx, *GetProductReq) (*GetProductResp, error)`
- `ListProduct(ctx, *ListProductReq) (*ListProductResp, error)`
- `DeleteProduct(ctx, *DeleteProductReq) error`

### 4.3 Order (`order.go`)

```go
type OrderInfo struct {
    OrderID    string `json:"order_id"`
    ProductID  string `json:"product_id"`
    Status     int    `json:"status"`
    CreateTime int64  `json:"create_time"`
    UpdateTime int64  `json:"update_time"`
}

type GetOrderReq struct {
    OrderID string `json:"order_id"`
}
type GetOrderResp struct {
    Order OrderInfo `json:"order"`
}

type ListOrderReq struct {
    Status    *int  `json:"status,omitempty"`
    StartTime int64 `json:"start_time,omitempty"`
    EndTime   int64 `json:"end_time,omitempty"`
    Offset    int   `json:"offset,omitempty"`
    Limit     int   `json:"limit,omitempty"`
}
type ListOrderResp struct {
    Orders []OrderInfo `json:"orders"`
    Total  int         `json:"total"`
}
```

Methods:
- `GetOrder(ctx, *GetOrderReq) (*GetOrderResp, error)`
- `ListOrder(ctx, *ListOrderReq) (*ListOrderResp, error)`

### 4.4 Data (`data.go`)

```go
type GetFinderLiveDataListReq struct {
    StartDate string `json:"start_date"`
    EndDate   string `json:"end_date"`
    Offset    int    `json:"offset,omitempty"`
    Limit     int    `json:"limit,omitempty"`
}
type FinderLiveData struct {
    Date       string `json:"date"`
    ViewCount  int64  `json:"view_count"`
    LikeCount  int64  `json:"like_count"`
    ShareCount int64  `json:"share_count"`
}
type GetFinderLiveDataListResp struct {
    Items []FinderLiveData `json:"items"`
    Total int              `json:"total"`
}

type GetFinderListReq struct {
    Offset int `json:"offset,omitempty"`
    Limit  int `json:"limit,omitempty"`
}
type FinderInfo struct {
    FinderID string `json:"finder_id"`
    Nickname string `json:"nickname"`
}
type GetFinderListResp struct {
    Items []FinderInfo `json:"items"`
    Total int          `json:"total"`
}
```

Methods:
- `GetFinderLiveDataList(ctx, *GetFinderLiveDataListReq) (*GetFinderLiveDataListResp, error)`
- `GetFinderList(ctx, *GetFinderListReq) (*GetFinderListResp, error)`

---

## 5. Testing

Same httptest pattern as mini-program:
- httptest.NewServer routing `/cgi-bin/token` for access_token + actual API path
- Validate request method (POST), access_token query param, request body JSON
- Return canned response, assert output struct fields

Helper function per test file:
```go
func newTestClient(t *testing.T, baseURL string) *Client
```

---

## 6. Error Prefix

All errors use `channels:` prefix (e.g., `fmt.Errorf("channels: %s errcode=%d errmsg=%s", path, resp.ErrCode, resp.ErrMsg)`).
