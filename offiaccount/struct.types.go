package offiaccount

import "fmt"

type Resp struct {
	ErrCode int    `json:"errcode"` // 错误码
	ErrMsg  string `json:"errmsg"`  // 错误描述
}

// DNS DNS解析结果
type DNS struct {
	IP           string `json:"ip"`            // 解析出来的ip
	RealOperator string `json:"real_operator"` // ip对应的运营商
}

// PING PING检测结果
type PING struct {
	IP           string `json:"ip"`            // ping的ip，执行命令为ping ip –c 1-w 1 -q
	FromOperator string `json:"from_operator"` // ping的源头的运营商，由请求中的check_operator控制
	PackageLoss  string `json:"package_loss"`  // ping的丢包率，0%表示无丢包，100%表示全部丢包。因为目前仅发送一个ping包，因此取值仅有0%或者100%两种可能。
}

// CallbackCheckResponse 网络通信检测结果
type CallbackCheckResponse struct {
	DNS  []*DNS  `json:"dns"`  // DNS解析结果列表
	PING []*PING `json:"ping"` // PING检测结果列表
}

type IpList struct {
	Resp
	IpList []string `json:"ip_list"` // ip列表
}

type AccessToken struct {
	AccessToken string `json:"access_token"` // access_token
	ExpiresIn   int64  `json:"expires_in"`   // access_token的过期时间
}

type RidInfoResp struct {
	Resp
	Request struct {
		InvokeTime   int    `json:"invoke_time"`   // 调用时间
		CostInMs     int    `json:"cost_in_ms"`    // 调用耗时
		RequestUrl   string `json:"request_url"`   // 请求的url
		RequestBody  string `json:"request_body"`  // 请求的body
		ResponseBody string `json:"response_body"` // 响应的body
		ClientIp     string `json:"client_ip"`     // 客户端ip
	} `json:"request"`
}
type ApiQuotaResp struct {
	Resp
	Quota struct {
		DailyLimit int `json:"daily_limit"` // 当天调用次数限制
		Used       int `json:"used"`        // 当天调用次数
		Remain     int `json:"remain"`      // 当天剩余调用次数
	} `json:"quota"`
}

// MenuNewsInfo 图文消息信息
type MenuNewsInfo struct {
	Title      string `json:"title"`       // 图文消息的标题
	Author     string `json:"author"`      // 作者
	Digest     string `json:"digest"`      // 摘要
	ShowCover  int    `json:"show_cover"`  // 是否显示封面，0为不显示，1为显示
	CoverURL   string `json:"cover_url"`   // 封面图片的URL
	ContentURL string `json:"content_url"` // 正文的URL
	SourceURL  string `json:"source_url"`  // 原文的URL，若置空则无查看原文入口
}

// MenuButton 菜单按钮结构体（通用）
type MenuButton struct {
	// 菜单的响应动作类型
	// click: 点击推事件
	// view: 跳转URL
	// miniprogram: 跳转小程序
	// scancode_push: 扫码推事件
	// scancode_waitmsg: 扫码推事件且弹出"消息接收中"提示框
	// pic_sysphoto: 弹出系统拍照发图
	// pic_photo_or_album: 弹出拍照或者相册发图
	// pic_weixin: 弹出微信相册发图器
	// location_select: 弹出地理位置选择器
	// media_id: 下发消息(除文本消息)
	// view_limited: 跳转图文消息URL
	// article_id: 发布后的图文消息
	// article_view_limited: 跳转发布后的图文消息URL
	Type string `json:"type,omitempty"`

	// 菜单标题，不超过16个字节，子菜单不超过60个字节
	Name string `json:"name"`

	// 菜单KEY值，用于消息接口推送，不超过128字节
	// click等点击类型必须
	Key string `json:"key,omitempty"`

	// 网页链接，用户点击菜单可打开链接，不超过1024字节
	// view、miniprogram类型必须
	// type为miniprogram时，不支持小程序的老版本客户端将打开本url
	URL string `json:"url,omitempty"`

	// 调用新增永久素材接口返回的合法media_id
	// media_id类型和view_limited类型必须
	MediaID string `json:"media_id,omitempty"`

	// 小程序的appid（仅认证公众号可配置）
	// miniprogram类型必须
	AppID string `json:"appid,omitempty"`

	// 小程序的页面路径
	// miniprogram类型必须
	PagePath string `json:"pagepath,omitempty"`

	// 发布后获得的合法 ArticleId
	// article_id类型和article_view_limited类型必须
	ArticleID string `json:"article_id,omitempty"`

	// 按钮的值，根据type不同而不同
	Value string `json:"value,omitempty"`

	// 图文消息的信息
	NewsInfo []*MenuNewsInfo `json:"news_info,omitempty"`

	// 二级菜单信息（查询时的特殊结构）
	SubButton *SubButtonWrapper `json:"sub_button,omitempty"`
}

// SubButtonWrapper 二级菜单包装器（查询时使用）
type SubButtonWrapper struct {
	List []*MenuButton `json:"list,omitempty"` // 二级菜单数组
}

// CreateMenuButton 创建自定义菜单按钮结构体
type CreateMenuButton struct {
	// 菜单的响应动作类型
	// click: 点击推事件
	// view: 跳转URL
	// miniprogram: 跳转小程序
	// scancode_push: 扫码推事件
	// scancode_waitmsg: 扫码推事件且弹出"消息接收中"提示框
	// pic_sysphoto: 弹出系统拍照发图
	// pic_photo_or_album: 弹出拍照或者相册发图
	// pic_weixin: 弹出微信相册发图器
	// location_select: 弹出地理位置选择器
	// media_id: 下发消息(除文本消息)
	// view_limited: 跳转图文消息URL
	// article_id: 发布后的图文消息
	// article_view_limited: 跳转发布后的图文消息URL
	Type string `json:"type,omitempty"`

	// 菜单标题，不超过16个字节，子菜单不超过60个字节
	Name string `json:"name"`

	// 菜单KEY值，用于消息接口推送，不超过128字节
	// click等点击类型必须
	Key string `json:"key,omitempty"`

	// 网页链接，用户点击菜单可打开链接，不超过1024字节
	// view、miniprogram类型必须
	// type为miniprogram时，不支持小程序的老版本客户端将打开本url
	URL string `json:"url,omitempty"`

	// 调用新增永久素材接口返回的合法media_id
	// media_id类型和view_limited类型必须
	MediaID string `json:"media_id,omitempty"`

	// 小程序的appid（仅认证公众号可配置）
	// miniprogram类型必须
	AppID string `json:"appid,omitempty"`

	// 小程序的页面路径
	// miniprogram类型必须
	PagePath string `json:"pagepath,omitempty"`

	// 发布后获得的合法 ArticleId
	// article_id类型和article_view_limited类型必须
	ArticleID string `json:"article_id,omitempty"`

	// 二级菜单数组，个数应为1~5个
	SubButton []*CreateMenuButton `json:"sub_button,omitempty"`
}

// Menu 默认菜单信息
type Menu struct {
	// 一级菜单数组(1-3个)
	Buttons []*MenuButton `json:"button"`
}

// MatchRule 菜单匹配规则
type MatchRule struct {
	// 用户标签的id，可通过用户标签管理接口获取
	TagID string `json:"tag_id,omitempty"`

	// 性别：男（1）女（2），不填则不做匹配
	Sex string `json:"sex,omitempty"`

	// 国家信息，是用户在微信中设置的国家，如中国
	Country string `json:"country,omitempty"`

	// 省份信息，是用户在微信中设置的省份，如广东
	Province string `json:"province,omitempty"`

	// 城市信息，是用户在微信中设置的城市，如广州
	City string `json:"city,omitempty"`

	// 客户端版本，当前只具体到系统型号：
	// IOS(1), Android(2), Others(3)
	// 不填则不做匹配
	ClientPlatformType string `json:"client_platform_type,omitempty"`

	// 语言信息，是用户在微信中设置的语言，如 zh_CN
	Language string `json:"language,omitempty"`
}

// ConditionalMenu 个性化菜单
type ConditionalMenu struct {
	// 一级菜单数组(1-3个)
	Buttons []*CreateMenuButton `json:"button"`

	// 菜单匹配规则(至少一个非空字段)
	MatchRule *MatchRule `json:"matchrule"`

	// 菜单ID（创建时不需要，查询时会返回）
	MenuID string `json:"menuid,omitempty"`
}

// GetMenuResponse 获取自定义菜单配置的响应结构体
type GetMenuResponse struct {
	Resp
	// 默认菜单信息
	Menu *Menu `json:"menu,omitempty"`

	// 个性化菜单列表
	ConditionalMenu []*ConditionalMenu `json:"conditionalmenu,omitempty"`
}

// CreateCustomMenu 创建自定义菜单结构体
type CreateCustomMenu struct {
	Buttons []*CreateMenuButton `json:"button"` // 菜单按钮列表
}

// QueryCustomMenu 查询自定义菜单结构体
type QueryCustomMenu struct {
	Resp
	Buttons []*MenuButton `json:"button"` // 菜单按钮列表
}

// SelfMenu 当前自定义菜单配置信息（查询结果）
type SelfMenu struct {
	// 菜单是否开启，0:未开启，1:开启
	IsMenuOpen int `json:"is_menu_open"`
	// 自定义菜单信息
	SelfMenuInfo QueryCustomMenu `json:"selfmenu_info"`
}

// AddConditionalMenuResponse 创建个性化菜单的响应结构体
type AddConditionalMenuResponse struct {
	Resp
	// 创建成功的个性化菜单ID
	MenuID string `json:"menuid,omitempty"`
}

// UploadImageResponse 上传图片响应结构
type UploadImageResponse struct {
	Url     string `json:"url"`
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type DeleteMassMsgRequest struct {
	// 发送出去的消息ID
	MsgId int `json:"msg_id,omitempty"`
	// 要删除的文章在图文消息中的位置，第一篇编号为1，该字段不填或填0会删除全部文章
	ArticleIdx int `json:"article_idx"`
	// 要删除的文章url，当msg_id未指定时该参数才生效
	Url string `json:"url,omitempty"`
}

type SpeedResp struct {
	Resp
	// 群发速度的级别
	Speed int `json:"speed"`
	// 群发速度的真实值 单位：万/分钟
	RealSpeed int `json:"realspeed"`
}

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

// 验证方法

// Validate 验证请求参数
func (r *MassSendByTagRequest) Validate() error {
	if r.Filter == nil {
		return fmt.Errorf("filter不能为空")
	}

	if r.Filter.IsToAll == "" {
		return fmt.Errorf("is_to_all不能为空")
	}

	if r.Filter.IsToAll != "true" && r.Filter.IsToAll != "false" {
		return fmt.Errorf("is_to_all必须为true或false")
	}

	if r.Filter.IsToAll == "false" && r.Filter.TagID == "" {
		return fmt.Errorf("当is_to_all为false时，tag_id不能为空")
	}

	if r.MsgType == "" {
		return fmt.Errorf("msgtype不能为空")
	}

	// 验证消息内容
	switch r.MsgType {
	case MsgTypeMPNews:
		if r.MPNews == nil || r.MPNews.MediaID == "" {
			return fmt.Errorf("图文消息的media_id不能为空")
		}
	case MsgTypeText:
		if r.Text == nil || r.Text.Content == "" {
			return fmt.Errorf("文本消息的content不能为空")
		}
	case MsgTypeVoice:
		if r.Voice == nil || r.Voice.MediaID == "" {
			return fmt.Errorf("语音消息的media_id不能为空")
		}
	case MsgTypeImage:
		if r.Images == nil || len(r.Images.MediaIDs) == 0 {
			return fmt.Errorf("图片消息的media_ids不能为空")
		}
	case MsgTypeVideo:
		if r.MPVideo == nil || r.MPVideo.MediaID == "" {
			return fmt.Errorf("视频消息的media_id不能为空")
		}
	case MsgTypeWxCard:
		if r.WxCard == nil || r.WxCard.CardID == "" {
			return fmt.Errorf("卡券消息的card_id不能为空")
		}
	}

	return nil
}

// AddNewsMaterialRequest 新增临时图文素材请求结构体
type AddNewsMaterialRequest struct {
	Articles []Article `json:"articles"` // 图文消息，一个图文消息支持1到8条图文
}

// Article 图文消息单条内容结构体
type Article struct {
	Title              string `json:"title"`                           // 图文消息的标题（必填）
	Author             string `json:"author,omitempty"`                // 图文消息的作者
	ThumbMediaID       string `json:"thumb_media_id"`                  // 图文消息缩略图的media_id（必填）
	Content            string `json:"content"`                         // 图文消息页面的内容，支持HTML标签（必填）
	ContentSourceURL   string `json:"content_source_url,omitempty"`    // 点击"阅读原文"后的页面链接
	Digest             string `json:"digest,omitempty"`                // 图文消息的描述，为空时默认抓取正文前64个字
	ShowCoverPic       int    `json:"show_cover_pic,omitempty"`        // 是否显示封面：1显示，0不显示
	NeedOpenComment    int    `json:"need_open_comment,omitempty"`     // 是否打开评论：0不打开，1打开
	OnlyFansCanComment int    `json:"only_fans_can_comment,omitempty"` // 是否粉丝才可评论：0所有人可评论，1粉丝才可评论
}

// AddNewsMaterialResponse 新增临时图文素材响应结构体
type AddNewsMaterialResponse struct {
	ErrCode   int    `json:"errcode"`    // 错误码
	ErrMsg    string `json:"errmsg"`     // 错误信息
	Type      string `json:"type"`       // 媒体文件类型，分别有图片（image）、语音（voice）、视频（video）和缩略图（thumb），图文消息为news
	MediaID   string `json:"media_id"`   // 媒体文件上传后，获取标识
	CreatedAt int64  `json:"created_at"` // 媒体文件上传时间戳
}

// 常量定义
const (
	// 显示封面设置
	ShowCoverPicNo  = 0 // 不显示封面
	ShowCoverPicYes = 1 // 显示封面

	// 评论设置
	CommentClosed = 0 // 不打开评论
	CommentOpen   = 1 // 打开评论

	// 评论权限设置
	CommentAllUsers = 0 // 所有人可评论
	CommentFansOnly = 1 // 粉丝才可评论

	// 图文消息数量限制
	MaxArticleCount = 8 // 最多8条图文
	MinArticleCount = 1 // 最少1条图文
)

// 构造函数和辅助方法

// NewAddNewsMaterialRequest 创建新增临时图文素材请求
func NewAddNewsMaterialRequest() *AddNewsMaterialRequest {
	return &AddNewsMaterialRequest{
		Articles: make([]Article, 0),
	}
}

// AddArticle 添加图文消息
func (r *AddNewsMaterialRequest) AddArticle(article Article) *AddNewsMaterialRequest {
	if len(r.Articles) < MaxArticleCount {
		r.Articles = append(r.Articles, article)
	}
	return r
}

// NewArticle 创建图文消息
func NewArticle(title, thumbMediaID, content string) Article {
	return Article{
		Title:        title,
		ThumbMediaID: thumbMediaID,
		Content:      content,
		ShowCoverPic: ShowCoverPicYes, // 默认显示封面
	}
}

// SetAuthor 设置作者
func (a *Article) SetAuthor(author string) *Article {
	a.Author = author
	return a
}

// SetContentSourceURL 设置阅读原文链接
func (a *Article) SetContentSourceURL(url string) *Article {
	a.ContentSourceURL = url
	return a
}

// SetDigest 设置摘要
func (a *Article) SetDigest(digest string) *Article {
	a.Digest = digest
	return a
}

// SetShowCoverPic 设置是否显示封面
func (a *Article) SetShowCoverPic(show bool) *Article {
	if show {
		a.ShowCoverPic = ShowCoverPicYes
	} else {
		a.ShowCoverPic = ShowCoverPicNo
	}
	return a
}

// SetComment 设置评论功能
func (a *Article) SetComment(needOpen bool, onlyFans bool) *Article {
	if needOpen {
		a.NeedOpenComment = CommentOpen
		if onlyFans {
			a.OnlyFansCanComment = CommentFansOnly
		} else {
			a.OnlyFansCanComment = CommentAllUsers
		}
	} else {
		a.NeedOpenComment = CommentClosed
		a.OnlyFansCanComment = CommentAllUsers // 不开启评论时，这个字段无意义
	}
	return a
}

// EnableComment 开启评论（所有人可评论）
func (a *Article) EnableComment() *Article {
	return a.SetComment(true, false)
}

// EnableCommentForFansOnly 开启评论（仅粉丝可评论）
func (a *Article) EnableCommentForFansOnly() *Article {
	return a.SetComment(true, true)
}

// DisableComment 关闭评论
func (a *Article) DisableComment() *Article {
	return a.SetComment(false, false)
}

// 验证方法

// Validate 验证请求参数
func (r *AddNewsMaterialRequest) Validate() error {
	if len(r.Articles) < MinArticleCount {
		return fmt.Errorf("图文消息数量不能少于%d条", MinArticleCount)
	}

	if len(r.Articles) > MaxArticleCount {
		return fmt.Errorf("图文消息数量不能超过%d条", MaxArticleCount)
	}

	for i, article := range r.Articles {
		if err := article.Validate(); err != nil {
			return fmt.Errorf("第%d条图文消息验证失败: %v", i+1, err)
		}
	}

	return nil
}

// Validate 验证单条图文消息参数
func (a *Article) Validate() error {
	if a.Title == "" {
		return fmt.Errorf("标题不能为空")
	}

	if a.ThumbMediaID == "" {
		return fmt.Errorf("缩略图media_id不能为空")
	}

	if a.Content == "" {
		return fmt.Errorf("内容不能为空")
	}

	// 验证评论设置的逻辑性
	if a.NeedOpenComment == CommentClosed && a.OnlyFansCanComment == CommentFansOnly {
		// 虽然不是错误，但逻辑上不合理，可以给出警告
		// 这里选择不报错，只是标准化设置
		a.OnlyFansCanComment = CommentAllUsers
	}

	return nil
}

// 辅助方法

// GetArticleCount 获取图文消息数量
func (r *AddNewsMaterialRequest) GetArticleCount() int {
	return len(r.Articles)
}

// IsEmpty 检查是否为空
func (r *AddNewsMaterialRequest) IsEmpty() bool {
	return len(r.Articles) == 0
}

// IsFull 检查是否已满
func (r *AddNewsMaterialRequest) IsFull() bool {
	return len(r.Articles) >= MaxArticleCount
}

// Clear 清空所有图文消息
func (r *AddNewsMaterialRequest) Clear() *AddNewsMaterialRequest {
	r.Articles = make([]Article, 0)
	return r
}

// RemoveArticle 移除指定位置的图文消息
func (r *AddNewsMaterialRequest) RemoveArticle(index int) *AddNewsMaterialRequest {
	if index >= 0 && index < len(r.Articles) {
		r.Articles = append(r.Articles[:index], r.Articles[index+1:]...)
	}
	return r
}

// NewSingleArticleRequest 创建单条图文消息请求
func NewSingleArticleRequest(title, thumbMediaID, content string) *AddNewsMaterialRequest {
	article := NewArticle(title, thumbMediaID, content)
	return NewAddNewsMaterialRequest().AddArticle(article)
}

// NewMultiArticleRequest 创建多条图文消息请求
func NewMultiArticleRequest(articles ...Article) *AddNewsMaterialRequest {
	req := NewAddNewsMaterialRequest()
	for _, article := range articles {
		if req.IsFull() {
			break
		}
		req.AddArticle(article)
	}
	return req
}

// SubscribeMessageRequest 订阅消息请求结构体
type SubscribeMessageRequest struct {
	ToUser      string                 `json:"touser"`                  // 接收者（用户）的 openid
	TemplateID  string                 `json:"template_id"`             // 所需下发的订阅模板id
	URL         string                 `json:"url,omitempty"`           // 模板跳转链接（可选）
	MiniProgram *MiniProgram           `json:"miniprogram,omitempty"`   // 跳转小程序时填写（可选）
	Data        map[string]interface{} `json:"data"`                    // 模板内容
	ClientMsgID string                 `json:"client_msg_id,omitempty"` // 防重入id（可选）
}

// MiniProgram 小程序跳转信息
type MiniProgram struct {
	AppID    string `json:"appid,omitempty"`    // 小程序appid
	PagePath string `json:"pagepath,omitempty"` // 小程序跳转路径
}

// TemplateData 模板数据基础结构
type TemplateData struct {
	Value string `json:"value"`
}

// 以下是各种数据类型的具体结构体，用于构建 data 字段

// ThingData 事物类型数据 - 20个以内字符
type ThingData struct {
	Value string `json:"value"` // 可汉字、数字、字母或符号组合，20个以内字符
}

// CharacterStringData 字符串类型数据 - 32位以内
type CharacterStringData struct {
	Value string `json:"value"` // 可数字、字母或符号组合，32位以内
}

// TimeData 时间类型数据
type TimeData struct {
	Value string `json:"value"` // 24小时制时间格式，支持时间段
}

// AmountData 金额类型数据
type AmountData struct {
	Value string `json:"value"` // 1个币种符号+10位以内纯数字，可带小数
}

// PhoneNumberData 电话号码类型数据
type PhoneNumberData struct {
	Value string `json:"value"` // 17位以内，数字、符号
}

// CarNumberData 车牌号码类型数据
type CarNumberData struct {
	Value string `json:"value"` // 8位以内，第一位与最后一位可为汉字
}

// ConstData 常量类型数据
type ConstData struct {
	Value string `json:"value"` // 20位以内字符，需要枚举审核
}

type AddTemplateResponse struct {
	Resp
	// 模板ID
	TemplateId string `json:"template_id"`
}

type QueryBlockTmplMsgReq struct {
	// 被拦截的模板消息id
	TmplMsgId string `json:"tmpl_msg_id"`
	// 上一页查询结果最大的id，用于翻页，第一次传0
	LargestId int `json:"largest_id"`
	// 单页查询的大小，最大100
	Limit int `json:"limit"`
}
type MsgInfo struct {
	// 记录唯一ID，用于多次查询的largest_id
	Id int `json:"id"`
	// 被拦截的模板消息id
	TmplMsgId string `json:"tmpl_msg_id"`
	// 模板消息的标题
	Title string `json:"title"`
	// 模板消息的内容
	Content string `json:"content"`
	// 下发的时间戳
	SendTimestamp int `json:"send_timestamp"`
	// 下发目标用户的openid
	Openid string `json:"openid"`
}

type QueryBlockTmplMsgResp struct {
	Resp
	MsgInfo *MsgInfo `json:"msginfo"`
}

type Template struct {
	// 模板ID
	TemplateId string `json:"template_id"`
	// 模板标题
	Title string `json:"title"`
	// 模板所属行业的一级行业
	PrimaryIndustry string `json:"primary_industry"`
	// 模板所属行业的二级行业
	DeputyIndustry string `json:"deputy_industry"`
	// 模板内容
	Content string `json:"content"`
	// 模板示例
	Example string `json:"example"`
}

type TemplateList struct {
	Resp
	TemplateList []*Template `json:"template_list"`
}

// PrimaryIndustry 账号设置的主营行业
type PrimaryIndustry struct {
	// 一级类目
	FirstClass string `json:"first_class"`
	// 二级类目
	SecondClass string `json:"second_class"`
}

// SecondaryIndustry 账号设置的副营行业
type SecondaryIndustry struct {
	// 一级类目
	FirstClass string `json:"first_class"`
	// 二级类目
	SecondClass string `json:"second_class"`
}

type IndustryResp struct {
	// 账号设置的主营行业
	PrimaryIndustry *PrimaryIndustry `json:"primary_industry"`
	// 账号设置的副营行业
	SecondaryIndustry *SecondaryIndustry `json:"secondary_industry"`
}
