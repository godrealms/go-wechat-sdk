# merchant/developed 模块（微信支付 · 商户模式）

`github.com/godrealms/go-wechat-sdk/merchant/developed` — 包名：`pay`

这是 SDK 里最完整的业务模块，覆盖微信支付 V3 协议的商户模式（直连商户）。所有请求都走相同的签名/验签通道：请求侧签 `WECHATPAY2-SHA256-RSA2048`，响应侧用平台证书校验 `Wechatpay-Signature`，回调通知额外做 AEAD_AES_256_GCM 解密。

推荐导入别名：

```go
import pay "github.com/godrealms/go-wechat-sdk/merchant/developed"
```

## 1. 类型与构造

### Config

```go
type Config struct {
    Appid             string            // 公众号 / APP / 小程序的 AppId
    Mchid             string            // 商户号
    CertificateNumber string            // 商户证书序列号
    APIv3Key          string            // 商户 APIv3 密钥（32 字节字符串）
    PrivateKey        *rsa.PrivateKey   // 商户私钥（PKCS#1 或 PKCS#8 均可）
    Certificate       *x509.Certificate // 商户证书（用于内部引用序列号等）
}
```

### Client

```go
type Client struct { /* 不导出字段 */ }

func NewClient(cfg Config) (*Client, error)
```

`NewClient` 会校验必填字段，缺任何一个都会立刻报错，避免等到实际请求时才发现。

### 访问器

```go
func (c *Client) Appid() string
func (c *Client) Mchid() string
func (c *Client) CertificateNumber() string
func (c *Client) PrivateKey() *rsa.PrivateKey
func (c *Client) Certificate() *x509.Certificate
func (c *Client) HTTP() *utils.HTTP
```

### 低层通道（供扩展使用）

```go
func (c *Client) PostV3Raw(ctx context.Context, urlPath string, body any, result any) error
func (c *Client) GetV3Raw(ctx context.Context, urlPath string, query url.Values, result any) error
func (c *Client) DoV3(
    ctx context.Context,
    method, urlPath string,
    query url.Values,
    body any,
    extraHeaders http.Header,
    result any,
) error
```

三个方法都暴露核心的 `doV3` 能力，`merchant/service` 就是用它们来实现服务商模式的：

- `PostV3Raw` / `GetV3Raw` 是最常用的两种方法的快捷封装；
- `DoV3` 是最通用的转发入口，支持任意 HTTP 方法以及**自定义请求头**。典型使用场景：
  - 需要 `PUT`/`PATCH`（例如修改结算账户）；
  - 需要通过 `Wechatpay-Serial` 头告诉服务端本次敏感字段是用哪一张平台证书加密的（进件、分账接收方加密姓名等）；
  - 需要塞 `Idempotency-Key` 等业务头。

调用方追加的头部会和 SDK 管理的头合并写入请求，但 `Accept` / `Authorization` / `User-Agent` / `Content-Type` 由 SDK 最终覆盖——调用方**无法**污染签名相关的关键头。

## 2. 下单接口

所有下单都接受一个 `*types.Transactions`，这是同一套结构体，微信文档里 JSAPI/APP/H5/Native 的参数也基本一致。

```go
func (c *Client) TransactionsJsapi(ctx context.Context, order *types.Transactions) (*types.TransactionsJsapiResp, error)
func (c *Client) TransactionsApp(ctx context.Context, order *types.Transactions)   (*types.TransactionsAppResponse, error)
func (c *Client) TransactionsH5(ctx context.Context, order *types.Transactions)    (*types.TransactionsH5Resp, error)
func (c *Client) TransactionsNative(ctx context.Context, order *types.Transactions) (*types.TransactionsNativeResp, error)

// Modify* 是旧版预留接口，语义和对应的下单一致，多返回一份前端签名好的参数
func (c *Client) ModifyTransactionsJsapi(ctx context.Context, order *types.Transactions) (*types.TransactionsJsapi, error)
func (c *Client) ModifyTransactionsApp(ctx context.Context, order *types.Transactions)   (*types.ModifyAppResponse, error)
```

`types.Transactions` 的核心字段（省略注释）：

```go
type Transactions struct {
    Appid, Mchid, Description, OutTradeNo, NotifyUrl string
    Amount    *Amount        // {Total int, Currency string}
    Payer     *Payer         // {Openid string}（JSAPI 时必填）
    TimeExpire, Attach, GoodsTag string
    SupportFapiao bool
    SceneInfo *SceneInfo
    // 还有代金券、结算信息等选填字段
}
```

### 使用案例：JSAPI 下单

```go
resp, err := client.TransactionsJsapi(ctx, &types.Transactions{
    Appid:       "wx1234567890",
    Mchid:       "1900000001",
    Description: "测试商品",
    OutTradeNo:  "ord-" + utils.RandomString(16),
    NotifyUrl:   "https://yourhost.com/wxpay/notify",
    Amount:      &types.Amount{Total: 1, Currency: "CNY"}, // 1 分
    Payer:       &types.Payer{Openid: "o-xxxxxx"},
})
if err != nil { return err }
// resp.PrepayId 交给前端 wx.chooseWXPay 唤起支付
```

## 3. 订单查询与关单

```go
func (c *Client) QueryTransactionId(ctx context.Context, transactionId string) (*types.QueryResponse, error)
func (c *Client) QueryOutTradeNo(ctx context.Context, outTradeNo string)       (*types.QueryResponse, error)
func (c *Client) TransactionsClose(ctx context.Context, outTradeNo string) error
```

`QueryResponse` 包含微信返回的完整字段（`TradeState` / `TransactionId` / `Amount` / `Payer` 等）。关单只关心 HTTP 是否成功，成功返回 204，SDK 内部已经正确处理了空响应体。

### 使用案例

```go
q, err := client.QueryOutTradeNo(ctx, "ord-abc123")
if err != nil { return err }
if q.TradeState == "SUCCESS" {
    // 已支付，做发货等业务逻辑
}

// 超时未支付，主动关单
if q.TradeState == "NOTPAY" && time.Since(createdAt) > 30*time.Minute {
    _ = client.TransactionsClose(ctx, "ord-abc123")
}
```

## 4. 退款

```go
func (c *Client) Refunds(ctx context.Context, refund *types.Refunds)     (*types.RefundResp, error)
func (c *Client) QueryRefunds(ctx context.Context, outRefundNo string)   (*types.RefundResp, error)
func (c *Client) ApplyAbnormalRefund(ctx context.Context, body any)      (*types.RefundResp, error) // 异常退款
```

### 使用案例

```go
refund := &types.Refunds{
    OutTradeNo:  "ord-abc123",
    OutRefundNo: "rf-" + utils.RandomString(16),
    Amount: &types.RefundAmount{
        Refund:   1,
        Total:    1,
        Currency: "CNY",
    },
    Reason: "测试退款",
}
r, err := client.Refunds(ctx, refund)
```

## 5. 对账单

```go
func (c *Client) TradeBill(ctx context.Context, q *types.TradeBillQuest)    (*types.BillResp, error)
func (c *Client) FundFlowBill(ctx context.Context, q *types.FundsBillQuest) (*types.BillResp, error)
```

返回 `BillResp` 里带有 `DownloadUrl`，按照该 URL 再发一次 GET 即可拿到 gzip 压缩的账单。

### 使用案例

```go
bill, err := client.TradeBill(ctx, &types.TradeBillQuest{
    BillDate: "2026-04-11",
    BillType: "ALL",
})
log.Println("download from:", bill.DownloadUrl)
```

## 6. 回调通知解析 ⭐

**这是本 SDK 相对官方仓库最重要的增值功能**。普通用法只要一个调用就能完成：验签 → 解密 → Unmarshal。

```go
func (c *Client) ParseNotification(
    ctx context.Context,
    r *http.Request,
    result any,  // 传 nil 则仅返回外层 Notify
) (*types.Notify, error)

// 回复微信
func AckNotification(w http.ResponseWriter)
func FailNotification(w http.ResponseWriter, message string)
```

内部流程：

1. 读 `r.Body` 全部字节（SDK 会自动关闭 body）。
2. 取 `Wechatpay-Timestamp / Nonce / Signature / Serial` 头。
3. 用本地缓存的平台证书验证响应签名；若 serial 不在缓存中，自动调 `/v3/certificates` 拉一次。
4. JSON 解析外层通知 `{id, event_type, resource{…}}`。
5. 对 `resource.ciphertext` 做 AEAD_AES_256_GCM 解密。
6. 如果调用方传了 `result`，把明文 Unmarshal 到 `result`。

### 使用案例：处理 JSAPI 支付成功回调

```go
http.HandleFunc("/wxpay/notify", func(w http.ResponseWriter, r *http.Request) {
    var txn struct {
        TransactionId string `json:"transaction_id"`
        OutTradeNo    string `json:"out_trade_no"`
        TradeState    string `json:"trade_state"`
        Amount        struct {
            Total int `json:"total"`
        } `json:"amount"`
    }
    notify, err := client.ParseNotification(r.Context(), r, &txn)
    if err != nil {
        pay.FailNotification(w, err.Error())
        return
    }

    if notify.EventType == "TRANSACTION.SUCCESS" && txn.TradeState == "SUCCESS" {
        if err := markOrderPaid(txn.OutTradeNo, txn.TransactionId, txn.Amount.Total); err != nil {
            pay.FailNotification(w, "internal") // 微信将重试
            return
        }
    }
    pay.AckNotification(w) // 200 OK + {"code":"SUCCESS","message":"成功"}
})
```

> 关键：只有在 `AckNotification` 之前业务处理完成（且幂等），才算安全闭环。返回 5xx 会让微信重推。

## 7. 平台证书管理 & 敏感字段加密

大多数时候不用手动调，`ParseNotification` 会按需拉取。但可以在启动阶段主动预热：

```go
func (c *Client) FetchPlatformCertificates(ctx context.Context) ([]*x509.Certificate, error)
func (c *Client) AddPlatformCertificate(cert *x509.Certificate)
```

### 敏感字段加密

微信支付大量接口（子商户进件、分账接收方姓名、退款申请人姓名等）要求对敏感字段做 RSA-OAEP(SHA256) 加密后再上送，同时请求头 `Wechatpay-Serial` 必须带上加密所用平台证书的序列号。SDK 提供两步到位的封装：

```go
// 从缓存取一张（或主动拉一次）可用的平台证书，同时返回它的序列号。
func (c *Client) PlatformCertForEncrypt(ctx context.Context) (*x509.Certificate, string, error)

// 用给定证书的公钥做 RSA-OAEP(SHA256) + base64。
func EncryptSensitiveField(cert *x509.Certificate, plaintext string) (string, error)
```

典型用法：

```go
cert, serial, err := client.PlatformCertForEncrypt(ctx)
if err != nil { return err }
encName, _ := pay.EncryptSensitiveField(cert, "张三")

headers := http.Header{"Wechatpay-Serial": []string{serial}}
_ = client.DoV3(ctx, http.MethodPost, "/v3/some-endpoint", nil, map[string]any{
    "name": encName,
    // ...
}, headers, &resp)
```

`merchant/service` 包把这一整套流程再封装一层（`EncryptSensitive` 一步返回密文+序列号）——服务商模式的用户应该优先使用那一层，见 [merchant-service.md](./merchant-service.md)。

### 使用案例

```go
// 应用启动时先拉一次
if _, err := client.FetchPlatformCertificates(ctx); err != nil {
    log.Printf("warn: preload platform certs failed: %v", err)
}
```

## 8. 分账（Profit Sharing）

```go
func (c *Client) ProfitSharingOrder(ctx context.Context, body any) (map[string]any, error)
func (c *Client) ProfitSharingQueryOrder(ctx context.Context, outOrderNo, transactionId string) (map[string]any, error)
func (c *Client) ProfitSharingReturn(ctx context.Context, body any) (map[string]any, error)
func (c *Client) ProfitSharingQueryReturn(ctx context.Context, outReturnNo, outOrderNo string) (map[string]any, error)
func (c *Client) ProfitSharingUnfreeze(ctx context.Context, body any) (map[string]any, error)
func (c *Client) ProfitSharingMerchantAmounts(ctx context.Context, transactionId string) (map[string]any, error)
func (c *Client) ProfitSharingAddReceiver(ctx context.Context, body any) (map[string]any, error)
func (c *Client) ProfitSharingDeleteReceiver(ctx context.Context, body any) (map[string]any, error)
func (c *Client) ProfitSharingBills(ctx context.Context, billDate, tarType string) (map[string]any, error)
```

请求体用 `any`（实际通常是 `map[string]any` 或你自己的 struct），响应体用 `map[string]any`——这样在微信不定期新增字段时也能透传，不会被 SDK 卡住。如果你需要强类型，在业务层自行定义 struct 再 `json.Unmarshal` 即可。

### 使用案例

```go
resp, err := client.ProfitSharingOrder(ctx, map[string]any{
    "appid":         "wx1234567890",
    "transaction_id": "42000012345",
    "out_order_no":  "ps-" + utils.RandomString(16),
    "receivers": []map[string]any{
        {
            "type":        "MERCHANT_ID",
            "account":     "1900000002",
            "amount":      80,
            "description": "运营服务",
        },
    },
    "unfreeze_unsplit": true,
})
```

## 9. 合单支付（Combine）

```go
func (c *Client) CombineTransactionsJsapi(ctx context.Context, body any)  (map[string]any, error)
func (c *Client) CombineTransactionsApp(ctx context.Context, body any)    (map[string]any, error)
func (c *Client) CombineTransactionsH5(ctx context.Context, body any)     (map[string]any, error)
func (c *Client) CombineTransactionsNative(ctx context.Context, body any) (map[string]any, error)
func (c *Client) QueryCombineOrder(ctx context.Context, combineOutTradeNo string)  (map[string]any, error)
func (c *Client) CloseCombineOrder(ctx context.Context, combineOutTradeNo string, body any) error
```

## 10. 代金券 / Favor

```go
func (c *Client) FavorCreateStock(ctx context.Context, body any) (map[string]any, error)
func (c *Client) FavorStartStock(ctx context.Context, stockId string, body any) (map[string]any, error)
func (c *Client) FavorPauseStock(ctx context.Context, stockId string, body any) error
func (c *Client) FavorRestartStock(ctx context.Context, stockId string, body any) error
func (c *Client) FavorSendCoupon(ctx context.Context, openid string, body any) (map[string]any, error)
func (c *Client) FavorQueryStock(ctx context.Context, stockId, stockCreatorMchid string) (map[string]any, error)
func (c *Client) FavorQueryUserCoupon(ctx context.Context, openid, couponId, appid string) (map[string]any, error)
```

## 11. 错误处理

HTTP 错误统一走 `*utils.HTTPError`：

```go
_, err := client.TransactionsJsapi(ctx, order)
var httpErr *utils.HTTPError
if errors.As(err, &httpErr) {
    // httpErr.Body 里有微信返回的详细 JSON 错误信息
    // httpErr.StatusCode 是 HTTP 状态码
}
```

响应验签失败会在 error 里包含 `"response signature invalid"` 字样。

## 12. 并发语义

`*pay.Client` 是线程安全的：

- 证书缓存用 `sync.RWMutex` 保护。
- 每次请求独立构造签名与 header，**不共享 header 对象**，所以多 goroutine 并发调用是安全的。
- 内部 HTTP 客户端的超时默认 30s，可以通过 `HTTP()` 访问器替换。

## 13. 完整接入案例

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"

    pay "github.com/godrealms/go-wechat-sdk/merchant/developed"
    "github.com/godrealms/go-wechat-sdk/merchant/developed/types"
    "github.com/godrealms/go-wechat-sdk/utils"
)

func main() {
    priv, _ := utils.LoadPrivateKeyWithPath("apiclient_key.pem")
    cert, _ := utils.LoadCertificateWithPath("apiclient_cert.pem")

    client, err := pay.NewClient(pay.Config{
        Appid:             "wx1234567890",
        Mchid:             "1900000001",
        CertificateNumber: utils.GetCertificateSerialNumber(*cert),
        APIv3Key:          "01234567890123456789012345678901",
        PrivateKey:        priv,
        Certificate:       cert,
    })
    if err != nil { log.Fatal(err) }

    ctx := context.Background()
    _, _ = client.FetchPlatformCertificates(ctx) // 预热

    http.HandleFunc("/api/pay/create", func(w http.ResponseWriter, r *http.Request) {
        resp, err := client.TransactionsJsapi(r.Context(), &types.Transactions{
            Appid:       "wx1234567890",
            Mchid:       "1900000001",
            Description: "测试商品",
            OutTradeNo:  "ord-" + utils.RandomString(16),
            NotifyUrl:   "https://example.com/wxpay/notify",
            Amount:      &types.Amount{Total: 1, Currency: "CNY"},
            Payer:       &types.Payer{Openid: r.URL.Query().Get("openid")},
        })
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        _ = json.NewEncoder(w).Encode(resp)
    })

    http.HandleFunc("/wxpay/notify", func(w http.ResponseWriter, r *http.Request) {
        var txn struct {
            TransactionId, OutTradeNo, TradeState string
        }
        _, err := client.ParseNotification(r.Context(), r, &txn)
        if err != nil {
            pay.FailNotification(w, err.Error())
            return
        }
        log.Printf("paid: %s %s %s", txn.TradeState, txn.OutTradeNo, txn.TransactionId)
        pay.AckNotification(w)
    })

    srv := &http.Server{Addr: ":8080", Handler: http.DefaultServeMux, ReadHeaderTimeout: 5 * time.Second}
    log.Fatal(srv.ListenAndServe())
}
```
