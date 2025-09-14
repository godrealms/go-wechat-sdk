package offiaccount

// UserField 授权页个人发票字段
type UserField struct {
	ShowTitle    int64          `json:"show_title,omitempty"`    // 是否填写抬头，0为否，1为是
	ShowPhone    int64          `json:"show_phone,omitempty"`    // 是否填写电话号码，0为否，1为是
	ShowEmail    int64          `json:"show_email,omitempty"`    // 是否填写邮箱，0为否，1为是
	RequirePhone int64          `json:"require_phone,omitempty"` // 电话号码是否必填,0为否，1为是
	RequireEmail int64          `json:"require_email,omitempty"` // 邮箱是否必填，0位否，1为是
	CustomField  []*CustomField `json:"custom_field,omitempty"`  // 自定义字段
}

// BizField 授权页单位发票字段
type BizField struct {
	ShowTitle       int64          `json:"show_title,omitempty"`        // 是否填写抬头，0为否，1为是
	ShowTaxNo       int64          `json:"show_tax_no,omitempty"`       // 是否填写税号，0为否，1为是
	ShowAddr        int64          `json:"show_addr,omitempty"`         // 是否填写单位地址，0为否，1为是
	ShowPhone       int64          `json:"show_phone,omitempty"`        // 是否填写电话号码，0为否，1为是
	ShowBankType    int64          `json:"show_bank_type,omitempty"`    // 是否填写开户银行，0为否，1为是
	ShowBankNo      int64          `json:"show_bank_no,omitempty"`      // 是否填写银行账号，0为否，1为是
	RequireTaxNo    int64          `json:"require_tax_no,omitempty"`    // 税号是否必填，0为否，1为是
	RequireAddr     int64          `json:"require_addr,omitempty"`      // 单位地址是否必填，0为否，1为是
	RequirePhone    int64          `json:"require_phone,omitempty"`     // 电话号码是否必填，0为否，1为是
	RequireBankType int64          `json:"require_bank_type,omitempty"` // 开户行是否必填，0为否，1为是
	RequireBankNo   int64          `json:"require_bank_no,omitempty"`   // 银行账号是否必填，0为否，1为是
	CustomField     []*CustomField `json:"custom_field,omitempty"`      // 自定义字段
}

// CustomField 自定义字段
type CustomField struct {
	Key       string `json:"key"`                  // 字段名
	IsRequire int64  `json:"is_require,omitempty"` // 0：否，1：是， 默认为0
	Notice    string `json:"notice,omitempty"`     // 提示文案
}

// AuthField 授权页字段
type AuthField struct {
	UserField *UserField `json:"user_field"` // 授权页个人发票字段
	BizField  *BizField  `json:"biz_field"`  // 授权页单位发票字段
}

// PayMchInfo 微信商户号与开票平台关系信息
type PayMchInfo struct {
	MchID   string `json:"mchid"`    // 微信支付商户号
	SPAppID string `json:"s_pappid"` // 为该商户提供开票服务的开票平台 id ，由开票平台提供给商户
}

// Contact 联系方式信息
type Contact struct {
	Timeout int64  `json:"time_out"` // 开票超时时间
	Phone   string `json:"phone"`    // 联系电话
}

// SetBizAttrRequest 设置授权页与商户信息请求参数
type SetBizAttrRequest struct {
	AuthField  *AuthField  `json:"auth_field,omitempty"`  // 授权页字段（set_auth_field使用）
	PayMchInfo *PayMchInfo `json:"paymch_info,omitempty"` // 微信商户号与开票平台关系信息（set_pay_mch使用）
	Contact    *Contact    `json:"contact,omitempty"`     // 联系方式信息(set_contact使用）
}

// GetAuthDataRequest 获取授权页数据请求参数
type GetAuthDataRequest struct {
	OrderID string `json:"order_id"` // 订单order_id
	SPAppID string `json:"s_pappid"` // 财政局id
}

// GetAuthDataResult 获取授权页数据结果
type GetAuthDataResult struct {
	Resp
	InvoiceStatus string `json:"invoice_status"` // 发票状态
	AuthTime      int64  `json:"auth_time"`      // 授权时间戳
}

// GetAuthUrlRequest 获取授权页链接请求参数
type GetAuthUrlRequest struct {
	SPAppID     string `json:"s_pappid"`               // 开票平台在微信的标识号，商户需要找开票平台提供
	OrderID     string `json:"order_id"`               // 订单id，在商户内单笔开票请求的唯一识别号，
	Money       int64  `json:"money"`                  // 订单金额，以分为单位
	Timestamp   int64  `json:"timestamp"`              // 时间戳
	Source      string `json:"source"`                 // 开票来源，app：app开票，web：微信h5开票，wxa：小程序开发票，wap：普通网页开票
	RedirectUrl string `json:"redirect_url,omitempty"` // 授权成功后跳转页面。本字段只有在source为H5的时候需要填写，引导用户在微信中进行下一步流程。app开票因为从外部app拉起微信授权页，授权完成后自动回到原来的app，故无需填写。
	Ticket      string `json:"ticket"`                 // 授权页ticket
	Type        int64  `json:"type"`                   // 授权类型，0：开票授权，1：填写字段开票授权，2：领票授权
}

// GetAuthUrlResult 获取授权页链接结果
type GetAuthUrlResult struct {
	Resp
	AuthUrl string `json:"auth_url"` // 授权链接
	AppID   string `json:"appid"`    // source为wxa时才有
}

// RejectInsertRequest 拒绝领取发票请求参数
type RejectInsertRequest struct {
	SPAppID string `json:"s_pappid"`      // 开票平台在微信上的标识，由开票平台告知商户
	OrderID string `json:"order_id"`      // 订单 id
	Reason  string `json:"reason"`        // 商家解释拒绝开票的原因，如重复开票，抬头无效、已退货无法开票等
	Url     string `json:"url,omitempty"` // 跳转链接，引导用户进行下一步处理，如重新发起开票、重新填写抬头、展示订单情况等
}

// SetUrlResult 设置商户联系方式结果
type SetUrlResult struct {
	Resp
	InvoiceUrl string `json:"invoice_url"` // 该开票平台专用的授权链接。开票平台须将 invoice_url 内的 s_pappid 给到服务的商户，商户在请求授权链接时会向微信传入该参数，标识所使用的开票平台是哪家
}

// GetPdfRequest 获取pdf文件请求参数
type GetPdfRequest struct {
	Action   string `json:"action"`     // 填"get_url"
	SMediaID string `json:"s_media_id"` // 发票s_media_id
}

// GetPdfResult 获取pdf文件结果
type GetPdfResult struct {
	Resp
	PdfUrl           string `json:"pdf_url"`             // pdf 的 url ，两个小时有效期
	PdfUrlExpireTime int64  `json:"pdf_url_expire_time"` // pdf_url 过期时间， 7200 秒
}

// UpdateInvoiceStatusRequest 更新发票状态请求参数
type UpdateInvoiceStatusRequest struct {
	CardID          string `json:"card_id"`          // 发票 id
	Code            string `json:"code"`             // 发票 code
	ReimburseStatus string `json:"reimburse_status"` // 发票报销状态
}

// SetPdfResult 设置pdf文件结果
type SetPdfResult struct {
	Resp
	SMediaID string `json:"s_media_id"` // 64位整数，在 将发票卡券插入用户卡包 时使用用于关联pdf和发票卡券，s_media_id有效期有3天，3天内若未将s_media_id关联到发票卡券，pdf将自动销毁
}

// CreateCardInvoiceInfo 创建发票卡券模板信息
type CreateCardInvoiceInfo struct {
	BaseInfo *struct {
		LogoUrl              string `json:"logo_url"`                          // 发票商家 LOGO，请使用永久素材接口
		Title                string `json:"title"`                             // 收款方（显示在列表），上限为 9 个汉字，建议填入商户简称
		CustomUrlName        string `json:"custom_url_name,omitempty"`         // 开票平台自定义入口名称，与 custom_url 字段共同使用，长度限制在 5 个汉字内
		CustomUrl            string `json:"custom_url,omitempty"`              // 开票平台自定义入口跳转外链的地址链接 , 发票外跳的链接会带有发票参数，用于标识是从哪张发票跳出的链接
		CustomUrlSubTitle    string `json:"custom_url_sub_title,omitempty"`    // 显示在入口右侧的 tips ，长度限制在 6 个汉字内
		PromotionUrlName     string `json:"promotion_url_name,omitempty"`      // 营销场景的自定义入口
		PromotionUrl         string `json:"promotion_url,omitempty"`           // 入口跳转外链的地址链接，发票外跳的链接会带有发票参数，用于标识是从那张发票跳出的链接
		PromotionUrlSubTitle string `json:"promotion_url_sub_title,omitempty"` // 显示在入口右侧的 tips ，长度限制在 6 个汉字内
	} `json:"base_info"` // 发票卡券模板基础信息
	Payee string `json:"payee"` // 收款方（开票方）全称，显示在发票详情内。故建议一个收款方对应一个发票卡券模板
	Type  string `json:"type"`  // 发票类型
}

// CreateCardRequest 创建发票卡券模板请求参数
type CreateCardRequest struct {
	InvoiceInfo *CreateCardInvoiceInfo `json:"invoice_info"` // 发票模板对象
}

// CreateCardResult 创建发票卡券模板结果
type CreateCardResult struct {
	Resp
	CardID string `json:"card_id"` // 当错误码为 0 时，返回发票卡券模板的编号，用于后续该商户发票生成后，作为必填参数在调用插卡接口时传入
}

// InvoiceItemInfo 商品详情结构
type InvoiceItemInfo struct {
	Name  string `json:"name"`           // 项目的名称
	Num   int64  `json:"num,omitempty"`  // 项目的数量
	Unit  string `json:"unit,omitempty"` // 项目的单位，如个
	Price int64  `json:"price"`          // 项目的单价
}

// InvoiceUserData 用户信息
type InvoiceUserData struct {
	Fee                   int64              `json:"fee"`                                // 发票的金额，以分为单位
	Title                 string             `json:"title"`                              // 发票的抬头
	BillingTime           int64              `json:"billing_time"`                       // 发票的开票时间，为10位时间戳（utc+8）
	BillingNo             string             `json:"billing_no"`                         // 发票的发票号码；数电发票传20位发票号码
	BillingCode           string             `json:"billing_code"`                       // 发票的发票代码；数电发票发票代码为空
	Info                  []*InvoiceItemInfo `json:"info,omitempty"`                     // 商品详情结构
	FeeWithoutTax         int64              `json:"fee_without_tax"`                    // 不含税金额，以分为单位
	Tax                   int64              `json:"tax"`                                // 税额，以分为单位
	SPdfMediaID           string             `json:"s_pdf_media_id"`                     // 发票pdf文件上传到微信发票平台后，会生成一个发票s_media_id，该s_media_id可以直接用于关联发票PDF和发票卡券。
	STripPdfMediaID       string             `json:"s_trip_pdf_media_id,omitempty"`      // 其它消费附件的PDF，如行程单、水单等
	CheckCode             string             `json:"check_code"`                         // 校验码，发票pdf右上角，开票日期下的校验码；数电发票发票校验码为空
	BuyerNumber           string             `json:"buyer_number,omitempty"`             // 购买方纳税人识别号
	BuyerAddressAndPhone  string             `json:"buyer_address_and_phone,omitempty"`  // 购买方地址、电话
	BuyerBankAccount      string             `json:"buyer_bank_account,omitempty"`       // 购买方开户行及账号
	SellerNumber          string             `json:"seller_number,omitempty"`            // 销售方纳税人识别号
	SellerAddressAndPhone string             `json:"seller_address_and_phone,omitempty"` // 销售方地址、电话
	SellerBankAccount     string             `json:"seller_bank_account,omitempty"`      // 销售方开户行及账号
	Remarks               string             `json:"remarks,omitempty"`                  // 备注，发票右下角初
	Cashier               string             `json:"cashier,omitempty"`                  // 收款人，发票左下角处
	Maker                 string             `json:"maker,omitempty"`                    // 开票人，发票下方处
}

// CardExt 发票具体内容
type CardExt struct {
	NonceStr string `json:"nonce_str"` // 随机字符串，防止重复
	UserCard *struct {
		InvoiceUserData *InvoiceUserData `json:"invoice_user_data"` // 用户信息
	} `json:"user_card"` // 用户信息结构体
}

// InsertInvoiceRequest 将电子发票卡券插入用户卡包请求参数
type InsertInvoiceRequest struct {
	OrderID string   `json:"order_id"` // 发票order_id，既商户给用户授权开票的订单号
	CardID  string   `json:"card_id"`  // 发票card_id
	AppID   string   `json:"appid"`    // 该订单号授权时使用的appid，一般为商户appid
	CardExt *CardExt `json:"card_ext"` // 发票具体内容
}

// InsertInvoiceResult 将电子发票卡券插入用户卡包结果
type InsertInvoiceResult struct {
	Resp
	Code    string `json:"code"`    // 发票code
	OpenID  string `json:"openid"`  // 获得发票用户的openid
	UnionID string `json:"unionid"` // 只有在用户将公众号绑定到微信开放平台账号后，才会出现该字段
}

// GetInvoiceInfoRequest 查询报销发票信息请求参数
type GetInvoiceInfoRequest struct {
	CardID      string `json:"card_id"`      // 发票id
	EncryptCode string `json:"encrypt_code"` // 发票卡券的加密code，和card_id共同构成一张发票卡券的唯一标识
}

// InvoiceUserInfo 用户可在发票票面看到的主要信息
type InvoiceUserInfo struct {
	Fee                   int64              `json:"fee"`                                // 发票的金额，以分为单位
	Title                 string             `json:"title"`                              // 发票的抬头
	BillingTime           int64              `json:"billing_time"`                       // 发票的开票时间，为10位时间戳（utc+8）
	BillingNo             string             `json:"billing_no"`                         // 发票的发票号码；数电发票传20位发票号码
	BillingCode           string             `json:"billing_code"`                       // 发票的发票代码；数电发票发票代码为空
	Info                  []*InvoiceItemInfo `json:"info,omitempty"`                     // 商品详情结构
	FeeWithoutTax         int64              `json:"fee_without_tax"`                    // 不含税金额，以分为单位
	Tax                   int64              `json:"tax"`                                // 税额，以分为单位
	PdfUrl                string             `json:"pdf_url"`                            // 这张发票对应的PDF_URL
	TripPdfUrl            string             `json:"trip_pdf_ur"`                        // 其它消费凭证附件对应的URL，如行程单、水单等
	CheckCode             string             `json:"check_code"`                         // 校验码，发票pdf右上角，开票日期下的校验码；数电发票发票校验码为空
	BuyerNumber           string             `json:"buyer_number,omitempty"`             // 购买方纳税人识别号
	BuyerAddressAndPhone  string             `json:"buyer_address_and_phone,omitempty"`  // 购买方地址、电话
	BuyerBankAccount      string             `json:"buyer_bank_account,omitempty"`       // 购买方开户行及账号
	SellerNumber          string             `json:"seller_number,omitempty"`            // 销售方纳税人识别号
	SellerAddressAndPhone string             `json:"seller_address_and_phone,omitempty"` // 销售方地址、电话
	SellerBankAccount     string             `json:"seller_bank_account,omitempty"`      // 销售方开户行及账号
	Remarks               string             `json:"remarks,omitempty"`                  // 备注，发票右下角初
	Cashier               string             `json:"cashier,omitempty"`                  // 收款人，发票左下角处
	Maker                 string             `json:"maker,omitempty"`                    // 开票人，发票下方处
	ReimburseStatus       string             `json:"reimburse_status"`                   // 发票报销状态
	OrderID               string             `json:"order_id,omitempty"`
}

// GetInvoiceInfoResult 查询报销发票信息结果
type GetInvoiceInfoResult struct {
	Resp
	CardID    string           `json:"card_id"`    // 发票id
	BeginTime int64            `json:"begin_time"` // 发票的有效期起始时间
	EndTime   int64            `json:"end_time"`   // 发票的有效期截止时间
	OpenID    string           `json:"openid"`     // 用户标识
	Type      string           `json:"type"`       // 发票的类型，如广东增值税普通发票
	Payee     string           `json:"payee"`      // 发票的收款方
	Detail    string           `json:"detail"`     // 发票详情
	UserInfo  *InvoiceUserInfo `json:"user_info"`  // 用户可在发票票面看到的主要信息
}

// UpdateInvoiceReimburseStatusRequest 报销方更新发票状态请求参数
type UpdateInvoiceReimburseStatusRequest struct {
	CardID          string `json:"card_id"`          // 发票卡券的card_id
	EncryptCode     string `json:"encrypt_code"`     // 发票卡券的加密code，和card_id共同构成一张发票卡券的唯一标识
	ReimburseStatus string `json:"reimburse_status"` // 发票报销状态
}

// InvoiceListItem 发票列表项
type InvoiceListItem struct {
	CardID      string `json:"card_id"`      // 发票卡券的card_id
	EncryptCode string `json:"encrypt_code"` // 发票卡券的加密code，和card_id共同构成一张发票卡券的唯一标识
}

// UpdateInvoiceReimburseStatusBatchRequest 批量更新报销发票状态请求参数
type UpdateInvoiceReimburseStatusBatchRequest struct {
	OpenID          string             `json:"openid"`           // 用户openid
	ReimburseStatus string             `json:"reimburse_status"` // 发票报销状态
	InvoiceList     []*InvoiceListItem `json:"invoice_list"`     // 发票列表
}

// GetInvoiceBatchRequest 批量获取报销发票信息请求参数
type GetInvoiceBatchRequest struct {
	ItemList []*InvoiceListItem `json:"item_list"` // 发票列表
}

// InvoiceBatchItem 批量获取发票信息项
type InvoiceBatchItem struct {
	CardID    string           `json:"card_id"`    // 发票id
	BeginTime int64            `json:"begin_time"` // 发票的有效期起始时间
	EndTime   int64            `json:"end_time"`   // 发票的有效期截止时间
	OpenID    string           `json:"openid"`     // 用户标识
	Type      string           `json:"type"`       // 发票的类型，如广东增值税普通发票
	Payee     string           `json:"payee"`      // 发票的收款方
	Detail    string           `json:"detail"`     // 发票详情
	UserInfo  *InvoiceUserInfo `json:"user_info"`  // 用户可在发票票面看到的主要信息
}

// GetInvoiceBatchResult 批量获取报销发票信息结果
type GetInvoiceBatchResult struct {
	Resp
	ItemList []*InvoiceBatchItem `json:"item_list"` // 发票信息
}

// GetUserTitleUrlRequest 获取添加发票链接请求参数
type GetUserTitleUrlRequest struct {
	Title      string `json:"title,omitempty"`        // 发票抬头，user_fill为0时必填
	Phone      string `json:"phone,omitempty"`        // 誁系方式，必须是数字或"-"
	TaxNo      string `json:"tax_no,omitempty"`       // 税号，必须是15-20位数字或者英文字母
	Addr       string `json:"addr,omitempty"`         // 地址
	BankType   string `json:"bank_type,omitempty"`    // 银行类型
	BankNo     string `json:"bank_no,omitempty"`      // 银行号码
	UserFill   int64  `json:"user_fill,omitempty"`    // 0:企业设置的抬头，1:用户自己填写抬头
	OutTitleID string `json:"out_title_id,omitempty"` // 开票码
}

// GetUserTitleUrlResult 获取添加发票链接结果
type GetUserTitleUrlResult struct {
	Resp
	Url string `json:"url"` // 用户确认链接
}

// GetSelectTitleUrlRequest 获取选择发票抬头链接请求参数
type GetSelectTitleUrlRequest struct {
	Attach  string `json:"attach,omitempty"`   // 附加字段，用户提交发票时会发给商户
	BizName string `json:"biz_name,omitempty"` // 将商户名称显示给用户看
}

// GetSelectTitleUrlResult 获取选择发票抬头链接结果
type GetSelectTitleUrlResult struct {
	Resp
	Url string `json:"url"` // 专属抬头链接
}

// ScanTitleRequest 扫描二维码获取抬头请求参数
type ScanTitleRequest struct {
	ScanText string `json:"scan_text"` // 扫码获取的原始数据
}

// ScanTitleResult 扫描二维码获取抬头结果
type ScanTitleResult struct {
	Resp
	TitleType int64  `json:"title_type"` // 0-单位抬头，1-个人抬头
	Title     string `json:"title"`      // 发票抬头
	Phone     string `json:"phone"`      // 联系方式
	TaxNo     string `json:"tax_no"`     // 税号
	Addr      string `json:"addr"`       // 地址
	BankType  string `json:"bank_type"`  // 银行类型
	BankNo    string `json:"bank_no"`    // 银行号码
}

// GetFiscalAuthDataRequest 查询财政电子票据授权信息请求参数
type GetFiscalAuthDataRequest struct {
	OrderID string `json:"order_id"` // 订单order_id
	SPAppID string `json:"s_pappid"` // 财政局id
}

// GetFiscalAuthDataResult 查询财政电子票据授权信息结果
type GetFiscalAuthDataResult struct {
	Resp
	InvoiceStatus string `json:"invoice_status"` // 发票状态
	AuthTime      int64  `json:"auth_time"`      // 授权时间戳
}

// GetTicketResult 获取sdk临时票据结果
type GetTicketResult struct {
	Resp
	Ticket    string `json:"ticket"`     // 临时票据
	ExpiresIn int64  `json:"expires_in"` // 有效期（秒）
}

// RejectInsertFiscalRequest 拒绝开票请求参数
type RejectInsertFiscalRequest struct {
	SPAppID string `json:"s_pappid"`      // 开票平台在微信上的标识，由开票平台告知商户
	OrderID string `json:"order_id"`      // 订单 id
	Reason  string `json:"reason"`        // 商家解释拒绝开票的原因，如重复开票，抬头无效、已退货无法开票等
	Url     string `json:"url,omitempty"` // 跳转链接，引导用户进行下一步处理，如重新发起开票、重新填写抬头、展示订单情况等
}

// SetInvoiceUrlResult 设置微信授权页链接结果
type SetInvoiceUrlResult struct {
	Resp
	InvoiceUrl string `json:"invoice_url"` // 该开票平台专用的授权链接。开票平台须将 invoice_url 内的 s_pappid 给到服务的商户，商户在请求授权链接时会向微信传入该参数，标识所使用的开票平台是哪家
}

// GetPlatformPdfRequest 查询已上传的PDF文件请求参数
type GetPlatformPdfRequest struct {
	Action   string `json:"action"`     // 填"get_url"
	SMediaID string `json:"s_media_id"` // 发票s_media_id
}

// GetPlatformPdfResult 查询已上传的PDF文件结果
type GetPlatformPdfResult struct {
	Resp
	PdfUrl           string `json:"pdf_url"`             // pdf 的 url ，两个小时有效期
	PdfUrlExpireTime int64  `json:"pdf_url_expire_time"` // pdf_url 过期时间， 7200 秒
}

// UpdateInvoicePlatformStatusRequest 更新发票状态请求参数
type UpdateInvoicePlatformStatusRequest struct {
	CardID          string `json:"card_id"`          // 发票 id
	Code            string `json:"code"`             // 发票 code
	ReimburseStatus string `json:"reimburse_status"` // 发票报销状态
}

// GetFiscalAuthUrlRequest 获取授权页链接请求参数
type GetFiscalAuthUrlRequest struct {
	SPAppID     string `json:"s_pappid"`               // 财政局id，需要找财政局提供
	OrderID     string `json:"order_id"`               // 订单id
	Money       int64  `json:"money"`                  // 订单金额，以分为单位
	Timestamp   int64  `json:"timestamp"`              // 时间戳
	Source      string `json:"source"`                 // 开票来源，web：公众号开票，app：app开票
	RedirectUrl string `json:"redirect_url,omitempty"` // 授权成功后跳转页面
	Ticket      string `json:"ticket"`                 // Api_ticket，参考获取api_ticket接口获取
}

// GetFiscalAuthUrlResult 获取授权页链接结果
type GetFiscalAuthUrlResult struct {
	Resp
	AuthUrl    string `json:"auth_url"`    // 授权链接
	ExpireTime int64  `json:"expire_time"` // 过期时间，单位为秒，授权链接会在一段时间之后过期
}

// CreateFiscalCardBaseInfo 财政电子票据信息
type CreateFiscalCardBaseInfo struct {
	LogoUrl string `json:"logo_url"` // 财政局LOGO，请参考上传图片接口
}

// CreateFiscalCardRequest 创建财政电子票据模板请求参数
type CreateFiscalCardRequest struct {
	BaseInfo *CreateFiscalCardBaseInfo `json:"base_info"` // 财政电子票据信息
	Payee    string                    `json:"payee"`     // 收款方（开票方）全称，显示在财政电子票据详情内
}

// CreateFiscalCardResult 创建财政电子票据模板结果
type CreateFiscalCardResult struct {
	Resp
	CardID string `json:"card_id"` // 票据card_id
}

// FiscalInvoiceUserData 财政电子票据用户信息
type FiscalInvoiceUserData struct {
	Fee         int64  `json:"fee"`            // 财政电子票据的金额，以分为单位
	Title       string `json:"title"`          // 财政电子票据的缴费单位
	BillingTime int64  `json:"billing_time"`   // 财政电子票据的开票时间，为10位时间戳（utc+8）
	BillingNo   string `json:"billing_no"`     // 财政电子票据代码
	BillingCode string `json:"billing_code"`   // 财政电子票据号码
	SPdfMediaID string `json:"s_pdf_media_id"` // 财政电子票据pdf文件上传到微信财政电子票据平台后，会生成一个财政电子票据s_media_id，该s_media_id可以直接用于开财政电子票据，上传参考"5、上传pdf"
}

// FiscalCardExt 财政电子票据具体内容
type FiscalCardExt struct {
	UserCard *struct {
		InvoiceUserData *FiscalInvoiceUserData `json:"invoice_user_data"` // 用户信息
	} `json:"user_card"` // 用户信息结构体
}

// InsertFiscalInvoiceRequest 票据插入用户卡包请求参数
type InsertFiscalInvoiceRequest struct {
	OrderID string         `json:"order_id"` // 财政电子票据order_id
	CardID  string         `json:"card_id"`  // 财政电子票据card_id
	AppID   string         `json:"appid"`    // 该订单号授权时使用的appid，一般为执收单位appid
	CardExt *FiscalCardExt `json:"card_ext"` // 财政电子票据具体内容
}
