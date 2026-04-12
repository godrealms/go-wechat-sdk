# work-wechat ISV 子项目 2:服务商自身接口设计

**Date:** 2026-04-12
**Status:** Draft
**Scope:** 企业微信 work-wechat/isv 包 —— 服务商(provider)自身接口家族
**Depends on:** 子项目 1(ISV 认证底座,已完成,commits e72594a..33e4e9e)

---

## 1. 目标

在已完成的 `work-wechat/isv` 包中追加 **4 个公开方法**,覆盖《企业微信服务商自身接口》中尚未实现的 3 个远程接口 + 1 个 provider_access_token 的公开包装。

追加后的 `work-wechat/isv` 公开 API 累计 **20 个方法**(子项目 1 贡献 16 个)。

### 1.1 必须交付的 4 个方法

| # | 签名 | 用途 | Token 类型 |
|---|---|---|---|
| 1 | `(c *Client) GetProviderAccessToken(ctx) (string, error)` | 公开包装现有 `getProviderAccessToken`,让调用方可以直接拿到 provider_access_token(用于自建请求、调试、或嵌入到下游子项目) | provider |
| 2 | `(c *Client) GetLoginInfo(ctx, authCode string) (*LoginInfoResp, error)` | `service/get_login_info` —— 用服务商管理端 OAuth 回跳的 auth_code 换取登录身份 | provider |
| 3 | `(c *Client) GetRegisterCode(ctx, req *GetRegisterCodeReq) (*RegisterCodeResp, error)` | `service/get_register_code` —— 生成注册企业微信的邀请链接 + register_code | provider |
| 4 | `(c *Client) GetRegistrationInfo(ctx, registerCode string) (*RegistrationInfoResp, error)` | `service/get_registration_info` —— 查询注册进度,拿到企业 corpid、管理员 userid、永久授权码 | provider |

### 1.2 非目标

- `service/contact_sync_success` —— 与通讯录同步强耦合,留到子项目 3 一起做。
- 单个 `userid_to_openuserid` —— 已有的 `UserIDToOpenUserID` 本身就调用批量接口,不需要额外单值版本。
- register_code 的前端 URL 构造辅助方法 —— `RegisterCodeResp.RegisterURI` 由接口直接返回,无需再造。
- 服务商管理端 OAuth2 授权 URL 构造(不是远程接口) —— 可以后续按需追加,本轮不做。

## 2. 架构决策

### 2.1 仍然追加到 `work-wechat/isv` 包

**不建子包。** 这 4 个方法都绑定在 `*Client` 上,强依赖 `providerMu` / `store` / `cfg` 等私有字段。拆出子包会被迫暴露内部字段,违反封装。

### 2.2 文件拆分

```
work-wechat/isv/
├── provider.id_convert.go   # 已存在 —— 追加 GetProviderAccessToken 一行公开包装
├── provider.login.go        # NEW —— GetLoginInfo + GetRegisterCode + GetRegistrationInfo
├── provider.login_test.go   # NEW
├── struct.provider.go       # NEW —— 4 个响应 DTO + 1 个请求 DTO
```

**理由:** 按 API 家族切分,与子项目 1 的 `suite.*` / `corp.*` 命名风格一致。单个 `provider.login.go` ~150 行,不会膨胀。DTO 单独拎进 `struct.provider.go` 与子项目 1 的 `struct.suite.go` / `struct.permanent.go` / `struct.corp.go` / `struct.notify.go` 对齐。

### 2.3 Token 选择

| 方法 | HTTP helper | 注入的 token |
|---|---|---|
| GetProviderAccessToken | N/A | — |
| GetLoginInfo | `providerDoPost` | provider_access_token |
| GetRegisterCode | `providerDoPost` | provider_access_token |
| GetRegistrationInfo | `providerDoPost` | provider_access_token |

全部 4 个远程方法都走 `providerDoPost`(子项目 1 已实现),**不新增任何 HTTP helper**。

### 2.4 错误处理

- 沿用子项目 1 的两阶段 `decodeRaw`:先探 `errcode`,非零则返回 `WeixinError`,零则 unmarshal 到目标类型。
- 未配置 `ProviderCorpID` / `ProviderSecret` 时,`getProviderAccessToken` 已返回 `ErrProviderCorpIDMissing` / `ErrProviderSecretMissing`;上层透传即可。
- `GetRegisterCode` 的请求 DTO 有必填字段校验 —— 不在客户端做校验,直接透传给服务端,让 WeChat 返回 `WeixinError`(保持简单)。

## 3. DTO 设计(`struct.provider.go`)

### 3.1 get_login_info 响应

```go
// LoginInfoResp 是 service/get_login_info 的响应。
// UserType: 1 = 企业管理员,2 = 企业成员,3 = 服务商(代开发)成员。
// 当 UserType ∈ {1,2} 时 AuthInfo 会返回对应身份。
type LoginInfoResp struct {
    UserType int                 `json:"usertype"`
    UserInfo LoginInfoUser       `json:"user_info"`
    CorpInfo LoginInfoCorp       `json:"corp_info"`
    Agent    []LoginInfoAgent    `json:"agent"`
    AuthInfo LoginInfoPermission `json:"auth_info"`
}

type LoginInfoUser struct {
    UserID     string `json:"userid"`
    OpenUserID string `json:"open_userid"`
    Name       string `json:"name"`
    Avatar     string `json:"avatar"`
}

type LoginInfoCorp struct {
    CorpID string `json:"corpid"`
}

type LoginInfoAgent struct {
    AgentID  int    `json:"agentid"`
    AuthType int    `json:"auth_type"` // 0 只使用 1 管理
}

// LoginInfoPermission —— 当登录者身份是企业管理员时,包含被管理的部门列表。
type LoginInfoPermission struct {
    Department []LoginInfoDepartment `json:"department"`
}

type LoginInfoDepartment struct {
    ID       int  `json:"id"`
    Writable bool `json:"writable"`
}
```

### 3.2 get_register_code 请求 + 响应

```go
// GetRegisterCodeReq 是 service/get_register_code 的请求体。
// TemplateID / CorpName / AdminName / AdminMobile / State 均为可选,
// 按官方文档的默认值语义处理。
type GetRegisterCodeReq struct {
    TemplateID  string `json:"template_id,omitempty"`
    CorpName    string `json:"corp_name,omitempty"`
    AdminName   string `json:"admin_name,omitempty"`
    AdminMobile string `json:"admin_mobile,omitempty"`
    State       string `json:"state,omitempty"`
}

type RegisterCodeResp struct {
    RegisterCode string `json:"register_code"`
    ExpiresIn    int    `json:"expires_in"`
}
```

### 3.3 get_registration_info 响应

```go
// RegistrationInfoResp —— 查询注册进度后返回已注册企业 + 管理员 + 授权信息。
type RegistrationInfoResp struct {
    CorpInfo      RegistrationCorpInfo  `json:"corp_info"`
    AuthUserInfo  RegistrationAdminInfo `json:"auth_user_info"`
    ContactSync   RegistrationContact   `json:"contact_sync"`
    AuthInfo      AuthInfoAgent         `json:"auth_info"` // 复用子项目 1 的 AuthInfoAgent
    PermanentCode string                `json:"permanent_code"`
}

type RegistrationCorpInfo struct {
    CorpID       string `json:"corpid"`
    CorpName     string `json:"corp_name"`
    CorpType     string `json:"corp_type"`
    CorpSquareLogoURL string `json:"corp_square_logo_url"`
    CorpUserMax  int    `json:"corp_user_max"`
    SubjectType  int    `json:"subject_type"`
    VerifiedEndTime int `json:"verified_end_time"`
    CorpWxqrcode string `json:"corp_wxqrcode"`
    CorpScale    string `json:"corp_scale"`
    CorpIndustry string `json:"corp_industry"`
    CorpSubIndustry string `json:"corp_sub_industry"`
}

type RegistrationAdminInfo struct {
    UserID string `json:"userid"`
    Name   string `json:"name"`
}

type RegistrationContact struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}
```

**复用决策:** `RegistrationInfoResp.AuthInfo` 复用子项目 1 `struct.permanent.go` 里的 `AuthInfoAgent`(字段布局 100% 一致)。`PermanentCode` 扁平在顶层,与官方文档对齐。

## 4. 实现细节

### 4.1 `GetProviderAccessToken`(`provider.id_convert.go` 追加)

```go
// GetProviderAccessToken 返回当前 provider_access_token(lazy 获取 + 自动缓存)。
// 如果未在 Config 里配置 ProviderCorpID / ProviderSecret,返回对应哨兵错误。
func (c *Client) GetProviderAccessToken(ctx context.Context) (string, error) {
    return c.getProviderAccessToken(ctx)
}
```

一行公开包装。测试通过现有 `provider.id_convert_test.go`(如无,新加到 `provider.login_test.go`)的 1 个 case 覆盖。

### 4.2 `GetLoginInfo`(`provider.login.go`)

```go
func (c *Client) GetLoginInfo(ctx context.Context, authCode string) (*LoginInfoResp, error) {
    body := map[string]string{"auth_code": authCode}
    var resp LoginInfoResp
    if err := c.providerDoPost(ctx, "/cgi-bin/service/get_login_info", body, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.3 `GetRegisterCode`(`provider.login.go`)

```go
func (c *Client) GetRegisterCode(ctx context.Context, req *GetRegisterCodeReq) (*RegisterCodeResp, error) {
    if req == nil {
        req = &GetRegisterCodeReq{}
    }
    var resp RegisterCodeResp
    if err := c.providerDoPost(ctx, "/cgi-bin/service/get_register_code", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.4 `GetRegistrationInfo`(`provider.login.go`)

```go
func (c *Client) GetRegistrationInfo(ctx context.Context, registerCode string) (*RegistrationInfoResp, error) {
    body := map[string]string{"register_code": registerCode}
    var resp RegistrationInfoResp
    if err := c.providerDoPost(ctx, "/cgi-bin/service/get_registration_info", body, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

## 5. 测试策略(`provider.login_test.go`)

沿用子项目 1 建立的 `newTestISVClientWithProvider` 模式(httptest + 预置 provider_token)。

### 5.1 测试矩阵(~7 cases)

| # | Case | 断言 |
|---|---|---|
| 1 | `TestGetProviderAccessToken_Exposed` | 调用公开方法,返回预置的 provider token |
| 2 | `TestGetProviderAccessToken_MissingConfig` | 未配置 ProviderCorpID → 返回 `ErrProviderCorpIDMissing` |
| 3 | `TestGetLoginInfo_Admin` | 模拟管理员登录响应,断言 UserType=1、UserID、CorpID、AuthInfo.Department 非空 |
| 4 | `TestGetLoginInfo_Member` | 模拟普通成员登录,断言 UserType=2 且 AuthInfo.Department 为空 |
| 5 | `TestGetLoginInfo_InjectsProviderToken` | 断言请求 URL 的 query 携带 `provider_access_token=<预置值>` |
| 6 | `TestGetRegisterCode_HappyPath` | POST body 字段映射正确,返回 RegisterCode + ExpiresIn |
| 7 | `TestGetRegistrationInfo_HappyPath` | 解析完整响应,断言 corpid / permanent_code / auth_info.agent[0].agentid |

覆盖率目标: ≥85%(与子项目 1 对齐)。

### 5.2 共享 helper

新增 helper `seedProviderToken(t, c)`(如子项目 1 没有):

```go
func seedProviderToken(t *testing.T, c *Client) {
    t.Helper()
    _ = c.store.PutProviderToken(context.Background(), "suite1", "PTOK", time.Now().Add(time.Hour))
}
```

如果子项目 1 已有同名 helper,直接复用。

## 6. 错误类型

本轮 **不引入新的哨兵错误**。复用:

- `ErrProviderCorpIDMissing` / `ErrProviderSecretMissing`(子项目 1)
- `WeixinError`(子项目 1)
- 裸 HTTP 错误(子项目 1)

## 7. 镜像关系

这一轮直接镜像 `oplatform` 的 `wxa-admin` 模式:小规模增量,复用已有 HTTP helper,不扩展 Config/Store/Client 结构。参考文件 `oplatform/wxa.admin.go`。

## 8. 交付规模估计

| 文件 | 生产行数 | 测试行数 |
|---|---|---|
| `provider.id_convert.go`(追加) | +4 | — |
| `provider.login.go`(新) | ~60 | — |
| `struct.provider.go`(新) | ~90 | — |
| `provider.login_test.go`(新) | — | ~220 |
| **合计** | **~154** | **~220** |

## 9. Commit 节奏预估

~4 个原子 commit(TDD 风格,与子项目 1 对齐):

1. DTO:`struct.provider.go` —— 所有响应类型
2. `GetProviderAccessToken` 公开包装 + test
3. `GetLoginInfo` + test(包含 usertype=1/2 两个 case)
4. `GetRegisterCode` + `GetRegistrationInfo` + tests

## 10. Self-Review Checklist(填充后续 plan 时勾选)

- [ ] 4 个公开方法全部实现
- [ ] 所有 DTO json tag 与官方文档字段名一致
- [ ] 未新增 HTTP helper(全部复用 `providerDoPost`)
- [ ] 未修改 Config / Store / Client 结构
- [ ] 未新增哨兵错误
- [ ] 测试全部走 httptest,无真实网络
- [ ] `go test -race ./work-wechat/isv/...` 通过
- [ ] 覆盖率 ≥85%
- [ ] `example/main.go` 展示 `GetLoginInfo` 用法(追加 ~10 行,可选)

---

**spec 完成,等待用户确认后进入 writing-plans。**
