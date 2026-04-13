# Documentation Update Design

**Date:** 2026-04-13  
**Scope:** Full documentation refresh for go-wechat-sdk  
**Target Audience:** External SDK users (developers integrating WeChat APIs)  
**Language Convention:** Chinese body text + English code/signatures (bilingual)

---

## Background

Following the completion of four SDK optimization sub-projects (code quality, test coverage, Godoc, stub package implementation), the codebase has changed significantly:

- Six previously stub-only packages are now fully implemented: `aispeech`, `mini-game`, `mini-store`, `xiaowei`, `channels`, `work-wechat/isv`
- All API methods now take `ctx context.Context` as the first argument
- `offiaccount` replaced `GetAccessToken()` with `AccessTokenE(ctx)` (the old method is deprecated)
- `utils` gained `DoRequest`, `DoRequestWithRawResponse`, `WithLogger`, `WithHTTPClient`, `HTTPError`
- Error types standardized: `*APIError` in mini-game/channels/mini-program, `*WeixinError` in offiaccount

The existing documentation does not reflect any of these changes, and six packages have no docs at all.

---

## Approach: Layered Documentation (B)

`README.md` stays lightweight (overview + install + quick-start + module index).  
`docs/` contains one detailed reference document per package.

This matches the existing `docs/offiaccount.md` / `docs/utils.md` pattern and scales cleanly as new packages are added.

---

## File Inventory

### Files to Update (5)

| File | Changes Required |
|------|-----------------|
| `README.md` | Remove verbose per-module descriptions; update module table to show all packages ✅; keep quick-start code snippets |
| `docs/README.md` | Update index table: all 12 packages ✅ with correct doc links |
| `docs/utils.md` | Add `DoRequest`, `DoRequestWithRawResponse`, `WithLogger`, `WithHTTPClient`, `HTTPError`; update `NewHTTP` option list |
| `docs/offiaccount.md` | Add `ctx context.Context` to all method signatures; mark `GetAccessToken()` deprecated; document `AccessTokenE(ctx)`; update `WeixinError` section |
| `docs/merchant-developed.md` | Add `ctx context.Context` to all transaction method signatures |
| `docs/mini-program.md` | Update `Code2Session`, `AccessToken`, `SendSubscribeMessage` error type to `*APIError` |

### Files to Create (6)

| File | Package | Methods |
|------|---------|---------|
| `docs/aispeech.md` | `aispeech` | 7: ASRLong, ASRShort, TextToSpeech, NLUUnderstand, NLUIntentRecognize, DialogQuery, DialogReset |
| `docs/mini-game.md` | `mini-game` | 12: Code2Session, GetDailySummary, GetDailyRetain, CreateGameRoom, GetRoomInfo, UnifyInvokePayment, ModifyInvokePayment, MsgSecCheck, SetUserStorage, RemoveUserStorage, GetUserStorage (+ AccessToken) |
| `docs/mini-store.md` | `mini-store` | 24: AddProduct, DelProduct, AuditProduct, CancelAuditProduct, UpdateProduct, GetProduct, GetProductList, ListCategories + 8 order methods + 4 coupon methods + 4 settlement methods |
| `docs/xiaowei.md` | `xiaowei` | 12: GetStoreInfo, UpdateStoreInfo, GetKYCStatus, SubmitKYC, AddMicroProduct, DelMicroProduct, GetMicroProduct, ListMicroProducts, GetMicroOrder, ListMicroOrders, ShipMicroOrder, RefundMicroOrder |
| `docs/channels.md` | `channels` | ~15: GetDailySummary, GetDailyRetain + live room methods + order methods + product CRUD |
| `docs/work-wechat-isv.md` | `work-wechat/isv` | 40+: NewClient, ParseNotify, ParseDataNotify, corp agent/approval/calendar/checkin/external-contact/jssdk/media/menu APIs |

---

## Per-Document Template

Every document (new and updated) follows this fixed structure:

```
# Package Name（包名）

> 一句话描述包的职责

## 适用场景

...（何时用这个包）

## 初始化 / Initialization

```go
// Go signature (English)
```
中文参数说明

## 错误处理 / Error Handling

错误类型说明，errors.As 用法示例

## API 参考 / API Reference

### MethodName

```go
func (c *Client) MethodName(ctx context.Context, ...) (..., error)
```
- **参数：** ...
- **返回：** ...
- **说明：** 中文一句话

（每个方法一节，按功能域分组）

## 完整示例 / Complete Example

```go
// 一个可直接运行的示例，涵盖该包的主要功能
```
```

---

## Content Guidelines

- **Code blocks and signatures:** English only (match Godoc)
- **Section headers:** English (machine-readable, SEO-friendly)
- **Body text and parameter descriptions:** Chinese
- **No translation of WeChat API field names** — keep them as-is (e.g., `appid`, `mchid`, `openid`)
- **Every method must show:** Go signature, parameter table (Chinese descriptions), return type, error conditions
- **Quick-start examples:** must compile against the current API (ctx-first signatures, AccessTokenE not GetAccessToken)

---

## README.md Target Structure

```markdown
# Go WeChat SDK
badges

> 一句话简介（中文）

## 安装 / Install
go get ...

## 快速入门 / Quick Start
（offiaccount 2段示例 + pay 1段示例）

## 支持的模块 / Supported Modules
| 包 | 描述 | 状态 | 文档 |
|---|---|---|---|
| utils | HTTP客户端、签名、证书 | ✅ | docs/utils.md |
| offiaccount | 公众号 | ✅ | docs/offiaccount.md |
| merchant/developed | 微信支付（商户模式）| ✅ | docs/merchant-developed.md |
| merchant/service | 微信支付（服务商模式）| ✅ | docs/merchant-service.md |
| mini-program | 小程序 | ✅ | docs/mini-program.md |
| mini-game | 小游戏 | ✅ | docs/mini-game.md |
| aispeech | 智能对话 | ✅ | docs/aispeech.md |
| mini-store | 微信小店 | ✅ | docs/mini-store.md |
| xiaowei | 微信小微 IoT | ✅ | docs/xiaowei.md |
| channels | 视频号 | ✅ | docs/channels.md |
| work-wechat/isv | 企业微信 ISV | ✅ | docs/work-wechat-isv.md |
| oplatform | 开放平台 | ✅ | — |

## 约定 / Conventions
（ctx透传、error类型、并发安全、日志注入）
```

---

## Implementation Plan Outline

Tasks can be executed in parallel by package:

1. **README.md + docs/README.md** — index files (no package knowledge needed)
2. **Update docs/utils.md** — add new HTTP methods
3. **Update docs/offiaccount.md** — ctx migration + WeixinError
4. **Update docs/merchant-developed.md + docs/mini-program.md** — ctx / APIError
5. **New: docs/aispeech.md + docs/mini-game.md**
6. **New: docs/mini-store.md + docs/xiaowei.md**
7. **New: docs/channels.md + docs/work-wechat-isv.md**

Total: 11 files changed/created.

---

## Success Criteria

- [ ] All 12 package entries in `docs/README.md` show ✅ with working links
- [ ] `README.md` module table is accurate and all links resolve
- [ ] Every public method in every package appears in its doc file with Go signature
- [ ] All code examples use `ctx context.Context` as first argument
- [ ] No `GetAccessToken()` in any example (replaced by `AccessTokenE(ctx)`)
- [ ] Six new doc files created for previously undocumented packages
- [ ] Chinese body text, English code/signatures throughout
