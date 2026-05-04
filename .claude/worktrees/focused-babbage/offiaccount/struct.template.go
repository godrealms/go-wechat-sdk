package offiaccount

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

type Category struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type CategoryResp struct {
	Resp
	Data []*Category `json:"data"`
}
type TemplateKeyWord struct {
	Kid     int    `json:"kid"`
	Name    string `json:"name"`
	Example string `json:"example"`
	Rule    string `json:"rule"`
}
type TemplateKeyWordsResp struct {
	Resp
	Data []*TemplateKeyWord `json:"data"`
}

type TemplateTitle struct {
	Tid        int    `json:"tid"`
	Title      string `json:"title"`
	Type       int    `json:"type"`
	CategoryId string `json:"categoryId"`
}

type TemplateTitlesResp struct {
	Resp
	Count int              `json:"count"`
	Data  []*TemplateTitle `json:"data"`
}

type KeywordEnumValue struct {
	EnumValueList []string `json:"enumValueList"`
	KeywordCode   string   `json:"keywordCode"`
}

type PubTemplate struct {
	PriTmplId            string              `json:"priTmplId"`
	Title                string              `json:"title"`
	Content              string              `json:"content"`
	Example              string              `json:"example"`
	Type                 int                 `json:"type"`
	KeywordEnumValueList []*KeywordEnumValue `json:"keywordEnumValueList,omitempty"`
}

type PubNewTemplateResp struct {
	Resp
	Data []*PubTemplate `json:"data"`
}

type AddWxaNewTemplateResp struct {
	Resp
	PriTmplId string `json:"priTmplId"`
}

type Value struct {
	Value string `json:"value"`
}

type SubscribeMsg struct {
	//接收者（用户）的 openid
	ToUser string `json:"touser"`
	//所需下发的订阅模板id
	TemplateId string `json:"template_id"`
	//点击模板卡片后的跳转页面，仅限本小程序内的页面。支持带参数,（示例index?foo=bar）。该字段不填则模板无跳转
	Page        string       `json:"page"`
	MiniProgram *MiniProgram `json:"miniprogram"`
	ClientMsgId string       `json:"client_msg_id"`
	//模板内容，格式形如{ "phrase3": { "value": "审核通过" }, "name1": { "value": "订阅" }, "date2": { "value": "2019-12-25 09:42" } }
	Data map[string]*Value `json:"data"`
}

type Content struct {
	Value string `json:"value,omitempty"`
	Color string `json:"color,omitempty"`
}

type TemplateSubscribeContent struct {
	Content *Content `json:"content"`
}

type TemplateSubscribeReq struct {
	ToUser      string                    `json:"touser"`
	TemplateId  string                    `json:"template_id"`
	MiniProgram *MiniProgram              `json:"miniprogram,omitempty"`
	Url         string                    `json:"url,omitempty"`
	Scene       string                    `json:"scene"`
	Title       string                    `json:"title"`
	Data        *TemplateSubscribeContent `json:"data"`
}
