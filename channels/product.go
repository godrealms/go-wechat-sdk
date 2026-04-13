package channels

import "context"

// ---------- 商品管理 ----------

type ProductInfo struct {
	ProductID  string   `json:"product_id,omitempty"`
	Title      string   `json:"title"`
	SubTitle   string   `json:"sub_title,omitempty"`
	HeadImgs   []string `json:"head_imgs,omitempty"`
	Status     *int     `json:"status,omitempty"`
	CreateTime int64    `json:"create_time,omitempty"`
}

type AddProductReq struct {
	Product ProductInfo `json:"product"`
}
type AddProductResp struct {
	ProductID string `json:"product_id"`
}

type UpdateProductReq struct {
	Product ProductInfo `json:"product"`
}

type GetProductReq struct {
	ProductID string `json:"product_id"`
}
type GetProductResp struct {
	Product ProductInfo `json:"product"`
}

type ListProductReq struct {
	Status *int `json:"status,omitempty"`
	Offset *int `json:"offset,omitempty"`
	Limit  *int `json:"limit,omitempty"`
}
type ListProductResp struct {
	Products []ProductInfo `json:"products"`
	Total    int           `json:"total"`
}

type DeleteProductReq struct {
	ProductID string `json:"product_id"`
}

// AddProduct 添加商品
func (c *Client) AddProduct(ctx context.Context, req *AddProductReq) (*AddProductResp, error) {
	var resp AddProductResp
	if err := c.doPost(ctx, "/channels/ec/product/add", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateProduct 更新商品
func (c *Client) UpdateProduct(ctx context.Context, req *UpdateProductReq) error {
	return c.doPost(ctx, "/channels/ec/product/update", req, nil)
}

// GetProduct 获取商品详情
func (c *Client) GetProduct(ctx context.Context, req *GetProductReq) (*GetProductResp, error) {
	var resp GetProductResp
	if err := c.doPost(ctx, "/channels/ec/product/get", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListProduct 获取商品列表
func (c *Client) ListProduct(ctx context.Context, req *ListProductReq) (*ListProductResp, error) {
	var resp ListProductResp
	if err := c.doPost(ctx, "/channels/ec/product/list", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteProduct 删除商品
func (c *Client) DeleteProduct(ctx context.Context, req *DeleteProductReq) error {
	return c.doPost(ctx, "/channels/ec/product/delete", req, nil)
}
