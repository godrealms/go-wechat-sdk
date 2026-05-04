package offiaccount

import "fmt"

// CardCreate 创建卡券
// POST /card/create (access_token in URL)
func (c *Client) CardCreate(req *CardCreateRequest) (*CardCreateResult, error) {
	path := fmt.Sprintf("/card/create?access_token=%s", c.GetAccessToken())
	result := &CardCreateResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

// CardGet 查看卡券详情
// POST /card/get (access_token in URL)
func (c *Client) CardGet(cardId string) (*CardGetResult, error) {
	path := fmt.Sprintf("/card/get?access_token=%s", c.GetAccessToken())
	body := map[string]string{"card_id": cardId}
	result := &CardGetResult{}
	err := c.Https.Post(c.Ctx, path, body, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

// CardUpdate 更改卡券信息
// POST /card/update (access_token in URL)
func (c *Client) CardUpdate(req *CardUpdateRequest) error {
	path := fmt.Sprintf("/card/update?access_token=%s", c.GetAccessToken())
	result := &Resp{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

// CardDelete 删除卡券
// POST /card/delete (access_token in URL)
func (c *Client) CardDelete(cardId string) error {
	path := fmt.Sprintf("/card/delete?access_token=%s", c.GetAccessToken())
	body := map[string]string{"card_id": cardId}
	result := &Resp{}
	err := c.Https.Post(c.Ctx, path, body, result)
	if err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

// CardQRCodeCreate 生成卡券二维码
// POST /card/qrcode/create (access_token in URL)
func (c *Client) CardQRCodeCreate(req *CardQRCodeRequest) (*CardQRCodeResult, error) {
	path := fmt.Sprintf("/card/qrcode/create?access_token=%s", c.GetAccessToken())
	result := &CardQRCodeResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

// CardCodeGet 查询 code 信息
// POST /card/code/get (access_token in URL)
func (c *Client) CardCodeGet(req *CardCodeRequest) (*CardCodeResult, error) {
	path := fmt.Sprintf("/card/code/get?access_token=%s", c.GetAccessToken())
	result := &CardCodeResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

// CardCodeConsume 核销 code
// POST /card/code/consume (access_token in URL)
func (c *Client) CardCodeConsume(req *CardCodeRequest) (*ConsumeCardCodeResult, error) {
	path := fmt.Sprintf("/card/code/consume?access_token=%s", c.GetAccessToken())
	result := &ConsumeCardCodeResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

// CardList 拉取卡券概况数据
// POST /card/batchget (access_token in URL)
func (c *Client) CardList(offset, count int, statusList []string) (*CardListResult, error) {
	path := fmt.Sprintf("/card/batchget?access_token=%s", c.GetAccessToken())
	body := map[string]interface{}{
		"offset":      offset,
		"count":       count,
		"status_list": statusList,
	}
	result := &CardListResult{}
	err := c.Https.Post(c.Ctx, path, body, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}
