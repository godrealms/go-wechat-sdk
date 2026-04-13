# xiaowei — 微信小微 IoT 平台

`github.com/godrealms/go-wechat-sdk/xiaowei` — 包名：`xiaowei`

> 为无需营业执照的个人卖家提供微信小微商户（IoT 设备经营）全生命周期管理：店铺信息维护、实名认证（KYC）、商品上下架与订单发货/退款。

## 适用场景

- **个人卖家小程序**：通过微信小微平台开店，免营业执照，快速上线商品销售。
- **IoT 设备商**：设备内嵌微信支付能力，通过小微账号管理设备侧商品与订单。
- **运营后台**：批量查询商品列表、订单列表，驱动自动化运营工作流。
- **合规审核流程**：对接 KYC 提交与状态轮询，完成开户实名认证闭环。

## 初始化 / Initialization

```go
func NewClient(cfg Config, opts ...Option) (*Client, error)
```

`NewClient` 验证 `AppId` 非空，并在未注入 `TokenSource` 时同时验证 `AppSecret` 非空。底层 HTTP 客户端默认 30 秒超时，baseURL 为 `https://api.weixin.qq.com`。

### Config 字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `AppId` | `string` | 是 | 小微商户 AppID |
| `AppSecret` | `string` | 注入 `TokenSource` 时可省略 | 小微商户 AppSecret，用于自动换取 access_token |

### Options

| Option | 说明 |
|--------|------|
| `WithHTTP(h *utils.HTTP)` | 注入自定义 HTTP 客户端（测试或自定义超时时使用） |
| `WithTokenSource(ts TokenSource)` | 注入外部 token 来源（开放平台代调用场景） |

### TokenSource 接口

```go
type TokenSource interface {
    AccessToken(ctx context.Context) (string, error)
}
```

实现此接口可完全接管 token 获取逻辑，例如从 Redis 中央缓存读取，或通过开放平台代调用链路取得。

### Token 自动刷新

`Client` 内置进程级 token 缓存（读写锁，并发安全）。刷新规则：

- 在 token 过期前 **60 秒**提前刷新。
- 若微信返回的 `expires_in ≤ 60`（含异常值 0），自动修正为 **120 秒**，避免立即过期造成竞争。

## 错误处理 / Error Handling

所有方法在以下情况下返回非 `nil` 的 `error`：

1. **网络层错误**：TCP 超时、DNS 解析失败等，`error` 包含原始网络错误（可通过 `errors.Unwrap` 或 `errors.As` 向下检查）。
2. **微信业务错误**：API 返回 `errcode != 0`，错误信息格式为：
   ```
   xiaowei: /wxaapi/wxamicrostore/xxx errcode=<code> errmsg=<msg>
   ```
3. **Token 获取失败**：`xiaowei: fetch token: ...` 或 `xiaowei: token errcode=<code> errmsg=<msg>`。

错误均为标准 `error` 类型，可用 `strings.Contains` 检查错误码，或通过 `fmt.Errorf("... %w", err)` 包装后向上传递。

```go
resp, err := client.GetStoreInfo(ctx)
if err != nil {
    // 区分业务错误与网络错误
    if strings.Contains(err.Error(), "errcode=") {
        log.Printf("微信业务错误: %v", err)
    } else {
        log.Printf("网络或系统错误: %v", err)
    }
    return
}
```

## API Reference

### 店铺 / Store

#### GetStoreInfo

```go
func (c *Client) GetStoreInfo(ctx context.Context) (*GetStoreInfoResp, error)
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文，支持超时/取消 |

返回商户的店铺信息，包含店铺名称、Logo URL 及店铺状态（`1`=正常，`2`=暂停）。

**返回结构**

```go
type GetStoreInfoResp struct {
    StoreInfo *StoreInfo `json:"store_info"`
}

type StoreInfo struct {
    StoreName   string `json:"store_name,omitempty"`
    StoreHead   string `json:"store_head_img,omitempty"` // logo URL
    StoreStatus int    `json:"store_status,omitempty"`  // 1=active, 2=suspended
}
```

---

#### UpdateStoreInfo

```go
func (c *Client) UpdateStoreInfo(ctx context.Context, req *UpdateStoreInfoReq) error
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.StoreName` | `string` | 新店铺名称（可选） |
| `req.StoreHead` | `string` | 新 Logo 图片 URL（可选） |

更新店铺名称和/或 Logo，成功返回 `nil`，失败返回包含错误码的 `error`。

---

#### GetKYCStatus

```go
func (c *Client) GetKYCStatus(ctx context.Context) (*GetKYCStatusResp, error)
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |

查询商户实名认证（KYC）状态。

**返回结构**

```go
type GetKYCStatusResp struct {
    Status int    `json:"kyc_status"` // 0=未提交, 1=审核中, 2=已通过, 3=已拒绝
    Reason string `json:"reject_reason,omitempty"`
}
```

| `Status` 值 | 含义 |
|------------|------|
| `0` | 未提交 |
| `1` | 审核中 |
| `2` | 已通过 |
| `3` | 已拒绝，拒绝原因见 `Reason` |

---

#### SubmitKYC

```go
func (c *Client) SubmitKYC(ctx context.Context, req *SubmitKYCReq) error
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.RealName` | `string` | 商户真实姓名（法定姓名） |
| `req.IDCardNo` | `string` | 身份证号码 |
| `req.IDCardFront` | `string` | 身份证正面照片的 media_id |
| `req.IDCardBack` | `string` | 身份证背面照片的 media_id |

提交实名认证材料，成功返回 `nil`。`media_id` 需事先通过微信素材接口上传获得。

---

### 商品 / Products

#### AddMicroProduct

```go
func (c *Client) AddMicroProduct(ctx context.Context, product *MicroProduct) (*AddMicroProductResp, error)
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `product.Title` | `string` | 商品标题（必填） |
| `product.Price` | `int64` | 商品价格，单位**分** |
| `product.Stock` | `int` | 库存数量 |
| `product.ImgURLs` | `[]string` | 商品图片 URL 列表 |

上架新商品，返回微信分配的 `product_id`。

**返回结构**

```go
type AddMicroProductResp struct {
    ProductID string `json:"product_id"`
}
```

---

#### DelMicroProduct

```go
func (c *Client) DelMicroProduct(ctx context.Context, req *DelMicroProductReq) error
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.ProductID` | `string` | 要删除的商品 ID |

从小微店铺下架并删除指定商品，成功返回 `nil`。

---

#### GetMicroProduct

```go
func (c *Client) GetMicroProduct(ctx context.Context, req *GetMicroProductReq) (*GetMicroProductResp, error)
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.ProductID` | `string` | 商品 ID |

查询单个商品详情，返回含完整字段的 `*MicroProduct`。

**返回结构**

```go
type GetMicroProductResp struct {
    Product *MicroProduct `json:"product"`
}

type MicroProduct struct {
    ProductID string   `json:"product_id,omitempty"`
    Title     string   `json:"title"`
    Price     int64    `json:"price"`            // 单位：分
    Stock     int      `json:"stock,omitempty"`
    ImgURLs   []string `json:"img_urls,omitempty"`
}
```

---

#### ListMicroProducts

```go
func (c *Client) ListMicroProducts(ctx context.Context, req *ListMicroProductsReq) (*ListMicroProductsResp, error)
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.Page` | `int` | 页码，从 `1` 开始（可选，默认 `1`） |
| `req.PageSize` | `int` | 每页数量（可选，默认由微信决定） |

分页获取店铺商品列表，返回商品切片及总数。

**返回结构**

```go
type ListMicroProductsResp struct {
    Products []*MicroProduct `json:"product_list"`
    Total    int             `json:"total"`
}
```

---

### 订单 / Orders

#### GetMicroOrder

```go
func (c *Client) GetMicroOrder(ctx context.Context, req *GetMicroOrderReq) (*GetMicroOrderResp, error)
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.OrderID` | `string` | 订单 ID |

查询单笔订单详情，返回含状态与金额的 `*MicroOrder`。

**返回结构**

```go
type GetMicroOrderResp struct {
    Order *MicroOrder `json:"order"`
}

type MicroOrder struct {
    OrderID string `json:"order_id,omitempty"`
    Status  int    `json:"status,omitempty"`
    Amount  int64  `json:"amount,omitempty"` // 单位：分
}
```

---

#### ListMicroOrders

```go
func (c *Client) ListMicroOrders(ctx context.Context, req *ListMicroOrdersReq) (*ListMicroOrdersResp, error)
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.Status` | `int` | 按订单状态筛选（`0` 表示不筛选） |
| `req.Page` | `int` | 页码（可选） |
| `req.PageSize` | `int` | 每页数量（可选） |
| `req.StartTime` | `int64` | 下单时间范围起始，Unix 时间戳（可选） |
| `req.EndTime` | `int64` | 下单时间范围截止，Unix 时间戳（可选） |

分页查询订单列表，支持按状态和时间范围过滤，返回订单切片及总数。

**返回结构**

```go
type ListMicroOrdersResp struct {
    Orders   []*MicroOrder `json:"order_list"`
    TotalNum int           `json:"total_num"`
}
```

---

#### ShipMicroOrder

```go
func (c *Client) ShipMicroOrder(ctx context.Context, req *ShipMicroOrderReq) error
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.OrderID` | `string` | 订单 ID |
| `req.DeliveryCompany` | `string` | 快递公司名称或编码 |
| `req.TrackingNumber` | `string` | 快递单号 |

标记订单为已发货并写入物流信息，成功返回 `nil`。

---

#### RefundMicroOrder

```go
func (c *Client) RefundMicroOrder(ctx context.Context, req *RefundMicroOrderReq) error
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.OrderID` | `string` | 订单 ID |
| `req.RefundAmount` | `int64` | 退款金额，单位**分**；传 `0` 表示全额退款 |
| `req.RefundReason` | `string` | 退款原因（可选） |

发起订单退款，成功返回 `nil`。全额退款时将 `RefundAmount` 设为 `0`。

## 完整示例 / Complete Example

以下示例展示了一个典型的小微商户工作流：初始化客户端 → 提交 KYC → 上架商品 → 查询订单 → 发货。

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/godrealms/go-wechat-sdk/xiaowei"
)

func main() {
    ctx := context.Background()

    // 1. 初始化客户端
    client, err := xiaowei.NewClient(xiaowei.Config{
        AppId:     "wx_your_appid",
        AppSecret: "your_app_secret",
    })
    if err != nil {
        log.Fatalf("init client: %v", err)
    }

    // 2. 查询店铺信息
    storeResp, err := client.GetStoreInfo(ctx)
    if err != nil {
        log.Fatalf("get store info: %v", err)
    }
    fmt.Printf("店铺名称: %s  状态: %d\n",
        storeResp.StoreInfo.StoreName, storeResp.StoreInfo.StoreStatus)

    // 3. 查询 KYC 状态，未通过则提交
    kycResp, err := client.GetKYCStatus(ctx)
    if err != nil {
        log.Fatalf("get kyc status: %v", err)
    }
    if kycResp.Status == 0 { // 未提交
        err = client.SubmitKYC(ctx, &xiaowei.SubmitKYCReq{
            RealName:    "张三",
            IDCardNo:    "110101199001011234",
            IDCardFront: "media_id_front_xxxx",
            IDCardBack:  "media_id_back_xxxx",
        })
        if err != nil {
            log.Fatalf("submit kyc: %v", err)
        }
        fmt.Println("KYC 已提交，等待审核")
    } else {
        fmt.Printf("KYC 状态: %d\n", kycResp.Status)
    }

    // 4. 上架一件商品
    addResp, err := client.AddMicroProduct(ctx, &xiaowei.MicroProduct{
        Title:   "手工编织帽子",
        Price:   3900, // 39.00 元
        Stock:   50,
        ImgURLs: []string{"https://example.com/hat.jpg"},
    })
    if err != nil {
        log.Fatalf("add product: %v", err)
    }
    fmt.Printf("商品已上架，product_id: %s\n", addResp.ProductID)

    // 5. 分页查询最近 7 天的订单
    now := time.Now()
    ordersResp, err := client.ListMicroOrders(ctx, &xiaowei.ListMicroOrdersReq{
        Page:      1,
        PageSize:  20,
        StartTime: now.Add(-7 * 24 * time.Hour).Unix(),
        EndTime:   now.Unix(),
    })
    if err != nil {
        log.Fatalf("list orders: %v", err)
    }
    fmt.Printf("近 7 天共 %d 笔订单\n", ordersResp.TotalNum)

    // 6. 对待发货订单批量发货
    for _, order := range ordersResp.Orders {
        if order.Status != 1 { // 假设 1 = 待发货
            continue
        }
        err = client.ShipMicroOrder(ctx, &xiaowei.ShipMicroOrderReq{
            OrderID:         order.OrderID,
            DeliveryCompany: "SF",
            TrackingNumber:  "SF1234567890",
        })
        if err != nil {
            log.Printf("ship order %s failed: %v", order.OrderID, err)
            continue
        }
        fmt.Printf("订单 %s 已标记发货\n", order.OrderID)
    }
}
```

### 注入外部 TokenSource（多进程/开放平台场景）

当多个进程共享同一个 access_token 时，建议通过 `WithTokenSource` 注入集中式 token 管理器，避免各进程并发刷新导致接口限频：

```go
type RedisTokenSource struct {
    rdb    *redis.Client
    appID  string
    secret string
}

func (r *RedisTokenSource) AccessToken(ctx context.Context) (string, error) {
    tok, err := r.rdb.Get(ctx, "wx:token:"+r.appID).Result()
    if err == nil && tok != "" {
        return tok, nil
    }
    // 从微信刷新并写回 Redis（此处省略实现）
    return refreshAndCache(ctx, r.appID, r.secret, r.rdb)
}

// 使用方式
client, err := xiaowei.NewClient(
    xiaowei.Config{AppId: "wx_your_appid"},
    xiaowei.WithTokenSource(&RedisTokenSource{rdb: rdb, appID: "wx_your_appid", secret: "secret"}),
)
```
