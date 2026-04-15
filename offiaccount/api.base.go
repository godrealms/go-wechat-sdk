package offiaccount

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// GetAccessToken retrieves the interface call credential (access token).
// The global unique backend interface call credential is valid for 7200s; developers must store it safely.
//
// Deprecated: GetAccessToken silently drops token errors. Use AccessTokenE instead.
func (c *Client) GetAccessToken() string {
	return c.getAccessToken()
}

// GetStableAccessToken returns a stable access_token that is not invalidated by concurrent calls
// to the regular token endpoint. Pass forceRefresh=true to rotate the token immediately.
//
// 获取稳定 AccessToken
// 获取全局后台接口调用凭据，有效期最长为7200s，开发者需要进行妥善保存；
// 有两种调用模式:
//  1. 普通模式，access_token 有效期内重复调用该接口不会更新 access_token，绝大部分场景下使用该模式；
//  2. 强制刷新模式，会导致上次获取的 access_token 失效，并返回新的 access_token；
//
// 该接口调用频率限制为 1万次 每分钟，每天限制调用 50万 次；
// 与getAccessToken获取的调用凭证完全隔离，互不影响。该接口仅支持 POST JSON 形式的调用；
func (c *Client) GetStableAccessToken(forceRefresh bool) (*AccessToken, error) {
	body := map[string]interface{}{
		"grant_type":    "client_credential",
		"appid":         c.Config.AppId,
		"secret":        c.Config.AppSecret,
		"force_refresh": forceRefresh,
	}
	result := &AccessToken{}
	if err := c.Https.Post(context.Background(), "/cgi-bin/stable_token", body, result); err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, &WeixinError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}
	c.tokenMutex.Lock()
	c.accessToken = result.AccessToken
	// Clamp TTL with a 60s floor (same as refreshAccessToken, see audit C7).
	ttl := result.ExpiresIn - 60
	if ttl < 60 {
		ttl = 60
	}
	c.expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
	c.tokenMutex.Unlock()
	return result, nil
}

// CallbackCheck performs network communication detection.
// To help developers diagnose callback connection failures, this API performs domain resolution
// and a ping operation on all IPs, returning packet loss rates and latency.
//
//	action: detection action: dns (domain resolution) / ping (ping check) / all (both)
//	checkOperator: network operator: CHINANET (Telecom) / UNICOM (Unicom) / CAP (Tencent) / DEFAULT (auto)
func (c *Client) CallbackCheck(ctx context.Context, action, checkOperator string) (*CallbackCheckResponse, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	query := url.Values{
		"access_token": {token},
	}
	body := map[string]interface{}{
		"action":         action,
		"check_operator": checkOperator,
	}
	result := &CallbackCheckResponse{}
	path := fmt.Sprintf("/cgi-bin/callback/check?%s", query.Encode())
	if err = c.doPost(ctx, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetCallbackIp retrieves the WeChat push server IP addresses.
// Returns the list of WeChat server IP addresses or IP segments used to push information
// to the developer's server.
func (c *Client) GetCallbackIp(ctx context.Context) ([]string, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	query := url.Values{
		"access_token": {token},
	}
	result := &IpList{}
	if err = c.doGet(ctx, "/cgi-bin/getcallbackip", query, result); err != nil {
		return nil, err
	}
	return result.IpList, nil
}

// GetApiDomainIP retrieves the WeChat API server IP addresses.
// Returns the list of remote IP addresses that the developer's server connects to at api.weixin.qq.com.
func (c *Client) GetApiDomainIP(ctx context.Context) ([]string, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	query := url.Values{
		"access_token": {token},
	}
	result := &IpList{}
	if err = c.doGet(ctx, "/cgi-bin/get_api_domain_ip", query, result); err != nil {
		return nil, err
	}
	return result.IpList, nil
}
