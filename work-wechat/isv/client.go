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

	"github.com/godrealms/go-wechat-sdk/utils/wxcrypto"
)

const defaultBaseURL = "https://qyapi.weixin.qq.com"

// Config 是 ISV Client 的运行时配置。
type Config struct {
	SuiteID        string // 第三方应用 suite_id
	SuiteSecret    string // 第三方应用 suite_secret
	ProviderCorpID string // 服务商自己的 corpid(provider 接口需要,可选)
	ProviderSecret string // 服务商 provider_secret(provider 接口需要,可选)
	Token          string // 回调 token
	EncodingAESKey string // 回调 AES key(43 字符)
}

// Client 是服务商级别的入口,无状态可共享。
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

// NewClient 校验配置并构造 Client。
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
		http:    http.DefaultClient,
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
func (c *Client) doPost(ctx context.Context, path string, body, out interface{}) error {
	tok, err := c.GetSuiteAccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"suite_access_token": {tok}}
	return c.doPostRaw(ctx, path, q, body, out)
}

// doGet 发送 GET 到 baseURL + path,query 自动注入 suite_access_token。
func (c *Client) doGet(ctx context.Context, path string, extra url.Values, out interface{}) error {
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
func (c *Client) doPostRaw(ctx context.Context, path string, query url.Values, body, out interface{}) error {
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

func (c *Client) doRequestRaw(ctx context.Context, method, path string, query url.Values, body io.Reader, out interface{}) error {
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
	return decodeRaw(raw, out)
}

// decodeRaw 实现"两阶段解码":先检查 errcode,再 unmarshal 到 out。
func decodeRaw(raw []byte, out interface{}) error {
	var probe struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return fmt.Errorf("isv: decode errcode: %w", err)
	}
	if probe.ErrCode != 0 {
		return &WeixinError{ErrCode: probe.ErrCode, ErrMsg: probe.ErrMsg}
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("isv: decode body: %w", err)
	}
	return nil
}
