package mini_game

import "context"

type GetGameAdDataReq struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	AdUnitID  string `json:"ad_unit_id,omitempty"`
}
type GameAdData struct {
	Date       string `json:"date"`
	AdUnitID   string `json:"ad_unit_id"`
	ReqCount   int64  `json:"req_count"`
	ShowCount  int64  `json:"show_count"`
	ClickCount int64  `json:"click_count"`
	Income     int64  `json:"income"`
}
type GetGameAdDataResp struct {
	Items []GameAdData `json:"items"`
}

func (c *Client) GetGameAdData(ctx context.Context, req *GetGameAdDataReq) (*GetGameAdDataResp, error) {
	var resp GetGameAdDataResp
	if err := c.doPost(ctx, "/wxa/game/getgameaddata", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
