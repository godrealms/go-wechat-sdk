package offiaccount

import (
	"errors"
	"fmt"
)

// GetApiQuota 获取接口调用次数
// 本接口用于查询公众号/服务号/小程序/小游戏/微信小店/带货助手/视频号助手/联盟带货机构/移动应用/网站应用/多端应用/第三方平台等接口的每日调用接口的额度，调用次数，频率限制。
func (c *Client) GetApiQuota(cgiPath string) (*ApiQuotaResp, error) {
	path := fmt.Sprintf("/cgi-bin/openapi/quota/get?access_token=%s", c.GetAccessToken())
	body := map[string]interface{}{
		"cgi_path": cgiPath,
	}
	result := &ApiQuotaResp{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return result, nil
}

// ClearQuota 重置API调用次数
// 本接口用于清空公众号/服务号/小程序/小游戏/微信小店/带货助手/视频号助手/联盟带货机构/移动应用/网站应用/多端应用/第三方平台等接口的每日调用接口次数。
func (c *Client) ClearQuota() error {
	path := fmt.Sprintf("/cgi-bin/clear_quota?access_token=%s", c.GetAccessToken())
	body := map[string]interface{}{
		"appid": c.Config.AppId,
	}
	result := &Resp{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return err
	} else if result.ErrCode != 0 {
		return errors.New(result.ErrMsg)
	}
	return nil
}

// GetRidInfo 查询rid信息
// 本接口用于查询调用公众号/服务号/小程序/小游戏/微信小店/带货助手/视频号助手/联盟带货机构/移动应用/网站应用/多端应用/第三方平台等接口报错返回的rid详情信息，辅助开发者高效定位问题。
// rid 为调用接口时返回的rid参数
func (c *Client) GetRidInfo(rid string) (*RidInfoResp, error) {
	path := fmt.Sprintf("/cgi-bin/openapi/rid/get?access_token=%s", c.GetAccessToken())
	body := map[string]interface{}{
		"rid": rid,
	}
	var result = &RidInfoResp{}
	err := c.Https.Post(c.ctx, path, body, &result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return result, nil
}

// ClearQuotaByAppSecret 使用AppSecret重置API调用次数
// 本接口用于清空公众号/服务号/小程序/小游戏/微信小店/带货助手/视频号助手/联盟带货机构/移动应用/网站应用/多端应用等接口的每日调用接口次数。
func (c *Client) ClearQuotaByAppSecret() error {
	path := fmt.Sprintf("/cgi-bin/clear_quota/v2?appid=%s&appsecret=%s", c.Config.AppId, c.Config.AppSecret)
	result := &Resp{}
	err := c.Https.Post(c.ctx, path, nil, result)
	if err != nil {
		return err
	} else if result.ErrCode != 0 {
		return errors.New(result.ErrMsg)
	}
	return nil
}
