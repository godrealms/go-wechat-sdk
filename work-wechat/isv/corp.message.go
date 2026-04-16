package isv

import (
	"context"
	"fmt"
)

// validateMessageHeader enforces the cross-cutting rules for /cgi-bin/message/send:
// non-nil request and a caller-provided AgentID (WeCom rejects agentid=0).
// We do NOT validate ToUser/ToParty/ToTag here — at least one must be set per
// WeCom docs, but many callers fill these dynamically right before sending,
// and forcing the check client-side would produce false positives.
func validateMessageHeader(method string, h *MessageHeader) error {
	if h == nil {
		return fmt.Errorf("isv: %s: req is required", method)
	}
	if h.AgentID <= 0 {
		return fmt.Errorf("isv: %s: AgentID must be > 0", method)
	}
	return nil
}

// SendText 发送文本消息。
func (cc *CorpClient) SendText(ctx context.Context, req *SendTextReq) (*SendMessageResp, error) {
	if req == nil {
		return nil, fmt.Errorf("isv: SendText: req is required")
	}
	if err := validateMessageHeader("SendText", &req.MessageHeader); err != nil {
		return nil, err
	}
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
	if req == nil {
		return nil, fmt.Errorf("isv: SendImage: req is required")
	}
	if err := validateMessageHeader("SendImage", &req.MessageHeader); err != nil {
		return nil, err
	}
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
	if req == nil {
		return nil, fmt.Errorf("isv: SendVoice: req is required")
	}
	if err := validateMessageHeader("SendVoice", &req.MessageHeader); err != nil {
		return nil, err
	}
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
	if req == nil {
		return nil, fmt.Errorf("isv: SendVideo: req is required")
	}
	if err := validateMessageHeader("SendVideo", &req.MessageHeader); err != nil {
		return nil, err
	}
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
	if req == nil {
		return nil, fmt.Errorf("isv: SendFile: req is required")
	}
	if err := validateMessageHeader("SendFile", &req.MessageHeader); err != nil {
		return nil, err
	}
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
	if req == nil {
		return nil, fmt.Errorf("isv: SendTextCard: req is required")
	}
	if err := validateMessageHeader("SendTextCard", &req.MessageHeader); err != nil {
		return nil, err
	}
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
	if req == nil {
		return nil, fmt.Errorf("isv: SendNews: req is required")
	}
	if err := validateMessageHeader("SendNews", &req.MessageHeader); err != nil {
		return nil, err
	}
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
	if req == nil {
		return nil, fmt.Errorf("isv: SendMpNews: req is required")
	}
	if err := validateMessageHeader("SendMpNews", &req.MessageHeader); err != nil {
		return nil, err
	}
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
	if req == nil {
		return nil, fmt.Errorf("isv: SendMarkdown: req is required")
	}
	if err := validateMessageHeader("SendMarkdown", &req.MessageHeader); err != nil {
		return nil, err
	}
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
	if req == nil {
		return nil, fmt.Errorf("isv: SendMiniProgramNotice: req is required")
	}
	if err := validateMessageHeader("SendMiniProgramNotice", &req.MessageHeader); err != nil {
		return nil, err
	}
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
	if req == nil {
		return nil, fmt.Errorf("isv: SendTemplateCard: req is required")
	}
	if err := validateMessageHeader("SendTemplateCard", &req.MessageHeader); err != nil {
		return nil, err
	}
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
