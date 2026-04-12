# Mini-Program Package Expansion — Design Spec

**Package:** `github.com/godrealms/go-wechat-sdk/mini-program`
**Date:** 2026-04-12
**Depends on:** Existing mini-program foundation (client.go, crypto.go, tokensource.go)

---

## 1. Scope

14 new public methods + 3 private helpers across 7 areas:

| # | Method | HTTP | Path | Returns |
|---|---|---|---|---|
| 1 | GetWxaCode | POST | `/wxa/getwxacode` | []byte (PNG) |
| 2 | GetWxaCodeUnlimit | POST | `/wxa/getwxacodeunlimit` | []byte (PNG) |
| 3 | CreateQRCode | POST | `/cgi-bin/wxaapp/createwxaqrcode` | []byte (PNG) |
| 4 | MsgSecCheck | POST | `/wxa/msg_sec_check` | JSON |
| 5 | MediaCheckAsync | POST | `/wxa/media_check_async` | JSON |
| 6 | GenerateScheme | POST | `/wxa/generatescheme` | JSON |
| 7 | GenerateUrlLink | POST | `/wxa/generate_urllink` | JSON |
| 8 | GenerateShortLink | POST | `/wxa/genwxashortlink` | JSON |
| 9 | GetPhoneNumber | POST | `/wxa/business/getuserphonenumber` | JSON |
| 10 | GetDailySummary | POST | `/datacube/getweanalysisappiddailysummarytrend` | JSON |
| 11 | GetVisitPage | POST | `/datacube/getweanalysisappidvisitpage` | JSON |
| 12 | GetDailyVisitTrend | POST | `/datacube/getweanalysisappiddailyvisittrend` | JSON |
| 13 | UploadTempMedia | POST | `/cgi-bin/media/upload` | JSON (multipart) |
| 14 | GetTempMedia | GET | `/cgi-bin/media/get` | []byte (binary) |

### 1.1 Private Helpers

| Helper | Purpose |
|---|---|
| `doGet(ctx, path, extra, out)` | GET with auto access_token |
| `doPost(ctx, path, body, out)` | POST JSON with auto access_token + errcode check |
| `doPostRaw(ctx, path, body) ([]byte, error)` | POST JSON returning raw bytes, with error sniffing |

---

## 2. File Layout

| File | Content |
|---|---|
| `helper.go` (new) | baseResp, doGet, doPost, doPostRaw |
| `wxacode.go` (new) | GetWxaCodeReq, GetWxaCodeUnlimitReq, CreateQRCodeReq, Color; 3 methods |
| `wxacode_test.go` (new) | 3 tests |
| `security.go` (new) | MsgSecCheckReq/Resp, SecCheckResult, SecCheckDetail, MediaCheckAsyncReq/Resp; 2 methods |
| `security_test.go` (new) | 2 tests |
| `urlscheme.go` (new) | GenerateSchemeReq/Resp, JumpWxa, GenerateUrlLinkReq/Resp, GenerateShortLinkReq/Resp; 3 methods |
| `urlscheme_test.go` (new) | 3 tests |
| `phone.go` (new) | GetPhoneNumberReq, PhoneInfo, Watermark, GetPhoneNumberResp; 1 method |
| `phone_test.go` (new) | 1 test |
| `analysis.go` (new) | AnalysisDateReq, DailySummaryItem, GetDailySummaryResp, VisitPageItem, GetVisitPageResp, DailyVisitTrendItem, GetDailyVisitTrendResp; 3 methods |
| `analysis_test.go` (new) | 3 tests |
| `media.go` (new) | UploadTempMediaResp; 2 methods (UploadTempMedia, GetTempMedia) |
| `media_test.go` (new) | 2 tests |

---

## 3. DTOs

### 3.1 Helper (`helper.go`)

```go
// baseResp 微信公共错误字段。
type baseResp struct {
    ErrCode int    `json:"errcode"`
    ErrMsg  string `json:"errmsg"`
}
```

### 3.2 Wxacode (`wxacode.go`)

```go
type GetWxaCodeReq struct {
    Path      string `json:"path"`
    Width     int    `json:"width,omitempty"`
    AutoColor bool   `json:"auto_color,omitempty"`
    LineColor *Color `json:"line_color,omitempty"`
    IsHyaline bool   `json:"is_hyaline,omitempty"`
}

type GetWxaCodeUnlimitReq struct {
    Scene      string `json:"scene"`
    Page       string `json:"page,omitempty"`
    Width      int    `json:"width,omitempty"`
    AutoColor  bool   `json:"auto_color,omitempty"`
    LineColor  *Color `json:"line_color,omitempty"`
    IsHyaline  bool   `json:"is_hyaline,omitempty"`
    CheckPath  bool   `json:"check_path,omitempty"`
    EnvVersion string `json:"env_version,omitempty"`
}

type CreateQRCodeReq struct {
    Path  string `json:"path"`
    Width int    `json:"width,omitempty"`
}

type Color struct {
    R int `json:"r"`
    G int `json:"g"`
    B int `json:"b"`
}
```

### 3.3 Security (`security.go`)

```go
type MsgSecCheckReq struct {
    Content   string `json:"content"`
    Version   int    `json:"version"`
    Scene     int    `json:"scene"`
    OpenID    string `json:"openid"`
    Title     string `json:"title,omitempty"`
    Nickname  string `json:"nickname,omitempty"`
    Signature string `json:"signature,omitempty"`
}

type MsgSecCheckResp struct {
    TraceID string           `json:"trace_id"`
    Result  SecCheckResult   `json:"result"`
    Detail  []SecCheckDetail `json:"detail"`
}

type SecCheckResult struct {
    Suggest string `json:"suggest"`
    Label   int    `json:"label"`
}

type SecCheckDetail struct {
    Strategy string `json:"strategy"`
    ErrCode  int    `json:"errcode"`
    Suggest  string `json:"suggest"`
    Label    int    `json:"label"`
    Prob     int    `json:"prob"`
    Keyword  string `json:"keyword"`
}

type MediaCheckAsyncReq struct {
    MediaURL  string `json:"media_url"`
    MediaType int    `json:"media_type"`
    Version   int    `json:"version"`
    Scene     int    `json:"scene"`
    OpenID    string `json:"openid"`
}

type MediaCheckAsyncResp struct {
    TraceID string `json:"trace_id"`
}
```

### 3.4 URL Scheme (`urlscheme.go`)

```go
type GenerateSchemeReq struct {
    JumpWxa        *JumpWxa `json:"jump_wxa,omitempty"`
    IsExpire       bool     `json:"is_expire,omitempty"`
    ExpireType     int      `json:"expire_type,omitempty"`
    ExpireTime     int64    `json:"expire_time,omitempty"`
    ExpireInterval int      `json:"expire_interval,omitempty"`
}

type JumpWxa struct {
    Path       string `json:"path"`
    Query      string `json:"query"`
    EnvVersion string `json:"env_version,omitempty"`
}

type GenerateSchemeResp struct {
    OpenLink string `json:"openlink"`
}

type GenerateUrlLinkReq struct {
    Path           string `json:"path,omitempty"`
    Query          string `json:"query,omitempty"`
    EnvVersion     string `json:"env_version,omitempty"`
    IsExpire       bool   `json:"is_expire,omitempty"`
    ExpireType     int    `json:"expire_type,omitempty"`
    ExpireTime     int64  `json:"expire_time,omitempty"`
    ExpireInterval int    `json:"expire_interval,omitempty"`
}

type GenerateUrlLinkResp struct {
    URLLink string `json:"url_link"`
}

type GenerateShortLinkReq struct {
    PageURL     string `json:"page_url"`
    PageTitle   string `json:"page_title,omitempty"`
    IsPermanent bool   `json:"is_permanent,omitempty"`
}

type GenerateShortLinkResp struct {
    Link string `json:"link"`
}
```

### 3.5 Phone (`phone.go`)

```go
type GetPhoneNumberReq struct {
    Code string `json:"code"`
}

type PhoneInfo struct {
    PhoneNumber     string    `json:"phoneNumber"`
    PurePhoneNumber string    `json:"purePhoneNumber"`
    CountryCode     string    `json:"countryCode"`
    Watermark       Watermark `json:"watermark"`
}

type Watermark struct {
    AppID     string `json:"appid"`
    Timestamp int64  `json:"timestamp"`
}

type GetPhoneNumberResp struct {
    PhoneInfo PhoneInfo `json:"phone_info"`
}
```

### 3.6 Analysis (`analysis.go`)

```go
type AnalysisDateReq struct {
    BeginDate string `json:"begin_date"`
    EndDate   string `json:"end_date"`
}

type DailySummaryItem struct {
    RefDate    string `json:"ref_date"`
    VisitTotal int    `json:"visit_total"`
    SharePV    int    `json:"share_pv"`
    ShareUV    int    `json:"share_uv"`
}

type GetDailySummaryResp struct {
    List []DailySummaryItem `json:"list"`
}

type VisitPageItem struct {
    PagePath       string  `json:"page_path"`
    PageVisitPV    int     `json:"page_visit_pv"`
    PageVisitUV    int     `json:"page_visit_uv"`
    PageStaytimePV float64 `json:"page_staytime_pv"`
    EntryPagePV    int     `json:"entrypage_pv"`
    ExitPagePV     int     `json:"exitpage_pv"`
    PageSharePV    int     `json:"page_share_pv"`
    PageShareUV    int     `json:"page_share_uv"`
}

type GetVisitPageResp struct {
    RefDate string          `json:"ref_date"`
    List    []VisitPageItem `json:"list"`
}

type DailyVisitTrendItem struct {
    RefDate         string  `json:"ref_date"`
    SessionCnt      int     `json:"session_cnt"`
    VisitPV         int     `json:"visit_pv"`
    VisitUV         int     `json:"visit_uv"`
    VisitUVNew      int     `json:"visit_uv_new"`
    StayTimeUV      float64 `json:"stay_time_uv"`
    StayTimeSession float64 `json:"stay_time_session"`
    VisitDepth      float64 `json:"visit_depth"`
}

type GetDailyVisitTrendResp struct {
    List []DailyVisitTrendItem `json:"list"`
}
```

### 3.7 Media (`media.go`)

```go
type UploadTempMediaResp struct {
    Type      string `json:"type"`
    MediaID   string `json:"media_id"`
    CreatedAt int64  `json:"created_at"`
}
```

---

## 4. Method Implementations

### 4.1 Helpers (`helper.go`)

```go
func (c *Client) doGet(ctx context.Context, path string, extra url.Values, out any) error {
    tok, err := c.AccessToken(ctx)
    if err != nil {
        return err
    }
    q := url.Values{"access_token": {tok}}
    for k, vs := range extra {
        q[k] = vs
    }
    return c.http.Get(ctx, path, q, out)
}

func (c *Client) doPost(ctx context.Context, path string, body any, out any) error {
    tok, err := c.AccessToken(ctx)
    if err != nil {
        return err
    }
    q := url.Values{"access_token": {tok}}
    fullPath := path + "?" + q.Encode()
    if out != nil {
        return c.http.Post(ctx, fullPath, body, out)
    }
    var resp baseResp
    if err := c.http.Post(ctx, fullPath, body, &resp); err != nil {
        return err
    }
    if resp.ErrCode != 0 {
        return fmt.Errorf("mini_program: %s errcode=%d errmsg=%s", path, resp.ErrCode, resp.ErrMsg)
    }
    return nil
}

func (c *Client) doPostRaw(ctx context.Context, path string, body any) ([]byte, error) {
    tok, err := c.AccessToken(ctx)
    if err != nil {
        return nil, err
    }
    q := url.Values{"access_token": {tok}}
    raw, err := json.Marshal(body)
    if err != nil {
        return nil, err
    }
    _, _, respBody, err := c.http.DoRequestWithRawResponse(ctx, http.MethodPost, path, q, raw, nil)
    if err != nil {
        return nil, err
    }
    if len(respBody) > 0 && respBody[0] == '{' {
        var resp baseResp
        if json.Unmarshal(respBody, &resp) == nil && resp.ErrCode != 0 {
            return nil, fmt.Errorf("mini_program: %s errcode=%d errmsg=%s", path, resp.ErrCode, resp.ErrMsg)
        }
    }
    return respBody, nil
}
```

### 4.2 Wxacode (`wxacode.go`)

```go
func (c *Client) GetWxaCode(ctx context.Context, req *GetWxaCodeReq) ([]byte, error) {
    return c.doPostRaw(ctx, "/wxa/getwxacode", req)
}

func (c *Client) GetWxaCodeUnlimit(ctx context.Context, req *GetWxaCodeUnlimitReq) ([]byte, error) {
    return c.doPostRaw(ctx, "/wxa/getwxacodeunlimit", req)
}

func (c *Client) CreateQRCode(ctx context.Context, req *CreateQRCodeReq) ([]byte, error) {
    return c.doPostRaw(ctx, "/cgi-bin/wxaapp/createwxaqrcode", req)
}
```

### 4.3 Security (`security.go`)

```go
func (c *Client) MsgSecCheck(ctx context.Context, req *MsgSecCheckReq) (*MsgSecCheckResp, error) {
    var resp MsgSecCheckResp
    if err := c.doPost(ctx, "/wxa/msg_sec_check", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

func (c *Client) MediaCheckAsync(ctx context.Context, req *MediaCheckAsyncReq) (*MediaCheckAsyncResp, error) {
    var resp MediaCheckAsyncResp
    if err := c.doPost(ctx, "/wxa/media_check_async", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.4 URL Scheme (`urlscheme.go`)

```go
func (c *Client) GenerateScheme(ctx context.Context, req *GenerateSchemeReq) (*GenerateSchemeResp, error) {
    var resp GenerateSchemeResp
    if err := c.doPost(ctx, "/wxa/generatescheme", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

func (c *Client) GenerateUrlLink(ctx context.Context, req *GenerateUrlLinkReq) (*GenerateUrlLinkResp, error) {
    var resp GenerateUrlLinkResp
    if err := c.doPost(ctx, "/wxa/generate_urllink", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

func (c *Client) GenerateShortLink(ctx context.Context, req *GenerateShortLinkReq) (*GenerateShortLinkResp, error) {
    var resp GenerateShortLinkResp
    if err := c.doPost(ctx, "/wxa/genwxashortlink", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.5 Phone (`phone.go`)

```go
func (c *Client) GetPhoneNumber(ctx context.Context, code string) (*GetPhoneNumberResp, error) {
    body := &GetPhoneNumberReq{Code: code}
    var resp GetPhoneNumberResp
    if err := c.doPost(ctx, "/wxa/business/getuserphonenumber", body, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.6 Analysis (`analysis.go`)

```go
func (c *Client) GetDailySummary(ctx context.Context, beginDate, endDate string) (*GetDailySummaryResp, error) {
    body := &AnalysisDateReq{BeginDate: beginDate, EndDate: endDate}
    var resp GetDailySummaryResp
    if err := c.doPost(ctx, "/datacube/getweanalysisappiddailysummarytrend", body, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

func (c *Client) GetVisitPage(ctx context.Context, beginDate, endDate string) (*GetVisitPageResp, error) {
    body := &AnalysisDateReq{BeginDate: beginDate, EndDate: endDate}
    var resp GetVisitPageResp
    if err := c.doPost(ctx, "/datacube/getweanalysisappidvisitpage", body, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

func (c *Client) GetDailyVisitTrend(ctx context.Context, beginDate, endDate string) (*GetDailyVisitTrendResp, error) {
    body := &AnalysisDateReq{BeginDate: beginDate, EndDate: endDate}
    var resp GetDailyVisitTrendResp
    if err := c.doPost(ctx, "/datacube/getweanalysisappiddailyvisittrend", body, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.7 Media (`media.go`)

```go
func (c *Client) UploadTempMedia(ctx context.Context, mediaType, fileName string, fileData io.Reader) (*UploadTempMediaResp, error) {
    tok, err := c.AccessToken(ctx)
    if err != nil {
        return nil, err
    }
    q := url.Values{"access_token": {tok}, "type": {mediaType}}
    var buf bytes.Buffer
    w := multipart.NewWriter(&buf)
    fw, err := w.CreateFormFile("media", fileName)
    if err != nil {
        return nil, err
    }
    if _, err := io.Copy(fw, fileData); err != nil {
        return nil, err
    }
    w.Close()
    _, _, respBody, err := c.http.DoRequestWithRawResponse(
        ctx, http.MethodPost, "/cgi-bin/media/upload", q,
        buf.Bytes(),
        http.Header{"Content-Type": {w.FormDataContentType()}},
    )
    if err != nil {
        return nil, err
    }
    var resp struct {
        baseResp
        UploadTempMediaResp
    }
    if err := json.Unmarshal(respBody, &resp); err != nil {
        return nil, err
    }
    if resp.ErrCode != 0 {
        return nil, fmt.Errorf("mini_program: upload media errcode=%d errmsg=%s", resp.ErrCode, resp.ErrMsg)
    }
    return &resp.UploadTempMediaResp, nil
}

func (c *Client) GetTempMedia(ctx context.Context, mediaID string) ([]byte, error) {
    tok, err := c.AccessToken(ctx)
    if err != nil {
        return nil, err
    }
    q := url.Values{"access_token": {tok}, "media_id": {mediaID}}
    _, _, respBody, err := c.http.DoRequestWithRawResponse(
        ctx, http.MethodGet, "/cgi-bin/media/get", q, nil, nil,
    )
    if err != nil {
        return nil, err
    }
    if len(respBody) > 0 && respBody[0] == '{' {
        var resp baseResp
        if json.Unmarshal(respBody, &resp) == nil && resp.ErrCode != 0 {
            return nil, fmt.Errorf("mini_program: get media errcode=%d errmsg=%s", resp.ErrCode, resp.ErrMsg)
        }
    }
    return respBody, nil
}
```

---

## 5. Testing

### 5.1 Test Matrix

| # | Test | File | Validates |
|---|---|---|---|
| 1 | TestGetWxaCode | wxacode_test.go | POST path, body, binary response |
| 2 | TestGetWxaCodeUnlimit | wxacode_test.go | POST path, scene in body, binary response |
| 3 | TestCreateQRCode | wxacode_test.go | POST path, body, binary response |
| 4 | TestMsgSecCheck | security_test.go | POST path, body, JSON response with result/detail |
| 5 | TestMediaCheckAsync | security_test.go | POST path, body, trace_id response |
| 6 | TestGenerateScheme | urlscheme_test.go | POST path, body, openlink response |
| 7 | TestGenerateUrlLink | urlscheme_test.go | POST path, body, url_link response |
| 8 | TestGenerateShortLink | urlscheme_test.go | POST path, body, link response |
| 9 | TestGetPhoneNumber | phone_test.go | POST path, code in body, phone_info response |
| 10 | TestGetDailySummary | analysis_test.go | POST path, date body, list response |
| 11 | TestGetVisitPage | analysis_test.go | POST path, date body, list response |
| 12 | TestGetDailyVisitTrend | analysis_test.go | POST path, date body, list response |
| 13 | TestUploadTempMedia | media_test.go | POST multipart, Content-Type, media field, JSON response |
| 14 | TestGetTempMedia | media_test.go | GET path, media_id query, binary response |

### 5.2 Test Helper Pattern

All tests use `NewClient(Config{AppId:"wx",AppSecret:"sec"}, WithHTTP(utils.NewHTTP(srv.URL, ...)))` — same as existing `client_test.go`. AccessToken is seeded by having the mock server respond to `/cgi-bin/token` with a valid token, or by using `WithTokenSource(fakeTokenSource)`.

### 5.3 Coverage Target

≥ 85%

---

## 6. Implementation Order

1. `helper.go` (foundation for all new methods)
2. `wxacode.go` + `wxacode_test.go` (tests doPostRaw)
3. `security.go` + `security_test.go`
4. `urlscheme.go` + `urlscheme_test.go`
5. `phone.go` + `phone_test.go`
6. `analysis.go` + `analysis_test.go`
7. `media.go` + `media_test.go` (multipart, last)
