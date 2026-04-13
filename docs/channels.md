# channels — 微信视频号

> 微信视频号服务端 SDK，覆盖数据分析、直播间管理、电商订单及商品四大能力。

## 适用场景

- **数据分析**：拉取视频号直播数据汇总（观看、点赞、分享量），查询关联的视频号账号列表
- **直播间管理**：创建 / 删除直播间，查询直播间状态，获取回放录像列表
- **订单管理**：查询单笔订单详情，按条件分页列举电商订单
- **商品管理**：新增 / 更新 / 查询 / 列举 / 删除视频号小店商品

---

## 初始化 / Initialization

```go
func NewClient(cfg Config, opts ...Option) (*Client, error)
```

`Config` 字段说明：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `AppId` | `string` | 是 | 视频号关联的微信应用 AppID |
| `AppSecret` | `string` | 条件必填 | 应用密钥；未注入 `TokenSource` 时必须提供 |

### Options

| Option | 说明 |
|--------|------|
| `WithHTTP(h *utils.HTTP)` | 注入自定义 HTTP 客户端（测试用） |
| `WithTokenSource(ts TokenSource)` | 注入外部 token 来源（开放平台代调用），注入后 `AppSecret` 可留空 |

**示例：**

```go
import "github.com/godrealms/go-wechat-sdk/channels"

// 直接使用 AppSecret 获取 access_token
c, err := channels.NewClient(channels.Config{
    AppId:     "wx_your_appid",
    AppSecret: "your_app_secret",
})
if err != nil {
    log.Fatal(err)
}

// 通过开放平台 TokenSource 代调用
c, err = channels.NewClient(
    channels.Config{AppId: "wx_your_appid"},
    channels.WithTokenSource(myTokenSource),
)
```

---

## 错误处理 / Error Handling

所有方法在微信 API 返回非零 `errcode` 时，返回 `*channels.APIError`，可用 `errors.As` 提取错误码：

```go
import "errors"

resp, err := c.GetOrder(ctx, &channels.GetOrderReq{OrderID: "xxx"})
if err != nil {
    var apiErr *channels.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("errcode=%d errmsg=%s path=%s\n",
            apiErr.ErrCode, apiErr.ErrMsg, apiErr.Path)
    }
    log.Fatal(err)
}
```

`APIError` 实现了 `Code() int` 和 `Message() string`，方便统一处理微信 API 错误。

---

## API Reference

### 数据分析 / Data Analytics

#### GetFinderLiveDataList

```go
func (c *Client) GetFinderLiveDataList(ctx context.Context, req *GetFinderLiveDataListReq) (*GetFinderLiveDataListResp, error)
```

获取指定日期范围内视频号直播数据列表（观看量、点赞量、分享量）。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.StartDate` | `string` | 开始日期，格式 `YYYY-MM-DD` |
| `req.EndDate` | `string` | 结束日期，格式 `YYYY-MM-DD` |
| `req.Offset` | `*int` | 分页偏移量（可选） |
| `req.Limit` | `*int` | 每页返回数量（可选） |

返回 `*GetFinderLiveDataListResp`，包含 `Items []FinderLiveData`（每日数据条目）和 `Total int`（记录总数）。

每条 `FinderLiveData` 包含：

| 字段 | 类型 | 说明 |
|------|------|------|
| `Date` | `string` | 日期 |
| `ViewCount` | `int64` | 观看量 |
| `LikeCount` | `int64` | 点赞量 |
| `ShareCount` | `int64` | 分享量 |

---

#### GetFinderList

```go
func (c *Client) GetFinderList(ctx context.Context, req *GetFinderListReq) (*GetFinderListResp, error)
```

获取与当前应用关联的视频号账号（Finder）列表。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.Offset` | `*int` | 分页偏移量（可选） |
| `req.Limit` | `*int` | 每页返回数量（可选） |

返回 `*GetFinderListResp`，包含 `Items []FinderInfo`（账号信息列表）和 `Total int`（账号总数）。

每条 `FinderInfo` 包含 `FinderID`（视频号唯一 ID）和 `Nickname`（显示名称）。

---

### 直播 / Live Streaming

#### CreateRoom

```go
func (c *Client) CreateRoom(ctx context.Context, req *CreateRoomReq) (*CreateRoomResp, error)
```

创建一个新的视频号直播间。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.Name` | `string` | 直播间名称（必填） |
| `req.CoverImg` | `string` | 封面图 URL（可选） |
| `req.StartTime` | `int64` | 预计开播时间（Unix 时间戳，秒） |
| `req.EndTime` | `int64` | 预计结束时间（Unix 时间戳，秒） |

返回 `*CreateRoomResp`，包含新建直播间的 `RoomID string`。

---

#### DeleteRoom

```go
func (c *Client) DeleteRoom(ctx context.Context, req *DeleteRoomReq) error
```

删除指定的视频号直播间。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.RoomID` | `string` | 待删除的直播间 ID |

删除成功返回 `nil`，否则返回 `*APIError`。

---

#### GetLiveInfo

```go
func (c *Client) GetLiveInfo(ctx context.Context, req *GetLiveInfoReq) (*GetLiveInfoResp, error)
```

查询指定直播间的当前状态及元数据。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.RoomID` | `string` | 直播间 ID |

返回 `*GetLiveInfoResp`，包含 `LiveInfo` 结构体：

| 字段 | 类型 | 说明 |
|------|------|------|
| `RoomID` | `string` | 直播间 ID |
| `Name` | `string` | 直播间名称 |
| `Status` | `int` | 直播状态（0 未开始，1 直播中，2 已结束等） |
| `StartTime` | `int64` | 开播时间（Unix 秒） |
| `EndTime` | `int64` | 结束时间（Unix 秒） |

---

#### GetLiveReplayList

```go
func (c *Client) GetLiveReplayList(ctx context.Context, req *GetLiveReplayListReq) (*GetLiveReplayListResp, error)
```

获取指定直播间的回放录像列表。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.RoomID` | `string` | 直播间 ID |
| `req.Offset` | `*int` | 分页偏移量（可选） |
| `req.Limit` | `*int` | 每页返回数量（可选） |

返回 `*GetLiveReplayListResp`，包含 `LiveReplayList []LiveReplay` 和 `Total int`。

每条 `LiveReplay` 包含：

| 字段 | 类型 | 说明 |
|------|------|------|
| `MediaURL` | `string` | 回放视频地址 |
| `ExpireTime` | `int64` | 链接有效期（Unix 秒） |
| `CreateTime` | `int64` | 录像创建时间（Unix 秒） |

---

### 订单 / Orders

#### GetOrder

```go
func (c *Client) GetOrder(ctx context.Context, req *GetOrderReq) (*GetOrderResp, error)
```

查询单笔视频号电商订单详情。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.OrderID` | `string` | 订单 ID |

返回 `*GetOrderResp`，包含 `Order OrderInfo` 结构体：

| 字段 | 类型 | 说明 |
|------|------|------|
| `OrderID` | `string` | 订单 ID |
| `ProductID` | `string` | 商品 ID |
| `Status` | `int` | 订单状态 |
| `CreateTime` | `int64` | 创建时间（Unix 秒） |
| `UpdateTime` | `int64` | 最后更新时间（Unix 秒） |

---

#### ListOrder

```go
func (c *Client) ListOrder(ctx context.Context, req *ListOrderReq) (*ListOrderResp, error)
```

按条件分页列举视频号电商订单。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.Status` | `*int` | 按订单状态过滤（可选） |
| `req.StartTime` | `int64` | 创建时间下界（Unix 秒，可选） |
| `req.EndTime` | `int64` | 创建时间上界（Unix 秒，可选） |
| `req.Offset` | `*int` | 分页偏移量（可选） |
| `req.Limit` | `*int` | 每页返回数量（可选） |

返回 `*ListOrderResp`，包含 `Orders []OrderInfo` 和 `Total int`。

---

### 商品 / Products

#### AddProduct

```go
func (c *Client) AddProduct(ctx context.Context, req *AddProductReq) (*AddProductResp, error)
```

在视频号小店新增一个商品。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.Product.Title` | `string` | 商品标题（必填） |
| `req.Product.SubTitle` | `string` | 商品副标题（可选） |
| `req.Product.HeadImgs` | `[]string` | 主图 URL 列表（可选） |

返回 `*AddProductResp`，包含平台分配的 `ProductID string`。

---

#### UpdateProduct

```go
func (c *Client) UpdateProduct(ctx context.Context, req *UpdateProductReq) error
```

更新已有视频号小店商品的信息。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.Product.ProductID` | `string` | 待更新的商品 ID（必填） |
| `req.Product.Title` | `string` | 商品标题（可选） |
| `req.Product.SubTitle` | `string` | 商品副标题（可选） |
| `req.Product.HeadImgs` | `[]string` | 主图 URL 列表（可选） |

更新成功返回 `nil`，否则返回 `*APIError`。

---

#### GetProduct

```go
func (c *Client) GetProduct(ctx context.Context, req *GetProductReq) (*GetProductResp, error)
```

查询单个视频号小店商品详情。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.ProductID` | `string` | 商品 ID |

返回 `*GetProductResp`，包含 `Product ProductInfo` 结构体：

| 字段 | 类型 | 说明 |
|------|------|------|
| `ProductID` | `string` | 商品 ID |
| `Title` | `string` | 商品标题 |
| `SubTitle` | `string` | 商品副标题 |
| `HeadImgs` | `[]string` | 主图 URL 列表 |
| `Status` | `*int` | 商品状态 |
| `CreateTime` | `int64` | 创建时间（Unix 秒） |

---

#### ListProduct

```go
func (c *Client) ListProduct(ctx context.Context, req *ListProductReq) (*ListProductResp, error)
```

分页列举视频号小店商品。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.Status` | `*int` | 按商品状态过滤（可选） |
| `req.Offset` | `*int` | 分页偏移量（可选） |
| `req.Limit` | `*int` | 每页返回数量（可选） |

返回 `*ListProductResp`，包含 `Products []ProductInfo` 和 `Total int`。

---

#### DeleteProduct

```go
func (c *Client) DeleteProduct(ctx context.Context, req *DeleteProductReq) error
```

删除指定视频号小店商品。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.ProductID` | `string` | 待删除的商品 ID |

删除成功返回 `nil`，否则返回 `*APIError`。

---

## 完整示例 / Complete Example

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/godrealms/go-wechat-sdk/channels"
)

func main() {
	ctx := context.Background()

	// 初始化客户端
	c, err := channels.NewClient(channels.Config{
		AppId:     "wx_your_appid",
		AppSecret: "your_app_secret",
	})
	if err != nil {
		log.Fatal(err)
	}

	// ── 数据分析：拉取直播数据 ──────────────────────────────────────────
	dataResp, err := c.GetFinderLiveDataList(ctx, &channels.GetFinderLiveDataListReq{
		StartDate: "2026-04-01",
		EndDate:   "2026-04-13",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("直播数据条数: %d\n", dataResp.Total)
	for _, d := range dataResp.Items {
		fmt.Printf("  %s 观看=%d 点赞=%d 分享=%d\n",
			d.Date, d.ViewCount, d.LikeCount, d.ShareCount)
	}

	// ── 直播：创建直播间 ────────────────────────────────────────────────
	roomResp, err := c.CreateRoom(ctx, &channels.CreateRoomReq{
		Name:      "春季新品发布直播",
		StartTime: 1744905600, // 2026-04-17 12:00:00 UTC+8
		EndTime:   1744916400, // 2026-04-17 15:00:00 UTC+8
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("直播间 ID: %s\n", roomResp.RoomID)

	// ── 直播：查询直播间状态 ────────────────────────────────────────────
	liveResp, err := c.GetLiveInfo(ctx, &channels.GetLiveInfoReq{
		RoomID: roomResp.RoomID,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("直播间「%s」状态: %d\n", liveResp.LiveInfo.Name, liveResp.LiveInfo.Status)

	// ── 商品：新增商品 ──────────────────────────────────────────────────
	addResp, err := c.AddProduct(ctx, &channels.AddProductReq{
		Product: channels.ProductInfo{
			Title:    "高品质蓝牙耳机",
			SubTitle: "主动降噪 30h续航",
			HeadImgs: []string{"https://example.com/img/headphone.jpg"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("新商品 ID: %s\n", addResp.ProductID)

	// ── 商品：查询商品列表 ──────────────────────────────────────────────
	offset, limit := 0, 20
	listResp, err := c.ListProduct(ctx, &channels.ListProductReq{
		Offset: &offset,
		Limit:  &limit,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("商品总数: %d\n", listResp.Total)
	for _, p := range listResp.Products {
		fmt.Printf("  [%s] %s\n", p.ProductID, p.Title)
	}

	// ── 订单：列举最近订单 ──────────────────────────────────────────────
	orderOffset, orderLimit := 0, 10
	ordersResp, err := c.ListOrder(ctx, &channels.ListOrderReq{
		StartTime: 1743436800, // 2026-04-01 00:00:00 UTC+8
		EndTime:   1744128000, // 2026-04-09 00:00:00 UTC+8
		Offset:    &orderOffset,
		Limit:     &orderLimit,
	})
	if err != nil {
		// 提取 APIError 详情
		var apiErr *channels.APIError
		if errors.As(err, &apiErr) {
			fmt.Printf("API 错误 errcode=%d: %s\n", apiErr.ErrCode, apiErr.ErrMsg)
		}
		log.Fatal(err)
	}
	fmt.Printf("订单总数: %d\n", ordersResp.Total)
	for _, o := range ordersResp.Orders {
		fmt.Printf("  订单=%s 商品=%s 状态=%d\n", o.OrderID, o.ProductID, o.Status)
	}

	// ── 清理：删除直播间 ────────────────────────────────────────────────
	if err := c.DeleteRoom(ctx, &channels.DeleteRoomReq{RoomID: roomResp.RoomID}); err != nil {
		log.Fatal(err)
	}
	fmt.Println("直播间已删除")
}
```
