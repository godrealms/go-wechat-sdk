package mini_program

import (
	"fmt"
	"net/url"
)

// GetLiveRooms 获取直播间列表及直播间信息
// GET /wxa/business/getliveinfo
func (c *Client) GetLiveRooms(start, limit int) (*GetLiveRoomsResult, error) {
	query := c.TokenQuery(url.Values{
		"start": {fmt.Sprintf("%d", start)},
		"limit": {fmt.Sprintf("%d", limit)},
	})
	result := &GetLiveRoomsResult{}
	err := c.Https.Get(c.Ctx, "/wxa/business/getliveinfo", query, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetLiveRoomInfo 获取指定直播间信息
// GET /wxa/business/getliveinfo?room_id=ROOMID
func (c *Client) GetLiveRoomInfo(roomId int64) (*GetLiveRoomsResult, error) {
	query := c.TokenQuery(url.Values{
		"room_id": {fmt.Sprintf("%d", roomId)},
	})
	result := &GetLiveRoomsResult{}
	err := c.Https.Get(c.Ctx, "/wxa/business/getliveinfo", query, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetLiveGoods 获取商品列表
// GET /wxa/business/getgoodswarehouse
func (c *Client) GetLiveGoods(start, limit int) (*LiveGoodsListResult, error) {
	query := c.TokenQuery(url.Values{
		"offset": {fmt.Sprintf("%d", start)},
		"limit":  {fmt.Sprintf("%d", limit)},
	})
	result := &LiveGoodsListResult{}
	err := c.Https.Get(c.Ctx, "/wxa/business/getgoodswarehouse", query, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// AddLiveGoods 直播间导入商品
// POST /wxa/business/add_goods
func (c *Client) AddLiveGoods(roomId int64, goodsIds []int64) error {
	path := fmt.Sprintf("/wxa/business/add_goods?access_token=%s", c.GetAccessToken())
	body := map[string]interface{}{
		"roomId":  roomId,
		"goodsId": goodsIds,
	}
	result := &struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}{}
	err := c.Https.Post(c.Ctx, path, body, result)
	if err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}
