package offiaccount

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// GetAccessToken 获取接口调用凭据
// 获取全局唯一后台接口调用凭据，token有效期为7200s，开发者需要进行妥善保存。
func (c *Client) GetAccessToken() string {
	return c.getAccessToken()
}

// GetStableAccessToken 获取稳定 AccessToken
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
	err := c.Https.Post(context.Background(), "/cgi-bin/stable_token", body, result)
	if err != nil {
		return nil, err
	}
	// 提前10秒过期，避免临界点问题
	result.ExpiresIn = result.ExpiresIn + time.Now().Unix() - 10
	c.tokenMutex.Lock()
	c.accessToken = result
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

// GetCallbackIp 获取微信推送服务器IP
// 该接口用于获取微信推送服务器 ip 地址（向开发者服务器推送信息的微信服务器来源地址）
// 如果开发者基于安全等考虑，需要获知微信服务器的IP地址列表，以便进行相关限制，可以通过该接口获得微信服务器IP地址列表或者IP网段信息。
func (c *Client) GetCallbackIp() ([]string, error) {
	query := url.Values{
		"access_token": {c.getAccessToken()},
	}
	var result = &IpList{}
	err := c.Https.Get(context.Background(), "/cgi-bin/getcallbackip", query, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}

	return result.IpList, nil
}

// GetApiDomainIP 获取微信API服务器IP
// 该接口用于获取微信 api 服务器 ip 地址（开发者服务器主动访问 api.weixin.qq.com 的远端地址）
// 如果开发者基于安全等考虑，需要获知微信服务器的IP地址列表，以便进行相关限制，可以通过该接口获得微信服务器IP地址列表或者IP网段信息。
func (c *Client) GetApiDomainIP() ([]string, error) {
	query := url.Values{
		"access_token": {c.getAccessToken()},
	}
	var result = &IpList{}
	err := c.Https.Get(context.Background(), "/cgi-bin/get_api_domain_ip", query, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return result.IpList, nil
}
