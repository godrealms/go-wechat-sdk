package offiaccount

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// UploadImage 上传图文消息图片
//
// 历史实现自己拼 multipart/http 请求并绕过了统一的 errcode 检查，这里统一走
// doPostMultipart，和 UploadImg / AddMaterial 保持一致。
func (c *Client) UploadImage(ctx context.Context, filename, mediaType string) (*UploadImageResponse, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("offiaccount: UploadImage: read file: %w", err)
	}

	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/media/uploadimg?access_token=%s", token)

	// 官方 /cgi-bin/media/uploadimg 接口实际不识别 type 字段（该字段属于
	// /cgi-bin/media/upload），保留原签名但 mediaType 为空时用 "image" 兜底，
	// 作为 multipart 的附加字段写入——和历史行为保持兼容。
	if mediaType == "" {
		mediaType = "image"
	}
	extra := map[string]string{"type": mediaType}

	var result UploadImageResponse
	if err = c.doPostMultipart(ctx, path, "media", filepath.Base(filename), data, extra, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteMassMsg 删除群发消息
// 群发之后，随时可以通过该接口删除群发。
func (c *Client) DeleteMassMsg(ctx context.Context, body *DeleteMassMsgRequest) error {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/cgi-bin/message/mass/delete?access_token=%s", token)
	result := &Resp{}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return err
	}
	return nil
}

// GetSpeed 获取群发速度
// 本接口用于获取消息的群发速度。官方接口不需要任何请求参数，传空 body。
// https://developers.weixin.qq.com/doc/service/api/notify/mass/api_getspeed.html
func (c *Client) GetSpeed(ctx context.Context) (*SpeedResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/message/mass/speed/get?access_token=%s", token)
	result := &SpeedResp{}
	if err = c.doPost(ctx, path, map[string]any{}, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetMassMsg 查询群发消息发送状态
// 本接口用于查询群发消息发送状态。官方 msg_id 为数字。
func (c *Client) GetMassMsg(ctx context.Context, msgId int64) (*MassMsgResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/message/mass/get?access_token=%s", token)
	body := map[string]any{"msg_id": msgId}
	result := &MassMsgResp{}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// MassSend 根据OpenID群发消息
// 本接口用于根据 openid 列表群发消息
func (c *Client) MassSend(ctx context.Context, body *MassSendRequest) (*MassSendResp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/message/mass/send?access_token=%s", token)
	result := &MassSendResp{}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// Preview 预览群发消息
// 本接口发送消息给指定用户，在手机端查看消息的样式和排版。
func (c *Client) Preview(ctx context.Context, body *MassSendRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/message/mass/preview?access_token=%s", token)
	result := &Resp{}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// SendAll 根据标签群发消息
// 本接口用于根据标签群发消息
func (c *Client) SendAll(ctx context.Context, body *MassSendByTagRequest) (*MassSendByTagResponse, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/message/mass/sendall?access_token=%s", token)
	result := &MassSendByTagResponse{}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetSpeed 设置群发速度
// 本接口用于设置消息的群发速度
func (c *Client) SetSpeed(ctx context.Context, speed int) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/message/mass/speed/set?access_token=%s", token)
	body := map[string]any{"speed": speed}
	result := &Resp{}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}

// UploadNewsMsg 上传图文消息素材
// 本接口用于上传图文消息，该能力已更新为草稿箱
func (c *Client) UploadNewsMsg(ctx context.Context, body *AddNewsMaterialRequest) (*AddNewsMaterialResponse, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/media/uploadnews?access_token=%s", token)
	result := &AddNewsMaterialResponse{}
	if err = c.doPost(ctx, path, body, result); err != nil {
		return nil, err
	}
	return result, nil
}
