package aispeech

import "context"

// DialogQueryReq is the request for a multi-turn dialog query.
type DialogQueryReq struct {
	Query     string `json:"query"`
	SessionID string `json:"session_id"`
	Lang      string `json:"lang,omitempty"`
}

// DialogQueryResp is the response from DialogQuery.
type DialogQueryResp struct {
	Answer    string `json:"answer"`
	SessionID string `json:"session_id"`
	EndFlag   bool   `json:"end_flag"`
}

// DialogQuery sends a user utterance and returns the system reply.
func (c *Client) DialogQuery(ctx context.Context, req *DialogQueryReq) (*DialogQueryResp, error) {
	var resp DialogQueryResp
	if err := c.doPost(ctx, "/aispeech/dialog/airequ", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DialogResetReq is the request to terminate and reset a dialog session.
type DialogResetReq struct {
	SessionID string `json:"session_id"`
}

// DialogResetResp is the response from DialogReset.
type DialogResetResp struct{}

// DialogReset terminates the dialog session identified by SessionID.
func (c *Client) DialogReset(ctx context.Context, req *DialogResetReq) error {
	return c.doPost(ctx, "/aispeech/dialog/aireset", req, nil)
}
