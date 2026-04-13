package aispeech

import "context"

// NLUUnderstandReq is the request for natural language understanding.
type NLUUnderstandReq struct {
	Query     string `json:"query"`
	SessionID string `json:"session_id,omitempty"`
	Lang      string `json:"lang,omitempty"`
}

// NLUEntity is a recognized named entity within the query.
type NLUEntity struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Begin int    `json:"begin"`
	End   int    `json:"end"`
}

// NLUUnderstandResp is the response from NLUUnderstand.
type NLUUnderstandResp struct {
	Intent    string      `json:"intent"`
	Slots     []NLUEntity `json:"slots"`
	SessionID string      `json:"session_id"`
}

// NLUUnderstand extracts intent and named entities from the input text.
func (c *Client) NLUUnderstand(ctx context.Context, req *NLUUnderstandReq) (*NLUUnderstandResp, error) {
	var resp NLUUnderstandResp
	if err := c.doPost(ctx, "/aispeech/nlu/airequ", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// NLUIntentRecognizeReq is the request for intent classification.
type NLUIntentRecognizeReq struct {
	Query     string   `json:"query"`
	IntentIDs []string `json:"intent_ids,omitempty"`
	SessionID string   `json:"session_id,omitempty"`
}

// NLUIntentRecognizeResp is the response from NLUIntentRecognize.
type NLUIntentRecognizeResp struct {
	IntentID   string  `json:"intent_id"`
	IntentName string  `json:"intent_name"`
	Confidence float64 `json:"confidence"`
	SessionID  string  `json:"session_id"`
}

// NLUIntentRecognize classifies the query against a configured intent set.
func (c *Client) NLUIntentRecognize(ctx context.Context, req *NLUIntentRecognizeReq) (*NLUIntentRecognizeResp, error) {
	var resp NLUIntentRecognizeResp
	if err := c.doPost(ctx, "/aispeech/nlu/aiintentrequ", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
