package channels

import "context"

// ---------- 直播间管理 ----------

type CreateRoomReq struct {
	Name      string `json:"name"`
	CoverImg  string `json:"cover_img,omitempty"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}
type CreateRoomResp struct {
	RoomID string `json:"room_id"`
}

type DeleteRoomReq struct {
	RoomID string `json:"room_id"`
}

type GetLiveInfoReq struct {
	RoomID string `json:"room_id"`
}
type LiveInfo struct {
	RoomID    string `json:"room_id"`
	Name      string `json:"name"`
	Status    int    `json:"status"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}
type GetLiveInfoResp struct {
	LiveInfo LiveInfo `json:"live_info"`
}

type GetLiveReplayListReq struct {
	RoomID string `json:"room_id"`
	Offset *int   `json:"offset,omitempty"`
	Limit  *int   `json:"limit,omitempty"`
}
type LiveReplay struct {
	MediaURL   string `json:"media_url"`
	ExpireTime int64  `json:"expire_time"`
	CreateTime int64  `json:"create_time"`
}
type GetLiveReplayListResp struct {
	LiveReplayList []LiveReplay `json:"live_replay_list"`
	Total          int          `json:"total"`
}

// CreateRoom creates a new live-streaming room for the Channels account.
func (c *Client) CreateRoom(ctx context.Context, req *CreateRoomReq) (*CreateRoomResp, error) {
	var resp CreateRoomResp
	if err := c.doPost(ctx, "/channels/ec/basics/live/createroom", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteRoom 删除直播间
func (c *Client) DeleteRoom(ctx context.Context, req *DeleteRoomReq) error {
	return c.doPost(ctx, "/channels/ec/basics/live/deleteroom", req, nil)
}

// GetLiveInfo retrieves the current status and metadata for the specified live room.
func (c *Client) GetLiveInfo(ctx context.Context, req *GetLiveInfoReq) (*GetLiveInfoResp, error) {
	var resp GetLiveInfoResp
	if err := c.doPost(ctx, "/channels/ec/basics/live/getliveinfo", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLiveReplayList 获取直播间回放列表
func (c *Client) GetLiveReplayList(ctx context.Context, req *GetLiveReplayListReq) (*GetLiveReplayListResp, error) {
	var resp GetLiveReplayListResp
	if err := c.doPost(ctx, "/channels/ec/basics/live/getlivereplaylist", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
