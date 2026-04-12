package isv

import "context"

// SendText 发送文本消息。
func (cc *CorpClient) SendText(ctx context.Context, req *SendTextReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendTextReq
	}
	wire.MsgType = "text"
	wire.SendTextReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendImage 发送图片消息。
func (cc *CorpClient) SendImage(ctx context.Context, req *SendImageReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendImageReq
	}
	wire.MsgType = "image"
	wire.SendImageReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendVoice 发送语音消息。
func (cc *CorpClient) SendVoice(ctx context.Context, req *SendVoiceReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendVoiceReq
	}
	wire.MsgType = "voice"
	wire.SendVoiceReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendVideo 发送视频消息。
func (cc *CorpClient) SendVideo(ctx context.Context, req *SendVideoReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendVideoReq
	}
	wire.MsgType = "video"
	wire.SendVideoReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendFile 发送文件消息。
func (cc *CorpClient) SendFile(ctx context.Context, req *SendFileReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendFileReq
	}
	wire.MsgType = "file"
	wire.SendFileReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendTextCard 发送文本卡片消息。
func (cc *CorpClient) SendTextCard(ctx context.Context, req *SendTextCardReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendTextCardReq
	}
	wire.MsgType = "textcard"
	wire.SendTextCardReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendNews 发送图文消息。
func (cc *CorpClient) SendNews(ctx context.Context, req *SendNewsReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendNewsReq
	}
	wire.MsgType = "news"
	wire.SendNewsReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendMpNews 发送图文消息（mpnews）。
func (cc *CorpClient) SendMpNews(ctx context.Context, req *SendMpNewsReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendMpNewsReq
	}
	wire.MsgType = "mpnews"
	wire.SendMpNewsReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendMarkdown 发送 Markdown 消息。
func (cc *CorpClient) SendMarkdown(ctx context.Context, req *SendMarkdownReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendMarkdownReq
	}
	wire.MsgType = "markdown"
	wire.SendMarkdownReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendMiniProgramNotice 发送小程序通知消息。
func (cc *CorpClient) SendMiniProgramNotice(ctx context.Context, req *SendMiniProgramNoticeReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendMiniProgramNoticeReq
	}
	wire.MsgType = "miniprogram_notice"
	wire.SendMiniProgramNoticeReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendTemplateCard 发送模板卡片消息。
func (cc *CorpClient) SendTemplateCard(ctx context.Context, req *SendTemplateCardReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendTemplateCardReq
	}
	wire.MsgType = "template_card"
	wire.SendTemplateCardReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
