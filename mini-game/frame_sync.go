package mini_game

import "context"

type CreateGameRoomReq struct {
	MaxNum     int    `json:"max_num"`
	AccessInfo string `json:"access_info,omitempty"`
}
type CreateGameRoomResp struct {
	RoomID string `json:"room_id"`
}

type GetRoomInfoReq struct {
	RoomID string `json:"room_id"`
}
type RoomMember struct {
	OpenID string `json:"openid"`
	Role   int    `json:"role"`
}
type GetRoomInfoResp struct {
	RoomID  string       `json:"room_id"`
	Status  int          `json:"status"`
	Members []RoomMember `json:"members"`
}

func (c *Client) CreateGameRoom(ctx context.Context, req *CreateGameRoomReq) (*CreateGameRoomResp, error) {
	var resp CreateGameRoomResp
	if err := c.doPost(ctx, "/wxa/game/createroom", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetRoomInfo(ctx context.Context, req *GetRoomInfoReq) (*GetRoomInfoResp, error) {
	var resp GetRoomInfoResp
	if err := c.doPost(ctx, "/wxa/game/getroominfo", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
