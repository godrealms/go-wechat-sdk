package offiaccount

import (
	"context"
	"fmt"
)

// SendTemplateMessage 发送模板消息
// 本接口用于向用户发送模板消息
func (c *Client) SendTemplateMessage(ctx context.Context, body *SubscribeMessageRequest) (*MassMsgResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/message/template/send?access_token=%s", token)
	result := &MassMsgResp{}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// AddTemplate 选用模板
// 本接口用于从类目模板库或行业模板库添加模板获得模板ID
func (c *Client) AddTemplate(ctx context.Context, templateIdShort string, keywordNameList []string) (*AddTemplateResponse, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	body := map[string]interface{}{
		"template_id_short": templateIdShort,
		"keyword_name_list": keywordNameList,
	}
	path := fmt.Sprintf("/cgi-bin/template/api_add_template?access_token=%s", token)
	result := &AddTemplateResponse{}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryBlockTmplMsg 查询被拦截的模板消息。
//
// 注意：此接口路径 /wxa/sec/queryblocktmplmsg 实际属于小程序（wxa）安全模块，
// 不属于公众号 API。保留在这里只是为了向后兼容；新代码应使用小程序 SDK。
//
// Deprecated: 此接口属于小程序安全模块而非公众号。
func (c *Client) QueryBlockTmplMsg(ctx context.Context, body *QueryBlockTmplMsgReq) (*QueryBlockTmplMsgResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/wxa/sec/queryblocktmplmsg?access_token=%s", token)
	result := &QueryBlockTmplMsgResp{}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteTemplate 删除模板
// 本接口用于删除账号下的指定模板。
func (c *Client) DeleteTemplate(ctx context.Context, templateId string) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	body := map[string]interface{}{"template_id": templateId}
	path := fmt.Sprintf("/cgi-bin/template/del_private_template?access_token=%s", token)
	result := &Resp{}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetAllTemplates 获取已选用模板列表
// 本接口用于获取已添加至账号下所有模板列表
func (c *Client) GetAllTemplates(ctx context.Context) (*TemplateList, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/template/get_all_private_template?access_token=%s", token)
	result := &TemplateList{}
	if err = c.doGet(ctx, path, nil, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetIndustry 获取行业信息
// 本接口可获取账号设置的行业信息，返回 primary_industry / secondary_industry。
// 文档：https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Template_Message_Interface.html
func (c *Client) GetIndustry(ctx context.Context) (*IndustryResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/template/get_industry?access_token=%s", token)
	result := &IndustryResp{}
	if err = c.doGet(ctx, path, nil, result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetIndustry 设置所属行业
// 本接口用于修改账号所属行业，修改行业后，你在原有行业中的模板将会被删除。
func (c *Client) SetIndustry(ctx context.Context, industryId1, industryId2 string) error {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return err
	}
	body := map[string]interface{}{
		"industry_id1": industryId1,
		"industry_id2": industryId2,
	}
	path := fmt.Sprintf("/cgi-bin/template/api_set_industry?access_token=%s", token)
	result := &Resp{}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return err
	}
	return nil
}
