# oplatform 代小程序运营 (WxaOps) — 设计文档

**Date:** 2026-04-12
**Status:** Draft — awaiting user review
**Scope:** 微信开放平台子项目 7 — 代小程序运营管理（插件、订阅消息、客服消息）

## 1. 背景

本文档是 oplatform 的第四轮实现迭代。已交付：
- 子项目 1+2+6（授权底座 / 代调用框架 / 扫码登录）
- 子项目 3（代小程序开发管理 WxaAdmin — account/category/domain/tester/code/release）
- 子项目 5（快速注册 FastRegister）

本轮交付子项目 7 —— 代运营阶段用于日常业务的 16 个接口，覆盖插件、订阅消息、客服消息三个家族。

### 1.1 目标

在现有 `WxaAdminClient` 上新增 **16 个方法**：

- **插件管理**（7）：使用方申请/查询/解绑 + 插件方审核/同意/拒绝/删除
- **订阅消息 & 模板库**（7）：类目、模板库、私有模板 CRUD、发送
- **客服消息**（2）：发送客服消息、输入状态

### 1.2 显式不做

- 小程序广告 API —— 官方已废弃或迁移到广告平台
- 附近小程序管理 —— 非主流场景
- 多客服 session 管理（`customservice/kfsession/*`）—— 复杂度高，独立特性
- 代运营事件回调解密（订阅消息发送结果等）—— 后续通过扩展 `ParseNotify` 实现

## 2. 架构

### 2.1 继续沿用 WxaAdminClient

所有 16 个方法都需要 `authorizer_access_token`，因此挂在现有 `WxaAdminClient` 上。子项目 3 已经建立了这个 wrapper + `doPost` / `doGet` / `doGetRaw` 共享助手，本轮完全复用，不引入新类型也不修改基础设施。

### 2.2 文件布局

```
oplatform/
  wxa.plugin.go           [NEW]  7 methods — 插件申请/审核
  wxa.plugin_test.go      [NEW]
  wxa.submsg.go           [NEW]  7 methods — 订阅消息 & 模板库
  wxa.submsg_test.go      [NEW]
  wxa.customer.go         [NEW]  2 methods — 客服消息
  wxa.customer_test.go    [NEW]
  wxa.struct.go           [MOD]  追加 plugin / submsg / customer DTO
```

一个子家族一对 `.go` + `_test.go`，和子项目 3 的节奏保持一致。

### 2.3 约定

- 所有方法都是 `func (w *WxaAdminClient) Xxx(ctx context.Context, ...) ...`
- HTTP 调用全部走 `w.doPost` / `w.doGet`（子项目 3 已定义）
- DTO 追加到 `wxa.struct.go` 对应 section
- 测试助手仍是 `newTestWxaAdmin(t, baseURL)`，不需要新 helper

## 3. 接口清单

### 3.1 wxa.plugin.go — 插件管理（7 个方法）

**使用方小程序**（3 个）

| 方法 | Endpoint | Body (action 在 body 而不是 query) |
|---|---|---|
| `ApplyPlugin(ctx, pluginAppID string) error` | `POST /wxa/plugin` | `{"action":"apply","plugin_appid":"..."}` |
| `ListPlugins(ctx) (*WxaPluginList, error)` | `POST /wxa/plugin` | `{"action":"list"}` |
| `UnbindPlugin(ctx, pluginAppID string) error` | `POST /wxa/plugin` | `{"action":"unbind","plugin_appid":"..."}` |

**插件方小程序**（4 个）

| 方法 | Endpoint | Body |
|---|---|---|
| `GetPluginDevApplyList(ctx, page, num int) (*WxaPluginDevApplyList, error)` | `POST /wxa/devplugin` | `{"action":"dev_apply_list","page":0,"num":10}` |
| `AgreeDevPlugin(ctx, userAppID string) error` | `POST /wxa/devplugin` | `{"action":"dev_agree","appid":"..."}` |
| `RefuseDevPlugin(ctx, reason string) error` | `POST /wxa/devplugin` | `{"action":"dev_refuse","reason":"..."}` |
| `DeleteDevPlugin(ctx, userAppID string) error` | `POST /wxa/devplugin` | `{"action":"dev_delete","appid":"..."}` |

**注意点**：这里的 `action` 字段在 **body** 而不是 URL query，和 FastRegister 不同。`doPost` helper 无需修改，调用方自行构造 body map。

### 3.2 wxa.submsg.go — 订阅消息 & 模板库（7 个方法）

| 方法 | Endpoint | Method |
|---|---|---|
| `GetSubscribeCategory(ctx) (*WxaSubscribeCategoryResp, error)` | `/wxaapi/newtmpl/getcategory` | GET |
| `GetPubTemplateTitles(ctx, ids string, start, limit int) (*WxaPubTemplateTitles, error)` | `/wxaapi/newtmpl/getpubtemplatetitles` | GET |
| `GetPubTemplateKeywords(ctx, tid string) (*WxaPubTemplateKeywords, error)` | `/wxaapi/newtmpl/getpubtemplatekeywords` | GET |
| `AddSubscribeTemplate(ctx, req *WxaAddSubscribeTemplateReq) (*WxaAddSubscribeTemplateResp, error)` | `/wxaapi/newtmpl/addtemplate` | POST |
| `DeleteSubscribeTemplate(ctx, priTmplID string) error` | `/wxaapi/newtmpl/deltemplate` | POST |
| `ListSubscribeTemplates(ctx) (*WxaSubscribeTemplateList, error)` | `/wxaapi/newtmpl/gettemplate` | GET |
| `SendSubscribeMessage(ctx, req *WxaSendSubscribeReq) error` | `/cgi-bin/message/subscribe/send` | POST |

**关于 `GetPubTemplateTitles` 的 `ids` 参数**：官方格式是用 `-` 连接的类目 ID 列表（例如 `"2-3-5"`）。我们保留原始字符串参数，不在 SDK 里做解析或拼接。

**关于 `SendSubscribeMessage` 的重复性**：`mini-program` 包已经有同名方法。通过 `auth.MiniProgramClient().SendSubscribeMessage(ctx, body)` 也能完成发送。我们仍在 WxaAdmin 上提供一个入口，因为：
1. 代运营场景里用户更可能从 WxaAdmin 入口调用
2. 两条路径共享同一个 token 源，行为一致
3. 类型命名空间独立，无冲突
4. 用户无需为一次发送再去构造 mini-program client

### 3.3 wxa.customer.go — 客服消息（2 个方法）

| 方法 | Endpoint |
|---|---|
| `SendCustomerMessage(ctx, req *WxaSendCustomerMessageReq) error` | `POST /cgi-bin/message/custom/send` |
| `SendCustomerTyping(ctx, toUser, command string) error` | `POST /cgi-bin/message/custom/typing` |

`command` 合法值：`"Typing"` 或 `"CancelTyping"`。

`WxaSendCustomerMessageReq` 支持 4 种 `MsgType`：`text` / `image` / `link` / `miniprogrampage`，对应的字段 `Text` / `Image` / `Link` / `MiniProgramPage` 都是 `omitempty` 可选指针。调用方按场景填一个，SDK 不做字段互斥校验（保持最薄封装）。

## 4. DTO 设计

追加到 `wxa.struct.go`：

```go

// ----- plugin -----

type WxaPluginListItem struct {
	AppID      string `json:"appid"`
	Status     int    `json:"status"` // 1=申请中 2=申请通过 3=已拒绝 4=已超时
	Nickname   string `json:"nickname,omitempty"`
	HeadImgURL string `json:"headimgurl,omitempty"`
}

type WxaPluginList struct {
	PluginList []WxaPluginListItem `json:"plugin_list"`
}

type WxaPluginDevApplyItem struct {
	AppID      string `json:"appid"`
	Status     int    `json:"status"`
	Nickname   string `json:"nickname,omitempty"`
	HeadImgURL string `json:"headimgurl,omitempty"`
	Categories []struct {
		First  string `json:"first"`
		Second string `json:"second"`
	} `json:"categories,omitempty"`
	CreateTime string `json:"create_time,omitempty"`
	ApplyURL   string `json:"apply_url,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type WxaPluginDevApplyList struct {
	ApplyList []WxaPluginDevApplyItem `json:"apply_list"`
}

// ----- subscribe message -----

type WxaSubscribeCategoryItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type WxaSubscribeCategoryResp struct {
	Data []WxaSubscribeCategoryItem `json:"data"`
}

type WxaPubTemplateTitleItem struct {
	TID        int    `json:"tid"`
	Title      string `json:"title"`
	Type       int    `json:"type"` // 2=一次性 3=长期
	CategoryID string `json:"categoryId"`
}

type WxaPubTemplateTitles struct {
	Count int                       `json:"count"`
	Data  []WxaPubTemplateTitleItem `json:"data"`
}

type WxaPubTemplateKeywordItem struct {
	KID     int    `json:"kid"`
	Name    string `json:"name"`
	Example string `json:"example"`
	Rule    string `json:"rule"`
}

type WxaPubTemplateKeywords struct {
	Count int                         `json:"count"`
	Data  []WxaPubTemplateKeywordItem `json:"data"`
}

type WxaAddSubscribeTemplateReq struct {
	TID       string `json:"tid"`
	KidList   []int  `json:"kidList"`
	SceneDesc string `json:"sceneDesc,omitempty"`
}

type WxaAddSubscribeTemplateResp struct {
	PriTmplID string `json:"priTmplId"`
}

type WxaSubscribeTemplateItem struct {
	PriTmplID string `json:"priTmplId"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Example   string `json:"example"`
	Type      int    `json:"type"`
}

type WxaSubscribeTemplateList struct {
	Data []WxaSubscribeTemplateItem `json:"data"`
}

type WxaSubscribeTemplateDataField struct {
	Value string `json:"value"`
}

type WxaSendSubscribeReq struct {
	ToUser           string                                   `json:"touser"`
	TemplateID       string                                   `json:"template_id"`
	Page             string                                   `json:"page,omitempty"`
	MiniprogramState string                                   `json:"miniprogram_state,omitempty"` // developer/trial/formal
	Lang             string                                   `json:"lang,omitempty"`
	Data             map[string]WxaSubscribeTemplateDataField `json:"data"`
}

// ----- customer message -----

type WxaCustomerTextPayload struct {
	Content string `json:"content"`
}

type WxaCustomerImagePayload struct {
	MediaID string `json:"media_id"`
}

type WxaCustomerLinkPayload struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	ThumbURL    string `json:"thumb_url"`
}

type WxaCustomerMiniProgramPagePayload struct {
	Title        string `json:"title"`
	Pagepath     string `json:"pagepath"`
	ThumbMediaID string `json:"thumb_media_id"`
}

type WxaSendCustomerMessageReq struct {
	ToUser          string                             `json:"touser"`
	MsgType         string                             `json:"msgtype"` // text/image/link/miniprogrampage
	Text            *WxaCustomerTextPayload            `json:"text,omitempty"`
	Image           *WxaCustomerImagePayload           `json:"image,omitempty"`
	Link            *WxaCustomerLinkPayload            `json:"link,omitempty"`
	MiniProgramPage *WxaCustomerMiniProgramPagePayload `json:"miniprogrampage,omitempty"`
}
```

## 5. 错误处理

- 业务错误 (`errcode != 0`) → `*oplatform.WeixinError`（沿用既有约定）
- HTTP/网络错误透传
- 不做特殊 errcode 分支；调用方用 `errors.As(*WeixinError)` 判断

## 6. 并发与生命周期

- `WxaAdminClient` 无状态，新增方法全部无副作用
- 所有方法接受 `ctx context.Context`，可取消/超时
- 可在多 goroutine 中共享

## 7. 测试策略

沿用 httptest 驱动模式 + 既有的 `newTestWxaAdmin(t, baseURL)` 助手。

### 测试矩阵

| 测试文件 | 覆盖 | 用例数 |
|---|---|---|
| `wxa.plugin_test.go` | 7 方法 happy path + 1 errcode (ApplyPlugin 重复申请) | 8 |
| `wxa.submsg_test.go` | 7 方法 happy path，AddSubscribeTemplate 验证 body/response 解析 | 7 |
| `wxa.customer_test.go` | SendCustomerMessage 3 种 msgtype (text/image/miniprogrampage) + SendCustomerTyping | 4 |

**合计 19 个新用例**。全部不需要额外 helper。

## 8. 兼容性

纯 additive：
- `WxaAdminClient` 获得 16 个新方法
- 新增约 16 个 struct DTO 类型
- 零 breaking change

## 9. 交付物

| 文件 | 行数估计 |
|---|---|
| `oplatform/wxa.plugin.go` | ~130 |
| `oplatform/wxa.plugin_test.go` | ~240 |
| `oplatform/wxa.submsg.go` | ~120 |
| `oplatform/wxa.submsg_test.go` | ~220 |
| `oplatform/wxa.customer.go` | ~60 |
| `oplatform/wxa.customer_test.go` | ~130 |
| `oplatform/wxa.struct.go` 追加 | ~160 |

**生产代码 ~470 行，测试 ~590 行，合计 ~1060 行**。预计 3 个原子 commit：

1. `wxa.plugin.*` + DTO 追加
2. `wxa.submsg.*` + DTO 追加
3. `wxa.customer.*` + DTO 追加

## 10. 非目标 / 后续工作

- 小程序广告 API（已废弃或已迁移）
- 附近小程序管理
- 多客服 session 管理（`customservice/kfsession/*`）
- 代运营事件回调解密
- 小程序码/二维码生成（`wxacode/*` family，自有 appid 也能调，归 mini-program）

## 11. 开放问题

无未决项。所有关键决策已锁定：

- 架构：复用 `WxaAdminClient`，无新类型 ✔
- 覆盖：16 方法（7 plugin + 7 submsg + 2 customer）✔
- 文件：3 对子家族文件 + `wxa.struct.go` 追加 ✔
- Action 字段位置：plugin 家族在 body（非 query）✔
- SendSubscribeMessage 重复：有意为之，两条路径共存 ✔
- 测试：沿用 newTestWxaAdmin helper ✔
- 兼容：纯 additive ✔
