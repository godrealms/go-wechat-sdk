package offiaccount

import (
	"fmt"
	"net/url"
)

// CreateQRCodeRequest 二维码请求参数结构体
type CreateQRCodeRequest struct {
	ExpireSeconds int64      `json:"expire_seconds,omitempty"`
	ActionName    string     `json:"action_name"`
	ActionInfo    ActionInfo `json:"action_info"`
}

// ActionInfo 包含二维码的具体信息
type ActionInfo struct {
	Scene Scene `json:"scene"`
}

// Scene 场景信息
type Scene struct {
	SceneID  int    `json:"scene_id,omitempty"`
	SceneStr string `json:"scene_str,omitempty"`
}

// CreateQRCodeResult 二维码结果结构体
type CreateQRCodeResult struct {
	Ticket        string `json:"ticket"`
	ExpireSeconds int    `json:"expire_seconds"`
	URL           string `json:"url"`
}

// CreateQRCode 创建二维码ticket
// req: 创建二维码请求参数
func (c *Client) CreateQRCode(req *CreateQRCodeRequest) (*CreateQRCodeResult, error) {
	// 构造请求URL，添加access_token参数
	path := fmt.Sprintf("/cgi-bin/qrcode/create?access_token=%s", c.GetAccessToken())

	// 发送请求
	var result CreateQRCodeResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetQRCodeURL 通过ticket换取二维码图片URL
// ticket: 二维码ticket
func (c *Client) GetQRCodeURL(ticket string) string {
	// 对ticket进行URL编码
	encodedTicket := url.QueryEscape(ticket)
	return "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=" + encodedTicket
}
