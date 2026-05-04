package offiaccount

import "fmt"

// MsgSecCheck 检查一段文本是否含有违法违规内容
// POST /wxa/msg_sec_check (access_token in URL)
func (c *Client) MsgSecCheck(req *MsgSecCheckRequest) (*MsgSecCheckResult, error) {
	path := fmt.Sprintf("/wxa/msg_sec_check?access_token=%s", c.GetAccessToken())
	result := &MsgSecCheckResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

// ImgSecCheck 校验一张图片是否含有违法违规内容 (synchronous, image < 1MB)
// POST /wxa/img_sec_check (multipart form, access_token in URL)
func (c *Client) ImgSecCheck(imageData []byte) (*MediaCheckAsyncResult, error) {
	path := fmt.Sprintf("/wxa/img_sec_check?access_token=%s", c.GetAccessToken())
	result := &MediaCheckAsyncResult{}
	err := c.Https.PostMultipart(c.Ctx, path, "media", "image.jpg", imageData, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

// MediaCheckAsync 异步校验图片/音频是否含有违法违规内容
// POST /wxa/media_check_async (access_token in URL)
func (c *Client) MediaCheckAsync(req *MediaCheckAsyncRequest) (*MediaCheckAsyncResult, error) {
	path := fmt.Sprintf("/wxa/media_check_async?access_token=%s", c.GetAccessToken())
	result := &MediaCheckAsyncResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}
