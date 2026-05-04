package mini_program

import "github.com/godrealms/go-wechat-sdk/core"

// LiveRoom represents a live streaming room
type LiveRoom struct {
	Name         string       `json:"name"`
	Roomid       int64        `json:"roomid"`
	CoverImg     string       `json:"cover_img"`
	ShareImg     string       `json:"share_img"`
	LiveStatus   int          `json:"live_status"` // 101=直播中 102=未开始 103=已结束 104=禁播 105=暂停
	StartTime    int64        `json:"start_time"`
	EndTime      int64        `json:"end_time"`
	AnchorName   string       `json:"anchor_name"`
	AnchorImgUrl string       `json:"anchor_img_url"`
	Goods        []*LiveGoods `json:"goods"`
	RoomType     int          `json:"room_type"` // 0=普通直播 1=小店直播
	SubRoomType  int          `json:"sub_room_type"`
}

// LiveGoods represents a product in a live room
type LiveGoods struct {
	GoodsId         int64  `json:"goods_id"`
	CoverImg        string `json:"cover_img"`
	Url             string `json:"url"`
	Price           int64  `json:"price"`
	Name            string `json:"name"`
	Price2          int64  `json:"price2"`
	PriceType       int    `json:"price_type"` // 1=一口价 2=价格区间 3=折扣价
	ThirdPartyAppid string `json:"third_party_appid,omitempty"`
}

// GetLiveRoomsResult is the result of GetLiveRooms
type GetLiveRoomsResult struct {
	core.Resp
	RoomInfo []*LiveRoom `json:"room_info"`
	Total    int64       `json:"total"`
}

// GetLiveRoomsRequest is the request for GetLiveRooms
type GetLiveRoomsRequest struct {
	Start int `json:"start"` // offset, starts from 0
	Limit int `json:"limit"` // max 10
}

// LiveGoodsListResult is the result of GetLiveGoods
type LiveGoodsListResult struct {
	core.Resp
	Goods    []*LiveGoods `json:"goods"`
	TotalNum int64        `json:"total_num"`
}
