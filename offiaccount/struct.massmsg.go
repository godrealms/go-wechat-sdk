package offiaccount

// DeleteMassMsgRequest 删除群发消息请求
type DeleteMassMsgRequest struct {
	// 发送出去的消息ID
	MsgId int `json:"msg_id,omitempty"`
	// 要删除的文章在图文消息中的位置，第一篇编号为1，该字段不填或填0会删除全部文章
	ArticleIdx int `json:"article_idx"`
	// 要删除的文章url，当msg_id未指定时该参数才生效
	Url string `json:"url,omitempty"`
}

// SpeedResp 群发速度响应
type SpeedResp struct {
	Resp
	// 群发速度的级别
	Speed int `json:"speed"`
	// 群发速度的真实值 单位：万/分钟
	RealSpeed int `json:"realspeed"`
}

// MassMsgResp 群发消息响应
type MassMsgResp struct {
	Resp
	// 群发消息后返回的消息id
	MsgId int `json:"msg_id"`
	// 消息发送后的状态，SEND_SUCCESS表示发送成功，SENDING表示发送中，SEND_FAIL表示发送失败，DELETE表示已删除
	MsgStatus string `json:"msg_status"`
}

// MassSendRequest 群发消息请求结构体
type MassSendRequest struct {
	ToUser            []string `json:"touser"`                        // 接收者OpenID列表，最少2个，最多10000个
	MsgType           string   `json:"msgtype"`                       // 消息类型：mpnews/text/voice/music/image/video/wxcard
	SendIgnoreReprint int      `json:"send_ignore_reprint,omitempty"` // 转载时是否继续群发：1继续，0停止，默认0
	MPNews            *MPNews  `json:"mpnews,omitempty"`              // 图文消息
	Text              *Text    `json:"text,omitempty"`                // 文本消息
	Voice             *Voice   `json:"voice,omitempty"`               // 语音消息
	Images            *Images  `json:"images,omitempty"`              // 图片消息
	MPVideo           *MPVideo `json:"mpvideo,omitempty"`             // 视频消息
	WxCard            *WxCard  `json:"wxcard,omitempty"`              // 卡券消息
	ClientMsgID       string   `json:"clientmsgid,omitempty"`         // 开发者侧群发msgid，长度限制32字节
}

// SetMPNews 设置图文消息
func (r *MassSendRequest) SetMPNews(mediaID string) *MassSendRequest {
	r.MsgType = MsgTypeMPNews
	r.MPNews = &MPNews{MediaID: mediaID}
	return r
}

// SetText 设置文本消息
func (r *MassSendRequest) SetText(content string) *MassSendRequest {
	r.MsgType = MsgTypeText
	r.Text = &Text{Content: content}
	return r
}

// SetVoice 设置语音消息
func (r *MassSendRequest) SetVoice(mediaID string) *MassSendRequest {
	r.MsgType = MsgTypeVoice
	r.Voice = &Voice{MediaID: mediaID}
	return r
}

// SetImages 设置图片消息
func (r *MassSendRequest) SetImages(mediaIDs []string, title, recommend string) *MassSendRequest {
	r.MsgType = MsgTypeImage
	r.Images = &Images{
		MediaIDs:  mediaIDs,
		Title:     title,
		Recommend: recommend,
	}
	return r
}

// SetImagesWithComment 设置带评论功能的图片消息
func (r *MassSendRequest) SetImagesWithComment(mediaIDs []string, title, recommend string, needOpenComment, onlyFansCanComment int) *MassSendRequest {
	r.MsgType = MsgTypeImage
	r.Images = &Images{
		MediaIDs:           mediaIDs,
		Title:              title,
		Recommend:          recommend,
		NeedOpenComment:    needOpenComment,
		OnlyFansCanComment: onlyFansCanComment,
	}
	return r
}

// SetMPVideo 设置视频消息
func (r *MassSendRequest) SetMPVideo(mediaID, title, description string) *MassSendRequest {
	r.MsgType = MsgTypeVideo
	r.MPVideo = &MPVideo{
		MediaID:     mediaID,
		Title:       title,
		Description: description,
	}
	return r
}

// SetWxCard 设置卡券消息
func (r *MassSendRequest) SetWxCard(cardID string) *MassSendRequest {
	r.MsgType = MsgTypeWxCard
	r.WxCard = &WxCard{CardID: cardID}
	return r
}

// SetSendIgnoreReprint 设置转载时是否继续群发
func (r *MassSendRequest) SetSendIgnoreReprint(ignore bool) *MassSendRequest {
	if ignore {
		r.SendIgnoreReprint = 1
	} else {
		r.SendIgnoreReprint = 0
	}
	return r
}

// SetClientMsgID 设置开发者侧群发msgid
func (r *MassSendRequest) SetClientMsgID(clientMsgID string) *MassSendRequest {
	if len(clientMsgID) <= 32 {
		r.ClientMsgID = clientMsgID
	}
	return r
}

// MPNews 图文消息结构体
type MPNews struct {
	MediaID string `json:"media_id,omitempty"` // 用于群发的图文消息的media_id
}

// Text 文本消息结构体
type Text struct {
	Content string `json:"content,omitempty"` // 文本内容
}

// Voice 语音消息结构体
type Voice struct {
	MediaID string `json:"media_id,omitempty"` // 用于群发的语音消息的media_id
}

// Images 图片消息结构体
type Images struct {
	MediaIDs           []string `json:"media_ids,omitempty"`             // 用于群发的图片消息的media_id列表
	Recommend          string   `json:"recommend,omitempty"`             // 推荐语
	Title              string   `json:"title,omitempty"`                 // 标题
	NeedOpenComment    int      `json:"need_open_comment,omitempty"`     // 开启评论：1开启，0关闭
	OnlyFansCanComment int      `json:"only_fans_can_comment,omitempty"` // 只有粉丝能评论：1开启，0关闭
}

// MPVideo 视频消息结构体
type MPVideo struct {
	MediaID     string `json:"media_id,omitempty"`    // 用于群发的视频消息的media_id
	Title       string `json:"title,omitempty"`       // 标题
	Description string `json:"description,omitempty"` // 描述
}

// WxCard 卡券消息结构体
type WxCard struct {
	CardID string `json:"card_id,omitempty"` // 卡券ID
}

// NewMassSendRequest 创建群发消息请求
func NewMassSendRequest(toUsers []string, msgType string) *MassSendRequest {
	return &MassSendRequest{
		ToUser:  toUsers,
		MsgType: msgType,
	}
}

// 消息类型常量
const (
	MsgTypeMPNews = "mpnews" // 图文消息
	MsgTypeText   = "text"   // 文本消息
	MsgTypeVoice  = "voice"  // 语音消息
	MsgTypeMusic  = "music"  // 音乐消息
	MsgTypeImage  = "image"  // 图片消息
	MsgTypeVideo  = "video"  // 视频消息
	MsgTypeWxCard = "wxcard" // 卡券消息
)

// MassSendResp 群发消息响应
type MassSendResp struct {
	Resp
	// 媒体文件类型，分别有图片（image）、语音（voice）、视频（video）和缩略图（thumb），次数为news，即图文消息
	Type string `json:"type"`
	// 消息发送任务的ID
	MsgId int `json:"msg_id"`
	// 消息的数据ID，，该字段只有在群发图文消息时，才会出现。
	// 可以用于在图文分析数据接口中，获取到对应的图文消息的数据，
	// 是图文分析数据接口中的msgid字段中的前半部分，
	// 详见图文分析数据接口中的msgid字段的介绍。
	MsgDataId int `json:"msg_data_id"`
}

// MassSendByTagRequest 按标签群发消息请求结构体
type MassSendByTagRequest struct {
	Filter      *Filter  `json:"filter"`                // 用于设定图文消息的接收者
	MsgType     string   `json:"msgtype"`               // 消息类型：mpnews/text/voice/music/image/video/wxcard
	MPNews      *MPNews  `json:"mpnews,omitempty"`      // 图文消息
	Text        *Text    `json:"text,omitempty"`        // 文本消息
	Voice       *Voice   `json:"voice,omitempty"`       // 语音消息
	Images      *Images  `json:"images,omitempty"`      // 图片消息
	MPVideo     *MPVideo `json:"mpvideo,omitempty"`     // 视频消息
	WxCard      *WxCard  `json:"wxcard,omitempty"`      // 卡券消息
	ClientMsgID string   `json:"clientmsgid,omitempty"` // 开发者侧群发msgid，长度限制32字节
}

// Filter 接收者过滤器结构体
type Filter struct {
	IsToAll string `json:"is_to_all,omitempty"` // 是否向全部用户发送：true或false
	TagID   string `json:"tag_id,omitempty"`    // 群发到的标签的tag_id，is_to_all为true时可不填
}

// MassSendByTagResponse 按标签群发消息响应结构体
type MassSendByTagResponse struct {
	Resp
	Type      string `json:"type"`        // 媒体文件类型，分别有图片（image）、语音（voice）、视频（video）和缩略图（thumb），次数为news，即图文消息
	MsgID     int64  `json:"msg_id"`      // 消息发送任务的ID
	MsgDataID int64  `json:"msg_data_id"` // 消息的数据ID
}

// 构造函数和辅助方法

// NewMassSendByTagRequest 创建按标签群发消息请求
func NewMassSendByTagRequest(msgType string) *MassSendByTagRequest {
	return &MassSendByTagRequest{
		MsgType: msgType,
		Filter:  &Filter{},
	}
}

// NewMassSendToTag 创建发送给指定标签用户的请求
func NewMassSendToTag(msgType, tagID string) *MassSendByTagRequest {
	return NewMassSendByTagRequest(msgType).SetToTag(tagID)
}

// SetToAll 设置发送给所有用户
func (r *MassSendByTagRequest) SetToAll() *MassSendByTagRequest {
	r.Filter.IsToAll = "true"
	r.Filter.TagID = "" // 发送给所有用户时清空tag_id
	return r
}

// SetToTag 设置发送给指定标签的用户
func (r *MassSendByTagRequest) SetToTag(tagID string) *MassSendByTagRequest {
	r.Filter.IsToAll = "false"
	r.Filter.TagID = tagID
	return r
}

// SetMPNews 设置图文消息
func (r *MassSendByTagRequest) SetMPNews(mediaID string) *MassSendByTagRequest {
	r.MsgType = MsgTypeMPNews
	r.MPNews = &MPNews{MediaID: mediaID}
	return r
}

// SetText 设置文本消息
func (r *MassSendByTagRequest) SetText(content string) *MassSendByTagRequest {
	r.MsgType = MsgTypeText
	r.Text = &Text{Content: content}
	return r
}

// SetVoice 设置语音消息
func (r *MassSendByTagRequest) SetVoice(mediaID string) *MassSendByTagRequest {
	r.MsgType = MsgTypeVoice
	r.Voice = &Voice{MediaID: mediaID}
	return r
}

// SetImages 设置图片消息
func (r *MassSendByTagRequest) SetImages(mediaIDs []string, title, recommend string) *MassSendByTagRequest {
	r.MsgType = MsgTypeImage
	r.Images = &Images{
		MediaIDs:  mediaIDs,
		Title:     title,
		Recommend: recommend,
	}
	return r
}

// SetImagesWithComment 设置带评论功能的图片消息
func (r *MassSendByTagRequest) SetImagesWithComment(mediaIDs []string, title, recommend string, needOpenComment, onlyFansCanComment int) *MassSendByTagRequest {
	r.MsgType = MsgTypeImage
	r.Images = &Images{
		MediaIDs:           mediaIDs,
		Title:              title,
		Recommend:          recommend,
		NeedOpenComment:    needOpenComment,
		OnlyFansCanComment: onlyFansCanComment,
	}
	return r
}

// SetMPVideo 设置视频消息
func (r *MassSendByTagRequest) SetMPVideo(mediaID string) *MassSendByTagRequest {
	r.MsgType = MsgTypeVideo
	r.MPVideo = &MPVideo{MediaID: mediaID}
	return r
}

// SetWxCard 设置卡券消息
func (r *MassSendByTagRequest) SetWxCard(cardID string) *MassSendByTagRequest {
	r.MsgType = MsgTypeWxCard
	r.WxCard = &WxCard{CardID: cardID}
	return r
}

// SetClientMsgID 设置开发者侧群发msgid
func (r *MassSendByTagRequest) SetClientMsgID(clientMsgID string) *MassSendByTagRequest {
	if len(clientMsgID) <= 32 {
		r.ClientMsgID = clientMsgID
	}
	return r
}
