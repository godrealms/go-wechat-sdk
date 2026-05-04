package offiaccount

// CardBaseInfo is the common base info for all card types
type CardBaseInfo struct {
	LogoUrl           string        `json:"logo_url"`
	BrandName         string        `json:"brand_name"`
	CodeType          string        `json:"code_type"`
	Title             string        `json:"title"`
	SubTitle          string        `json:"sub_title,omitempty"`
	Color             string        `json:"color"`
	Notice            string        `json:"notice"`
	ServicePhone      string        `json:"service_phone,omitempty"`
	Description       string        `json:"description"`
	UseCondition      *UseCondition `json:"use_condition,omitempty"`
	Abstract          *Abstract     `json:"abstract,omitempty"`
	TextImageList     []*TextImage  `json:"text_image_list,omitempty"`
	TimeInfo          *TimeInfo     `json:"time_info"`
	Sku               *Sku          `json:"sku"`
	LocationIdList    []int64       `json:"location_id_list,omitempty"`
	CenterTitle       string        `json:"center_title,omitempty"`
	CenterSubTitle    string        `json:"center_sub_title,omitempty"`
	CenterUrl         string        `json:"center_url,omitempty"`
	CustomUrl         string        `json:"custom_url,omitempty"`
	CustomUrlName     string        `json:"custom_url_name,omitempty"`
	CustomUrlSubTitle string        `json:"custom_url_sub_title,omitempty"`
	PromotionUrl      string        `json:"promotion_url,omitempty"`
	PromotionUrlName  string        `json:"promotion_url_name,omitempty"`
	GetLimit          int           `json:"get_limit,omitempty"`
	CanShare          bool          `json:"can_share,omitempty"`
	CanGiveFriend     bool          `json:"can_give_friend,omitempty"`
}

// UseCondition specifies card usage conditions
type UseCondition struct {
	AcceptCategory          string `json:"accept_category,omitempty"`
	RejectCategory          string `json:"reject_category,omitempty"`
	CanUseWithOtherDiscount bool   `json:"can_use_with_other_discount,omitempty"`
}

// Abstract is a short description with icon
type Abstract struct {
	Abstract    string   `json:"abstract"`
	IconUrlList []string `json:"icon_url_list,omitempty"`
}

// TextImage is one line in the details section
type TextImage struct {
	ImageUrl string `json:"image_url"`
	Text     string `json:"text"`
}

// TimeInfo specifies the validity period
type TimeInfo struct {
	Type           string `json:"type"`
	BeginTimestamp int64  `json:"begin_timestamp,omitempty"`
	EndTimestamp   int64  `json:"end_timestamp,omitempty"`
	FixedTerm      int    `json:"fixed_term,omitempty"`
	FixedBeginTerm int    `json:"fixed_begin_term,omitempty"`
}

// Sku specifies quantity info
type Sku struct {
	Quantity int64 `json:"quantity"`
}

// DiscountCard is a discount card (打折券)
type DiscountCard struct {
	BaseInfo *CardBaseInfo `json:"base_info"`
	Discount int           `json:"discount"`
}

// CashCard is a cash voucher (代金券)
type CashCard struct {
	BaseInfo   *CardBaseInfo `json:"base_info"`
	LeastCost  int64         `json:"least_cost"`
	ReduceCost int64         `json:"reduce_cost"`
}

// MemberCard is a membership card (会员卡)
type MemberCard struct {
	BaseInfo                 *CardBaseInfo `json:"base_info"`
	SupplyBonus              bool          `json:"supply_bonus"`
	SupplyBalance            bool          `json:"supply_balance"`
	BonusCleaned             bool          `json:"bonus_cleared,omitempty"`
	BonusRules               string        `json:"bonus_rules,omitempty"`
	BalanceRules             string        `json:"balance_rules,omitempty"`
	Prerogative              string        `json:"prerogative"`
	AutoActivate             bool          `json:"auto_activate,omitempty"`
	WxActivate               bool          `json:"wx_activate,omitempty"`
	ActivateUrl              string        `json:"activate_url,omitempty"`
	ActivateAppBrandUserName string        `json:"activate_app_brand_user_name,omitempty"`
	ActivateAppBrandPass     string        `json:"activate_app_brand_pass,omitempty"`
}

// CardCreateRequest is the request for CardCreate
type CardCreateRequest struct {
	Card *CardSpec `json:"card"`
}

// CardSpec specifies which card type and its data
type CardSpec struct {
	CardType   string        `json:"card_type"`
	Discount   *DiscountCard `json:"discount,omitempty"`
	Cash       *CashCard     `json:"cash,omitempty"`
	MemberCard *MemberCard   `json:"member_card,omitempty"`
}

// CardCreateResult is the result of CardCreate
type CardCreateResult struct {
	Resp
	CardId string `json:"card_id"`
}

// CardGetResult is the result of CardGet
type CardGetResult struct {
	Resp
	Card *CardSpec `json:"card"`
}

// CardUpdateRequest is the request for CardUpdate
type CardUpdateRequest struct {
	CardId string    `json:"card_id"`
	Card   *CardSpec `json:"card"`
}

// CardQRCodeRequest is the request for GetCardQRCode
type CardQRCodeRequest struct {
	ActionName    string            `json:"action_name"`
	ActionInfo    *CardQRCodeAction `json:"action_info"`
	ExpireSeconds int               `json:"expire_seconds,omitempty"`
}

// CardQRCodeAction specifies which cards go in the QR code
type CardQRCodeAction struct {
	Card  *CardQRItem  `json:"card,omitempty"`
	Scene *CardQRScene `json:"scene,omitempty"`
}

// CardQRItem is one card in the QR code
type CardQRItem struct {
	CardId       string `json:"card_id"`
	Code         string `json:"code,omitempty"`
	OpenId       string `json:"openid,omitempty"`
	IsUniqueCode bool   `json:"is_unique_code,omitempty"`
}

// CardQRScene is scene info for the QR code
type CardQRScene struct {
	SceneStr string `json:"scene_str,omitempty"`
	SceneId  int    `json:"scene_id,omitempty"`
}

// CardQRCodeResult is the result of GetCardQRCode
type CardQRCodeResult struct {
	Resp
	Ticket        string `json:"ticket"`
	ExpireSeconds int    `json:"expire_seconds"`
	Url           string `json:"url"`
	ShowQrcodeUrl string `json:"show_qrcode_url"`
}

// CardCodeRequest is the request for GetCardCode / ConsumeCardCode
type CardCodeRequest struct {
	Code   string `json:"code"`
	CardId string `json:"card_id,omitempty"`
}

// CardCodeResult is the result of GetCardCode
type CardCodeResult struct {
	Resp
	Card       *CardCodeInfo `json:"card"`
	OpenId     string        `json:"openid"`
	CanConsume bool          `json:"can_consume"`
}

// CardCodeInfo is the card info returned with a code
type CardCodeInfo struct {
	CardId    string `json:"card_id"`
	BeginTime int64  `json:"begin_time"`
	EndTime   int64  `json:"end_time"`
}

// ConsumeCardCodeResult is the result of ConsumeCardCode
type ConsumeCardCodeResult struct {
	Resp
	Card   *CardCodeInfo `json:"card"`
	OpenId string        `json:"openid"`
}

// CardListResult is the result of GetCardList
type CardListResult struct {
	Resp
	CardIdList []string `json:"card_id_list"`
	TotalNum   int      `json:"total_num"`
}
