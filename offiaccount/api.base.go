package offiaccount

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// GetStableAccessToken returns a stable access_token that is not
// invalidated by concurrent calls to the regular /cgi-bin/token endpoint.
// Pass forceRefresh=true to rotate the token immediately (which WILL
// invalidate the previous stable token).
//
// The stable_token endpoint is rate-limited to 10,000 calls/minute and
// 500,000 calls/day. Its token pool is isolated from the one populated
// by AccessTokenE / /cgi-bin/token — the two do not interfere.
//
// On success, the returned token is also written to the client's
// in-process token cache with a 60s TTL floor, so subsequent
// AccessTokenE calls will see it until expiry.
//
// Breaking change vs pre-audit signature: this method used to ignore
// its context and call context.Background() internally. It now takes a
// ctx parameter like every other method on the client.
func (c *Client) GetStableAccessToken(ctx context.Context, forceRefresh bool) (*AccessToken, error) {
	if c.Config == nil || c.Config.AppId == "" || c.Config.AppSecret == "" {
		return nil, fmt.Errorf("offiaccount: AppId and AppSecret are required")
	}
	body := map[string]interface{}{
		"grant_type":    "client_credential",
		"appid":         c.Config.AppId,
		"secret":        c.Config.AppSecret,
		"force_refresh": forceRefresh,
	}
	result := &AccessToken{}
	if err := c.Https.Post(ctx, "/cgi-bin/stable_token", body, result); err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, &WeixinError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}
	if result.AccessToken == "" {
		return nil, fmt.Errorf("offiaccount: stable_token returned empty access_token")
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
