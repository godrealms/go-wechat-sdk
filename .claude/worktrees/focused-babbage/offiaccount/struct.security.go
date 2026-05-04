package offiaccount

// MsgSecCheckResult is the result of MsgSecCheck
type MsgSecCheckResult struct {
	Resp
	TraceId string               `json:"trace_id"`
	Result  *MsgSecCheckSummary  `json:"result"`
	Detail  []*MsgSecCheckDetail `json:"detail"`
}

// MsgSecCheckSummary is the overall judgment
type MsgSecCheckSummary struct {
	Suggest string `json:"suggest"` // pass/review/risky
	Label   int    `json:"label"`
}

// MsgSecCheckDetail contains one label's detailed judgment
type MsgSecCheckDetail struct {
	Strategy string `json:"strategy"`
	ErrCode  int    `json:"errcode"`
	Suggest  string `json:"suggest"` // pass/review/risky
	Label    int    `json:"label"`
	Level    int    `json:"level"`
	Prob     int    `json:"prob"`
	KeyWord  string `json:"keyword"`
}

// MsgSecCheckRequest is the request for MsgSecCheck
type MsgSecCheckRequest struct {
	Content   string `json:"content"`
	Version   int    `json:"version"` // 1 or 2
	Scene     int    `json:"scene"`   // 1=资料 2=评论 3=论坛 4=社交日志
	Openid    string `json:"openid"`
	Title     string `json:"title,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// MediaCheckAsyncResult is the result of MediaCheckAsync
type MediaCheckAsyncResult struct {
	Resp
	TraceId string `json:"trace_id"`
}

// MediaCheckAsyncRequest is the request for MediaCheckAsync
type MediaCheckAsyncRequest struct {
	MediaUrl  string `json:"media_url"`
	MediaType int    `json:"media_type"` // 1=音频 2=图片
	Version   int    `json:"version"`
	Scene     int    `json:"scene"`
	Openid    string `json:"openid"`
}
