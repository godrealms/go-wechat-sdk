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

// ---- template_card ----

// TCSource 卡片来源样式。
type TCSource struct {
	IconURL   string `json:"icon_url,omitempty"`
	Desc      string `json:"desc,omitempty"`
	DescColor int    `json:"desc_color,omitempty"` // 0 灰 / 1 黑 / 2 红 / 3 绿
}

// TCActionMenuItem 右上角菜单项。
type TCActionMenuItem struct {
	Text string `json:"text"`
	Key  string `json:"key"`
}

// TCActionMenu 右上角菜单。
type TCActionMenu struct {
	Desc       string             `json:"desc,omitempty"`
	ActionList []TCActionMenuItem `json:"action_list"`
}

// TCMainTitle 一级标题。
type TCMainTitle struct {
	Title string `json:"title,omitempty"`
	Desc  string `json:"desc,omitempty"`
}

// TCEmphasisContent 关键数据。
type TCEmphasisContent struct {
	Title string `json:"title,omitempty"`
	Desc  string `json:"desc,omitempty"`
}

// TCQuoteArea 引用。
type TCQuoteArea struct {
	Type      int    `json:"type,omitempty"` // 0 文本 / 1 链接
	URL       string `json:"url,omitempty"`
	Title     string `json:"title,omitempty"`
	QuoteText string `json:"quote_text,omitempty"`
}

// TCHorizontalContent 二级标题 + 文本列表。
type TCHorizontalContent struct {
	KeyName string `json:"keyname"`
	Value   string `json:"value,omitempty"`
	Type    int    `json:"type,omitempty"` // 0 文本 / 1 链接 / 2 附件 / 3 @人
	URL     string `json:"url,omitempty"`
	MediaID string `json:"media_id,omitempty"`
	UserID  string `json:"userid,omitempty"`
}

// TCJumpItem 跳转列表项。
type TCJumpItem struct {
	Type     int    `json:"type,omitempty"` // 0 链接 / 1 小程序
	Title    string `json:"title"`
	URL      string `json:"url,omitempty"`
	AppID    string `json:"appid,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// TCCardAction 整体卡片跳转。
type TCCardAction struct {
	Type     int    `json:"type"` // 1 链接 / 2 小程序
	URL      string `json:"url,omitempty"`
	AppID    string `json:"appid,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// TCOption 选项（投票/多选共用）。
type TCOption struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsChecked bool   `json:"is_checked,omitempty"`
}

// TCButton 按钮。
type TCButton struct {
	Text  string `json:"text"`
	Style int    `json:"style,omitempty"` // 1 常规 / 2 强调
	Key   string `json:"key"`
}

// TCButtonSelection 下拉按钮。
type TCButtonSelection struct {
	QuestionKey string     `json:"question_key"`
	Title       string     `json:"title,omitempty"`
	OptionList  []TCOption `json:"option_list"`
	SelectedID  string     `json:"selected_id,omitempty"`
}

// TCSelectItem 多选列表单项。
type TCSelectItem struct {
	QuestionKey string     `json:"question_key"`
	Title       string     `json:"title,omitempty"`
	SelectedID  string     `json:"selected_id,omitempty"`
	OptionList  []TCOption `json:"option_list"`
}

// TCCheckbox 多选框。
type TCCheckbox struct {
	QuestionKey string     `json:"question_key"`
	OptionList  []TCOption `json:"option_list"`
	Mode        int        `json:"mode,omitempty"` // 0 多选 / 1 单选
}

// TCSubmitButton 提交按钮。
type TCSubmitButton struct {
	Text string `json:"text"`
	Key  string `json:"key"`
}

// TCCardImage 卡片图片（news_notice 子类型）。
type TCCardImage struct {
	URL         string  `json:"url"`
	AspectRatio float64 `json:"aspect_ratio,omitempty"`
}

// TCImageTextArea 左图右文（news_notice 子类型）。
type TCImageTextArea struct {
	Type     int    `json:"type,omitempty"`
	URL      string `json:"url,omitempty"`
	Title    string `json:"title,omitempty"`
	Desc     string `json:"desc,omitempty"`
	ImageURL string `json:"image_url"`
}

// TCVerticalContent 竖向内容。
type TCVerticalContent struct {
	Title string `json:"title"`
	Desc  string `json:"desc,omitempty"`
}

// TemplateCardContent 是 template_card 消息的 content 结构。
// card_type 决定哪些字段有效:
//   - text_notice / news_notice: 基本字段
//   - button_interaction: + ButtonSelection + ButtonList
//   - vote_interaction: + ButtonSelection + ButtonList
//   - multiple_interaction: + SelectList + Checkbox + SubmitButton
type TemplateCardContent struct {
	CardType              string                `json:"card_type"`
	Source                *TCSource             `json:"source,omitempty"`
	ActionMenu            *TCActionMenu         `json:"action_menu,omitempty"`
	TaskID                string                `json:"task_id,omitempty"`
	MainTitle             TCMainTitle           `json:"main_title"`
	EmphasisContent       *TCEmphasisContent    `json:"emphasis_content,omitempty"`
	QuoteArea             *TCQuoteArea          `json:"quote_area,omitempty"`
	SubTitleText          string                `json:"sub_title_text,omitempty"`
	HorizontalContentList []TCHorizontalContent `json:"horizontal_content_list,omitempty"`
	JumpList              []TCJumpItem          `json:"jump_list,omitempty"`
	CardAction            TCCardAction          `json:"card_action"`
	ButtonSelection       *TCButtonSelection    `json:"button_selection,omitempty"`
	ButtonList            []TCButton            `json:"button_list,omitempty"`
	SelectList            []TCSelectItem        `json:"select_list,omitempty"`
	Checkbox              *TCCheckbox           `json:"checkbox,omitempty"`
	SubmitButton          *TCSubmitButton       `json:"submit_button,omitempty"`
	CardImage             *TCCardImage          `json:"card_image,omitempty"`
	ImageTextArea         *TCImageTextArea      `json:"image_text_area,omitempty"`
	VerticalContentList   []TCVerticalContent   `json:"vertical_content_list,omitempty"`
}

// SendTemplateCardReq 模板卡片消息请求。
type SendTemplateCardReq struct {
	MessageHeader
	TemplateCard TemplateCardContent `json:"template_card"`
}
