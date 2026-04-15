# offiaccount 模块（微信公众号）

`github.com/godrealms/go-wechat-sdk/offiaccount`

公众号模块提供两部分能力：

1. **客户端 API**：大约 200 个经过拆分的 `api.*.go` 文件，涵盖菜单、消息（客服/模板/群发/订阅）、素材、用户与标签、草稿箱、二维码、数据统计、JS-SDK、网页授权等。由于数量巨大，本文档不逐个枚举每一个方法，而是按大类给出入口点和调用模式。
2. **消息加解密 (Biz Msg Crypt) + 回调解析**：`crypto.go` + `notify.go`。这是做公众号接入必须的"服务器校验 + 接收消息"环节，本 SDK 已经完整实现。

## 1. Client

### Config

```go
type Config struct {
    AppId          string
    AppSecret      string
    Token          string // 公众号后台"开发 → 基本配置"里的 Token
    EncodingAESKey string // 43 字符；加密模式必填，明文模式可空
}
```

### 构造

```go
type Client struct {
    ctx    context.Context
    Config *Config
    Https  *utils.HTTP
    // ...
}

func NewClient(ctx context.Context, config *Config) *Client
```

ctx 用于长期持有请求时的默认上下文（例如 access_token 刷新）。建议传 `context.Background()` 或一个应用级 `ctx`。

### access_token

#### AccessTokenE *(推荐)*

```go
func (c *Client) AccessTokenE(ctx context.Context) (string, error)
```

返回当前有效的 access_token（自动刷新，双检锁缓存）。推荐在所有 API 调用前使用此方法。

#### GetStableAccessToken

```go
// 稳定版 access_token（双通道，与 /cgi-bin/token 完全隔离）
func (c *Client) GetStableAccessToken(ctx context.Context, forceRefresh bool) (*AccessToken, error)
```

调用 `/cgi-bin/stable_token`，返回稳定 token。`forceRefresh=true` 会立即作废上次的稳定 token 并返回新的。成功后会把 token 写入客户端内部缓存（60 秒安全 TTL 下限），后续 `AccessTokenE` 调用可直接命中。

> ⚠️ **Breaking change (2026-04):** 此方法原本忽略 ctx 并内部使用 `context.Background()`；现在接受显式 ctx，与其它方法保持一致。

> ⚠️ **Removed (2026-04):** 原 `GetAccessToken() string` 已删除。它静默吞掉 token 刷新错误，属于审计 2026-04-14 标记的 P0 安全隐患。请改用 `AccessTokenE(ctx)`。

### 错误类型

```go
type WeixinError struct {
    ErrCode int
    ErrMsg  string
}
func (e *WeixinError) Error() string
func (e *WeixinError) Code() int
func (e *WeixinError) Message() string

func CheckResp(r *Resp) error // errcode!=0 时返回 *WeixinError
```

几乎所有 API 的响应结构都内嵌 `Resp{ErrCode, ErrMsg}`。绝大多数 `api.*.go` 里已经帮你做了 errcode 判断；如果你自己扩展了接口调用，用 `CheckResp(&resp.Resp)` 就能快速转成 `*WeixinError`。

`WeixinError` 实现了 `utils.WechatAPIError` 接口（`Code() int` + `Message() string`）。

## 错误处理 / Error Handling

公众号 API 返回非零 errcode 时，SDK 统一返回 `*WeixinError`：

```go
var wxErr *offiaccount.WeixinError
if errors.As(err, &wxErr) {
    fmt.Println(wxErr.Code(), wxErr.Message())
}
```

`WeixinError` 实现了 `utils.WechatAPIError` 接口（`Code() int` + `Message() string`）。

旧版按字段访问依然有效：

```go
var werr *offiaccount.WeixinError
if errors.As(err, &werr) && werr.ErrCode == 40001 {
    // invalid credential，需要刷新 token
}
```

## 2. API 方法按类组织

下列表只列"大类 + 主要入口点"。每个方法的签名请直接在对应 `api.*.go` 看代码（都有中文注释）。

### 基础 (`api.base.go`, `api.api-manage.go`)
`AccessTokenE` / `GetStableAccessToken` / `CallbackCheck` / `GetCallbackIp` / `GetApiDomainIP` / `GetApiQuota` / `ClearQuota` / `GetRidInfo`。

### 自定义菜单 (`api.custom-menu.go`)
`CreateCustomMenu` / `GetMenu` / `DeleteMenu` / `AddConditionalMenu` / `DeleteConditionalMenu` / `TryMatchMenu`。

### 客服消息 (`api.customer.*.go`)
`SendKFMessage` / `AddKFAccount` / `UpdateKFAccount` / `DelKFAccount` / `GetKFList` / `GetKFMsgList` / `CreateKFSession` / `CloseKFSession` / `SetKFTyping`。

### 群发 / 自动回复 / 模板 / 订阅 (`api.notify.*.go`)
- 群发：`Massmsg*`
- 自动回复：`GetAutoReplyInfo`
- 模板：`Send` (模板消息) / `GetAllTemplates` / `DeleteTemplate` / `GetIndustry` / `SetIndustry`
- 订阅：`SubscribeSend`（订阅通知）

注意：`QueryBlockTmplMsg` 实际是小程序安全接口，已打上 `Deprecated` 标签；不要在公众号上下文使用。
`GetIndustry` 的 URL 已从旧版本的错误路径修正为 `/cgi-bin/template/get_industry`。

### 素材 (`api.material.*.go`)
- 临时素材：`UploadTempMedia` / `GetTempMedia`
- 永久素材：`UploadPermanentMaterial` / `GetPermanentMaterial` / `DelPermanentMaterial` / `GetMaterialCount` / `BatchGetMaterial`

### 用户与标签 (`api.user.manage.*.go`)
`CreateTag` / `UpdateTag` / `DeleteTag` / `GetTags` / `TagUser` / `UntagUser` / `GetUserTags` / `GetUserInfo` / `BatchGetUserInfo` / `GetFollowers` / `UpdateRemark`。

### 草稿箱 (`api.draft-box.*.go`)
`AddDraft` / `GetDraft` / `UpdateDraft` / `DeleteDraft` / `GetDraftCount` / `BatchGetDraft` / `DraftSwitch`。

### 二维码 (`api.qr-code.*.go`)
`CreateQrCode` / `CreateQrCodeJump` / `ShortenGenerate` / `ShortenFetch`。

### JS-SDK / 网页授权 (`api.web-dev.*.go`)
`GetJsapiTicket` / `GetSnsAccessToken` / `RefreshSnsAccessToken` / `CheckSnsAccessToken` / `GetSnsUserInfo`。

### 发票、数据统计、门店、医疗、留言等
数量较多，每个都有对应的 `api.*.go` 文件，使用方式一致：构造请求 struct → 调方法 → 检查 error。

## 3. 消息加解密（Biz Msg Crypt）⭐

这是做公众号接入必须的模块。微信后台可以选择"明文模式"或"安全模式/兼容模式"；安全模式下所有推送消息都被 AES-256-CBC 加密，需要服务端解密才能看到内容。

### 类型

```go
type MsgCrypto struct { /* 不导出 */ }

func NewMsgCrypto(token, encodingAESKey, appid string) (*MsgCrypto, error)
```

- `token`：公众号后台的 Token（与 `Config.Token` 一致）。
- `encodingAESKey`：43 字符，公众号后台的 EncodingAESKey。
- `appid`：公众号的 AppId，用于校验解密后消息的发送方。

### 核心方法

```go
func (m *MsgCrypto) Signature(timestamp, nonce, encryptedMsg string) string
func (m *MsgCrypto) VerifySignature(sig, timestamp, nonce, encryptedMsg string) bool

func (m *MsgCrypto) Encrypt(plaintext []byte) (base64Ciphertext string, err error)
func (m *MsgCrypto) Decrypt(encryptedMsg string) (plaintext []byte, fromAppid string, err error)

func (m *MsgCrypto) BuildEncryptedReply(plaintext []byte, timestamp, nonce string) ([]byte, error)
```

算法细节：

- **签名**：对 `[token, timestamp, nonce, encryptedMsg]` 字典序排序、拼接、SHA1，结果为十六进制字符串。使用常数时间比较防止时序攻击。
- **加密**：`16 字节随机前缀 + 4 字节大端长度 + 明文 + appid` → PKCS#7 补齐 → AES-256-CBC，IV 使用 AESKey 的前 16 字节。
- **解密**：上述过程的逆向；解密后会校验 `appid` 是否匹配 `Config.AppId`，不匹配返回错误——这一步避免跨账号伪造消息。

### 服务器接入校验 + 推送解析

```go
func VerifyServerToken(token, signature, timestamp, nonce string) error

type EncryptedEnvelope struct {
    XMLName    xml.Name `xml:"xml"`
    ToUserName string   `xml:"ToUserName"`
    Encrypt    string   `xml:"Encrypt"`
}

func ParseNotify(r *http.Request, crypto *MsgCrypto) ([]byte, error)
func ParseNotifyPlaintext(r *http.Request, token string) ([]byte, error)
```

`ParseNotify`（加密模式 / 兼容模式）的语义：

- 先校验 `timestamp` 必须在当前时间 ±5 分钟以内（防重放），超窗或缺失直接返回 error。
- `r.Method == GET`（接入校验阶段）：用 `crypto.Token()` 校验 `signature / timestamp / nonce`，通过则返回 `echostr` 字节给调用方原样写回。
- `r.Method == POST`：读 body → 解析 envelope → 用 `msg_signature` 验签 → AES 解密 → 返回明文 XML。
- `crypto` **必须非 nil**。历史上 `crypto == nil` 时会直接返回 body，这个行为已经废弃（会返回 `ErrNotifyNoCrypto`），因为它会放行任何未签名的请求；如果你用的是「明文模式」请改用 `ParseNotifyPlaintext` 并把公众号后台配置的 Token 传进去。

`ParseNotifyPlaintext`（明文模式）的语义：

- 校验 `timestamp` 在 ±5 分钟以内。
- 用传入的 `token` 校验 `signature / timestamp / nonce`（与 GET 接入签名同一个算法）。
- `r.Method == GET`：返回 `echostr`。
- `r.Method == POST`：返回原始 body（签名已验证）。

## 4. 完整使用案例：公众号服务器

```go
package main

import (
    "encoding/xml"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/godrealms/go-wechat-sdk/offiaccount"
)

const (
    appId     = "wx1234567890"
    appSecret = "abcdefg..."
    token     = "my-token"
    aesKey    = "0123456789abcdef0123456789abcdef0123456789a" // 43 chars
)

// 微信推送的文本消息结构
type incomingText struct {
    XMLName      xml.Name `xml:"xml"`
    ToUserName   string   `xml:"ToUserName"`
    FromUserName string   `xml:"FromUserName"`
    CreateTime   int64    `xml:"CreateTime"`
    MsgType      string   `xml:"MsgType"`
    Content      string   `xml:"Content"`
    MsgId        int64    `xml:"MsgId"`
}

// 要回给用户的文本消息
type outgoingText struct {
    XMLName      xml.Name `xml:"xml"`
    ToUserName   string   `xml:"ToUserName"`
    FromUserName string   `xml:"FromUserName"`
    CreateTime   int64    `xml:"CreateTime"`
    MsgType      string   `xml:"MsgType"`
    Content      string   `xml:"Content"`
}

func main() {
    crypto, err := offiaccount.NewMsgCrypto(token, aesKey, appId)
    if err != nil {
        log.Fatal(err)
    }

    http.HandleFunc("/wx/callback", func(w http.ResponseWriter, r *http.Request) {
        plaintext, err := offiaccount.ParseNotify(r, crypto)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        if r.Method == http.MethodGet {
            _, _ = w.Write(plaintext) // echostr
            return
        }

        var in incomingText
        if err := xml.Unmarshal(plaintext, &in); err != nil {
            http.Error(w, "bad xml", 400)
            return
        }

        out := outgoingText{
            ToUserName:   in.FromUserName,
            FromUserName: in.ToUserName,
            CreateTime:   time.Now().Unix(),
            MsgType:      "text",
            Content:      "你说的是：" + in.Content,
        }
        replyXML, _ := xml.Marshal(out)

        // 安全模式下需要再加密
        encrypted, err := crypto.BuildEncryptedReply(replyXML,
            r.URL.Query().Get("timestamp"),
            r.URL.Query().Get("nonce"))
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        w.Header().Set("Content-Type", "application/xml")
        _, _ = w.Write(encrypted)
    })

    // API 调用：发送模板消息、管理菜单等
    client := offiaccount.NewClient(nil, &offiaccount.Config{
        AppId:          appId,
        AppSecret:      appSecret,
        Token:          token,
        EncodingAESKey: aesKey,
    })

    token, err := client.AccessTokenE(r.Context()) // 举例
    _ = token
    _ = err
    _ = client

    _ = fmt.Sprint // 避免未使用
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## 5. 并发语义

`*offiaccount.Client` 是线程安全的。access_token 缓存用 `sync.RWMutex` 保护，多 goroutine 可以共享同一个 Client 实例。`MsgCrypto` 是无状态（除了初始化时设置的密钥）因此也可自由共享。

## 6. 已知注意事项

- `GetIndustry`：历史代码里的 URL 曾拷贝自 `GetAllTemplates`（都是 `/cgi-bin/template/get_all_private_template`），已改为官方文档正确的 `/cgi-bin/template/get_industry`。
- `QueryBlockTmplMsg`：这个接口 URL 是 `/wxa/sec/queryblocktmplmsg`，属于**小程序**安全模块，不属于公众号。保留是为向后兼容，已打 `Deprecated` 标签，新代码请勿使用。
- access_token 缓存是**进程内**的。如果你有多个应用实例（水平扩展），每个实例会各自持有自己的 token。要做分布式共享请在业务层包一层自定义缓存（后续版本计划提供 `TokenStore` 接口）。
- 默认 HTTP 超时 30s，可以通过 `client.Https.SetBaseURL` 或构造自己的 `utils.HTTP` 再赋值进去来定制（见 `crypto_test.go` 里的测试 client 构造）。
