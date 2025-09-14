package offiaccount

// KFInfo 客服账号信息
type KFInfo struct {
	KfAccount        string `json:"kf_account"`         // 完整客服账号，格式为：账号前缀@公众号微信号
	KfNick           string `json:"kf_nick"`            // 客服昵称
	KfID             string `json:"kf_id"`              // 客服编号
	KfHeadImgURL     string `json:"kf_headimgurl"`      // 客服头像
	KfWX             string `json:"kf_wx"`              // 如果客服账号已绑定了客服人员微信号，则此处显示微信号
	InviteWX         string `json:"invite_wx"`          // 如果客服账号尚未绑定微信号，但是已经发起了一个绑定邀请，则此处显示绑定邀请的微信号
	InviteExpireTime string `json:"invite_expire_time"` // 如果客服账号尚未绑定微信号，但是已经发起过一个绑定邀请，邀请的过期时间，为unix 时间戳
	InviteStatus     string `json:"invite_status"`      // 邀请的状态，有等待确认"waiting"，被拒绝"rejected"，过期"expired"
}

// KFOnlineInfo 在线客服信息
type KFOnlineInfo struct {
	KfAccount    string `json:"kf_account"`    // 完整客服账号，格式为：账号前缀@公众号微信号
	Status       int    `json:"status"`        // 客服在线状态，目前为：0-不在线，1-web 在线
	KfID         string `json:"kf_id"`         // 客服编号
	AcceptedCase int    `json:"accepted_case"` // 客服当前正在接待的会话数
}

// KFListResp 客服列表响应
type KFListResp struct {
	Resp
	KFList []*KFInfo `json:"kf_list"` // 客服列表
}

// KFOnlineListResp 在线客服列表响应
type KFOnlineListResp struct {
	Resp
	KFOnlineList []*KFOnlineInfo `json:"kf_online_list"` // 在线客服列表
}

// KFSessionInfo 客服会话信息
type KFSessionInfo struct {
	CreateTime int64  `json:"createtime"` // 会话接入时间
	OpenID     string `json:"openid"`     // 用户openid
}

// KFCustomerSessionInfo 客户会话状态信息
type KFCustomerSessionInfo struct {
	Resp
	CreateTime int64  `json:"createtime"` // 会话接入时间
	KfAccount  string `json:"kf_account"` // 接待客服账号
}

// WaitCaseInfo 未接入会话信息
type WaitCaseInfo struct {
	LatestTime int64  `json:"latest_time"` // 用户的最后一条消息的时间
	OpenID     string `json:"openid"`      // 用户openid
}

// WaitCaseListResp 未接入会话列表响应
type WaitCaseListResp struct {
	Resp
	Count        int             `json:"count"`        // 未接入会话数量
	WaitCaseList []*WaitCaseInfo `json:"waitcaselist"` // 未接入会话列表
}

// MsgRecord 消息记录
type MsgRecord struct {
	OpenID   string `json:"openid"`   // 用户openid
	OperCode int    `json:"opercode"` // 操作码，2002（客服发送信息），2003（客服接收消息）
	Text     string `json:"text"`     // 聊天记录
	Time     int64  `json:"time"`     // 操作时间
	Worker   string `json:"worker"`   // 完整客服账号，格式为：账号前缀@公众号微信号
}

// MsgListResp 消息列表响应
type MsgListResp struct {
	Resp
	RecordList []*MsgRecord `json:"recordlist"` // 消息内容
	Number     int          `json:"number"`     // 消息数量
	MsgID      int64        `json:"msgid"`      // 消息id
}

// TypingCommand 客服输入状态命令
type TypingCommand string

const (
	TypingCommandTyping       TypingCommand = "Typing"       // 对用户下发"正在输入"状态
	TypingCommandCancelTyping TypingCommand = "CancelTyping" // 取消对用户的"正在输入"状态
)

// KFMsgType 客服消息类型
type KFMsgType string

const (
	KFMsgTypeText          KFMsgType = "text"            // 文本消息
	KFMsgTypeImage         KFMsgType = "image"           // 图片消息
	KFMsgTypeVoice         KFMsgType = "voice"           // 语音消息
	KFMsgTypeVideo         KFMsgType = "video"           // 视频消息
	KFMsgTypeMusic         KFMsgType = "music"           // 音乐消息
	KFMsgTypeNews          KFMsgType = "news"            // 图文消息（点击跳转到外链）
	KFMsgTypeMpNews        KFMsgType = "mpnews"          // 图文消息（点击跳转到图文消息页面）
	KFMsgTypeMpNewsArticle KFMsgType = "mpnewsarticle"   // 图文消息（点击跳转到图文消息页面）
	KFMsgTypeMsgMenu       KFMsgType = "msgmenu"         // 菜单消息
	KFMsgTypeWxCard        KFMsgType = "wxcard"          // 卡券信息
	KFMsgTypeMiniProgram   KFMsgType = "miniprogrampage" // 小程序消息
)

// KFText 文本消息内容
type KFText struct {
	Content string `json:"content"` // 文本内容
}

// KFImage 图片消息内容
type KFImage struct {
	MediaID string `json:"media_id"` // 媒体ID
}

// KFVoice 语音消息内容
type KFVoice struct {
	MediaID string `json:"media_id"` // 媒体ID
}

// KFVideo 视频消息内容
type KFVideo struct {
	MediaID      string `json:"media_id"`              // 媒体ID
	ThumbMediaID string `json:"thumb_media_id"`        // 缩略图媒体ID
	Title        string `json:"title,omitempty"`       // 视频标题
	Description  string `json:"description,omitempty"` // 视频描述
}

// KFMusic 音乐消息内容
type KFMusic struct {
	Title        string `json:"title"`                // 音乐标题
	Description  string `json:"description"`          // 音乐描述
	MusicURL     string `json:"musicurl"`             // 音乐链接
	HQMusicURL   string `json:"hqmusicurl,omitempty"` // 高质量音乐链接
	ThumbMediaID string `json:"thumb_media_id"`       // 缩略图媒体ID
}

// KFNewsArticle 图文消息条目
type KFNewsArticle struct {
	Title       string `json:"title"`       // 消息标题
	Description string `json:"description"` // 消息描述
	PicURL      string `json:"picurl"`      // 封面图片url
	URL         string `json:"url"`         // 跳转url
}

// KFNews 图文消息内容（点击跳转到外链）
type KFNews struct {
	Articles []*KFNewsArticle `json:"articles"` // 图文消息条数限制在1条以内
}

// KFMsgMenu 菜单消息内容
type KFMsgMenu struct {
	HeadContent string        `json:"head_content,omitempty"` // 菜单描述
	List        []*KFMenuList `json:"list"`                   // 菜单内容
	TailContent string        `json:"tail_content,omitempty"` // 菜单结尾
}

// KFMenuList 菜单列表项
type KFMenuList struct {
	ID      string `json:"id"`      // 菜单值
	Content string `json:"content"` // 菜单项
}

// KFWxCard 卡券消息内容
type KFWxCard struct {
	CardID string `json:"card_id"` // 卡券ID
}

// KFMiniProgramPage 小程序消息内容
type KFMiniProgramPage struct {
	Title        string `json:"title"`          // 小程序卡片标题
	AppID        string `json:"appid"`          // 小程序APPID
	PagePath     string `json:"pagepath"`       // 小程序的页面路径
	ThumbMediaID string `json:"thumb_media_id"` // 小程序消息卡片的封面
}

// KFCustomService 客服信息
type KFCustomService struct {
	KfAccount string `json:"kf_account"` // 客服账号
}

// KFMessage 客服消息
type KFMessage struct {
	ToUser          string             `json:"touser"`                    // 用户的 OpenID
	MsgType         KFMsgType          `json:"msgtype"`                   // 消息类型
	Text            *KFText            `json:"text,omitempty"`            // 文本消息
	Image           *KFImage           `json:"image,omitempty"`           // 图片消息
	Voice           *KFVoice           `json:"voice,omitempty"`           // 语音消息
	Video           *KFVideo           `json:"video,omitempty"`           // 视频消息
	Music           *KFMusic           `json:"music,omitempty"`           // 音乐消息
	News            *KFNews            `json:"news,omitempty"`            // 图文消息（点击跳转到外链）
	MpNews          *KFImage           `json:"mpnews,omitempty"`          // 图文消息（点击跳转到图文消息页面）
	MpNewsArticle   *KFMPNewsArticle   `json:"mpnewsarticle,omitempty"`   // 图文消息（点击跳转到图文消息页面）
	MsgMenu         *KFMsgMenu         `json:"msgmenu,omitempty"`         // 菜单消息
	WxCard          *KFWxCard          `json:"wxcard,omitempty"`          // 卡券信息
	MiniProgramPage *KFMiniProgramPage `json:"miniprogrampage,omitempty"` // 小程序消息
	CustomService   *KFCustomService   `json:"customservice,omitempty"`   // 以某个客服账号来发消息
}

// KFMPNewsArticle 图文消息文章
type KFMPNewsArticle struct {
	ArticleID string `json:"article_id"` // 发布文章ID
}

// KFGetMsgListRequest 获取聊天记录请求参数
type KFGetMsgListRequest struct {
	StartTime int64 `json:"starttime"` // 起始时间
	EndTime   int64 `json:"endtime"`   // 结束时间
	MsgID     int64 `json:"msgid"`     // 消息id
	Number    int   `json:"number"`    // 获取数量
}

// KFTypingRequest 客服输入状态请求参数
type KFTypingRequest struct {
	ToUser  string        `json:"touser"`  // 用户的 OpenID
	Command TypingCommand `json:"command"` // 命令
}

// KFSessionListResp 客服会话列表响应
type KFSessionListResp struct {
	Resp
	SessionList []*KFSessionInfo `json:"sessionlist"` // 会话列表
}
