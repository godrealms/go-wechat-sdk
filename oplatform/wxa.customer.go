package oplatform

import "context"

// SendCustomerMessage 发送客服消息。req.MsgType 合法值：
// "text" / "image" / "link" / "miniprogrampage"，与之对应的 payload
// 字段（Text/Image/Link/MiniProgramPage）填一个即可；其它字段会被
// omitempty 忽略。SDK 不做字段互斥校验，保持最薄封装。
// POST /cgi-bin/message/custom/send
func (w *WxaAdminClient) SendCustomerMessage(ctx context.Context, req *WxaSendCustomerMessageReq) error {
	return w.doPost(ctx, "/cgi-bin/message/custom/send", req, nil)
}

// SendCustomerTyping 下发"正在输入"状态。command 合法值：
// "Typing" 或 "CancelTyping"。
// POST /cgi-bin/message/custom/typing
func (w *WxaAdminClient) SendCustomerTyping(ctx context.Context, toUser, command string) error {
	body := map[string]string{
		"touser":  toUser,
		"command": command,
	}
	return w.doPost(ctx, "/cgi-bin/message/custom/typing", body, nil)
}
