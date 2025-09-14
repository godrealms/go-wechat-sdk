package offiaccount

// Resp 基础响应结构体
type Resp struct {
	ErrCode int    `json:"errcode"` // 错误码
	ErrMsg  string `json:"errmsg"`  // 错误描述
}

// Tag 标签信息
type Tag struct {
	Id    int64  `json:"id"`    // 标签ID
	Name  string `json:"name"`  // 标签名
	Count int64  `json:"count"` // 标签下粉丝数
}

// GetTagsResult 获取标签列表结果
type GetTagsResult struct {
	Resp
	Tags []*Tag `json:"tags"` // 标签列表
}

// CreateTagRequest 创建标签请求参数
type CreateTagRequest struct {
	Tag *CreateTag `json:"tag"` // 标签信息
}

// CreateTag 创建标签信息
type CreateTag struct {
	Name string `json:"name"` // 标签名（30个字符以内）
}

// CreateTagResult 创建标签结果
type CreateTagResult struct {
	Resp
	Tag *Tag `json:"tag"` // 标签信息
}

// UpdateTagRequest 更新标签请求参数
type UpdateTagRequest struct {
	Tag *UpdateTag `json:"tag"` // 标签信息
}

// UpdateTag 更新标签信息
type UpdateTag struct {
	Id   int64  `json:"id"`   // 标签ID
	Name string `json:"name"` // 标签名
}

// DeleteTagRequest 删除标签请求参数
type DeleteTagRequest struct {
	Tag *DeleteTag `json:"tag"` // 标签信息
}

// DeleteTag 删除标签信息
type DeleteTag struct {
	Id int64 `json:"id"` // 标签ID
}

// GetTagFansRequest 获取标签下粉丝列表请求参数
type GetTagFansRequest struct {
	TagId      int64  `json:"tagid"`       // 标签ID
	NextOpenid string `json:"next_openid"` // 第一个拉取的OPENID，不填默认从头开始拉取
}

// GetTagFansResult 获取标签下粉丝列表结果
type GetTagFansResult struct {
	Resp
	Count      int64       `json:"count"`       // 本次获取的粉丝数量
	Data       *OpenidData `json:"data"`        // 标签下粉丝数据
	NextOpenid string      `json:"next_openid"` // 拉取列表最后一个用户的openid
}

// OpenidData 粉丝openid数据
type OpenidData struct {
	Openid []string `json:"openid"` // 粉丝openid列表
}

// BatchTaggingRequest 批量为用户打标签请求参数
type BatchTaggingRequest struct {
	OpenidList []string `json:"openid_list"` // 粉丝openid列表，最多50个
	Tagid      int64    `json:"tagid"`       // 标签id
}

// BatchUntaggingRequest 批量为用户取消标签请求参数
type BatchUntaggingRequest struct {
	OpenidList []string `json:"openid_list"` // 粉丝openid列表
	Tagid      int64    `json:"tagid"`       // 标签id
}

// GetTagidListRequest 获取用户身上的标签列表请求参数
type GetTagidListRequest struct {
	Openid string `json:"openid"` // 用户openid
}

// GetTagidListResult 获取用户身上的标签列表结果
type GetTagidListResult struct {
	Resp
	TagidList []int64 `json:"tagid_list"` // 用户的标签列表
}

// UserInfo 用户信息
type UserInfo struct {
	Subscribe      int     `json:"subscribe"`       // 用户是否订阅该公众号标识，值为0时，代表此用户没有关注该公众号，拉取不到其余信息。
	Openid         string  `json:"openid"`          // 用户的标识，对当前公众号唯一
	Language       string  `json:"language"`        // 用户的语言，简体中文为zh_CN
	SubscribeTime  int64   `json:"subscribe_time"`  // 用户关注时间，为时间戳。如果用户曾多次关注，则取最后关注时间
	Unionid        string  `json:"unionid"`         // 只有在用户将公众号绑定到微信开放平台账号后，才会出现该字段。
	Remark         string  `json:"remark"`          // 公众号运营者对粉丝的备注，公众号运营者可在微信公众平台用户管理界面对粉丝添加备注
	Groupid        int64   `json:"groupid"`         // 用户所在的分组ID（兼容旧的用户分组接口）
	TagidList      []int64 `json:"tagid_list"`      // 用户被打上的标签ID列表
	SubscribeScene string  `json:"subscribe_scene"` // 返回用户关注的渠道来源
	QrScene        int64   `json:"qr_scene"`        // 二维码扫码场景（开发者自定义）
	QrSceneStr     string  `json:"qr_scene_str"`    // 二维码扫码场景描述（开发者自定义）
}

// GetFansResult 获取用户列表结果
type GetFansResult struct {
	Resp
	Total      int64       `json:"total"`       // 关注该公众账号的总用户数
	Count      int64       `json:"count"`       // 拉取的OPENID个数，最大值为10000
	Data       *OpenidData `json:"data"`        // 列表数据，OPENID的列表
	NextOpenid string      `json:"next_openid"` // 拉取列表后一个用户的OPENID
}

// UserListItem 用户列表项
type UserListItem struct {
	Openid   string `json:"openid"` // 用户的标识，对当前公众号唯一；必须是已关注的用户的 openid
	Language string `json:"lang"`   // 国家地区语言版本，zh_CN 简体，zh_TW 繁体，en 英语，默认为zh-CN
}

// BatchGetUserInfoRequest 批量获取用户基本信息请求参数
type BatchGetUserInfoRequest struct {
	UserList []*UserListItem `json:"user_list"` // 用户列表
}

// BatchGetUserInfoResult 批量获取用户基本信息结果
type BatchGetUserInfoResult struct {
	Resp
	UserInfoList []*UserInfo `json:"user_info_list"` // 用户列表
}

// UpdateRemarkRequest 设置用户备注名请求参数
type UpdateRemarkRequest struct {
	Openid string `json:"openid"` // 用户的标识，对当前公众号唯一
	Remark string `json:"remark"` // 备注名
}

// GetBlacklistRequest 获取公众号的黑名单列表请求参数
type GetBlacklistRequest struct {
	BeginOpenid string `json:"begin_openid"` // 起始OpenID，为空时从开头拉取
}

// GetBlacklistResult 获取公众号的黑名单列表结果
type GetBlacklistResult struct {
	Resp
	Total      int64       `json:"total"`       // 用户总数
	Count      int64       `json:"count"`       // 本次返回的用户数
	Data       *OpenidData `json:"data"`        // 用户数据
	NextOpenid string      `json:"next_openid"` // 本次列表后一位openid
}

// QRCodeJumpRule 二维码跳转规则
type QRCodeJumpRule struct {
	Prefix      string   `json:"prefix"`       // 二维码规则
	Path        string   `json:"path"`         // 小程序功能页面
	State       int      `json:"state"`        // 发布标志位，1 表示未发布，2 表示已发布
	OpenVersion int      `json:"open_version"` // 测试范围
	DebugURL    []string `json:"debug_url"`    // 测试链接
}

// GetQRCodeJumpResult 获取二维码跳转规则结果
type GetQRCodeJumpResult struct {
	Resp
	RuleList           []*QRCodeJumpRule `json:"rule_list"`            // 二维码规则详情列表
	QRCodeJumpOpen     int               `json:"qrcodejump_open"`      // 是否已经打开二维码跳转链接设置
	ListSize           int               `json:"list_size"`            // 二维码规则数量
	QRCodeJumpPubQuota int               `json:"qrcodejump_pub_quota"` // 本月还可发布的次数
	TotalCount         int               `json:"total_count"`          // 二维码规则总数据量，用于分页查询
}

// GetQRCodeJumpRequest 获取二维码跳转规则请求参数
type GetQRCodeJumpRequest struct {
	AppID      string   `json:"appid"`       // 小程序的appid
	GetType    int      `json:"get_type"`    // 默认值为0。 0：查询最近新增 10000 条；1：prefix查询；2：分页查询，按新增顺序返回
	PrefixList []string `json:"prefix_list"` // prefix查询，get_type=1 必传，最多传 200 个前缀
	PageNum    int      `json:"page_num"`    // 页码，get_type=2 必传，从 1 开始
	PageSize   int      `json:"page_size"`   // 每页数量，get_type=2 必传，最大为 200
}

// AddQRCodeJumpRequest 增加二维码规则请求参数
type AddQRCodeJumpRequest struct {
	Prefix string `json:"prefix"`  // 二维码规则，填服务号的带参二维码url
	AppID  string `json:"appid"`   // 要扫了服务号二维码之后要跳转的小程序的appid
	Path   string `json:"path"`    // 小程序功能页面
	IsEdit int    `json:"is_edit"` // 编辑标志位，0 表示新增二维码规则，1 表示修改已有二维码规则
}

// PublishQRCodeJumpRequest 发布二维码规则请求参数
type PublishQRCodeJumpRequest struct {
	Prefix string `json:"prefix"` // 二维码规则
}

// DeleteQRCodeJumpRequest 删除二维码规则请求参数
type DeleteQRCodeJumpRequest struct {
	Prefix string `json:"prefix"` // 服务号的带参的二维码url
	AppID  string `json:"appid"`  // 服务号二维码跳转的小程序的appid
}

// GenShortKeyRequest 生成短key请求参数
type GenShortKeyRequest struct {
	LongData      string `json:"long_data"`      // 需要转换的长信息，不超过4KB
	ExpireSeconds int64  `json:"expire_seconds"` // 过期秒数，最大值为2592000（即30天），默认为2592000
}

// GenShortKeyResult 生成短key结果
type GenShortKeyResult struct {
	Resp
	ShortKey string `json:"short_key"` // 短key，15字节，base62编码(0-9/a-z/A-Z)
}

// FetchShortenRequest 还原短key请求参数
type FetchShortenRequest struct {
	ShortKey string `json:"short_key"` // 短key
}

// FetchShortenResult 还原短key结果
type FetchShortenResult struct {
	Resp
	LongData      string `json:"long_data"`      // 长信息
	CreateTime    int64  `json:"create_time"`    // 创建的时间戳
	ExpireSeconds int64  `json:"expire_seconds"` // 剩余的过期秒数
}

// UserSummary 用户增减数据
type UserSummary struct {
	RefDate    string `json:"ref_date"`    // 数据的日期
	UserSource int64  `json:"user_source"` // 用户的渠道
	NewUser    int64  `json:"new_user"`    // 新增的用户数量
	CancelUser int64  `json:"cancel_user"` // 取消关注的用户数量
}

// GetUserSummaryResult 获取用户增减数据结果
type GetUserSummaryResult struct {
	Resp
	List []*UserSummary `json:"list"` // 数据列表
}

// UserCumulate 累计用户数据
type UserCumulate struct {
	RefDate      string `json:"ref_date"`      // 数据的日期
	CumulateUser int64  `json:"cumulate_user"` // 总用户量
}

// GetUserCumulateResult 获取累计用户数据结果
type GetUserCumulateResult struct {
	Resp
	List []*UserCumulate `json:"list"` // 数据列表
}

// GetDataRequest 获取数据请求参数
type GetDataRequest struct {
	BeginDate string `json:"begin_date"` // 起始日期(格式yyyy-MM-dd)
	EndDate   string `json:"end_date"`   // 结束日期(最大跨度7天)
}

// ArticleSummary 图文群发每日数据
type ArticleSummary struct {
	RefDate          string `json:"ref_date"`            // 数据的日期
	MsgID            string `json:"msgid"`               // 图文消息id
	Title            string `json:"title"`               // 图文消息的标题
	IntPageReadUser  int64  `json:"int_page_read_user"`  // 图文页（点击群发图文卡片进入的页面）的阅读人数
	IntPageReadCount int64  `json:"int_page_read_count"` // 图文页的阅读次数
	OriPageReadUser  int64  `json:"ori_page_read_user"`  // 原文页（点击图文页"阅读原文"进入的页面）的阅读人数
	OriPageReadCount int64  `json:"ori_page_read_count"` // 原文页的阅读次数
	ShareUser        int64  `json:"share_user"`          // 分享的人数
	ShareCount       int64  `json:"share_count"`         // 分享的次数
	AddToFavUser     int64  `json:"add_to_fav_user"`     // 收藏的人数
	AddToFavCount    int64  `json:"add_to_fav_count"`    // 收藏的次数
}

// GetArticleSummaryResult 获取图文群发每日数据结果
type GetArticleSummaryResult struct {
	Resp
	List []*ArticleSummary `json:"list"` // 数据列表
}

// UserReadHour 图文阅读分时数据
type UserReadHour struct {
	RefDate          string `json:"ref_date"`            // 数据的日期
	RefHour          int64  `json:"ref_hour"`            // 数据小时
	UserSource       int64  `json:"user_source"`         // 用户从哪里进入来阅读该图文
	IntPageReadUser  int64  `json:"int_page_read_user"`  // 图文页（点击群发图文卡片进入的页面）的阅读人数
	IntPageReadCount int64  `json:"int_page_read_count"` // 图文页的阅读次数
	OriPageReadUser  int64  `json:"ori_page_read_user"`  // 原文页（点击图文页"阅读原文"进入的页面）的阅读人数
	OriPageReadCount int64  `json:"ori_page_read_count"` // 原文页的阅读次数
	ShareUser        int64  `json:"share_user"`          // 分享的人数
	ShareCount       int64  `json:"share_count"`         // 分享的次数
	AddToFavUser     int64  `json:"add_to_fav_user"`     // 收藏的人数
	AddToFavCount    int64  `json:"add_to_fav_count"`    // 收藏的次数
}

// GetUserReadHourResult 获取图文阅读分时数据结果
type GetUserReadHourResult struct {
	Resp
	List []*UserReadHour `json:"list"` // 数据列表
}

// UserShareHour 图文分享转发分时数据
type UserShareHour struct {
	RefDate    string `json:"ref_date"`    // 数据的日期
	RefHour    int64  `json:"ref_hour"`    // 数据小时
	ShareUser  int64  `json:"share_user"`  // 分享的人数
	ShareCount int64  `json:"share_count"` // 分享的次数
	ShareScene int64  `json:"share_scene"` // 分享的场景 1代表好友转发 2代表朋友圈 255代表其他
}

// GetUserShareHourResult 获取图文分享转发分时数据结果
type GetUserShareHourResult struct {
	Resp
	List []*UserShareHour `json:"list"` // 数据列表
}

// UserRead 图文阅读数据
type UserRead struct {
	RefDate          string `json:"ref_date"`            // 数据的日期
	UserSource       int64  `json:"user_source"`         // 用户从哪里进入来阅读该图文
	IntPageReadUser  int64  `json:"int_page_read_user"`  // 图文页（点击群发图文卡片进入的页面）的阅读人数
	IntPageReadCount int64  `json:"int_page_read_count"` // 图文页的阅读次数
	OriPageReadUser  int64  `json:"ori_page_read_user"`  // 原文页（点击图文页"阅读原文"进入的页面）的阅读人数
	OriPageReadCount int64  `json:"ori_page_read_count"` // 原文页的阅读次数
	ShareUser        int64  `json:"share_user"`          // 分享的人数
	ShareCount       int64  `json:"share_count"`         // 分享的次数
	AddToFavUser     int64  `json:"add_to_fav_user"`     // 收藏的人数
	AddToFavCount    int64  `json:"add_to_fav_count"`    // 收藏的次数
}

// GetUserReadResult 获取图文统计数据结果
type GetUserReadResult struct {
	Resp
	List []*UserRead `json:"list"` // 数据列表
}

// ArticleTotalDetail 图文群发总数据详情
type ArticleTotalDetail struct {
	StatDate                     string `json:"stat_date"`                         // 统计的日期
	TargetUser                   int64  `json:"target_user"`                       // 送达人数
	IntPageReadUser              int64  `json:"int_page_read_user"`                // 图文页（点击群发图文卡片进入的页面）的阅读人数
	IntPageReadCount             int64  `json:"int_page_read_count"`               // 图文页的阅读次数
	OriPageReadUser              int64  `json:"ori_page_read_user"`                // 原文页（点击图文页"阅读原文"进入的页面）的阅读人数
	OriPageReadCount             int64  `json:"ori_page_read_count"`               // 原文页的阅读次数
	ShareUser                    int64  `json:"share_user"`                        // 分享的人数
	ShareCount                   int64  `json:"share_count"`                       // 分享的次数
	AddToFavUser                 int64  `json:"add_to_fav_user"`                   // 收藏的人数
	AddToFavCount                int64  `json:"add_to_fav_count"`                  // 收藏的次数
	IntPageFromSessionReadUser   int64  `json:"int_page_from_session_read_user"`   // 公众号会话阅读人数
	IntPageFromSessionReadCount  int64  `json:"int_page_from_session_read_count"`  // 公众号会话阅读次数
	IntPageFromHistMsgReadUser   int64  `json:"int_page_from_hist_msg_read_user"`  // 历史消息页阅读人数
	IntPageFromHistMsgReadCount  int64  `json:"int_page_from_hist_msg_read_count"` // 历史消息页阅读次数
	IntPageFromFeedReadUser      int64  `json:"int_page_from_feed_read_user"`      // 朋友圈阅读人数
	IntPageFromFeedReadCount     int64  `json:"int_page_from_feed_read_count"`     // 朋友圈阅读次数
	IntPageFromFriendsReadUser   int64  `json:"int_page_from_friends_read_user"`   // 好友转发阅读人数
	IntPageFromFriendsReadCount  int64  `json:"int_page_from_friends_read_count"`  // 好友转发阅读次数
	IntPageFromOtherReadUser     int64  `json:"int_page_from_other_read_user"`     // 其他场景阅读人数
	IntPageFromOtherReadCount    int64  `json:"int_page_from_other_read_count"`    // 其他场景阅读次数
	FeedShareFromSessionUser     int64  `json:"feed_share_from_session_user"`      // 公众号会话转发朋友圈人数
	FeedShareFromSessionCnt      int64  `json:"feed_share_from_session_cnt"`       // 公众号会话转发朋友圈次数
	FeedShareFromFeedUser        int64  `json:"feed_share_from_feed_user"`         // 朋友圈转发朋友圈人数
	FeedShareFromFeedCnt         int64  `json:"feed_share_from_feed_cnt"`          // 朋友圈转发朋友圈次数
	FeedShareFromOtherUser       int64  `json:"feed_share_from_other_user"`        // 其他场景转发朋友圈人数
	FeedShareFromOtherCnt        int64  `json:"feed_share_from_other_cnt"`         // 其他场景转发朋友圈次数
	IntPageFromKanyikanReadUser  int64  `json:"int_page_from_kanyikan_read_user"`  // 看一看来源阅读人数
	IntPageFromKanyikanReadCount int64  `json:"int_page_from_kanyikan_read_count"` // 看一看来源阅读次数
	IntPageFromSouyisouReadUser  int64  `json:"int_page_from_souyisou_read_user"`  // 搜一搜来源阅读人数
	IntPageFromSouyisouReadCount int64  `json:"int_page_from_souyisou_read_count"` // 搜一搜来源阅读次数
}

// ArticleTotal 图文群发总数据
type ArticleTotal struct {
	RefDate string                `json:"ref_date"` // 数据的日期
	MsgID   string                `json:"msgid"`    // 图文消息id
	Title   string                `json:"title"`    // 图文消息的标题
	Details []*ArticleTotalDetail `json:"details"`  // 详情列表
}

// GetArticleTotalResult 获取图文群发总数据结果
type GetArticleTotalResult struct {
	Resp
	List []*ArticleTotal `json:"list"` // 数据列表
}

// UserShare 图文分享转发数据
type UserShare struct {
	RefDate    string `json:"ref_date"`    // 数据的日期
	ShareUser  int64  `json:"share_user"`  // 分享的人数
	ShareCount int64  `json:"share_count"` // 分享的次数
	ShareScene int64  `json:"share_scene"` // 分享的场景 1代表好友转发 2代表朋友圈 255代表其他
}

// GetUserShareResult 获取图文分享转发数据结果
type GetUserShareResult struct {
	Resp
	List []*UserShare `json:"list"` // 数据列表
}

// UpstreamMsg 消息发送概况数据
type UpstreamMsg struct {
	RefDate  string `json:"ref_date"`  // 数据的日期
	MsgType  int64  `json:"msg_type"`  // 消息类型，1代表文字 2代表图片 3代表语音 4代表视频 6代表第三方应用消息
	MsgUser  int64  `json:"msg_user"`  // 上行发送了消息的用户数
	MsgCount int64  `json:"msg_count"` // 上行发送了消息的消息总数
}

// GetUpstreamMsgResult 获取消息发送概况数据结果
type GetUpstreamMsgResult struct {
	Resp
	List []*UpstreamMsg `json:"list"` // 数据列表
}

// UpstreamMsgHour 消息发送分时数据
type UpstreamMsgHour struct {
	RefDate  string `json:"ref_date"`  // 数据的日期
	RefHour  int64  `json:"ref_hour"`  // 数据的小时
	MsgType  int64  `json:"msg_type"`  // 消息类型，1代表文字 2代表图片 3代表语音 4代表视频 6代表第三方应用消息
	MsgUser  int64  `json:"msg_user"`  // 上行发送了消息的用户数
	MsgCount int64  `json:"msg_count"` // 上行发送了消息的消息总数
}

// GetUpstreamMsgHourResult 获取消息发送分时数据结果
type GetUpstreamMsgHourResult struct {
	Resp
	List []*UpstreamMsgHour `json:"list"` // 数据列表
}

// UpstreamMsgDist 消息发送分布数据
type UpstreamMsgDist struct {
	RefDate       string `json:"ref_date"`       // 数据的日期
	CountInterval int64  `json:"count_interval"` // 当日发送消息量分布的区间，0代表 "0"，1代表"1-5"，2代表"6-10"，3代表"10次以上"
	MsgUser       int64  `json:"msg_user"`       // 上行发送了消息的用户数
}

// GetUpstreamMsgDistResult 获取消息发送分布数据结果
type GetUpstreamMsgDistResult struct {
	Resp
	List []*UpstreamMsgDist `json:"list"` // 数据列表
}

// InterfaceSummary 接口分析数据
type InterfaceSummary struct {
	RefDate       string `json:"ref_date"`        // 数据的日期
	CallbackCount int64  `json:"callback_count"`  // 通过服务器配置地址获得消息后，被动回复用户消息的次数
	FailCount     int64  `json:"fail_count"`      // 上述动作的失败次数
	TotalTimeCost int64  `json:"total_time_cost"` // 总耗时，除以callback_count即为平均耗时
	MaxTimeCost   int64  `json:"max_time_cost"`   // 最大耗时
}

// GetInterfaceSummaryResult 获取接口分析数据结果
type GetInterfaceSummaryResult struct {
	Resp
	List []*InterfaceSummary `json:"list"` // 数据列表
}

// InterfaceSummaryHour 接口分析分时数据
type InterfaceSummaryHour struct {
	RefDate       string `json:"ref_date"`        // 数据的日期
	RefHour       int64  `json:"ref_hour"`        // 数据的小时
	CallbackCount int64  `json:"callback_count"`  // 通过服务器配置地址获得消息后，被动回复用户消息的次数
	FailCount     int64  `json:"fail_count"`      // 上述动作的失败次数
	TotalTimeCost int64  `json:"total_time_cost"` // 总耗时，除以callback_count即为平均耗时
	MaxTimeCost   int64  `json:"max_time_cost"`   // 最大耗时
}

// GetInterfaceSummaryHourResult 获取接口分析分时数据结果
type GetInterfaceSummaryHourResult struct {
	Resp
	List []*InterfaceSummaryHour `json:"list"` // 数据列表
}

// SnsAccessToken 网页授权access_token信息
type SnsAccessToken struct {
	Resp
	AccessToken    string `json:"access_token"`      // 网页授权接口调用凭证
	ExpiresIn      int64  `json:"expires_in"`        // access_token接口调用凭证超时时间，单位（秒）
	RefreshToken   string `json:"refresh_token"`     // 用户刷新access_token
	OpenID         string `json:"openid"`            // 用户唯一标识
	UnionID        string `json:"unionid,omitempty"` // 用户统一标识
	IsSnapshotUser int64  `json:"is_snapshotuser"`   // 是否为快照页模式虚拟账号
	Scope          string `json:"scope,omitempty"`   // 用户授权的作用域
}

// SnsUserInfo 网页授权用户信息
type SnsUserInfo struct {
	Resp
	OpenID     string   `json:"openid"`            // 用户的唯一标识
	Nickname   string   `json:"nickname"`          // 用户昵称
	Sex        int64    `json:"sex"`               // 用户的性别，值为1时是男性，值为2时是女性，值为0时是未知
	Province   string   `json:"province"`          // 用户个人资料填写的省份
	City       string   `json:"city"`              // 普通用户个人资料填写的城市
	Country    string   `json:"country"`           // 国家，如中国为CN
	HeadImgURL string   `json:"headimgurl"`        // 用户头像
	Privilege  []string `json:"privilege"`         // 用户特权信息
	UnionID    string   `json:"unionid,omitempty"` // 只有在用户将公众号绑定到微信开放平台账号后，才会出现该字段
}

// Ticket JS-SDK票据
type Ticket struct {
	Resp
	Ticket    string `json:"ticket"`     // 临时票据
	ExpiresIn int64  `json:"expires_in"` // 有效期（秒）
}

// TranslateContentResult 翻译内容结果
type TranslateContentResult struct {
	Resp
	FromContent string `json:"from_content"` // 原文内容
	ToContent   string `json:"to_content"`   // 译文内容
}

// QueryRecoResultForTextResult 语音识别结果
type QueryRecoResultForTextResult struct {
	Resp
	Result string `json:"result"` // 识别结果
}

// MenuItem 菜单项目
type MenuItem struct {
	Name  string  `json:"name"`  // 菜单名
	Price float64 `json:"price"` // 价格
}

// MenuOcrResult 菜单识别结果
type MenuOcrResult struct {
	Resp
	Content struct {
		MenuItems []*MenuItem `json:"menu_items"` // 菜单内容列表
	} `json:"content"` // 识别的信息
}

// Position 位置信息
type Position struct {
	LeftTop     Coordinate `json:"left_top"`     // 左上角位置
	RightTop    Coordinate `json:"right_top"`    // 右上角位置
	RightBottom Coordinate `json:"right_bottom"` // 右下角位置
	LeftBottom  Coordinate `json:"left_bottom"`  // 左下角位置
}

// Coordinate 坐标信息
type Coordinate struct {
	X int64 `json:"x"` // x坐标
	Y int64 `json:"y"` // y坐标
}

// ImgSize 图片大小
type ImgSize struct {
	W int64 `json:"w"` // 宽度
	H int64 `json:"h"` // 高度
}

// CommOcrItem 通用印刷体识别项目
type CommOcrItem struct {
	Text string    `json:"text"` // 识别文本
	Pos  *Position `json:"pos"`  // 位置信息
}

// CommOcrResult 通用印刷体识别结果
type CommOcrResult struct {
	Resp
	Items   []*CommOcrItem `json:"items"`    // 识别结果
	ImgSize *ImgSize       `json:"img_size"` // 图片大小
}

// DrivingOcrResult 行驶证识别结果
type DrivingOcrResult struct {
	Resp
	PlateNum          string    `json:"plate_num"`           // 车牌号码
	VehicleType       string    `json:"vehicle_type"`        // 车辆类型
	Owner             string    `json:"owner"`               // 所有人
	Addr              string    `json:"addr"`                // 住址
	UseCharacter      string    `json:"use_character"`       // 使用性质
	Model             string    `json:"model"`               // 品牌型号
	Vin               string    `json:"vin"`                 // 车辆识别代号
	EngineNum         string    `json:"engine_num"`          // 发动机号码
	RegisterDate      string    `json:"register_date"`       // 注册日期
	IssueDate         string    `json:"issue_date"`          // 发证日期
	PlateNumB         string    `json:"plate_num_b"`         // 车牌号码
	Record            string    `json:"record"`              // 号牌
	PassengersNum     string    `json:"passengers_num"`      // 核定载人数
	TotalQuality      string    `json:"total_quality"`       // 总质量
	PrepareQuality    string    `json:"prepare_quality"`     // 整备质量
	OverallSize       string    `json:"overall_size"`        // 外廓尺寸
	CardPositionFront *Position `json:"card_position_front"` // 卡片正面位置
	CardPositionBack  *Position `json:"card_position_back"`  // 卡片反面位置
	ImgSize           *ImgSize  `json:"img_size"`            // 图片大小
}

// BankcardOcrResult 银行卡识别结果
type BankcardOcrResult struct {
	Resp
	Number string `json:"number"` // 银行卡号
}

// CertPosition 营业执照位置
type CertPosition struct {
	Pos *Position `json:"pos"` // 位置信息
}

// BizLicenseOcrResult 营业执照识别结果
type BizLicenseOcrResult struct {
	Resp
	RegNum              string        `json:"reg_num"`              // 注册号
	Serial              string        `json:"serial"`               // 编号
	LegalRepresentative string        `json:"legal_representative"` // 法定代表人姓名
	EnterpriseName      string        `json:"enterprise_name"`      // 企业名称
	TypeOfOrganization  string        `json:"type_of_organization"` // 组成形式
	Address             string        `json:"address"`              // 经营场所/企业住所
	TypeOfEnterprise    string        `json:"type_of_enterprise"`   // 公司类型
	BusinessScope       string        `json:"business_scope"`       // 经营范围
	RegisteredCapital   string        `json:"registered_capital"`   // 注册资本
	PaidInCapital       string        `json:"paid_in_capital"`      // 实收资本
	ValidPeriod         string        `json:"valid_period"`         // 营业期限
	RegisteredDate      string        `json:"registered_date"`      // 注册日期/成立日期
	CertPosition        *CertPosition `json:"cert_position"`        // 营业执照位置
	ImgSize             *ImgSize      `json:"img_size"`             // 图片大小
}

// DrivingLicenseOcrResult 驾驶证识别结果
type DrivingLicenseOcrResult struct {
	Resp
	IdNum        string `json:"id_num"`        // 证号
	Name         string `json:"name"`          // 姓名
	Sex          string `json:"sex"`           // 性别
	Address      string `json:"address"`       // 地址
	BirthDate    string `json:"birth_date"`    // 出生日期
	IssueDate    string `json:"issue_date"`    // 初次领证日期
	CarClass     string `json:"car_class"`     // 准驾车型
	ValidFrom    string `json:"valid_from"`    // 有效期限起始日
	ValidTo      string `json:"valid_to"`      // 有效期限终止日
	OfficialSeal string `json:"official_seal"` // 印章文构
}

// IdCardOcrResult 身份证识别结果
type IdCardOcrResult struct {
	Resp
	Type        string `json:"type"`        // 正面或背面，Front / Back
	Name        string `json:"name"`        // 正面返回，姓名
	Id          string `json:"id"`          // 正面返回，身份证号
	ValidDate   string `json:"valid_date"`  // 背面返回，有效期
	Addr        string `json:"addr"`        // 正面返回，地址
	Gender      string `json:"gender"`      // 正面返回，性别
	Nationality string `json:"nationality"` // 正面返回，民族
}

// ImgAiCropResult 图片智能裁剪结果
type ImgAiCropResult struct {
	Resp
	Results []*struct {
		CropLeft   int64 `json:"crop_left"`   // 左上角x
		CropTop    int64 `json:"crop_top"`    // 左上角y
		CropRight  int64 `json:"crop_right"`  // 右下角x
		CropBottom int64 `json:"crop_bottom"` // 右下角y
	} `json:"results"` // 智能裁剪结果
	ImgSize *ImgSize `json:"img_size"` // 图片大小
}

// CodeResult 二维码/条码识别结果
type CodeResult struct {
	TypeName string    `json:"type_name"` // 码的类型
	Data     string    `json:"data"`      // 码的信息
	Pos      *Position `json:"pos"`       // 码的坐标
}

// ImgQrcodeResult 二维码/条码识别结果
type ImgQrcodeResult struct {
	Resp
	CodeResults []*CodeResult `json:"code_results"` // 处理结果
	ImgSize     *ImgSize      `json:"img_size"`     // 图片大小
}

// MerchantCategory 门店小程序类目
type MerchantCategory struct {
	ID       int64   `json:"id"`                 // 类目id
	Name     string  `json:"name,omitempty"`     // 类目名称
	Level    int64   `json:"level"`              // 类目的级别，一级或者二级类目
	Father   int64   `json:"father,omitempty"`   // 父级类目id
	Children []int64 `json:"children,omitempty"` // 子类目id列表
	Qualify  *struct {
		ExterList []struct {
			InnerList []struct {
				Name string `json:"name"` // 证件要求说明
			} `json:"inner_list"`
		} `json:"exter_list"`
	} `json:"qualify,omitempty"` // 资质要求
	Scene         int64 `json:"scene,omitempty"` // 场景
	SensitiveType int64 `json:"sensitive_type"`  // 0或者1， 0表示不用特殊处理 1表示创建该类目的门店小程序时，需要添加相关证件
}

// GetWxaStoreCateListResult 拉取门店小程序类目结果
type GetWxaStoreCateListResult struct {
	Resp
	Data struct {
		AllCategoryInfo struct {
			Categories []*MerchantCategory `json:"categories"` // 类目列表
		} `json:"all_category_info"` // 类目信息
	} `json:"data"` // 数据
}

// ApplyWxaStoreRequest 创建门店小程序请求参数
type ApplyWxaStoreRequest struct {
	FirstCatID        int64  `json:"first_catid"`           // 一级类目id
	SecondCatID       int64  `json:"second_catid"`          // 二级类目id
	HeadImgMediaID    string `json:"headimg_mediaid"`       // 头像临时素材mediaid( 支持jpg和png格式的图片)
	Nickname          string `json:"nickname"`              // 门店小程序的昵称 名称长度为4-30个字符（中文算两个字符）
	Intro             string `json:"intro"`                 // 门店小程序的介绍
	QualificationList string `json:"qualification_list"`    // 类目相关证件，临时素材mediaid
	OrgCode           string `json:"org_code,omitempty"`    // 营业执照或组织代码证，临时素材mediaid
	OtherFiles        string `json:"other_files,omitempty"` // 补充材料，临时素材mediaid
}

// ApplyWxaStoreResult 创建门店小程序结果
type ApplyWxaStoreResult struct {
	Resp
}

// GetWxaStoreAuditInfoResult 查询门店小程序审核结果
type GetWxaStoreAuditInfoResult struct {
	Resp
	Data struct {
		AuditID int64  `json:"audit_id"` // 审核单id
		Status  int64  `json:"status"`   // 审核状态，0：未提交审核，1：审核成功，2：审核中，3：审核失败，4：管理员拒绝
		Reason  string `json:"reason"`   // 审核状态为3或者4时，reason列出审核失败的原因
	} `json:"data"` // 审核结果
}

// ModifyWxaStoreRequest 修改门店小程序信息请求参数
type ModifyWxaStoreRequest struct {
	HeadImgMediaID string `json:"headimg_mediaid"` // 门店头像的临时素材mediaid，不改可传空值
	Intro          string `json:"intro"`           // 门店小程序的介绍，不改可传空值
}

// District 省市区信息
type District struct {
	ID       string   `json:"id"`                 // 区域id，也叫做 districtid
	Name     string   `json:"name,omitempty"`     // 省市区简要名称
	Fullname string   `json:"fullname,omitempty"` // 省市区完整名称
	Pinyin   []string `json:"pinyin,omitempty"`   // 省市区拼音列表
	Location *struct {
		Lat string `json:"lat"` // 纬度坐标
		Lng string `json:"lng"` // 经度坐标
	} `json:"location,omitempty"` // 坐标
	Cidx []int64 `json:"cidx,omitempty"` // 下属地区所有id，可通过此id获取下属地区
}

// GetDistrictListResult 获取省市区信息结果
type GetDistrictListResult struct {
	Status      int64         `json:"status"`       // 状态码
	Message     string        `json:"message"`      // 状态描述
	DataVersion string        `json:"data_version"` // 数据版本
	Result      [][]*District `json:"result"`       // 数据内容，此数组是二维数组，分别代表省市区的信息
}

// SearchMapPoiRequest 搜索门店地图信息请求参数
type SearchMapPoiRequest struct {
	DistrictID int64  `json:"districtid"` // 对应 拉取省市区信息接口 中的id字段
	Keyword    string `json:"keyword"`    // 搜索的关键词
}

// MapPoi 门店地图信息
type MapPoi struct {
	BranchName    string   `json:"branch_name"`            // 门店名称
	Address       string   `json:"address"`                // 地址描述
	Longitude     float64  `json:"longitude"`              // 经度
	Latitude      float64  `json:"latitude"`               // 纬度
	Telephone     string   `json:"telephone"`              // 电话号码
	Category      string   `json:"category"`               // 类目
	SosomapPoiUID string   `json:"sosomap_poi_uid"`        // 地图点位id
	DataSupply    int64    `json:"data_supply"`            // 地图数据
	PicUrls       []string `json:"pic_urls,omitempty"`     // 门店图片列表
	CardIDList    []string `json:"card_id_list,omitempty"` // 门店相应卡券列表
}

// SearchMapPoiResult 搜索门店地图信息结果
type SearchMapPoiResult struct {
	Resp
	Data struct {
		Item []*MapPoi `json:"item"` // 信息数组
	} `json:"data"` // 地图信息
}

// AddStoreRequest 新增门店请求参数
type AddStoreRequest struct {
	MapPoiID          string `json:"map_poi_id"`                   // 从腾讯地图换取的位置点id， 即search_map_poi接口返回的sosomap_poi_uid字段
	PicList           string `json:"pic_list"`                     // 门店图片，可传多张图片，字段是一个 json 字符串
	ContractPhone     string `json:"contract_phone"`               // 联系电话
	Hour              string `json:"hour"`                         // 营业时间，格式11:11-12:12
	Credential        string `json:"credential"`                   // 经营资质证件号
	CompanyName       string `json:"company_name,omitempty"`       // 主体名字，如果复用公众号主体，则company_name为空，如果不复用公众号主体，则company_name为具体的主体名字
	CardID            string `json:"card_id"`                      // 卡券id，如果不需要添加卡券，该参数可为空 目前仅开放支持会员卡、买单和刷卡支付券，不支持自定义code，需要先去公众平台卡券后台创建cardid
	QualificationList string `json:"qualification_list,omitempty"` // 相关证明材料   临时素材mediaid 不复用公众号主体时，才需要填 支持0~5个mediaid，例如mediaid1 或 mediaid2
	PoiID             string `json:"poi_id,omitempty"`             // 如果是从门店管理迁移门店到门店小程序，则需要填该字段
}

// AddStoreResult 新增门店结果
type AddStoreResult struct {
	Resp
	Data struct {
		AuditID int64 `json:"audit_id"` // 审核id
	} `json:"data"` // 新增信息
}

// GetStoreInfoRequest 查询门店详情请求参数
type GetStoreInfoRequest struct {
	PoiID string `json:"poi_id"` // 为门店小程序添加门店，审核成功后返回的门店id
}

// StoreInfo 门店信息
type StoreInfo struct {
	BaseInfo *struct {
		BusinessName string `json:"business_name"` // 门店名称
		Address      string `json:"address"`       // 详细地址
		Telephone    string `json:"telephone"`     // 电话，可多个，使用英文分号间隔
		City         string `json:"city"`          // 城市
		Province     string `json:"province"`      // 省份
		Longitude    string `json:"longitude"`     // 经度
		Latitude     string `json:"latitude"`      // 纬度
		PhotoList    []*struct {
			PhotoURL string `json:"photo_url"` // 图片url
		} `json:"photo_list"` // 门店图片列表
		OpenTime          string `json:"open_time"`          // 门店开放时间
		PoiID             string `json:"poi_id"`             // 门店id
		Status            int64  `json:"status"`             // 审核结果，1-审核通过，2-审核中，3-审核失败
		District          string `json:"district"`           // 区
		QualificationNum  string `json:"qualification_num"`  // 营业执照号
		QualificationName string `json:"qualification_name"` // 营业执照的名称
	} `json:"base_info"`
}

// GetStoreInfoResult 查询门店详情结果
type GetStoreInfoResult struct {
	Resp
	Business *StoreInfo `json:"business"`
}

// GetStoreListRequest 查询门店列表请求参数
type GetStoreListRequest struct {
	Offset int64 `json:"offset"` // 获取门店列表的初始偏移位置，从0开始计数
	Limit  int64 `json:"limit"`  // 获取门店个数
}

// GetStoreListResult 查询门店列表结果
type GetStoreListResult struct {
	Resp
	BusinessList []*StoreInfo `json:"business_list"` // 门店列表
	TotalCount   int64        `json:"total_count"`   // 门店总数
}

// DelStoreRequest 删除门店请求参数
type DelStoreRequest struct {
	PoiID string `json:"poi_id"` // 为门店小程序添加门店，审核成功后返回的门店id
}

// UpdateStoreRequest 更新门店信息请求参数
type UpdateStoreRequest struct {
	PoiID         string `json:"poi_id"`         // 为门店小程序添加门店，审核成功后返回的门店id
	PicList       string `json:"pic_list"`       // 门店图片，可传多张图片，字段是一个 json 字符串
	ContractPhone string `json:"contract_phone"` // 联系电话
	Hour          string `json:"hour"`           // 营业时间，格式11:11-12:12
	CardID        string `json:"card_id"`        // 卡券id，如果不需要添加卡券，该参数可为空 目前仅开放支持会员卡、买单和刷卡支付券，不支持自定义code，需要先去公众平台卡券后台创建cardid
}

// UpdateStoreResult 更新门店信息结果
type UpdateStoreResult struct {
	Resp
	Data struct {
		HasAuditID int64 `json:"has_audit_id"` // 是否需要审核(1表示需要，0表示不需要)
		AuditID    int64 `json:"audit_id"`     // 审核单id
	} `json:"data"` // 更新信息
}

// CreateMapPoiRequest 在地图中创建门店请求参数
type CreateMapPoiRequest struct {
	Name       string `json:"name"`       // 名字
	Longitude  string `json:"longitude"`  // 经度
	Latitude   string `json:"latitude"`   // 纬度
	Province   string `json:"province"`   // 省份
	City       string `json:"city"`       // 城市
	District   string `json:"district"`   // 区
	Address    string `json:"address"`    // 详细地址
	Category   string `json:"category"`   // 类目，比如美食:中餐厅
	Telephone  string `json:"telephone"`  // 电话，可多个，使用英文分号间隔
	Photo      string `json:"photo"`      // 门店图片url
	License    string `json:"license"`    // 营业执照url
	Introduct  string `json:"introduct"`  // 介绍
	DistrictID string `json:"districtid"` // 腾讯地图拉取省市区信息接口返回的id
	PoiID      string `json:"poi_id"`     // 如果是迁移门店， 必须填 poi_id字段
}

// CreateMapPoiResult 在地图中创建门店结果
type CreateMapPoiResult struct {
	Resp
	Data struct {
		BaseID int64 `json:"base_id"` // 审核单id
		RichID int64 `json:"rich_id"`
	} `json:"data"` // 数据
}

// DNS DNS解析结果
type DNS struct {
	IP           string `json:"ip"`            // 解出来的ip
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
