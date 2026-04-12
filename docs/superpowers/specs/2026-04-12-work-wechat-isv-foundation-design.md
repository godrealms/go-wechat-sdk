# work-wechat ISV 认证底座 — 设计文档

**Date:** 2026-04-12
**Status:** Draft — awaiting user review
**Scope:** 企业微信(work-wechat)第一轮 —— 服务商(ISV)生态子项目 1:认证底座 + 消息回调解密

## 1. 背景

`work-wechat/` 模块当前只有一个空 stub。企业微信(qyapi.weixin.qq.com)作为独立于公众号/小程序的生态,拥有 150+ 个官方 API,按业务域可分解为 9 个子项目(见 §11)。本轮只做**服务商(ISV)生态**中最基础的子项目 1 —— 认证底座 + 消息回调解密。

这对应 `oplatform` 的子项目 1+2+6(代公众号/代小程序授权底座),两者在模式上高度镜像。本文档的大部分架构决策都直接沿用了 `oplatform` 沉淀下来的范式。

### 1.1 目标

在新的 `work-wechat/isv` 子包中提供 **16 个公开方法**,覆盖:

- **suite 凭证生命周期**(2):`GetSuiteAccessToken` / `RefreshSuiteToken`
- **预授权与授权 URL**(3):`GetPreAuthCode` / `SetSessionInfo` / `AuthorizeURL`
- **永久授权与授权信息**(3):`GetPermanentCode` / `GetAuthInfo` / `GetAdminList`
- **corp_token 生命周期**(4):`GetCorpToken` / `CorpClient` 工厂 / `(*CorpClient).AccessToken` / `(*CorpClient).Refresh`
- **运维与下游注入**(1):`(*Client).RefreshAll`
- **回调事件解密与分发**(1):`ParseNotify`(9 种强类型事件 + 1 种 RawEvent 兜底)
- **ID 转换**(2):`CorpIDToOpenCorpID` / `UserIDToOpenUserID`

### 1.2 显式不做

- 代企业调用通讯录/应用/消息(子项目 3+4)
- 服务商自身管理界面登录(子项目 2,但本轮已经包含了 ID 转换这两个 provider 接口)
- 代开发自建应用(子项目 5)
- 家校沟通 / 政民通 / 会议室 / 微盘 等垂直行业(独立子项目)
- 自建企业自用场景(非 ISV,将来单独做 `work-wechat/enterprise` 子包)

## 2. 架构

### 2.1 定位

`work-wechat/isv` 是企业微信第三方应用服务商的授权底座。它管理:

- 服务商级别的 `suite_access_token`(由 suite_id + suite_secret + suite_ticket 换取)
- 每个被授权企业的 `permanent_code` + `corp_access_token`
- 服务商级别的 `provider_access_token`(由 ProviderCorpID + ProviderSecret 换取,用于 ID 转换接口)
- 所有授权事件回调的解密与分发

它的输出是:
- 一组用于授权流程的方法(拉 pre_auth_code → 生成授权 URL → 换 permanent_code)
- 一组用于代企业调用的 `*CorpClient`,实现 `isv.TokenSource` 接口,供**未来**的下游子项目(通讯录/应用/消息)注入
- 一个 `ParseNotify` 入口,消费回调并自动持久化 `suite_ticket`

### 2.2 依赖与复用

- `utils/wxcrypto` —— 消息加解密。企业微信和公众号用的是同一套 Biz Msg Crypt(AES-256-CBC + PKCS#7 + SHA1 签名),可以直接复用。
- `oplatform` 的架构范式 —— 全部照搬:
  - `Config` + `Client` + `Option` 函数式配置
  - `Store` 接口 + `MemoryStore` 默认实现
  - `TokenSource` 注入抽象
  - Lazy + 显式 `RefreshAll` 的 token 策略
  - `doPost` / `doGet` / `doPostRaw` 共享助手 + 两阶段 JSON 解码
  - `httptest` 驱动的单元测试 + `newTestISVClient(t, baseURL)` helper
  - `ParseNotify` 强类型事件分发

**零新基础设施** —— 所有模式都从 oplatform 搬过来,调用方和维护方的学习成本都是零。

### 2.3 包结构

**顶层选择**:`work-wechat/` 下按家族分子包(与 oplatform 的单包平铺不同)。

**理由**:企业微信最终会有 9 个子项目(§11),家族之间耦合很低。如果全部平铺到 `work-wechat/` 单包会导致 60+ 文件和命名前缀拥挤(`ISVClient` / `ContactClient` / `ExternalContactClient` ...)。分子包后,每个家族独立 import path,互不污染。

第一轮只建 `work-wechat/isv`,import path 为 `github.com/godrealms/go-wechat-sdk/work-wechat/isv`。

## 3. 文件布局

```
work-wechat/
  isv/
    doc.go                      [NEW] 包文档
    errors.go                   [NEW] WeixinError + 哨兵错误
    store.go                    [NEW] Store 接口 + MemoryStore
    store_test.go               [NEW]
    tokensource.go              [NEW] TokenSource 接口
    client.go                   [NEW] Config / Client / Options / 共享 HTTP 助手
    client_test.go              [NEW]
    suite.token.go              [NEW] GetSuiteAccessToken / RefreshSuiteToken
    suite.token_test.go         [NEW]
    suite.preauth.go            [NEW] GetPreAuthCode / SetSessionInfo / AuthorizeURL
    suite.preauth_test.go       [NEW]
    suite.permanent.go          [NEW] GetPermanentCode / GetAuthInfo / GetAdminList / ID 转换 2 个
    suite.permanent_test.go     [NEW]
    corp.token.go               [NEW] GetCorpToken / CorpClient / AccessToken / Refresh / RefreshAll
    corp.token_test.go          [NEW]
    authorizer.go               [NEW] CorpClient 编译期断言为 TokenSource
    notify.go                   [NEW] ParseNotify + 9 种事件 + RawEvent
    notify_test.go              [NEW]
    struct.suite.go             [NEW] suite_token / pre_auth_code / session_info DTO
    struct.permanent.go         [NEW] permanent_code / auth_info / admin_list / ID 转换 DTO
    struct.corp.go              [NEW] corp_token DTO
    struct.notify.go            [NEW] 9 种 InfoType 事件结构体 + Event 接口
    example/
      main.go                   [NEW] 编译级 demo
```

每个家族一对 `.go` + `_test.go`,DTO 按家族拆成 4 个 `struct.*.go`(而不是一个大杂烩),这样未来做子项目 2+3 时追加 DTO 不会让某个文件膨胀。

## 4. 核心类型

### 4.1 Config

```go
type Config struct {
    SuiteID        string // 第三方应用 suite_id
    SuiteSecret    string // 第三方应用 suite_secret
    ProviderCorpID string // 服务商自己的 corpid(provider 接口需要,可选)
    ProviderSecret string // 服务商 provider_secret(provider 接口需要,可选)
    Token          string // 回调 token
    EncodingAESKey string // 回调 AES key(43 字符)
}
```

**校验规则**(`NewClient` 里执行):
- `SuiteID` / `SuiteSecret` / `Token` / `EncodingAESKey` 必填
- `EncodingAESKey` 长度必须 43
- `ProviderCorpID` 和 `ProviderSecret` **要么都填,要么都空**。都空时 ID 转换接口返回 `ErrProviderSecretMissing` / `ErrProviderCorpIDMissing`。

### 4.2 Client

```go
type Client struct {
    cfg        Config
    store      Store
    http       *http.Client
    crypto     *wxcrypto.MsgCrypto
    baseURL    string     // 默认 "https://qyapi.weixin.qq.com",WithBaseURL 可注入 httptest URL

    suiteMu    sync.Mutex // suite_access_token 刷新单飞
    providerMu sync.Mutex // provider_access_token 刷新单飞
    corpMu     sync.Map   // map[corpid]*sync.Mutex,corp_token 按 corpid 单飞
}

type Option func(*Client)

func WithStore(s Store) Option
func WithHTTPClient(h *http.Client) Option
func WithBaseURL(u string) Option

func NewClient(cfg Config, opts ...Option) (*Client, error)
```

`Client` 无状态可共享,所有状态都在 Store 中。无后台 goroutine / 析构函数。

### 4.3 Store

```go
type Store interface {
    // suite_ticket(回调推送,必须持久化)
    GetSuiteTicket(ctx context.Context, suiteID string) (string, error)
    PutSuiteTicket(ctx context.Context, suiteID, ticket string) error

    // suite_access_token(缓存)
    GetSuiteToken(ctx context.Context, suiteID string) (token string, expiresAt time.Time, err error)
    PutSuiteToken(ctx context.Context, suiteID, token string, expiresAt time.Time) error

    // provider_access_token(缓存)
    GetProviderToken(ctx context.Context, suiteID string) (token string, expiresAt time.Time, err error)
    PutProviderToken(ctx context.Context, suiteID, token string, expiresAt time.Time) error

    // 企业授权信息(permanent_code + corp_token + 过期时间)
    GetAuthorizer(ctx context.Context, suiteID, corpID string) (*AuthorizerTokens, error)
    PutAuthorizer(ctx context.Context, suiteID, corpID string, tokens *AuthorizerTokens) error
    DeleteAuthorizer(ctx context.Context, suiteID, corpID string) error
    ListAuthorizers(ctx context.Context, suiteID string) ([]string, error)
}

type AuthorizerTokens struct {
    CorpID            string
    PermanentCode     string
    CorpAccessToken   string
    CorpTokenExpireAt time.Time
}
```

**设计要点**
- 所有方法的第一 key 都是 `suiteID`,允许同一体系内跑多个 Client 共享同一 Store,或未来做多服务商托管
- `AuthorizerTokens` 把 permanent_code + corp_token + 过期时间打包到一个实体,换 token 时只做一次 `Put`
- `GetAuthorizer` 在 key 不存在时返回 `ErrNotFound`,Client 据此决定是否初始化

**默认实现 `MemoryStore`** 用 `sync.Map` + `sync.RWMutex` 保证线程安全。

### 4.4 TokenSource 与 CorpClient

```go
// tokensource.go
type TokenSource interface {
    AccessToken(ctx context.Context) (string, error)
}

// corp.token.go
type CorpClient struct {
    parent *Client
    corpID string
}

func (c *Client) CorpClient(corpID string) *CorpClient
func (cc *CorpClient) AccessToken(ctx context.Context) (string, error) // TokenSource
func (cc *CorpClient) Refresh(ctx context.Context) error
func (c *Client) RefreshAll(ctx context.Context) error

// authorizer.go
var _ TokenSource = (*CorpClient)(nil) // 编译期断言
```

`CorpClient.AccessToken` 实现 **lazy + 双检锁**:
1. 从 Store 读 `AuthorizerTokens`,未过期直接返回
2. 获取 `corpMu` 里对应 corpid 的锁
3. 再从 Store 读一次(可能已被其它 goroutine 刷新)
4. 仍过期则调 `service/get_corp_token`,写回 Store

`RefreshAll` 遍历 `Store.ListAuthorizers`,逐个调 `CorpClient.Refresh`。任意一个失败都继续下一个,最后聚合错误返回(`errors.Join`)。

### 4.5 错误

```go
var (
    ErrNotFound               = errors.New("isv: not found")
    ErrSuiteTicketMissing     = errors.New("isv: suite_ticket missing in store")
    ErrProviderCorpIDMissing  = errors.New("isv: provider corpid not configured")
    ErrProviderSecretMissing  = errors.New("isv: provider secret not configured")
    ErrAuthorizerRevoked      = errors.New("isv: authorizer revoked")
)

type WeixinError struct {
    ErrCode int    `json:"errcode"`
    ErrMsg  string `json:"errmsg"`
}

func (e *WeixinError) Error() string { return fmt.Sprintf("isv: weixin error %d: %s", e.ErrCode, e.ErrMsg) }
```

业务错误(`errcode != 0`)包装为 `*WeixinError`,调用方用 `errors.As` 判断。HTTP/网络错误从 `http.Client.Do` 透传。

## 5. HTTP 助手

`client.go` 内部提供三个助手,所有业务方法调用它们而不直接操作 `net/http`。

```go
// doPost: 以 suite_access_token 作为 query 参数发 POST JSON
func (c *Client) doPost(ctx context.Context, path string, body, out interface{}) error

// doGet: 以 suite_access_token 作为 query 参数发 GET
func (c *Client) doGet(ctx context.Context, path string, query url.Values, out interface{}) error

// doPostRaw: 不自动附 suite_token(用于需要特殊 query 的请求)
func (c *Client) doPostRaw(ctx context.Context, path string, query url.Values, body, out interface{}) error
```

**实现细节**
- `baseURL + path` 拼 URL
- `doPost` / `doGet` 自动注入 `access_token=<suite_token>` 到 query(`get_suite_token` 本身用 `doPostRaw`,因为它是用来获取 token 的)
- body 序列化为 `application/json; charset=utf-8`
- **两阶段 JSON 解码**:
  1. 先解进 `map[string]json.RawMessage`,取 `errcode`
  2. 非零返回 `*WeixinError{ErrCode, ErrMsg}`
  3. 为零再用 `json.Unmarshal(raw, out)` 解进 `out`
- HTTP 状态码非 2xx 返回包装错误

**CorpClient 版本**:`corp.token.go` 里有 `CorpClient.doPost` / `doGet` 注入 `corp_access_token`,第一轮内部使用,为下游子项目铺路。

## 6. 16 个公开方法

### 6.1 suite_token(2)

```go
// 返回 suite_access_token(lazy + 缓存命中直接返回,否则双检锁刷新)
func (c *Client) GetSuiteAccessToken(ctx context.Context) (string, error)

// 强制刷新 suite_access_token,写回 Store
func (c *Client) RefreshSuiteToken(ctx context.Context) error
```

**endpoint**:`POST /cgi-bin/service/get_suite_token`
**body**:`{"suite_id":"...","suite_secret":"...","suite_ticket":"..."}`
**response**:`{"suite_access_token":"...","expires_in":7200}`
**缓存过期时间**:`expires_in - 300s`,写回 `Store.PutSuiteToken`

若 Store 没有 suite_ticket,返回 `ErrSuiteTicketMissing`。

### 6.2 pre_auth_code & session(3,其中 1 个无 HTTP)

```go
func (c *Client) GetPreAuthCode(ctx context.Context) (*PreAuthCodeResp, error)
func (c *Client) SetSessionInfo(ctx context.Context, preAuthCode string, info *SessionInfo) error
func (c *Client) AuthorizeURL(preAuthCode, redirectURI, state string) string
```

```go
type PreAuthCodeResp struct {
    PreAuthCode string `json:"pre_auth_code"`
    ExpiresIn   int    `json:"expires_in"`
}

type SessionInfo struct {
    AppID    []int `json:"appid,omitempty"`    // 限定授权的应用 ID 列表
    AuthType int   `json:"auth_type,omitempty"`// 0=管理员授权,1=成员授权
}
```

**`AuthorizeURL` 拼接**:
```
https://open.work.weixin.qq.com/3rdapp/install?suite_id=<>&pre_auth_code=<>&redirect_uri=<>&state=<>
```

所有参数 url.QueryEscape。

**endpoints**
- `POST /cgi-bin/service/get_pre_auth_code`(GET 也可,官方文档里是 GET;这里统一用 GET 减少分支)
- `POST /cgi-bin/service/set_session_info`

注:`get_pre_auth_code` 是 GET 方法,`set_session_info` 是 POST,需要在实现时区分。

### 6.3 永久授权与授权信息(3)

```go
// auth_code → 永久授权码 + 首次 corp_token + 授权信息
// 自动持久化 AuthorizerTokens 到 Store
func (c *Client) GetPermanentCode(ctx context.Context, authCode string) (*PermanentCodeResp, error)

// 查询企业授权信息(不缓存,每次拉新)
func (c *Client) GetAuthInfo(ctx context.Context, corpID, permanentCode string) (*AuthInfoResp, error)

// 获取应用管理员列表
func (c *Client) GetAdminList(ctx context.Context, corpID, agentID string) (*AdminListResp, error)
```

**endpoints**
- `POST /cgi-bin/service/get_permanent_code`
- `POST /cgi-bin/service/get_auth_info`
- `POST /cgi-bin/service/get_admin_list`

**DTO 骨架**(完整嵌套结构按官方文档补齐):
```go
type PermanentCodeResp struct {
    AccessToken   string        `json:"access_token"`
    ExpiresIn     int           `json:"expires_in"`
    PermanentCode string        `json:"permanent_code"`
    AuthCorpInfo  AuthCorpInfo  `json:"auth_corp_info"`
    AuthInfo      AuthInfoAgent `json:"auth_info"`
    AuthUserInfo  AuthUserInfo  `json:"auth_user_info"`
}

type AuthCorpInfo struct {
    CorpID           string `json:"corpid"`
    CorpName         string `json:"corp_name"`
    CorpType         string `json:"corp_type"`
    CorpSquareLogoURL string `json:"corp_square_logo_url"`
    CorpUserMax      int    `json:"corp_user_max"`
    CorpFullName     string `json:"corp_full_name"`
    VerifiedEndTime  int64  `json:"verified_end_time"`
    SubjectType      int    `json:"subject_type"`
    CorpWxqrcode     string `json:"corp_wxqrcode"`
    CorpScale        string `json:"corp_scale"`
    CorpIndustry     string `json:"corp_industry"`
    CorpSubIndustry  string `json:"corp_sub_industry"`
    Location         string `json:"location"`
}

type AuthInfoAgent struct {
    Agent []AuthAgent `json:"agent"`
}

type AuthAgent struct {
    AgentID       int    `json:"agentid"`
    Name          string `json:"name"`
    RoundLogoURL  string `json:"round_logo_url"`
    SquareLogoURL string `json:"square_logo_url"`
    AppID         int    `json:"appid"`
    Privilege     AgentPrivilege `json:"privilege,omitempty"`
    SharedFrom    *SharedFromInfo `json:"shared_from,omitempty"`
}

type AgentPrivilege struct {
    Level      int      `json:"level"`
    AllowParty []int    `json:"allow_party,omitempty"`
    AllowUser  []string `json:"allow_user,omitempty"`
    AllowTag   []int    `json:"allow_tag,omitempty"`
    ExtraParty []int    `json:"extra_party,omitempty"`
    ExtraUser  []string `json:"extra_user,omitempty"`
    ExtraTag   []int    `json:"extra_tag,omitempty"`
}

type SharedFromInfo struct {
    CorpID string `json:"corpid"`
    ShareType int `json:"share_type"`
}

type AuthUserInfo struct {
    UserID     string `json:"userid"`
    OpenUserID string `json:"open_userid"`
    Name       string `json:"name"`
    Avatar     string `json:"avatar"`
}

type AuthInfoResp struct {
    AuthCorpInfo AuthCorpInfo  `json:"auth_corp_info"`
    AuthInfo     AuthInfoAgent `json:"auth_info"`
}

type AdminListResp struct {
    Admin []AdminInfo `json:"admin"`
}

type AdminInfo struct {
    UserID      string `json:"userid"`
    OpenUserID  string `json:"open_userid"`
    AuthType    int    `json:"auth_type"` // 0=普通管理员 1=超级管理员
}
```

### 6.4 corp_token(4)

```go
// 底层接口(CorpClient 内部会调用它,也对外暴露供手动使用)
func (c *Client) GetCorpToken(ctx context.Context, corpID, permanentCode string) (*CorpTokenResp, error)

// CorpClient 工厂
func (c *Client) CorpClient(corpID string) *CorpClient

// 实现 TokenSource,lazy + 双检锁,写回 Store
func (cc *CorpClient) AccessToken(ctx context.Context) (string, error)

// 强制刷新单个企业的 corp_token
func (cc *CorpClient) Refresh(ctx context.Context) error
```

```go
type CorpTokenResp struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}
```

**endpoint**:`POST /cgi-bin/service/get_corp_token`
**body**:`{"auth_corpid":"...","permanent_code":"..."}`

### 6.5 运维(1)

```go
// 遍历 Store.ListAuthorizers,逐个调 CorpClient.Refresh
// 任意失败继续下一个,聚合错误用 errors.Join
func (c *Client) RefreshAll(ctx context.Context) error
```

### 6.6 ID 转换(2,provider 体系)

```go
// corpid → open_corpid
func (c *Client) CorpIDToOpenCorpID(ctx context.Context, corpID string) (string, error)

// 批量 userid → open_userid
func (c *Client) UserIDToOpenUserID(ctx context.Context, corpID string, userIDs []string) (*UserIDConvertResp, error)
```

```go
type UserIDConvertResp struct {
    OpenUserIDList    []UserIDOpenUserIDPair `json:"open_userid_list"`
    InvalidUserIDList []string               `json:"invalid_userid_list"`
}

type UserIDOpenUserIDPair struct {
    UserID     string `json:"userid"`
    OpenUserID string `json:"open_userid"`
}
```

**endpoints**(用 provider_access_token)
- `POST /cgi-bin/service/corpid_to_opencorpid`
- `POST /cgi-bin/service/batch/userid_to_openuserid`

**provider_access_token 的获取**:内部 `getProviderAccessToken(ctx)` 函数,lazy + `providerMu` 单飞,逻辑同 suite_token。官方 endpoint:`POST /cgi-bin/service/get_provider_token`,body `{"corpid":"<ProviderCorpID>","provider_secret":"<ProviderSecret>"}`。

**前置校验**:调用这 2 个方法前,若 `cfg.ProviderCorpID` 或 `cfg.ProviderSecret` 为空,立即返回 `ErrProviderCorpIDMissing` / `ErrProviderSecretMissing`。

### 6.7 ParseNotify(1 个公开入口)

```go
// 接收回调 HTTP 请求,解密并返回强类型事件
// 对 suite_ticket 事件自动持久化到 Store
func (c *Client) ParseNotify(r *http.Request) (Event, error)
```

**流程**
1. 读 query:`msg_signature` / `timestamp` / `nonce`
2. 读 body,XML unmarshal 到 `componentEnvelope{ToUserName, Encrypt}`
3. `wxcrypto.MsgCrypto.VerifySignature(token, timestamp, nonce, encrypt, msgSignature)` → 失败返回签名错误
4. `wxcrypto.MsgCrypto.Decrypt(encrypt)` → 解出 innerXML
5. XML unmarshal 到 `componentInner{InfoType, SuiteID, ...}`
6. 按 `InfoType` 分发到具体事件类型
7. 若 `InfoType == "suite_ticket"`,在返回前调用 `c.store.PutSuiteTicket(ctx, suiteID, ticket)`
8. 未知 `InfoType` 返回 `*RawEvent{InfoType, RawXML}`,不是错误

### 6.8 9 种事件类型

```go
// struct.notify.go

type Event interface{ isEvent() }

type baseEvent struct {
    SuiteID   string
    ReceiveAt time.Time
}

func (baseEvent) isEvent() {}

type SuiteTicketEvent struct {
    baseEvent
    SuiteTicket string
}

type CreateAuthEvent struct {
    baseEvent
    AuthCode string
}

type ChangeAuthEvent struct {
    baseEvent
    AuthCorpID string
}

type CancelAuthEvent struct {
    baseEvent
    AuthCorpID string
}

type ResetPermanentCodeEvent struct {
    baseEvent
    AuthCorpID string
}

type ChangeContactEvent struct {
    baseEvent
    AuthCorpID string
    ChangeType string // create_user / update_user / delete_user / create_party / update_party / delete_party / update_tag
    UserID     string
    Name       string
    Department string
    NewUserID  string // update_user 时的 userid 变更
}

type ChangeExternalContactEvent struct {
    baseEvent
    AuthCorpID     string
    ChangeType     string
    UserID         string
    ExternalUserID string
}

type ShareAgentChangeEvent struct {
    baseEvent
    AuthCorpID string
    AgentID    string
}

type ChangeAppAdminEvent struct {
    baseEvent
    AuthCorpID string
    UserID     string
    IsAdmin    bool
}

// 未知 InfoType 兜底
type RawEvent struct {
    baseEvent
    InfoType string
    RawXML   string
}
```

## 7. 并发与生命周期

- `Client` 无状态,可在多 goroutine 共享
- 三把锁:
  - `suiteMu` — suite_token 刷新单飞
  - `providerMu` — provider_token 刷新单飞
  - `corpMu sync.Map` — 每个 corpid 一把独立 `*sync.Mutex`,`corp_token` 刷新按 corpid 单飞
- 所有方法接受 `ctx context.Context`,可取消/超时
- Store 的实现方负责自己的线程安全,`MemoryStore` 用 `sync.RWMutex`
- 无后台 goroutine / 定时器 / 析构函数,`Client` 被 GC 不会泄露资源

## 8. 测试策略

沿用 `oplatform` 的 `httptest` 模式。

### 8.1 Helper

```go
// testing_helpers_test.go(仅测试包内)

func newTestISVClient(t *testing.T, baseURL string) *Client {
    // 建 MemoryStore,预种子 suite_ticket = "TICKET"
    // 构造 Client 指向 baseURL
    // 返回 Client
}

func newTestISVClientWithProvider(t *testing.T, baseURL string) *Client {
    // 同上 + ProviderCorpID + ProviderSecret
}
```

### 8.2 用例矩阵

| 测试文件 | 覆盖 | 用例数 |
|---|---|---|
| `store_test.go` | MemoryStore CRUD + 过期检测 + ListAuthorizers | 8 |
| `client_test.go` | NewClient 参数校验 + Options 应用 + 哨兵错误 | 4 |
| `suite.token_test.go` | GetSuiteAccessToken 首次/缓存命中/过期 / RefreshSuiteToken / ErrSuiteTicketMissing / errcode 42001 | 5 |
| `suite.preauth_test.go` | GetPreAuthCode / SetSessionInfo / AuthorizeURL 拼接 | 3 |
| `suite.permanent_test.go` | GetPermanentCode(+ 验证 Store 写入) / GetAuthInfo / GetAdminList / CorpIDToOpenCorpID / UserIDToOpenUserID / ErrProviderSecretMissing | 6 |
| `corp.token_test.go` | GetCorpToken / CorpClient.AccessToken 首次+命中+过期 / Refresh / RefreshAll(2 企业) / 并发单飞(多 goroutine 只触发 1 次 HTTP) | 6 |
| `notify_test.go` | 9 种 InfoType happy path + 1 签名失败 + 1 未知 InfoType → RawEvent + suite_ticket 自动持久化 | 12 |

**合计约 44 用例**。

### 8.3 并发测试要点

`corp.token_test.go` 中的并发单飞测试:

```go
// 启 10 个 goroutine 并发调 CorpClient.AccessToken
// httptest handler 计数请求次数
// 断言:10 次 Go 调用 → 只有 1 次 HTTP 请求到 get_corp_token
```

这个测试确保双检锁正确性,避免 thundering herd。

## 9. 错误处理(汇总)

**三类错误分层**:
1. HTTP/网络错误 — 从 `http.Client.Do` 透传,包装成 `fmt.Errorf("isv: http: %w", err)`
2. 业务错误(`errcode != 0`)— `*WeixinError`
3. 哨兵错误 — `ErrNotFound` / `ErrSuiteTicketMissing` / `ErrProviderCorpIDMissing` / `ErrProviderSecretMissing` / `ErrAuthorizerRevoked`

**不做 errcode 分支**(和 oplatform 一致)。调用方有特判需求自行 `errors.As(&we)`。

## 10. 兼容性

纯 additive:
- 新增 `work-wechat/isv` 子包
- 不修改任何现有模块的公开 API
- `utils/wxcrypto` 已是稳定公共包,零修改
- 零 breaking change

## 11. 非目标 / 后续工作

企业微信的完整路线图(本轮只做子项目 1):

| # | 子项目 | 状态 |
|---|---|---|
| **1** | **ISV 认证底座** | **本轮** |
| 2 | 服务商自身接口(provider 登录等) | 待做 |
| 3 | 代企业通讯录管理 | 待做 |
| 4 | 代企业应用管理 & 消息发送 | 待做 |
| 5 | 代开发自建应用 | 待做 |
| 6 | 第三方应用事件回调扩展(超出 9 种) | 待做 |
| 7 | 服务商小程序 / 网页登录 | 待做 |
| 8 | 自建企业自用(非 ISV) | 待做 |
| 9 | 家校沟通 / 政民通 / 会议室等垂直行业 | 独立,非 ISV 核心 |

注:ID 转换(corpid_to_opencorpid / userid_to_openuserid)原本归属子项目 2,本轮因调用方主动要求提前纳入,对应 Config 新增 `ProviderCorpID` + `ProviderSecret` 两字段。后续子项目 2 会在此基础上追加 `service/get_login_info` 等服务商登录接口。

## 12. 交付物规模估计

| 文件 | 生产行数 | 测试行数 |
|---|---|---|
| `doc.go` | 20 | — |
| `errors.go` | 30 | — |
| `store.go` | 150 | 180 |
| `tokensource.go` | 10 | — |
| `client.go` | 130 | 80 |
| `suite.token.go` | 100 | 150 |
| `suite.preauth.go` | 90 | 90 |
| `suite.permanent.go` | 230 | 200 |
| `corp.token.go` | 200 | 180 |
| `authorizer.go` | 20 | — |
| `notify.go` | 260 | 280 |
| `struct.suite.go` | 40 | — |
| `struct.permanent.go` | 150 | — |
| `struct.corp.go` | 20 | — |
| `struct.notify.go` | 100 | — |
| `example/main.go` | 80 | — |

**生产代码 ~1630 行,测试 ~1160 行,合计 ~2790 行**。

## 13. 预期 commit 节奏

约 11 个原子 commit:

1. 脚手架 — `doc.go` / `errors.go` / `store.go` + test / `tokensource.go`
2. `client.go` + test + Options/NewClient
3. suite token 家族
4. pre_auth_code 家族
5. permanent_code 家族核心 3 方法
6. permanent_code 家族 provider ID 转换 2 方法
7. corp token 家族 + authorizer.go
8. notify.go 骨架 + suite_ticket / 授权类 4 事件
9. notify.go 补 5 个变更类事件 + RawEvent 兜底
10. example + README 片段
11. 最终 `go vet` / `go test -race` / 覆盖率检查

## 14. 开放问题

无未决项。所有关键决策已在本设计中锁定:

- 包结构:`work-wechat/isv` 独立子包 ✔
- Store 接口:复刻 oplatform(suite_ticket / suite_token / provider_token / authorizer CRUD)✔
- TokenSource:`CorpClient` 实现 `isv.TokenSource`,编译期断言 ✔
- Token 策略:Lazy + 显式 RefreshAll ✔
- 测试:httptest + 预种子 token helper ✔
- ParseNotify:全量 9 种 InfoType 强类型 + RawEvent 兜底 ✔
- 接口范围:16 个公开方法(含 provider ID 转换 2 个)✔
- Config:6 字段(ProviderCorpID + ProviderSecret 可选)✔
- 兼容:纯 additive ✔
