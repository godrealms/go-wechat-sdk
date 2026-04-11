# oplatform 代小程序开发管理 (WxaAdmin) — 设计文档

**Date:** 2026-04-12
**Status:** Draft — awaiting user review
**Scope:** 微信开放平台子项目 3 — 代小程序开发管理

## 1. 背景

本文档是 oplatform 的第二个实现迭代。前一轮 (`2026-04-12-oplatform-auth-foundation-design.md`) 已交付子项目 1+2+6（授权底座、authorizer 代调用框架、扫码登录）。本轮专注子项目 3：通过已授权 authorizer 代小程序完成开发与发布全流程。

### 1.1 目标

覆盖官方"代小程序实现业务"文档中"开发管理"相关的所有接口，共 6 个子族 31 个方法：

- 账号管理（5）：名称、头像、签名
- 类目管理（5）：增删改查
- 域名管理（4）：服务器域名、业务域名、快速配置
- 成员管理（3）：绑定 / 解绑 / 列出体验者
- 代码管理（4）：上传、页面、体验二维码、可选类目
- 审核发布（10）：提交审核、撤回、加急、发布、回退、可见状态、支持版本

### 1.2 显式不做

- 快速注册小程序 / 公众号（子项目 5）
- 代运营：插件、模板库、广告、订阅消息、客服（子项目 7）
- 隐私接口 / 用户隐私设置
- 小程序码 `wxacode/*`（自有 appid 也能调，归 mini-program 模块）
- 审核回调事件解密（应通过扩展 `ParseNotify` 实现，属于 notify.go 的后续迭代）

## 2. 架构

### 2.1 WxaAdminClient 包装器

```go
// oplatform/wxa.client.go
package oplatform

// WxaAdminClient 代小程序开发管理客户端。
// 所有方法都以 AuthorizerClient.AccessToken() 作为 token 源。
// 无状态，线程安全，可在多 goroutine 共享。
type WxaAdminClient struct {
    auth *AuthorizerClient
}

// WxaAdmin 从 AuthorizerClient 构造开发管理客户端。
func (a *AuthorizerClient) WxaAdmin() *WxaAdminClient {
    return &WxaAdminClient{auth: a}
}
```

- `AuthorizerClient` 上只增加一个方法 `WxaAdmin()`
- `WxaAdminClient` 包一层，持有 `*AuthorizerClient` 的引用
- 所有业务方法挂在 `WxaAdminClient` 上，不污染 `AuthorizerClient`
- 复用 `AuthorizerClient` 的 per-appid 锁和 token 缓存

### 2.2 共用 doPost / doGet / doGetRaw

```go
// doPost 通用 POST 辅助：token 自动拼接到 query，errcode != 0 自动转错误。
// out 为 nil 时不解析响应体（适用于"仅检查 errcode"的接口）。
func (w *WxaAdminClient) doPost(ctx context.Context, path string, body, out any) error {
    ctx = touchContext(ctx)
    token, err := w.auth.AccessToken(ctx)
    if err != nil {
        return err
    }
    fullPath := path + "?access_token=" + url.QueryEscape(token)

    var raw json.RawMessage
    if err := w.auth.c.http.Post(ctx, fullPath, body, &raw); err != nil {
        return fmt.Errorf("oplatform: %s: %w", path, err)
    }
    var base struct {
        ErrCode int    `json:"errcode"`
        ErrMsg  string `json:"errmsg"`
    }
    _ = json.Unmarshal(raw, &base)
    if err := checkWeixinErr(base.ErrCode, base.ErrMsg); err != nil {
        return err
    }
    if out != nil {
        if err := json.Unmarshal(raw, out); err != nil {
            return fmt.Errorf("oplatform: %s decode: %w", path, err)
        }
    }
    return nil
}

// doGet — GET with typed query params, JSON response
func (w *WxaAdminClient) doGet(ctx context.Context, path string, q url.Values, out any) error

// doGetRaw — binary response; used by GetQrcode only
func (w *WxaAdminClient) doGetRaw(ctx context.Context, path string, q url.Values) (body []byte, contentType string, err error)
```

**关键决策**：`doPost` 先把响应读成 `json.RawMessage`，然后做两次 unmarshal —— 一次拿 `errcode`/`errmsg`，一次拿 caller 的 typed struct。这避免了"业务响应里混杂 errcode 字段"时的字段冲突。

### 2.3 文件布局

```
oplatform/
  wxa.client.go          WxaAdminClient + WxaAdmin() + doPost/doGet/doGetRaw
  wxa.struct.go          所有子族共用 DTO
  wxa.account.go         账号子族 (5 方法)
  wxa.category.go        类目子族 (5 方法)
  wxa.domain.go          域名子族 (4 方法)
  wxa.tester.go          成员子族 (3 方法)
  wxa.code.go            代码子族 (4 方法)
  wxa.release.go         审核发布子族 (10 方法)
  wxa.client_test.go     doPost/doGet/doGetRaw 单元测试
  wxa.account_test.go
  wxa.category_test.go
  wxa.domain_test.go
  wxa.tester_test.go
  wxa.code_test.go
  wxa.release_test.go
```

- 一个子族一对 `.go` + `_test.go`
- 平均每文件 ~80 行生产代码，~130 行测试
- 没有单文件超过 200 行（`wxa.release.go` 最大）

### 2.4 命名约定

- 方法名：Pascal Case 翻译官方 endpoint。`api_wxa_modify_domain` → `ModifyServerDomain`
- DTO 统一 `Wxa` 前缀：`WxaCommitReq`, `WxaAuditStatus`, `WxaCategoryItem`
- 请求结构体后缀 `Req`，响应结构体后缀 `Resp`

## 3. 接口清单

### 3.1 wxa.account.go — 账号管理

| 方法 | Endpoint | 说明 |
|---|---|---|
| `SetNickname(ctx, req *WxaSetNicknameReq) (*WxaSetNicknameResp, error)` | `/wxa/setnickname` | 设置小程序名称，含法人/非法人双模式（所有字段 omitempty） |
| `QueryNickname(ctx, auditID string) (*WxaQueryNicknameResp, error)` | `/wxa/api_wxa_querynickname` | 查询改名审核状态 |
| `CheckNickname(ctx, nickname string) (*WxaCheckNicknameResp, error)` | `/cgi-bin/wxverify/checkwxverifynickname` | 名称合法性预检 |
| `ModifyHeadImage(ctx, mediaID string) error` | `/cgi-bin/account/modifyheadimage` | 修改头像 |
| `ModifySignature(ctx, signature string) error` | `/cgi-bin/account/modifysignature` | 修改功能介绍 |

### 3.2 wxa.category.go — 类目管理

| 方法 | Endpoint |
|---|---|
| `GetCategory(ctx) (*WxaGetCategoryResp, error)` | `/cgi-bin/wxopen/getcategory` |
| `GetAllCategories(ctx) (*WxaGetAllCategoriesResp, error)` | `/cgi-bin/wxopen/getallcategories` |
| `AddCategory(ctx, req *WxaAddCategoryReq) error` | `/cgi-bin/wxopen/addcategory` |
| `DeleteCategory(ctx, first, second int) error` | `/cgi-bin/wxopen/deletecategory` |
| `ModifyCategory(ctx, req *WxaModifyCategoryReq) error` | `/cgi-bin/wxopen/modifycategory` |

### 3.3 wxa.domain.go — 域名管理

| 方法 | Endpoint |
|---|---|
| `ModifyServerDomain(ctx, req *WxaModifyServerDomainReq) (*WxaServerDomainResp, error)` | `/wxa/modify_domain` |
| `SetWebviewDomain(ctx, req *WxaSetWebviewDomainReq) error` | `/wxa/setwebviewdomain` |
| `GetDomainConfirmFile(ctx) (*WxaDomainConfirmFile, error)` | `/wxa/get_webview_confirmfile` |
| `ModifyDomainDirectly(ctx, req *WxaModifyDomainDirectlyReq) (*WxaServerDomainResp, error)` | `/wxa/modify_domain_directly` |

`WxaModifyServerDomainReq` 包含 `action` (add/delete/set/get/delete_legal_domain) + 4 类域名数组（request/wsrequest/upload/download），所有字段 `omitempty`。

### 3.4 wxa.tester.go — 成员管理

| 方法 | Endpoint |
|---|---|
| `BindTester(ctx, wechatID string) (*WxaBindTesterResp, error)` | `/wxa/bind_tester` |
| `UnbindTester(ctx, wechatID, userStr string) error` | `/wxa/unbind_tester` |
| `ListTesters(ctx) (*WxaListTestersResp, error)` | `/wxa/memberauth` |

`UnbindTester` 接受 `wechatID` 或 `userStr`（二选一），内部组装 body。

### 3.5 wxa.code.go — 代码管理

| 方法 | Endpoint |
|---|---|
| `Commit(ctx, req *WxaCommitReq) error` | `/wxa/commit` |
| `GetPage(ctx) (*WxaGetPageResp, error)` | `/wxa/get_page` |
| `GetQrcode(ctx, path string) (body []byte, contentType string, err error)` | `/wxa/get_qrcode` |
| `GetCodeCategory(ctx) (*WxaGetCodeCategoryResp, error)` | `/wxa/get_category` |

- `GetQrcode` 是本轮**唯一一个非 JSON 接口**：返回二进制图片，需要 `utils.HTTP.DoRequestWithRawResponse`
- `GetCodeCategory` 与 `GetCategory`（账号类目）是两个不同接口，分别是 `/wxa/get_category` 与 `/cgi-bin/wxopen/getcategory`

### 3.6 wxa.release.go — 审核与发布

| 方法 | Endpoint |
|---|---|
| `SubmitAudit(ctx, req *WxaSubmitAuditReq) (*WxaSubmitAuditResp, error)` | `/wxa/submit_audit` |
| `GetAuditStatus(ctx, auditID int64) (*WxaAuditStatus, error)` | `/wxa/get_auditstatus` |
| `GetLatestAuditStatus(ctx) (*WxaAuditStatus, error)` | `/wxa/get_latest_auditstatus` |
| `UndoCodeAudit(ctx) error` | `/wxa/undocodeaudit` |
| `SpeedupAudit(ctx, auditID int64) error` | `/wxa/speedupaudit` |
| `Release(ctx) error` | `/wxa/release` |
| `RevertCodeRelease(ctx) error` | `/wxa/revertcoderelease` |
| `ChangeVisitStatus(ctx, action string) error` | `/wxa/change_visitstatus` — open/close |
| `GetSupportVersion(ctx) (*WxaSupportVersionResp, error)` | `/cgi-bin/wxopen/getweappsupportversion` |
| `SetSupportVersion(ctx, version string) error` | `/cgi-bin/wxopen/setweappsupportversion` |

## 4. 错误处理

- 所有业务错误 (`errcode != 0`) → `*oplatform.WeixinError`
- HTTP / 网络错误透传 `*utils.HTTPError` 或 `net.Error`
- `ComponentAccessToken` / `AuthorizerClient.AccessToken` 的错误透传（可能是 `ErrVerifyTicketMissing` 或 `ErrAuthorizerRevoked`）
- 特殊 errcode（如审核中 85013、频率限制 45009）直接透传，不做特殊分支；调用方用 `errors.As(*WeixinError)` 判断
- 禁止 panic；禁止吞错误

## 5. 并发与生命周期

- `WxaAdminClient` 无自己的状态，复用 `AuthorizerClient.AccessToken` 的 per-appid `sync.Mutex`
- 所有方法接受 `ctx context.Context`，可取消 / 超时
- 构造不做 I/O：`auth.WxaAdmin()` 纯粹返回 struct 包装
- 可在多 goroutine 中共享单个 `WxaAdminClient` 实例

## 6. 测试策略

沿用现有模块（offiaccount、oplatform 已有 34+ 测试）的 `httptest` 驱动风格。

| 测试文件 | 覆盖 | 预计用例数 |
|---|---|---|
| `wxa.client_test.go` | doPost 成功路径 / errcode 路径 / unmarshal 失败；doGet 成功；doGetRaw 成功路径 + content-type 断言 | ~4 |
| `wxa.account_test.go` | 5 个 account 方法 happy path + SetNickname 失败分支 | ~5 |
| `wxa.category_test.go` | 5 个类目方法 happy path + 1 个 errcode 分支 | ~5 |
| `wxa.domain_test.go` | 4 个域名方法 + action 参数断言 | ~4 |
| `wxa.tester_test.go` | 3 个成员方法 | ~3 |
| `wxa.code_test.go` | 4 个代码方法，`GetQrcode` 验证二进制字节 + content-type | ~4 |
| `wxa.release_test.go` | 10 个发布方法 happy path + 审核中 errcode + Release 错误分支 | ~10 |

**合计约 35 个测试用例**。每个测试：
1. 启动 `httptest.Server` 配置具体 endpoint 的 handler
2. 构造 `MemoryStore` 预置 `TICKET` + 一个未过期的 `wxBiz` authorizer
3. 用 `newTestClient(t, baseURL, WithStore(store)).Authorizer("wxBiz").WxaAdmin()`
4. 调用具体方法，断言请求 path / query / body 和响应解析

引入一个小的共享 helper（在 `wxa.client_test.go` 中）以消除重复：

```go
func newTestWxaAdmin(t *testing.T, baseURL string) *WxaAdminClient {
    t.Helper()
    store := NewMemoryStore()
    _ = store.SetVerifyTicket(context.Background(), "TICKET")
    _ = store.SetAuthorizer(context.Background(), "wxBiz", AuthorizerTokens{
        AccessToken: "ATOK", RefreshToken: "R",
        ExpireAt: time.Now().Add(time.Hour),
    })
    c := newTestClient(t, baseURL, WithStore(store))
    return c.Authorizer("wxBiz").WxaAdmin()
}
```

**重点**：预置 `AccessToken = "ATOK"` 且 `ExpireAt` 在未来 → `AuthorizerClient.AccessToken` 走缓存路径，不触发 `/cgi-bin/component/api_authorizer_token` refresh，测试 mock 只需 handle 业务 endpoint。

测试深度与 §9 一致，覆盖主要路径 + 少量错误分支，不强求覆盖率数字。

## 7. 兼容性

**零破坏**：
- `AuthorizerClient` 新增一个方法 `WxaAdmin()`，签名完全 additive
- 不动 `offiaccount` / `mini-program` / `utils`
- 不动 oplatform 其它任何现有类型或函数
- `WxaAdminClient` 是全新类型

**新增导出符号**：
- `WxaAdminClient` type
- `*AuthorizerClient.WxaAdmin() *WxaAdminClient`
- 31 个 `*WxaAdminClient` 方法
- 对应的 `Wxa*Req` / `Wxa*Resp` DTO 类型（约 20 个 struct）

## 8. 交付物

1. `oplatform/wxa.client.go` (~80 生产行)
2. `oplatform/wxa.struct.go` (~200 行 DTO)
3. `oplatform/wxa.account.go` (~80 行)
4. `oplatform/wxa.category.go` (~80 行)
5. `oplatform/wxa.domain.go` (~90 行)
6. `oplatform/wxa.tester.go` (~60 行)
7. `oplatform/wxa.code.go` (~100 行)
8. `oplatform/wxa.release.go` (~180 行)
9. 7 个配对 `*_test.go`（合计 ~900 行）
10. `README.md` / `docs/README.md` 注释更新（本轮已支持代小程序发布全流程）

**生产代码 ~870 行，测试 ~900 行，合计 ~1770 行**。预计 12-14 个原子 commit（每个子族 1-2 个）。

## 9. 非目标 / 后续工作

- 子项目 5（快速注册小程序 / 公众号）
- 子项目 7（代运营：插件、模板库、广告、订阅消息、客服）
- 隐私设置接口
- 审核回调事件的 `ParseNotify` 扩展
- 小程序商户基础库、生成小程序码 / 二维码图片（非 `get_qrcode` 那种）

## 10. 开放问题

无未决项。所有关键决策已锁定：

- 架构：`WxaAdminClient` wrapper via `AuthorizerClient.WxaAdmin()` ✔
- 覆盖：全量 6 子族 31 方法 ✔
- 文件：一个子族一对 `.go` + `_test.go` ✔
- 通用助手：`doPost` / `doGet` / `doGetRaw` ✔
- 错误：统一 `*WeixinError`，不做特殊 errcode 分支 ✔
- 测试：httptest 驱动，~35 用例，覆盖主要路径 ✔
- 兼容：零破坏，纯 additive ✔
