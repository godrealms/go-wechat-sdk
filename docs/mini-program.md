# mini-program 模块（微信小程序）

`github.com/godrealms/go-wechat-sdk/mini-program` — 包名：`mini_program`

小程序服务端 SDK，当前实现了做小程序登录和订阅消息所必需的最小闭环：

1. `Code2Session` — 用 `wx.login` 的临时 `code` 换 `openid` + `session_key`
2. `AccessToken` — 获取/缓存全局 `access_token`（进程内缓存、并发安全）
3. `DecryptUserData` — 解密 `wx.getUserInfo` / `wx.getPhoneNumber` 返回的 `encryptedData`
4. `SendSubscribeMessage` — 发送订阅消息

其它能力（数据分析、云开发、直播、虚拟支付等）没有逐个封装；通过 `Client.HTTP()` 暴露出来的 `*utils.HTTP` 可以直接手工调用任何 `https://api.weixin.qq.com` 下的接口。

## 1. Config 与 Client

```go
type Config struct {
    AppId     string
    AppSecret string
}

type Client struct { /* 不导出字段 */ }

func NewClient(cfg Config, opts ...Option) (*Client, error)

type Option func(*Client)
func WithHTTP(h *utils.HTTP) Option     // 替换底层 HTTP（单测常用）

func (c *Client) HTTP() *utils.HTTP     // 暴露底层 HTTP，便于扩展
```

`NewClient` 会校验 `AppId`、`AppSecret` 非空，默认构造一个 30 秒超时、baseURL 指向 `https://api.weixin.qq.com` 的 `*utils.HTTP`。

`HTTP()` 返回的是内部使用的同一实例——如果要定制请求头、日志、超时，建议在 `NewClient` 之前用 `WithHTTP(utils.NewHTTP(...))` 注入自己的版本，而不是事后修改共享实例。

## 2. 登录：Code2Session

```go
type Code2SessionResp struct {
    OpenId     string `json:"openid"`
    SessionKey string `json:"session_key"`
    UnionId    string `json:"unionid,omitempty"`
    ErrCode    int    `json:"errcode,omitempty"`
    ErrMsg     string `json:"errmsg,omitempty"`
}

func (c *Client) Code2Session(ctx context.Context, jsCode string) (*Code2SessionResp, error)
```

- 调用的是 `GET /sns/jscode2session`。
- `jsCode` 为空会直接返回错误，不会产生 HTTP 请求。
- 当响应里 `errcode != 0` 时返回 `*APIError`；调用方可以用 `errors.As` 拿到结构化错误码（见下方错误处理章节）。
- `session_key` **必须在服务端保管**，绝不能下发到小程序前端。建议用 `openid` 作为 key 写进你的 session / Redis。

### 使用案例：小程序登录

```go
func loginHandler(w http.ResponseWriter, r *http.Request) {
    jsCode := r.URL.Query().Get("code")
    resp, err := client.Code2Session(r.Context(), jsCode)
    if err != nil {
        http.Error(w, err.Error(), 400)
        return
    }

    // 把 session_key 存到服务端（不能下发给前端！）
    redis.Set(r.Context(),
        "wx:session:"+resp.OpenId, resp.SessionKey, 2*time.Hour)

    // 只把自己的 token 发回前端
    myToken := issueJWT(resp.OpenId, resp.UnionId)
    json.NewEncoder(w).Encode(map[string]string{"token": myToken})
}
```

## 3. access_token 缓存

```go
func (c *Client) AccessToken(ctx context.Context) (string, error)
```

语义：

- 第一次调用时请求 `GET /cgi-bin/token?grant_type=client_credential`，把返回的 `access_token` + 到期时间存入 Client 内部。
- 到期时间 = `now + (expires_in - 60s)`，提前 60 秒视为过期，规避临界点。
- 后续调用走 `sync.RWMutex` 的读锁路径，零网络；过期时走写锁路径的双重检查，确保并发情况下只刷一次。
- `errcode != 0` 或空 token 都会返回 `*APIError`（见下方错误处理章节）。

这个缓存是**进程内**的。水平扩展多实例时，每个实例独立持有自己的 token。如需分布式共享，请在业务层包一层，用 `WithHTTP` 注入自己的 HTTP 客户端并在中间件里接管 token 行为，或者直接放弃 `AccessToken` 走自己的方案。

## 4. 订阅消息

```go
func (c *Client) SendSubscribeMessage(ctx context.Context, body any) error
```

调用 `POST /cgi-bin/message/subscribe/send?access_token=...`，`body` 会被 JSON 序列化为请求体。签名跟微信官方文档一致，字段由调用方自己组装：

```go
err := client.SendSubscribeMessage(ctx, map[string]any{
    "touser":      openid,
    "template_id": "abcd1234",
    "page":        "pages/order/detail?id=100",
    "lang":        "zh_CN",
    "miniprogram_state": "formal",
    "data": map[string]any{
        "thing1":  map[string]any{"value": "拿铁"},
        "amount2": map[string]any{"value": "￥28.00"},
        "time3":   map[string]any{"value": "2026-04-12 10:30"},
    },
})
```

`errcode != 0` 会返回 `*APIError`；成功返回 `nil`。

## 5. 解密 encryptedData（WxBizDataCrypt）

```go
func DecryptUserData(sessionKey, encryptedData, iv string, result any) ([]byte, error)
```

这是一个**包级函数**，不依赖 `Client`——因为它只需要 `session_key`，而 `session_key` 是你从 Redis/DB 自己读出来的，跟 Client 的生命周期无关。

参数：

- `sessionKey`：`Code2Session` 返回的 `SessionKey` 字段，base64 字符串。
- `encryptedData`：前端 `wx.getUserInfo` / `wx.getPhoneNumber` 回调里的 `encryptedData`，base64。
- `iv`：前端给的 `iv`，base64。
- `result`：可选的 `*struct` 指针，非 nil 会把解密后的 JSON 反序列化进去；传 nil 则只返回明文字节。

算法：base64 解码 → AES-128-CBC（key 必须 16 字节，iv 必须 16 字节）→ PKCS#7 去填充 → 可选 JSON Unmarshal。

⚠️ **安全提示**：微信规定 `session_key` 绝不能出现在客户端。上述参数应当全部由**你的**服务端接收：`encryptedData` 和 `iv` 由小程序前端上传，`session_key` 由服务端从存储里取出。

### 使用案例：解密手机号

```go
// 前端上送 encryptedData + iv；服务端根据用户 id 取出 session_key
type PhoneInfo struct {
    PhoneNumber     string `json:"phoneNumber"`
    PurePhoneNumber string `json:"purePhoneNumber"`
    CountryCode     string `json:"countryCode"`
    Watermark       struct {
        AppId     string `json:"appid"`
        Timestamp int64  `json:"timestamp"`
    } `json:"watermark"`
}

func decryptPhone(openid, encryptedData, iv string) (*PhoneInfo, error) {
    sessionKey, err := redis.Get(ctx, "wx:session:"+openid).Result()
    if err != nil {
        return nil, fmt.Errorf("session expired, please re-login: %w", err)
    }
    out := &PhoneInfo{}
    if _, err := mini_program.DecryptUserData(sessionKey, encryptedData, iv, out); err != nil {
        return nil, err
    }
    // 建议校验 watermark.appid 与自己的 appid 一致
    if out.Watermark.AppId != myAppId {
        return nil, fmt.Errorf("watermark appid mismatch")
    }
    return out, nil
}
```

## 6. 扩展：调用未封装的接口

如果你要调的接口没封装，但接口协议是 `token in query + JSON body`，直接用 `Client.HTTP()`：

```go
token, err := client.AccessToken(ctx)
if err != nil { return err }

q := url.Values{"access_token": {token}}
body := map[string]any{ /* 业务字段 */ }
out := map[string]any{}

// path 直接带 query，utils.HTTP 会正确拼接
if err := client.HTTP().Post(ctx,
    "/wxa/business/getuserphonenumber?"+q.Encode(),
    body, &out); err != nil {
    return err
}
```

常用扩展点：

- `POST /wxa/msg_sec_check` 内容安全
- `POST /wxa/img_sec_check` 图片安全
- `POST /wxa/getwxacodeunlimit` 生成无限制小程序码
- `GET  /cgi-bin/analysis/*` 数据分析系列

## 7. 完整使用案例：登录 + 发订阅消息

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"

    mp "github.com/godrealms/go-wechat-sdk/mini-program"
)

var client *mp.Client

func main() {
    var err error
    client, err = mp.NewClient(mp.Config{
        AppId:     "wx1234567890",
        AppSecret: "your-app-secret",
    })
    if err != nil { log.Fatal(err) }

    http.HandleFunc("/api/login", loginHandler)
    http.HandleFunc("/api/order/paid", orderPaidHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// 1. 登录：前端 wx.login 拿到 code，调这个接口换 openid
func loginHandler(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    resp, err := client.Code2Session(r.Context(), code)
    if err != nil {
        http.Error(w, err.Error(), 400)
        return
    }
    // 业务上：把 session_key 存到 Redis，用 openid 作 key
    // 返回给前端你自己签发的 token（不要返回 session_key）
    _ = json.NewEncoder(w).Encode(map[string]string{
        "openid": resp.OpenId,
    })
}

// 2. 下单成功后发订阅消息
func orderPaidHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    err := client.SendSubscribeMessage(ctx, map[string]any{
        "touser":      "oXYZ-openid",
        "template_id": "your-template-id",
        "page":        "pages/order/detail?id=100",
        "data": map[string]any{
            "thing1":  map[string]any{"value": "拿铁"},
            "amount2": map[string]any{"value": "￥28.00"},
            "time3":   map[string]any{"value": time.Now().Format("2006-01-02 15:04")},
        },
    })
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
```

## 8. 错误处理 / Error Handling

小程序 API 返回非零 errcode 时，SDK 统一返回 `*mini_program.APIError`：

```go
var apiErr *mini_program.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.Code(), apiErr.Message(), apiErr.Path)
}
```

`APIError` 实现了 `utils.WechatAPIError` 接口（`Code() int` + `Message() string`）。

`Code2Session`、`AccessToken` 和 `SendSubscribeMessage` 均会在 `errcode != 0` 时返回 `*APIError`，可以统一用 `errors.As` 分支处理。

## 9. 并发语义

- `*mini_program.Client` 是并发安全的，可以在多 goroutine 间共享。
- `AccessToken` 使用 `sync.RWMutex` + 双重检查，缓存命中走读锁路径；未命中只有一个 goroutine 会真正去刷 token。
- `Code2Session` 和 `SendSubscribeMessage` 是无状态的，完全依赖 `*utils.HTTP`——后者也是并发安全的（见 `utils.md`）。
- `DecryptUserData` 是纯函数，可以随意并发调用。

## 10. 已知注意事项

- `session_key` 只能留在服务端。下发到前端 = 把你小程序的加密通信拱手交出去。
- access_token 的缓存是**进程内**的。多实例部署时每个进程会各自持有一份。
- 微信的 `access_token` 有调用次数限制，频繁重启服务会触发配额问题；生产环境建议用稳定版 `access_token` 或在中间层做缓存。本 SDK 暂未封装稳定版 access_token 接口，可以通过 `Client.HTTP()` 手工调 `POST /cgi-bin/stable_token`。
- `SendSubscribeMessage` 的 `body` 类型是 `any`，你可以传 `struct` 或 `map[string]any`，只要 JSON 序列化出来的字段名符合微信文档即可。
