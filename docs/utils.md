# utils 模块

`github.com/godrealms/go-wechat-sdk/utils`

`utils` 是整个 SDK 的基础设施包，提供四类能力：HTTP 客户端、RSA 签名/验签、PEM（证书 + 私钥）加载、加密安全随机串。所有上层业务模块都构建在它之上。

## 1. HTTP 客户端

### 类型

```go
type HTTP struct { /* 不导出字段 */ }

type Logger interface {
    Debugf(format string, args ...any)
}

type HTTPError struct {
    StatusCode int
    Body       []byte
    Header     http.Header
}
```

`HTTP` 是对 `net/http.Client` 的薄封装，提供统一的 JSON 请求/响应处理、per-request header 注入、状态码校验。`Logger` 默认是 `nopLogger`（什么都不打）。`HTTPError` 实现了 `error` 接口，当微信返回 4xx/5xx 时 HTTP 客户端会返回这个类型，方便调用方用 `errors.As` 解构。

### 构造与选项

```go
func NewHTTP(baseURL string, opts ...Option) *HTTP

func WithTimeout(d time.Duration) Option
func WithHeaders(h map[string]string) Option    // 默认追加到每个请求
func WithLogger(l Logger) Option
func WithHTTPClient(c *http.Client) Option      // 替换底层 *http.Client
```

`baseURL` 会被拼接到 path 前面，路径里可以带 query string（比如 `/cgi-bin/foo?access_token=X`），`NewHTTP` 和 `buildURL` 会正确处理。

### 常用方法

```go
func (h *HTTP) Get(ctx, path string, query url.Values, result any) error
func (h *HTTP) Post(ctx, path string, body any, result any) error
func (h *HTTP) Put(ctx, path string, body any, result any) error
func (h *HTTP) Patch(ctx, path string, body any, result any) error
func (h *HTTP) Delete(ctx, path string, result any) error
func (h *HTTP) PostForm(ctx, path string, form url.Values, result any) error

func (h *HTTP) DoRequest(
    ctx context.Context,
    method, path string,
    query url.Values,
    body []byte,
    headers http.Header,
    result any,
) error

func (h *HTTP) DoRequestWithRawResponse(
    ctx context.Context,
    method, path string,
    query url.Values,
    body []byte,
    headers http.Header,
) (statusCode int, respHeader http.Header, respBody []byte, err error)
```

语义说明：

- `Get/Post/Put/Patch/Delete/PostForm`：便捷封装，自动 JSON 编解码。`result` 传 `nil` 表示不解析响应体。
- `DoRequest`：完全控制版本，`body` 是已序列化好的字节，`headers` 会每次请求单独使用，**不会修改客户端级别的默认 header**——这是支付 V3 Authorization 签名所需要的性质（不同请求签名不同）。
- `DoRequestWithRawResponse`：在 `DoRequest` 基础上把原始响应 body 也返回给调用方，用于后续的签名校验等额外处理。
- 所有方法都会校验 HTTP 状态码必须是 2xx；否则返回 `*HTTPError`。

### 使用案例

```go
import (
    "context"
    "errors"
    "fmt"
    "net/url"
    "time"

    "github.com/godrealms/go-wechat-sdk/utils"
)

type weather struct {
    City string  `json:"city"`
    Temp float64 `json:"temp"`
}

func fetchWeather(ctx context.Context, city string) (*weather, error) {
    http := utils.NewHTTP("https://example.com/api",
        utils.WithTimeout(5*time.Second),
        utils.WithHeaders(map[string]string{"X-Client": "go-wechat-sdk"}),
    )

    out := &weather{}
    err := http.Get(ctx, "/weather", url.Values{"city": {city}}, out)
    if err != nil {
        var httpErr *utils.HTTPError
        if errors.As(err, &httpErr) {
            return nil, fmt.Errorf("weather api %d: %s", httpErr.StatusCode, httpErr.Body)
        }
        return nil, err
    }
    return out, nil
}
```

### Logger 与 PII 风险

`HTTP.Logger` 默认是 `nopLogger`——即 SDK 不会把任何请求/响应内容打印到标准输出，因此**在不启用 `WithLogger` 的情况下并不会泄漏 token 或 openid**。

但是一旦你通过 `WithLogger(...)` 注入自己的 `Logger`，SDK 会把请求 URL 与请求/响应 body 原样 `Debugf` 出去。WeChat Pay V3 的请求 URL 与请求头本身并不承载敏感凭证，但 `refund` / `profitsharing` / `transfer` 等接口的请求/响应 body 会包含加密后的 openid、银行信息等 PII。**接入自定义 logger 前，请先确认你的 logger sink（日志文件、日志中心、第三方采集等）是否满足合规要求**，避免把 PII 写入不受控的存储。

## 2. 签名 / 验签

### 函数

```go
func SignSHA256WithRSA(source string, privateKey *rsa.PrivateKey) (string, error)
func VerifySHA256WithRSA(source, signatureBase64 string, publicKey *rsa.PublicKey) error
```

- `SignSHA256WithRSA`：对 `source` 做 SHA-256，然后用私钥 PKCS1v15 签名，最后 Base64 编码返回。微信支付 V3 的 `Authorization` 头就是这么算出来的。
- `VerifySHA256WithRSA`：反向操作，验证失败返回非 nil `error`。

### 使用案例

```go
priv, _ := utils.LoadPrivateKeyWithPath("apiclient_key.pem")

source := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n",
    "POST", "/v3/pay/transactions/jsapi", timestamp, nonce, body)
sig, err := utils.SignSHA256WithRSA(source, priv)
if err != nil {
    log.Fatal(err)
}
// sig 是 Base64 字符串，直接拼进 Authorization header
```

## 3. PEM 加载（证书 + 私钥 + 公钥）

### 函数

```go
// 从字符串加载
func LoadCertificate(pem string) (*x509.Certificate, error)
func LoadPrivateKey(pem string) (*rsa.PrivateKey, error)
func LoadPublicKey(pem string) (*rsa.PublicKey, error)

// 从文件加载（推荐，不用自己 os.ReadFile）
func LoadCertificateWithPath(path string) (*x509.Certificate, error)
func LoadPrivateKeyWithPath(path string) (*rsa.PrivateKey, error)
func LoadPublicKeyWithPath(path string) (*rsa.PublicKey, error)

// 工具
func GetCertificateSerialNumber(c x509.Certificate) string
func IsCertValid(c x509.Certificate, now time.Time) bool
func IsCertExpired(c x509.Certificate, now time.Time) bool
```

特性：

- `LoadPrivateKey` **同时支持 PKCS#1 (`-----BEGIN RSA PRIVATE KEY-----`) 和 PKCS#8 (`-----BEGIN PRIVATE KEY-----`)**，会根据 `pem.Block.Type` 自动选择解析器，无须你手动区分。
- `GetCertificateSerialNumber` 返回的是大写十六进制字符串，可以直接用于微信支付的 `Wechatpay-Serial` 头比对。

### 使用案例

```go
cert, err := utils.LoadCertificateWithPath("apiclient_cert.pem")
if err != nil {
    log.Fatal(err)
}
serial := utils.GetCertificateSerialNumber(*cert)
log.Printf("using cert serial=%s valid=%v", serial, utils.IsCertValid(*cert, time.Now()))
```

## 4. 随机串

```go
func RandomString(n int) string          // 默认字符集：大小写字母+数字
func RandomStringWithCharset(n int, charset string) string
func ShuffleString(s string) string
```

**重要：所有随机串都使用 `crypto/rand`**（不是 `math/rand`），可以安全用于：微信支付 V3 的 `nonce_str`、防重放 token、订单号生成等对抗可预测性的场景。

### 使用案例

```go
nonce := utils.RandomString(32) // 32 个字符的 nonce_str，签名时直接用
```

## 5. 错误类型

```go
type HTTPError struct {
    StatusCode int
    Body       []byte
    Header     http.Header
}

func (e *HTTPError) Error() string
```

所有 HTTP 请求在状态码不是 2xx 时返回 `*HTTPError`。可以用 `errors.As` 拆开：

```go
err := client.Get(ctx, "/foo", nil, &out)
var httpErr *utils.HTTPError
if errors.As(err, &httpErr) && httpErr.StatusCode == 404 {
    // 业务上的"不存在"
}
```

## 6. 并发语义

`*HTTP` 可以在多 goroutine 间自由共享。它内部没有可变状态，`DoRequest` / `DoRequestWithRawResponse` 会为每次请求构造独立的 `http.Request`，不会修改传入的 `headers` 或客户端级别的默认 header。

## 7. 完整使用案例：手写一个 V3 POST

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/godrealms/go-wechat-sdk/utils"
)

func main() {
    ctx := context.Background()

    priv, err := utils.LoadPrivateKeyWithPath("apiclient_key.pem")
    if err != nil { log.Fatal(err) }

    httpC := utils.NewHTTP("https://api.mch.weixin.qq.com",
        utils.WithTimeout(10*time.Second))

    body, _ := json.Marshal(map[string]any{
        "mchid":        "1900000001",
        "out_trade_no": "demo-001",
    })

    ts    := time.Now().Unix()
    nonce := utils.RandomString(32)
    source := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n",
        "POST", "/v3/pay/transactions/jsapi", ts, nonce, body)
    sig, err := utils.SignSHA256WithRSA(source, priv)
    if err != nil { log.Fatal(err) }

    h := http.Header{}
    h.Set("Authorization", fmt.Sprintf(
        `WECHATPAY2-SHA256-RSA2048 mchid="1900000001",nonce_str="%s",signature="%s",timestamp="%d",serial_no="CERT_SERIAL"`,
        nonce, sig, ts))

    var resp map[string]any
    err = httpC.DoRequest(ctx, "POST", "/v3/pay/transactions/jsapi",
        nil, body, h, &resp)
    if err != nil { log.Fatal(err) }
    fmt.Printf("%+v\n", resp)
}
```

> 真实接入时不需要这么写——直接用 `merchant/developed.Client` 就好，这段代码只是为了演示 utils 的原语。
