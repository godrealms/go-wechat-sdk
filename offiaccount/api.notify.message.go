package offiaccount

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// UploadImage 上传图文消息图片
func (c *Client) UploadImage(filename, mediaType string) (*UploadImageResponse, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("the file does not exist: %s", filename)
	}

	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file: %v", err)
	}
	defer file.Close()

	// 创建multipart表单
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 添加type字段
	if mediaType == "" {
		mediaType = "image"
	}
	err = writer.WriteField("type", mediaType)
	if err != nil {
		return nil, fmt.Errorf("writing type field failed: %v", err)
	}

	// 添加文件字段
	fileName := filepath.Base(filename)
	part, err := writer.CreateFormFile("media", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create a file field: %v", err)
	}

	// 复制文件内容
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("copying file contents failed: %v", err)
	}

	// 关闭writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("closing writer failed: %v", err)
	}

	// 构建请求URL
	uploadURL := fmt.Sprintf("%s/cgi-bin/media/uploadimg?access_token=%s", c.Https.BaseURL, c.GetAccessToken())

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", uploadURL, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("the http request was created failed: %v", err)
	}

	// 设置Content-Type
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := c.Https.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending http request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %v", err)
	}
	log.Println(string(respBody))
	// 解析响应
	var uploadResp UploadImageResponse
	err = json.Unmarshal(respBody, &uploadResp)
	if err != nil {
		return nil, fmt.Errorf("the parsing response failed: %v", err)
	}

	// 检查业务错误
	if uploadResp.ErrCode != 0 {
		return &uploadResp, fmt.Errorf("wechat api error: %d - %s", uploadResp.ErrCode, uploadResp.ErrMsg)
	}

	return &uploadResp, nil
}

// DeleteMassMsg 删除群发消息
// 群发之后，随时可以通过该接口删除群发。
func (c *Client) DeleteMassMsg(body *DeleteMassMsgRequest) error {
	path := fmt.Sprintf("/cgi-bin/message/mass/delete?access_token=%s", c.GetAccessToken())
	result := &Resp{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return err
	} else if result.ErrCode != 0 {
		return errors.New(result.ErrMsg)
	}
	return nil
}

// GetSpeed 获取群发速度
// 本接口用于获取消息的群发速度
func (c *Client) getSpeed(speed int) (*SpeedResp, error) {
	path := fmt.Sprintf("/cgi-bin/message/mass/speed/get?access_token=%s", c.GetAccessToken())
	body := map[string]interface{}{"speed": speed}
	result := &SpeedResp{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return result, nil
}

// GetMassMsg 查询群发消息发送状态
// 本接口用于查询群发消息发送状态。
func (c *Client) GetMassMsg(msgId string) (*MassMsgResp, error) {
	path := fmt.Sprintf("/cgi-bin/message/mass/get?access_token=%s", c.GetAccessToken())
	body := map[string]interface{}{"msg_id": msgId}
	result := &MassMsgResp{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return result, nil
}

// MassSend 根据OpenID群发消息
// 本接口用于根据 openid 列表群发消息
func (c *Client) MassSend(body *MassSendRequest) (*MassSendResp, error) {
	path := fmt.Sprintf("/cgi-bin/message/mass/send?access_token=%s", c.GetAccessToken())
	result := &MassSendResp{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return result, nil
}

// Preview 预览群发消息
// 本接口发送消息给指定用户，在手机端查看消息的样式和排版。
func (c *Client) Preview(body *MassSendRequest) (*Resp, error) {
	path := fmt.Sprintf("/cgi-bin/message/mass/preview?access_token=%s", c.GetAccessToken())
	result := &Resp{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return result, nil
}

// SendAll 根据标签群发消息
// 本接口用于根据标签群发消息
func (c *Client) SendAll(body *MassSendByTagRequest) (*MassSendByTagResponse, error) {
	path := fmt.Sprintf("/cgi-bin/message/mass/sendall?access_token=%s", c.GetAccessToken())
	result := &MassSendByTagResponse{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return result, nil
}

// SetSpeed 设置群发速度
// 本接口用于设置消息的群发速度
func (c *Client) SetSpeed(speed int) (*Resp, error) {
	path := fmt.Sprintf("/cgi-bin/message/mass/speed/set?access_token=%s", c.GetAccessToken())
	body := map[string]interface{}{"speed": speed}
	result := &Resp{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return result, nil
}

// UploadNewsMsg 上传图文消息素材
// 本接口用于上传图文消息，该能力已更新为草稿箱
func (c *Client) UploadNewsMsg(body *AddNewsMaterialRequest) (*AddNewsMaterialResponse, error) {
	path := fmt.Sprintf("/cgi-bin/media/uploadnews?access_token=%s", c.GetAccessToken())
	result := &AddNewsMaterialResponse{}
	err := c.Https.Post(c.ctx, path, body, result)
	if err != nil {
		return nil, err
	} else if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return result, nil
}
