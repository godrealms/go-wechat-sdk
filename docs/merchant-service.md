# merchant/service 模块（微信支付 · 服务商模式）

`github.com/godrealms/go-wechat-sdk/merchant/service` — 包名：`service`

服务商模式与直连商户模式在 V3 协议层面几乎完全相同：都用**服务商自己的**商户号（`sp_mchid`）+ 证书签名，区别仅在于下单等请求体里要额外携带 `sub_mchid` / `sub_appid` 字段，以及部分接口路径从 `/v3/pay/transactions/*` 变成 `/v3/pay/partner/transactions/*`。

基于这一点，本包**没有重写一整套签名/验签逻辑**，而是封装 `merchant/developed.Client`，在调用前自动把 `sp_/sub_` 字段注入到请求体里。

## 1. 类型

### Config

```go
type Config struct {
    SpMchid           string            // 服务商商户号
    SpAppid           string            // 服务商 AppId
    SubMchid          string            // 默认子商户号（必填，可在单次调用时覆盖）
    SubAppid          string            // 默认子商户 AppId（选填）
    CertificateNumber string
    APIv3Key          string
    PrivateKey        *rsa.PrivateKey
    Certificate       *x509.Certificate
}
```

### Client

```go
type Client struct { /* 不导出字段 */ }

func NewClient(cfg Config) (*Client, error)

func (c *Client) Inner() *pay.Client  // 返回底层的 developed.Client
func (c *Client) SubMchid() string
func (c *Client) SubAppid() string
```

`Inner()` 非常重要——**所有商户模式已经封装好的方法都可以通过 `Inner()` 直接调用**（比如退款、对账单、分账、代金券、平台证书管理、`ParseNotification` 等）。本包只额外提供下单接口的 partner 版本。

## 2. 服务商下单接口

```go
func (c *Client) PartnerTransactionsJsapi(ctx, body map[string]any)  (map[string]any, error)
func (c *Client) PartnerTransactionsApp(ctx, body map[string]any)    (map[string]any, error)
func (c *Client) PartnerTransactionsH5(ctx, body map[string]any)     (map[string]any, error)
func (c *Client) PartnerTransactionsNative(ctx, body map[string]any) (map[string]any, error)
```

调用前，SDK 会自动检查 `body` 里是否已经含有以下键；**没有的才注入默认值**，有的话尊重调用方：

| 字段 | 默认值来源 |
|---|---|
| `sp_mchid` | `Config.SpMchid` |
| `sp_appid` | `Config.SpAppid` |
| `sub_mchid` | `Config.SubMchid` |
| `sub_appid` | `Config.SubAppid`（仅当非空） |

这样你在写业务代码时，**不需要关心服务商自己的字段**，只填本次业务的订单信息就够了。

## 3. 使用案例：服务商 JSAPI 下单

```go
package main

import (
    "context"
    "log"

    pay "github.com/godrealms/go-wechat-sdk/merchant/developed"
    "github.com/godrealms/go-wechat-sdk/merchant/service"
    "github.com/godrealms/go-wechat-sdk/utils"
)

func main() {
    priv, _ := utils.LoadPrivateKeyWithPath("sp_apiclient_key.pem")
    cert, _ := utils.LoadCertificateWithPath("sp_apiclient_cert.pem")

    client, err := service.NewClient(service.Config{
        SpMchid:           "1900000001", // 服务商自己的商户号
        SpAppid:           "wxSpAppId",
        SubMchid:          "1900000100", // 子商户
        SubAppid:          "wxSubAppId",
        CertificateNumber: utils.GetCertificateSerialNumber(*cert),
        APIv3Key:          "01234567890123456789012345678901",
        PrivateKey:        priv,
        Certificate:       cert,
    })
    if err != nil { log.Fatal(err) }

    ctx := context.Background()

    // 这里只填业务字段——sp_mchid/sp_appid/sub_mchid 会自动补上
    resp, err := client.PartnerTransactionsJsapi(ctx, map[string]any{
        "description":  "服务商代下单",
        "out_trade_no": "sp-" + utils.RandomString(16),
        "notify_url":   "https://example.com/wxpay/notify",
        "amount":       map[string]any{"total": 1, "currency": "CNY"},
        "payer":        map[string]any{"sub_openid": "o-xxxxxx"},
    })
    if err != nil { log.Fatal(err) }

    prepayId := resp["prepay_id"]
    log.Printf("prepay_id = %v", prepayId)

    // 需要关单 / 退款 / 回调处理？直接用 Inner() 复用商户接口
    _ = client.Inner().TransactionsClose(ctx, "sp-abc123")
    _, _ = client.Inner().FetchPlatformCertificates(ctx)

    _ = pay.Config{} // 只是示意这里也能直接 new 一个 developed 的
}
```

## 4. 使用案例：服务商模式下的回调处理

回调处理代码和商户模式**完全一致**，直接用 `Inner().ParseNotification` 即可。因为 V3 回调的验签机制不区分商户/服务商，只看平台证书和 APIv3 密钥。

```go
http.HandleFunc("/wxpay/notify", func(w http.ResponseWriter, r *http.Request) {
    var txn struct {
        TransactionId string `json:"transaction_id"`
        OutTradeNo    string `json:"out_trade_no"`
        SpMchid       string `json:"sp_mchid"`
        SubMchid      string `json:"sub_mchid"`
        TradeState    string `json:"trade_state"`
    }
    _, err := client.Inner().ParseNotification(r.Context(), r, &txn)
    if err != nil {
        pay.FailNotification(w, err.Error())
        return
    }
    log.Printf("sub=%s ord=%s state=%s", txn.SubMchid, txn.OutTradeNo, txn.TradeState)
    pay.AckNotification(w)
})
```

## 5. 使用案例：为多个子商户按需切换

`Config.SubMchid` 是"默认"子商户号。如果你需要在同一个服务商账户下，为多个子商户分别下单，最直接的做法有两种：

**方案 A：每个子商户一个 Client（推荐，语义清晰）**

```go
clientByMerchant := map[string]*service.Client{}
for _, sub := range subMerchants {
    c, _ := service.NewClient(service.Config{
        SpMchid:           spMchid,
        SpAppid:           spAppid,
        SubMchid:          sub.Mchid,
        SubAppid:          sub.Appid,
        CertificateNumber: certSerial,
        APIv3Key:          apiV3,
        PrivateKey:        priv,
        Certificate:       cert,
    })
    clientByMerchant[sub.Mchid] = c
}
```

**方案 B：单个 Client + 手动覆盖字段**

```go
// 用默认 Client 单次调用时显式传 sub_mchid，会覆盖 Config 里的值
_, _ = client.PartnerTransactionsJsapi(ctx, map[string]any{
    "sub_mchid":   "1900000999", // 显式指定，SDK 不再自动注入
    "description": "一次性下单",
    // ...
})
```

## 6. 局限

本包目前**只封装了下单接口**。如果服务商业务需要：

- 子商户进件
- 子商户查询
- 服务商侧的分账配置

请直接通过 `Inner().PostV3Raw` / `Inner().GetV3Raw` 手动调用对应的微信文档 URL。基础设施（签名/验签/证书管理）全都复用。

## 7. 并发语义

与 `pay.Client` 完全一致——线程安全，可在多 goroutine 共享。所有核心签名/验签逻辑都继承自 `pay.Client`，本包只是做了业务层的字段注入。
