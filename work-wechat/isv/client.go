// Package isv provides a client for the WeCom (企业微信) third-party ISV API.
// Create a Client with NewClient to manage suite tokens and handle
// authorized-enterprise API calls on behalf of third-party applications.
package isv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
	"github.com/godrealms/go-wechat-sdk/utils/wxcrypto"
)

const (
	defaultBaseURL    = "https://qyapi.weixin.qq.com"
	defaultHTTPTimout = 10 * time.Second
)

// newDefaultHTTPClient returns a private *http.Client with a sane total timeout.
// Avoids sharing http.DefaultClient (which has no timeout and can leak goroutines
// on network hangs). Callers can still override with WithHTTPClient.
func newDefaultHTTPClient() *http.Client {
	return &http.Client{Timeout: defaultHTTPTimout}
}

// Config holds the WeCom ISV suite credentials.
type Config struct {
	SuiteID        string // 第三方应用 suite_id
	SuiteSecret    string // 第三方应用 suite_secret
	ProviderCorpID string // 服务商自己的 corpid(provider 接口需要,可选)
	ProviderSecret string // 服务商 provider_secret(provider 接口需要,可选)
	Token          string // 回调 token
	EncodingAESKey string // 回调 AES key(43 字符)
}

// Client is the WeCom ISV API client. Safe for concurrent use.
type Client struct {
	cfg     Config
	store   Store
	http    *http.Client
	crypto  *wxcrypto.MsgCrypto
	baseURL string

	suiteMu    sync.Mutex
	providerMu sync.Mutex
	corpMu     sync.Map // map[corpid]*sync.Mutex
}

// Option 是函数式配置项。
type Option func(*Client)

func WithStore(s Store) Option             { return func(c *Client) { c.store = s } }
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }
func WithBaseURL(u string) Option          { return func(c *Client) { c.baseURL = u } }

// NewClient constructs a WeCom ISV client.
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	if cfg.SuiteID == "" {
		return nil, fmt.Errorf("isv: SuiteID required")
	}
	if cfg.SuiteSecret == "" {
		return nil, fmt.Errorf("isv: SuiteSecret required")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("isv: Token required")
	}
	if len(cfg.EncodingAESKey) != 43 {
		return nil, fmt.Errorf("isv: EncodingAESKey must be 43 chars")
	}
	// ProviderCorpID 与 ProviderSecret 要么都填要么都空
	if (cfg.ProviderCorpID == "") != (cfg.ProviderSecret == "") {
		return nil, fmt.Errorf("isv: ProviderCorpID and ProviderSecret must both be set or both empty")
	}

	cry, err := wxcrypto.New(cfg.Token, cfg.EncodingAESKey, cfg.SuiteID)
	if err != nil {
		return nil, fmt.Errorf("isv: init crypto: %w", err)
	}

	c := &Client{
		cfg:     cfg,
		store:   NewMemoryStore(),
		http:    newDefaultHTTPClient(),
		crypto:  cry,
		baseURL: defaultBaseURL,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// ---- shared HTTP helpers ----

// doPost 发送 JSON POST 到 baseURL + path,query 自动注入 suite_access_token。
func (c *Client) doPost(ctx context.Context, path string, body, out any) error {
	tok, err := c.GetSuiteAccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"suite_access_token": {tok}}
	return c.doPostRaw(ctx, path, q, body, out)
}

// doGet 发送 GET 到 baseURL + path,query 自动注入 suite_access_token。
func (c *Client) doGet(ctx context.Context, path string, extra url.Values, out any) error {
	tok, err := c.GetSuiteAccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{}
	for k, v := range extra {
		q[k] = v
	}
	q.Set("suite_access_token", tok)
	return c.doRequestRaw(ctx, http.MethodGet, path, q, nil, out)
}

// doPostRaw 不自动获取 suite_token,query 由调用方完全控制。
func (c *Client) doPostRaw(ctx context.Context, path string, query url.Values, body, out any) error {
	var buf io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("isv: marshal body: %w", err)
		}
		buf = bytes.NewReader(raw)
	}
	return c.doRequestRaw(ctx, http.MethodPost, path, query, buf, out)
}

func (c *Client) doRequestRaw(ctx context.Context, method, path string, query url.Values, body io.Reader, out any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return fmt.Errorf("isv: new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("isv: http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("isv: read body: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("isv: http %d: %s", resp.StatusCode, string(raw))
	}
	return decodeRaw(path, raw, out)
}

// decodeRaw delegates to utils.DecodeEnvelope for two-stage JSON decode:
// check errcode first, then unmarshal into out.
func decodeRaw(path string, raw []byte, out any) error {
	return utils.DecodeEnvelope("isv", path, raw, out, func(code int, msg, _ string) error {
		return &WeixinError{ErrCode: code, ErrMsg: msg}
	})
}
