package oplatform

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// QueryAuth 用 authorization_code 换取 authorizer 的 access_token / refresh_token。
// 成功后自动写入 Store（key = authorizer_appid）。
func (c *Client) QueryAuth(ctx context.Context, authCode string) (*AuthorizationInfo, error) {
	ctx = touchContext(ctx)
	if authCode == "" {
		return nil, fmt.Errorf("oplatform: authCode is required")
	}
	token, err := c.ComponentAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{"component_access_token": {token}}
	body := map[string]string{
		"component_appid":    c.cfg.ComponentAppID,
		"authorization_code": authCode,
	}
	var resp queryAuthResp
	path := "/cgi-bin/component/api_query_auth?" + q.Encode()
	if err := c.http.Post(ctx, path, body, &resp); err != nil {
		return nil, fmt.Errorf("oplatform: api_query_auth: %w", err)
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return nil, err
	}
	info := resp.AuthorizationInfo
	if info.AuthorizerAppID == "" {
		return nil, fmt.Errorf("oplatform: api_query_auth returned empty authorizer_appid")
	}
	// Clamp TTL with a floor (see authorizer token refresh for rationale).
	expiresIn := info.ExpiresIn
	if expiresIn < 120 {
		expiresIn = 120
	}
	tokens := AuthorizerTokens{
		AccessToken:  info.AuthorizerAccessToken,
		RefreshToken: info.AuthorizerRefreshToken,
		ExpireAt:     time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
	if err := c.store.SetAuthorizer(ctx, info.AuthorizerAppID, tokens); err != nil {
		return nil, fmt.Errorf("oplatform: store set authorizer: %w", err)
	}
	return &info, nil
}

// GetAuthorizerInfo /cgi-bin/component/api_get_authorizer_info
func (c *Client) GetAuthorizerInfo(ctx context.Context, authorizerAppID string) (*AuthorizerInfoResp, error) {
	ctx = touchContext(ctx)
	token, err := c.ComponentAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{"component_access_token": {token}}
	body := map[string]string{
		"component_appid":  c.cfg.ComponentAppID,
		"authorizer_appid": authorizerAppID,
	}
	var resp AuthorizerInfoResp
	path := "/cgi-bin/component/api_get_authorizer_info?" + q.Encode()
	if err := c.http.Post(ctx, path, body, &resp); err != nil {
		return nil, fmt.Errorf("oplatform: api_get_authorizer_info: %w", err)
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAuthorizerOption /cgi-bin/component/api_get_authorizer_option
func (c *Client) GetAuthorizerOption(ctx context.Context, authorizerAppID, optionName string) (*AuthorizerOption, error) {
	ctx = touchContext(ctx)
	token, err := c.ComponentAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{"component_access_token": {token}}
	body := map[string]string{
		"component_appid":  c.cfg.ComponentAppID,
		"authorizer_appid": authorizerAppID,
		"option_name":      optionName,
	}
	var resp AuthorizerOption
	path := "/cgi-bin/component/api_get_authorizer_option?" + q.Encode()
	if err := c.http.Post(ctx, path, body, &resp); err != nil {
		return nil, fmt.Errorf("oplatform: api_get_authorizer_option: %w", err)
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetAuthorizerOption /cgi-bin/component/api_set_authorizer_option
func (c *Client) SetAuthorizerOption(ctx context.Context, authorizerAppID, optionName, optionValue string) error {
	ctx = touchContext(ctx)
	token, err := c.ComponentAccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"component_access_token": {token}}
	body := map[string]string{
		"component_appid":  c.cfg.ComponentAppID,
		"authorizer_appid": authorizerAppID,
		"option_name":      optionName,
		"option_value":     optionValue,
	}
	var resp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	path := "/cgi-bin/component/api_set_authorizer_option?" + q.Encode()
	if err := c.http.Post(ctx, path, body, &resp); err != nil {
		return fmt.Errorf("oplatform: api_set_authorizer_option: %w", err)
	}
	return checkWeixinErr(resp.ErrCode, resp.ErrMsg)
}

// GetAuthorizerList /cgi-bin/component/api_get_authorizer_list
func (c *Client) GetAuthorizerList(ctx context.Context, offset, count int) (*AuthorizerList, error) {
	ctx = touchContext(ctx)
	token, err := c.ComponentAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{"component_access_token": {token}}
	payload := struct {
		ComponentAppID string `json:"component_appid"`
		Offset         int    `json:"offset"`
		Count          int    `json:"count"`
	}{c.cfg.ComponentAppID, offset, count}

	var resp AuthorizerList
	path := "/cgi-bin/component/api_get_authorizer_list?" + q.Encode()
	if err := c.http.Post(ctx, path, payload, &resp); err != nil {
		return nil, fmt.Errorf("oplatform: api_get_authorizer_list: %w", err)
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return nil, err
	}
	return &resp, nil
}
