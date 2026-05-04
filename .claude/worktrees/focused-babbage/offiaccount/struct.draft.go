package offiaccount

// DraftSwitchResult 草稿箱开关设置结果
type DraftSwitchResult struct {
	Resp
	IsOpen int `json:"is_open"` // 仅 errcode==0 (即调用成功) 时返回，0 表示开关处于关闭，1 表示开启成功（或开关已开启）
}

// ArticleType 文章类型
type ArticleType string

const (
	ArticleTypeNews    ArticleType = "news"    // 图文消息
	ArticleTypeNewsPic ArticleType = "newspic" // 图片消息
)

// DraftArticle 草稿文章
type DraftArticle struct {
	ArticleType        ArticleType       `json:"article_type,omitempty"`          // 文章类型
	Title              string            `json:"title"`                           // 标题
	Author             string            `json:"author,omitempty"`                // 作者
	Digest             string            `json:"digest,omitempty"`                // 图文消息的摘要
	Content            string            `json:"content"`                         // 图文消息的具体内容
	ContentSourceURL   string            `json:"content_source_url,omitempty"`    // 图文消息的原文地址
	ThumbMediaID       string            `json:"thumb_media_id,omitempty"`        // 图文消息的封面图片素材id
	NeedOpenComment    int               `json:"need_open_comment,omitempty"`     // 是否打开评论，0不打开(默认)，1打开
	OnlyFansCanComment int               `json:"only_fans_can_comment,omitempty"` // 是否粉丝才可评论，0所有人可评论(默认)，1粉丝才可评论
	PicCrop2351        string            `json:"pic_crop_235_1,omitempty"`        // 图文消息封面裁剪为2.35:1规格的坐标字段
	PicCrop11          string            `json:"pic_crop_1_1,omitempty"`          // 图文消息封面裁剪为1:1规格的坐标字段
	ImageInfo          *DraftImageInfo   `json:"image_info,omitempty"`            // 图片消息里的图片相关信息
	CoverInfo          *DraftCoverInfo   `json:"cover_info,omitempty"`            // 图片消息的封面信息
	ProductInfo        *DraftProductInfo `json:"product_info,omitempty"`          // 商品信息
}

// DraftImageInfo 图片消息里的图片相关信息
type DraftImageInfo struct {
	ImageList []*DraftImage `json:"image_list"` // 图片列表
}

// DraftImage 图片列表项
type DraftImage struct {
	ImageMediaID string `json:"image_media_id"` // 图片消息里的图片素材id
}

// DraftCoverInfo 图片消息的封面信息
type DraftCoverInfo struct {
	CropPercentList []*DraftCropPercent `json:"crop_percent_list"` // 封面裁剪信息
}

// DraftCropPercent 封面裁剪信息
type DraftCropPercent struct {
	Ratio string `json:"ratio,omitempty"` // 裁剪比例
	X1    string `json:"x1,omitempty"`    // 左上角X坐标
	Y1    string `json:"y1,omitempty"`    // 左上角Y坐标
	X2    string `json:"x2,omitempty"`    // 右下角X坐标
	Y2    string `json:"y2,omitempty"`    // 右下角Y坐标
}

// DraftProductInfo 商品信息
type DraftProductInfo struct {
	FooterProductInfo *DraftFooterProductInfo `json:"footer_product_info"` // 文末插入商品相关信息
}

// DraftFooterProductInfo 文末插入商品相关信息
type DraftFooterProductInfo struct {
	ProductKey string `json:"product_key"` // 商品key
}

// AddDraftResult 新增草稿结果
type AddDraftResult struct {
	Resp
	MediaID string `json:"media_id"` // 上传后的获取标志
}

// GetDraftResult 获取草稿结果
type GetDraftResult struct {
	Resp
	NewsItem []*DraftNewsItem `json:"news_item"` // 图文素材列表
}

// DraftNewsItem 图文素材列表项
type DraftNewsItem struct {
	ArticleType        ArticleType       `json:"article_type"`          // 文章类型
	Title              string            `json:"title"`                 // 标题
	Author             string            `json:"author"`                // 作者
	Digest             string            `json:"digest"`                // 图文消息的摘要
	Content            string            `json:"content"`               // 图文消息的具体内容
	ContentSourceURL   string            `json:"content_source_url"`    // 图文消息的原文地址
	ThumbMediaID       string            `json:"thumb_media_id"`        // 图文消息的封面图片素材id
	NeedOpenComment    int               `json:"need_open_comment"`     // 是否打开评论
	OnlyFansCanComment int               `json:"only_fans_can_comment"` // 是否粉丝才可评论
	ImageInfo          *DraftImageInfo   `json:"image_info"`            // 图片消息里的图片相关信息
	ProductInfo        *DraftProductInfo `json:"product_info"`          // 商品信息
	URL                string            `json:"url"`                   // 草稿的临时链接
}

// UpdateDraftRequest 更新草稿请求参数
type UpdateDraftRequest struct {
	MediaID  string        `json:"media_id"` // 要修改的图文消息的id
	Index    int           `json:"index"`    // 要更新的文章在图文消息中的位置
	Articles *DraftArticle `json:"articles"` // 图文信息
}

// DraftCountResult 获取草稿总数结果
type DraftCountResult struct {
	Resp
	TotalCount int `json:"total_count"` // 草稿的总数
}

// BatchGetDraftRequest 批量获取草稿请求参数
type BatchGetDraftRequest struct {
	Offset    int `json:"offset"`     // 从全部素材的该偏移位置开始返回，0表示从第一个素材返回
	Count     int `json:"count"`      // 返回素材的数量，取值在1到20之间
	NoContent int `json:"no_content"` // 1 表示不返回 content 字段，0 表示正常返回，默认为 0
}

// BatchGetDraftResult 批量获取草稿结果
type BatchGetDraftResult struct {
	Resp
	TotalCount int              `json:"total_count"` // 草稿素材的总数
	ItemCount  int              `json:"item_count"`  // 本次调用获取的素材的数量
	Item       []*DraftListItem `json:"item"`        // 素材列表
}

// DraftListItem 素材列表项
type DraftListItem struct {
	MediaID    string        `json:"media_id"`    // 图文消息的id
	Content    *DraftContent `json:"content"`     // 图文消息内容
	UpdateTime int64         `json:"update_time"` // 图文消息更新时间
}

// DraftContent 图文消息内容
type DraftContent struct {
	NewsItem []*DraftNewsItem `json:"news_item"` // 图文内容列表
}

// DeleteDraftRequest 删除草稿请求参数
type DeleteDraftRequest struct {
	MediaID string `json:"media_id"` // 要删除的草稿的media_id
}

// ProductCardType 卡片类型
type ProductCardType int

const (
	ProductCardTypeLarge ProductCardType = 0 // 大卡
	ProductCardTypeSmall ProductCardType = 1 // 小卡
	ProductCardTypeText  ProductCardType = 2 // 文字链接
	ProductCardTypeStrip ProductCardType = 3 // 条卡
)

// GetProductCardInfoRequest 获取商品卡片信息请求参数
type GetProductCardInfoRequest struct {
	ProductID   string          `json:"product_id"`   // 商品id
	ArticleType ArticleType     `json:"article_type"` // 文章类型
	CardType    ProductCardType `json:"card_type"`    // 卡片类型
}

// GetProductCardInfoResult 获取商品卡片信息结果
type GetProductCardInfoResult struct {
	Resp
	ProductKey string `json:"product_key"` // 商品 key
	DOM        string `json:"DOM"`         // 商品卡DOM结构
}
