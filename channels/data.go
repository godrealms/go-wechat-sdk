package channels

import "context"

// ---------- 数据统计 ----------

type GetFinderLiveDataListReq struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Offset    *int   `json:"offset,omitempty"`
	Limit     *int   `json:"limit,omitempty"`
}
type FinderLiveData struct {
	Date       string `json:"date"`
	ViewCount  int64  `json:"view_count"`
	LikeCount  int64  `json:"like_count"`
	ShareCount int64  `json:"share_count"`
}
type GetFinderLiveDataListResp struct {
	Items []FinderLiveData `json:"items"`
	Total int              `json:"total"`
}

type GetFinderListReq struct {
	Offset *int `json:"offset,omitempty"`
	Limit  *int `json:"limit,omitempty"`
}
type FinderInfo struct {
	FinderID string `json:"finder_id"`
	Nickname string `json:"nickname"`
}
type GetFinderListResp struct {
	Items []FinderInfo `json:"items"`
	Total int          `json:"total"`
}

// GetFinderLiveDataList 获取直播数据列表
func (c *Client) GetFinderLiveDataList(ctx context.Context, req *GetFinderLiveDataListReq) (*GetFinderLiveDataListResp, error) {
	var resp GetFinderLiveDataListResp
	if err := c.doPost(ctx, "/channels/ec/basics/getfinderlivedata", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFinderList 获取视频号列表
func (c *Client) GetFinderList(ctx context.Context, req *GetFinderListReq) (*GetFinderListResp, error) {
	var resp GetFinderListResp
	if err := c.doPost(ctx, "/channels/ec/basics/getfinderlist", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
