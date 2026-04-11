# oplatform 授权底座 + 代调用框架 + 扫码登录 — 设计文档

**Date:** 2026-04-12
**Status:** Draft — awaiting user review
**Scope:** 微信开放平台子项目 1（component 授权底座）+ 子项目 2（authorizer 代调用框架）+ 子项目 6（网站扫码登录 QR Login）

## 1. 背景与范围

### 1.1 为什么拆分

微信"开放平台"官方文档包含 200+ 接口，跨 7 个独立子系统，无法在单个 spec / 单次实现中覆盖。本设计将 oplatform 拆为以下子项目：

| # | 子项目 | 依赖 |
|---|---|---|
| **1** | component 授权底座：verify_ticket、component_access_token、pre_auth_code、authorization_code 换 authorizer token、授权事件 | — |
| **2** | authorizer 代调用框架：authorizer_access_token 自动刷新、代调用 HTTP 通路 | 1 |
| 3 | 代小程序·开发管理 | 1, 2 |
| 4 | 代公众号调用接口清单 | 1, 2 |
| 5 | 快速注册小程序 / 公众号 | 1 |
| **6** | 开放平台网站扫码登录（snsapi_login） | — |
| 7 | 代小程序·运营管理 | 1, 2 |

**本轮实现 1 + 2 + 6**。3 / 4 / 5 / 7 留作后续子项目，每个单独 brainstorm → spec → plan。

### 1.2 为什么这三项一起做

- **1 是硬前置**：component_access_token 依赖推送过来的 verify_ticket，没有它一切代调用都无从谈起
- **2 只是在 1 之上加一层薄 wrapper**：通过 TokenSource 注入把 authorizer 身份接入现有 offiaccount / mini-program 的全部 API，打通代调用通路后，后续子项目 3/4/7 等同于纯铺接口
- **6 与 1/2 完全独立**：属于另一条开放平台 OAuth 流程，但体量很小（5 个方法），顺手做掉
- 合起来仍然是一个可独立交付、有测试、有 example 的完整功能，不会留半截工程

### 1.3 显式不做

- 代小程序·开发管理（代码上传 / 审核 / 发布）
- 代小程序·运营管理（插件、订阅消息、客服、广告）
- 快速注册小程序 / 公众号
- 开放平台账号管理（绑定 / 解绑 UnionID 等管理接口）
- 代公众号调用的具体业务接口清单（通路打通后 offiaccount 现有方法自动可用，无需重写）
- `Store` 的 Redis / SQL 实现（接口留给用户）
- SDK 内部后台 goroutine 刷新（维护成本 > 收益）

## 2. 架构与包结构

### 2.1 顶层文件布局

```
oplatform/                         单包，平铺
  client.go                        component Client + Config + 构造
  store.go                         Store 接口 + MemoryStore 实现
  component.token.go               component_access_token (lazy + RefreshComponentToken)
  component.preauth.go             pre_auth_code + 授权跳转 URL
  component.authorize.go           authorization_code 换 authorizer token、授权方信息查询
  component.authorizer.token.go    authorizer_access_token 刷新 (lazy + RefreshAuthorizer/RefreshAll)
  authorizer.go                    AuthorizerClient wrapper: OffiaccountClient()/MiniProgramClient()
  notify.go                        ParseComponentNotify: verify_ticket / authorized / updateauthorized / unauthorized
  qrlogin.go                       开放平台网站扫码登录 OAuth (snsapi_login)
  struct.component.go              component 流程响应结构体
  struct.authorizer.go             authorizer / authorization_info 结构体
  struct.qrlogin.go                qrlogin 响应结构体
  errors.go                        WeixinError + ErrNotFound + ErrAuthorizerRevoked
  client_test.go                   Client 构造与 token 流程测试
  store_test.go                    MemoryStore 并发与语义测试
  component_test.go                component 流程集成测试
  authorizer_test.go               authorizer 刷新 + wrapper 测试
  notify_test.go                   回调解密测试（含 fixture）
  qrlogin_test.go                  qrlogin 测试
  example/
    main.go                        可编译 demo：回调接收 + 扫码登录

utils/wxcrypto/                    新增：从 offiaccount/crypto.go 抽出
  msgcrypt.go                      MsgCrypto、BizMsgVerify、加解密核心
  msgcrypt_test.go                 沿用 offiaccount 已有测试用例
  doc.go                           包 doc

offiaccount/                       改动
  crypto.go                        改为薄代理 + 类型别名（保持现有公开 API 不变）
  client.go                        新增 WithTokenSource Option；AccessTokenE 优先走 TokenSource
  tokensource.go                   新增：TokenSource interface + 构造校验调整
  client_test.go                   新增 TokenSource 注入用例

mini-program/                      改动
  client.go                        新增 WithTokenSource Option；AccessToken 优先走 TokenSource
  tokensource.go                   新增：TokenSource interface + 构造校验调整
  client_test.go                   新增 TokenSource 注入用例
```

### 2.2 基本规则

- 包名 `oplatform`（沿用目录名）
- `NewClient` 不做任何网络 IO（构造不 block）
- 复用 `utils.HTTP` 而非自己造轮子
- 所有外部 API 接收 `context.Context`
- 并发安全：Client 可在多 goroutine 共享
- 日志遵循 `utils.Logger` 默认静默的约定

## 3. Store 持久化契约

### 3.1 为什么必须持久化

- 微信服务器每 ~10 分钟向第三方平台回调 URL 推送 `component_verify_ticket`
- 这是推导 `component_access_token` 的**唯一**种子
- 如果 SDK 重启后内存态丢失，在下一次推送到达前（最多 ~10 分钟）**所有代调用全部瘫痪**
- 同理 `authorizer_refresh_token` 一旦丢失，只能让对方重新扫码授权

### 3.2 接口

```go
package oplatform

// Store 持久化 component_verify_ticket、component_access_token
// 及每个 authorizer 的 refresh_token / access_token。
// SDK 内置 MemoryStore；生产环境应自行实现 Redis/MySQL 版本。
type Store interface {
    // Verify ticket：微信每 ~10min 推一次
    GetVerifyTicket(ctx context.Context) (ticket string, err error)
    SetVerifyTicket(ctx context.Context, ticket string) error

    // component access token（2h TTL）
    GetComponentToken(ctx context.Context) (token string, expireAt time.Time, err error)
    SetComponentToken(ctx context.Context, token string, expireAt time.Time) error

    // authorizer：每个被授权的小程序/公众号一份
    GetAuthorizer(ctx context.Context, appid string) (AuthorizerTokens, error)
    SetAuthorizer(ctx context.Context, appid string, tokens AuthorizerTokens) error
    DeleteAuthorizer(ctx context.Context, appid string) error
    ListAuthorizerAppIDs(ctx context.Context) ([]string, error)
}

type AuthorizerTokens struct {
    AccessToken  string
    RefreshToken string
    ExpireAt     time.Time
}

// ErrNotFound —— Store 实现应返回该哨兵错误表示 key 不存在（非 I/O 错误）。
var ErrNotFound = errors.New("oplatform: not found")
```

### 3.3 MemoryStore

- `sync.RWMutex` + map 实现
- 用于测试和本地开发
- 不持久到磁盘 —— 和内存模型一致
- 完整单元测试：Get/Set/Delete/List 语义 + 并发读写

## 4. Component Client 与授权流程

### 4.1 类型定义

```go
type Config struct {
    ComponentAppID     string
    ComponentAppSecret string
    Token              string // 回调签名 Token
    EncodingAESKey     string // 43字符 AESKey
}

type Client struct {
    cfg    Config
    http   *utils.HTTP
    store  Store
    crypto *wxcrypto.MsgCrypto

    componentMu sync.Mutex                   // 保护 component token 刷新单飞
    authMu      sync.Map                     // map[string]*sync.Mutex，per-appid 刷新锁
}

type Option func(*Client)

func NewClient(cfg Config, opts ...Option) (*Client, error)

func WithStore(s Store) Option               // 默认 MemoryStore
func WithHTTP(h *utils.HTTP) Option          // 测试注入
```

构造时校验：`ComponentAppID`、`ComponentAppSecret`、`Token`、`EncodingAESKey` 非空；`EncodingAESKey` 长度 = 43。

### 4.2 回调解析

```go
type ComponentNotify struct {
    AppID      string    // ComponentAppID
    CreateTime int64
    InfoType   string    // component_verify_ticket / authorized / updateauthorized / unauthorized
    // 以下字段按 InfoType 填充
    ComponentVerifyTicket string
    AuthorizerAppID       string
    AuthorizationCode     string
    AuthorizationCodeExpiredTime int64
    PreAuthCode           string
    Raw                   []byte    // 解密后的原始 XML，供调用方兜底
}

// ParseNotify 处理回调请求：
// - 校验 msg_signature
// - 用 wxcrypto 解密
// - 解析 InfoType
// - 如果是 component_verify_ticket，自动写入 Store
// 成功时返回解析结果；调用方对 4 种 InfoType 各自处理业务逻辑。
func (c *Client) ParseNotify(r *http.Request, body []byte) (*ComponentNotify, error)
```

**关键行为**：`component_verify_ticket` 类型的回调由 SDK 自动 `store.SetVerifyTicket`，调用方零感知。其它三种类型是纯通知，调用方自行决定业务处理（比如在 `unauthorized` 时 `store.DeleteAuthorizer`）。

### 4.3 Token 管理

```go
// Lazy 获取：过期则刷新。
func (c *Client) ComponentAccessToken(ctx context.Context) (string, error)

// 强制刷新（忽略缓存）。
func (c *Client) RefreshComponentToken(ctx context.Context) error
```

实现：
1. 读 `store.GetComponentToken`
2. 未过期（预留 60s）则直接返回
3. 否则加锁（`componentMu`）+ 双重检查
4. 读 `store.GetVerifyTicket`；缺失返回明确错误（提示调用方 Store 未初始化或 ticket 推送未到）
5. 调 `/cgi-bin/component/api_component_token`
6. 成功后 `store.SetComponentToken` 写回，返回

### 4.4 授权引导

```go
// 换 pre_auth_code
func (c *Client) PreAuthCode(ctx context.Context) (string, error)

// 构造 PC / 移动 授权跳转 URL
// authType: 1=公众号 2=小程序 3=全部
// bizAppid: 指定授权方 appid，不传为空字符串
func (c *Client) AuthorizeURL(preAuthCode, redirectURI string, authType int, bizAppid string) string
func (c *Client) MobileAuthorizeURL(preAuthCode, redirectURI string, authType int, bizAppid string) string
```

### 4.5 授权码换 authorizer token

```go
// 用户在微信客户端完成授权后，redirectURI 会带 auth_code 回调。
// 调用方拿 auth_code 调用本方法；SDK 自动写 Store。
func (c *Client) QueryAuth(ctx context.Context, authCode string) (*AuthorizationInfo, error)
```

调用 `/cgi-bin/component/api_query_auth`。成功后：
1. 调用 `store.SetAuthorizer(ctx, authorizerAppID, AuthorizerTokens{...})`
2. 返回解析后的 `AuthorizationInfo`（含 funcInfo 权限集）

### 4.6 授权方信息查询

```go
func (c *Client) GetAuthorizerInfo(ctx context.Context, authorizerAppID string) (*AuthorizerInfo, error)
func (c *Client) GetAuthorizerOption(ctx context.Context, authorizerAppID, optionName string) (*AuthorizerOption, error)
func (c *Client) SetAuthorizerOption(ctx context.Context, authorizerAppID, optionName, optionValue string) error
func (c *Client) GetAuthorizerList(ctx context.Context, offset, count int) (*AuthorizerList, error)
```

对应 `/cgi-bin/component/api_get_authorizer_info` / `api_get_authorizer_option` / `api_set_authorizer_option` / `api_get_authorizer_list`。

## 5. Authorizer 代调用框架

### 5.1 核心想法

在 `offiaccount` 和 `mini_program` 里加一个**可选的** `TokenSource`：

- 默认行为不变（调用方传 `AppSecret`，Client 自己调 `/cgi-bin/token`）
- 如果注入 `TokenSource`，Client 就不再自己刷 token，全部代理给它
- `oplatform.AuthorizerClient` 同时实现 `offiaccount.TokenSource` 和 `mini_program.TokenSource`
- 结果：**同一套 offiaccount / mini_program API 同时支持"自有身份"和"代授权身份"两种角色**，调用方代码零差异

### 5.2 TokenSource 接口（两个包各定义一份）

```go
// offiaccount/tokensource.go
package offiaccount
type TokenSource interface {
    AccessToken(ctx context.Context) (string, error)
}

// mini-program/tokensource.go
package mini_program
type TokenSource interface {
    AccessToken(ctx context.Context) (string, error)
}
```

两个包的 interface 签名相同但类型独立 —— 这样 `oplatform` 可以同时实现两者，而不用引入任一方的依赖。

### 5.3 offiaccount 改动

```go
// client.go 新增字段
type Client struct {
    // ...已有字段...
    tokenSource TokenSource
}

// 新增 Option
func WithTokenSource(ts TokenSource) Option {
    return func(c *Client) { c.tokenSource = ts }
}

// AccessTokenE 改造
func (c *Client) AccessTokenE(ctx context.Context) (string, error) {
    if c.tokenSource != nil {
        return c.tokenSource.AccessToken(ctx)
    }
    // ...原有 /cgi-bin/token 逻辑完全不动...
}
```

**兼容性保证**：
- 调用方不传 `TokenSource`，行为 100% 一致 —— 零破坏
- 构造校验改为：`AppSecret != "" || tokenSource != nil`
- `NewClient` 签名不变；`Option` 变长列表可附加
- 现有的 `GetAccessToken() string` 兼容入口和返回值行为保持

> **注意**：当前 offiaccount `NewClient(ctx, *Config)` 接受 Config 但**没有**可变 Options。本次需要在保持现有调用点能编译的前提下追加 Options：可以通过增加新方法 `NewClientWithOptions(ctx, *Config, ...Option)` 或者给 `NewClient` 附加 `opts ...Option` 可变参数（Go 允许向现有函数尾部追加 variadic）。采用后者更简洁，不会破坏已有调用点。

`mini_program.Client` 同样改造。

### 5.4 oplatform 这边的 AuthorizerClient

```go
type AuthorizerClient struct {
    c     *Client        // 指回 component client
    appID string
}

// 便捷构造
func (c *Client) Authorizer(appID string) *AuthorizerClient {
    return &AuthorizerClient{c: c, appID: appID}
}

// Lazy：读 Store，过期则刷新
func (a *AuthorizerClient) AccessToken(ctx context.Context) (string, error)

// 强制刷新
func (a *AuthorizerClient) Refresh(ctx context.Context) error

// 返回一个预配 TokenSource 的 offiaccount.Client
func (a *AuthorizerClient) OffiaccountClient(opts ...offiaccount.Option) *offiaccount.Client

// 返回一个预配 TokenSource 的 mini_program.Client
func (a *AuthorizerClient) MiniProgramClient(opts ...mini_program.Option) (*mini_program.Client, error)
```

### 5.5 刷新语义

- `AuthorizerClient.AccessToken(ctx)`：
  1. 读 `store.GetAuthorizer(ctx, appID)`
  2. 未过期（预留 60s）直接返回
  3. 否则取 per-appid 锁 + 双重检查
  4. 读 `refresh_token`，调用 `/cgi-bin/component/api_authorizer_token`
  5. 成功：`store.SetAuthorizer` 写回新的 AccessToken/RefreshToken/ExpireAt（微信可能返回新的 refresh_token）
  6. 失败且 `errcode == 61023`：返回 `ErrAuthorizerRevoked`（调用方应清除 Store 并引导重新授权）

- `Client.RefreshAll(ctx)`：遍历 `store.ListAuthorizerAppIDs`，对每个 appid 构造 `AuthorizerClient` 并调用其 `Refresh(ctx)`；用于启动预热或外部 cron 触发。单个 appid 失败不中断整个循环，错误汇总后返回

- SDK 不自动恢复失效的 refresh_token —— 错误透出给调用方

### 5.6 并发与单飞

- component token：`Client.componentMu sync.Mutex` 保护
- authorizer token：`Client.authMu sync.Map`，per-appid 一把锁，避免一个慢 authorizer 阻塞其它

### 5.7 用户代码示例

```go
op, _ := oplatform.NewClient(oplatform.Config{...}, oplatform.WithStore(myRedisStore))

// 某个被授权的公众号
auth := op.Authorizer("wx-biz-appid")
official := auth.OffiaccountClient()
// 之后所有 offiaccount 现有方法直接用，token 自动来自开放平台
official.SetMenu(ctx, menu)
official.SendTemplate(msg)

// 某个被授权的小程序
mp, _ := auth.MiniProgramClient()
mp.SendSubscribeMessage(ctx, body)
```

## 6. QR Login（子项目 6）

和 component 流程完全独立，5 个方法 + 结构体一个文件搞定。

```go
type QRLoginClient struct {
    appID     string
    appSecret string
    http      *utils.HTTP
}

type QRLoginOption func(*QRLoginClient)
func NewQRLoginClient(appID, appSecret string, opts ...QRLoginOption) *QRLoginClient

// 1) 构造网页授权跳转 URL（scope: snsapi_login / snsapi_base / snsapi_userinfo）
func (q *QRLoginClient) AuthorizeURL(redirectURI, scope, state string) string

// 2) code 换 access_token
type QRLoginToken struct {
    AccessToken  string `json:"access_token"`
    ExpiresIn    int64  `json:"expires_in"`
    RefreshToken string `json:"refresh_token"`
    OpenID       string `json:"openid"`
    Scope        string `json:"scope"`
    UnionID      string `json:"unionid,omitempty"`
}
func (q *QRLoginClient) Code2Token(ctx context.Context, code string) (*QRLoginToken, error)

// 3) refresh_token 续期
func (q *QRLoginClient) RefreshToken(ctx context.Context, refreshToken string) (*QRLoginToken, error)

// 4) 获取用户信息（仅 snsapi_userinfo）
type QRLoginUserInfo struct {
    OpenID     string   `json:"openid"`
    Nickname   string   `json:"nickname"`
    Sex        int      `json:"sex"`
    Province   string   `json:"province"`
    City       string   `json:"city"`
    Country    string   `json:"country"`
    HeadImgURL string   `json:"headimgurl"`
    Privilege  []string `json:"privilege"`
    UnionID    string   `json:"unionid,omitempty"`
}
func (q *QRLoginClient) UserInfo(ctx context.Context, accessToken, openID string) (*QRLoginUserInfo, error)

// 5) 检查 access_token 是否有效
func (q *QRLoginClient) Auth(ctx context.Context, accessToken, openID string) error
```

- 所有请求走 `/sns/oauth2/access_token`、`/sns/oauth2/refresh_token`、`/sns/userinfo`、`/sns/auth`
- 无状态，不依赖 Store，不依赖 component Client

## 7. utils/wxcrypto 抽取

### 7.1 目标

消除 offiaccount 和 oplatform 之间的消息加解密代码重复，同时**保持 offiaccount 现有外部 API 不变**。

### 7.2 步骤

1. 新建 `utils/wxcrypto/msgcrypt.go`，把 `offiaccount/crypto.go` 的实现原样挪过来，包名改为 `wxcrypto`
2. 挪 `offiaccount/crypto_test.go` → `utils/wxcrypto/msgcrypt_test.go`
3. `offiaccount/crypto.go` 改为薄代理 + 类型别名，保证外部已有代码不变：

```go
package offiaccount

import "github.com/godrealms/go-wechat-sdk/utils/wxcrypto"

// 类型别名：外部代码中 *offiaccount.MsgCrypto 的用法继续生效
type MsgCrypto = wxcrypto.MsgCrypto

// 构造器转发
func NewMsgCrypto(token, aesKey, appID string) (*MsgCrypto, error) {
    return wxcrypto.New(token, aesKey, appID)
}

// 其它已经导出的顶层函数（如果有）逐一别名/转发
```

4. `offiaccount/crypto_test.go` 保留轻量 smoke test，确认别名链路
5. `oplatform` 直接 `import "github.com/godrealms/go-wechat-sdk/utils/wxcrypto"`

### 7.3 非目标

- 不修改 wxcrypto 核心算法
- 不改变 offiaccount 对外的函数/类型名

## 8. 错误处理

```go
// oplatform/errors.go
type WeixinError struct {
    ErrCode int
    ErrMsg  string
}
func (e *WeixinError) Error() string

var (
    ErrNotFound          = errors.New("oplatform: not found")
    ErrAuthorizerRevoked = errors.New("oplatform: authorizer refresh_token revoked (61023)")
    ErrVerifyTicketMissing = errors.New("oplatform: component_verify_ticket not yet received; wait for weixin push")
)
```

规则：
- 所有微信业务错误（`errcode != 0`）→ `*WeixinError`
- HTTP 错误 → 透传 `*utils.HTTPError`
- Store 错误 → 透传
- 禁止 panic；禁止吞错误

## 9. 测试矩阵

| 层 | 覆盖内容 | 方式 |
|---|---|---|
| `utils/wxcrypto` | 加解密/签名 round-trip、tamper 检测、非法 base64/AES padding | 纯单测（沿用 offiaccount 已有用例） |
| `oplatform.MemoryStore` | Get/Set/Delete/List 语义 + 并发读写 | 纯单测 |
| `oplatform` component token | lazy 拉取、缓存命中、过期重拉、单飞、verify_ticket 缺失报错 | httptest |
| `oplatform` PreAuthCode | 正常路径、errcode 非零 | httptest |
| `oplatform` AuthorizeURL | query 参数顺序/编码 | 字符串断言 |
| `oplatform` QueryAuth | 换 token 后写入 Store、返回结构解析 | httptest + in-mem store |
| `oplatform` authorizer refresh | lazy 刷新、expired、61023 → ErrAuthorizerRevoked、per-appid 锁不互相阻塞 | httptest |
| `oplatform` ParseNotify | 四种 InfoType 解密 + verify_ticket 自动写 Store | httptest + 固定密文 fixture |
| `oplatform` AuthorizerClient wrapper | `OffiaccountClient()` 通过 TokenSource 走 oplatform 而非 /cgi-bin/token | httptest |
| `oplatform.QRLoginClient` | 5 个方法全路径（成功 + errcode 分支） | httptest |
| `oplatform/example` | `go build ./oplatform/example/...` 可编译 | go build |
| `offiaccount` TokenSource 注入 | 注入后 AccessTokenE 走注入源；未注入时走旧逻辑；AppSecret 校验条件 | 新增 client_test 用例 |
| `mini-program` TokenSource 注入 | 同上 | 新增 client_test 用例 |
| `offiaccount` 兼容 smoke | `NewMsgCrypto` 类型别名链路仍工作 | 保留轻量测试 |

覆盖策略与现有 `utils`、`merchant/developed`、`offiaccount`、`mini-program` 对齐：httptest 驱动主要路径，不强求覆盖率数字。

## 10. 兼容性保证

**外部 API 零破坏**：
- `offiaccount.Client` 所有现有方法签名不变
- `offiaccount.Config` 不变
- `offiaccount.NewMsgCrypto` 及返回类型的方法集不变
- `offiaccount.NewClient(ctx, *Config)` 签名追加 `opts ...Option` —— Go 允许对现有调用点透明
- `mini_program.Client` 所有现有方法签名不变
- `mini_program.NewClient(Config, ...Option)` 已经是 variadic options，追加新 Option 零影响
- `utils/*` 现有 API 不变

**新增**：
- `utils/wxcrypto` 新包
- `offiaccount.TokenSource` interface 与 `WithTokenSource` Option
- `mini_program.TokenSource` interface 与 `WithTokenSource` Option
- `oplatform` 全部 API

## 11. 交付物清单

1. `utils/wxcrypto/` 新包（实现 + 测试 + doc）
2. `offiaccount/crypto.go` 薄代理
3. `offiaccount/tokensource.go`、`offiaccount/client.go` 改动
4. `offiaccount/client_test.go` 追加 TokenSource 用例
5. `mini-program/tokensource.go`、`mini-program/client.go` 改动
6. `mini-program/client_test.go` 追加 TokenSource 用例
7. `oplatform/` 完整实现（按 §2.1 文件清单）
8. `oplatform/example/main.go` 可编译 demo
9. `README.md` 把 `oplatform` 从 🚧 改为 ✅（标注本轮范围）
10. `docs/README.md` 同步状态表

## 12. 非目标 / 后续工作

- 子项目 3（代小程序开发管理）：需要单独 spec，涉及代码包管理
- 子项目 4（代公众号调用业务接口清单）：通路打通后按需铺
- 子项目 5（快速注册小程序/公众号）：需要单独 spec
- 子项目 7（代小程序运营管理）：需要单独 spec
- `Store` 的 Redis/SQL 实现：由使用方提供
- SDK 内部后台刷新 goroutine：后续若有强需求再加

## 13. 开放问题

目前无未决问题。所有关键架构决策已在 brainstorm 阶段锁定：

- 持久化：`Store` 接口 + `MemoryStore` 默认实现 ✔
- 包布局：单包 `oplatform` 平铺 ✔
- 消息加解密：抽到 `utils/wxcrypto` 公共包 ✔
- 代调用形态：`TokenSource` 注入到 offiaccount / mini_program ✔
- Token 刷新：Lazy + 显式 `RefreshAll` ✔
- 测试深度：与现有模块一致，httptest 覆盖主要路径 ✔
