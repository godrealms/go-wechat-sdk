package core

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// BaseClient provides shared token management and HTTP functionality
// for all WeChat SDK sub-packages.
type BaseClient struct {
	Ctx         context.Context
	Config      *BaseConfig
	Https       *utils.HTTP
	accessToken *AccessToken
	tokenMutex  sync.RWMutex
	TokenURL    string // e.g. "/cgi-bin/token"
	TokenMethod string // "GET" or "POST"
}

// NewBaseClient creates and returns an initialised BaseClient.
// baseURL is the HTTP base URL (e.g. "https://api.weixin.qq.com").
// tokenURL is the path used to fetch an access token.
// tokenMethod is either "GET" or "POST".
func NewBaseClient(
	ctx context.Context,
	config *BaseConfig,
	baseURL string,
	tokenURL string,
	tokenMethod string,
) *BaseClient {
	return &BaseClient{
		Ctx:         ctx,
		Config:      config,
		Https:       utils.NewHTTP(baseURL, utils.WithTimeout(30*time.Second)),
		TokenURL:    tokenURL,
		TokenMethod: tokenMethod,
	}
}

// getAccessToken returns a valid access token, refreshing if necessary.
// Uses double-checked locking to avoid redundant refreshes.
func (c *BaseClient) getAccessToken() string {
	c.tokenMutex.RLock()
	if c.accessToken != nil && c.accessToken.ExpiresIn > time.Now().Unix() {
		token := c.accessToken.AccessToken
		c.tokenMutex.RUnlock()
		return token
	}
	c.tokenMutex.RUnlock()

	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	if c.accessToken != nil && c.accessToken.ExpiresIn > time.Now().Unix() {
		return c.accessToken.AccessToken
	}
	token, err := c.refreshAccessToken()
	if err != nil {
		return ""
	}
	c.accessToken = token
	return c.accessToken.AccessToken
}

// refreshAccessToken fetches a new access token from the WeChat API.
func (c *BaseClient) refreshAccessToken() (*AccessToken, error) {
	result := &AccessToken{}
	var err error
	if c.TokenMethod == "POST" {
		body := map[string]string{
			"grant_type": "client_credential",
			"appid":      c.Config.AppId,
			"secret":     c.Config.AppSecret,
		}
		err = c.Https.Post(c.Ctx, c.TokenURL, body, result)
	} else {
		query := url.Values{
			"grant_type": {"client_credential"},
			"appid":      {c.Config.AppId},
			"secret":     {c.Config.AppSecret},
		}
		err = c.Https.Get(c.Ctx, c.TokenURL, query, result)
	}
	if err != nil {
		return nil, fmt.Errorf("core: refresh access token: %w", err)
	}
	result.ExpiresIn = time.Now().Unix() + result.ExpiresIn - 10
	return result, nil
}

// GetAccessToken returns the current valid access token.
func (c *BaseClient) GetAccessToken() string {
	return c.getAccessToken()
}

// GetAccessTokenWithError returns the current valid access token or an error.
func (c *BaseClient) GetAccessTokenWithError() (string, error) {
	c.tokenMutex.RLock()
	if c.accessToken != nil && c.accessToken.ExpiresIn > time.Now().Unix() {
		token := c.accessToken.AccessToken
		c.tokenMutex.RUnlock()
		return token, nil
	}
	c.tokenMutex.RUnlock()

	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	if c.accessToken != nil && c.accessToken.ExpiresIn > time.Now().Unix() {
		return c.accessToken.AccessToken, nil
	}
	token, err := c.refreshAccessToken()
	if err != nil {
		return "", err
	}
	c.accessToken = token
	return c.accessToken.AccessToken, nil
}

// SetAccessToken replaces the stored access token.
// Needed by callers that obtain a token via a separate mechanism (e.g. GetStableAccessToken).
func (c *BaseClient) SetAccessToken(token *AccessToken) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.accessToken = token
}

// TokenQuery returns url.Values containing the current access token merged
// with any additional values provided via extra.
func (c *BaseClient) TokenQuery(extra ...url.Values) url.Values {
	q := url.Values{
		"access_token": {c.GetAccessToken()},
	}
	for _, v := range extra {
		for key, vals := range v {
			q[key] = vals
		}
	}
	return q
}
