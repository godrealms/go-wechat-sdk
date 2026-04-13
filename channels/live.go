package channels

import "context"

// CreateRoomReq holds the parameters for creating a new Channels live-streaming room.
type CreateRoomReq struct {
	Name      string `json:"name"`
	CoverImg  string `json:"cover_img,omitempty"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}

// CreateRoomResp is the response returned by CreateRoom.
type CreateRoomResp struct {
	RoomID string `json:"room_id"`
}

// DeleteRoomReq holds the room ID of the Channels live room to delete.
type DeleteRoomReq struct {
	RoomID string `json:"room_id"`
}

// GetLiveInfoReq holds the room ID for querying live room details.
type GetLiveInfoReq struct {
	RoomID string `json:"room_id"`
}

// LiveInfo contains the current status and scheduling details of a Channels live room.
type LiveInfo struct {
	RoomID    string `json:"room_id"`
	Name      string `json:"name"`
	Status    int    `json:"status"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}

// GetLiveInfoResp is the response returned by GetLiveInfo.
type GetLiveInfoResp struct {
	LiveInfo LiveInfo `json:"live_info"`
}

// GetLiveReplayListReq holds the parameters for listing replay recordings of a live room.
type GetLiveReplayListReq struct {
	RoomID string `json:"room_id"`
	Offset *int   `json:"offset,omitempty"`
	Limit  *int   `json:"limit,omitempty"`
}

// LiveReplay describes a single replay recording for a Channels live room.
type LiveReplay struct {
	MediaURL   string `json:"media_url"`
	ExpireTime int64  `json:"expire_time"`
	CreateTime int64  `json:"create_time"`
}

// GetLiveReplayListResp is the response returned by GetLiveReplayList.
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

// DeleteRoom deletes the specified Channels live-streaming room.
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

// GetLiveReplayList retrieves the list of replay recordings for the specified Channels live room.
func (c *Client) GetLiveReplayList(ctx context.Context, req *GetLiveReplayListReq) (*GetLiveReplayListResp, error) {
	var resp GetLiveReplayListResp
	if err := c.doPost(ctx, "/channels/ec/basics/live/getlivereplaylist", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
