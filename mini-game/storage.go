package mini_game

import "context"

type KVData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SetUserStorageReq struct {
	OpenID    string   `json:"openid"`
	KVList    []KVData `json:"kv_list"`
	SigMethod string   `json:"sig_method"`
	Signature string   `json:"signature"`
}

type GetUserStorageReq struct {
	OpenID    string   `json:"openid"`
	KeyList   []string `json:"key_list"`
	SigMethod string   `json:"sig_method"`
	Signature string   `json:"signature"`
}
type GetUserStorageResp struct {
	KVList []KVData `json:"kv_list"`
}

func (c *Client) SetUserStorage(ctx context.Context, req *SetUserStorageReq) error {
	return c.doPost(ctx, "/wxa/set_user_storage", req, nil)
}

func (c *Client) GetUserStorage(ctx context.Context, req *GetUserStorageReq) (*GetUserStorageResp, error) {
	var resp GetUserStorageResp
	if err := c.doPost(ctx, "/wxa/get_user_storage", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
