# Go WeChat SDK

一个功能完整的 Go 语言微信开发工具包，支持微信生态的主要服务接口。

[![Go Reference](https://pkg.go.dev/badge/github.com/godrealms/go-wechat-sdk.svg)](https://pkg.go.dev/github.com/godrealms/go-wechat-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/godrealms/go-wechat-sdk)](https://goreportcard.com/report/github.com/godrealms/go-wechat-sdk)
[![License](https://img.shields.io/github/license/godrealms/go-wechat-sdk)](LICENSE)

## 安装 / Install

```bash
go get github.com/godrealms/go-wechat-sdk
```

## 快速入门 / Quick Start

### 公众号

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/godrealms/go-wechat-sdk/offiaccount"
)

func main() {
    ctx := context.Background()
    c, err := offiaccount.NewClient(ctx, &offiaccount.Config{
        AppID:     "wx_your_appid",
        AppSecret: "your_app_secret",
    })
    if err != nil {
        log.Fatal(err)
    }
    ip, err := c.GetCallbackIp(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(ip)
}
```

### 微信支付

```go
package main

import (
    "context"
    "fmt"
    "log"

    pay "github.com/godrealms/go-wechat-sdk/merchant/developed"
    "github.com/godrealms/go-wechat-sdk/merchant/developed/types"
)

func main() {
    ctx := context.Background()
    c, err := pay.NewClient(pay.Config{
        AppID:             "wx_your_appid",
        MchID:             "your_mchid",
        APIv3Key:          "your_apiv3key",
        SerialNo:          "your_serial_no",
        PrivateKeyPEMPath: "/path/to/apiclient_key.pem",
    })
    if err != nil {
        log.Fatal(err)
    }
    resp, err := c.TransactionsJsapi(ctx, &types.Transactions{
        Description: "商品描述",
        OutTradeNo:  "order_20260413_001",
        Amount:      &types.Amount{Total: 100, Currency: "CNY"},
        Payer:       &types.Payer{OpenID: "oUser_xxxx"},
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(resp.PrepayID)
}
```

## 支持的模块 / Supported Modules

| 包 | 描述 | 状态 | 文档 |
|---|---|---|---|
| `utils` | HTTP 客户端、签名工具、PEM 加载、随机串 | ✅ | [docs/utils.md](docs/utils.md) |
| `offiaccount` | 微信公众号服务端 API | ✅ | [docs/offiaccount.md](docs/offiaccount.md) |
| `merchant/developed` | 微信支付（直连商户模式）| ✅ | [docs/merchant-developed.md](docs/merchant-developed.md) |
| `merchant/service` | 微信支付（服务商模式）| ✅ | [docs/merchant-service.md](docs/merchant-service.md) |
| `mini-program` | 微信小程序服务端 API | ✅ | [docs/mini-program.md](docs/mini-program.md) |
| `mini-game` | 微信小游戏服务端 API | ✅ | [docs/mini-game.md](docs/mini-game.md) |
| `aispeech` | 微信 AI 语音（ASR / TTS / NLU / 对话）| ✅ | [docs/aispeech.md](docs/aispeech.md) |
| `mini-store` | 微信小店（商品 / 订单 / 优惠券 / 结算）| ✅ | [docs/mini-store.md](docs/mini-store.md) |
| `xiaowei` | 微信小微 IoT 平台 | ✅ | [docs/xiaowei.md](docs/xiaowei.md) |
| `channels` | 视频号（数据 / 直播 / 订单 / 商品）| ✅ | [docs/channels.md](docs/channels.md) |
| `work-wechat/isv` | 企业微信 ISV 服务商模式 | ✅ | [docs/work-wechat-isv.md](docs/work-wechat-isv.md) |
| `oplatform` | 微信开放平台（第三方平台授权）| ✅ | — |

## 约定 / Conventions

- **Context：** 所有公开 API 第一参数均为 `ctx context.Context`，请透传上游 ctx，SDK 内部不创建额外上下文。
- **错误：** 所有可能失败的操作返回 `error`。WeChat API 业务错误通过 `errors.As` 提取（见各包文档）。
- **并发：** 所有 `Client` 并发安全，可多 goroutine 共享。
- **日志：** 默认 `nopLogger`（静默），通过 `utils.WithLogger(l)` 注入实现了 `utils.Logger` 接口的日志器。
- **Token：** SDK 自动管理 access_token 缓存（双检锁 + 60 秒提前刷新），无需手动调用刷新。
