package mini_program

import "github.com/godrealms/go-wechat-sdk/core"

// AddExpressOrderRequest is the request to create an express waybill
type AddExpressOrderRequest struct {
	AddSource    int             `json:"add_source"` // 0=微信侧 2=自定义
	WxAppId      string          `json:"wx_appid,omitempty"`
	OrderId      string          `json:"order_id"`
	OpenId       string          `json:"openid,omitempty"`
	DeliveryId   string          `json:"delivery_id"`
	BizId        string          `json:"biz_id"`
	CustomRemark string          `json:"custom_remark,omitempty"`
	Tagid        int             `json:"tagid,omitempty"`
	Sender       *ExpressContact `json:"sender"`
	Receiver     *ExpressContact `json:"receiver"`
	Cargo        *ExpressCargo   `json:"cargo"`
	Shop         *ExpressShop    `json:"shop,omitempty"`
	SubBizId     string          `json:"sub_biz_id,omitempty"`
}

// ExpressContact represents sender or receiver info
type ExpressContact struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Company  string `json:"company,omitempty"`
	PostCode string `json:"post_code,omitempty"`
	Country  string `json:"country,omitempty"`
	Province string `json:"province"`
	City     string `json:"city"`
	Area     string `json:"area"`
	Address  string `json:"address"`
}

// ExpressCargo describes the parcel contents
type ExpressCargo struct {
	Count      int                   `json:"count"`
	Weight     float64               `json:"weight"`
	SpaceX     float64               `json:"space_x"`
	SpaceY     float64               `json:"space_y"`
	SpaceZ     float64               `json:"space_z"`
	DetailList []*ExpressCargoDetail `json:"detail_list"`
}

// ExpressCargoDetail is one item in the parcel
type ExpressCargoDetail struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ExpressShop holds shop info for print
type ExpressShop struct {
	WxaPath    string `json:"wxa_path"`
	ImgUrl     string `json:"img_url,omitempty"`
	GoodsName  string `json:"goods_name,omitempty"`
	GoodsCount int    `json:"goods_count,omitempty"`
}

// AddExpressOrderResult is the result of AddExpressOrder
type AddExpressOrderResult struct {
	core.Resp
	OrderId     string         `json:"order_id"`
	WaybillId   string         `json:"waybill_id"`
	WaybillData []*WaybillData `json:"waybill_data"`
}

// WaybillData is one field on the printed waybill
type WaybillData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetExpressOrderResult is the result of GetExpressOrder
type GetExpressOrderResult struct {
	core.Resp
	OrderId      string             `json:"order_id"`
	OpenId       string             `json:"openid,omitempty"`
	DeliveryId   string             `json:"delivery_id"`
	WaybillId    string             `json:"waybill_id"`
	WaybillData  []*WaybillData     `json:"waybill_data"`
	PathItemNum  int                `json:"path_item_num"`
	PathItemList []*ExpressPathItem `json:"path_item_list"`
}

// ExpressPathItem is one tracking update
type ExpressPathItem struct {
	ActionTime int64  `json:"action_time"`
	ActionType int    `json:"action_type"`
	ActionMsg  string `json:"action_msg"`
}

// CancelExpressOrderRequest is the request to cancel a waybill
type CancelExpressOrderRequest struct {
	OrderId    string `json:"order_id"`
	OpenId     string `json:"openid,omitempty"`
	DeliveryId string `json:"delivery_id"`
	WaybillId  string `json:"waybill_id"`
}

// CancelExpressOrderResult is the result of CancelExpressOrder
type CancelExpressOrderResult struct {
	core.Resp
	Count int `json:"count"` // 0=失败 1=成功
}

// ExpressDelivery represents an express delivery company
type ExpressDelivery struct {
	DeliveryId   string `json:"delivery_id"`
	DeliveryName string `json:"delivery_name"`
}

// GetAllDeliveryResult is the result of GetAllDelivery
type GetAllDeliveryResult struct {
	core.Resp
	Count int                `json:"count"`
	Data  []*ExpressDelivery `json:"data"`
}
