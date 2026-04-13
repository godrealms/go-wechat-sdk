# go-wechat-sdk 接口文档

本目录按模块拆分了详细的接口说明与使用案例。每份文档包含：

- 模块职责与适用场景
- 所有公开类型、函数、方法的签名和参数说明
- 返回值、错误模型
- 并发与线程安全语义
- 完整的、可直接运行的使用案例

## 目录

| 模块 | 状态 | 文档 |
|---|---|---|
| `utils` — HTTP 基础设施、签名、PEM、随机串 | ✅ | [utils.md](utils.md) |
| `merchant/developed` — 微信支付（商户模式）| ✅ | [merchant-developed.md](merchant-developed.md) |
| `merchant/service` — 微信支付（服务商模式）| ✅ | [merchant-service.md](merchant-service.md) |
| `offiaccount` — 微信公众号 | ✅ | [offiaccount.md](offiaccount.md) |
| `mini-program` — 微信小程序 | ✅ | [mini-program.md](mini-program.md) |
| `mini-game` — 小游戏 | ✅ | [mini-game.md](mini-game.md) |
| `aispeech` — 智能对话（ASR / TTS / NLU）| ✅ | [aispeech.md](aispeech.md) |
| `mini-store` — 微信小店 | ✅ | [mini-store.md](mini-store.md) |
| `xiaowei` — 微信小微 IoT | ✅ | [xiaowei.md](xiaowei.md) |
| `channels` — 视频号 | ✅ | [channels.md](channels.md) |
| `work-wechat/isv` — 企业微信 ISV | ✅ | [work-wechat-isv.md](work-wechat-isv.md) |
| `oplatform` — 开放平台 | ✅ | 第三方平台授权底座；代小程序发布等待实现 |

## 阅读建议

如果你是第一次接入本 SDK，建议顺序：

1. 先读 [utils.md](utils.md)——其他模块的 HTTP 客户端、日志、错误模型、证书加载全都来自这里。
2. 按业务挑一个模块的文档通读：
   - 做支付接入 → [merchant-developed.md](merchant-developed.md)
   - 做服务商支付 → 先看 merchant-developed 再看 [merchant-service.md](merchant-service.md)
   - 做公众号 → [offiaccount.md](offiaccount.md)
   - 做小程序后端 → [mini-program.md](mini-program.md)
   - 做小游戏后端 → [mini-game.md](mini-game.md)
   - 接入 AI 语音（ASR / TTS / NLU）→ [aispeech.md](aispeech.md)
   - 做微信小店（商品 / 订单 / 优惠券）→ [mini-store.md](mini-store.md)
   - 做微信小微 IoT → [xiaowei.md](xiaowei.md)
   - 做视频号（数据 / 直播 / 订单）→ [channels.md](channels.md)
   - 做企业微信 ISV 服务商 → [work-wechat-isv.md](work-wechat-isv.md)
3. 每份模块文档末尾都有一个"完整使用案例"段落，可以直接抄代码起步。

## 约定

- 所有公开 API 都接收 `context.Context`，请务必把上游 ctx 透传进来，SDK 内部不会创建 `context.Background()` 以外的上下文。
- 所有可能失败的操作都返回 `error`。特殊错误类型会在对应模块文档中说明（例如 `utils.HTTPError`、`offiaccount.WeixinError`）。
- SDK 不会把敏感信息打到 stdout。默认使用 `nopLogger`，需要日志请用 `utils.WithLogger(...)` 注入自己的实现。
- 所有 `Client` 都是并发安全的，可以多 goroutine 共享。
