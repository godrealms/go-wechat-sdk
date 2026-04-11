# oplatform 快速注册 (FastRegister) — 设计文档

**Date:** 2026-04-12
**Status:** Draft — awaiting user review
**Scope:** 微信开放平台子项目 5 — 代注册小程序

## 1. 背景

本文档是 oplatform 的第三轮实现迭代。已交付：
- 子项目 1+2+6（授权底座 / 代调用框架 / 扫码登录）— spec `2026-04-12-oplatform-auth-foundation-design.md`
- 子项目 3（代小程序开发管理 WxaAdmin）— spec `2026-04-12-oplatform-wxa-admin-design.md`

本轮交付子项目 5 —— 通过第三方平台代商户完成小程序的快速注册与基础账号信息管理。

### 1.1 目标

覆盖官方"代注册小程序"家族的核心接口，共 **8 个方法**：

- **企业快速注册**（2）：创建任务 + 查询任务状态
- **个人类型小程序**（2）：创建任务 + 查询任务状态
- **复用主体试用版**（2）：创建任务 + 查询任务状态
- **管理员变更**（1）：生成管理员变更二维码
- **账号基本信息**（1）：获取小程序基本信息（authorizer 级别）

### 1.2 显式不做

- 账号注销 family（`apply_logoff` / `check_logoff_cond` / `logoff_bind_wechat`）—— 留到下一轮作为独立子项目 5.5
- 公众号快速创建（`api_create_miniprogram` 等）—— 微信该 API 已与小程序路径合并，不单独实现
- 管理员变更任务查询 / 完成（`createbindlink` 等辅助接口）—— 上述"生成二维码" 本身是主入口，后续管理可通过管理员端完成，暂不封装
- 复用公众号/小程序资质创建新账号的其它罕见路径

## 2. 架构

### 2.1 FastRegisterClient 包装器

快速注册的绝大多数接口需要 `component_access_token` 而不是 `authorizer_access_token`（因为此时 authorizer 关系还不存在），因此必须挂在 `*Client` 而非 `*AuthorizerClient`/`*WxaAdminClient` 上。

```go
// oplatform/fastregister.go
package oplatform

// FastRegisterClient 提供开放平台代注册小程序相关的 component 级接口。
// 所有方法都以 Client.ComponentAccessToken() 作为 token 源。
// 无状态，线程安全，可在多 goroutine 共享。
type FastRegisterClient struct {
	c *Client
}

// FastRegister 从 Client 构造 FastRegisterClient。构造无 I/O。
func (c *Client) FastRegister() *FastRegisterClient {
	return &FastRegisterClient{c: c}
}
```

### 2.2 共享 doPost 助手

和 `WxaAdminClient.doPost` 结构一致，但 token 源不同：从 `Client.ComponentAccessToken(ctx)` 取，而非 `AuthorizerClient.AccessToken(ctx)`。

```go
// fastregister.go 内部
func (f *FastRegisterClient) doPost(ctx context.Context, path string, body, out any) error {
	ctx = touchContext(ctx)
	token, err := f.c.ComponentAccessToken(ctx)
	if err != nil {
		return err
	}
	fullPath := path + "?component_access_token=" + url.QueryEscape(token)

	var raw json.RawMessage
	if err := f.c.http.Post(ctx, fullPath, body, &raw); err != nil {
		return fmt.Errorf("oplatform: %s: %w", path, err)
	}
	return decodeRaw(path, raw, out)
}
```

复用 `wxa.client.go` 中已定义的包级 `decodeRaw` 助手（两段式 JSON 解码 + errcode 折叠）。

### 2.3 Query 参数差异

快速注册接口使用的 query key 是 `component_access_token`，**不是** WxaAdmin 的 `access_token`。这是必须分开 doPost 的核心原因 —— 二者不能简单复用。

### 2.4 action query 参数

大多数 fastregister endpoint 会在路径后带 `?action=create` 或 `?action=search`。我们把 `action` 写进方法名里的 endpoint 常量，而不是参数化 —— 每个"创建/查询"是独立的语义动作，分别是独立方法比一个 enum 参数更清楚。

例如 `CreateEnterpriseAccount` 对应固定的 `?action=create`，`QueryEnterpriseAccount` 对应固定的 `?action=search`。action 参数会和 `component_access_token` 一起拼到 URL：

```go
f.doPost(ctx, "/cgi-bin/component/fastregisterweapp?action=create", req, &resp)
```

`doPost` 里的 `fullPath := path + "?component_access_token=..."` 需要用 `&` 而非 `?` 当 path 已经含 query。让 `doPost` 自己处理：

```go
sep := "?"
if strings.Contains(path, "?") {
	sep = "&"
}
fullPath := path + sep + "component_access_token=" + url.QueryEscape(token)
```

这个小分支合进 doPost 实现里。`WxaAdminClient.doPost` 不需要同样处理，因为 WxaAdmin 的 endpoint 都是纯路径（action 在 body 里）。

### 2.5 文件布局

```
oplatform/
  fastregister.go          [NEW] FastRegisterClient + doPost + 7 个 component 级方法
  fastregister.struct.go   [NEW] FastRegister 的 DTO
  fastregister_test.go     [NEW] 8 用例（7 happy path + 1 errcode）
  wxa.account.go           [MOD] 追加 GetAccountBasicInfo
  wxa.account_test.go      [MOD] 追加 TestWxaAdmin_GetAccountBasicInfo
  wxa.struct.go            [MOD] 追加 WxaAccountBasicInfo DTO
```

FastRegister 的 DTO 单独放 `fastregister.struct.go`，不混进 `wxa.struct.go` —— 它们语义上不属于 WxaAdmin 家族。

## 3. 接口清单

### 3.1 FastRegisterClient（7 个 component 级方法）

| 方法 | Endpoint | 说明 |
|---|---|---|
| `CreateEnterpriseAccount(ctx, req *FastRegEnterpriseReq) (*FastRegEnterpriseResp, error)` | `POST /cgi-bin/component/fastregisterweapp?action=create` | 企业快速注册（通过法人身份证 + 企业代码） |
| `QueryEnterpriseAccount(ctx, legalPersonaWechat, legalPersonaName string) (*FastRegEnterpriseStatus, error)` | `POST /cgi-bin/component/fastregisterweapp?action=search` | 查询创建任务状态 |
| `CreatePersonalAccount(ctx, req *FastRegPersonalReq) (*FastRegPersonalResp, error)` | `POST /cgi-bin/component/fastregisterpersonalweapp?action=create` | 个人类型小程序注册 |
| `QueryPersonalAccount(ctx, taskID string) (*FastRegPersonalStatus, error)` | `POST /cgi-bin/component/fastregisterpersonalweapp?action=query` | 查询个人注册任务 |
| `CreateBetaAccount(ctx, req *FastRegBetaReq) (*FastRegBetaResp, error)` | `POST /cgi-bin/component/fastregisterbetaweapp?action=create` | 复用主体创建试用版小程序 |
| `QueryBetaAccount(ctx, uniqueID string) (*FastRegBetaStatus, error)` | `POST /cgi-bin/component/fastregisterbetaweapp?action=search` | 查询试用版创建任务 |
| `GenerateAdminRebindQrcode(ctx, redirectURI string) (*RebindAdminQrcode, error)` | `POST /cgi-bin/account/componentrebindadmin` | 生成小程序管理员变更二维码 |

### 3.2 WxaAdminClient（1 个 authorizer 级方法）

| 方法 | Endpoint |
|---|---|
| `GetAccountBasicInfo(ctx) (*WxaAccountBasicInfo, error)` | `GET /cgi-bin/account/getaccountbasicinfo` |

此方法复用 `wxa.client.go` 的 `doGet` 助手；挂在 `WxaAdminClient` 上因为它需要 `authorizer_access_token`。

## 4. DTO 设计

### 4.1 `fastregister.struct.go`

```go
package oplatform

// ----- enterprise -----

type FastRegEnterpriseReq struct {
	Name               string `json:"name"`
	Code               string `json:"code"`
	CodeType           int    `json:"code_type"` // 1=统一社会信用代码 2=组织机构代码 3=营业执照注册号
	LegalPersonaWechat string `json:"legal_persona_wechat"`
	LegalPersonaName   string `json:"legal_persona_name"`
	ComponentPhone     string `json:"component_phone"`
}

type FastRegEnterpriseResp struct{}

type FastRegEnterpriseStatus struct {
	Status          int    `json:"status"`
	AuthCode        string `json:"auth_code,omitempty"`
	AuthorizerAppid string `json:"authorizer_appid,omitempty"`
	IsWxVerify      bool   `json:"is_wx_verify,omitempty"`
	IsLinkMp        bool   `json:"is_link_mp,omitempty"`
}

// ----- personal -----

type FastRegPersonalReq struct {
	IDName         string `json:"idname"`
	WxUser         string `json:"wxuser"`
	ComponentPhone string `json:"component_phone,omitempty"`
}

type FastRegPersonalResp struct {
	TaskID string `json:"taskid"`
}

type FastRegPersonalStatus struct {
	Status            int    `json:"status"`
	AppID             string `json:"appid,omitempty"`
	AuthorizationCode string `json:"authorization_code,omitempty"`
}

// ----- beta -----

type FastRegBetaReq struct {
	Name               string `json:"name"`
	Code               string `json:"code"`
	CodeType           int    `json:"code_type"`
	LegalPersonaWechat string `json:"legal_persona_wechat"`
	LegalPersonaName   string `json:"legal_persona_name"`
	ComponentPhone     string `json:"component_phone"`
}

type FastRegBetaResp struct {
	UniqueID string `json:"unique_id"`
}

type FastRegBetaStatus struct {
	Status            int    `json:"status"`
	AppID             string `json:"appid,omitempty"`
	AuthorizationCode string `json:"authorization_code,omitempty"`
}

// ----- admin rebind -----

type RebindAdminQrcode struct {
	TaskID    string `json:"taskid"`
	QrcodeURL string `json:"qrcode_url"`
}
```

### 4.2 `wxa.struct.go` 追加

```go

// ----- account basic info -----

type WxaAccountBasicInfo struct {
	AppID             string `json:"appid"`
	AccountType       int    `json:"account_type"`
	PrincipalType     int    `json:"principal_type"`
	PrincipalName     string `json:"principal_name"`
	RealnameStatus    int    `json:"realname_status"`
	Nickname          string `json:"nickname,omitempty"`
	HeadImg           string `json:"head_img,omitempty"`
	Signature         string `json:"signature,omitempty"`
	RegisteredCountry int    `json:"registered_country,omitempty"`
	WxVerifyInfo      struct {
		QualificationVerify bool `json:"qualification_verify"`
		NamingVerify        bool `json:"naming_verify"`
	} `json:"wx_verify_info,omitempty"`
	SignatureInfo struct {
		Signature       string `json:"signature"`
		ModifyUsedCount int    `json:"modify_used_count"`
		ModifyQuota     int    `json:"modify_quota"`
	} `json:"signature_info,omitempty"`
	HeadImageInfo struct {
		HeadImageURL    string `json:"head_image_url"`
		ModifyUsedCount int    `json:"modify_used_count"`
		ModifyQuota     int    `json:"modify_quota"`
	} `json:"head_image_info,omitempty"`
}
```

## 5. 错误处理

- 业务错误 (`errcode != 0`) → `*oplatform.WeixinError`（和现有接口一致）
- `ComponentAccessToken` / `AuthorizerClient.AccessToken` 的错误透传（可能 `ErrVerifyTicketMissing` / `ErrAuthorizerRevoked`）
- 特殊 errcode（如 89251 法人验证中）不做特殊分支；调用方用 `errors.As(*WeixinError)` 判断

## 6. 并发与生命周期

- `FastRegisterClient` 无状态，复用 `Client` 的 `componentMu` 单飞
- 所有方法接受 `ctx context.Context`
- 构造不做 I/O：`c.FastRegister()` 返回 struct 包装即可
- 可在多 goroutine 中共享同一 `FastRegisterClient` 实例

## 7. 测试策略

沿用 httptest 驱动模式。

### 7.1 测试矩阵

| 测试文件 | 覆盖 | 用例数 |
|---|---|---|
| `fastregister_test.go` | 7 个 FastRegister 方法 happy path + 1 errcode 分支 | 8 |
| `wxa.account_test.go` | 追加 `TestWxaAdmin_GetAccountBasicInfo` | 1 |

**合计 9 个新用例**。

### 7.2 FastRegister 测试助手

```go
// fastregister_test.go
func newTestFastRegister(t *testing.T, baseURL string) *FastRegisterClient {
	t.Helper()
	store := NewMemoryStore()
	_ = store.SetVerifyTicket(context.Background(), "TICKET")
	// 预置未过期的 component_access_token，这样 ComponentAccessToken 命中缓存
	_ = store.SetComponentToken(context.Background(), "CTOK", time.Now().Add(time.Hour))
	c := newTestClient(t, baseURL, WithStore(store))
	return c.FastRegister()
}
```

**关键**：预置 `CTOK` 为未过期 → `ComponentAccessToken()` 不触发 `/cgi-bin/component/api_component_token` 刷新，测试 mock 只需 handle 具体业务 endpoint。

`GetAccountBasicInfo` 测试则继续用现有的 `newTestWxaAdmin` 助手（预置 authorizer tokens）。

## 8. 兼容性

纯 additive：
- `Client` 新增一个方法 `FastRegister() *FastRegisterClient`
- 新增类型 `FastRegisterClient` + 7 个方法
- 新增 DTO 类型（约 10 个 struct）
- `WxaAdminClient` 新增一个方法 `GetAccountBasicInfo`
- 无任何 breaking change

## 9. 交付物

1. `oplatform/fastregister.go` (~140 生产行，含 doPost helper + 7 methods)
2. `oplatform/fastregister.struct.go` (~80 行 DTO)
3. `oplatform/fastregister_test.go` (~220 行，8 用例)
4. `oplatform/wxa.account.go` 追加 `GetAccountBasicInfo`（~12 行）
5. `oplatform/wxa.account_test.go` 追加 `TestWxaAdmin_GetAccountBasicInfo`（~25 行）
6. `oplatform/wxa.struct.go` 追加 `WxaAccountBasicInfo`（~30 行 DTO）

**生产代码 ~260 行，测试 ~245 行，合计 ~500 行**。预计 3 个原子 commit：

1. **commit 1** — FastRegisterClient 骨架 (type + FastRegister() factory + doPost helper + fastregister.struct.go DTO + fastregister_test.go doPost 基础测试)
2. **commit 2** — FastRegister 7 个业务方法 + 对应 7 个测试
3. **commit 3** — WxaAdmin.GetAccountBasicInfo + DTO + 测试

## 10. 非目标 / 后续工作

- 账号注销 family（下一轮或合并进子项目 5.5）
- 子项目 7（代运营：插件、模板库、订阅消息、客服、广告）
- 复用公众号主体创建新账号（罕见路径）

## 11. 开放问题

无未决项。所有关键决策已锁定：

- 架构：`FastRegisterClient` wrapper via `Client.FastRegister()` ✔
- 覆盖：8 方法（7 component + 1 authorizer）✔
- 文件：独立 `fastregister.go`/`fastregister.struct.go`；`GetAccountBasicInfo` 挂在 WxaAdmin ✔
- DTO 分家：`fastregister.struct.go` 独立，不混进 `wxa.struct.go` ✔
- doPost：query key 用 `component_access_token`；处理 path 已含 `?action=xxx` 的情况 ✔
- 测试：sharing decodeRaw from wxa.client.go ✔
- 兼容：纯 additive ✔
