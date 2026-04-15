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

## 6. 子商户进件（特约商户进件 / applyment4sub）

服务商把子商户资料提交给微信支付审核。文档：<https://pay.weixin.qq.com/docs/partner/apis/partner-applyment/applyments.html>

```go
// 提交进件单
func (c *Client) ApplymentSubmit(ctx context.Context, body any, platformSerial string) (*ApplymentSubmitResponse, error)

// 用业务申请编号查询
func (c *Client) ApplymentQueryByBusinessCode(ctx context.Context, businessCode string) (*ApplymentQueryResponse, error)

// 用微信侧 applyment_id 查询
func (c *Client) ApplymentQueryByID(ctx context.Context, applymentID int64) (*ApplymentQueryResponse, error)

// 敏感字段加密助手：一步返回 (ciphertext, platformSerial)
func (c *Client) EncryptSensitive(ctx context.Context, plaintext string) (string, string, error)
```

返回类型：

```go
type ApplymentSubmitResponse struct {
    ApplymentID int64 `json:"applyment_id"`
}

type ApplymentQueryResponse struct {
    BusinessCode      string        `json:"business_code,omitempty"`
    ApplymentID       int64         `json:"applyment_id,omitempty"`
    SubMchid          string        `json:"sub_mchid,omitempty"`    // 审核通过后返回
    SignURL           string        `json:"sign_url,omitempty"`     // 超管签约链接
    ApplymentState    string        `json:"applyment_state"`        // 例如 APPLYMENT_STATE_FINISHED
    ApplymentStateMsg string        `json:"applyment_state_msg"`
    AuditDetail       []AuditDetail `json:"audit_detail,omitempty"` // 被驳回的字段明细
}

type AuditDetail struct {
    ParamName    string `json:"param_name"`
    RejectReason string `json:"reject_reason"`
}
```

### 使用案例：提交进件

所有姓名、身份证号、手机号、银行卡号等敏感字段**必须先加密**再放进请求体。`EncryptSensitive` 每次返回的 `platformSerial` 在同一次提交里必须是**同一张**平台证书（否则服务端解不出来）——简单起见，在准备一整个请求体前只调一次就行：

```go
ctx := context.Background()

// 1) 先拉一张平台证书用于加密本次所有敏感字段
_, serial, err := client.EncryptSensitive(ctx, "") // 随便加密一次触发拉取；忽略结果
if err != nil { log.Fatal(err) }

encrypt := func(s string) string {
    ct, _, err := client.EncryptSensitive(ctx, s)
    if err != nil { log.Fatal(err) }
    return ct
}

body := map[string]any{
    "business_code": "1900013511_10000",
    "contact_info": map[string]any{
        "contact_type":      "LEGAL",
        "contact_name":      encrypt("张三"),
        "contact_id_number": encrypt("110101199001011234"),
        "mobile_phone":      encrypt("13800138000"),
        "contact_email":     encrypt("zs@example.com"),
    },
    "subject_info": map[string]any{
        "subject_type": "SUBJECT_TYPE_ENTERPRISE",
        "business_license_info": map[string]any{
            "license_copy":   "MEDIA_ID_xxx", // 用 /v3/merchant/media/upload 上传得到
            "license_number": "91440300MA5xxxxxxx",
            "merchant_name":  "深圳示例有限公司",
            "legal_person":   "张三",
        },
    },
    "business_info": map[string]any{
        "merchant_shortname": "示例小店",
        "service_phone":      "075588888888",
        "sales_info": map[string]any{
            "sales_scenes_type": []string{"SALES_SCENES_STORE"},
        },
    },
    "settlement_info": map[string]any{
        "settlement_id":        "719",
        "qualification_type":   "餐饮",
    },
    "bank_account_info": map[string]any{
        "bank_account_type": "BANK_ACCOUNT_TYPE_CORPORATE",
        "account_name":      encrypt("深圳示例有限公司"),
        "account_bank":      "工商银行",
        "bank_address_code": "110000",
        "account_number":    encrypt("6222021234567890123"),
    },
}

resp, err := client.ApplymentSubmit(ctx, body, serial)
if err != nil { log.Fatal(err) }
log.Printf("applyment_id = %d", resp.ApplymentID)
```

### 使用案例：轮询进件状态

```go
// 刚提交时还没有 applyment_id，先用 business_code 查一次拿到
st, err := client.ApplymentQueryByBusinessCode(ctx, "1900013511_10000")
if err != nil { return err }
log.Printf("state=%s msg=%s", st.ApplymentState, st.ApplymentStateMsg)

// 之后可以直接走 applyment_id
st, _ = client.ApplymentQueryByID(ctx, st.ApplymentID)

// 如果被驳回，audit_detail 会告诉你是哪些字段
for _, d := range st.AuditDetail {
    log.Printf("  %s: %s", d.ParamName, d.RejectReason)
}

// 审核通过后拿到 sub_mchid 和超管签约链接
if st.ApplymentState == "APPLYMENT_STATE_FINISHED" {
    log.Printf("sub_mchid=%s sign_url=%s", st.SubMchid, st.SignURL)
}
```

## 7. 服务商侧的分账配置

> 若只是"发起分账"（`ProfitSharingOrder`）、查询分账结果等，请直接用 `Inner()` 调用 `pay.Client` 的既有方法——服务商和直连商户走同一套路径，只是 body 里要带 `sub_mchid`。
>
> 本节专门覆盖的是**给子商户配置分账**的三个管理类接口。

```go
// 添加分账接收方（不带敏感字段加密）
func (c *Client) ProfitSharingAddReceiver(ctx context.Context, body map[string]any) (map[string]any, error)

// 添加分账接收方，同时携带 Wechatpay-Serial 头——用于 receiver.name 已加密的场景
func (c *Client) ProfitSharingAddReceiverWithSerial(ctx context.Context, body map[string]any, platformSerial string) (map[string]any, error)

// 删除分账接收方
func (c *Client) ProfitSharingDeleteReceiver(ctx context.Context, body map[string]any) (map[string]any, error)

// 查询某个子商户的最大分账比例（max_ratio 单位为万分比，例如 2000 表示 20%）
func (c *Client) ProfitSharingMerchantConfig(ctx context.Context, subMchid string) (map[string]any, error)
```

这四个方法会自动把 `appid` / `sub_mchid` 填成 `Client` 初始化时配置的默认值；调用方显式提供的字段优先。`ProfitSharingMerchantConfig` 的 `subMchid` 为空时使用默认 sub_mchid。

### 使用案例：添加一个商户号型接收方

```go
_, err := client.ProfitSharingAddReceiver(ctx, map[string]any{
    "type":          "MERCHANT_ID",
    "account":       "1900000100",
    "relation_type": "SUPPLIER",
})
```

### 使用案例：添加一个个人 openid 接收方（name 需要加密）

```go
encName, serial, err := client.EncryptSensitive(ctx, "张三")
if err != nil { return err }

_, err = client.ProfitSharingAddReceiverWithSerial(ctx, map[string]any{
    "type":          "PERSONAL_OPENID",
    "account":       "oUpF8uMuAJO_M2pxb1Q9zNjWeS6o",
    "name":          encName,         // 已加密的姓名
    "relation_type": "USER",
}, serial)
```

### 使用案例：查询最大分账比例

```go
resp, err := client.ProfitSharingMerchantConfig(ctx, "") // 空 -> 用默认 sub_mchid
if err != nil { return err }
log.Printf("max_ratio = %v (万分比)", resp["max_ratio"])
```

### 使用案例：删除接收方

```go
_, err := client.ProfitSharingDeleteReceiver(ctx, map[string]any{
    "type":    "PERSONAL_OPENID",
    "account": "oUpF8uMuAJO_M2pxb1Q9zNjWeS6o",
})
```

## 8. 子商户结算账户查询 / 修改

进件完成后，需要查看或更改子商户的收款结算账户时使用。

```go
// 查询
func (c *Client) SettlementQuery(ctx context.Context, subMchid string) (*SettlementInfo, error)

// 修改（account_number 需为调用方已经用平台证书加密过的密文，serial 为加密所用证书序列号）
func (c *Client) SettlementModify(
    ctx context.Context,
    subMchid string,
    req *SettlementModifyRequest,
    platformSerial string,
) error

// 修改（高层封装，接受明文 account_number，SDK 自动加密并填 Wechatpay-Serial 头）
func (c *Client) SettlementModifyEncrypted(
    ctx context.Context,
    subMchid string,
    req *SettlementModifyRequest,
    plaintextAccountNumber string,
) error
```

类型：

```go
type SettlementInfo struct {
    AccountType      string `json:"account_type"`                 // ACCOUNT_TYPE_BUSINESS / ACCOUNT_TYPE_PRIVATE
    AccountBank      string `json:"account_bank"`                 // "工商银行"
    BankAddressCode  string `json:"bank_address_code,omitempty"`  // 行政区划代码
    BankBranchID     string `json:"bank_branch_id,omitempty"`     // 联行号
    BankName         string `json:"bank_name,omitempty"`          // 支行全称
    AccountNumber    string `json:"account_number"`               // 查询返回：平台证书加密后的密文
    VerifyResult     string `json:"verify_result,omitempty"`      // 最近一次打款验证结果
    VerifyFailReason string `json:"verify_fail_reason,omitempty"` // 失败原因
}

type SettlementModifyRequest struct {
    ModifyBalance   bool   `json:"modify_balance"`              // 是否同时修改出款账户
    AccountType     string `json:"account_type"`
    AccountBank     string `json:"account_bank"`
    BankAddressCode string `json:"bank_address_code,omitempty"`
    BankName        string `json:"bank_name,omitempty"`
    BankBranchID    string `json:"bank_branch_id,omitempty"`
    AccountNumber   string `json:"account_number"`              // 平台证书加密后的密文
}
```

`subMchid` 为空时均使用 `Config.SubMchid`。修改接口成功时服务端返回 200 + 空响应体，SDK 会当作成功处理。

> **注意**：`SettlementQuery` 返回的 `AccountNumber` 是**平台证书加密后的密文**；调用方拿到之后需要**用商户 API 证书私钥**做 RSA-OAEP(SHA256) 解密才能得到明文。SDK 故意**不做自动解密**，避免把敏感明文长期保留在进程内存中。

### 使用案例：查询并解密结算账号

```go
info, err := client.SettlementQuery(ctx, "") // 使用默认 sub_mchid
if err != nil { return err }
log.Printf("bank=%s type=%s verify=%s", info.AccountBank, info.AccountType, info.VerifyResult)

// 调用方用自己保存的商户 API 私钥解密 account_number（SDK 不替你做）
plain, err := utils.DecryptOAEP(info.AccountNumber, myMerchantPrivateKey)
if err != nil { return err }
defer zero(plain) // 用完及时清零
```

### 使用案例：修改结算账号（明文 → SDK 自动加密）

```go
err := client.SettlementModifyEncrypted(ctx, "", &service.SettlementModifyRequest{
    ModifyBalance:   true,
    AccountType:     "ACCOUNT_TYPE_BUSINESS",
    AccountBank:     "工商银行",
    BankAddressCode: "110000",
    BankName:        "工商银行股份有限公司上海市分行营业部",
    BankBranchID:    "402713354941",
    // 不填 AccountNumber —— 由下面的 plaintext 参数接管
}, "6222021234567890123")
```

### 使用案例：自己加密后调用（想复用同一张平台证书多字段加密）

```go
cipher, serial, err := client.EncryptSensitive(ctx, "6222021234567890123")
if err != nil { return err }

err = client.SettlementModify(ctx, "", &service.SettlementModifyRequest{
    AccountType:   "ACCOUNT_TYPE_BUSINESS",
    AccountBank:   "工商银行",
    AccountNumber: cipher, // 已是密文
}, serial)
```

## 9. 服务商侧分账 — 动态接口（partner 变体）

第 7 节覆盖了**管理类**的三个静态接口（add/delete receiver、查询最大分账比例）。本节补齐与**具体订单 / 回退单 / 剩余金额 / 账单**相关的动态接口。

这组接口与直连商户共享同一组 REST 路径；服务商版的差异只在：请求体 / query 里必须带 `sub_mchid`（POST 类还要带 `appid`）。SDK 会自动填充，调用方显式提供的字段优先。

```go
// 发起分账
func (c *Client) ProfitSharingOrder(ctx context.Context, body map[string]any) (map[string]any, error)

// 发起分账 + 携带 Wechatpay-Serial 头（当 receivers 中有加密过的 name 时必填）
func (c *Client) ProfitSharingOrderWithSerial(ctx context.Context, body map[string]any, platformSerial string) (map[string]any, error)

// 查询分账结果
func (c *Client) ProfitSharingQueryOrder(ctx context.Context, subMchid, outOrderNo, transactionId string) (map[string]any, error)

// 请求分账回退
func (c *Client) ProfitSharingReturn(ctx context.Context, body map[string]any) (map[string]any, error)

// 查询分账回退
func (c *Client) ProfitSharingQueryReturn(ctx context.Context, subMchid, outReturnNo, outOrderNo string) (map[string]any, error)

// 解冻剩余金额
func (c *Client) ProfitSharingUnfreeze(ctx context.Context, body map[string]any) (map[string]any, error)

// 查询某笔交易的剩余待分金额
func (c *Client) ProfitSharingMerchantAmounts(ctx context.Context, subMchid, transactionId string) (map[string]any, error)

// 申请分账账单（按日）
func (c *Client) ProfitSharingBills(ctx context.Context, billDate, subMchid, tarType string) (map[string]any, error)
```

所有 `subMchid` 参数为空时使用 `Config.SubMchid`。POST 类方法的 body 中 `appid`/`sub_mchid` 缺省时会自动填为默认值，调用方显式写了则保留。

### 使用案例：发起一次分账（同步解冻）

```go
resp, err := client.ProfitSharingOrder(ctx, map[string]any{
    "transaction_id": "4208450740201411110007820472",
    "out_order_no":   "P" + utils.RandomString(16),
    "receivers": []map[string]any{
        {
            "type":        "MERCHANT_ID",
            "account":     "1900000109",
            "amount":      100,
            "description": "分给供应商",
        },
    },
    "unfreeze_unsplit": true,
})
if err != nil { return err }
log.Printf("order_id = %v state = %v", resp["order_id"], resp["state"])
```

### 使用案例：发起分账 + 加密的个人 openid 接收方

```go
encName, serial, err := client.EncryptSensitive(ctx, "张三")
if err != nil { return err }

_, err = client.ProfitSharingOrderWithSerial(ctx, map[string]any{
    "transaction_id": txId,
    "out_order_no":   "P" + utils.RandomString(16),
    "receivers": []map[string]any{
        {
            "type":        "PERSONAL_OPENID",
            "account":     "oUpF8uMuAJO_M2pxb1Q9zNjWeS6o",
            "name":        encName, // 已加密
            "amount":      50,
            "description": "分给分销员",
        },
    },
    "unfreeze_unsplit": false,
}, serial)
```

### 使用案例：查询分账结果 / 回退 / 剩余金额

```go
// 查询分账结果
orderResp, err := client.ProfitSharingQueryOrder(ctx, "", "P20150806125346", "4208450740201411110007820472")

// 请求回退
_, err = client.ProfitSharingReturn(ctx, map[string]any{
    "order_id":      orderResp["order_id"],
    "out_return_no": "R" + utils.RandomString(16),
    "return_mchid":  "1900000109",
    "amount":        30,
    "description":   "协商退款",
})

// 查询回退
_, err = client.ProfitSharingQueryReturn(ctx, "", "R20150806125346", "P20150806125346")

// 查询某笔交易的剩余待分金额
amounts, err := client.ProfitSharingMerchantAmounts(ctx, "", "4208450740201411110007820472")
log.Printf("unsplit_amount = %v", amounts["unsplit_amount"])

// 解冻剩余资金（不再继续分账时调用）
_, err = client.ProfitSharingUnfreeze(ctx, map[string]any{
    "transaction_id": "4208450740201411110007820472",
    "out_order_no":   "U" + utils.RandomString(16),
    "description":    "剩余资金解冻",
})
```

### 使用案例：下载子商户分账账单

```go
billResp, err := client.ProfitSharingBills(ctx, "2026-04-15", "", "GZIP")
if err != nil { return err }
log.Printf("download_url = %v", billResp["download_url"])
// 下一步：用 Inner() 的 HTTP 通道下载账单文件（账单下载接口不走 V3 验签）
```

## 10. 局限

本包现在已经覆盖了服务商模式最常用的接口集合（下单、进件、结算、分账）。剩余仍需要走 `Inner().DoV3` 手动调用的场景，典型是那些**发布节奏较慢、字段定义仍在变化**的管理类接口（例如商家券、代金券发放、电子发票等——它们目前在直连商户侧的封装也不稳定）。

遇到这类接口时：签名 / 验签 / 平台证书拉取 / 敏感字段加密全都可以复用 `Inner()` 已经提供的基础设施——具体到 `DoV3(method, path, query, body, headers, result)` + `PlatformCertForEncrypt` / `EncryptSensitiveField`。

## 11. 并发语义

与 `pay.Client` 完全一致——线程安全，可在多 goroutine 共享。所有核心签名/验签逻辑都继承自 `pay.Client`，本包只是做了业务层的字段注入。
