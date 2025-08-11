package offiaccount

import (
	"errors"
	"fmt"
	"net/url"
)

// CreateCustomMenu 创建自定义菜单
// 该接口用于创建公众号/服务号的自定义菜单。
func (c *Client) CreateCustomMenu(body *CreateMenuButton) error {
	path := fmt.Sprintf("/cgi-bin/menu/create?access_token=%s", c.GetAccessToken())
	result := &Resp{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return err
	} else if result.ErrCode != 0 {
		return errors.New(result.ErrMsg)
	}
	return nil
}

// GetCurrentSelfMenuInfo 查询自定义菜单信息
// 本接口提供公众号当前使用的自定义菜单的配置，如果公众号是通过API调用设置的菜单，则返回菜单的开发配置，
// 而如果公众号是在公众平台官网通过网站功能发布菜单，则本接口返回运营者设置的菜单配置。
func (c *Client) GetCurrentSelfMenuInfo() (*SelfMenu, error) {
	query := url.Values{
		"access_token": {c.GetAccessToken()},
	}
	result := &SelfMenu{}
	err := c.Https.Get(c.ctx, "/cgi-bin/get_current_selfmenu_info", query, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetMenu 获取自定义菜单配置
// 使用接口创建自定义菜单后，开发者还可使用接口查询自定义菜单的结构。
func (c *Client) GetMenu() (*QueryCustomMenu, error) {
	query := url.Values{
		"access_token": {c.GetAccessToken()},
	}
	result := &QueryCustomMenu{}
	err := c.Https.Get(c.ctx, "/cgi-bin/menu/get", query, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteMenu 删除自定义菜单
// 删除当前使用的自定义菜单。注意：调用此接口会删除默认菜单及全部个性化菜单。
func (c *Client) DeleteMenu() error {
	query := url.Values{
		"access_token": {c.GetAccessToken()},
	}
	result := &Resp{}
	err := c.Https.Get(c.ctx, "/cgi-bin/menu/delete", query, result)
	if err != nil {
		return err
	} else if result.ErrCode != 0 {
		return errors.New(result.ErrMsg)
	}
	return nil
}

// AddConditionalMenu 创建个性化菜单
// 为了帮助公众号实现灵活的业务运营，微信公众平台新增了个性化菜单接口，开发者可以通过该接口，让公众号的不同用户群体看到不一样的自定义菜单。
//
// 开发者可以通过以下条件来设置用户看到的菜单：
//
// 用户标签（开发者的业务需求可以借助用户标签来完成）
// 使用普通自定义菜单查询接口可以获取默认菜单和全部个性化菜单信息，请见自定义菜单查询接口的说明。
// 使用普通自定义菜单删除接口可以删除所有自定义菜单（包括默认菜单和全部个性化菜单），请见自定义菜单删除接口的说明。
func (c *Client) AddConditionalMenu(body *ConditionalMenu) (*AddConditionalMenuResponse, error) {
	path := fmt.Sprintf("/cgi-bin/menu/addconditional?access_token=%s", c.GetAccessToken())
	result := &AddConditionalMenuResponse{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return result, nil
}

// DeleteConditionalMenu 删除个性化菜单
// 删除指定个性化菜单
func (c *Client) DeleteConditionalMenu(menuId string) (*Resp, error) {
	body := map[string]string{"menuid": menuId}
	path := fmt.Sprintf("/cgi-bin/menu/delconditional?access_token=%s", c.GetAccessToken())
	result := &Resp{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return result, nil
}

// TryMatchMenu 测试个性化菜单匹配结果
// 测试个性化菜单，测试用户看到的菜单配置。
func (c *Client) TryMatchMenu(userId string) (*QueryCustomMenu, error) {
	body := map[string]string{"user_id": userId}
	path := fmt.Sprintf("/cgi-bin/menu/trymatch?access_token=%s", c.GetAccessToken())
	result := &QueryCustomMenu{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return result, nil
}
