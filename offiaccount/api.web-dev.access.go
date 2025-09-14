package offiaccount

import (
	"fmt"
	"net/url"
)

// GetSnsAccessToken 通过code换取网页授权access_token
// code: 填写第一步获取的code参数
func (c *Client) GetSnsAccessToken(code string) (*SnsAccessToken, error) {
	// 构造请求URL
	path := fmt.Sprintf("/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		c.Config.AppId, c.Config.AppSecret, code)

	// 发送请求
	var result SnsAccessToken
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// RefreshSnsAccessToken 刷新用户授权凭证
// refreshToken: 填写通过access_token获取到的refresh_token参数
func (c *Client) RefreshSnsAccessToken(refreshToken string) (*SnsAccessToken, error) {
	// 构造请求URL
	path := fmt.Sprintf("/sns/oauth2/refresh_token?appid=%s&grant_type=refresh_token&refresh_token=%s",
		c.Config.AppId, refreshToken)

	// 发送请求
	var result SnsAccessToken
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSnsUserInfo 拉取用户信息(需scope为 snsapi_userinfo)
// accessToken: 网页授权接口调用凭证
// openID: 用户的唯一标识
// lang: 国家地区语言版本，zh_CN 简体，zh_TW 繁体，en 英语
func (c *Client) GetSnsUserInfo(accessToken, openID, lang string) (*SnsUserInfo, error) {
	// 如果语言参数为空，默认使用简体中文
	if lang == "" {
		lang = "zh_CN"
	}

	// 构造请求URL
	path := fmt.Sprintf("/sns/userinfo?access_token=%s&openid=%s&lang=%s",
		url.QueryEscape(accessToken), openID, lang)

	// 发送请求
	var result SnsUserInfo
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CheckSnsAccessToken 检验授权凭证access_token是否有效
// accessToken: 网页授权接口调用凭证
// openID: 用户的唯一标识
func (c *Client) CheckSnsAccessToken(accessToken, openID string) (*Resp, error) {
	// 构造请求URL
	path := fmt.Sprintf("/sns/auth?access_token=%s&openid=%s",
		url.QueryEscape(accessToken), openID)

	// 发送请求
	var result Resp
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
