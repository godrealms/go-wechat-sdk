# Sub-project C: Stub Package Implementations Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fully implement three stub packages: aispeech (7 methods), mini-store (24 methods), and xiaowei (12 methods), each following the channels/mini-program architecture pattern.

**Architecture:** Each package has Config+NewClient with token caching, a helper.go with doPost, feature files per domain, and test files with httptest mocking. Godoc written inline.

**Tech Stack:** Go 1.23.1, net/http/httptest for tests

---

## Reference Architecture

All three packages use the same architecture as `channels/`. Key invariants:

- `client.go`: `Config`, `TokenSource` interface, `Client` struct (cfg + http + sync.RWMutex + accessToken + expiresAt + tokenSource), `Option` type, `WithHTTP`, `WithTokenSource`, `NewClient`, `AccessToken`.
- `helper.go`: `baseResp` struct, `doPost` method — always decodes to `json.RawMessage` first, checks `errcode != 0`, then unmarshals into `out`.
- Feature files: one file per domain, containing request/response struct types and one method per API endpoint.
- Test files: `_test.go` next to each feature file, using `httptest.NewServer` with a switch on `r.URL.Path` routing `/cgi-bin/token` and the API path, table-driven tests covering success and `errcode != 0` cases.

### Shared test helper pattern

Every package defines a `newTestClient` helper in `client_test.go`:

```go
func newTestClient(t *testing.T, baseURL string) *Client {
    t.Helper()
    c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
        WithHTTP(utils.NewHTTP(baseURL, utils.WithTimeout(3*time.Second))))
    if err != nil {
        t.Fatal(err)
    }
    return c
}
```

### Shared fakeTokenSource pattern

```go
type fakeTokenSource struct{ token string; err error; calls int }
func (f *fakeTokenSource) AccessToken(_ context.Context) (string, error) {
    f.calls++; return f.token, f.err
}
```

---

## C1: `aispeech` package

**Package name:** `aispeech`
**Module path:** `github.com/godrealms/go-wechat-sdk/aispeech`
**Base URL:** `https://openai.weixin.qq.com`
**WeChat doc:** 微信AI语音 (ASR, TTS, NLU, dialog)

---

### Task C1-1: `aispeech/client.go` + `aispeech/helper.go`

**Steps:**

- [ ] Replace the stub `aispeech/client.go` with the full implementation below. This is the complete file; copy exactly:

```go
// Package aispeech provides a client for the WeChat AI Speech (智能对话) API,
// covering automatic speech recognition (ASR), text-to-speech (TTS),
// natural language understanding (NLU), and dialog management.
// Create a Client with NewClient; token refresh is automatic.
package aispeech

import (
    "context"
    "fmt"
    "net/url"
    "sync"
    "time"

    "github.com/godrealms/go-wechat-sdk/utils"
)

// Config holds the aispeech app credentials.
type Config struct {
    AppId     string
    AppSecret string
}

// TokenSource is an injectable access_token provider. Configure it via
// WithTokenSource to delegate token management (e.g. open-platform flows)
// without calling /cgi-bin/token.
type TokenSource interface {
    AccessToken(ctx context.Context) (string, error)
}

// Client manages the aispeech API. Safe for concurrent use.
type Client struct {
    cfg         Config
    http        *utils.HTTP
    mu          sync.RWMutex
    accessToken string
    expiresAt   time.Time
    tokenSource TokenSource
}

// Option is a functional option applied during NewClient.
type Option func(*Client)

// WithHTTP injects a custom HTTP client, primarily for testing.
func WithHTTP(h *utils.HTTP) Option { return func(c *Client) { c.http = h } }

// WithTokenSource injects an external access_token provider.
// When set, AccessToken() delegates to it without calling /cgi-bin/token.
func WithTokenSource(ts TokenSource) Option { return func(c *Client) { c.tokenSource = ts } }

// NewClient constructs an aispeech client. Returns an error if AppId is
// empty or if AppSecret is empty and no TokenSource is provided.
func NewClient(cfg Config, opts ...Option) (*Client, error) {
    if cfg.AppId == "" {
        return nil, fmt.Errorf("aispeech: AppId is required")
    }
    c := &Client{
        cfg:  cfg,
        http: utils.NewHTTP("https://openai.weixin.qq.com", utils.WithTimeout(30*time.Second)),
    }
    for _, o := range opts {
        o(c)
    }
    if c.tokenSource == nil && cfg.AppSecret == "" {
        return nil, fmt.Errorf("aispeech: AppSecret is required when no TokenSource is injected")
    }
    return c, nil
}

// HTTP returns the underlying HTTP client for making custom requests.
func (c *Client) HTTP() *utils.HTTP { return c.http }

type accessTokenResp struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int64  `json:"expires_in"`
    ErrCode     int    `json:"errcode,omitempty"`
    ErrMsg      string `json:"errmsg,omitempty"`
}

// AccessToken returns a valid access_token, refreshing 60 s before expiry.
// When a TokenSource is configured, the call is forwarded to it.
func (c *Client) AccessToken(ctx context.Context) (string, error) {
    if c.tokenSource != nil {
        return c.tokenSource.AccessToken(ctx)
    }
    c.mu.RLock()
    if c.accessToken != "" && time.Now().Before(c.expiresAt) {
        t := c.accessToken
        c.mu.RUnlock()
        return t, nil
    }
    c.mu.RUnlock()
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.accessToken != "" && time.Now().Before(c.expiresAt) {
        return c.accessToken, nil
    }
    q := url.Values{
        "grant_type": {"client_credential"},
        "appid":      {c.cfg.AppId},
        "secret":     {c.cfg.AppSecret},
    }
    out := &accessTokenResp{}
    if err := c.http.Get(ctx, "/cgi-bin/token", q, out); err != nil {
        return "", fmt.Errorf("aispeech: fetch token: %w", err)
    }
    if out.ErrCode != 0 {
        return "", fmt.Errorf("aispeech: token errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
    }
    if out.AccessToken == "" {
        return "", fmt.Errorf("aispeech: empty access_token")
    }
    c.accessToken = out.AccessToken
    c.expiresAt = time.Now().Add(time.Duration(out.ExpiresIn-60) * time.Second)
    return c.accessToken, nil
}
```

- [ ] Create `aispeech/helper.go` with the following content:

```go
package aispeech

import (
    "context"
    "encoding/json"
    "fmt"
    "net/url"
)

// baseResp holds the common WeChat error fields present in every API response.
type baseResp struct {
    ErrCode int    `json:"errcode"`
    ErrMsg  string `json:"errmsg"`
}

// doPost sends a POST JSON request to path with access_token in the query,
// always checks errcode before unmarshalling into out.
func (c *Client) doPost(ctx context.Context, path string, body any, out any) error {
    tok, err := c.AccessToken(ctx)
    if err != nil {
        return err
    }
    q := url.Values{"access_token": {tok}}
    fullPath := path + "?" + q.Encode()
    var raw json.RawMessage
    if err := c.http.Post(ctx, fullPath, body, &raw); err != nil {
        return err
    }
    var base baseResp
    _ = json.Unmarshal(raw, &base)
    if base.ErrCode != 0 {
        return fmt.Errorf("aispeech: %s errcode=%d errmsg=%s", path, base.ErrCode, base.ErrMsg)
    }
    if out != nil {
        return json.Unmarshal(raw, out)
    }
    return nil
}
```

- [ ] Create `aispeech/client_test.go` with `TestNewClient`, `TestAccessToken` (caching: 3 calls → 1 HTTP hit), `TestAccessToken_UsesInjectedTokenSource`, and `newTestClient`/`fakeTokenSource` helpers:

```go
package aispeech

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/godrealms/go-wechat-sdk/utils"
)

func newTestClient(t *testing.T, baseURL string) *Client {
    t.Helper()
    c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
        WithHTTP(utils.NewHTTP(baseURL, utils.WithTimeout(3*time.Second))))
    if err != nil {
        t.Fatal(err)
    }
    return c
}

type fakeTokenSource struct {
    token string
    err   error
    calls int
}

func (f *fakeTokenSource) AccessToken(_ context.Context) (string, error) {
    f.calls++
    return f.token, f.err
}

func TestNewClient(t *testing.T) {
    if _, err := NewClient(Config{}); err == nil {
        t.Error("expected error for empty AppId")
    }
    if _, err := NewClient(Config{AppId: "wx"}); err == nil {
        t.Error("expected error for empty AppSecret without TokenSource")
    }
    c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"})
    if err != nil {
        t.Fatal(err)
    }
    if c.HTTP() == nil {
        t.Error("HTTP() must not be nil")
    }
}

func TestAccessToken(t *testing.T) {
    var calls int
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        calls++
        _, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
    }))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    for i := 0; i < 3; i++ {
        tok, err := c.AccessToken(context.Background())
        if err != nil {
            t.Fatal(err)
        }
        if tok != "TOK" {
            t.Errorf("got %q, want TOK", tok)
        }
    }
    if calls != 1 {
        t.Errorf("expected 1 fetch, got %d", calls)
    }
}

func TestAccessToken_UsesInjectedTokenSource(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        t.Errorf("must not call /cgi-bin/token when TokenSource injected: %s", r.URL.Path)
    }))
    defer srv.Close()
    fake := &fakeTokenSource{token: "INJECTED"}
    c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
        WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))),
        WithTokenSource(fake))
    if err != nil {
        t.Fatal(err)
    }
    tok, err := c.AccessToken(context.Background())
    if err != nil {
        t.Fatal(err)
    }
    if tok != "INJECTED" {
        t.Errorf("got %q, want INJECTED", tok)
    }
    if fake.calls != 1 {
        t.Errorf("expected 1 call, got %d", fake.calls)
    }
}
```

- [ ] Run `go build ./aispeech/...` and `go test ./aispeech/...` — must pass.

- [ ] Commit:

```bash
git add aispeech/client.go aispeech/helper.go aispeech/client_test.go
git commit -m "feat(aispeech): implement client, token caching, helper doPost"
```

---

### Task C1-2: `aispeech/asr.go` + `aispeech/asr_test.go`

WeChat API paths:
- `POST /aispeech/asr/aiasrlong` — long audio ASR (async)
- `POST /aispeech/asr/aiasrshort` — short audio ASR (sync, ≤60 s)

**Steps:**

- [ ] Create `aispeech/asr.go`:

```go
package aispeech

import "context"

// ASRLongReq is the request for long-audio asynchronous speech recognition.
type ASRLongReq struct {
    VoiceID   string `json:"voice_id"`            // caller-assigned unique ID for this job
    VoiceURL  string `json:"voice_url"`            // URL of the audio file (mp3/wav/amr, ≤300 s)
    Format    string `json:"voice_format"`         // audio format: mp3 | wav | amr
    Lang      string `json:"lang,omitempty"`       // language: zh_CN (default) | en_US
    CallbackURL string `json:"callback_url,omitempty"` // optional callback on completion
}

// ASRLongResp is the response from ASRLong. The task is queued
// asynchronously; the final transcript arrives via callback or polling.
type ASRLongResp struct {
    TaskID string `json:"task_id"` // opaque identifier for polling
}

// ASRLong submits a long-audio speech recognition job (≤300 s).
// Results are delivered asynchronously; use TaskID to poll for status.
func (c *Client) ASRLong(ctx context.Context, req *ASRLongReq) (*ASRLongResp, error) {
    var resp ASRLongResp
    if err := c.doPost(ctx, "/aispeech/asr/aiasrlong", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// ASRShortReq is the request for short-audio synchronous speech recognition.
type ASRShortReq struct {
    VoiceID  string `json:"voice_id"`        // caller-assigned unique ID
    VoiceData string `json:"voice_data"`     // base64-encoded audio data (≤1 MB)
    Format   string `json:"voice_format"`    // audio format: mp3 | wav | amr | speex
    Rate     int    `json:"voice_rate"`      // sample rate in Hz, e.g. 16000
    Bits     int    `json:"voice_bits"`      // bit depth, e.g. 16
    Lang     string `json:"lang,omitempty"`  // language: zh_CN (default) | en_US
}

// ASRShortResp is the response from ASRShort, containing the transcript.
type ASRShortResp struct {
    Result string `json:"result"` // recognized text
}

// ASRShort performs synchronous speech recognition on audio data ≤60 s.
// The recognized text is returned in the response.
func (c *Client) ASRShort(ctx context.Context, req *ASRShortReq) (*ASRShortResp, error) {
    var resp ASRShortResp
    if err := c.doPost(ctx, "/aispeech/asr/aiasrshort", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

- [ ] Create `aispeech/asr_test.go` with table-driven tests:

```go
package aispeech

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestASRLong(t *testing.T) {
    tests := []struct {
        name    string
        handler func(w http.ResponseWriter, r *http.Request)
        wantErr bool
        wantID  string
    }{
        {
            name: "success",
            handler: func(w http.ResponseWriter, r *http.Request) {
                switch r.URL.Path {
                case "/cgi-bin/token":
                    _, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
                case "/aispeech/asr/aiasrlong":
                    var req ASRLongReq
                    _ = json.NewDecoder(r.Body).Decode(&req)
                    _, _ = w.Write([]byte(`{"task_id":"task_001","errcode":0,"errmsg":"ok"}`))
                }
            },
            wantErr: false,
            wantID:  "task_001",
        },
        {
            name: "api_error",
            handler: func(w http.ResponseWriter, r *http.Request) {
                switch r.URL.Path {
                case "/cgi-bin/token":
                    _, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
                case "/aispeech/asr/aiasrlong":
                    _, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"invalid credential"}`))
                }
            },
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            srv := httptest.NewServer(http.HandlerFunc(tt.handler))
            defer srv.Close()
            c := newTestClient(t, srv.URL)
            resp, err := c.ASRLong(context.Background(), &ASRLongReq{
                VoiceID:  "v1",
                VoiceURL: "https://example.com/audio.mp3",
                Format:   "mp3",
            })
            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
            }
            if !tt.wantErr && resp.TaskID != tt.wantID {
                t.Errorf("got task_id=%s, want %s", resp.TaskID, tt.wantID)
            }
        })
    }
}

func TestASRShort(t *testing.T) {
    tests := []struct {
        name       string
        handler    func(w http.ResponseWriter, r *http.Request)
        wantErr    bool
        wantResult string
    }{
        {
            name: "success",
            handler: func(w http.ResponseWriter, r *http.Request) {
                switch r.URL.Path {
                case "/cgi-bin/token":
                    _, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
                case "/aispeech/asr/aiasrshort":
                    _, _ = w.Write([]byte(`{"result":"你好世界","errcode":0,"errmsg":"ok"}`))
                }
            },
            wantErr:    false,
            wantResult: "你好世界",
        },
        {
            name: "api_error",
            handler: func(w http.ResponseWriter, r *http.Request) {
                switch r.URL.Path {
                case "/cgi-bin/token":
                    _, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
                case "/aispeech/asr/aiasrshort":
                    _, _ = w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
                }
            },
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            srv := httptest.NewServer(http.HandlerFunc(tt.handler))
            defer srv.Close()
            c := newTestClient(t, srv.URL)
            resp, err := c.ASRShort(context.Background(), &ASRShortReq{
                VoiceID:   "v1",
                VoiceData: "base64data==",
                Format:    "wav",
                Rate:      16000,
                Bits:      16,
            })
            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
            }
            if !tt.wantErr && resp.Result != tt.wantResult {
                t.Errorf("got result=%q, want %q", resp.Result, tt.wantResult)
            }
        })
    }
}
```

- [ ] Run `go test ./aispeech/... -run TestASR` — both tests pass.

- [ ] Commit:

```bash
git add aispeech/asr.go aispeech/asr_test.go
git commit -m "feat(aispeech): implement ASRLong and ASRShort with tests"
```

---

### Task C1-3: `aispeech/tts.go` + `aispeech/tts_test.go`

WeChat API path:
- `POST /aispeech/tts/aitts` — text-to-speech synthesis

**Steps:**

- [ ] Create `aispeech/tts.go`:

```go
package aispeech

import "context"

// TextToSpeechReq is the request for TTS synthesis.
type TextToSpeechReq struct {
    Text   string `json:"text"`              // plain text to synthesize (≤300 chars)
    Speed  int    `json:"speed,omitempty"`   // speech rate: -500 to 500 (0=default)
    Volume int    `json:"volume,omitempty"`  // volume: -10 to 10 (0=default)
    Pitch  int    `json:"pitch,omitempty"`   // pitch: -500 to 500 (0=default)
    VoiceType int `json:"voice_type,omitempty"` // voice ID (0=default female Mandarin)
}

// TextToSpeechResp is the response from TextToSpeech.
// AudioData contains the synthesized audio as a base64-encoded MP3.
type TextToSpeechResp struct {
    AudioData  string `json:"audio_data"`   // base64-encoded MP3 audio
    AudioSize  int    `json:"audio_size"`   // size in bytes before encoding
    SessionID  string `json:"session_id"`   // opaque session identifier
}

// TextToSpeech converts text to speech and returns the audio as base64 MP3.
// Text must be ≤300 characters; longer input should be split by the caller.
func (c *Client) TextToSpeech(ctx context.Context, req *TextToSpeechReq) (*TextToSpeechResp, error) {
    var resp TextToSpeechResp
    if err := c.doPost(ctx, "/aispeech/tts/aitts", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

- [ ] Create `aispeech/tts_test.go`:

```go
package aispeech

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestTextToSpeech(t *testing.T) {
    tests := []struct {
        name      string
        handler   func(w http.ResponseWriter, r *http.Request)
        wantErr   bool
        wantAudio string
    }{
        {
            name: "success",
            handler: func(w http.ResponseWriter, r *http.Request) {
                switch r.URL.Path {
                case "/cgi-bin/token":
                    _, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
                case "/aispeech/tts/aitts":
                    if r.Method != http.MethodPost {
                        t.Errorf("expected POST, got %s", r.Method)
                    }
                    if r.URL.Query().Get("access_token") != "TOK" {
                        t.Error("missing access_token")
                    }
                    _, _ = w.Write([]byte(`{"audio_data":"AAEC","audio_size":3,"session_id":"s1","errcode":0,"errmsg":"ok"}`))
                }
            },
            wantErr:   false,
            wantAudio: "AAEC",
        },
        {
            name: "api_error",
            handler: func(w http.ResponseWriter, r *http.Request) {
                switch r.URL.Path {
                case "/cgi-bin/token":
                    _, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
                case "/aispeech/tts/aitts":
                    _, _ = w.Write([]byte(`{"errcode":45009,"errmsg":"reach max api daily quota limit"}`))
                }
            },
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            srv := httptest.NewServer(http.HandlerFunc(tt.handler))
            defer srv.Close()
            c := newTestClient(t, srv.URL)
            resp, err := c.TextToSpeech(context.Background(), &TextToSpeechReq{Text: "你好"})
            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
            }
            if !tt.wantErr && resp.AudioData != tt.wantAudio {
                t.Errorf("got audio_data=%q, want %q", resp.AudioData, tt.wantAudio)
            }
        })
    }
}
```

- [ ] Run `go test ./aispeech/... -run TestTextToSpeech` — passes.

- [ ] Commit:

```bash
git add aispeech/tts.go aispeech/tts_test.go
git commit -m "feat(aispeech): implement TextToSpeech with tests"
```

---

### Task C1-4: `aispeech/nlu.go` + `aispeech/nlu_test.go`

WeChat API paths:
- `POST /aispeech/nlu/airequ` — natural language understanding
- `POST /aispeech/nlu/aiintentrequ` — intent recognition

**Steps:**

- [ ] Create `aispeech/nlu.go`:

```go
package aispeech

import "context"

// NLUUnderstandReq is the request for natural language understanding.
type NLUUnderstandReq struct {
    Query     string `json:"query"`                // input text to analyse (≤512 chars)
    SessionID string `json:"session_id,omitempty"` // conversation session for context
    Lang      string `json:"lang,omitempty"`       // language code; default zh_CN
}

// NLUEntity is a recognized named entity within the query.
type NLUEntity struct {
    Type  string `json:"type"`  // entity type, e.g. "PERSON", "LOCATION"
    Value string `json:"value"` // matched text
    Begin int    `json:"begin"` // start byte offset
    End   int    `json:"end"`   // end byte offset (exclusive)
}

// NLUUnderstandResp is the response from NLUUnderstand.
type NLUUnderstandResp struct {
    Intent    string      `json:"intent"`    // top-level intent label
    Slots     []NLUEntity `json:"slots"`     // named entities extracted from the query
    SessionID string      `json:"session_id"`
}

// NLUUnderstand extracts intent and named entities from the input text.
func (c *Client) NLUUnderstand(ctx context.Context, req *NLUUnderstandReq) (*NLUUnderstandResp, error) {
    var resp NLUUnderstandResp
    if err := c.doPost(ctx, "/aispeech/nlu/airequ", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// NLUIntentRecognizeReq is the request for intent classification.
type NLUIntentRecognizeReq struct {
    Query     string   `json:"query"`                // input text
    IntentIDs []string `json:"intent_ids,omitempty"` // restrict to these intent IDs
    SessionID string   `json:"session_id,omitempty"`
}

// NLUIntentRecognizeResp is the response from NLUIntentRecognize.
type NLUIntentRecognizeResp struct {
    IntentID   string  `json:"intent_id"`   // matched intent identifier
    IntentName string  `json:"intent_name"` // human-readable intent label
    Confidence float64 `json:"confidence"`  // score in [0, 1]
    SessionID  string  `json:"session_id"`
}

// NLUIntentRecognize classifies the query against a configured intent set.
// It returns the best-matching intent and a confidence score.
func (c *Client) NLUIntentRecognize(ctx context.Context, req *NLUIntentRecognizeReq) (*NLUIntentRecognizeResp, error) {
    var resp NLUIntentRecognizeResp
    if err := c.doPost(ctx, "/aispeech/nlu/aiintentrequ", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

- [ ] Create `aispeech/nlu_test.go`:

```go
package aispeech

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func handler200(path, body string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/cgi-bin/token":
            _, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
        case path:
            _, _ = w.Write([]byte(body))
        default:
            w.WriteHeader(http.StatusNotFound)
        }
    }
}

func TestNLUUnderstand(t *testing.T) {
    tests := []struct {
        name       string
        respBody   string
        wantErr    bool
        wantIntent string
    }{
        {
            name:       "success",
            respBody:   `{"intent":"weather_query","slots":[{"type":"LOCATION","value":"北京","begin":2,"end":4}],"session_id":"s1","errcode":0,"errmsg":"ok"}`,
            wantErr:    false,
            wantIntent: "weather_query",
        },
        {
            name:     "api_error",
            respBody: `{"errcode":40001,"errmsg":"invalid credential"}`,
            wantErr:  true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            srv := httptest.NewServer(handler200("/aispeech/nlu/airequ", tt.respBody))
            defer srv.Close()
            c := newTestClient(t, srv.URL)
            resp, err := c.NLUUnderstand(context.Background(), &NLUUnderstandReq{Query: "北京天气"})
            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
            }
            if !tt.wantErr && resp.Intent != tt.wantIntent {
                t.Errorf("got intent=%q, want %q", resp.Intent, tt.wantIntent)
            }
        })
    }
}

func TestNLUIntentRecognize(t *testing.T) {
    tests := []struct {
        name       string
        respBody   string
        wantErr    bool
        wantIntent string
    }{
        {
            name:       "success",
            respBody:   `{"intent_id":"i1","intent_name":"查天气","confidence":0.95,"session_id":"s1","errcode":0,"errmsg":"ok"}`,
            wantErr:    false,
            wantIntent: "i1",
        },
        {
            name:     "api_error",
            respBody: `{"errcode":40003,"errmsg":"invalid openid"}`,
            wantErr:  true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            srv := httptest.NewServer(handler200("/aispeech/nlu/aiintentrequ", tt.respBody))
            defer srv.Close()
            c := newTestClient(t, srv.URL)
            resp, err := c.NLUIntentRecognize(context.Background(), &NLUIntentRecognizeReq{Query: "查天气"})
            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
            }
            if !tt.wantErr && resp.IntentID != tt.wantIntent {
                t.Errorf("got intent_id=%q, want %q", resp.IntentID, tt.wantIntent)
            }
        })
    }
}
```

- [ ] Run `go test ./aispeech/... -run TestNLU` — passes.

- [ ] Commit:

```bash
git add aispeech/nlu.go aispeech/nlu_test.go
git commit -m "feat(aispeech): implement NLUUnderstand and NLUIntentRecognize with tests"
```

---

### Task C1-5: `aispeech/dialog.go` + `aispeech/dialog_test.go`

WeChat API paths:
- `POST /aispeech/dialog/airequ` — dialog query (multi-turn)
- `POST /aispeech/dialog/aireset` — dialog reset (end session)

**Steps:**

- [ ] Create `aispeech/dialog.go`:

```go
package aispeech

import "context"

// DialogQueryReq is the request for a multi-turn dialog query.
type DialogQueryReq struct {
    Query     string `json:"query"`      // current user utterance (≤512 chars)
    SessionID string `json:"session_id"` // conversation session; WeChat assigns on first turn
    Lang      string `json:"lang,omitempty"`
}

// DialogQueryResp is the response from DialogQuery.
type DialogQueryResp struct {
    Answer    string `json:"answer"`     // system reply text
    SessionID string `json:"session_id"` // session ID (same or updated)
    EndFlag   bool   `json:"end_flag"`   // true if dialog is complete
}

// DialogQuery sends a user utterance and returns the system reply.
// Pass the SessionID from the previous response to maintain context.
// An empty SessionID starts a new dialog session.
func (c *Client) DialogQuery(ctx context.Context, req *DialogQueryReq) (*DialogQueryResp, error) {
    var resp DialogQueryResp
    if err := c.doPost(ctx, "/aispeech/dialog/airequ", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// DialogResetReq is the request to terminate and reset a dialog session.
type DialogResetReq struct {
    SessionID string `json:"session_id"` // session to terminate
}

// DialogResetResp is the response from DialogReset.
type DialogResetResp struct {
    // WeChat returns only errcode/errmsg; no additional fields.
}

// DialogReset terminates the dialog session identified by SessionID,
// freeing any server-side conversation state.
func (c *Client) DialogReset(ctx context.Context, req *DialogResetReq) error {
    return c.doPost(ctx, "/aispeech/dialog/aireset", req, nil)
}
```

- [ ] Create `aispeech/dialog_test.go`:

```go
package aispeech

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestDialogQuery(t *testing.T) {
    tests := []struct {
        name       string
        respBody   string
        wantErr    bool
        wantAnswer string
    }{
        {
            name:       "success",
            respBody:   `{"answer":"今天北京天气晴","session_id":"s1","end_flag":false,"errcode":0,"errmsg":"ok"}`,
            wantErr:    false,
            wantAnswer: "今天北京天气晴",
        },
        {
            name:     "api_error",
            respBody: `{"errcode":40001,"errmsg":"invalid credential"}`,
            wantErr:  true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            srv := httptest.NewServer(handler200("/aispeech/dialog/airequ", tt.respBody))
            defer srv.Close()
            c := newTestClient(t, srv.URL)
            resp, err := c.DialogQuery(context.Background(), &DialogQueryReq{
                Query:     "北京天气",
                SessionID: "s1",
            })
            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
            }
            if !tt.wantErr && resp.Answer != tt.wantAnswer {
                t.Errorf("got answer=%q, want %q", resp.Answer, tt.wantAnswer)
            }
        })
    }
}

func TestDialogReset(t *testing.T) {
    tests := []struct {
        name     string
        respBody string
        wantErr  bool
    }{
        {
            name:     "success",
            respBody: `{"errcode":0,"errmsg":"ok"}`,
            wantErr:  false,
        },
        {
            name:     "api_error",
            respBody: `{"errcode":40001,"errmsg":"invalid credential"}`,
            wantErr:  true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            srv := httptest.NewServer(handler200("/aispeech/dialog/aireset", tt.respBody))
            defer srv.Close()
            c := newTestClient(t, srv.URL)
            err := c.DialogReset(context.Background(), &DialogResetReq{SessionID: "s1"})
            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
            }
        })
    }
}
```

- [ ] Run `go test ./aispeech/...` — all 9 test cases pass.

- [ ] Commit:

```bash
git add aispeech/dialog.go aispeech/dialog_test.go
git commit -m "feat(aispeech): implement DialogQuery and DialogReset with tests"
```

---

## C2: `mini-store` package

**Package name:** `mini_store`
**Module path:** `github.com/godrealms/go-wechat-sdk/mini-store`
**Base URL:** `https://api.weixin.qq.com`
**Path prefix:** `/shop/`
**WeChat doc:** 微信小店 (formerly 微信小商店)

---

### Task C2-6: `mini-store/client.go` + `mini-store/helper.go`

**Steps:**

- [ ] Replace the stub `mini-store/client.go` with:

```go
// Package mini_store provides a client for the WeChat Mini Store (微信小店) API,
// covering product management, order management, delivery, settlement,
// coupons, and after-sale service. Create a Client with NewClient.
package mini_store

import (
    "context"
    "fmt"
    "net/url"
    "sync"
    "time"

    "github.com/godrealms/go-wechat-sdk/utils"
)

// Config holds the Mini Store app credentials.
type Config struct {
    AppId     string
    AppSecret string
}

// TokenSource is an injectable access_token provider.
type TokenSource interface {
    AccessToken(ctx context.Context) (string, error)
}

// Client manages the Mini Store API. Safe for concurrent use.
type Client struct {
    cfg         Config
    http        *utils.HTTP
    mu          sync.RWMutex
    accessToken string
    expiresAt   time.Time
    tokenSource TokenSource
}

// Option is a functional option applied during NewClient.
type Option func(*Client)

// WithHTTP injects a custom HTTP client, primarily for testing.
func WithHTTP(h *utils.HTTP) Option { return func(c *Client) { c.http = h } }

// WithTokenSource injects an external access_token provider.
func WithTokenSource(ts TokenSource) Option { return func(c *Client) { c.tokenSource = ts } }

// NewClient constructs a Mini Store client. Returns an error if AppId is
// empty or if AppSecret is empty and no TokenSource is provided.
func NewClient(cfg Config, opts ...Option) (*Client, error) {
    if cfg.AppId == "" {
        return nil, fmt.Errorf("mini_store: AppId is required")
    }
    c := &Client{
        cfg:  cfg,
        http: utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(30*time.Second)),
    }
    for _, o := range opts {
        o(c)
    }
    if c.tokenSource == nil && cfg.AppSecret == "" {
        return nil, fmt.Errorf("mini_store: AppSecret is required when no TokenSource is injected")
    }
    return c, nil
}

// HTTP returns the underlying HTTP client for custom requests.
func (c *Client) HTTP() *utils.HTTP { return c.http }

type accessTokenResp struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int64  `json:"expires_in"`
    ErrCode     int    `json:"errcode,omitempty"`
    ErrMsg      string `json:"errmsg,omitempty"`
}

// AccessToken returns a valid access_token, refreshing 60 s before expiry.
func (c *Client) AccessToken(ctx context.Context) (string, error) {
    if c.tokenSource != nil {
        return c.tokenSource.AccessToken(ctx)
    }
    c.mu.RLock()
    if c.accessToken != "" && time.Now().Before(c.expiresAt) {
        t := c.accessToken
        c.mu.RUnlock()
        return t, nil
    }
    c.mu.RUnlock()
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.accessToken != "" && time.Now().Before(c.expiresAt) {
        return c.accessToken, nil
    }
    q := url.Values{
        "grant_type": {"client_credential"},
        "appid":      {c.cfg.AppId},
        "secret":     {c.cfg.AppSecret},
    }
    out := &accessTokenResp{}
    if err := c.http.Get(ctx, "/cgi-bin/token", q, out); err != nil {
        return "", fmt.Errorf("mini_store: fetch token: %w", err)
    }
    if out.ErrCode != 0 {
        return "", fmt.Errorf("mini_store: token errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
    }
    if out.AccessToken == "" {
        return "", fmt.Errorf("mini_store: empty access_token")
    }
    c.accessToken = out.AccessToken
    c.expiresAt = time.Now().Add(time.Duration(out.ExpiresIn-60) * time.Second)
    return c.accessToken, nil
}
```

- [ ] Create `mini-store/helper.go`:

```go
package mini_store

import (
    "context"
    "encoding/json"
    "fmt"
    "net/url"
)

type baseResp struct {
    ErrCode int    `json:"errcode"`
    ErrMsg  string `json:"errmsg"`
}

// doPost sends a POST JSON request with access_token, checks errcode,
// and unmarshals the response into out.
func (c *Client) doPost(ctx context.Context, path string, body any, out any) error {
    tok, err := c.AccessToken(ctx)
    if err != nil {
        return err
    }
    q := url.Values{"access_token": {tok}}
    fullPath := path + "?" + q.Encode()
    var raw json.RawMessage
    if err := c.http.Post(ctx, fullPath, body, &raw); err != nil {
        return err
    }
    var base baseResp
    _ = json.Unmarshal(raw, &base)
    if base.ErrCode != 0 {
        return fmt.Errorf("mini_store: %s errcode=%d errmsg=%s", path, base.ErrCode, base.ErrMsg)
    }
    if out != nil {
        return json.Unmarshal(raw, out)
    }
    return nil
}
```

- [ ] Create `mini-store/client_test.go` with `newTestClient`, `fakeTokenSource`, `TestNewClient`, `TestAccessToken`, `TestAccessToken_UsesInjectedTokenSource` (same pattern as aispeech).

- [ ] Run `go test ./mini-store/... -run TestNewClient` — passes.

- [ ] Commit:

```bash
git add mini-store/client.go mini-store/helper.go mini-store/client_test.go
git commit -m "feat(mini-store): implement client, token caching, helper doPost"
```

---

### Task C2-7: `mini-store/product.go` + `mini-store/product_test.go`

WeChat API paths under `/shop/spu/`:
- `POST /shop/spu/add` — AddProduct
- `POST /shop/spu/update` — UpdateProduct
- `POST /shop/spu/del` — DeleteProduct
- `POST /shop/spu/get` — GetProduct
- `POST /shop/spu/get_list` — ListProduct
- `POST /shop/spu/update_status` — UpdateProductStatus

**Steps:**

- [ ] Create `mini-store/product.go`:

```go
package mini_store

import "context"

// SpuInfo describes a Mini Store product (SPU = Standard Product Unit).
type SpuInfo struct {
    SpuID      string   `json:"spu_id,omitempty"`
    Title      string   `json:"title,omitempty"`
    SubTitle   string   `json:"sub_title,omitempty"`
    HeadImgUris []string `json:"head_img_uris,omitempty"`
    Status     int      `json:"status,omitempty"`  // 0=draft 1=listing 2=delisted
    CreateTime int64    `json:"create_time,omitempty"`
    UpdateTime int64    `json:"update_time,omitempty"`
}

// AddProductReq is the request body for AddProduct.
type AddProductReq struct {
    SpuInfo SpuInfo `json:"spu_info"`
}

// AddProductResp is the response from AddProduct.
type AddProductResp struct {
    SpuID string `json:"spu_id"` // WeChat-assigned product ID
}

// UpdateProductReq is the request body for UpdateProduct.
type UpdateProductReq struct {
    SpuID   string  `json:"spu_id"`
    SpuInfo SpuInfo `json:"spu_info"`
}

// DeleteProductReq is the request body for DeleteProduct.
type DeleteProductReq struct {
    SpuID string `json:"spu_id"`
}

// GetProductReq is the request body for GetProduct.
type GetProductReq struct {
    SpuID string `json:"spu_id"`
}

// GetProductResp is the response from GetProduct.
type GetProductResp struct {
    SpuInfo SpuInfo `json:"spu_info"`
}

// ListProductReq is the request body for ListProduct.
type ListProductReq struct {
    Status int `json:"status,omitempty"` // filter by status; 0 = all
    Page   int `json:"page,omitempty"`   // 1-based page number
    PageSize int `json:"page_size,omitempty"` // items per page (max 100)
}

// ListProductResp is the response from ListProduct.
type ListProductResp struct {
    SpuList  []SpuInfo `json:"spu_list"`
    TotalNum int       `json:"total_num"`
}

// UpdateProductStatusReq is the request body for UpdateProductStatus.
type UpdateProductStatusReq struct {
    SpuID  string `json:"spu_id"`
    Status int    `json:"status"` // 1=list 2=delist
}

// AddProduct creates a new product in the Mini Store catalog.
// Returns the WeChat-assigned spu_id on success.
func (c *Client) AddProduct(ctx context.Context, req *AddProductReq) (*AddProductResp, error) {
    var resp AddProductResp
    if err := c.doPost(ctx, "/shop/spu/add", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// UpdateProduct updates an existing product. SpuID must be set in req.
func (c *Client) UpdateProduct(ctx context.Context, req *UpdateProductReq) error {
    return c.doPost(ctx, "/shop/spu/update", req, nil)
}

// DeleteProduct permanently removes the product identified by SpuID.
func (c *Client) DeleteProduct(ctx context.Context, req *DeleteProductReq) error {
    return c.doPost(ctx, "/shop/spu/del", req, nil)
}

// GetProduct retrieves the full product details for the given spu_id.
func (c *Client) GetProduct(ctx context.Context, req *GetProductReq) (*GetProductResp, error) {
    var resp GetProductResp
    if err := c.doPost(ctx, "/shop/spu/get", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// ListProduct returns a paginated list of products, optionally filtered by status.
func (c *Client) ListProduct(ctx context.Context, req *ListProductReq) (*ListProductResp, error) {
    var resp ListProductResp
    if err := c.doPost(ctx, "/shop/spu/get_list", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// UpdateProductStatus changes the listing status of the product identified by SpuID.
// Pass Status=1 to list and Status=2 to delist.
func (c *Client) UpdateProductStatus(ctx context.Context, req *UpdateProductStatusReq) error {
    return c.doPost(ctx, "/shop/spu/update_status", req, nil)
}
```

- [ ] Create `mini-store/product_test.go` with table-driven tests for all six methods. Each test table has a "success" row and an "api_error" row:

```go
package mini_store

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func shopHandler(path, successBody string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/cgi-bin/token":
            _, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
        case path:
            if r.Method != http.MethodPost {
                w.WriteHeader(http.StatusMethodNotAllowed)
                return
            }
            _, _ = w.Write([]byte(successBody))
        default:
            w.WriteHeader(http.StatusNotFound)
        }
    }
}

func TestAddProduct(t *testing.T) {
    srv := httptest.NewServer(shopHandler("/shop/spu/add", `{"spu_id":"spu_001","errcode":0,"errmsg":"ok"}`))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    resp, err := c.AddProduct(context.Background(), &AddProductReq{
        SpuInfo: SpuInfo{Title: "Test Product"},
    })
    if err != nil {
        t.Fatal(err)
    }
    if resp.SpuID != "spu_001" {
        t.Errorf("got spu_id=%s, want spu_001", resp.SpuID)
    }
}

func TestAddProduct_Error(t *testing.T) {
    srv := httptest.NewServer(shopHandler("/shop/spu/add", `{"errcode":40001,"errmsg":"invalid credential"}`))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    _, err := c.AddProduct(context.Background(), &AddProductReq{})
    if err == nil {
        t.Fatal("expected error")
    }
}

func TestUpdateProduct(t *testing.T) {
    srv := httptest.NewServer(shopHandler("/shop/spu/update", `{"errcode":0,"errmsg":"ok"}`))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    err := c.UpdateProduct(context.Background(), &UpdateProductReq{SpuID: "spu_001", SpuInfo: SpuInfo{Title: "New Title"}})
    if err != nil {
        t.Fatal(err)
    }
}

func TestDeleteProduct(t *testing.T) {
    srv := httptest.NewServer(shopHandler("/shop/spu/del", `{"errcode":0,"errmsg":"ok"}`))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    if err := c.DeleteProduct(context.Background(), &DeleteProductReq{SpuID: "spu_001"}); err != nil {
        t.Fatal(err)
    }
}

func TestGetProduct(t *testing.T) {
    srv := httptest.NewServer(shopHandler("/shop/spu/get", `{"spu_info":{"spu_id":"spu_001","title":"Test Product","status":1},"errcode":0,"errmsg":"ok"}`))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    resp, err := c.GetProduct(context.Background(), &GetProductReq{SpuID: "spu_001"})
    if err != nil {
        t.Fatal(err)
    }
    if resp.SpuInfo.SpuID != "spu_001" {
        t.Errorf("got spu_id=%s, want spu_001", resp.SpuInfo.SpuID)
    }
}

func TestListProduct(t *testing.T) {
    srv := httptest.NewServer(shopHandler("/shop/spu/get_list", `{"spu_list":[{"spu_id":"spu_001"},{"spu_id":"spu_002"}],"total_num":2,"errcode":0,"errmsg":"ok"}`))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    resp, err := c.ListProduct(context.Background(), &ListProductReq{Page: 1, PageSize: 20})
    if err != nil {
        t.Fatal(err)
    }
    if resp.TotalNum != 2 {
        t.Errorf("got total_num=%d, want 2", resp.TotalNum)
    }
    if len(resp.SpuList) != 2 {
        t.Errorf("got %d products, want 2", len(resp.SpuList))
    }
}

func TestUpdateProductStatus(t *testing.T) {
    srv := httptest.NewServer(shopHandler("/shop/spu/update_status", `{"errcode":0,"errmsg":"ok"}`))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    if err := c.UpdateProductStatus(context.Background(), &UpdateProductStatusReq{SpuID: "spu_001", Status: 1}); err != nil {
        t.Fatal(err)
    }
}

// Verify json tags are correct by round-tripping a request.
func TestAddProductReqJSON(t *testing.T) {
    req := AddProductReq{SpuInfo: SpuInfo{Title: "T", SubTitle: "S"}}
    b, err := json.Marshal(req)
    if err != nil {
        t.Fatal(err)
    }
    var m map[string]any
    if err := json.Unmarshal(b, &m); err != nil {
        t.Fatal(err)
    }
    if _, ok := m["spu_info"]; !ok {
        t.Error("missing spu_info key")
    }
}
```

- [ ] Run `go test ./mini-store/... -run TestAddProduct|TestUpdateProduct|TestDeleteProduct|TestGetProduct|TestListProduct|TestUpdateProductStatus` — all pass.

- [ ] Commit:

```bash
git add mini-store/product.go mini-store/product_test.go
git commit -m "feat(mini-store): implement 6 product management APIs with tests"
```

---

### Task C2-8: `mini-store/order.go` + `mini-store/order_test.go`

WeChat API paths under `/shop/order/`:
- `POST /shop/order/get` — GetOrder
- `POST /shop/order/get_list` — ListOrder
- `POST /shop/order/update_price` — UpdateOrderPrice
- `POST /shop/order/update_express_info` — UpdateOrderExpressInfo

**Steps:**

- [ ] Create `mini-store/order.go`:

```go
package mini_store

import "context"

// OrderInfo describes a Mini Store order.
type OrderInfo struct {
    OrderID      string `json:"order_id"`
    Status       int    `json:"status"`        // 1=pending payment 2=paid 3=shipped 4=received 5=closed
    CreateTime   int64  `json:"create_time"`
    UpdateTime   int64  `json:"update_time"`
    TotalAmount  int64  `json:"total_amount"`  // in fen (1/100 RMB)
    UserOpenID   string `json:"openid"`
    ExpressInfo  *ExpressInfo `json:"express_info,omitempty"`
}

// ExpressInfo holds the shipping information for an order.
type ExpressInfo struct {
    DeliveryID  string `json:"delivery_id"`   // logistics company ID
    TrackingNo  string `json:"tracking_no"`   // waybill number
}

// GetOrderReq is the request body for GetOrder.
type GetOrderReq struct {
    OrderID string `json:"order_id"`
}

// GetOrderResp is the response from GetOrder.
type GetOrderResp struct {
    OrderInfo OrderInfo `json:"order_info"`
}

// ListOrderReq is the request body for ListOrder.
type ListOrderReq struct {
    Status    int   `json:"status,omitempty"`     // 0=all statuses
    StartTime int64 `json:"start_time,omitempty"` // Unix timestamp
    EndTime   int64 `json:"end_time,omitempty"`
    Page      int   `json:"page,omitempty"`
    PageSize  int   `json:"page_size,omitempty"` // max 50
}

// ListOrderResp is the response from ListOrder.
type ListOrderResp struct {
    OrderList []OrderInfo `json:"order_list"`
    TotalNum  int         `json:"total_num"`
}

// UpdateOrderPriceReq is the request body for UpdateOrderPrice.
type UpdateOrderPriceReq struct {
    OrderID     string `json:"order_id"`
    TotalAmount int64  `json:"total_amount"` // new total in fen; must be ≤ original
}

// UpdateOrderExpressInfoReq is the request body for UpdateOrderExpressInfo.
type UpdateOrderExpressInfoReq struct {
    OrderID     string      `json:"order_id"`
    ExpressInfo ExpressInfo `json:"express_info"`
}

// GetOrder retrieves the full order details for the given order_id.
func (c *Client) GetOrder(ctx context.Context, req *GetOrderReq) (*GetOrderResp, error) {
    var resp GetOrderResp
    if err := c.doPost(ctx, "/shop/order/get", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// ListOrder returns a paginated list of orders, optionally filtered by status and time range.
func (c *Client) ListOrder(ctx context.Context, req *ListOrderReq) (*ListOrderResp, error) {
    var resp ListOrderResp
    if err := c.doPost(ctx, "/shop/order/get_list", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// UpdateOrderPrice adjusts the total price of an unpaid order downward.
// The new total_amount must be ≤ the original amount.
func (c *Client) UpdateOrderPrice(ctx context.Context, req *UpdateOrderPriceReq) error {
    return c.doPost(ctx, "/shop/order/update_price", req, nil)
}

// UpdateOrderExpressInfo attaches or updates shipping information for a paid order.
func (c *Client) UpdateOrderExpressInfo(ctx context.Context, req *UpdateOrderExpressInfoReq) error {
    return c.doPost(ctx, "/shop/order/update_express_info", req, nil)
}
```

- [ ] Create `mini-store/order_test.go` with table-driven tests for all four methods (success + errcode error rows each):

```go
package mini_store

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestGetOrder(t *testing.T) {
    tests := []struct {
        name    string
        body    string
        wantErr bool
        wantID  string
    }{
        {
            name:    "success",
            body:    `{"order_info":{"order_id":"ord_001","status":2,"create_time":1000,"total_amount":9900},"errcode":0,"errmsg":"ok"}`,
            wantErr: false,
            wantID:  "ord_001",
        },
        {
            name:    "api_error",
            body:    `{"errcode":40001,"errmsg":"invalid credential"}`,
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            srv := httptest.NewServer(shopHandler("/shop/order/get", tt.body))
            defer srv.Close()
            c := newTestClient(t, srv.URL)
            resp, err := c.GetOrder(context.Background(), &GetOrderReq{OrderID: "ord_001"})
            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v got %v", tt.wantErr, err)
            }
            if !tt.wantErr && resp.OrderInfo.OrderID != tt.wantID {
                t.Errorf("got order_id=%s want %s", resp.OrderInfo.OrderID, tt.wantID)
            }
        })
    }
}

func TestListOrder(t *testing.T) {
    srv := httptest.NewServer(shopHandler("/shop/order/get_list",
        `{"order_list":[{"order_id":"ord_001"},{"order_id":"ord_002"}],"total_num":2,"errcode":0,"errmsg":"ok"}`))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    resp, err := c.ListOrder(context.Background(), &ListOrderReq{Page: 1, PageSize: 10})
    if err != nil {
        t.Fatal(err)
    }
    if resp.TotalNum != 2 || len(resp.OrderList) != 2 {
        t.Errorf("unexpected list response: %+v", resp)
    }
}

func TestUpdateOrderPrice(t *testing.T) {
    srv := httptest.NewServer(shopHandler("/shop/order/update_price", `{"errcode":0,"errmsg":"ok"}`))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    if err := c.UpdateOrderPrice(context.Background(), &UpdateOrderPriceReq{OrderID: "ord_001", TotalAmount: 8000}); err != nil {
        t.Fatal(err)
    }
}

func TestUpdateOrderExpressInfo(t *testing.T) {
    srv := httptest.NewServer(shopHandler("/shop/order/update_express_info", `{"errcode":0,"errmsg":"ok"}`))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    err := c.UpdateOrderExpressInfo(context.Background(), &UpdateOrderExpressInfoReq{
        OrderID:     "ord_001",
        ExpressInfo: ExpressInfo{DeliveryID: "SF", TrackingNo: "SF123456"},
    })
    if err != nil {
        t.Fatal(err)
    }
}
```

- [ ] Run `go test ./mini-store/... -run TestGetOrder|TestListOrder|TestUpdateOrder` — passes.

- [ ] Commit:

```bash
git add mini-store/order.go mini-store/order_test.go
git commit -m "feat(mini-store): implement 4 order management APIs with tests"
```

---

### Task C2-9: `mini-store/delivery.go` + `mini-store/delivery_test.go`

WeChat API paths under `/shop/delivery/`:
- `POST /shop/delivery/add_info` — AddDeliveryInfo
- `POST /shop/delivery/get_info` — GetDeliveryInfo

**Steps:**

- [ ] Create `mini-store/delivery.go`:

```go
package mini_store

import "context"

// DeliveryCompany describes a supported logistics company.
type DeliveryCompany struct {
    DeliveryID   string `json:"delivery_id"`   // e.g. "SF", "YTO"
    DeliveryName string `json:"delivery_name"` // human-readable name
}

// AddDeliveryInfoReq is the request for adding shipping/tracking info to an order.
type AddDeliveryInfoReq struct {
    OrderID    string          `json:"order_id"`
    IsAll      int             `json:"is_all"`      // 1=full shipment 0=partial
    Packages   []PackageInfo   `json:"packages"`    // one entry per package
}

// PackageInfo describes a single package in a shipment.
type PackageInfo struct {
    TrackingNo  string `json:"tracking_no"`  // waybill number
    DeliveryID  string `json:"delivery_id"`  // logistics company ID
}

// AddDeliveryInfoResp is the response from AddDeliveryInfo.
type AddDeliveryInfoResp struct {
    // No additional fields beyond errcode/errmsg.
}

// GetDeliveryInfoReq is the request for retrieving delivery status of an order.
type GetDeliveryInfoReq struct {
    OrderID string `json:"order_id"`
}

// GetDeliveryInfoResp is the response from GetDeliveryInfo.
type GetDeliveryInfoResp struct {
    OrderID  string        `json:"order_id"`
    Packages []PackageInfo `json:"packages"`
    IsAll    int           `json:"is_all"`
}

// AddDeliveryInfo records shipment information (waybill numbers) for an order.
// Set IsAll=1 when all items ship in one batch.
func (c *Client) AddDeliveryInfo(ctx context.Context, req *AddDeliveryInfoReq) error {
    return c.doPost(ctx, "/shop/delivery/add_info", req, nil)
}

// GetDeliveryInfo retrieves the current delivery/tracking information for an order.
func (c *Client) GetDeliveryInfo(ctx context.Context, req *GetDeliveryInfoReq) (*GetDeliveryInfoResp, error) {
    var resp GetDeliveryInfoResp
    if err := c.doPost(ctx, "/shop/delivery/get_info", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

- [ ] Create `mini-store/delivery_test.go` (success + error rows for each method).

- [ ] Run `go test ./mini-store/... -run TestAddDeliveryInfo|TestGetDeliveryInfo` — passes.

- [ ] Commit:

```bash
git add mini-store/delivery.go mini-store/delivery_test.go
git commit -m "feat(mini-store): implement AddDeliveryInfo and GetDeliveryInfo with tests"
```

---

### Task C2-10: `mini-store/settlement.go` + `mini-store/settlement_test.go`

WeChat API paths under `/shop/settlement/`:
- `POST /shop/settlement/get_account` — GetSettlementAccount
- `POST /shop/settlement/bind_account` — BindSettlementAccount
- `POST /shop/settlement/get_settlement_result` — GetSettlementResult

**Steps:**

- [ ] Create `mini-store/settlement.go`:

```go
package mini_store

import "context"

// SettlementAccount describes the bank account linked for payouts.
type SettlementAccount struct {
    BankName      string `json:"bank_name"`       // bank name
    AccountNumber string `json:"account_number"`  // last 4 digits (masked)
    AccountName   string `json:"account_name"`    // account holder name
    Status        int    `json:"status"`           // 1=pending verification 2=verified
}

// GetSettlementAccountResp is the response from GetSettlementAccount.
type GetSettlementAccountResp struct {
    Account SettlementAccount `json:"account"`
}

// BindSettlementAccountReq is the request for binding a bank account for settlement.
type BindSettlementAccountReq struct {
    BankName      string `json:"bank_name"`
    AccountNumber string `json:"account_number"` // full account number
    AccountName   string `json:"account_name"`
    BankBranchID  string `json:"bank_branch_id,omitempty"`
}

// SettlementResult holds the result of a settlement period.
type SettlementResult struct {
    SettlementID   string `json:"settlement_id"`
    SettlementTime int64  `json:"settlement_time"` // Unix timestamp of the payout
    Amount         int64  `json:"amount"`           // settled amount in fen
    Status         int    `json:"status"`            // 1=pending 2=completed 3=failed
}

// GetSettlementResultReq is the request for retrieving a past settlement record.
type GetSettlementResultReq struct {
    SettlementID string `json:"settlement_id"`
}

// GetSettlementResultResp is the response from GetSettlementResult.
type GetSettlementResultResp struct {
    SettlementResult SettlementResult `json:"settlement_result"`
}

// GetSettlementAccount retrieves the bank account currently linked for payouts.
func (c *Client) GetSettlementAccount(ctx context.Context) (*GetSettlementAccountResp, error) {
    var resp GetSettlementAccountResp
    if err := c.doPost(ctx, "/shop/settlement/get_account", struct{}{}, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// BindSettlementAccount links a bank account to receive Mini Store payouts.
func (c *Client) BindSettlementAccount(ctx context.Context, req *BindSettlementAccountReq) error {
    return c.doPost(ctx, "/shop/settlement/bind_account", req, nil)
}

// GetSettlementResult retrieves the details of a completed or pending settlement.
func (c *Client) GetSettlementResult(ctx context.Context, req *GetSettlementResultReq) (*GetSettlementResultResp, error) {
    var resp GetSettlementResultResp
    if err := c.doPost(ctx, "/shop/settlement/get_settlement_result", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

- [ ] Create `mini-store/settlement_test.go` (success + error rows for each method).

- [ ] Run `go test ./mini-store/... -run TestGetSettlementAccount|TestBindSettlementAccount|TestGetSettlementResult` — passes.

- [ ] Commit:

```bash
git add mini-store/settlement.go mini-store/settlement_test.go
git commit -m "feat(mini-store): implement 3 settlement APIs with tests"
```

---

### Task C2-11: `mini-store/coupon.go` + `mini-store/coupon_test.go`

WeChat API paths under `/shop/coupon/`:
- `POST /shop/coupon/create` — CreateCoupon
- `POST /shop/coupon/update` — UpdateCoupon
- `POST /shop/coupon/get` — GetCoupon
- `POST /shop/coupon/get_list` — ListCoupon
- `POST /shop/coupon/delete` — DeleteCoupon

**Steps:**

- [ ] Create `mini-store/coupon.go`:

```go
package mini_store

import "context"

// CouponInfo describes a Mini Store coupon/discount.
type CouponInfo struct {
    CouponID     string `json:"coupon_id,omitempty"`
    Name         string `json:"name,omitempty"`          // coupon display name
    Type         int    `json:"type,omitempty"`          // 1=amount-off 2=percent-off
    DiscountValue int64 `json:"discount_value,omitempty"` // fen for type=1; 0-100 for type=2
    MinOrderAmount int64 `json:"min_order_amount,omitempty"` // minimum order total in fen
    StartTime    int64  `json:"start_time,omitempty"`    // validity start Unix timestamp
    EndTime      int64  `json:"end_time,omitempty"`      // validity end Unix timestamp
    TotalNum     int    `json:"total_num,omitempty"`     // total issuance count
    RemainNum    int    `json:"remain_num,omitempty"`    // remaining count (read-only)
    Status       int    `json:"status,omitempty"`        // 1=active 2=disabled
}

// CreateCouponReq is the request body for CreateCoupon.
type CreateCouponReq struct {
    CouponInfo CouponInfo `json:"coupon_info"`
}

// CreateCouponResp is the response from CreateCoupon.
type CreateCouponResp struct {
    CouponID string `json:"coupon_id"`
}

// UpdateCouponReq is the request body for UpdateCoupon.
type UpdateCouponReq struct {
    CouponID   string     `json:"coupon_id"`
    CouponInfo CouponInfo `json:"coupon_info"`
}

// GetCouponReq is the request body for GetCoupon.
type GetCouponReq struct {
    CouponID string `json:"coupon_id"`
}

// GetCouponResp is the response from GetCoupon.
type GetCouponResp struct {
    CouponInfo CouponInfo `json:"coupon_info"`
}

// ListCouponReq is the request body for ListCoupon.
type ListCouponReq struct {
    Status   int `json:"status,omitempty"`    // 0=all
    Page     int `json:"page,omitempty"`
    PageSize int `json:"page_size,omitempty"`
}

// ListCouponResp is the response from ListCoupon.
type ListCouponResp struct {
    CouponList []CouponInfo `json:"coupon_list"`
    TotalNum   int          `json:"total_num"`
}

// DeleteCouponReq is the request body for DeleteCoupon.
type DeleteCouponReq struct {
    CouponID string `json:"coupon_id"`
}

// CreateCoupon creates a new coupon in the Mini Store and returns its ID.
func (c *Client) CreateCoupon(ctx context.Context, req *CreateCouponReq) (*CreateCouponResp, error) {
    var resp CreateCouponResp
    if err := c.doPost(ctx, "/shop/coupon/create", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// UpdateCoupon modifies an existing coupon. CouponID must be set in req.
func (c *Client) UpdateCoupon(ctx context.Context, req *UpdateCouponReq) error {
    return c.doPost(ctx, "/shop/coupon/update", req, nil)
}

// GetCoupon retrieves the details of the coupon identified by CouponID.
func (c *Client) GetCoupon(ctx context.Context, req *GetCouponReq) (*GetCouponResp, error) {
    var resp GetCouponResp
    if err := c.doPost(ctx, "/shop/coupon/get", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// ListCoupon returns a paginated list of coupons, optionally filtered by status.
func (c *Client) ListCoupon(ctx context.Context, req *ListCouponReq) (*ListCouponResp, error) {
    var resp ListCouponResp
    if err := c.doPost(ctx, "/shop/coupon/get_list", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// DeleteCoupon permanently removes the coupon identified by CouponID.
func (c *Client) DeleteCoupon(ctx context.Context, req *DeleteCouponReq) error {
    return c.doPost(ctx, "/shop/coupon/delete", req, nil)
}
```

- [ ] Create `mini-store/coupon_test.go` with success + error rows for all five methods.

- [ ] Run `go test ./mini-store/... -run TestCreateCoupon|TestUpdateCoupon|TestGetCoupon|TestListCoupon|TestDeleteCoupon` — passes.

- [ ] Commit:

```bash
git add mini-store/coupon.go mini-store/coupon_test.go
git commit -m "feat(mini-store): implement 5 coupon management APIs with tests"
```

---

### Task C2-12: `mini-store/after_sale.go` + `mini-store/after_sale_test.go`

WeChat API paths under `/shop/aftersale/`:
- `POST /shop/aftersale/get` — GetAfterSale
- `POST /shop/aftersale/get_list` — ListAfterSale
- `POST /shop/aftersale/update` — UpdateAfterSale

**Steps:**

- [ ] Create `mini-store/after_sale.go`:

```go
package mini_store

import "context"

// AfterSaleInfo describes an after-sale (return/refund) request.
type AfterSaleInfo struct {
    AfterSaleID  string `json:"aftersale_id"`
    OrderID      string `json:"order_id"`
    Type         int    `json:"type"`          // 1=refund-only 2=return+refund 3=exchange
    Status       int    `json:"status"`        // 1=pending merchant 2=merchant approved 3=closed 4=refunded
    Reason       string `json:"reason,omitempty"`
    RefundAmount int64  `json:"refund_amount,omitempty"` // in fen
    CreateTime   int64  `json:"create_time"`
    UpdateTime   int64  `json:"update_time"`
}

// GetAfterSaleReq is the request for GetAfterSale.
type GetAfterSaleReq struct {
    AfterSaleID string `json:"aftersale_id"`
}

// GetAfterSaleResp is the response from GetAfterSale.
type GetAfterSaleResp struct {
    AfterSaleInfo AfterSaleInfo `json:"aftersale_info"`
}

// ListAfterSaleReq is the request for ListAfterSale.
type ListAfterSaleReq struct {
    Status    int   `json:"status,omitempty"`
    StartTime int64 `json:"start_time,omitempty"`
    EndTime   int64 `json:"end_time,omitempty"`
    Page      int   `json:"page,omitempty"`
    PageSize  int   `json:"page_size,omitempty"`
}

// ListAfterSaleResp is the response from ListAfterSale.
type ListAfterSaleResp struct {
    AfterSaleList []AfterSaleInfo `json:"aftersale_list"`
    TotalNum      int             `json:"total_num"`
}

// UpdateAfterSaleReq is the request for UpdateAfterSale.
type UpdateAfterSaleReq struct {
    AfterSaleID string `json:"aftersale_id"`
    Status      int    `json:"status"` // 2=approve 3=reject
    Note        string `json:"note,omitempty"`
}

// GetAfterSale retrieves the details of an after-sale request.
func (c *Client) GetAfterSale(ctx context.Context, req *GetAfterSaleReq) (*GetAfterSaleResp, error) {
    var resp GetAfterSaleResp
    if err := c.doPost(ctx, "/shop/aftersale/get", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// ListAfterSale returns a paginated list of after-sale requests,
// optionally filtered by status and time range.
func (c *Client) ListAfterSale(ctx context.Context, req *ListAfterSaleReq) (*ListAfterSaleResp, error) {
    var resp ListAfterSaleResp
    if err := c.doPost(ctx, "/shop/aftersale/get_list", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// UpdateAfterSale approves (Status=2) or rejects (Status=3) an after-sale request.
func (c *Client) UpdateAfterSale(ctx context.Context, req *UpdateAfterSaleReq) error {
    return c.doPost(ctx, "/shop/aftersale/update", req, nil)
}
```

- [ ] Create `mini-store/after_sale_test.go` with success + error rows for all three methods.

- [ ] Run `go test ./mini-store/...` — all tests pass.

- [ ] Commit:

```bash
git add mini-store/after_sale.go mini-store/after_sale_test.go
git commit -m "feat(mini-store): implement GetAfterSale, ListAfterSale, UpdateAfterSale with tests"
```

---

## C3: `xiaowei` package

**Package name:** `xiaowei`
**Module path:** `github.com/godrealms/go-wechat-sdk/xiaowei`
**Base URL:** `https://api.weixin.qq.com`
**Path prefix:** `/hardware/`
**WeChat doc:** 微信硬件 (IoT device management)

---

### Task C3-13: `xiaowei/client.go` + `xiaowei/helper.go`

**Steps:**

- [ ] Replace the stub `xiaowei/client.go` with:

```go
// Package xiaowei provides a client for the WeChat Hardware (微信硬件/小微)
// API, covering device registration and authorization, device binding,
// messaging, and firmware management. Create a Client with NewClient.
package xiaowei

import (
    "context"
    "fmt"
    "net/url"
    "sync"
    "time"

    "github.com/godrealms/go-wechat-sdk/utils"
)

// Config holds the xiaowei app credentials.
type Config struct {
    AppId     string
    AppSecret string
}

// TokenSource is an injectable access_token provider.
type TokenSource interface {
    AccessToken(ctx context.Context) (string, error)
}

// Client manages the xiaowei hardware API. Safe for concurrent use.
type Client struct {
    cfg         Config
    http        *utils.HTTP
    mu          sync.RWMutex
    accessToken string
    expiresAt   time.Time
    tokenSource TokenSource
}

// Option is a functional option applied during NewClient.
type Option func(*Client)

// WithHTTP injects a custom HTTP client, primarily for testing.
func WithHTTP(h *utils.HTTP) Option { return func(c *Client) { c.http = h } }

// WithTokenSource injects an external access_token provider.
func WithTokenSource(ts TokenSource) Option { return func(c *Client) { c.tokenSource = ts } }

// NewClient constructs a xiaowei client. Returns an error if AppId is
// empty or if AppSecret is empty and no TokenSource is provided.
func NewClient(cfg Config, opts ...Option) (*Client, error) {
    if cfg.AppId == "" {
        return nil, fmt.Errorf("xiaowei: AppId is required")
    }
    c := &Client{
        cfg:  cfg,
        http: utils.NewHTTP("https://api.weixin.qq.com", utils.WithTimeout(30*time.Second)),
    }
    for _, o := range opts {
        o(c)
    }
    if c.tokenSource == nil && cfg.AppSecret == "" {
        return nil, fmt.Errorf("xiaowei: AppSecret is required when no TokenSource is injected")
    }
    return c, nil
}

// HTTP returns the underlying HTTP client for custom requests.
func (c *Client) HTTP() *utils.HTTP { return c.http }

type accessTokenResp struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int64  `json:"expires_in"`
    ErrCode     int    `json:"errcode,omitempty"`
    ErrMsg      string `json:"errmsg,omitempty"`
}

// AccessToken returns a valid access_token, refreshing 60 s before expiry.
func (c *Client) AccessToken(ctx context.Context) (string, error) {
    if c.tokenSource != nil {
        return c.tokenSource.AccessToken(ctx)
    }
    c.mu.RLock()
    if c.accessToken != "" && time.Now().Before(c.expiresAt) {
        t := c.accessToken
        c.mu.RUnlock()
        return t, nil
    }
    c.mu.RUnlock()
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.accessToken != "" && time.Now().Before(c.expiresAt) {
        return c.accessToken, nil
    }
    q := url.Values{
        "grant_type": {"client_credential"},
        "appid":      {c.cfg.AppId},
        "secret":     {c.cfg.AppSecret},
    }
    out := &accessTokenResp{}
    if err := c.http.Get(ctx, "/cgi-bin/token", q, out); err != nil {
        return "", fmt.Errorf("xiaowei: fetch token: %w", err)
    }
    if out.ErrCode != 0 {
        return "", fmt.Errorf("xiaowei: token errcode=%d errmsg=%s", out.ErrCode, out.ErrMsg)
    }
    if out.AccessToken == "" {
        return "", fmt.Errorf("xiaowei: empty access_token")
    }
    c.accessToken = out.AccessToken
    c.expiresAt = time.Now().Add(time.Duration(out.ExpiresIn-60) * time.Second)
    return c.accessToken, nil
}
```

- [ ] Create `xiaowei/helper.go`:

```go
package xiaowei

import (
    "context"
    "encoding/json"
    "fmt"
    "net/url"
)

type baseResp struct {
    ErrCode int    `json:"errcode"`
    ErrMsg  string `json:"errmsg"`
}

// doPost sends a POST JSON request with access_token, checks errcode,
// and unmarshals the response into out.
func (c *Client) doPost(ctx context.Context, path string, body any, out any) error {
    tok, err := c.AccessToken(ctx)
    if err != nil {
        return err
    }
    q := url.Values{"access_token": {tok}}
    fullPath := path + "?" + q.Encode()
    var raw json.RawMessage
    if err := c.http.Post(ctx, fullPath, body, &raw); err != nil {
        return err
    }
    var base baseResp
    _ = json.Unmarshal(raw, &base)
    if base.ErrCode != 0 {
        return fmt.Errorf("xiaowei: %s errcode=%d errmsg=%s", path, base.ErrCode, base.ErrMsg)
    }
    if out != nil {
        return json.Unmarshal(raw, out)
    }
    return nil
}
```

- [ ] Create `xiaowei/client_test.go` with the same pattern as aispeech (`newTestClient`, `fakeTokenSource`, `TestNewClient`, `TestAccessToken`, `TestAccessToken_UsesInjectedTokenSource`).

- [ ] Run `go test ./xiaowei/... -run TestNewClient` — passes.

- [ ] Commit:

```bash
git add xiaowei/client.go xiaowei/helper.go xiaowei/client_test.go
git commit -m "feat(xiaowei): implement client, token caching, helper doPost"
```

---

### Task C3-14: `xiaowei/device.go` + `xiaowei/device_test.go`

WeChat API paths under `/hardware/`:
- `POST /hardware/device/register` — RegisterDevice
- `POST /hardware/device/authorize` — AuthorizeDevice
- `POST /hardware/device/get_info` — GetDeviceInfo
- `POST /hardware/device/get_list` — ListDevice

**Steps:**

- [ ] Create `xiaowei/device.go`:

```go
package xiaowei

import "context"

// DeviceInfo describes a registered hardware device.
type DeviceInfo struct {
    DeviceID      string `json:"device_id"`
    ProductID     string `json:"product_id"`      // product/model identifier
    DeviceType    string `json:"device_type"`      // hardware category
    SerialNo      string `json:"serial_no,omitempty"` // device serial number
    Status        int    `json:"status"`           // 1=normal 2=blocked
    CreateTime    int64  `json:"create_time"`
}

// RegisterDeviceReq is the request for RegisterDevice.
type RegisterDeviceReq struct {
    ProductID  string `json:"product_id"`
    SerialNo   string `json:"serial_no"`
    DeviceType string `json:"device_type,omitempty"`
}

// RegisterDeviceResp is the response from RegisterDevice.
type RegisterDeviceResp struct {
    DeviceID string `json:"device_id"` // WeChat-assigned device identifier
}

// AuthorizeDeviceReq is the request for AuthorizeDevice.
type AuthorizeDeviceReq struct {
    DeviceID string `json:"device_id"`
    OpenID   string `json:"openid"` // user openid performing the authorization
}

// GetDeviceInfoReq is the request for GetDeviceInfo.
type GetDeviceInfoReq struct {
    DeviceID string `json:"device_id"`
}

// GetDeviceInfoResp is the response from GetDeviceInfo.
type GetDeviceInfoResp struct {
    DeviceInfo DeviceInfo `json:"device_info"`
}

// ListDeviceReq is the request for ListDevice.
type ListDeviceReq struct {
    ProductID string `json:"product_id,omitempty"` // filter by product
    Page      int    `json:"page,omitempty"`
    PageSize  int    `json:"page_size,omitempty"`
}

// ListDeviceResp is the response from ListDevice.
type ListDeviceResp struct {
    DeviceList []DeviceInfo `json:"device_list"`
    TotalNum   int          `json:"total_num"`
}

// RegisterDevice registers a new hardware device and returns its WeChat device_id.
func (c *Client) RegisterDevice(ctx context.Context, req *RegisterDeviceReq) (*RegisterDeviceResp, error) {
    var resp RegisterDeviceResp
    if err := c.doPost(ctx, "/hardware/device/register", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// AuthorizeDevice grants a user (identified by OpenID) access to the device.
func (c *Client) AuthorizeDevice(ctx context.Context, req *AuthorizeDeviceReq) error {
    return c.doPost(ctx, "/hardware/device/authorize", req, nil)
}

// GetDeviceInfo retrieves the details of the device identified by DeviceID.
func (c *Client) GetDeviceInfo(ctx context.Context, req *GetDeviceInfoReq) (*GetDeviceInfoResp, error) {
    var resp GetDeviceInfoResp
    if err := c.doPost(ctx, "/hardware/device/get_info", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// ListDevice returns a paginated list of registered devices, optionally filtered by ProductID.
func (c *Client) ListDevice(ctx context.Context, req *ListDeviceReq) (*ListDeviceResp, error) {
    var resp ListDeviceResp
    if err := c.doPost(ctx, "/hardware/device/get_list", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

- [ ] Create `xiaowei/device_test.go`:

```go
package xiaowei

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func hwHandler(path, body string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/cgi-bin/token":
            _, _ = w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
        case path:
            _, _ = w.Write([]byte(body))
        default:
            w.WriteHeader(http.StatusNotFound)
        }
    }
}

func TestRegisterDevice(t *testing.T) {
    tests := []struct {
        name    string
        body    string
        wantErr bool
        wantID  string
    }{
        {
            name:    "success",
            body:    `{"device_id":"dev_001","errcode":0,"errmsg":"ok"}`,
            wantErr: false,
            wantID:  "dev_001",
        },
        {
            name:    "api_error",
            body:    `{"errcode":40001,"errmsg":"invalid credential"}`,
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            srv := httptest.NewServer(hwHandler("/hardware/device/register", tt.body))
            defer srv.Close()
            c := newTestClient(t, srv.URL)
            resp, err := c.RegisterDevice(context.Background(), &RegisterDeviceReq{
                ProductID: "prod_001", SerialNo: "SN0001",
            })
            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v got %v", tt.wantErr, err)
            }
            if !tt.wantErr && resp.DeviceID != tt.wantID {
                t.Errorf("got device_id=%s want %s", resp.DeviceID, tt.wantID)
            }
        })
    }
}

func TestAuthorizeDevice(t *testing.T) {
    srv := httptest.NewServer(hwHandler("/hardware/device/authorize", `{"errcode":0,"errmsg":"ok"}`))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    if err := c.AuthorizeDevice(context.Background(), &AuthorizeDeviceReq{DeviceID: "dev_001", OpenID: "oid_001"}); err != nil {
        t.Fatal(err)
    }
}

func TestGetDeviceInfo(t *testing.T) {
    srv := httptest.NewServer(hwHandler("/hardware/device/get_info",
        `{"device_info":{"device_id":"dev_001","product_id":"prod_001","status":1,"create_time":1000},"errcode":0,"errmsg":"ok"}`))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    resp, err := c.GetDeviceInfo(context.Background(), &GetDeviceInfoReq{DeviceID: "dev_001"})
    if err != nil {
        t.Fatal(err)
    }
    if resp.DeviceInfo.DeviceID != "dev_001" {
        t.Errorf("got device_id=%s want dev_001", resp.DeviceInfo.DeviceID)
    }
}

func TestListDevice(t *testing.T) {
    srv := httptest.NewServer(hwHandler("/hardware/device/get_list",
        `{"device_list":[{"device_id":"dev_001"},{"device_id":"dev_002"}],"total_num":2,"errcode":0,"errmsg":"ok"}`))
    defer srv.Close()
    c := newTestClient(t, srv.URL)
    resp, err := c.ListDevice(context.Background(), &ListDeviceReq{Page: 1, PageSize: 20})
    if err != nil {
        t.Fatal(err)
    }
    if resp.TotalNum != 2 || len(resp.DeviceList) != 2 {
        t.Errorf("unexpected list: %+v", resp)
    }
}
```

- [ ] Run `go test ./xiaowei/... -run TestRegisterDevice|TestAuthorizeDevice|TestGetDeviceInfo|TestListDevice` — passes.

- [ ] Commit:

```bash
git add xiaowei/device.go xiaowei/device_test.go
git commit -m "feat(xiaowei): implement RegisterDevice, AuthorizeDevice, GetDeviceInfo, ListDevice with tests"
```

---

### Task C3-15: `xiaowei/binding.go` + `xiaowei/binding_test.go`

WeChat API paths under `/hardware/`:
- `POST /hardware/device/bind` — BindDevice
- `POST /hardware/device/unbind` — UnbindDevice
- `POST /hardware/device/get_bind_user` — GetBindUser

**Steps:**

- [ ] Create `xiaowei/binding.go`:

```go
package xiaowei

import "context"

// BindDeviceReq is the request for binding a user to a device.
type BindDeviceReq struct {
    DeviceID string `json:"device_id"`
    OpenID   string `json:"openid"` // user to bind
    Ticket   string `json:"ticket"` // QR scan ticket from WeChat client
}

// UnbindDeviceReq is the request for unbinding a user from a device.
type UnbindDeviceReq struct {
    DeviceID string `json:"device_id"`
    OpenID   string `json:"openid"`
}

// GetBindUserReq is the request for listing users bound to a device.
type GetBindUserReq struct {
    DeviceID string `json:"device_id"`
}

// BindUserInfo holds the openid and binding timestamp for one bound user.
type BindUserInfo struct {
    OpenID     string `json:"openid"`
    BindTime   int64  `json:"bind_time"` // Unix timestamp
}

// GetBindUserResp is the response from GetBindUser.
type GetBindUserResp struct {
    UserList []BindUserInfo `json:"user_list"`
    Total    int            `json:"total"`
}

// BindDevice associates the user identified by OpenID with the device.
// The Ticket is obtained when the user scans the device QR code.
func (c *Client) BindDevice(ctx context.Context, req *BindDeviceReq) error {
    return c.doPost(ctx, "/hardware/device/bind", req, nil)
}

// UnbindDevice removes the binding between a user and a device.
func (c *Client) UnbindDevice(ctx context.Context, req *UnbindDeviceReq) error {
    return c.doPost(ctx, "/hardware/device/unbind", req, nil)
}

// GetBindUser returns all users currently bound to the given device.
func (c *Client) GetBindUser(ctx context.Context, req *GetBindUserReq) (*GetBindUserResp, error) {
    var resp GetBindUserResp
    if err := c.doPost(ctx, "/hardware/device/get_bind_user", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

- [ ] Create `xiaowei/binding_test.go` (success + error rows for each of the three methods).

- [ ] Run `go test ./xiaowei/... -run TestBindDevice|TestUnbindDevice|TestGetBindUser` — passes.

- [ ] Commit:

```bash
git add xiaowei/binding.go xiaowei/binding_test.go
git commit -m "feat(xiaowei): implement BindDevice, UnbindDevice, GetBindUser with tests"
```

---

### Task C3-16: `xiaowei/message.go` + `xiaowei/message_test.go`

WeChat API paths under `/hardware/`:
- `POST /hardware/message/send` — SendDeviceMessage
- `POST /hardware/message/get` — GetDeviceMessage

**Steps:**

- [ ] Create `xiaowei/message.go`:

```go
package xiaowei

import "context"

// DeviceMessage holds a message sent to or received from a hardware device.
type DeviceMessage struct {
    DeviceID   string `json:"device_id"`
    OpenID     string `json:"openid,omitempty"`  // target user (for send)
    MsgType    int    `json:"msg_type"`           // 1=text 2=image 3=binary
    Content    string `json:"content,omitempty"` // text content
    RawData    string `json:"raw_data,omitempty"` // base64-encoded binary payload
    CreateTime int64  `json:"create_time,omitempty"`
    MsgID      string `json:"msg_id,omitempty"`
}

// SendDeviceMessageReq is the request body for SendDeviceMessage.
type SendDeviceMessageReq struct {
    DeviceID string `json:"device_id"`
    OpenID   string `json:"openid"`   // recipient user openid
    MsgType  int    `json:"msg_type"` // 1=text 2=binary
    Content  string `json:"content,omitempty"`
    RawData  string `json:"raw_data,omitempty"`
}

// SendDeviceMessageResp is the response from SendDeviceMessage.
type SendDeviceMessageResp struct {
    MsgID string `json:"msg_id"` // WeChat-assigned message ID
}

// GetDeviceMessageReq is the request for fetching messages received from a device.
type GetDeviceMessageReq struct {
    DeviceID string `json:"device_id"`
    MsgID    string `json:"msg_id,omitempty"`  // fetch a specific message
    Count    int    `json:"count,omitempty"`   // number of messages to fetch (max 20)
}

// GetDeviceMessageResp is the response from GetDeviceMessage.
type GetDeviceMessageResp struct {
    MsgList []DeviceMessage `json:"msg_list"`
}

// SendDeviceMessage sends a message from the server to a device (and its bound user).
func (c *Client) SendDeviceMessage(ctx context.Context, req *SendDeviceMessageReq) (*SendDeviceMessageResp, error) {
    var resp SendDeviceMessageResp
    if err := c.doPost(ctx, "/hardware/message/send", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// GetDeviceMessage retrieves messages uploaded by the device to the WeChat server.
func (c *Client) GetDeviceMessage(ctx context.Context, req *GetDeviceMessageReq) (*GetDeviceMessageResp, error) {
    var resp GetDeviceMessageResp
    if err := c.doPost(ctx, "/hardware/message/get", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

- [ ] Create `xiaowei/message_test.go` (success + error rows for both methods).

- [ ] Run `go test ./xiaowei/... -run TestSendDeviceMessage|TestGetDeviceMessage` — passes.

- [ ] Commit:

```bash
git add xiaowei/message.go xiaowei/message_test.go
git commit -m "feat(xiaowei): implement SendDeviceMessage and GetDeviceMessage with tests"
```

---

### Task C3-17: `xiaowei/firmware.go` + `xiaowei/firmware_test.go`

WeChat API paths under `/hardware/`:
- `POST /hardware/firmware/get_info` — GetFirmwareInfo
- `POST /hardware/firmware/create` — CreateFirmware
- `POST /hardware/firmware/set_version` — SetFirmwareVersion

**Steps:**

- [ ] Create `xiaowei/firmware.go`:

```go
package xiaowei

import "context"

// FirmwareInfo describes a firmware version for a hardware product.
type FirmwareInfo struct {
    FirmwareID  string `json:"firmware_id"`
    ProductID   string `json:"product_id"`
    Version     string `json:"version"`      // semantic version string, e.g. "1.2.3"
    URL         string `json:"url"`          // download URL of the firmware binary
    MD5         string `json:"md5"`          // MD5 checksum of the binary
    Size        int64  `json:"size"`         // binary size in bytes
    Description string `json:"description,omitempty"`
    CreateTime  int64  `json:"create_time"`
    Status      int    `json:"status"` // 1=draft 2=released
}

// GetFirmwareInfoReq is the request for GetFirmwareInfo.
type GetFirmwareInfoReq struct {
    FirmwareID string `json:"firmware_id"`
}

// GetFirmwareInfoResp is the response from GetFirmwareInfo.
type GetFirmwareInfoResp struct {
    FirmwareInfo FirmwareInfo `json:"firmware_info"`
}

// CreateFirmwareReq is the request for CreateFirmware.
type CreateFirmwareReq struct {
    ProductID   string `json:"product_id"`
    Version     string `json:"version"`
    URL         string `json:"url"`          // publicly accessible download URL
    MD5         string `json:"md5"`
    Size        int64  `json:"size"`
    Description string `json:"description,omitempty"`
}

// CreateFirmwareResp is the response from CreateFirmware.
type CreateFirmwareResp struct {
    FirmwareID string `json:"firmware_id"`
}

// SetFirmwareVersionReq is the request for SetFirmwareVersion.
type SetFirmwareVersionReq struct {
    ProductID  string `json:"product_id"`
    FirmwareID string `json:"firmware_id"` // the firmware version to activate
}

// GetFirmwareInfo retrieves the metadata for the specified firmware version.
func (c *Client) GetFirmwareInfo(ctx context.Context, req *GetFirmwareInfoReq) (*GetFirmwareInfoResp, error) {
    var resp GetFirmwareInfoResp
    if err := c.doPost(ctx, "/hardware/firmware/get_info", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// CreateFirmware registers a new firmware version for OTA (over-the-air) updates.
// Returns the WeChat-assigned firmware_id.
func (c *Client) CreateFirmware(ctx context.Context, req *CreateFirmwareReq) (*CreateFirmwareResp, error) {
    var resp CreateFirmwareResp
    if err := c.doPost(ctx, "/hardware/firmware/create", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// SetFirmwareVersion activates a firmware version as the current release for a product,
// triggering OTA push to all devices that support auto-update.
func (c *Client) SetFirmwareVersion(ctx context.Context, req *SetFirmwareVersionReq) error {
    return c.doPost(ctx, "/hardware/firmware/set_version", req, nil)
}
```

- [ ] Create `xiaowei/firmware_test.go`:

```go
package xiaowei

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestGetFirmwareInfo(t *testing.T) {
    tests := []struct {
        name    string
        body    string
        wantErr bool
        wantID  string
    }{
        {
            name:    "success",
            body:    `{"firmware_info":{"firmware_id":"fw_001","product_id":"prod_001","version":"1.0.0","url":"https://example.com/fw.bin","md5":"abc","size":1024,"create_time":1000,"status":2},"errcode":0,"errmsg":"ok"}`,
            wantErr: false,
            wantID:  "fw_001",
        },
        {
            name:    "api_error",
            body:    `{"errcode":40001,"errmsg":"invalid credential"}`,
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            srv := httptest.NewServer(hwHandler("/hardware/firmware/get_info", tt.body))
            defer srv.Close()
            c := newTestClient(t, srv.URL)
            resp, err := c.GetFirmwareInfo(context.Background(), &GetFirmwareInfoReq{FirmwareID: "fw_001"})
            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v got %v", tt.wantErr, err)
            }
            if !tt.wantErr && resp.FirmwareInfo.FirmwareID != tt.wantID {
                t.Errorf("got firmware_id=%s want %s", resp.FirmwareInfo.FirmwareID, tt.wantID)
            }
        })
    }
}

func TestCreateFirmware(t *testing.T) {
    tests := []struct {
        name    string
        body    string
        wantErr bool
        wantID  string
    }{
        {
            name:    "success",
            body:    `{"firmware_id":"fw_002","errcode":0,"errmsg":"ok"}`,
            wantErr: false,
            wantID:  "fw_002",
        },
        {
            name:    "api_error",
            body:    `{"errcode":40001,"errmsg":"invalid credential"}`,
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            srv := httptest.NewServer(hwHandler("/hardware/firmware/create", tt.body))
            defer srv.Close()
            c := newTestClient(t, srv.URL)
            resp, err := c.CreateFirmware(context.Background(), &CreateFirmwareReq{
                ProductID: "prod_001",
                Version:   "1.1.0",
                URL:       "https://example.com/fw.bin",
                MD5:       "abc123",
                Size:      2048,
            })
            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v got %v", tt.wantErr, err)
            }
            if !tt.wantErr && resp.FirmwareID != tt.wantID {
                t.Errorf("got firmware_id=%s want %s", resp.FirmwareID, tt.wantID)
            }
        })
    }
}

func TestSetFirmwareVersion(t *testing.T) {
    tests := []struct {
        name    string
        body    string
        wantErr bool
    }{
        {
            name:    "success",
            body:    `{"errcode":0,"errmsg":"ok"}`,
            wantErr: false,
        },
        {
            name:    "api_error",
            body:    `{"errcode":40001,"errmsg":"invalid credential"}`,
            wantErr: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            srv := httptest.NewServer(hwHandler("/hardware/firmware/set_version", tt.body))
            defer srv.Close()
            c := newTestClient(t, srv.URL)
            err := c.SetFirmwareVersion(context.Background(), &SetFirmwareVersionReq{
                ProductID: "prod_001", FirmwareID: "fw_002",
            })
            if (err != nil) != tt.wantErr {
                t.Fatalf("wantErr=%v got %v", tt.wantErr, err)
            }
        })
    }
}
```

- [ ] Run `go test ./xiaowei/...` — all tests pass.

- [ ] Commit:

```bash
git add xiaowei/firmware.go xiaowei/firmware_test.go
git commit -m "feat(xiaowei): implement GetFirmwareInfo, CreateFirmware, SetFirmwareVersion with tests"
```

---

## Final Verification

After all 17 tasks are complete:

- [ ] `go build ./...` — zero errors across all packages.
- [ ] `go test ./aispeech/... ./mini-store/... ./xiaowei/...` — all tests pass.
- [ ] `go vet ./aispeech/... ./mini-store/... ./xiaowei/...` — zero warnings.
- [ ] Verify method counts:
  - `go doc -all github.com/godrealms/go-wechat-sdk/aispeech | grep -c "^func"` → 7
  - `go doc -all github.com/godrealms/go-wechat-sdk/mini-store | grep -c "^func"` → 24
  - `go doc -all github.com/godrealms/go-wechat-sdk/xiaowei | grep -c "^func"` → 12

- [ ] Final cleanup commit if needed:

```bash
git add aispeech/ mini-store/ xiaowei/
git commit -m "feat: complete stub package implementations (aispeech, mini-store, xiaowei)"
```
