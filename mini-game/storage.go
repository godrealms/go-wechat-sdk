package mini_game

import "context"

// KVData holds a single key-value pair used in Mini Game cloud storage operations.
type KVData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SetUserStorageReq holds the parameters for writing key-value pairs to a user's cloud storage.
type SetUserStorageReq struct {
	OpenID    string   `json:"openid"`
	KVList    []KVData `json:"kv_list"`
	SigMethod string   `json:"sig_method"`
	Signature string   `json:"signature"`
}

// GetUserStorageReq holds the parameters for reading key-value pairs from a user's cloud storage.
type GetUserStorageReq struct {
	OpenID    string   `json:"openid"`
	KeyList   []string `json:"key_list"`
	SigMethod string   `json:"sig_method"`
	Signature string   `json:"signature"`
}

// GetUserStorageResp is the response returned by GetUserStorage.
type GetUserStorageResp struct {
	KVList []KVData `json:"kv_list"`
}

// SetUserStorage writes the provided key-value pairs to the specified user's Mini Game cloud storage.
func (c *Client) SetUserStorage(ctx context.Context, req *SetUserStorageReq) error {
	return c.doPost(ctx, "/wxa/set_user_storage", req, nil)
}

// GetUserStorage retrieves the stored key-value pairs for the specified user from Mini Game cloud storage.
func (c *Client) GetUserStorage(ctx context.Context, req *GetUserStorageReq) (*GetUserStorageResp, error) {
	var resp GetUserStorageResp
	if err := c.doPost(ctx, "/wxa/get_user_storage", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
