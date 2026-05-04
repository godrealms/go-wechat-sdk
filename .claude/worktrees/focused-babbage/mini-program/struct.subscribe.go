package mini_program

import "github.com/godrealms/go-wechat-sdk/core"

// SubscribeMessageValue holds the template keyword value
type SubscribeMessageValue struct {
	Value string `json:"value"`
}

// SendSubscribeMessageRequest is the request body for SendSubscribeMessage
type SendSubscribeMessageRequest struct {
	ToUser           string                            `json:"touser"`
	TemplateId       string                            `json:"template_id"`
	Page             string                            `json:"page,omitempty"`
	MiniProgramState string                            `json:"miniprogram_state,omitempty"` // developer/trial/formal
	Lang             string                            `json:"lang,omitempty"`              // zh_CN/en_US/zh_HK/zh_TW
	Data             map[string]*SubscribeMessageValue `json:"data"`
}

// SendSubscribeMessageResult is the result of SendSubscribeMessage
type SendSubscribeMessageResult struct {
	core.Resp
}
