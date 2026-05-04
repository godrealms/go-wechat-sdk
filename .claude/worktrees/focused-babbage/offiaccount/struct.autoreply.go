package offiaccount

type AddFriendAutoReplyInfo struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type MessageDefaultAutoReplyInfo struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}
type KeywordListInfo struct {
	Type      string `json:"type"`
	MatchMode string `json:"match_mode"`
	Content   string `json:"content"`
}

type NewsInfo struct {
	Title      string `json:"title"`
	Author     string `json:"author"`
	Digest     string `json:"digest"`
	ShowCover  int    `json:"show_cover"`
	CoverUrl   string `json:"cover_url"`
	ContentUrl string `json:"content_url"`
	SourceUrl  string `json:"source_url"`
}

type NewsInfoList struct {
	List []*NewsInfo `json:"list"`
}

type ReplyListInfo struct {
	Type     string        `json:"type"`
	NewsInfo *NewsInfoList `json:"news_info,omitempty"`
	Content  string        `json:"content,omitempty"`
}

type KeywordAutoReplyInfo struct {
	RuleName        string             `json:"rule_name"`
	CreateTime      int                `json:"create_time"`
	ReplyMode       string             `json:"reply_mode"`
	KeywordListInfo []*KeywordListInfo `json:"keyword_list_info"`
	ReplyListInfo   []*ReplyListInfo   `json:"reply_list_info"`
}

type KeywordAutoReplyInfoList struct {
	List []*KeywordAutoReplyInfo `json:"list"`
}

type ReplyResp struct {
	IsAddFriendReplyOpen        int                          `json:"is_add_friend_reply_open"`
	IsAutoReplyOpen             int                          `json:"is_autoreply_open"`
	AddFriendAutoReplyInfo      *AddFriendAutoReplyInfo      `json:"add_friend_autoreply_info"`
	MessageDefaultAutoReplyInfo *MessageDefaultAutoReplyInfo `json:"message_default_autoreply_info"`
	KeywordAutoReplyInfo        *KeywordAutoReplyInfoList    `json:"keyword_autoreply_info"`
}
