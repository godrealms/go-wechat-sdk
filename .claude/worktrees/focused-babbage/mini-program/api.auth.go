package mini_program

import (
	"net/url"
	"github.com/godrealms/go-wechat-sdk/core"
)

// Code2Session 登录凭证校验
// Uses appid+secret directly, no access_token needed
// GET /sns/jscode2session
func (c *Client) Code2Session(jsCode string) (*Code2SessionResult, error) {
	query := url.Values{
		"appid":      {c.Config.AppId},
		"secret":     {c.Config.AppSecret},
		"js_code":    {jsCode},
		"grant_type": {"authorization_code"},
	}
	result := &Code2SessionResult{}
	err := c.Https.Get(c.Ctx, "/sns/jscode2session", query, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// CheckSessionKey 检验登录态
// GET /wxa/checksession
func (c *Client) CheckSessionKey(openid, signature, sigMethod string) error {
	query := c.TokenQuery(url.Values{
		"openid":     {openid},
		"signature":  {signature},
		"sig_method": {sigMethod},
	})
	result := &core.Resp{}
	err := c.Https.Get(c.Ctx, "/wxa/checksession", query, result)
	if err != nil {
		return err
	}
	return result.GetError()
}

// ResetSessionKey 重置登录态
// GET /wxa/resetusersessionkey
func (c *Client) ResetSessionKey(openid, signature, sigMethod string) error {
	query := c.TokenQuery(url.Values{
		"openid":     {openid},
		"signature":  {signature},
		"sig_method": {sigMethod},
	})
	result := &core.Resp{}
	err := c.Https.Get(c.Ctx, "/wxa/resetusersessionkey", query, result)
	if err != nil {
		return err
	}
	return result.GetError()
}
