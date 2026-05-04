package offiaccount

// CommentType 评论类型
type CommentType int

const (
	CommentTypeAll      CommentType = 0 // 普通评论&精选评论
	CommentTypeNormal   CommentType = 1 // 普通评论
	CommentTypeSelected CommentType = 2 // 精选评论
)

// OpenArticleCommentRequest 打开已群发文章评论请求参数
type OpenArticleCommentRequest struct {
	MsgDataID int64 `json:"msg_data_id"` // 群发返回的msg_data_id
	Index     int   `json:"index"`       // 多图文时，用来指定第几篇图文，从0开始，不带默认操作该msg_data_id的第一篇图文
}

// CloseArticleCommentRequest 关闭已群发文章评论请求参数
type CloseArticleCommentRequest struct {
	MsgDataID int64 `json:"msg_data_id"` // 群发返回的msg_data_id
	Index     int   `json:"index"`       // 多图文时，用来指定第几篇图文，从0开始，不带默认操作该msg_data_id的第一篇图文
}

// ListCommentRequest 查看指定文章的评论数据请求参数
type ListCommentRequest struct {
	MsgDataID int64       `json:"msg_data_id"` // 群发返回的msg_data_id
	Index     int         `json:"index"`       // 多图文时，用来指定第几篇图文，从0开始，不带默认返回该msg_data_id的第一篇图文
	Begin     int         `json:"begin"`       // 起始位置
	Count     int         `json:"count"`       // 获取数目（>=50会被拒绝）
	Type      CommentType `json:"type"`        // type=0 普通评论&精选评论 type=1 普通评论 type=2 精选评论
}

// CommentReply 评论回复信息
type CommentReply struct {
	Content    string `json:"content"`     // 回复内容
	CreateTime int64  `json:"create_time"` // 回复时间
}

// Comment 评论信息
type Comment struct {
	UserCommentID int64         `json:"user_comment_id"` // 用户评论id
	CreateTime    int64         `json:"create_time"`     // 评论时间
	Content       string        `json:"content"`         // 评论内容
	CommentType   int           `json:"comment_type"`    // 是否精选评论，0为即非精选，1为true，即精选
	OpenID        string        `json:"openid"`          // openid，用户如果用非微信身份评论，不返回openid
	Reply         *CommentReply `json:"reply"`           // 回复信息
}

// ListCommentResult 查看指定文章的评论数据结果
type ListCommentResult struct {
	Resp
	Total   int        `json:"total"`   // 评论总数
	Comment []*Comment `json:"comment"` // 评论列表
}

// ElectCommentRequest 评论标记精选请求参数
type ElectCommentRequest struct {
	MsgDataID     int64 `json:"msg_data_id"`     // 群发返回的msg_data_id
	Index         int   `json:"index"`           // 多图文时，用来指定第几篇图文，从0开始，不带默认操作该msg_data_id的第一篇图文
	UserCommentID int64 `json:"user_comment_id"` // 用户评论id
}

// UnElectCommentRequest 评论取消精选请求参数
type UnElectCommentRequest struct {
	MsgDataID     int64 `json:"msg_data_id"`     // 群发返回的msg_data_id
	Index         int   `json:"index"`           // 多图文时，用来指定第几篇图文，从0开始，不带默认操作该msg_data_id的第一篇图文
	UserCommentID int64 `json:"user_comment_id"` // 用户评论id
}

// DeleteCommentRequest 删除评论请求参数
type DeleteCommentRequest struct {
	MsgDataID     int64 `json:"msg_data_id"`     // 群发返回的msg_data_id
	Index         int   `json:"index"`           // 多图文时，用来指定第几篇图文，从0开始，不带默认操作该msg_data_id的第一篇图文
	UserCommentID int64 `json:"user_comment_id"` // 评论id
}

// ReplyCommentRequest 回复评论请求参数
type ReplyCommentRequest struct {
	MsgDataID     int64  `json:"msg_data_id"`     // 群发返回的msg_data_id
	Index         int    `json:"index"`           // 多图文时，用来指定第几篇图文，从0开始，不带默认操作该msg_data_id的第一篇图文
	UserCommentID int64  `json:"user_comment_id"` // 评论id
	Content       string `json:"content"`         // 回复内容
}

// DeleteReplyCommentRequest 删除回复请求参数
type DeleteReplyCommentRequest struct {
	MsgDataID     int64 `json:"msg_data_id"`     // 群发返回的msg_data_id
	Index         int   `json:"index"`           // 多图文时，用来指定第几篇图文，从0开始，不带默认操作该msg_data_id的第一篇图文
	UserCommentID int64 `json:"user_comment_id"` // 评论id
}
