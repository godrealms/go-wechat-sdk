package aispeech

import "context"

// TextToSpeechReq is the request for TTS synthesis.
type TextToSpeechReq struct {
	Text      string `json:"text"`
	Speed     int    `json:"speed,omitempty"`
	Volume    int    `json:"volume,omitempty"`
	Pitch     int    `json:"pitch,omitempty"`
	VoiceType int    `json:"voice_type,omitempty"`
}

// TextToSpeechResp is the response from TextToSpeech.
type TextToSpeechResp struct {
	AudioData string `json:"audio_data"`
	AudioSize int    `json:"audio_size"`
	SessionID string `json:"session_id"`
}

// TextToSpeech converts text to speech and returns the audio as base64 MP3.
func (c *Client) TextToSpeech(ctx context.Context, req *TextToSpeechReq) (*TextToSpeechResp, error) {
	var resp TextToSpeechResp
	if err := c.doPost(ctx, "/aispeech/tts/aitts", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
