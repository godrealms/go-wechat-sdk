package offiaccount

// NonTaxItem 缴费子项目详情
type NonTaxItem struct {
	No       int64  `json:"no"`        // 项目号，如1，2，3
	ItemID   string `json:"item_id"`   // 项目编码
	ItemName string `json:"item_name"` // 项目名称
	Overdue  int64  `json:"overdue"`   // 滞纳金（单位是分）
	Penalty  int64  `json:"penalty"`   // 加罚金额（单位是分）
	Fee      int64  `json:"fee"`       // 金额（包含滞纳金和加罚金额，单位是分）
}

// QueryFeeRequest 查询应收信息请求参数
type QueryFeeRequest struct {
	AppID             string `json:"appid"`                         // appid
	ServiceID         int64  `json:"service_id"`                    // 服务id
	BankID            string `json:"bank_id,omitempty"`             // 银行id（由微信非税平台分配的全局唯一id），不指定时在已配置的银行列表中随机选择
	PaymentNoticeNo   string `json:"payment_notice_no"`             // 缴费通知书编号
	DepartmentCode    string `json:"department_code"`               // 执收单位编码
	PaymentNoticeType int64  `json:"payment_notice_type,omitempty"` // 通知书类型，1：普通通知书类型；2：处罚通知书类型
	RegionCode        string `json:"region_code"`                   // 行政区划代码
}

// QueryFeeResult 查询应收信息结果
type QueryFeeResult struct {
	Resp
	UserName                string       `json:"user_name"`                  // 用户姓名
	Fee                     int64        `json:"fee"`                        // 总金额（单位是分）
	Items                   []NonTaxItem `json:"items"`                      // 缴费子项目详情
	PaymentNoticeNo         string       `json:"payment_notice_no"`          // 缴费通知书编号
	DepartmentCode          string       `json:"department_code"`            // 执收单位编码
	DepartmentName          string       `json:"department_name"`            // 执收单位名称
	PaymentNoticeType       int64        `json:"payment_notice_type"`        // 通知书类型
	RegionCode              string       `json:"region_code"`                // 行政区划代码
	PaymentNoticeCreateTime int64        `json:"payment_notice_create_time"` // 缴费通知书创建时间（时间戳，单位是秒）
	PaymentExpireDate       string       `json:"payment_expire_date"`        // 限缴日期，格式YYYYMMDD
}

// QueryFee 查询应收信息
// https://developers.weixin.qq.com/doc/service/api/nontaxpay/api_nontaxqueryfee.html
func (c *Client) QueryFee(request *QueryFeeRequest) (*QueryFeeResult, error) {
	result := &QueryFeeResult{}
	err := c.Https.Post(c.ctx, "/nontax/queryfee", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UnifiedOrderRequest 缴费支付下单请求参数
type UnifiedOrderRequest struct {
	AppID                   string       `json:"appid"`                         // appid
	ServiceID               int64        `json:"service_id,omitempty"`          // 服务id
	BankID                  string       `json:"bank_id,omitempty"`             // 银行id（由微信非税平台分配的全局唯一id），不指定时在已配置的银行列表中随机选择；不填时默认生产环境，测试环境需填上
	BankAccount             string       `json:"bank_account,omitempty"`        // 清分银行账号（不使用清分机制的不用填）
	MchID                   string       `json:"mch_id,omitempty"`              // 指定资金结算到mch_id。只能结算到 bank_id 下绑定的mch_id。不填时自动从 bank_id 下绑定的 mch_id 选择一个。
	OpenID                  string       `json:"openid,omitempty"`              // 用户标识。当 trade_type 为 MWEB 时，不需要填；其它情况，必填。
	Desc                    string       `json:"desc"`                          // 描述（服务名称）
	Fee                     int64        `json:"fee"`                           // 总金额（单位是分）
	ReturnURL               string       `json:"return_url,omitempty"`          // 支付中间页支付完成后跳转的页面（小程序下单非必填，其他必填）
	IP                      string       `json:"ip"`                            // 用户端ip
	OrderNo                 string       `json:"order_no,omitempty"`            // 订单号。（缴费通知书编号和订单号必须二选一。如果没有缴费通知书编号，则填订单号）
	PaymentNoticeNo         string       `json:"payment_notice_no,omitempty"`   // 缴费通知书编号（缴费通知书编号和订单号必须二选一。如果没有缴费通知书编号，则填订单号）
	DepartmentCode          string       `json:"department_code"`               // 执收单位编码
	DepartmentName          string       `json:"department_name"`               // 执收单位名称
	PaymentNoticeType       int64        `json:"payment_notice_type,omitempty"` // 通知书类型
	RegionCode              string       `json:"region_code"`                   // 行政区划代码
	UserName                string       `json:"user_name,omitempty"`           // 用户姓名
	Items                   []NonTaxItem `json:"items"`                         // 缴费子项目详情
	PaymentNoticeCreateTime int64        `json:"payment_notice_create_time"`    // 缴款通知书创建时间（时间戳，单位是秒）
	PaymentExpireDate       string       `json:"payment_expire_date,omitempty"` // 限缴日期，格式YYYYMMDD
	Scene                   string       `json:"scene"`                         // 场景。"biz":微信公众号"ctiyservice":城市服务"miniprogram":小程序
	AppAppID                string       `json:"app_appid,omitempty"`           // app的appid，app下单时必填
	TradeType               string       `json:"trade_type,omitempty"`          // 默认是JSAPI类型，非微信浏览器的H5下单请填MWEB。
	AutoCallPay             bool         `json:"auto_call_pay,omitempty"`       // 支付中间页/小程序是否自动调起支付
}

// UnifiedOrderResult 缴费支付下单结果
type UnifiedOrderResult struct {
	Resp
	UserName                string       `json:"user_name"`                  // 用户姓名
	Fee                     int64        `json:"fee"`                        // 总金额（单位是分）
	Items                   []NonTaxItem `json:"items"`                      // 缴费子项目详情
	PaymentNoticeNo         string       `json:"payment_notice_no"`          // 缴费通知书编号
	DepartmentCode          string       `json:"department_code"`            // 执收单位编码
	DepartmentName          string       `json:"department_name"`            // 执收单位名称
	PaymentNoticeType       int64        `json:"payment_notice_type"`        // 通知书类型
	RegionCode              string       `json:"region_code"`                // 行政区划代码
	PaymentNoticeCreateTime int64        `json:"payment_notice_create_time"` // 缴费通知书创建时间（时间戳，单位是秒）
	PaymentExpireDate       string       `json:"payment_expire_date"`        // 限缴日期，格式YYYYMMDD
}

// UnifiedOrder 缴费支付下单
// https://developers.weixin.qq.com/doc/service/api/nontaxpay/api_nontaxunifiedorder.html
func (c *Client) UnifiedOrder(request *UnifiedOrderRequest) (*UnifiedOrderResult, error) {
	result := &UnifiedOrderResult{}
	err := c.Https.Post(c.ctx, "/nontax/unifiedorder", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DownloadBillRequest 下载缴费对帐单请求参数
type DownloadBillRequest struct {
	AppID    string `json:"appid"`               // appid
	MchID    string `json:"mch_id"`              // 商户号
	BillDate string `json:"bill_date"`           // 对账单日期，格式：20170903
	BillType string `json:"bill_type,omitempty"` // ALL，返回当日所有订单信息，默认值；SUCCESS，返回当日成功支付的订单；REFUND，返回当日退款订单
}

// DownloadBillResult 下载缴费对帐单结果
type DownloadBillResult struct {
	Resp
	// 返回的是CSV格式的对账单数据，以字符串形式返回
}

// DownloadBill 下载缴费对帐单
// https://developers.weixin.qq.com/doc/service/api/nontaxpay/api_nontaxdownloadbill.html
func (c *Client) DownloadBill(request *DownloadBillRequest) (*DownloadBillResult, error) {
	result := &DownloadBillResult{}
	err := c.Https.Post(c.ctx, "/nontax/downloadbill", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// NotifyInconsistentOrderRequest 通知不一致订单请求参数
type NotifyInconsistentOrderRequest struct {
	AppID   string `json:"appid"`    // appid
	OrderID string `json:"order_id"` // 订单号
}

// NotifyInconsistentOrder 通知不一致订单
// https://developers.weixin.qq.com/doc/service/api/nontaxpay/api_nontaxnotifyinconsistentorder.html
func (c *Client) NotifyInconsistentOrder(request *NotifyInconsistentOrderRequest) (*Resp, error) {
	result := &Resp{}
	err := c.Https.Post(c.ctx, "/nontax/notifyinconsistentorder", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// MockNotificationRequest 模拟支付结果通知请求参数
type MockNotificationRequest struct {
	AppID   string `json:"appid"`   // appid
	URL     string `json:"url"`     // 接收通知的url
	Version int64  `json:"version"` // 协议版本号，默认为1
}

// MockNotification 模拟支付结果通知
// https://developers.weixin.qq.com/doc/service/api/nontaxpay/api_nontaxmocknotification.html
func (c *Client) MockNotification(request *MockNotificationRequest) (*Resp, error) {
	result := &Resp{}
	err := c.Https.Post(c.ctx, "/nontax/mocknotification", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// MockQueryFeeRequest 模拟查询应收信息请求参数
type MockQueryFeeRequest struct {
	AppID   string `json:"appid"`   // appid
	URL     string `json:"url"`     // 接收通知的url
	Version int64  `json:"version"` // 协议版本号，默认为1
}

// MockQueryFee 模拟查询应收信息
// https://developers.weixin.qq.com/doc/service/api/nontaxpay/api_nontaxmockqueryfee.html
func (c *Client) MockQueryFee(request *MockQueryFeeRequest) (*Resp, error) {
	result := &Resp{}
	err := c.Https.Post(c.ctx, "/nontax/mockqueryfee", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// MicroPayRequest 提交刷卡支付请求参数
type MicroPayRequest struct {
	AppID                   string       `json:"appid"`                         // appid
	BankID                  string       `json:"bank_id,omitempty"`             // 银行id（由微信非税平台分配的全局唯一id），不指定时在已配置的银行列表中随机选择；不填时默认生产环境，测试环境需填上
	BankAccount             string       `json:"bank_account,omitempty"`        // 清分银行账号（不使用清分机制的不用填）
	MchID                   string       `json:"mch_id,omitempty"`              // 指定资金结算到mch_id。只能结算到 bank_id 下绑定的mch_id。不填时自动从 bank_id 下绑定的 mch_id 选择一个。
	Desc                    string       `json:"desc"`                          // 描述（服务名称）
	Fee                     int64        `json:"fee"`                           // 总金额（单位是分）
	UserName                string       `json:"user_name,omitempty"`           // 用户姓名
	Items                   []NonTaxItem `json:"items"`                         // 缴费子项目详情
	PaymentNoticeCreateTime int64        `json:"payment_notice_create_time"`    // 缴款通知书创建时间（时间戳，单位是秒）
	PaymentExpireDate       string       `json:"payment_expire_date,omitempty"` // 限缴日期，格式YYYYMMDD
	PaymentNoticeNo         string       `json:"payment_notice_no,omitempty"`   // 缴费通知书编号（缴费通知书编号和订单号必须二选一。如果没有缴费通知书编号，则填订单号）
	OrderNo                 string       `json:"order_no,omitempty"`            // 订单号。（缴费通知书编号和订单号必须二选一。如果没有缴费通知书编号，则填订单号）
	DepartmentCode          string       `json:"department_code"`               // 执收单位编码
	DepartmentName          string       `json:"department_name"`               // 执收单位名称
	PaymentNoticeType       int64        `json:"payment_notice_type,omitempty"` // 通知书类型
	RegionCode              string       `json:"region_code"`                   // 行政区划代码
	AuthCode                string       `json:"auth_code"`                     // 扫码支付授权码，设备读取用户微信中的条码或者二维码信息（注：用户刷卡条形码规则：18位纯数字，以10、11、12、13、14、15开头）
	OrderID                 string       `json:"order_id,omitempty"`            // 订单号（之前请求有返回订单号则填上）
}

// MicroPayResult 提交刷卡支付结果
type MicroPayResult struct {
	Resp
	OrderID string `json:"order_id"` // 订单号
}

// MicroPay 提交刷卡支付
// https://developers.weixin.qq.com/doc/service/api/nontaxpay/api_nontaxmicropay.html
func (c *Client) MicroPay(request *MicroPayRequest) (*MicroPayResult, error) {
	result := &MicroPayResult{}
	err := c.Https.Post(c.ctx, "/nontax/micropay", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetOrderListRequest 获取缴费订单列表请求参数
type GetOrderListRequest struct {
	AppID           string `json:"appid"`                       // appid
	RegionCode      string `json:"region_code"`                 // 行政区划代码
	DepartmentCode  string `json:"department_code"`             // 执收单位编码
	PaymentNoticeNo string `json:"payment_notice_no,omitempty"` // 缴费通知书编号（缴费通知书编号和订单号必须二选一。如果没有缴费通知书编号，则填订单号）
}

// GetOrderListResult 获取缴费订单列表结果
type GetOrderListResult struct {
	Resp
	OrderIDList []string `json:"order_id_list"` // 已下单的订单id列表
	PaidOrderID string   `json:"paid_order_id"` // 已支付的订单id
}

// GetOrderList 获取缴费订单列表
// https://developers.weixin.qq.com/doc/service/api/nontaxpay/api_nontaxgetorderlist.html
func (c *Client) GetOrderList(request *GetOrderListRequest) (*GetOrderListResult, error) {
	result := &GetOrderListResult{}
	err := c.Https.Post(c.ctx, "/nontax/getorderlist", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// RefundRequest 缴费订单退款请求参数
type RefundRequest struct {
	AppID       string `json:"appid"`                   // appid
	OrderID     string `json:"order_id"`                // 订单号
	Reason      string `json:"reason"`                  // 退款原因
	RefundFee   int64  `json:"refund_fee,omitempty"`    // 退款金额（单位是分），部分退款时必填
	RefundOutID string `json:"refund_out_id,omitempty"` // 退款单号（每笔部分退款唯一），部分退款时必填
}

// RefundResult 缴费订单退款结果
type RefundResult struct {
	Resp
	RefundOrderID string `json:"refund_order_id"` // 退款订单号
}

// Refund 缴费订单退款
// https://developers.weixin.qq.com/doc/service/api/nontaxpay/api_nontaxrefund.html
func (c *Client) Refund(request *RefundRequest) (*RefundResult, error) {
	result := &RefundResult{}
	err := c.Https.Post(c.ctx, "/nontax/refund", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetOrderRequest 获取缴费订单详情请求参数
type GetOrderRequest struct {
	AppID     string `json:"appid"`                // appid
	ServiceID int64  `json:"service_id,omitempty"` // 服务id
	OrderID   string `json:"order_id"`             // 订单id
}

// PartialRefundInfo 部分退款信息
type PartialRefundInfo struct {
	RefundOrderID    string `json:"refund_order_id"`    // 退款订单号
	RefundReason     string `json:"refund_reason"`      // 退款原因
	RefundFee        int64  `json:"refund_fee"`         // 退款金额（单位是分）
	RefundFinishTime int64  `json:"refund_finish_time"` // 退款完成时间（时间戳，单位是秒）
	RefundOutID      string `json:"refund_out_id"`      // 退款时传入的外部单号
	RefundStatus     int64  `json:"refund_status"`      // 退款状态;5：已退款;6：退款中
}

// NotifyDetail 通知详情（第一次和最后一次通知）
type NotifyDetail struct {
	NotifyTime    int64  `json:"notify_time"`     // 通知时间（时间戳，单位是秒）
	Ret           int64  `json:"ret"`             // 微信后台通知总返回码
	RetErrmsg     string `json:"ret_errmsg"`      // 微信后台通知总返回信息
	CostTime      int64  `json:"cost_time"`       // 耗时（单位是毫秒）
	WxNonTaxStr   string `json:"wxnontaxstr"`     // 带在url参数上的一次请求的随机字符串
	Status        int64  `json:"status"`          // 订单状态;3或4：支付成功;5：已退款
	URL           string `json:"url"`             // 第三方接收通知的url
	ErrCode       int64  `json:"errcode"`         // 第三方返回码;0– 成功;210 – 数据格式错误;232 – 缴款通知书已缴费;236 – 不允许在该银行缴费;298 – 解密失败;299 – 系统错误;300 – 签名错误
	ErrMsg        string `json:"errmsg"`          // 第三方返回信息，如非空，为错误原因
	ThirdResp     string `json:"third_resp"`      // 第三方的返回
	ThirdRespData string `json:"third_resp_data"` // 第三方的返回解密出的data
}

// NotifyHistory 通知历史
type NotifyHistory struct {
	AppID        string         `json:"appid"`         // 第三方appid
	Name         string         `json:"name"`          // 第三方名字
	NotifyDetail []NotifyDetail `json:"notify_detail"` // 通知详情（第一次和最后一次通知）
	NotifyCnt    int64          `json:"notify_cnt"`    // 通知次数
}

// GetOrderResult 获取缴费订单详情结果
type GetOrderResult struct {
	Resp
	AppID             string              `json:"appid"`               // appid
	OpenID            string              `json:"openid"`              // 用户标识
	OrderID           string              `json:"order_id"`            // 订单号
	CreateTime        int64               `json:"create_time"`         // 订单创建时间（时间戳，单位是秒）
	PayFinishTime     int64               `json:"pay_finish_time"`     // 订单支付成功时间（时间戳，单位是秒）
	Desc              string              `json:"desc"`                // 描述（服务名称）
	Fee               int64               `json:"fee"`                 // 总金额(单位是分)
	FeeType           int64               `json:"fee_type"`            // 币种1：人民币2：美元
	TransID           string              `json:"trans_id"`            // 支付交易单号
	Status            int64               `json:"status"`              // 订单总状态1：还没支付;3或4：支付成功;5：已退款;6：退款中;12：超时未支付订单自动关闭（若部分退款只退了一部分金额，订单总状态不会变，只有全部退完总状态才会变成已退款）
	BankID            string              `json:"bank_id"`             // 银行id（由微信非税平台分配的全局唯一id）
	BankName          string              `json:"bank_name"`           // 银行名称
	BankAccount       string              `json:"bank_account"`        // 银行账号
	RefundFinishTime  int64               `json:"refund_finish_time"`  // 退款完成时间（时间戳，单位是秒）
	RefundReason      string              `json:"refund_reason"`       // 退款原因
	RefundOrderID     string              `json:"refund_order_id"`     // 退款订单号
	RefundOutID       string              `json:"refund_out_id"`       // 退款时传入的外部单号
	PaymentNoticeNo   string              `json:"payment_notice_no"`   // 缴费通知书编号（根据下单请求的参数返回）
	OrderNo           string              `json:"order_no"`            // 订单号。（根据下单请求的参数返回）
	DepartmentCode    string              `json:"department_code"`     // 执收单位编码
	DepartmentName    string              `json:"department_name"`     // 执收单位名称
	PaymentNoticeType int64               `json:"payment_notice_type"` // 通知书类型
	RegionCode        string              `json:"region_code"`         // 行政区划代码
	Items             []NonTaxItem        `json:"items"`               // 缴费子项目详情
	BillTypeCode      string              `json:"bill_type_code"`      // 票据类型编码
	BillNo            string              `json:"bill_no"`             // 票据号码
	PaymentInfoSource int64               `json:"payment_info_source"` // 应收款信息来源，1：财政2：委办局
	PartialRefundInfo []PartialRefundInfo `json:"partial_refund_info"` // 部分退款信息
	NotifyHistory     []NotifyHistory     `json:"notify_history"`      // 通知历史
	Scene             string              `json:"scene"`               // 场景。"biz":微信公众号"ctiyservice":城市服务"miniprogram":小程序"offline":线下二维码"pc":pc机"app":手机app"other":其它
}

// GetOrder 获取缴费订单详情
// https://developers.weixin.qq.com/doc/service/api/nontaxpay/api_nontaxgetorder.html
func (c *Client) GetOrder(request *GetOrderRequest) (*GetOrderResult, error) {
	result := &GetOrderResult{}
	err := c.Https.Post(c.ctx, "/nontax/getorder", request, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
