package isv

// ---- message/send common ----

// MessageHeader 是所有消息发送请求共用的头部字段。
type MessageHeader struct {
	ToUser                 string `json:"touser,omitempty"`
	ToParty                string `json:"toparty,omitempty"`
	ToTag                  string `json:"totag,omitempty"`
	AgentID                int    `json:"agentid"`
	Safe                   int    `json:"safe,omitempty"`
	EnableIDTrans          int    `json:"enable_id_trans,omitempty"`
	EnableDuplicateCheck   int    `json:"enable_duplicate_check,omitempty"`
	DuplicateCheckInterval int    `json:"duplicate_check_interval,omitempty"`
}

// SendMessageResp 是 message/send 的统一响应。
type SendMessageResp struct {
	InvalidUser    string `json:"invaliduser"`
	InvalidParty   string `json:"invalidparty"`
	InvalidTag     string `json:"invalidtag"`
	UnlicensedUser string `json:"unlicenseduser"`
	MsgID          string `json:"msgid"`
	ResponseCode   string `json:"response_code"`
}

// ---- text ----

// TextContent 文本消息内容。
type TextContent struct {
	Content string `json:"content"`
}

// SendTextReq 文本消息请求。
type SendTextReq struct {
	MessageHeader
	Text TextContent `json:"text"`
}

// ---- image ----

// ImageContent 图片消息内容。
type ImageContent struct {
	MediaID string `json:"media_id"`
}

// SendImageReq 图片消息请求。
type SendImageReq struct {
	MessageHeader
	Image ImageContent `json:"image"`
}

// ---- voice ----

// VoiceContent 语音消息内容。
type VoiceContent struct {
	MediaID string `json:"media_id"`
}

// SendVoiceReq 语音消息请求。
type SendVoiceReq struct {
	MessageHeader
	Voice VoiceContent `json:"voice"`
}

// ---- video ----

// VideoContent 视频消息内容。
type VideoContent struct {
	MediaID     string `json:"media_id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// SendVideoReq 视频消息请求。
type SendVideoReq struct {
	MessageHeader
	Video VideoContent `json:"video"`
}

// ---- file ----

// FileContent 文件消息内容。
type FileContent struct {
	MediaID string `json:"media_id"`
}

// SendFileReq 文件消息请求。
type SendFileReq struct {
	MessageHeader
	File FileContent `json:"file"`
}

// ---- textcard ----

// TextCardContent 文本卡片消息内容。
type TextCardContent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	BtnTxt      string `json:"btntxt,omitempty"`
}

// SendTextCardReq 文本卡片消息请求。
type SendTextCardReq struct {
	MessageHeader
	TextCard TextCardContent `json:"textcard"`
}

// ---- news ----

// NewsArticle 图文消息条目。
type NewsArticle struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	PicURL      string `json:"picurl,omitempty"`
	AppID       string `json:"appid,omitempty"`
	PagePath    string `json:"pagepath,omitempty"`
}

// NewsContent 图文消息内容。
type NewsContent struct {
	Articles []NewsArticle `json:"articles"`
}

// SendNewsReq 图文消息请求。
type SendNewsReq struct {
	MessageHeader
	News NewsContent `json:"news"`
}

// ---- mpnews ----

// MpNewsArticle 图文消息（mpnews）条目。
type MpNewsArticle struct {
	Title            string `json:"title"`
	ThumbMediaID     string `json:"thumb_media_id"`
	Author           string `json:"author,omitempty"`
	ContentSourceURL string `json:"content_source_url,omitempty"`
	Content          string `json:"content"`
	Digest           string `json:"digest,omitempty"`
}

// MpNewsContent 图文消息（mpnews）内容。
type MpNewsContent struct {
	Articles []MpNewsArticle `json:"articles"`
}

// SendMpNewsReq 图文消息（mpnews）请求。
type SendMpNewsReq struct {
	MessageHeader
	MpNews MpNewsContent `json:"mpnews"`
}

// ---- markdown ----

// MarkdownContent Markdown 消息内容。
type MarkdownContent struct {
	Content string `json:"content"`
}

// SendMarkdownReq Markdown 消息请求。
type SendMarkdownReq struct {
	MessageHeader
	Markdown MarkdownContent `json:"markdown"`
}

// ---- miniprogram_notice ----

// MiniProgramContentItem 小程序通知的 content_item 条目。
type MiniProgramContentItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// MiniProgramNoticeContent 小程序通知消息内容。
type MiniProgramNoticeContent struct {
	AppID             string                   `json:"appid"`
	Page              string                   `json:"page,omitempty"`
	Title             string                   `json:"title"`
	Description       string                   `json:"description,omitempty"`
	EmphasisFirstItem bool                     `json:"emphasis_first_item,omitempty"`
	ContentItem       []MiniProgramContentItem `json:"content_item,omitempty"`
}

// SendMiniProgramNoticeReq 小程序通知消息请求。
type SendMiniProgramNoticeReq struct {
	MessageHeader
	MiniProgramNotice MiniProgramNoticeContent `json:"miniprogram_notice"`
}
