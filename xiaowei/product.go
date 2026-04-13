package xiaowei

import "context"

// MicroProduct represents a Xiaowei product item.
type MicroProduct struct {
	ProductID string   `json:"product_id,omitempty"`
	Title     string   `json:"title"`
	Price     int64    `json:"price"`            // price in fen
	Stock     int      `json:"stock,omitempty"`
	ImgURLs   []string `json:"img_urls,omitempty"`
}

// AddMicroProductResp is the response from AddMicroProduct.
type AddMicroProductResp struct {
	ProductID string `json:"product_id"`
}

// AddMicroProduct adds a new product to the Xiaowei store.
func (c *Client) AddMicroProduct(ctx context.Context, product *MicroProduct) (*AddMicroProductResp, error) {
	var resp AddMicroProductResp
	if err := c.doPost(ctx, "/wxaapi/wxamicrostore/add_product", product, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DelMicroProductReq is the request to delete a product.
type DelMicroProductReq struct {
	ProductID string `json:"product_id"`
}

// DelMicroProduct removes a product from the Xiaowei store.
func (c *Client) DelMicroProduct(ctx context.Context, req *DelMicroProductReq) error {
	return c.doPost(ctx, "/wxaapi/wxamicrostore/del_product", req, nil)
}

// GetMicroProductReq is the request to get product details.
type GetMicroProductReq struct {
	ProductID string `json:"product_id"`
}

// GetMicroProductResp is the response from GetMicroProduct.
type GetMicroProductResp struct {
	Product *MicroProduct `json:"product"`
}

// GetMicroProduct returns the details of a Xiaowei product.
func (c *Client) GetMicroProduct(ctx context.Context, req *GetMicroProductReq) (*GetMicroProductResp, error) {
	var resp GetMicroProductResp
	if err := c.doPost(ctx, "/wxaapi/wxamicrostore/get_product", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListMicroProductsReq is the request to list products.
type ListMicroProductsReq struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
}

// ListMicroProductsResp is the response from ListMicroProducts.
type ListMicroProductsResp struct {
	Products []*MicroProduct `json:"product_list"`
	Total    int             `json:"total"`
}

// ListMicroProducts returns a paginated list of Xiaowei store products.
func (c *Client) ListMicroProducts(ctx context.Context, req *ListMicroProductsReq) (*ListMicroProductsResp, error) {
	var resp ListMicroProductsResp
	if err := c.doPost(ctx, "/wxaapi/wxamicrostore/get_product_list", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
