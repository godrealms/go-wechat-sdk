package offiaccount

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// GetAccessToken returns the current cached access token without triggering a refresh.
// Returns an empty string if no token has been fetched yet.
//
// Deprecated: use AccessTokenE for error propagation.
func (c *Client) GetAccessToken() string {
	return c.getAccessToken()
}

// GetStableAccessToken returns a stable access_token that is not invalidated by concurrent calls
// to the regular token endpoint. Pass forceRefresh=true to rotate the token immediately.
func (c *Client) GetStableAccessToken(forceRefresh bool) (*AccessToken, error) {
	body := map[string]interface{}{
		"grant_type":    "client_credential",
		"appid":         c.Config.AppId,
		"secret":        c.Config.AppSecret,
		"force_refresh": forceRefresh,
	}
	result := &AccessToken{}
	err := c.Https.Post(context.Background(), "/cgi-bin/stable_token", body, result)
	if err != nil {
		return nil, err
	}
	c.tokenMutex.Lock()
	c.accessToken = result.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(result.ExpiresIn-10) * time.Second)
	c.tokenMutex.Unlock()
	return result, nil
}

// CallbackCheck 网络通信检测
// 为了帮助开发者排查回调连接失败的问题，提供这个网络检测的API。它可以对开发者URL做域名解析，然后对所有IP进行一次ping操作，得到丢包率和耗时。
//
//	action: 检测动作：dns(域名解析)/ping(ping检测)/all(全部)
//	check_operator:	检测运营商：CHINANET(电信)/UNICOM(联通)/CAP(腾讯)/DEFAULT(自动)
func (c *Client) CallbackCheck(action, checkOperator string) (*CallbackCheckResponse, error) {
	query := url.Values{
		"access_token": {c.getAccessToken()},
	}
	body := map[string]interface{}{
		"action":         action,
		"check_operator": checkOperator,
	}
	result := &CallbackCheckResponse{}
	path := fmt.Sprintf("/cgi-bin/callback/check?%s", query.Encode())
	err := c.Https.Post(context.Background(), path, body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetCallbackIp returns the list of WeChat server IP addresses used for callback (push message)
// delivery to the official account's server.
func (c *Client) GetCallbackIp() ([]string, error) {
	query := url.Values{
		"access_token": {c.getAccessToken()},
	}
	var result = &IpList{}
	err := c.Https.Get(context.Background(), "/cgi-bin/getcallbackip", query, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, &WeixinError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}

	return result.IpList, nil
}

// GetApiDomainIP returns the IP addresses of the WeChat API servers,
// used for outbound IP allowlisting.
func (c *Client) GetApiDomainIP() ([]string, error) {
	query := url.Values{
		"access_token": {c.getAccessToken()},
	}
	var result = &IpList{}
	err := c.Https.Get(context.Background(), "/cgi-bin/get_api_domain_ip", query, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, &WeixinError{ErrCode: result.ErrCode, ErrMsg: result.ErrMsg}
	}
	return result.IpList, nil
}
