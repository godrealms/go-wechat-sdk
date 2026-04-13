# Go WeChat SDK

一个功能完整的 Go 语言微信开发工具包，支持微信生态的主要服务接口。

[![Go Reference](https://pkg.go.dev/badge/github.com/godrealms/go-wechat-sdk.svg)](https://pkg.go.dev/github.com/godrealms/go-wechat-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/godrealms/go-wechat-sdk)](https://goreportcard.com/report/github.com/godrealms/go-wechat-sdk)
[![License](https://img.shields.io/github/license/godrealms/go-wechat-sdk)](LICENSE)

## 功能特性

- 支持微信生态全系列服务
- 简洁易用的 API 设计
- 完善的错误处理机制
- 详细的使用文档
- 丰富的示例代码
- 自动 Access Token 管理

## 支持的服务

> ✅ = 已实现并有测试覆盖；🚧 = 目录已存在但仅有占位/未实现

- 💰 微信支付 (`merchant/developed`) ✅：JSAPI / APP / H5 / Native 下单，订单查询、关单、退款、对账单、分账、合单、代金券；含响应签名校验、平台证书自动管理、**支付回调通知解析**
- 🤝 微信支付服务商 (`merchant/service`) ✅：服务商模式薄封装，复用商户侧全部签名/验签逻辑
- 📢 公众号 (`offiaccount`) ✅：菜单、客服/模板/群发消息、素材、用户/标签管理、JS-SDK、网页授权；**完整的消息加解密 (Biz Msg Crypt)** 和回调解析
- 📱 小程序 (`mini-program`) ✅：登录 (`Code2Session`)、access_token 缓存、订阅消息、`DecryptUserData`（解密 encryptedData）
- 🎮 小游戏 (`mini-game`) 🚧：目录占位，尚未实现
- 🤖 智能对话 (`aispeech`) 🚧：目录占位，尚未实现
- 🌐 开放平台 (`oplatform`) ✅：第三方平台授权底座（verify_ticket 回调、component_access_token、pre_auth_code、authorization_code 换 authorizer token、授权事件解析）+ authorizer 代调用框架（通过 TokenSource 注入到 offiaccount / mini-program，token 自动来自开放平台）+ 网站应用扫码登录 (snsapi_login)。代小程序开发/运营管理、快速注册等子模块后续实现
- 💼 企业微信 (`work-wechat`) 🚧：目录占位，尚未实现
- 📺 视频号 (`channels`) 🚧：目录占位，尚未实现
- 🏪 微信小店 (`mini-store`) 🚧：目录占位，尚未实现
- 🤖 微信小微 (`xiaowei`) 🚧：目录占位，尚未实现

## 快速开始

### 安装

```bash
go get github.com/godrealms/go-wechat-sdk
```

### 基础配置

```go
package main

import (
    "context"
    "log"

    "github.com/godrealms/go-wechat-sdk/offiaccount"
    pay "github.com/godrealms/go-wechat-sdk/merchant/developed"
    "github.com/godrealms/go-wechat-sdk/utils"
)

func main() {
    ctx := context.Background()

    // 公众号
    official := offiaccount.NewClient(ctx, &offiaccount.Config{
        AppId:     "your-app-id",
        AppSecret: "your-app-secret",
    })
    token, err := official.AccessTokenE(ctx) // 推荐：错误会被返回
    if err != nil {
        log.Fatal(err)
    }
    _ = token

    // 微信支付：推荐使用 NewClient + Config
    privateKey, _ := utils.LoadPrivateKeyWithPath("apiclient_key.pem")
    cert, _ := utils.LoadCertificateWithPath("apiclient_cert.pem")
    payClient, err := pay.NewClient(pay.Config{
        Appid:             "your-app-id",
        Mchid:             "your-merchant-id",
        CertificateNumber: "your-cert-serial",
        APIv3Key:          "your-api-v3-key-32-bytes",
        PrivateKey:        privateKey,
        Certificate:       cert,
    })
    if err != nil {
        log.Fatal(err)
    }
    // 可选：启动时主动拉取一次平台证书，后续响应自动验签
    _, _ = payClient.FetchPlatformCertificates(ctx)
}
```

## 设计与运行时注意事项

- **并发安全**：`merchant/developed.Client` 与 `offiaccount.Client` 都可在多 goroutine 中共享。所有签名/认证字段都按请求生成，不会修改共享 header。
- **响应签名校验**：`merchant/developed` 默认会用平台证书校验微信支付返回的 `Wechatpay-Signature`；首次调用未注入证书时会自动拉取 `/v3/certificates`。
- **错误处理**：公众号 API 错误请使用 `errors.As` 解构 `*offiaccount.WeixinError`；旧的 `GetAccessToken() string` 仍可用，但建议改为 `AccessTokenE(ctx)` 以拿到错误。
- **日志**：utils.HTTP 不再向 stdout 打印请求日志，可通过 `utils.WithLogger(...)` 注入自己的实现。
- **超时**：默认超时 30s，可通过 `utils.WithTimeout` / `utils.WithHTTPClient` 自定义。

## 常用片段

### 解析微信支付回调通知

```go
http.HandleFunc("/wxpay/notify", func(w http.ResponseWriter, r *http.Request) {
    var txn struct {
        TransactionId string `json:"transaction_id"`
        OutTradeNo    string `json:"out_trade_no"`
        TradeState    string `json:"trade_state"`
    }
    notify, err := payClient.ParseNotification(r.Context(), r, &txn)
    if err != nil {
        pay.FailNotification(w, err.Error())
        return
    }
    log.Printf("event=%s txn=%+v", notify.EventType, txn)
    pay.AckNotification(w)
})
```

### 公众号回调消息解密

```go
crypto, _ := offiaccount.NewMsgCrypto(token, encodingAESKey, appId)

http.HandleFunc("/wx/callback", func(w http.ResponseWriter, r *http.Request) {
    plaintext, err := offiaccount.ParseNotify(r, crypto)
    if err != nil { http.Error(w, err.Error(), 400); return }
    if r.Method == http.MethodGet { // 接入校验
        w.Write(plaintext) // echostr
        return
    }
    var msg struct { MsgType, Content string }
    xml.Unmarshal(plaintext, &msg)
    // ... 构造回复 ...
    reply, _ := crypto.BuildEncryptedReply(replyXML,
        r.URL.Query().Get("timestamp"), r.URL.Query().Get("nonce"))
    w.Write(reply)
})
```

### 小程序登录

```go
mp, _ := mini_program.NewClient(mini_program.Config{AppId: "wx", AppSecret: "sec"})
sess, _ := mp.Code2Session(ctx, jsCode)
// 解密用户信息
var user struct{ OpenId, NickName string }
mini_program.DecryptUserData(sess.SessionKey, encryptedData, iv, &user)
```

## 测试

```bash
go build ./...
go test ./...
```

当前有测试覆盖的包：`utils`, `merchant/developed`, `offiaccount`, `mini-program`。

## 模块说明

### 微信支付 (merchant)

微信支付模块支持完整的支付功能，包括：
- JSAPI支付
- APP支付
- H5支付
- Native支付
- 小程序支付
- 退款处理
- 订单查询
- 账单下载

支付模块采用链式调用设计，配置更加灵活。

### 公众号 (offiaccount)

公众号模块提供完整的公众号API支持，包括：

#### 核心功能
- 自定义菜单管理
- 消息管理（客服消息、模板消息、群发消息）
- 素材管理（临时素材、永久素材、草稿箱）
- 用户管理（用户信息、标签管理、黑名单）
- 账号管理（二维码、短Key URL、数据统计）

#### 高级功能
- 发票功能（电子发票、财政电子票据）
- 非税支付
- 一码通
- 医疗助手
- 留言功能
- 数据分析
- 微信门店
- 微信AI开放平台（OCR识别、语音识别）
- JS-SDK支持
- 网页授权
- 评论管理

#### 特色功能
- 自动 Access Token 管理
- 完善的错误处理机制
- 详细的接口文档和示例代码

### 其他模块

其他模块如小程序、小游戏、企业微信等目前提供基础框架，可根据需要进行扩展。

## 示例

### 小程序登录

```go
import "github.com/godrealms/go-wechat-sdk/mini-program"

func login(code string) (*miniprogram.LoginResponse, error) {
    client := miniprogram.NewClient(&miniprogram.Config{
        AppID:     "your-app-id",
        AppSecret: "your-app-secret",
    })
    
    return client.Login(code)
}
```

### 发送模板消息

```go
import "github.com/godrealms/go-wechat-sdk/offiaccount"

func sendTemplate(openID, templateID string, data interface{}) error {
    client := offiaccount.NewClient(context.Background(), config)
    
    msg := &offiaccount.TemplateMessage{
        ToUser:     openID,
        TemplateID: templateID,
        Data:       data,
    }
    
    return client.SendTemplate(msg)
}
```

## 文档

详细文档请访问我们的 [Wiki](https://github.com/godrealms/go-wechat-sdk/wiki)

每个模块都有详细的 README 文档：
- [公众号模块文档](offiaccount/README.md)

## 贡献指南

欢迎提交 Pull Request 或创建 Issue。

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 联系我们

- 项目地址：[https://github.com/godrealms/go-wechat-sdk](https://github.com/godrealms/go-wechat-sdk)
- 问题反馈：[Issues](https://github.com/godrealms/go-wechat-sdk/issues)

## 致谢

感谢所有贡献者对本项目的支持！