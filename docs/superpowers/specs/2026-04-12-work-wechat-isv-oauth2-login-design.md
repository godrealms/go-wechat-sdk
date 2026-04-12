# work-wechat ISV 子项目 7:小程序/网页 OAuth2 登录设计

**Date:** 2026-04-12
**Status:** Draft
**Scope:** 企业微信 work-wechat/isv 包 —— 第三方网页/小程序 OAuth2 登录三件套
**Depends on:** 子项目 1(ISV 认证底座)+ 子项目 2(provider 接口家族)

---

## 1. 目标

在已完成的 `work-wechat/isv` 包中追加 **3 个公开方法**,覆盖企业微信第三方网页授权(OAuth2)三件套:

1. 构造 OAuth2 授权 URL(调用方直接 302 重定向到企业微信)
2. 用回调带回的 `auth_code` 换取 `UserId` + `user_ticket`
3. 用 `user_ticket` 换取用户敏感详情(姓名/邮箱/头像/手机号)

追加后 `work-wechat/isv` 公开 API 累计 **23 个方法**(子项目 1:16,子项目 2:4)。

### 1.1 必须交付的 3 个方法

| # | 签名 | 用途 | Token |
|---|---|---|---|
| 1 | `(c *Client) OAuth2URL(redirectURI, state string, opts ...OAuth2Option) string` | 构造第三方网页授权 URL,调用方把它塞到 302 header 即可 | — (纯本地) |
| 2 | `(c *Client) GetUserInfo3rd(ctx, authCode string) (*UserInfo3rdResp, error)` | `/cgi-bin/service/auth/getuserinfo3rd` 换身份 | provider |
| 3 | `(c *Client) GetUserDetail3rd(ctx, userTicket string) (*UserDetail3rdResp, error)` | `/cgi-bin/service/auth/getuserdetail3rd` 换敏感详情 | provider |

### 1.2 非目标

- **JS-SDK ticket 与 signature**:需要 `corp_access_token` 作用域和额外的 Store 字段,留到后续子项目。
- **agent_config ticket**:同上。
- **小程序 `code2Session`**:企业微信小程序走 `jscode2session`(需要 suite_token),与 OAuth2 三件套的数据流不同,留到独立的子项目。
- **扫码登录(wwopen/sso/3rd_qrConnect)**:URL 形态与 OAuth2 不同,可作为 OAuth2URL 的一个 Option,但本轮 YAGNI 先不做。

## 2. 架构决策

### 2.1 追加到 `work-wechat/isv` 包

与子项目 2 一致 —— 3 个方法都绑在 `*Client` 上,依赖 `cfg.SuiteID`(URL 构造)和 `providerDoPost`(HTTP),拆子包会无谓暴露私有字段。

### 2.2 文件拆分

```
work-wechat/isv/
├── oauth2.go             # NEW —— OAuth2URL + OAuth2Option + GetUserInfo3rd + GetUserDetail3rd
├── oauth2_test.go        # NEW —— ~5 tests
└── struct.oauth2.go      # NEW —— UserInfo3rdResp + UserDetail3rdResp
```

单个 `oauth2.go` 容纳 URL 构造 + 2 个 HTTP 方法,总 ~90 行,不超过子项目 1 的同量级文件(`suite.permanent.go` 就做了 3 个方法)。

### 2.3 OAuth2URL 设计

**Base:** `https://open.weixin.qq.com/connect/oauth2/authorize`

**必填 query(按官方顺序组装):**
- `appid` = `cfg.SuiteID`
- `redirect_uri` = URL-encoded `redirectURI`
- `response_type` = `code`
- `scope` = 默认 `snsapi_privateinfo`(可经 `WithOAuth2Scope` 覆盖)
- `state` = 调用方传入
- `agentid` = 可选,经 `WithOAuth2AgentID(int)` 设置。**仅当 `scope=snsapi_privateinfo` 时必填**,由调用方负责正确性,本库不做强校验(保持 YAGNI)。

**尾部片段:** `#wechat_redirect`(浏览器规避缓存 + 微信识别)

**Functional Option 定义:**

```go
type OAuth2Option func(*oauth2Params)

type oauth2Params struct {
    scope    string
    agentID  int
    hasAgent bool
}

func WithOAuth2Scope(scope string) OAuth2Option {
    return func(p *oauth2Params) { p.scope = scope }
}

func WithOAuth2AgentID(agentID int) OAuth2Option {
    return func(p *oauth2Params) {
        p.agentID = agentID
        p.hasAgent = true
    }
}
```

默认 `scope = "snsapi_privateinfo"`,`agentid` 未设置则不出现在 query 中(不是 `agentid=0`)。

### 2.4 Token 选择

| 方法 | HTTP helper | 注入的 token |
|---|---|---|
| OAuth2URL | — | — |
| GetUserInfo3rd | `providerDoPost` | provider_access_token |
| GetUserDetail3rd | `providerDoPost` | provider_access_token |

官方文档:两个 `auth/getuser*3rd` 接口都使用服务商 provider_access_token 而非 suite_access_token。`providerDoPost` 已经在子项目 1 实现,无需新建 helper。

### 2.5 请求 / 响应方式的特殊性

- `GetUserInfo3rd` 是 **GET** `?code=<authCode>`,不是 POST。
- `GetUserDetail3rd` 是 **POST** `{"user_ticket": "..."}`。

这破坏了 `providerDoPost` 的假设(后者只做 POST)。解决方案:**新增一个 `providerDoGet` helper**,与现有 `providerDoPost` 对称。开销 ~10 行,职责清晰,值得写。

```go
// providerDoGet 对 GET 请求注入 provider_access_token。
func (c *Client) providerDoGet(ctx context.Context, path string, extra url.Values, out interface{}) error {
    tok, err := c.getProviderAccessToken(ctx)
    if err != nil {
        return err
    }
    q := url.Values{"provider_access_token": {tok}}
    for k, vs := range extra {
        q[k] = vs
    }
    return c.doRequestRaw(ctx, http.MethodGet, path, q, nil, out)
}
```

注意:不能复用 `c.doGet` —— 后者会自动注入 `suite_access_token`,与 provider 作用域冲突。直接调用底层 `doRequestRaw`。

## 3. DTO 设计(`struct.oauth2.go`)

### 3.1 UserInfo3rdResp

```go
// UserInfo3rdResp 是 service/auth/getuserinfo3rd 的响应。
// 企业成员 / 非企业成员 / 外部联系人返回的字段不同,全部用 omitempty 填进同一个结构体。
type UserInfo3rdResp struct {
    CorpID         string `json:"CorpId"`
    UserID         string `json:"UserId"`          // 企业成员
    DeviceID       string `json:"DeviceId"`
    UserTicket     string `json:"user_ticket"`     // 企业成员才返回,用于后续换详情
    ExpiresIn      int    `json:"expires_in"`      // user_ticket 有效期(秒)
    OpenUserID     string `json:"open_userid"`     // 跨服务商匿名 id
    OpenID         string `json:"OpenId"`          // 非企业成员时的微信 openid
    ExternalUserID string `json:"external_userid"` // 外部联系人
}
```

官方字段名大小写混用(`CorpId` / `UserId` / `DeviceId` / `OpenId` 为首字母大写,其它为下划线),必须原样映射。

### 3.2 UserDetail3rdResp

```go
// UserDetail3rdResp 是 service/auth/getuserdetail3rd 的响应。
// 注意:此接口对敏感字段有调用者备案要求,调用前请确认合规。
type UserDetail3rdResp struct {
    CorpID  string `json:"corpid"`
    UserID  string `json:"userid"`
    Gender  string `json:"gender"`   // 1 男 / 2 女
    Avatar  string `json:"avatar"`
    QRCode  string `json:"qr_code"`
    Mobile  string `json:"mobile"`
    Email   string `json:"email"`
    BizMail string `json:"biz_mail"`
    Address string `json:"address"`
}
```

## 4. 实现细节

### 4.1 OAuth2URL

```go
func (c *Client) OAuth2URL(redirectURI, state string, opts ...OAuth2Option) string {
    p := &oauth2Params{scope: "snsapi_privateinfo"}
    for _, opt := range opts {
        opt(p)
    }
    q := url.Values{}
    q.Set("appid", c.cfg.SuiteID)
    q.Set("redirect_uri", redirectURI)
    q.Set("response_type", "code")
    q.Set("scope", p.scope)
    q.Set("state", state)
    if p.hasAgent {
        q.Set("agentid", strconv.Itoa(p.agentID))
    }
    return "https://open.weixin.qq.com/connect/oauth2/authorize?" + q.Encode() + "#wechat_redirect"
}
```

**质量要点:**
- `url.Values.Encode()` 保证 key 按字典序排列(便于测试断言)。
- `#wechat_redirect` 片段必须放在 query 之后。
- `redirect_uri` 由 `url.Values.Encode()` 自动百分号编码。

### 4.2 GetUserInfo3rd

```go
func (c *Client) GetUserInfo3rd(ctx context.Context, authCode string) (*UserInfo3rdResp, error) {
    extra := url.Values{"code": {authCode}}
    var resp UserInfo3rdResp
    if err := c.providerDoGet(ctx, "/cgi-bin/service/auth/getuserinfo3rd", extra, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.3 GetUserDetail3rd

```go
func (c *Client) GetUserDetail3rd(ctx context.Context, userTicket string) (*UserDetail3rdResp, error) {
    body := map[string]string{"user_ticket": userTicket}
    var resp UserDetail3rdResp
    if err := c.providerDoPost(ctx, "/cgi-bin/service/auth/getuserdetail3rd", body, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

## 5. 测试策略(`oauth2_test.go`)

~5 cases,httptest 驱动(URL 构造 case 不需要服务器),沿用 `newTestISVClientWithProvider`。

| # | Case | 断言 |
|---|---|---|
| 1 | `TestOAuth2URL_Default` | 解析返回 URL,断言 host、path、appid/state/response_type/scope 默认 `snsapi_privateinfo`、fragment `wechat_redirect`、redirect_uri 被编码 |
| 2 | `TestOAuth2URL_WithOptions` | `WithOAuth2Scope("snsapi_base") + WithOAuth2AgentID(1000001)` 覆盖 scope 并添加 agentid |
| 3 | `TestGetUserInfo3rd_Member` | mock 企业成员响应:UserID / UserTicket / OpenUserID 全部映射,断言 query 含 `code=AUTH1` 和 `provider_access_token=PTOK` |
| 4 | `TestGetUserInfo3rd_NonMember` | mock 非企业成员响应:只返回 `OpenId`,断言 UserID 为空、OpenID 非空 |
| 5 | `TestGetUserDetail3rd` | POST body 字段 `user_ticket="TICKET"`,响应字段映射到 `Mobile` / `Email` / `BizMail` |

覆盖率目标 ≥85%。

## 6. 错误处理

- `OAuth2URL` 不返回 error —— 调用方传入的 `redirectURI` 可能已经被 URL-encode 过一次,库不做二次判断,交给 WeChat 返回错误。
- HTTP 方法沿用 `providerDoPost` / `providerDoGet` 的 `decodeRaw` 两阶段解码 → `WeixinError`。
- 未配置 provider 凭据时透传 `ErrProviderCorpIDMissing` / `ErrProviderSecretMissing`。

## 7. 交付规模估计

| 文件 | 生产行数 | 测试行数 |
|---|---|---|
| `oauth2.go` | ~75 | — |
| `struct.oauth2.go` | ~30 | — |
| `oauth2_test.go` | — | ~200 |
| `provider.id_convert.go`(追加 `providerDoGet`) | +12 | — |
| **合计** | **~117** | **~200** |

## 8. Commit 节奏

~4 个原子 commit:

1. DTO:`struct.oauth2.go`
2. `providerDoGet` helper(+ 单测复用现有 `provider.id_convert_test.go` 的 mock 风格)
3. `OAuth2URL` + `OAuth2Option` + 2 个单元测试
4. `GetUserInfo3rd` + `GetUserDetail3rd` + 3 个 httptest 用例(最后一步跑全量 `-race`、`-cover`、`./...`)

## 9. Self-Review Checklist(plan 阶段勾选)

- [ ] 3 个公开方法全部实现,外加一个私有 `providerDoGet` helper
- [ ] DTO json tag 与官方文档字段名大小写一致(`CorpId` / `UserId` 等)
- [ ] OAuth2URL 默认 scope = `snsapi_privateinfo`
- [ ] OAuth2URL 使用 `url.Values.Encode()`,保证测试可断言
- [ ] 未新增哨兵错误
- [ ] 未修改 Config / Store / Client 结构
- [ ] `go test -race ./work-wechat/isv/...` 通过
- [ ] 覆盖率 ≥85%

---

**spec 完成。**
