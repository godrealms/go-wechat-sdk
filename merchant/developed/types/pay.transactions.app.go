package types

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/godrealms/go-wechat-sdk/utils"
)

// Amount
// 【订单金额】订单金额信息
type Amount struct {
	//【总金额】 订单总金额，单位为分，整型。
	//	示例：1元应填写 100
	Total int `json:"total,omitempty"`
	//【货币类型】符合ISO 4217标准的三位字母代码，固定传：CNY，代表人民币。
	Currency string `json:"currency,omitempty"`
	//【用户支付金额】用户实际支付金额，整型，单位为分，用户支付金额=总金额-代金券金额。
	PayerTotal int `json:"payer_total,omitempty"`
	//【用户支付币种】 订单支付成功后固定返回：CNY，代表人民币。
	PayerCurrency string `json:"payer_currency,omitempty"`
	//【退款金额】 退款金额，币种的最小单位，只能为整数，不能超过原订单支付金额。
	Refund int `json:"refund,omitempty"`
	//【退款出资账户及金额】退款需从指定账户出资时，可传递该参数以指定出资金额（币种最小单位，仅限整数）。
	//	多账户出资退款需满足：1、未开通退款支出分离功能；2、订单为待分账或分账中的分账订单。
	//	传递参数需确保：1、基本账户可用与不可用余额之和等于退款金额；2、账户类型不重复。不符条件将返回错误。
	From []*RefundFrom `json:"from,omitempty"`
	//【用户退款金额】 指用户实际收到的现金退款金额，数据类型为整型，单位为分。
	//	例如在一个10元的订单中，用户使用了2元的全场代金券，若商户申请退款5元，则用户将收到4元的现金退款(即该字段所示金额)和1元的代金券退款。
	//	注：部分退款用户无法继续使用代金券，只有在订单全额退款且代金券未过期的情况下，且全场券属于银行立减金用户才能继续使用代金券。
	//	详情参考含优惠退款说明。
	PayerRefund int `json:"payer_refund,omitempty"`
	//【应结退款金额】 去掉免充值代金券退款金额后的退款金额，整型，单位为分，
	//	例如10元订单用户使用了2元全场代金券(一张免充值1元 + 一张预充值1元)，商户申请退款5元，则该金额为 退款金额5元 - 0.5元免充值代金券退款金额 = 4.5元。
	SettlementRefund int `json:"settlement_refund,omitempty"`
	//【应结订单金额】去除免充值代金券金额后的订单金额，整型，单位为分，
	//	例如10元订单用户使用了2元全场代金券(一张免充值1元 + 一张预充值1元)，则该金额为 订单金额10元 - 免充值代金券金额1元 = 9元。
	SettlementTotal int `json:"settlement_total,omitempty"`
	//【优惠退款金额】 申请退款后用户收到的代金券退款金额，整型，单位为分，
	//	例如10元订单用户使用了2元全场代金券，商户申请退款5元，用户收到的是4元现金 + 1元代金券退款金额(该字段) 。
	DiscountRefund int `json:"discount_refund,omitempty"`
	//【手续费退款金额】 订单退款时退还的手续费金额，整型，单位为分，
	//	例如一笔100元的订单收了0.6元手续费，商户申请退款50元，该金额为等比退还的0.3元手续费。
	RefundFee int `json:"refund_fee,omitempty"`
}

// GoodsDetail
// 【单品列表】 单品列表信息
//
//	条目个数限制：【1，6000】
type GoodsDetail struct {
	//【商户侧商品编码】 由半角的大小写字母、数字、中划线、下划线中的一种或几种组成。
	MerchantGoodsId string `json:"merchant_goods_id"`
	//【微信支付商品编码】 微信支付定义的统一商品编码（没有可不传）
	WechatpayGoodsId string `json:"wechatpay_goods_id,omitempty"`
	//【商品编码】 商品编码。
	GoodsID string `json:"goods_id,omitempty"`
	//【商品名称】 商品的实际名称
	GoodsName string `json:"goods_name,omitempty"`
	//【商品数量】 用户购买的数量
	Quantity int `json:"quantity"`
	//【商品单价】整型，单位为：分。
	//	如果商户有优惠，需传输商户优惠后的单价(例如：用户对一笔100元的订单使用了商场发的纸质优惠券100-50，则活动商品的单价应为原单价-50)
	UnitPrice int `json:"unit_price"`
	//【商品优惠金额】 商品优惠金额。
	DiscountAmount int `json:"discount_amount,omitempty"`
	//【商品备注】 商品备注。创券商户在商户平台创建单品券时，若设置了商品备注则会返回。
	GoodsRemark string `json:"goods_remark,omitempty"`
	//【商品退款金额】 商品退款金额，单位为分
	RefundAmount int `json:"refund_amount,omitempty"`
	//【商品退货数量】 对应商品的退货数量
	RefundQuantity int `json:"refund_quantity,omitempty"`
}

// Detail
// 【优惠功能】 优惠功能
type Detail struct {
	//【订单原价】
	//	1、商户侧一张小票订单可能被分多次支付，订单原价用于记录整张小票的交易金额。
	//	2、当订单原价与支付金额不相等，则不享受优惠。
	//	3、该字段主要用于防止同一张小票分多次支付，以享受多次优惠的情况，正常支付订单不必上传此参数。
	CostPrice int `json:"cost_price"`
	//【商品小票ID】 商家小票ID
	InvoiceId string `json:"invoice_id"`
	//【单品列表】 单品列表信息
	//  条目个数限制：【1，6000】
	GoodsDetail []*GoodsDetail `json:"goods_detail"`
}

// StoreInfo
// 【商户门店信息】 商户门店信息
type StoreInfo struct {
	//【门店编号】商户侧门店编号，总长度不超过32字符，由商户自定义。
	Id string `json:"id"`
	//【门店名称】 商户侧门店名称，由商户自定义。
	Name string `json:"name,omitempty"`
	//【地区编码】 地区编码，详细请见省市区编号对照表: https://pay.weixin.qq.com/doc/v3/merchant/4012076371
	AreaCode string `json:"area_code,omitempty"`
	//【详细地址】 详细的商户门店地址，由商户自定义。
	Address string `json:"address,omitempty"`
}

// SceneInfo
// 【场景信息】 场景信息
type SceneInfo struct {
	//【用户终端IP】 用户的客户端IP，支持IPv4和IPv6两种格式的IP地址。
	PayerClientIp string `json:"payer_client_ip,omitempty"`
	//【商户端设备号】 商户端设备号（门店号或收银设备ID）。
	DeviceId string `json:"device_id,omitempty"`
	//【商户门店信息】 商户门店信息
	StoreInfo *StoreInfo `json:"store_info,omitempty"`
}

// SettleInfo
// 【结算信息】 结算信息
type SettleInfo struct {
	//【分账标识】订单的分账标识在下单时设置，传入true表示在订单支付成功后可进行分账操作。以下是详细说明：
	//	需要分账（传入true）：
	//	订单收款成功后，资金将被冻结并转入基本账户的不可用余额。商户可通过请求分账API，将收款资金分配给其他商户或用户。
	//	完成分账操作后，可通过接口解冻剩余资金，或在支付成功30天后自动解冻。
	//
	//	不需要分账（传入false或不传，默认为false）：
	//	订单收款成功后，资金不会被冻结，而是直接转入基本账户的可用余额。
	ProfitSharing bool `json:"profit_sharing"`
}

// Transactions APP下单
type Transactions struct {
	//【公众账号ID】APPID是商户移动应用唯一标识，在开放平台(移动应用)申请。
	// 	此处需填写与mchid完成绑定的appid，详见：https://pay.weixin.qq.com/doc/v3/merchant/4013070756。
	Appid string `json:"appid"`
	//【商户号】是由微信支付系统生成并分配给每个商户的唯一标识符，商户号获取方式请参考商户模式开发必要参数说明。
	Mchid string `json:"mchid"`
	//【商品描述】商品信息描述，用户微信账单的商品字段中可见，商户需传递能真实代表商品信息的描述，不能超过127个字符。
	Description string `json:"description"`
	//【商户订单号】商户系统内部订单号，要求6-32个字符内，只能是数字、大小写字母_-|* 且在同一个商户号下唯一。
	OutTradeNo string `json:"out_trade_no"`
	//【支付结束时间】
	// 1、定义：支付结束时间是指用户能够完成该笔订单支付的最后时限，并非订单关闭的时间。超过此时间后，用户将无法对该笔订单进行支付。如需关闭订单，请调用关闭订单API接口。
	// 2、格式要求：支付结束时间需遵循rfc3339标准格式：yyyy-MM-DDTHH:mm:ss+TIMEZONE。yyyy-MM-DD 表示年月日；T 字符用于分隔日期和时间部分；HH:mm:ss 表示具体的时分秒；TIMEZONE 表示时区（例如，+08:00 对应东八区时间，即北京时间）。
	// 示例：2015-05-20T13:29:35+08:00 表示北京时间2015年5月20日13点29分35秒。
	// 3、注意事项：
	// time_expire 参数仅在用户首次下单时可设置，且不允许后续修改，尝试修改将导致错误。
	// 若用户实际进行支付的时间超过了订单设置的支付结束时间，商户需使用新的商户订单号下单，生成新的订单供用户进行支付。若未超过支付结束时间，则可使用原参数重新请求下单接口，以获取当前订单最新的prepay_id 进行支付。
	// 支付结束时间不能早于下单时间后1分钟，若设置的支付结束时间早于该时间，系统将自动调整为下单时间后1分钟作为支付结束时间。
	TimeExpire string `json:"time_expire,omitempty"`
	//【附加数据】附加数据，在查询API和支付通知中原样返回，可作为自定义参数使用，长度限制为1,128个字符。
	Attach string `json:"attach,omitempty"`
	//【商户回调地址】商户接收支付成功回调通知的地址，需按照notify_url填写注意事项规范填写。
	NotifyUrl string `json:"notify_url"`
	//【订单优惠标记】代金券在创建时可以配置多个订单优惠标记，标记的内容由创券商户自定义设置。
	//	详细参考：创建代金券批次API: https://pay.weixin.qq.com/doc/v3/merchant/4012534633。
	//	如果代金券有配置订单优惠标记，则必须在该参数传任意一个配置的订单优惠标记才能使用券。
	//	如果代金券没有配置订单优惠标记，则可以不传该参数。
	//	示例：
	//	如有两个活动，活动A设置了两个优惠标记：WXG1、WXG2；活动B设置了两个优惠标记：WXG1、WXG3；
	//	下单时优惠标记传WXG2，则订单参与活动A的优惠；
	//	下单时优惠标记传WXG3，则订单参与活动B的优惠；
	//	下单时优惠标记传共同的WXG1，则订单参与活动A、B两个活动的优惠；
	GoodsTag string `json:"goods_tag,omitempty"`
	//【电子发票入口开放标识】 传入true时，支付成功消息和支付详情页将出现开票入口。
	//	需要在微信支付商户平台或微信公众平台开通电子发票功能，传此字段才可生效。
	//	详细参考：电子发票介绍: https://pay.weixin.qq.com/doc/v3/merchant/4012064743
	//	true：是
	//	false：否
	SupportFapiao bool `json:"support_fapiao"`
	//【订单金额】订单金额信息
	Amount *Amount `json:"amount"`
	//【优惠功能】 优惠功能
	Detail *Detail `json:"detail,omitempty"`
	//【场景信息】 场景信息
	SceneInfo *SceneInfo `json:"scene_info,omitempty"`
	//【结算信息】 结算信息
	SettleInfo *SettleInfo `json:"settle_info,omitempty"`
}

func (a *Transactions) ToString() string {
	marshal, _ := json.Marshal(a)
	return string(marshal)
}

type TransactionsAppResponse struct {
	//【预支付交易会话标识】预支付交易会话标识，APP调起支付时需要使用的参数，
	//	有效期为2小时，失效后需要重新请求该接口以获取新的prepay_id。
	PrepayId string `json:"prepay_id"`
}

// ModifyAppResponse APP调起支付
type ModifyAppResponse struct {
	// 填写下单时传入的【公众账号ID】appid。
	AppId string `json:"app_id"`
	// 填写下单时传入的【商户号】mchid。
	PartnerId string `json:"partner_id"`
	// 预支付交易会话标识。APP下单接口返回的prepay_id，
	//	该值有效期为2小时，超过有效期需要重新请求APP下单接口以获取新的prepay_id。
	PrepayId string `json:"prepay_id"`
	// 填写固定值Sign=WXPay
	PackageValue string `json:"package_value"`
	// 随机字符串，不长于32位。该值建议使用随机数算法生成。
	NonceStr string `json:"nonce_str"`
	// Unix时间戳，是从1970年1月1日（UTC/GMT的午夜）开始所经过的秒数。
	// 	注意：常见时间戳为秒级或毫秒级，该处必需传秒级时间戳。
	TimeStamp string `json:"timestamp"`
	// 签名，使用字段appId、timeStamp、nonceStr、prepayId以及商户API证书私钥生成的RSA签名值，
	//	详细参考: https://pay.weixin.qq.com/doc/v3/merchant/4012365340。
	Sign string `json:"sign"`
}

func (r *ModifyAppResponse) GenerateSignature(privateKey *rsa.PrivateKey) error {
	str := fmt.Sprintf("%s\n%s\n%s\n%s\n", r.AppId, r.TimeStamp, r.NonceStr, r.PrepayId)
	sign, err := utils.SignSHA256WithRSA(str, privateKey)
	if err != nil {
		return err
	}
	r.Sign = sign
	return nil
}
