package offiaccount

import "fmt"

// AddNewsMaterialRequest 新增临时图文素材请求结构体
type AddNewsMaterialRequest struct {
	Articles []Article `json:"articles"` // 图文消息，一个图文消息支持1到8条图文
}

// Article 图文消息单条内容结构体
type Article struct {
	Title              string `json:"title"`                           // 图文消息的标题（必填）
	Author             string `json:"author,omitempty"`                // 图文消息的作者
	ThumbMediaID       string `json:"thumb_media_id"`                  // 图文消息缩略图的media_id（必填）
	Content            string `json:"content"`                         // 图文消息页面的内容，支持HTML标签（必填）
	ContentSourceURL   string `json:"content_source_url,omitempty"`    // 点击"阅读原文"后的页面链接
	Digest             string `json:"digest,omitempty"`                // 图文消息的描述，为空时默认抓取正文前64个字
	ShowCoverPic       int    `json:"show_cover_pic,omitempty"`        // 是否显示封面：1显示，0不显示
	NeedOpenComment    int    `json:"need_open_comment,omitempty"`     // 是否打开评论：0不打开，1打开
	OnlyFansCanComment int    `json:"only_fans_can_comment,omitempty"` // 是否粉丝才可评论：0所有人可评论，1粉丝才可评论
}

// AddNewsMaterialResponse 新增临时图文素材响应结构体
type AddNewsMaterialResponse struct {
	ErrCode   int    `json:"errcode"`    // 错误码
	ErrMsg    string `json:"errmsg"`     // 错误信息
	Type      string `json:"type"`       // 媒体文件类型，分别有图片（image）、语音（voice）、视频（video）和缩略图（thumb），图文消息为news
	MediaID   string `json:"media_id"`   // 媒体文件上传后，获取标识
	CreatedAt int64  `json:"created_at"` // 媒体文件上传时间戳
}

// 常量定义
const (
	// 显示封面设置
	ShowCoverPicNo  = 0 // 不显示封面
	ShowCoverPicYes = 1 // 显示封面

	// 评论设置
	CommentClosed = 0 // 不打开评论
	CommentOpen   = 1 // 打开评论

	// 评论权限设置
	CommentAllUsers = 0 // 所有人可评论
	CommentFansOnly = 1 // 粉丝才可评论

	// 图文消息数量限制
	MaxArticleCount = 8 // 最多8条图文
	MinArticleCount = 1 // 最少1条图文
)

// 构造函数和辅助方法

// NewAddNewsMaterialRequest 创建新增临时图文素材请求
func NewAddNewsMaterialRequest() *AddNewsMaterialRequest {
	return &AddNewsMaterialRequest{
		Articles: make([]Article, 0),
	}
}

// AddArticle 添加图文消息
func (r *AddNewsMaterialRequest) AddArticle(article Article) *AddNewsMaterialRequest {
	if len(r.Articles) < MaxArticleCount {
		r.Articles = append(r.Articles, article)
	}
	return r
}

// NewArticle 创建图文消息
func NewArticle(title, thumbMediaID, content string) Article {
	return Article{
		Title:        title,
		ThumbMediaID: thumbMediaID,
		Content:      content,
		ShowCoverPic: ShowCoverPicYes, // 默认显示封面
	}
}

// SetAuthor 设置作者
func (a *Article) SetAuthor(author string) *Article {
	a.Author = author
	return a
}

// SetContentSourceURL 设置阅读原文链接
func (a *Article) SetContentSourceURL(url string) *Article {
	a.ContentSourceURL = url
	return a
}

// SetDigest 设置摘要
func (a *Article) SetDigest(digest string) *Article {
	a.Digest = digest
	return a
}

// SetShowCoverPic 设置是否显示封面
func (a *Article) SetShowCoverPic(show bool) *Article {
	if show {
		a.ShowCoverPic = ShowCoverPicYes
	} else {
		a.ShowCoverPic = ShowCoverPicNo
	}
	return a
}

// SetComment 设置评论功能
func (a *Article) SetComment(needOpen bool, onlyFans bool) *Article {
	if needOpen {
		a.NeedOpenComment = CommentOpen
		if onlyFans {
			a.OnlyFansCanComment = CommentFansOnly
		} else {
			a.OnlyFansCanComment = CommentAllUsers
		}
	} else {
		a.NeedOpenComment = CommentClosed
		a.OnlyFansCanComment = CommentAllUsers // 不开启评论时，这个字段无意义
	}
	return a
}

// EnableComment 开启评论（所有人可评论）
func (a *Article) EnableComment() *Article {
	return a.SetComment(true, false)
}

// EnableCommentForFansOnly 开启评论（仅粉丝可评论）
func (a *Article) EnableCommentForFansOnly() *Article {
	return a.SetComment(true, true)
}

// DisableComment 关闭评论
func (a *Article) DisableComment() *Article {
	return a.SetComment(false, false)
}

// 验证方法

// Validate 验证请求参数
func (r *AddNewsMaterialRequest) Validate() error {
	if len(r.Articles) < MinArticleCount {
		return fmt.Errorf("图文消息数量不能少于%d条", MinArticleCount)
	}

	if len(r.Articles) > MaxArticleCount {
		return fmt.Errorf("图文消息数量不能超过%d条", MaxArticleCount)
	}

	for i, article := range r.Articles {
		if err := article.Validate(); err != nil {
			return fmt.Errorf("第%d条图文消息验证失败: %v", i+1, err)
		}
	}

	return nil
}

// Validate 验证单条图文消息参数
func (a *Article) Validate() error {
	if a.Title == "" {
		return fmt.Errorf("标题不能为空")
	}

	if a.ThumbMediaID == "" {
		return fmt.Errorf("缩略图media_id不能为空")
	}

	if a.Content == "" {
		return fmt.Errorf("内容不能为空")
	}

	// 验证评论设置的逻辑性
	if a.NeedOpenComment == CommentClosed && a.OnlyFansCanComment == CommentFansOnly {
		// 虽然不是错误，但逻辑上不合理，可以给出警告
		// 这里选择不报错，只是标准化设置
		a.OnlyFansCanComment = CommentAllUsers
	}

	return nil
}

// 辅助方法

// GetArticleCount 获取图文消息数量
func (r *AddNewsMaterialRequest) GetArticleCount() int {
	return len(r.Articles)
}

// IsEmpty 检查是否为空
func (r *AddNewsMaterialRequest) IsEmpty() bool {
	return len(r.Articles) == 0
}

// IsFull 检查是否已满
func (r *AddNewsMaterialRequest) IsFull() bool {
	return len(r.Articles) >= MaxArticleCount
}

// Clear 清空所有图文消息
func (r *AddNewsMaterialRequest) Clear() *AddNewsMaterialRequest {
	r.Articles = make([]Article, 0)
	return r
}

// RemoveArticle 移除指定位置的图文消息
func (r *AddNewsMaterialRequest) RemoveArticle(index int) *AddNewsMaterialRequest {
	if index >= 0 && index < len(r.Articles) {
		r.Articles = append(r.Articles[:index], r.Articles[index+1:]...)
	}
	return r
}

// NewSingleArticleRequest 创建单条图文消息请求
func NewSingleArticleRequest(title, thumbMediaID, content string) *AddNewsMaterialRequest {
	article := NewArticle(title, thumbMediaID, content)
	return NewAddNewsMaterialRequest().AddArticle(article)
}

// NewMultiArticleRequest 创建多条图文消息请求
func NewMultiArticleRequest(articles ...Article) *AddNewsMaterialRequest {
	req := NewAddNewsMaterialRequest()
	for _, article := range articles {
		if req.IsFull() {
			break
		}
		req.AddArticle(article)
	}
	return req
}

// MaterialType 素材类型
type MaterialType string

const (
	MaterialTypeImage MaterialType = "image" // 图片
	MaterialTypeVoice MaterialType = "voice" // 语音
	MaterialTypeVideo MaterialType = "video" // 视频
	MaterialTypeNews  MaterialType = "news"  // 图文
	MaterialTypeThumb MaterialType = "thumb" // 缩略图
)

// MaterialCount 素材总数
type MaterialCount struct {
	Resp
	VoiceCount int `json:"voice_count"` // 语音总数量
	VideoCount int `json:"video_count"` // 视频总数量
	ImageCount int `json:"image_count"` // 图片总数量
	NewsCount  int `json:"news_count"`  // 图文总数量
}

// NewsItem 图文消息条目
type NewsItem struct {
	Title            string `json:"title"`              // 图文消息的标题
	ThumbMediaID     string `json:"thumb_media_id"`     // 图文消息的封面图片素材id（必须是永久mediaID）
	ShowCoverPic     int    `json:"show_cover_pic"`     // 是否显示封面，0为false，即不显示，1为true，即显示
	Author           string `json:"author"`             // 作者
	Digest           string `json:"digest"`             // 图文消息的摘要，仅有单图文消息才有摘要，多图文此处为空
	Content          string `json:"content"`            // 图文消息的具体内容，支持HTML标签，必须少于2万字符，小于1M，且此处会去除JS
	URL              string `json:"url"`                // 图文页的URL
	ContentSourceURL string `json:"content_source_url"` // 图文消息的原文地址，即点击"阅读原文"后的URL
}

// GetMaterialNewsResult 获取图文素材结果
type GetMaterialNewsResult struct {
	Resp
	NewsItem []*NewsItem `json:"news_item"` // 图文素材，内容
}

// GetMaterialVideoResult 获取视频素材结果
type GetMaterialVideoResult struct {
	Resp
	Title       string `json:"title"`       // 视频素材，标题
	Description string `json:"description"` // 视频素材，描述
	DownURL     string `json:"down_url"`    // 视频下载，地址
}

// BatchGetMaterialRequest 批量获取素材请求参数
type BatchGetMaterialRequest struct {
	Type   MaterialType `json:"type"`   // 素材的类型
	Offset int          `json:"offset"` // 从全部素材的该偏移位置开始返回，0表示从第一个素材返回
	Count  int          `json:"count"`  // 返回素材的数量，取值在1到20之间
}

// BatchNewsItem 批量获取图文素材条目
type BatchNewsItem struct {
	Title            string `json:"title"`              // 图文消息的标题
	Author           string `json:"author"`             // 作者
	Digest           string `json:"digest"`             // 图文消息的摘要，仅有单图文消息才有摘要，多图文此处为空
	Content          string `json:"content"`            // 图文消息的具体内容，支持HTML标签，必须少于2万字符，小于1M，且此处会去除JS
	ContentSourceURL string `json:"content_source_url"` // 图文消息的原文地址，即点击"阅读原文"后的URL
	ThumbMediaID     string `json:"thumb_media_id"`     // 图文消息的封面图片素材id（必须是永久mediaID）
	ShowCoverPic     int    `json:"show_cover_pic"`     // 是否显示封面，0为false，即不显示，1为true，即显示
	URL              string `json:"url"`                // 图文页的URL，或者，当获取的列表是图片素材列表时，该字段是图片的URL
	ThumbURL         string `json:"thumb_url"`          // 图文消息的封面图片素材id（必须是永久mediaID）
}

// BatchNewsContent 批量获取图文素材内容
type BatchNewsContent struct {
	NewsItem []*BatchNewsItem `json:"news_item"` // 图文消息内的1篇或多篇文章
}

// BatchMaterialItem 批量获取素材条目
type BatchMaterialItem struct {
	MediaID    string            `json:"media_id"`    // 消息ID
	Content    *BatchNewsContent `json:"content"`     // 图文消息，内容
	UpdateTime int64             `json:"update_time"` // 更新日期
	Name       string            `json:"name"`        // 图片、语音、视频素材的名字
	URL        string            `json:"url"`         // 图片、语音、视频素材URL
}

// BatchGetMaterialResult 批量获取素材结果
type BatchGetMaterialResult struct {
	Resp
	Item       []*BatchMaterialItem `json:"item"`        // 多个图文消息
	TotalCount int                  `json:"total_count"` // 该类型的素材的总数
	ItemCount  int                  `json:"item_count"`  // 本次调用获取的素材的数量
}

// UploadImgResult 上传图文消息图片结果
type UploadImgResult struct {
	Resp
	URL string `json:"url"` // 图片URL
}

// AddMaterialVideoDescription 新增视频素材描述信息
type AddMaterialVideoDescription struct {
	Title        string `json:"title"`        // 视频标题
	Introduction string `json:"introduction"` // 视频简介
}

// AddMaterialResult 新增素材结果
type AddMaterialResult struct {
	Resp
	MediaID string `json:"media_id"` // 新增的永久素材media_id
	URL     string `json:"url"`      // 图片素材URL(仅图片返回)
}

// DelMaterialRequest 删除素材请求参数
type DelMaterialRequest struct {
	MediaID string `json:"media_id"` // 要删除的素材media_id
}

// TempMediaType 临时素材类型
type TempMediaType string

const (
	TempMediaTypeImage TempMediaType = "image" // 图片
	TempMediaTypeVoice TempMediaType = "voice" // 语音
	TempMediaTypeVideo TempMediaType = "video" // 视频
	TempMediaTypeThumb TempMediaType = "thumb" // 缩略图
)

// UploadTempMediaResult 上传临时素材结果
type UploadTempMediaResult struct {
	Resp
	Type      TempMediaType `json:"type"`       // 媒体文件类型
	MediaID   string        `json:"media_id"`   // 媒体文件标识
	CreatedAt int64         `json:"created_at"` // 上传时间戳
}

// GetTempMediaVideoResult 获取临时视频素材结果
type GetTempMediaVideoResult struct {
	Resp
	VideoURL string `json:"video_url"` // 视频消息素材下载地址
}
