# mini-game（小游戏）

> 小游戏服务端 SDK，覆盖登录认证、广告数据、数据分析、帧同步房间、虚拟支付与内容安全检测六大领域。

## 适用场景

| 场景 | 说明 |
|------|------|
| 登录认证 | 服务端校验 `wx.login` 返回的临时 code，换取 `openid` / `session_key` |
| 广告变现 | 拉取游戏广告位的展示、点击及收入数据 |
| 数据分析 | 查询每日访问汇总趋势与用户留存数据 |
| 帧同步 | 创建实时对战房间，查询房间状态与成员列表 |
| 虚拟支付 | 创建并查询游戏虚拟货币购买订单 |
| 内容安全 | 对用户生成的文字内容执行违规检测 |
| 云存储 | 读写用户维度的游戏存档键值对 |

## 初始化 / Initialization

```go
func NewClient(cfg Config, opts ...Option) (*Client, error)
```

`NewClient` 按照提供的 `Config` 和可选项构造客户端。`AppId` 必填；若未注入 `TokenSource` 则 `AppSecret` 也必填。

### Config 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `AppId` | `string` | 小游戏 AppID，必填 |
| `AppSecret` | `string` | 小游戏 AppSecret；注入 `TokenSource` 时可省略 |

### Options

| Option | 说明 |
|--------|------|
| `WithHTTP(h *utils.HTTP)` | 注入自定义 HTTP 客户端，主要用于单元测试 |
| `WithTokenSource(ts TokenSource)` | 注入外部 token 来源（如开放平台代调用），设置后 `AccessToken()` 不再直接请求 `/cgi-bin/token` |

```go
client, err := mini_game.NewClient(mini_game.Config{
    AppId:     "wx_your_appid",
    AppSecret: "your_appsecret",
})
if err != nil {
    log.Fatal(err)
}
```

## 错误处理 / Error Handling

当微信 API 返回业务错误（`errcode != 0`）时，SDK 返回 `*APIError`：

```go
type APIError struct {
    ErrCode int
    ErrMsg  string
    Path    string // 触发错误的 API 路径
}
```

使用 `errors.As` 判断并获取错误详情：

```go
resp, err := client.Code2Session(ctx, jsCode)
if err != nil {
    var apiErr *mini_game.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("WeChat API error: code=%d msg=%s path=%s\n",
            apiErr.ErrCode, apiErr.ErrMsg, apiErr.Path)
    } else {
        fmt.Printf("network/transport error: %v\n", err)
    }
    return
}
```

`Code2Session`、`GetGameAdData`、`GetDailySummary`、`GetDailyRetain`、`CreateGameRoom`、`GetRoomInfo`、`CreateOrder`、`QueryOrder`、`MsgSecCheck`、`SetUserStorage`、`GetUserStorage` 均在微信返回非零 `errcode` 时返回 `*APIError`。

## API Reference

### 认证 / Auth

#### Code2Session

```go
func (c *Client) Code2Session(ctx context.Context, jsCode string) (*Code2SessionResp, error)
```

将 `wx.login` 获取的临时登录凭证 `js_code` 换取用户的 `openid`、`session_key` 及（适用时的）`unionid`。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `jsCode` | `string` | `wx.login` 返回的临时 code，必填 |

**返回：** `*Code2SessionResp`，包含以下字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `OpenId` | `string` | 用户唯一标识 |
| `SessionKey` | `string` | 会话密钥 |
| `UnionId` | `string` | 用户在开放平台的唯一标识（仅当小游戏已绑定开放平台账号时返回） |

微信返回业务错误时返回 `*APIError`。

---

#### AccessToken

```go
func (c *Client) AccessToken(ctx context.Context) (string, error)
```

获取有效的全局 `access_token`。当有效期剩余不足 60 秒时自动刷新。若注入了 `TokenSource`，则委托给外部实现。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |

**返回：** 有效的 `access_token` 字符串。

---

### 广告 / Ad

#### GetGameAdData

```go
func (c *Client) GetGameAdData(ctx context.Context, req *GetGameAdDataReq) (*GetGameAdDataResp, error)
```

获取小游戏指定日期范围内的广告表现数据（请求量、展示量、点击量、收入）。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.StartDate` | `string` | 起始日期，格式 `YYYYMMDD` |
| `req.EndDate` | `string` | 截止日期，格式 `YYYYMMDD` |
| `req.AdUnitID` | `string` | 广告单元 ID，可选；为空时返回全部广告位数据 |

**返回：** `*GetGameAdDataResp`，包含 `Items []GameAdData`。每条 `GameAdData` 字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `Date` | `string` | 日期 |
| `AdUnitID` | `string` | 广告单元 ID |
| `ReqCount` | `int64` | 广告请求次数 |
| `ShowCount` | `int64` | 广告展示次数 |
| `ClickCount` | `int64` | 广告点击次数 |
| `Income` | `int64` | 广告收入（单位：分） |

---

### 数据分析 / Analysis

#### GetDailySummary

```go
func (c *Client) GetDailySummary(ctx context.Context, req *AnalysisDateReq) (*GetDailySummaryResp, error)
```

获取小游戏指定日期范围内每日访问汇总趋势数据。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.BeginDate` | `string` | 起始日期，格式 `YYYYMMDD` |
| `req.EndDate` | `string` | 截止日期，格式 `YYYYMMDD` |

**返回：** `*GetDailySummaryResp`，包含 `List []DailySummaryItem`。每条 `DailySummaryItem` 字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `RefDate` | `string` | 日期 |
| `VisitTotal` | `int64` | 累计访问次数 |
| `SharePV` | `int64` | 转发次数 |
| `ShareUV` | `int64` | 转发人数 |

---

#### GetDailyRetain

```go
func (c *Client) GetDailyRetain(ctx context.Context, req *AnalysisDateReq) (*GetDailyRetainResp, error)
```

获取小游戏指定日期范围内的用户留存数据。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.BeginDate` | `string` | 起始日期，格式 `YYYYMMDD` |
| `req.EndDate` | `string` | 截止日期，格式 `YYYYMMDD` |

**返回：** `*GetDailyRetainResp`，字段说明：

| 字段 | 类型 | 说明 |
|------|------|------|
| `RefDate` | `string` | 日期 |
| `VisitUVNew` | `[]DailyRetainItem` | 新增用户留存，每项含 `DateKey`（天数偏移）和 `Value`（人数） |
| `VisitUV` | `[]DailyRetainItem` | 活跃用户留存，结构同上 |

---

### 帧同步 / Frame Sync

#### CreateGameRoom

```go
func (c *Client) CreateGameRoom(ctx context.Context, req *CreateGameRoomReq) (*CreateGameRoomResp, error)
```

创建一个帧同步对战房间，返回房间 ID。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.MaxNum` | `int` | 房间最大人数 |
| `req.AccessInfo` | `string` | 自定义扩展信息，可选 |

**返回：** `*CreateGameRoomResp`，包含 `RoomID string`（新房间的唯一标识）。

---

#### GetRoomInfo

```go
func (c *Client) GetRoomInfo(ctx context.Context, req *GetRoomInfoReq) (*GetRoomInfoResp, error)
```

查询帧同步房间的当前状态及成员列表。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.RoomID` | `string` | 房间 ID |

**返回：** `*GetRoomInfoResp`，字段说明：

| 字段 | 类型 | 说明 |
|------|------|------|
| `RoomID` | `string` | 房间 ID |
| `Status` | `int` | 房间状态 |
| `Members` | `[]RoomMember` | 成员列表，每项含 `OpenID` 和 `Role` |

---

### 支付 / Payment

#### CreateOrder

```go
func (c *Client) CreateOrder(ctx context.Context, req *CreateOrderReq) (*CreateOrderResp, error)
```

创建小游戏虚拟货币购买订单。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.OpenID` | `string` | 用户 OpenID |
| `req.Env` | `int` | 环境标识（0：正式，1：沙箱） |
| `req.ZoneID` | `string` | 分区 ID |
| `req.ProductID` | `string` | 商品 ID |
| `req.Quantity` | `int` | 购买数量 |

**返回：** `*CreateOrderResp`，包含 `OrderID string`（新订单 ID）和 `Balance int64`（用户当前余额）。

---

#### QueryOrder

```go
func (c *Client) QueryOrder(ctx context.Context, req *QueryOrderReq) (*QueryOrderResp, error)
```

查询已有小游戏订单的状态与详情。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.OrderID` | `string` | 订单 ID |
| `req.OpenID` | `string` | 用户 OpenID |

**返回：** `*QueryOrderResp`，字段说明：

| 字段 | 类型 | 说明 |
|------|------|------|
| `OrderID` | `string` | 订单 ID |
| `Status` | `int` | 订单状态 |
| `PayAmount` | `int64` | 支付金额（单位：分） |
| `CreateTime` | `int64` | 订单创建时间（Unix 时间戳） |

---

### 安全 / Security

#### MsgSecCheck

```go
func (c *Client) MsgSecCheck(ctx context.Context, req *MsgSecCheckReq) (*MsgSecCheckResp, error)
```

将文本内容提交至微信安全检测 API，返回内容审核结果。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.Content` | `string` | 待检测的文本内容 |
| `req.Version` | `int` | 接口版本号（建议填 `2`） |
| `req.Scene` | `int` | 场景值（参考微信文档） |
| `req.OpenID` | `string` | 用户 OpenID |

**返回：** `*MsgSecCheckResp`，包含 `Result SecCheckResult`：

| 字段 | 类型 | 说明 |
|------|------|------|
| `Suggest` | `string` | 建议操作：`pass`（通过）、`review`（待人工审核）、`risky`（违规） |
| `Label` | `int` | 违规类型标签 |

---

### 存储 / Storage

#### SetUserStorage

```go
func (c *Client) SetUserStorage(ctx context.Context, req *SetUserStorageReq) error
```

将指定的键值对写入用户的小游戏云存档。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.OpenID` | `string` | 用户 OpenID |
| `req.KVList` | `[]KVData` | 待写入的键值对列表，每项含 `Key` 和 `Value` |
| `req.SigMethod` | `string` | 签名方法（如 `hmac_sha256`） |
| `req.Signature` | `string` | 请求签名 |

**返回：** 写入成功时返回 `nil`，失败时返回 `*APIError` 或网络错误。

---

#### GetUserStorage

```go
func (c *Client) GetUserStorage(ctx context.Context, req *GetUserStorageReq) (*GetUserStorageResp, error)
```

从用户的小游戏云存档中读取指定键的值。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req.OpenID` | `string` | 用户 OpenID |
| `req.KeyList` | `[]string` | 待读取的键名列表 |
| `req.SigMethod` | `string` | 签名方法（如 `hmac_sha256`） |
| `req.Signature` | `string` | 请求签名 |

**返回：** `*GetUserStorageResp`，包含 `KVList []KVData`（查询结果键值对列表）。

---

## 完整示例 / Complete Example

以下示例演示如何初始化客户端并完成 `wx.login` 登录流程：

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"

    mini_game "github.com/godrealms/go-wechat-sdk/mini-game"
)

func main() {
    // 1. 初始化客户端
    client, err := mini_game.NewClient(mini_game.Config{
        AppId:     "wx_your_appid",
        AppSecret: "your_appsecret",
    })
    if err != nil {
        log.Fatalf("NewClient: %v", err)
    }

    ctx := context.Background()

    // 2. 前端调用 wx.login() 后将 code 传递至服务端
    jsCode := "code_from_wx_login"

    // 3. 换取 openid / session_key
    resp, err := client.Code2Session(ctx, jsCode)
    if err != nil {
        var apiErr *mini_game.APIError
        if errors.As(err, &apiErr) {
            // 微信业务错误，例如 code 无效或已过期
            fmt.Printf("WeChat API error: code=%d msg=%s path=%s\n",
                apiErr.ErrCode, apiErr.ErrMsg, apiErr.Path)
        } else {
            // 网络或传输层错误
            fmt.Printf("transport error: %v\n", err)
        }
        return
    }

    fmt.Printf("openid:      %s\n", resp.OpenId)
    fmt.Printf("session_key: %s\n", resp.SessionKey)
    if resp.UnionId != "" {
        fmt.Printf("unionid:     %s\n", resp.UnionId)
    }

    // 4. 可选：获取全局 access_token（用于调用其他需要 token 的接口）
    token, err := client.AccessToken(ctx)
    if err != nil {
        log.Fatalf("AccessToken: %v", err)
    }
    fmt.Printf("access_token: %s\n", token)
}
```
