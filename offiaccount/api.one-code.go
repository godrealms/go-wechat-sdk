package offiaccount

// ApplyCodeRequest 申请二维码请求参数
type ApplyCodeRequest struct {
	CodeCount        int64  `json:"code_count"`         // 申请码的数量，≥10000，≤20000000，10000的整数倍
	IsvApplicationID string `json:"isv_application_id"` // 外部单号，相同isv_application_id视为同一申请单
}

// ApplyCodeResult 申请二维码结果
type ApplyCodeResult struct {
	Resp
	ApplicationID int64 `json:"application_id"` // 申请单号
}

// ApplyCode 申请二维码接口用于批量生成指定数量的营销码
// https://developers.weixin.qq.com/doc/service/api/onecode/api_intp_marketcode_applycode.html
func (c *Client) ApplyCode(request *ApplyCodeRequest) (*ApplyCodeResult, error) {
	result := &ApplyCodeResult{}
	err := c.Https.Post(c.ctx, "/intp/marketcode/applycode", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ApplyCodeQueryRequest 查询二维码申请单请求参数
type ApplyCodeQueryRequest struct {
	ApplicationID    int64  `json:"application_id,omitempty"`     // 申请单号
	IsvApplicationID string `json:"isv_application_id,omitempty"` // 外部单号，相同isv_application_id视为同一申请单
}

// CodeGenerateInfo 二维码信息
type CodeGenerateInfo struct {
	CodeStart int64 `json:"code_start"` // 开始位置
	CodeEnd   int64 `json:"code_end"`   // 结束位置
}

// ApplyCodeQueryResult 查询二维码申请单结果
type ApplyCodeQueryResult struct {
	Resp
	Status           string             `json:"status"`             // 申请单状态
	ApplicationID    int64              `json:"application_id"`     // 申请单号
	IsvApplicationID string             `json:"isv_application_id"` // 外部单号
	CodeGenerateList []CodeGenerateInfo `json:"code_generate_list"` // 二维码信息
	CreateTime       int64              `json:"create_time"`        // 创建时间
	UpdateTime       int64              `json:"update_time"`        // 更新时间
}

// ApplyCodeQuery 查询二维码申请单状态及详细信息
// https://developers.weixin.qq.com/doc/service/api/onecode/api_intp_marketcode_applycodequery.html
func (c *Client) ApplyCodeQuery(request *ApplyCodeQueryRequest) (*ApplyCodeQueryResult, error) {
	result := &ApplyCodeQueryResult{}
	err := c.Https.Post(c.ctx, "/intp/marketcode/applycodequery", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ApplyCodeDownloadRequest 下载二维码包请求参数
type ApplyCodeDownloadRequest struct {
	ApplicationID int64 `json:"application_id"` // 申请单号
	CodeStart     int64 `json:"code_start"`     // 开始位置，来自查询二维码申请接口
	CodeEnd       int64 `json:"code_end"`       // 结束位置，来自查询二维码申请接口
}

// ApplyCodeDownloadResult 下载二维码包结果
type ApplyCodeDownloadResult struct {
	Resp
	Buffer string `json:"buffer"` // 文件buffer，需要先base64 decode，再做解密操作
}

// ApplyCodeDownload 下载生成的二维码数据包
// https://developers.weixin.qq.com/doc/service/api/onecode/api_intp_marketcode_applycodedownload.html
func (c *Client) ApplyCodeDownload(request *ApplyCodeDownloadRequest) (*ApplyCodeDownloadResult, error) {
	result := &ApplyCodeDownloadResult{}
	err := c.Https.Post(c.ctx, "/intp/marketcode/applycodedownload", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CodeActiveRequest 激活二维码请求参数
type CodeActiveRequest struct {
	ApplicationID int64  `json:"application_id"`     // 申请单号
	ActivityName  string `json:"activity_name"`      // 活动名称,数据分析活动区分依据，请规范命名
	ProductBrand  string `json:"product_brand"`      // 商品品牌,数据分析品牌区分依据，请规范命名
	ProductTitle  string `json:"product_title"`      // 商品标题,数据分析商品区分依据，请规范命名
	ProductCode   string `json:"product_code"`       // 商品条码,EAN商品条码，请规范填写
	WxaAppid      string `json:"wxa_appid"`          // 小程序的appid,扫码跳转小程序的appid
	WxaPath       string `json:"wxa_path"`           // 小程序的path,扫码跳转小程序的path
	WxaType       int64  `json:"wxa_type,omitempty"` // 小程序版本,默认为0正式版，开发版为1，体验版为2
	CodeStart     int64  `json:"code_start"`         // 激活码段的起始位,如0（包含该值）
	CodeEnd       int64  `json:"code_end"`           // 激活码段的结束位,如9999（包含该值）
}

// CodeActive 激活指定范围的二维码用于实际营销活动
// https://developers.weixin.qq.com/doc/service/api/onecode/api_intp_marketcode_codeactive.html
func (c *Client) CodeActive(request *CodeActiveRequest) (*Resp, error) {
	result := &Resp{}
	err := c.Https.Post(c.ctx, "/intp/marketcode/codeactive", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CodeActiveQueryRequest 查询二维码激活状态请求参数
type CodeActiveQueryRequest struct {
	ApplicationID int64  `json:"application_id,omitempty"` // 申请单号
	CodeIndex     int64  `json:"code_index,omitempty"`     // 该码在批次中的偏移量,传入application_id时必填
	CodeURL       string `json:"code_url,omitempty"`       // 28位普通码字符,code与code_url二选一
	Code          string `json:"code,omitempty"`           // 九位的字符串原始码,code与code_url二选一
}

// CodeActiveQueryResult 查询二维码激活状态结果
type CodeActiveQueryResult struct {
	Resp
	Code             string `json:"code"`               // 原始码,返回原始码数据，并返回对应的激活信息
	ApplicationID    int64  `json:"application_id"`     // 申请单号
	IsvApplicationID string `json:"isv_application_id"` // 外部单号
	ActivityName     string `json:"activity_name"`      // 活动名称
	ProductBrand     string `json:"product_brand"`      // 商品品牌
	ProductTitle     string `json:"product_title"`      // 商品标题
	WxaAppid         string `json:"wxa_appid"`          // 小程序的appid
	WxaPath          string `json:"wxa_path"`           // 小程序的path
	WxaType          int64  `json:"wxa_type"`           // 小程序版本
	CodeStart        int64  `json:"code_start"`         // 激活码段的起始位,如0（包含该值）
	CodeEnd          int64  `json:"code_end"`           // 激活码段的结束位,如9999（包含该值）
}

// CodeActiveQuery 查询二维码激活状态及关联信息
// https://developers.weixin.qq.com/doc/service/api/onecode/api_intp_marketcode_codeactivequery.html
func (c *Client) CodeActiveQuery(request *CodeActiveQueryRequest) (*CodeActiveQueryResult, error) {
	result := &CodeActiveQueryResult{}
	err := c.Https.Post(c.ctx, "/intp/marketcode/codeactivequery", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// TicketToCodeRequest CODE_TICKET换CODE请求参数
type TicketToCodeRequest struct {
	OpenID     string `json:"openid"`      // 用户的openid
	CodeTicket string `json:"code_ticket"` // 跳转时带上的code_ticket参数
}

// TicketToCodeResult CODE_TICKET换CODE结果
type TicketToCodeResult struct {
	Resp
	Code             string `json:"code"`               // 原始码,返回原始码数据，并返回对应的激活信息
	ApplicationID    int64  `json:"application_id"`     // 申请单号
	IsvApplicationID string `json:"isv_application_id"` // 外部单号
	ActivityName     string `json:"activity_name"`      // 活动名称
	ProductBrand     string `json:"product_brand"`      // 商品品牌
	ProductTitle     string `json:"product_title"`      // 商品标题
	WxaAppid         string `json:"wxa_appid"`          // 小程序的appid
	WxaPath          string `json:"wxa_path"`           // 小程序的path
	CodeStart        int64  `json:"code_start"`         // 激活码段的起始位,如0（包含该值）
	CodeEnd          int64  `json:"code_end"`           // 激活码段的结束位,如9999（包含该值）
}

// TicketToCode 将用户扫码获得的临时票据转换为正式营销码
// https://developers.weixin.qq.com/doc/service/api/onecode/api_intp_marketcode_tickettocode.html
func (c *Client) TicketToCode(request *TicketToCodeRequest) (*TicketToCodeResult, error) {
	result := &TicketToCodeResult{}
	// 注意：这个接口不需要access_token参数
	err := c.Https.Post(c.ctx, "/intp/marketcode/tickettocode", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
