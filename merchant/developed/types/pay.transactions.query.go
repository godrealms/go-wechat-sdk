package types

import "time"

type TradeType string

const (
	TradeTypeJSAPI    TradeType = "JSAPI"    // 公众号支付、小程序支付
	TradeTypeNATIVE   TradeType = "NATIVE"   // Native支付
	TradeTypeAPP      TradeType = "APP"      // APP支付
	TradeTypeMICROPAY TradeType = "MICROPAY" // 付款码支付
	TradeTypeMWEB     TradeType = "MWEB"     // H5支付
	TradeTypeFACEPAY  TradeType = "FACEPAY"  // 刷脸支付
)

type TradeState string

const (
	TradeStateSUCCESS    TradeState = "SUCCESS"    // 支付成功
	TradeStateREFUND     TradeState = "REFUND"     // 转入退款
	TradeStateNOTPAY     TradeState = "NOTPAY"     // 未支付
	TradeStateCLOSED     TradeState = "CLOSED"     // 已关闭
	TradeStateREVOKED    TradeState = "REVOKED"    // 已撤销（仅付款码支付会返回）
	TradeStateUSERPAYING TradeState = "USERPAYING" // 用户支付中（仅付款码支付会返回）
	TradeStatePAYERROR   TradeState = "PAYERROR"   // 支付失败（仅付款码支付会返回）
)

// Payer
// 【支付者信息】 订单的支付者信息。
type Payer struct {
	//【用户标识】用户在商户下单的appid下唯一标识。
	Openid string `json:"openid"`
}

// PromotionDetail
// 【优惠功能】 代金券信息，当订单有使用代金券时，该字段将返回所使用的代金券信息。
type PromotionDetail struct {
	//【券ID】 代金券id，微信为代金券分配的唯一标识，创券商户调用发放指定批次的代金券时返回的代金券ID coupon_id。
	CouponId string `json:"coupon_id,omitempty"`
	//【优惠名称】 优惠名称，创券商户创建代金券批次时传入的批次名称stock_name。
	Name string `json:"name,omitempty"`
	//【优惠范围】优惠活动中代金券的适用范围，分为两种类型：
	//	1、GLOBAL：全场代金券-以订单整体可优惠的金额为优惠门槛的代金券；
	//	2、SINGLE：单品优惠-以订单中具体某个单品的总金额为优惠门槛的代金券
	Scope string `json:"scope,omitempty"`
	//【优惠类型】代金券资金类型，优惠活动中代金券的结算资金类型，分为两种类型：
	//	1、CASH：预充值-带有结算资金的代金券，会随订单结算给订单收款商户；
	//	2、NOCASH：免充值-不带有结算资金的代金券，无资金结算给订单收款商户。
	Type string `json:"type,omitempty"`
	//【优惠券面额】代金券优惠的金额。
	Amount int `json:"amount,omitempty"`
	//【活动ID】单张代金券所对应的批次号
	StockId string `json:"stock_id,omitempty"`
	//【微信出资】 代金券有三种出资类型：微信出资、商户出资和其他出资。本参数将返回选择“微信出资类型”时的优惠券面额。
	//	1、创建代金券后默认为商户出资类型。如需使用其他两种类型，请与相关行业运营进行沟通。
	//	2、在 wechatpay_contribute、merchant_contribute 和 other_contribute 这三个字段中，仅有一个字段会返回出资金额。
	//	具体返回哪个字段取决于代金券批次的配置。
	WechatpayContribute int `json:"wechatpay_contribute,omitempty"`
	//【商户出资】代金券有三种出资类型：微信出资、商户出资和其他出资。本参数将返回选择“商户出资类型”时的优惠券面额。
	//	1、创建代金券后默认为商户出资类型。如需使用其他两种类型，请与相关行业运营进行沟通。
	//	2、在 wechatpay_contribute、merchant_contribute 和 other_contribute 这三个字段中，仅有一个字段会返回出资金额。
	//	具体返回哪个字段取决于代金券批次的配置。
	MerchantContribute int `json:"merchant_contribute,omitempty"`
	//【其他出资】代金券有三种出资类型：微信出资、商户出资和其他出资。本参数将返回选择“其他出资类型”时的优惠券面额。
	//	1、创建代金券后默认为商户出资类型。如需使用其他两种类型，请与相关行业运营进行沟通。
	//	2、在 wechatpay_contribute、merchant_contribute 和 other_contribute 这三个字段中，仅有一个字段会返回出资金额。
	//	具体返回哪个字段取决于代金券批次的配置。
	OtherContribute int `json:"other_contribute,omitempty"`
	//【优惠币种】 代金券金额所对应的货币种类：固定为：CNY，人民币。
	Currency string `json:"currency,omitempty"`
	//【单品列表】 单品列表。scope为SINGLE（单品优惠）时返回该参数
	GoodsDetail  []*GoodsDetail `json:"goods_detail,omitempty"`
	PromotionId  string         `json:"promotion_id,omitempty"`
	RefundAmount int            `json:"refund_amount,omitempty"`
}

// QueryResponse 微信支付订单号查询订单
type QueryResponse struct {
	//【公众账号ID】商户下单时传入的公众账号ID。
	Appid string `json:"appid"`
	//【商户号】商户下单时传入的商户号。
	Mchid string `json:"mchid"`
	//【商户订单号】商户下单时传入的商户系统内部订单号。
	OutTradeNo string `json:"out_trade_no"`
	//【微信支付订单号】 微信支付侧订单的唯一标识。
	TransactionId string `json:"transaction_id"`
	//【交易类型】 返回当前订单的交易类型，枚举值：
	// 	JSAPI：公众号支付、小程序支付
	// 	NATIVE：Native支付
	// 	APP：APP支付
	// 	MICROPAY：付款码支付
	// 	MWEB：H5支付
	// 	FACEPAY：刷脸支付
	TradeType TradeType `json:"trade_type"`
	//【交易状态】 返回订单当前交易状态。详细业务流转状态处理请参考开发指引-订单状态流转图。枚举值：
	//	SUCCESS：支付成功
	//	REFUND：转入退款
	//	NOTPAY：未支付
	//	CLOSED：已关闭
	//	REVOKED：已撤销（仅付款码支付会返回）
	//	USERPAYING：用户支付中（仅付款码支付会返回）
	//	PAYERROR：支付失败（仅付款码支付会返回）
	TradeState TradeState `json:"trade_state"`
	//【交易状态描述】 对交易状态的详细说明。
	TradeStateDesc string `json:"trade_state_desc"`
	//【银行类型】 用户支付方式说明，订单支付成功后返回，
	//	格式为银行简码_具体类型(DEBIT借记卡/CREDIT信用卡/ECNY数字人民币)，
	//	例如ICBC_DEBIT代表工商银行借记卡，
	//	非银行卡支付类型(例如余额/零钱通等)统一为OTHERS，
	//	具体请参考《银行类型对照表》。
	BankType string `json:"bank_type"`
	//【商户数据包】商户下单时传入的自定义数据包，用户不可见，长度不超过128字符，
	//	若下单传入该参数，则订单支付成功后此接口和支付成功回调通知以及交易账单中会原样返回；
	//	若下单未传该参数，则不会返回。
	Attach string `json:"attach"`
	//【支付完成时间】
	//	1、定义：用户完成订单支付的时间。该参数在订单支付成功后返回。
	//	2、格式：遵循rfc3339标准格式：yyyy-MM-DDTHH:mm:ss+TIMEZONE。
	//	yyyy-MM-DD 表示年月日；
	//	T 字符用于分隔日期和时间部分；
	//	HH:mm:ss 表示具体的时分秒；
	//	TIMEZONE 表示时区（例如，+08:00 对应东八区时间，即北京时间）。
	//	示例：2015-05-20T13:29:35+08:00 表示北京时间2015年5月20日13点29分35秒。
	SuccessTime time.Time `json:"success_time"`
	//【支付者信息】 订单的支付者信息。
	Payer *Payer `json:"payer"`
	//【订单金额】 订单金额信息。
	Amount *Amount `json:"amount"`
	//【场景信息】 下单时传入的支付场景描述，若下单传入该参数，则原样返回；若下单未传该参数，则不会返回。
	SceneInfo *SceneInfo `json:"scene_info"`
	//【优惠功能】 代金券信息，当订单有使用代金券时，该字段将返回所使用的代金券信息。
	PromotionDetail []*PromotionDetail `json:"promotion_detail"`
}
