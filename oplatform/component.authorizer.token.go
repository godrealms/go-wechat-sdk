package oplatform

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"
)

const authorizerTokenSafetyWindow = 60 * time.Second

// errcodeAuthorizerRevoked 微信返回 refresh_token 失效。
const errcodeAuthorizerRevoked = 61023

// AuthorizerClient 是代单个 authorizer 调用微信接口的句柄。
// 同时实现 offiaccount.TokenSource 和 mini_program.TokenSource。
type AuthorizerClient struct {
	c     *Client
	appID string
}

// Authorizer 返回一个指向某 authorizer 的句柄。构造不做 I/O。
func (c *Client) Authorizer(authorizerAppID string) *AuthorizerClient {
	return &AuthorizerClient{c: c, appID: authorizerAppID}
}

// AppID 返回被授权方 appid。
func (a *AuthorizerClient) AppID() string { return a.appID }

// AccessToken 返回 authorizer_access_token，过期自动刷新。
// 签名匹配 offiaccount.TokenSource 与 mini_program.TokenSource。
func (a *AuthorizerClient) AccessToken(ctx context.Context) (string, error) {
	ctx = touchContext(ctx)

	tokens, err := a.c.store.GetAuthorizer(ctx, a.appID)
	if err == nil && time.Now().Add(authorizerTokenSafetyWindow).Before(tokens.ExpireAt) {
		return tokens.AccessToken, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", fmt.Errorf("oplatform: store get authorizer: %w", err)
	}

	mu := a.lockFor(a.appID)
	mu.Lock()
	defer mu.Unlock()

	tokens, err = a.c.store.GetAuthorizer(ctx, a.appID)
	if err == nil && time.Now().Add(authorizerTokenSafetyWindow).Before(tokens.ExpireAt) {
		return tokens.AccessToken, nil
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", fmt.Errorf("oplatform: store get authorizer: %w", err)
	}

	return a.refreshLocked(ctx, tokens.RefreshToken)
}

// Refresh 强制刷新。
func (a *AuthorizerClient) Refresh(ctx context.Context) error {
	ctx = touchContext(ctx)

	mu := a.lockFor(a.appID)
	mu.Lock()
	defer mu.Unlock()

	tokens, err := a.c.store.GetAuthorizer(ctx, a.appID)
	if err != nil {
		return fmt.Errorf("oplatform: store get authorizer: %w", err)
	}
	_, err = a.refreshLocked(ctx, tokens.RefreshToken)
	return err
}

// refreshLocked must be called with the per-appid mutex held.
func (a *AuthorizerClient) refreshLocked(ctx context.Context, refreshToken string) (string, error) {
	if refreshToken == "" {
		return "", fmt.Errorf("oplatform: no refresh_token for authorizer %s", a.appID)
	}
	componentToken, err := a.c.ComponentAccessToken(ctx)
	if err != nil {
		return "", err
	}
	q := url.Values{"component_access_token": {componentToken}}
	body := map[string]string{
		"component_appid":          a.c.cfg.ComponentAppID,
		"authorizer_appid":         a.appID,
		"authorizer_refresh_token": refreshToken,
	}
	var resp authorizerTokenResp
	path := "/cgi-bin/component/api_authorizer_token?" + q.Encode()
	if err := a.c.http.Post(ctx, path, body, &resp); err != nil {
		return "", fmt.Errorf("oplatform: api_authorizer_token: %w", err)
	}
	if resp.ErrCode == errcodeAuthorizerRevoked {
		return "", ErrAuthorizerRevoked
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return "", err
	}
	if resp.AuthorizerAccessToken == "" {
		return "", fmt.Errorf("oplatform: empty authorizer_access_token")
	}
	tokens := AuthorizerTokens{
		AccessToken:  resp.AuthorizerAccessToken,
		RefreshToken: resp.AuthorizerRefreshToken,
		ExpireAt:     time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}
	if tokens.RefreshToken == "" {
		tokens.RefreshToken = refreshToken
	}
	if err := a.c.store.SetAuthorizer(ctx, a.appID, tokens); err != nil {
		return "", fmt.Errorf("oplatform: store set authorizer: %w", err)
	}
	return tokens.AccessToken, nil
}

func (a *AuthorizerClient) lockFor(appid string) *sync.Mutex {
	if mu, ok := a.c.authMu.Load(appid); ok {
		return mu.(*sync.Mutex)
	}
	mu := &sync.Mutex{}
	actual, _ := a.c.authMu.LoadOrStore(appid, mu)
	return actual.(*sync.Mutex)
}
