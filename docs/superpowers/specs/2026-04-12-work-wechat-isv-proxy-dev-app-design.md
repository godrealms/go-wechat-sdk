# Sub-project 5: 代开发自建应用 — Design Spec

**Package:** `github.com/godrealms/go-wechat-sdk/work-wechat/isv`
**Date:** 2026-04-12
**Depends on:** Sub-project 1 (认证底座), Sub-project 4 (CorpClient HTTP helpers)

---

## 1. Scope

Three areas of functionality, all on `*CorpClient`:

1. **应用管理** — GetAgent (获取应用详情) + SetAgent (设置应用)
2. **自定义菜单** — CreateMenu + GetMenu + DeleteMenu
3. **素材上传** — UploadMedia (multipart/form-data) + doUpload private helper

### 1.1 Method Summary

| # | Method | HTTP | Path | Notes |
|---|---|---|---|---|
| 1 | GetAgent | GET | `/cgi-bin/agent/get?agentid=xxx` | |
| 2 | SetAgent | POST | `/cgi-bin/agent/set` | JSON body |
| 3 | CreateMenu | POST | `/cgi-bin/menu/create?agentid=xxx` | JSON body + agentid query |
| 4 | GetMenu | GET | `/cgi-bin/menu/get?agentid=xxx` | |
| 5 | DeleteMenu | GET | `/cgi-bin/menu/delete?agentid=xxx` | |
| 6 | UploadMedia | POST | `/cgi-bin/media/upload?type=xxx` | multipart/form-data |

### 1.2 File Summary

- 1 new DTO file (`struct.agent.go`)
- 3 new implementation files (`corp.agent.go`, `corp.menu.go`, `corp.media.go`)
- 3 new test files
- 6 public methods + 1 private helper (`doUpload`)

---

## 2. Architecture

### 2.1 Method Placement

All 6 methods are on `*CorpClient`. They use `doPost`/`doGet` for JSON APIs and a new `doUpload` for multipart.

### 2.2 doUpload Helper

```go
// doUpload sends a multipart/form-data POST with access_token injected.
func (cc *CorpClient) doUpload(ctx context.Context, path string, extra url.Values, fieldName, fileName string, fileData io.Reader, out interface{}) error {
    tok, err := cc.AccessToken(ctx)
    if err != nil {
        return err
    }
    q := url.Values{"access_token": {tok}}
    for k, vs := range extra {
        q[k] = vs
    }

    var buf bytes.Buffer
    w := multipart.NewWriter(&buf)
    fw, err := w.CreateFormFile(fieldName, fileName)
    if err != nil {
        return fmt.Errorf("isv: create form file: %w", err)
    }
    if _, err := io.Copy(fw, fileData); err != nil {
        return fmt.Errorf("isv: copy file data: %w", err)
    }
    w.Close()

    fullURL := cc.parent.baseURL + path + "?" + q.Encode()

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, &buf)
    if err != nil {
        return fmt.Errorf("isv: new request: %w", err)
    }
    req.Header.Set("Content-Type", w.FormDataContentType())

    resp, err := cc.parent.http.Do(req)
    if err != nil {
        return fmt.Errorf("isv: upload: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("isv: read body: %w", err)
    }
    return decodeRaw(body, out)
}
```

Key points:
- Accepts `extra url.Values` for additional query params (e.g., `type=image`)
- Constructs multipart body in memory using `bytes.Buffer`
- Uses `cc.parent.baseURL` and `cc.parent.http` (same field names as `doRequestRaw`)
- Merges `access_token` + extra params into single query string
- Reads response body into `[]byte` then calls `decodeRaw` (matches `doRequestRaw` pattern)

### 2.3 File Layout

| File | Content |
|---|---|
| `struct.agent.go` (new) | All DTOs for agent, menu, and media |
| `corp.http.go` (modify) | Add doPostExtra helper |
| `corp.agent.go` (new) | GetAgent, SetAgent |
| `corp.agent_test.go` (new) | 2 tests |
| `corp.menu.go` (new) | CreateMenu, GetMenu, DeleteMenu |
| `corp.menu_test.go` (new) | 3 tests |
| `corp.media.go` (new) | doUpload, UploadMedia |
| `corp.media_test.go` (new) | 2 tests (happy path + WeixinError) |

---

## 3. DTOs

### 3.1 Agent Management

```go
// AgentDetail is the response from GetAgent.
type AgentDetail struct {
    AgentID            int          `json:"agentid"`
    Name               string       `json:"name"`
    Description        string       `json:"description"`
    SquareLogoURL      string       `json:"square_logo_url"`
    RoundLogoURL       string       `json:"round_logo_url"`
    HomeURL            string       `json:"home_url"`
    RedirectDomain     string       `json:"redirect_domain"`
    IsReportEnter      int          `json:"isreportenter"`
    ReportLocationFlag int          `json:"report_location_flag"`
    AllowUserInfos     AllowUsers   `json:"allow_userinfos"`
    AllowParties       AllowParties `json:"allow_partys"`
    AllowTags          AllowTags    `json:"allow_tags"`
}

// AllowUsers contains the list of users in the agent's visibility scope.
type AllowUsers struct {
    User []AllowUser `json:"user"`
}

// AllowUser represents a single user in the visibility scope.
type AllowUser struct {
    UserID string `json:"userid"`
}

// AllowParties contains the list of departments in the agent's visibility scope.
type AllowParties struct {
    PartyID []int `json:"partyid"`
}

// AllowTags contains the list of tags in the agent's visibility scope.
type AllowTags struct {
    TagID []int `json:"tagid"`
}

// SetAgentReq is the request body for SetAgent.
type SetAgentReq struct {
    AgentID            int    `json:"agentid"`
    Name               string `json:"name,omitempty"`
    Description        string `json:"description,omitempty"`
    LogoMediaID        string `json:"logo_mediaid,omitempty"`
    HomeURL            string `json:"home_url,omitempty"`
    RedirectDomain     string `json:"redirect_domain,omitempty"`
    IsReportEnter      *int   `json:"isreportenter,omitempty"`
    ReportLocationFlag *int   `json:"report_location_flag,omitempty"`
}
```

Note: `IsReportEnter` and `ReportLocationFlag` use `*int` so zero values can be explicitly sent (JSON `omitempty` skips zero-value ints but not nil pointers to int).

### 3.2 Custom Menu

```go
// MenuButton represents a single menu button (supports nesting via SubButton).
type MenuButton struct {
    Type      string       `json:"type,omitempty"`
    Name      string       `json:"name"`
    Key       string       `json:"key,omitempty"`
    URL       string       `json:"url,omitempty"`
    AppID     string       `json:"appid,omitempty"`
    PagePath  string       `json:"pagepath,omitempty"`
    SubButton []MenuButton `json:"sub_button,omitempty"`
}

// CreateMenuReq is the request body for CreateMenu.
type CreateMenuReq struct {
    Button []MenuButton `json:"button"`
}

// MenuResp is the response from GetMenu.
type MenuResp struct {
    Button []MenuButton `json:"button"`
}
```

### 3.3 Media Upload

```go
// UploadMediaResp is the response from UploadMedia.
type UploadMediaResp struct {
    Type      string `json:"type"`
    MediaID   string `json:"media_id"`
    CreatedAt string `json:"created_at"`
}
```

---

## 4. Method Implementations

### 4.1 GetAgent

```go
// GetAgent retrieves the details of an agent (application) by ID.
func (cc *CorpClient) GetAgent(ctx context.Context, agentID int) (*AgentDetail, error) {
    extra := url.Values{"agentid": {strconv.Itoa(agentID)}}
    var resp AgentDetail
    if err := cc.doGet(ctx, "/cgi-bin/agent/get", extra, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.2 SetAgent

```go
// SetAgent updates an agent's properties (name, description, homepage, etc.).
func (cc *CorpClient) SetAgent(ctx context.Context, req *SetAgentReq) error {
    return cc.doPost(ctx, "/cgi-bin/agent/set", req, nil)
}
```

### 4.3 CreateMenu

`CreateMenu` needs both a JSON body AND an extra `agentid` query param. The existing `doPost` only injects `access_token`. Add a private `doPostExtra` helper in `corp.http.go`:

```go
// doPostExtra is like doPost but merges extra query params alongside access_token.
func (cc *CorpClient) doPostExtra(ctx context.Context, path string, extra url.Values, body, out interface{}) error {
    tok, err := cc.AccessToken(ctx)
    if err != nil {
        return err
    }
    q := url.Values{"access_token": {tok}}
    for k, vs := range extra {
        q[k] = vs
    }
    return cc.parent.doPostRaw(ctx, path, q, body, out)
}
```

Then:

```go
// CreateMenu creates a custom menu for the specified agent.
func (cc *CorpClient) CreateMenu(ctx context.Context, agentID int, req *CreateMenuReq) error {
    extra := url.Values{"agentid": {strconv.Itoa(agentID)}}
    return cc.doPostExtra(ctx, "/cgi-bin/menu/create", extra, req, nil)
}
```

### 4.4 GetMenu

```go
// GetMenu retrieves the current custom menu for the specified agent.
func (cc *CorpClient) GetMenu(ctx context.Context, agentID int) (*MenuResp, error) {
    extra := url.Values{"agentid": {strconv.Itoa(agentID)}}
    var resp MenuResp
    if err := cc.doGet(ctx, "/cgi-bin/menu/get", extra, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.5 DeleteMenu

```go
// DeleteMenu deletes the custom menu for the specified agent.
func (cc *CorpClient) DeleteMenu(ctx context.Context, agentID int) error {
    extra := url.Values{"agentid": {strconv.Itoa(agentID)}}
    return cc.doGet(ctx, "/cgi-bin/menu/delete", extra, nil)
}
```

### 4.6 UploadMedia

```go
// UploadMedia uploads a temporary media file (image, voice, video, or file).
func (cc *CorpClient) UploadMedia(ctx context.Context, mediaType, fileName string, fileData io.Reader) (*UploadMediaResp, error) {
    extra := url.Values{"type": {mediaType}}
    var resp UploadMediaResp
    if err := cc.doUpload(ctx, "/cgi-bin/media/upload", extra, "media", fileName, fileData, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

---

## 5. Testing

### 5.1 Test Matrix

| # | Test Function | File | Validates |
|---|---|---|---|
| 1 | TestGetAgent | corp.agent_test.go | GET path, agentid query, access_token, response parsing (AgentDetail fields) |
| 2 | TestSetAgent | corp.agent_test.go | POST path, access_token, request body JSON, *int fields serialization |
| 3 | TestCreateMenu | corp.menu_test.go | POST path with agentid query, access_token, button array in body |
| 4 | TestGetMenu | corp.menu_test.go | GET path, agentid query, response parsing (MenuResp with SubButton nesting) |
| 5 | TestDeleteMenu | corp.menu_test.go | GET path, agentid query, no error on success |
| 6 | TestUploadMedia | corp.media_test.go | POST multipart, Content-Type contains "multipart/form-data", type query param, field name "media", file content, response parsing |
| 7 | TestUploadMedia_WeixinError | corp.media_test.go | Upload returning errcode!=0, verify *WeixinError with errors.As |

### 5.2 Test Patterns

- All tests use `newTestCorpClient(t, srv.URL)` with pre-seeded token "CTOK"
- JSON API tests follow the established pattern from `corp.department_test.go` / `corp.user_test.go`
- Upload test: httptest server reads multipart form, verifies field name and file content
- WeixinError test: httptest server returns `{"errcode": 40004, "errmsg": "invalid media type"}`, verify `errors.As(err, &we)` and `we.ErrCode == 40004`

### 5.3 Coverage Target

≥ 85% for the package after changes.

---

## 6. Implementation Order

1. Create `struct.agent.go` with all DTOs (AgentDetail, SetAgentReq, MenuButton, CreateMenuReq, MenuResp, UploadMediaResp)
2. Create `corp.agent.go` with GetAgent + SetAgent, add tests in `corp.agent_test.go`
3. Create `corp.menu.go` with CreateMenu + GetMenu + DeleteMenu, add tests in `corp.menu_test.go`
4. Create `corp.media.go` with doUpload + UploadMedia, add tests in `corp.media_test.go`
