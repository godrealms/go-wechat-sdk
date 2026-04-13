package mini_game

import "context"

// CreateGameRoomReq holds the parameters for creating a frame-sync game room.
type CreateGameRoomReq struct {
	MaxNum     int    `json:"max_num"`
	AccessInfo string `json:"access_info,omitempty"`
}

// CreateGameRoomResp is the response returned by CreateGameRoom.
type CreateGameRoomResp struct {
	RoomID string `json:"room_id"`
}

// GetRoomInfoReq holds the room ID needed to query frame-sync room details.
type GetRoomInfoReq struct {
	RoomID string `json:"room_id"`
}

// RoomMember describes a single member in a frame-sync game room.
type RoomMember struct {
	OpenID string `json:"openid"`
	Role   int    `json:"role"`
}

// GetRoomInfoResp is the response returned by GetRoomInfo.
type GetRoomInfoResp struct {
	RoomID  string       `json:"room_id"`
	Status  int          `json:"status"`
	Members []RoomMember `json:"members"`
}

// CreateGameRoom creates a new frame-sync game room and returns its room ID.
func (c *Client) CreateGameRoom(ctx context.Context, req *CreateGameRoomReq) (*CreateGameRoomResp, error) {
	var resp CreateGameRoomResp
	if err := c.doPost(ctx, "/wxa/game/createroom", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRoomInfo retrieves the current status and member list of a frame-sync game room.
func (c *Client) GetRoomInfo(ctx context.Context, req *GetRoomInfoReq) (*GetRoomInfoResp, error) {
	var resp GetRoomInfoResp
	if err := c.doPost(ctx, "/wxa/game/getroominfo", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
