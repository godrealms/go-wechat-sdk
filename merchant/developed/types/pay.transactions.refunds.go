package types

import (
	"encoding/json"
	"time"
)

// RefundFrom
// 【退款出资账户及金额】退款需从指定账户出资时，可传递该参数以指定出资金额（币种最小单位，仅限整数）。
// 多账户出资退款需满足：1、未开通退款支出分离功能；2、订单为待分账或分账中的分账订单。
// 传递参数需确保：1、基本账户可用与不可用余额之和等于退款金额；2、账户类型不重复。不符条件将返回错误。
type RefundFrom struct {
	//【出资账户类型】 退款出资的账户类型。
	//	可选取值：
	//	AVAILABLE : 可用余额
	//	UNAVAILABLE : 不可用余额
	Account string `json:"account"`
	//【出资金额】对应账户出资金额
	Amount int `json:"amount"`
}

// Refunds
// 退款申请
type Refunds struct {
	//【微信支付订单号】 微信支付侧订单的唯一标识，订单支付成功后，查询订单和支付成功回调通知会返回该参数。
	//	transaction_id和out_trade_no必须二选一进行传参。
	TransactionId string `json:"transaction_id,omitempty"`
	//【商户订单号】 商户下单时传入的商户系统内部订单号。
	//	transaction_id和out_trade_no必须二选一进行传参。
	OutTradeNo string `json:"out_trade_no"`
	//【商户退款单号】 商户系统内部的退款单号，商户系统内部唯一，只能是数字、大小写字母_-|*@ ，同一商户退款单号多次请求只退一笔。不可超过64个字节数。
	OutRefundNo string `json:"out_refund_no"`
	//【退款原因】 若商户传了退款原因，该原因将在下发给用户的退款消息中显示，具体展示可参见退款通知UI示意图。
	//	请注意：
	//	1、该退款原因参数的长度不得超过80个字节；
	//	2、当订单退款金额小于等于1元且为部分退款时，退款原因将不会在消息中体现。
	Reason string `json:"reason"`
	//【退款结果回调url】 异步接收微信支付退款结果通知的回调地址，通知url必须为外网可访问的url，不能携带参数。
	//	如果传了该参数，则商户平台上配置的回调地址（商户平台-交易中心-退款管理-退款配置）将不会生效，优先回调当前传的这个地址。
	NotifyUrl string `json:"notify_url"`
	//【退款资金来源】 若传递此参数则使用对应的资金账户退款。
	//	可选取值：
	//	AVAILABLE: 仅对旧资金流商户适用(请参考旧资金流介绍区分)，传此枚举指定从可用余额账户出资，否则默认使用未结算资金退款。
	//	UNSETTLED: 仅对出行预付押金退款适用，指定从未结算资金出资。
	FundsAccount string `json:"funds_account"`
	//【金额信息】订单退款金额信息
	Amount *Amount `json:"amount"`
	//【退款商品】 请填写需要指定退款的商品信息，所指定的商品信息需要与下单时传入的单品列表goods_detail中的对应商品信息一致 ，
	//	如无需按照指定商品退款，本字段不填。
	GoodsDetail []*GoodsDetail `json:"goods_detail"`
}

func (r *Refunds) ToString() string {
	marshal, _ := json.Marshal(r)
	return string(marshal)
}

// RefundResp 应答参数
type RefundResp struct {
	//【微信支付退款单号】申请退款受理成功时，该笔退款单在微信支付侧生成的唯一标识。
	RefundId string `json:"refund_id"`
	//【商户退款单号】 商户申请退款时传的商户系统内部退款单号。
	OutRefundNo string `json:"out_refund_no"`
	//【微信支付订单号】微信支付侧订单的唯一标识。
	TransactionId string `json:"transaction_id"`
	//【商户订单号】 商户下单时传入的商户系统内部订单号。
	OutTradeNo string `json:"out_trade_no"`
	//【退款渠道】 订单退款渠道
	//	以下枚举：
	//	ORIGINAL: 原路退款
	//	BALANCE: 退回到余额
	//	OTHER_BALANCE: 原账户异常退到其他余额账户
	//	OTHER_BANKCARD: 原银行卡异常退到其他银行卡(发起异常退款成功后返回)
	Channel string `json:"channel"`
	//【退款入账账户】 取当前退款单的退款入账方，有以下几种情况：
	//	1）退回银行卡：{银行名称}{卡类型}{卡尾号}
	//	2）退回支付用户零钱:支付用户零钱
	//	3）退还商户:商户基本账户商户结算银行账户
	//	4）退回支付用户零钱通:支付用户零钱通
	//	5）退回支付用户银行电子账户:支付用户银行电子账户
	//	6）退回支付用户零花钱:支付用户零花钱
	//	7）退回用户经营账户:用户经营账户
	//	8）退回支付用户来华零钱包:支付用户来华零钱包
	//	9）退回企业支付商户:企业支付商户
	UserReceivedAccount string `json:"user_received_account"`
	//【退款成功时间】
	//	1、定义：退款成功的时间，该字段在退款状态status为SUCCESS（退款成功）时返回。
	//	2、格式：遵循rfc3339标准格式：yyyy-MM-DDTHH:mm:ss+TIMEZONE。
	//	yyyy-MM-DD 表示年月日；
	//	T 字符用于分隔日期和时间部分；
	//	HH:mm:ss 表示具体的时分秒；
	//	TIMEZONE 表示时区（例如，+08:00 对应东八区时间，即北京时间）。
	//	示例：2015-05-20T13:29:35+08:00 表示北京时间2015年5月20日13点29分35秒。
	SuccessTime time.Time `json:"success_time"`
	//【退款创建时间】
	//	1、定义：提交退款申请成功，微信受理退款申请单的时间。
	//	2、格式：遵循rfc3339标准格式：yyyy-MM-DDTHH:mm:ss+TIMEZONE。
	//	yyyy-MM-DD 表示年月日；
	//	T 字符用于分隔日期和时间部分；
	//	HH:mm:ss 表示具体的时分秒；
	//	TIMEZONE 表示时区（例如，+08:00 对应东八区时间，即北京时间）。
	//	示例：2015-05-20T13:29:35+08:00 表示北京时间2015年5月20日13点29分35秒。
	CreateTime time.Time `json:"create_time"`
	//【退款状态】退款单的退款处理状态。
	//	SUCCESS: 退款成功
	//	CLOSED: 退款关闭
	//	PROCESSING: 退款处理中
	//	ABNORMAL: 退款异常，退款到银行发现用户的卡作废或者冻结了，导致原路退款银行卡失败，可前往商户平台-交易中心，手动处理此笔退款，
	//	可参考： 退款异常的处理，或者通过发起异常退款接口进行处理。
	//	注：状态流转说明请参考状态流转图
	Status string `json:"status"`
	//【资金账户】 退款所使用资金对应的资金账户类型
	//	UNSETTLED: 未结算资金
	//	AVAILABLE: 可用余额
	//	UNAVAILABLE: 不可用余额
	//	OPERATION: 运营账户
	//	BASIC: 基本账户（含可用余额和不可用余额）
	//	ECNY_BASIC: 数字人民币基本账户
	FundsAccount string `json:"funds_account"`
	//【金额信息】订单退款金额信息
	Amount *Amount `json:"amount"`
	//【优惠退款详情】 订单各个代金券的退款详情，订单使用了代金券且代金券发生退款时返回。
	PromotionDetail []*PromotionDetail `json:"promotion_detail"`
}

// AbnormalRefund 发起异常退款
type AbnormalRefund struct {
	//【商户退款单号】商户申请退款时传入的商户系统内部退款单号。
	OutRefundNo string `json:"out_refund_no"`
	//【异常退款处理方式】 可选：退款至用户银行卡、退款至交易商户银行账户
	// 	可选取值
	// 	USER_BANK_CARD: 退款到用户银行卡
	// 	MERCHANT_BANK_CARD: 退款至交易商户银行账户
	Type string `json:"type"`
	//【开户银行】 银行类型，采用字符串类型的银行标识，值列表详见银行类型。
	//	仅支持招行、交通银行、农行、建行、工商、中行、平安、浦发、中信、光大、民生、兴业、广发、邮储、宁波银行的借记卡。
	//	若退款至用户此字段必填。
	BankType string `json:"bank_type"`
	//【收款银行卡号】用户的银行卡账号，该字段需要使用微信支付公钥加密（推荐），
	//	请参考获取微信支付公钥ID说明(https://pay.weixin.qq.com/doc/v3/merchant/4012153196)
	//	以及如何使用微信支付公钥加密敏感字段(https://pay.weixin.qq.com/doc/v3/merchant/4013053257)，
	//	也可以使用微信支付平台证书公钥加密，
	//	参考平台证书简介与使用说明、如何使用平台证书加密敏感字段(https://pay.weixin.qq.com/doc/v3/merchant/4012068814)
	//	若退款至用户此字段必填。
	BankAccount string `json:"bank_account"`
	//【收款用户姓名】 收款用户姓名，该字段需要使用微信支付公钥加密（推荐），
	//	请参考获取微信支付公钥ID说明(https://pay.weixin.qq.com/doc/v3/merchant/4012153196)
	//	以及如何使用微信支付公钥加密敏感字段(https://pay.weixin.qq.com/doc/v3/merchant/4013053257)，
	//	也可以使用微信支付平台证书公钥加密，
	//	参考平台证书简介与使用说明、如何使用平台证书加密敏感字段(https://pay.weixin.qq.com/doc/v3/merchant/4012068814)
	//	若退款至用户此字段必填。
	RealName string `json:"real_name"`
}

func (r *AbnormalRefund) ToString() string {
	marshal, _ := json.Marshal(r)
	return string(marshal)
}
