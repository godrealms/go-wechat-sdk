# mini-store — 微信小店

`github.com/godrealms/go-wechat-sdk/mini-store` — 包名：`mini_store`

微信小店（原微信小商店）Open API 的 Go 封装，覆盖商品管理、订单管理、优惠券、结算账期以及商家信息查询，所有方法都以 `context.Context` 作为第一个参数，`*Client` 并发安全。

推荐导入别名：

```go
import ministore "github.com/godrealms/go-wechat-sdk/mini-store"
```

---

## 适用场景

| 场景 | 说明 |
|------|------|
| **商品管理** | 创建/更新/删除商品，查询商品列表，提交或撤回平台审核，修改上下架状态 |
| **订单管理** | 查询订单详情与列表，修改价格，关闭订单，上传发货信息，处理售后退款 |
| **优惠券** | 创建折扣券/满减券，查询与分页列表，激活/停用优惠券 |
| **结算** | 查询指定时间段内的结算账期，获取商家注册信息、品牌目录、商品分类 |

---

## 初始化

### Config

```go
type Config struct {
    AppId     string
    AppSecret string
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `AppId` | `string` | 是 | 小程序/小店的 AppId |
| `AppSecret` | `string` | 条件必填 | 未注入 `TokenSource` 时必填；SDK 用它自动换取并缓存 access_token（到期前 60s 刷新） |

### Options

```go
func WithHTTP(h *utils.HTTP) Option
func WithTokenSource(ts TokenSource) Option
```

| Option | 说明 |
|--------|------|
| `WithHTTP` | 注入自定义 HTTP 客户端，主要用于测试 |
| `WithTokenSource` | 注入外部 token 提供方（实现 `TokenSource` 接口即可），注入后 `AppSecret` 可留空 |

### TokenSource 接口

```go
type TokenSource interface {
    AccessToken(ctx context.Context) (string, error)
}
```

### NewClient

```go
func NewClient(cfg Config, opts ...Option) (*Client, error)
```

`NewClient` 在以下情况返回错误：

- `AppId` 为空
- `AppSecret` 为空且未注入 `TokenSource`

**示例**

```go
client, err := ministore.NewClient(ministore.Config{
    AppId:     "wx1234567890",
    AppSecret: "your_app_secret",
})
if err != nil {
    log.Fatal(err)
}
```

---

## 错误处理

所有方法在 API 返回 `errcode != 0` 时都会返回包含路径、errcode 和 errmsg 的格式化错误，例如：

```
mini_store: /shop/spu/add errcode=40001 errmsg=invalid credential
```

HTTP 层错误为 `*utils.HTTPError`，可用 `errors.As` 检查：

```go
var httpErr *utils.HTTPError
if errors.As(err, &httpErr) {
    // httpErr.StatusCode — HTTP 状态码
    // httpErr.Body      — 响应原始 body
}
```

---

## API Reference

### 商品管理（Products）

#### AddProduct

```go
func (c *Client) AddProduct(ctx context.Context, product *Product) (*AddProductResp, error)
```

创建新商品，返回平台分配的 `product_id`。

**Product 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `Title` | `string` | 商品标题（必填） |
| `SubTitle` | `string` | 副标题 |
| `HeadImgs` | `[]string` | 商品主图 URL 列表 |
| `Description` | `string` | 商品详情描述（JSON 字段 `desc_info`） |
| `Status` | `int` | 商品状态 |

**AddProductResp**

| 字段 | 类型 | 说明 |
|------|------|------|
| `ProductID` | `string` | 平台分配的商品 ID |

---

#### UpdateProduct

```go
func (c *Client) UpdateProduct(ctx context.Context, req *UpdateProductReq) error
```

更新已有商品信息。

**UpdateProductReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `ProductID` | `string` | 要更新的商品 ID（必填） |
| `Product` | `*Product` | 更新后的商品信息（JSON 字段 `spu_info`） |

---

#### DelProduct

```go
func (c *Client) DelProduct(ctx context.Context, req *DelProductReq) error
```

删除指定商品。

**DelProductReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `ProductID` | `string` | 要删除的商品 ID（必填） |

---

#### GetProduct

```go
func (c *Client) GetProduct(ctx context.Context, req *GetProductReq) (*GetProductResp, error)
```

查询单个商品详情。

**GetProductReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `ProductID` | `string` | 商品 ID（必填） |

**GetProductResp**

| 字段 | 类型 | 说明 |
|------|------|------|
| `SPU` | `*Product` | 商品详情 |

---

#### ListProducts

```go
func (c *Client) ListProducts(ctx context.Context, req *ListProductsReq) (*ListProductsResp, error)
```

分页查询商品列表。

**ListProductsReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `Status` | `int` | 过滤状态（0 表示不过滤） |
| `PageSize` | `int` | 每页数量 |
| `Page` | `int` | 页码（从 1 开始） |

**ListProductsResp**

| 字段 | 类型 | 说明 |
|------|------|------|
| `SPUs` | `[]*Product` | 商品列表 |
| `TotalNum` | `int` | 总数量 |

---

#### UpdateProductStatus

```go
func (c *Client) UpdateProductStatus(ctx context.Context, req *UpdateProductStatusReq) error
```

修改商品上下架状态（无需重新审核）。

**UpdateProductStatusReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `ProductID` | `string` | 商品 ID（必填） |
| `Status` | `int` | `0` = 下架，`1` = 上架 |

---

#### SubmitProductAudit

```go
func (c *Client) SubmitProductAudit(ctx context.Context, req *SubmitProductAuditReq) error
```

将商品提交平台审核，审核通过后方可上架。

**SubmitProductAuditReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `ProductID` | `string` | 商品 ID（必填） |

---

#### CancelProductAudit

```go
func (c *Client) CancelProductAudit(ctx context.Context, req *CancelProductAuditReq) error
```

撤回尚在审核中的商品审核申请。

**CancelProductAuditReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `ProductID` | `string` | 商品 ID（必填） |

---

### 订单管理（Orders）

#### GetOrder

```go
func (c *Client) GetOrder(ctx context.Context, req *GetOrderReq) (*GetOrderResp, error)
```

按 `order_id` 查询单笔订单详情。

**GetOrderReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `OrderID` | `string` | 订单 ID（必填） |

**GetOrderResp**

| 字段 | 类型 | 说明 |
|------|------|------|
| `Order` | `*Order` | 订单详情 |

**Order 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `OrderID` | `string` | 订单 ID |
| `Status` | `int` | 订单状态 |
| `UserOpenID` | `string` | 买家 openid（JSON 字段 `openid`） |

---

#### ListOrders

```go
func (c *Client) ListOrders(ctx context.Context, req *ListOrdersReq) (*ListOrdersResp, error)
```

分页查询订单列表，可按状态和时间范围过滤。

**ListOrdersReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `Status` | `int` | 订单状态过滤（0 表示不过滤） |
| `Page` | `int` | 页码 |
| `PageSize` | `int` | 每页数量 |
| `StartTime` | `int64` | 起始时间（Unix 秒级时间戳） |
| `EndTime` | `int64` | 结束时间（Unix 秒级时间戳） |

**ListOrdersResp**

| 字段 | 类型 | 说明 |
|------|------|------|
| `Orders` | `[]*Order` | 订单列表（JSON 字段 `order_list`） |
| `TotalNum` | `int` | 总数量 |

---

#### UpdateOrderPrice

```go
func (c *Client) UpdateOrderPrice(ctx context.Context, req *UpdateOrderPriceReq) error
```

在买家付款前修改订单总价（仅限待付款状态）。

**UpdateOrderPriceReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `OrderID` | `string` | 订单 ID（必填） |
| `NewPrice` | `int64` | 新价格，单位：分（人民币） |

---

#### CloseOrder

```go
func (c *Client) CloseOrder(ctx context.Context, req *CloseOrderReq) error
```

关闭开放中的订单，关闭后买家无法支付。

**CloseOrderReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `OrderID` | `string` | 订单 ID（必填） |

---

#### UploadShipping

```go
func (c *Client) UploadShipping(ctx context.Context, req *UploadShippingReq) error
```

为已支付订单上传发货快递信息。

**UploadShippingReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `OrderID` | `string` | 订单 ID（必填） |
| `DeliveryCompany` | `string` | 快递公司编码 |
| `DeliverySN` | `string` | 快递单号 |

---

#### GetAfterSaleOrder

```go
func (c *Client) GetAfterSaleOrder(ctx context.Context, req *GetAfterSaleOrderReq) (*GetAfterSaleOrderResp, error)
```

查询售后（退款/退货）单详情。

**GetAfterSaleOrderReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `AfterSaleOrderID` | `string` | 售后单 ID（必填） |

**AfterSaleOrderDetail 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `AfterSaleOrderID` | `string` | 售后单 ID |
| `OrderID` | `string` | 原始订单 ID |
| `RefundAmount` | `int64` | 退款金额，单位：分 |
| `Status` | `int` | 售后单状态 |
| `Reason` | `string` | 售后原因 |

---

#### AcceptRefund

```go
func (c *Client) AcceptRefund(ctx context.Context, req *AcceptRefundReq) error
```

同意买家退款申请。

**AcceptRefundReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `AfterSaleOrderID` | `string` | 售后单 ID（必填） |

---

#### RejectRefund

```go
func (c *Client) RejectRefund(ctx context.Context, req *RejectRefundReq) error
```

拒绝买家退款申请，可附拒绝原因。

**RejectRefundReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `AfterSaleOrderID` | `string` | 售后单 ID（必填） |
| `RejectReason` | `string` | 拒绝原因（可选） |

---

### 优惠券（Coupons）

#### AddCoupon

```go
func (c *Client) AddCoupon(ctx context.Context, coupon *Coupon) (*AddCouponResp, error)
```

创建新优惠券，返回平台分配的 `coupon_id`。

**Coupon 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `Name` | `string` | 优惠券名称（JSON 字段 `coupon_name`） |
| `Type` | `int` | `1` = 固定金额减，`2` = 折扣百分比（JSON 字段 `coupon_type`） |
| `Discount` | `int64` | 优惠力度：金额减时为分，折扣时为基点（万分之一） |
| `MinAmount` | `int64` | 最低消费门槛，单位：分 |

**AddCouponResp**

| 字段 | 类型 | 说明 |
|------|------|------|
| `CouponID` | `string` | 平台分配的优惠券 ID |

---

#### GetCoupon

```go
func (c *Client) GetCoupon(ctx context.Context, req *GetCouponReq) (*GetCouponResp, error)
```

查询单张优惠券详情。

**GetCouponReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `CouponID` | `string` | 优惠券 ID（必填） |

**GetCouponResp**

| 字段 | 类型 | 说明 |
|------|------|------|
| `Coupon` | `*Coupon` | 优惠券详情 |

---

#### UpdateCouponStatus

```go
func (c *Client) UpdateCouponStatus(ctx context.Context, req *UpdateCouponStatusReq) error
```

激活或停用优惠券。

**UpdateCouponStatusReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `CouponID` | `string` | 优惠券 ID（必填） |
| `Status` | `int` | `0` = 停用，`1` = 激活 |

---

#### ListCoupons

```go
func (c *Client) ListCoupons(ctx context.Context, req *ListCouponsReq) (*ListCouponsResp, error)
```

分页查询优惠券列表。

**ListCouponsReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `Status` | `int` | 状态过滤（0 表示不过滤） |
| `Page` | `int` | 页码 |
| `PageSize` | `int` | 每页数量 |

**ListCouponsResp**

| 字段 | 类型 | 说明 |
|------|------|------|
| `Coupons` | `[]*Coupon` | 优惠券列表（JSON 字段 `coupon_list`） |
| `TotalNum` | `int` | 总数量 |

---

### 结算与商家信息（Settlement & Merchant）

#### GetMerchantInfo

```go
func (c *Client) GetMerchantInfo(ctx context.Context) (*GetMerchantInfoResp, error)
```

查询商家注册及结算信息，无需额外请求参数。

**MerchantInfo 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `Name` | `string` | 商家名称 |
| `MerchantID` | `string` | 商家 ID |
| `Status` | `int` | 商家状态 |
| `SettlementBankNo` | `string` | 结算银行账号 |

---

#### GetSettlement

```go
func (c *Client) GetSettlement(ctx context.Context, req *GetSettlementReq) (*GetSettlementResp, error)
```

查询指定时间范围内的结算账期记录。

**GetSettlementReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `StartTime` | `int64` | 起始时间（Unix 秒级时间戳，必填） |
| `EndTime` | `int64` | 结束时间（Unix 秒级时间戳，必填） |

**Settlement 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `ID` | `string` | 结算记录 ID |
| `SettleTime` | `int64` | 结算时间（Unix 时间戳） |
| `SettleAmount` | `int64` | 结算金额，单位：分 |
| `Status` | `int` | 结算状态 |

---

#### GetBrandList

```go
func (c *Client) GetBrandList(ctx context.Context) (*GetBrandListResp, error)
```

获取可用于商品发布的品牌列表，无需额外请求参数。

**Brand 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `ID` | `int` | 品牌 ID |
| `Name` | `string` | 品牌名称 |

---

#### GetCategoryList

```go
func (c *Client) GetCategoryList(ctx context.Context, req *GetCategoryListReq) (*GetCategoryListResp, error)
```

获取商品分类列表，可传父分类 ID 获取子分类。

**GetCategoryListReq 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `ParentCatID` | `int` | 父分类 ID（JSON 字段 `f_cat_id`）；传 `0` 获取一级分类 |

**Category 字段**

| 字段 | 类型 | 说明 |
|------|------|------|
| `CatID` | `int` | 分类 ID（JSON 字段 `f_cat_id`） |
| `Name` | `string` | 分类名称 |

---

## 并发语义

`*Client` 并发安全：

- access_token 缓存使用 `sync.RWMutex` 保护，过期前 60s 自动刷新。
- 每次请求独立携带 token，不共享状态，多 goroutine 并发调用安全。
- 底层 HTTP 客户端默认超时 30s，可通过 `client.HTTP()` 访问器替换。

---

## 完整示例

以下示例演示初始化客户端，并依次完成商品发布、订单发货和优惠券创建的完整流程。

```go
package main

import (
	"context"
	"log"
	"time"

	ministore "github.com/godrealms/go-wechat-sdk/mini-store"
)

func main() {
	ctx := context.Background()

	// 1. 初始化客户端
	client, err := ministore.NewClient(ministore.Config{
		AppId:     "wx1234567890",
		AppSecret: "your_app_secret",
	})
	if err != nil {
		log.Fatal(err)
	}

	// 2. 查询可用分类（选填商品时使用）
	cats, err := client.GetCategoryList(ctx, &ministore.GetCategoryListReq{ParentCatID: 0})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("root categories: %d", len(cats.CategoryList))

	// 3. 创建商品
	addResp, err := client.AddProduct(ctx, &ministore.Product{
		Title:       "测试商品",
		SubTitle:    "副标题文案",
		HeadImgs:    []string{"https://example.com/img1.jpg"},
		Description: "详细描述",
	})
	if err != nil {
		log.Fatal(err)
	}
	productID := addResp.ProductID
	log.Printf("created product: %s", productID)

	// 4. 提交审核后上架
	if err := client.SubmitProductAudit(ctx, &ministore.SubmitProductAuditReq{ProductID: productID}); err != nil {
		log.Fatal(err)
	}
	// 审核通过后修改状态为上架
	if err := client.UpdateProductStatus(ctx, &ministore.UpdateProductStatusReq{
		ProductID: productID,
		Status:    1,
	}); err != nil {
		log.Fatal(err)
	}

	// 5. 查询最新订单列表（近 24 小时）
	now := time.Now()
	orders, err := client.ListOrders(ctx, &ministore.ListOrdersReq{
		Page:      1,
		PageSize:  20,
		StartTime: now.Add(-24 * time.Hour).Unix(),
		EndTime:   now.Unix(),
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("recent orders: %d", orders.TotalNum)

	// 6. 对第一笔订单上传发货信息
	if len(orders.Orders) > 0 {
		order := orders.Orders[0]
		if err := client.UploadShipping(ctx, &ministore.UploadShippingReq{
			OrderID:         order.OrderID,
			DeliveryCompany: "SF",
			DeliverySN:      "SF1234567890",
		}); err != nil {
			log.Printf("upload shipping failed: %v", err)
		}
	}

	// 7. 创建满减优惠券（满 100 减 10，单位：分）
	couponResp, err := client.AddCoupon(ctx, &ministore.Coupon{
		Name:      "满100减10",
		Type:      1,
		Discount:  1000,
		MinAmount: 10000,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("created coupon: %s", couponResp.CouponID)

	// 8. 查询结算记录
	settle, err := client.GetSettlement(ctx, &ministore.GetSettlementReq{
		StartTime: now.AddDate(0, -1, 0).Unix(),
		EndTime:   now.Unix(),
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("settlement records: %d", len(settle.SettlementList))
}
```
