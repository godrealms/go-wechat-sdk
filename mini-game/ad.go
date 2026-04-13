package mini_game

import "context"

// GetGameAdDataReq holds the request parameters for fetching game advertisement data.
type GetGameAdDataReq struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	AdUnitID  string `json:"ad_unit_id,omitempty"`
}

// GameAdData contains the advertising metrics for a single ad unit on a given date.
type GameAdData struct {
	Date       string `json:"date"`
	AdUnitID   string `json:"ad_unit_id"`
	ReqCount   int64  `json:"req_count"`
	ShowCount  int64  `json:"show_count"`
	ClickCount int64  `json:"click_count"`
	Income     int64  `json:"income"`
}

// GetGameAdDataResp is the response returned by GetGameAdData.
type GetGameAdDataResp struct {
	Items []GameAdData `json:"items"`
}

// GetGameAdData retrieves advertising performance data for the Mini Game within the given date range.
func (c *Client) GetGameAdData(ctx context.Context, req *GetGameAdDataReq) (*GetGameAdDataResp, error) {
	var resp GetGameAdDataResp
	if err := c.doPost(ctx, "/wxa/game/getgameaddata", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
