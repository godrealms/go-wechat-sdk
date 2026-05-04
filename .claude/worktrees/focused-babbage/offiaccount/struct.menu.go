package offiaccount

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
