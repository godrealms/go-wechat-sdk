package channels

import "context"

// GetFinderLiveDataListReq holds the parameters for querying Channels live-streaming analytics data.
type GetFinderLiveDataListReq struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Offset    *int   `json:"offset,omitempty"`
	Limit     *int   `json:"limit,omitempty"`
}

// FinderLiveData contains the aggregated view, like, and share counts for a single day.
type FinderLiveData struct {
	Date       string `json:"date"`
	ViewCount  int64  `json:"view_count"`
	LikeCount  int64  `json:"like_count"`
	ShareCount int64  `json:"share_count"`
}

// GetFinderLiveDataListResp is the response returned by GetFinderLiveDataList.
type GetFinderLiveDataListResp struct {
	Items []FinderLiveData `json:"items"`
	Total int              `json:"total"`
}

// GetFinderListReq holds the pagination parameters for listing Channels accounts (Finders).
type GetFinderListReq struct {
	Offset *int `json:"offset,omitempty"`
	Limit  *int `json:"limit,omitempty"`
}

// FinderInfo describes a single Channels account (Finder) with its ID and display name.
type FinderInfo struct {
	FinderID string `json:"finder_id"`
	Nickname string `json:"nickname"`
}

// GetFinderListResp is the response returned by GetFinderList.
type GetFinderListResp struct {
	Items []FinderInfo `json:"items"`
	Total int          `json:"total"`
}

// GetFinderLiveDataList retrieves the live-streaming analytics data list for the Channels account.
func (c *Client) GetFinderLiveDataList(ctx context.Context, req *GetFinderLiveDataListReq) (*GetFinderLiveDataListResp, error) {
	var resp GetFinderLiveDataListResp
	if err := c.doPost(ctx, "/channels/ec/basics/getfinderlivedata", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetFinderList retrieves the list of Channels accounts (Finders) associated with the application.
func (c *Client) GetFinderList(ctx context.Context, req *GetFinderListReq) (*GetFinderListResp, error) {
	var resp GetFinderListResp
	if err := c.doPost(ctx, "/channels/ec/basics/getfinderlist", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
