package offiaccount

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

type SubButton struct {
	// 菜单的响应动作类型（与 sub_button 互斥）
	//	click
	//	view
	//	scancode_push
	//	scancode_waitmsg
	//	pic_sysphoto
	//	pic_photo_or_album
	//	pic_weixin
	//	location_select
	//	media_id
	//	article_id
	//	article_view_limited
	Type string `json:"type"` // 按钮类型
	// 菜单标题，不超过16个字节，子菜单不超过60个字节
	Name string `json:"name"`
	// 菜单KEY值，用于消息接口推送，不超过128字节。click等点击类型必须。
	Key string `json:"key,omitempty"` // 按钮的key
	// 网页链接，用户点击菜单可打开链接，不超过1024字节。 type为miniprogram时，不支持小程序的老版本客户端将打开本url。view、miniprogram类型必填。
	Url string `json:"url,omitempty"`
	// 调用新增永久素材接口返回的合法media_id。media_id类型和view_limited类型必须
	MediaId string `json:"media_id,omitempty"`
	// 小程序的appid（仅认证公众号可配置），miniprogram类型必须
	Appid string `json:"appid,omitempty"`
	// 小程序的页面路径，miniprogram类型必须
	PagePath string `json:"pagepath,omitempty"`
	// 发布后获得的合法 ArticleId，article_id类型和article_view_limited类型必须
	ArticleId string `json:"article_Id,omitempty"`
}
type Button struct {
	// 菜单的响应动作类型（与 sub_button 互斥）
	//	click
	//	view
	//	scancode_push
	//	scancode_waitmsg
	//	pic_sysphoto
	//	pic_photo_or_album
	//	pic_weixin
	//	location_select
	//	media_id
	//	article_id
	//	article_view_limited
	Type string `json:"type,omitempty"`
	// 菜单标题，不超过16个字节，子菜单不超过60个字节
	Name string `json:"name"`
	// 菜单KEY值，用于消息接口推送，不超过128字节。click等点击类型必须。
	Key string `json:"key,omitempty"` // 按钮的key
	// 网页链接，用户点击菜单可打开链接，不超过1024字节。 type为miniprogram时，不支持小程序的老版本客户端将打开本url。view、miniprogram类型必填。
	Url string `json:"url,omitempty"`
	// 调用新增永久素材接口返回的合法media_id。media_id类型和view_limited类型必须
	MediaId string `json:"media_id,omitempty"`
	// 小程序的appid（仅认证公众号可配置），miniprogram类型必须
	Appid string `json:"appid,omitempty"`
	// 小程序的页面路径，miniprogram类型必须
	PagePath string `json:"pagepath,omitempty"`
	// 发布后获得的合法 ArticleId，article_id类型和article_view_limited类型必须
	ArticleId string `json:"article_Id,omitempty"`
	// 二级菜单结构体数组
	SubButton []*SubButton `json:"sub_button,omitempty"`
}
type CustomMenu struct {
	Button []*Button `json:"button"` // 按钮列表
}
