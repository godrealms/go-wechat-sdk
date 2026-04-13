package aispeech

import "context"

// ASRLongReq is the request for long-audio asynchronous speech recognition.
type ASRLongReq struct {
	VoiceID     string `json:"voice_id"`
	VoiceURL    string `json:"voice_url"`
	Format      string `json:"voice_format"`
	Lang        string `json:"lang,omitempty"`
	CallbackURL string `json:"callback_url,omitempty"`
}

// ASRLongResp is the response from ASRLong.
type ASRLongResp struct {
	TaskID string `json:"task_id"`
}

// ASRLong submits a long-audio speech recognition job (≤300 s).
func (c *Client) ASRLong(ctx context.Context, req *ASRLongReq) (*ASRLongResp, error) {
	var resp ASRLongResp
	if err := c.doPost(ctx, "/aispeech/asr/aiasrlong", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ASRShortReq is the request for short-audio synchronous speech recognition.
type ASRShortReq struct {
	VoiceID   string `json:"voice_id"`
	VoiceData string `json:"voice_data"`
	Format    string `json:"voice_format"`
	Rate      int    `json:"voice_rate"`
	Bits      int    `json:"voice_bits"`
	Lang      string `json:"lang,omitempty"`
}

// ASRShortResp is the response from ASRShort.
type ASRShortResp struct {
	Result string `json:"result"`
}

// ASRShort performs synchronous speech recognition on audio data ≤60 s.
func (c *Client) ASRShort(ctx context.Context, req *ASRShortReq) (*ASRShortResp, error) {
	var resp ASRShortResp
	if err := c.doPost(ctx, "/aispeech/asr/aiasrshort", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
