package mini_program

import "github.com/godrealms/go-wechat-sdk/core"

// KfAccount represents a customer service account
type KfAccount struct {
	KfAccount    string `json:"kf_account"`
	NickName     string `json:"nickname"`
	KfHeadImgUrl string `json:"kf_headimgurl"`
}

// KfAccountListResult is the result of GetKfAccountList
type KfAccountListResult struct {
	core.Resp
	KfList []*KfAccount `json:"kf_list"`
}

// CustomerMsgText holds text message content
type CustomerMsgText struct {
	Content string `json:"content"`
}

// CustomerMsgImage holds image message content
type CustomerMsgImage struct {
	MediaId string `json:"media_id"`
}

// CustomerMsgLink holds link message content
type CustomerMsgLink struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	ThumbUrl    string `json:"thumb_url"`
}

// CustomerMsgMiniProgram holds mini-program card message content
type CustomerMsgMiniProgram struct {
	Title        string `json:"title"`
	Pagepath     string `json:"pagepath"`
	ThumbMediaId string `json:"thumb_media_id"`
}

// SendCustomerMessageRequest is the request for SendCustomerMessage
type SendCustomerMessageRequest struct {
	ToUser      string                  `json:"touser"`
	MsgType     string                  `json:"msgtype"` // text/image/link/miniprogrampage
	Text        *CustomerMsgText        `json:"text,omitempty"`
	Image       *CustomerMsgImage       `json:"image,omitempty"`
	Link        *CustomerMsgLink        `json:"link,omitempty"`
	MiniProgram *CustomerMsgMiniProgram `json:"miniprogrampage,omitempty"`
}

// SendCustomerMessageResult is the result of SendCustomerMessage
type SendCustomerMessageResult struct {
	core.Resp
}

// TypingStatus represents typing status command
type TypingStatus string

const (
	TypingStatusTyping       TypingStatus = "Typing"
	TypingStatusCancelTyping TypingStatus = "CancelTyping"
)

// SetTypingRequest is the request for SetTyping
type SetTypingRequest struct {
	ToUser  string       `json:"touser"`
	Command TypingStatus `json:"command"`
}
