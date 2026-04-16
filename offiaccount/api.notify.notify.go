package offiaccount

import (
	"context"
	"fmt"
	"net/url"
)

// DelWxAnewTemplate 删除模板
// 删除私有模板库中的模板
func (c *Client) DelWxAnewTemplate(ctx context.Context, priTmplId string) error {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/wxaapi/newtmpl/deltemplate?access_token=%s", token)
	result := &Resp{}
	if err = c.doPost(ctx, path, map[string]any{"priTmplId": priTmplId}, result); err != nil {
		return err
	}
	return nil
}

// GetCategory 获取类目
// 本接口用于获取小程序、公众号所属类目用于查询公共模板
func (c *Client) GetCategory(ctx context.Context) (*CategoryResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/wxaapi/newtmpl/getcategory?access_token=%s", token)
	result := &CategoryResp{}
	if err = c.doGet(ctx, path, nil, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetPubNewTemplateKeyWords 获取模板中的关键词
// 该接口用于获取模板标题下的关键词列表。
func (c *Client) GetPubNewTemplateKeyWords(ctx context.Context, tid string) (*TemplateKeyWordsResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := "/wxaapi/newtmpl/getpubtemplatekeywords"
	query := url.Values{
		"access_token": []string{token},
		"tid":          []string{tid},
	}
	result := &TemplateKeyWordsResp{}
	if err = c.doGet(ctx, path, query, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetPubNewTemplateTitles 获取类目下的公共模板
// 该接口用于获取帐号所属类目下的公共模板，可从中选用模板使用
// ids	string	是	类目 id，多个用逗号隔开
// start	number	是	用于分页，表示从 start 开始。从 0 开始计数
// limit	number	是	用于分页，表示拉取 limit 条记录。最大为 30
func (c *Client) GetPubNewTemplateTitles(ctx context.Context, ids string, start, limit int) (*TemplateTitlesResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := "/wxaapi/newtmpl/getpubtemplatetitles"
	query := url.Values{
		"access_token": []string{token},
		"ids":          []string{ids},
		"start":        []string{fmt.Sprintf("%d", start)},
		"limit":        []string{fmt.Sprintf("%d", limit)},
	}
	result := &TemplateTitlesResp{}
	if err = c.doGet(ctx, path, query, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetWxaPubNewTemplate 获取已有模板列表
func (c *Client) GetWxaPubNewTemplate(ctx context.Context) (*PubNewTemplateResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := "/wxaapi/newtmpl/gettemplate"
	query := url.Values{
		"access_token": []string{token},
	}
	result := &PubNewTemplateResp{}
	if err = c.doGet(ctx, path, query, result); err != nil {
		return nil, err
	}
	return result, nil
}

// AddWxaNewTemplate 选用模板
//
// tid string 是 模板标题 id，可通过接口获取，也可登录小程序后台查看获取
// kidList numarray 是 开发者自行组合好的模板关键词列表，关键词顺序可以自由搭配（例如 [3,5,4] 或 [4,5,3]），最多支持5个，最少2个关键词组合
// sceneDesc string 是 服务场景描述，15个字以内
func (c *Client) AddWxaNewTemplate(ctx context.Context, tid string, kidList []int, sceneDesc string) (*AddWxaNewTemplateResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/wxaapi/newtmpl/addtemplate?access_token=%s", token)
	result := &AddWxaNewTemplateResp{}
	body := map[string]any{"tid": tid, "kidList": kidList, "sceneDesc": sceneDesc}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// SendNewSubscribeMsg 发送订阅通知
func (c *Client) SendNewSubscribeMsg(ctx context.Context, body *SubscribeMsg) error {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/cgi-bin/message/subscribe/bizsend?access_token=%s", token)
	result := &Resp{}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return err
	}
	return nil
}
