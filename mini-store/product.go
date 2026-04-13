package mini_store

import "context"

// Product represents a Mini Store product listing.
type Product struct {
	ProductID   string   `json:"product_id,omitempty"`
	Title       string   `json:"title"`
	SubTitle    string   `json:"sub_title,omitempty"`
	HeadImgs    []string `json:"head_imgs,omitempty"`
	Description string   `json:"desc_info,omitempty"`
	Status      int      `json:"status,omitempty"`
}

// AddProductResp is the response from AddProduct.
type AddProductResp struct {
	ProductID string `json:"product_id"`
}

// AddProduct creates a new product and returns its product_id.
func (c *Client) AddProduct(ctx context.Context, product *Product) (*AddProductResp, error) {
	var resp AddProductResp
	if err := c.doPost(ctx, "/shop/spu/add", product, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DelProductReq is the request to delete a product.
type DelProductReq struct {
	ProductID string `json:"product_id"`
}

// DelProduct deletes the product identified by product_id.
func (c *Client) DelProduct(ctx context.Context, req *DelProductReq) error {
	return c.doPost(ctx, "/shop/spu/del", req, nil)
}

// UpdateProductReq is the request to update a product.
type UpdateProductReq struct {
	ProductID string   `json:"product_id"`
	Product   *Product `json:"spu_info"`
}

// UpdateProduct updates an existing product.
func (c *Client) UpdateProduct(ctx context.Context, req *UpdateProductReq) error {
	return c.doPost(ctx, "/shop/spu/update", req, nil)
}

// GetProductReq is the request to get a single product.
type GetProductReq struct {
	ProductID string `json:"product_id"`
}

// GetProductResp is the response from GetProduct.
type GetProductResp struct {
	SPU *Product `json:"spu"`
}

// GetProduct returns the details of a single product.
func (c *Client) GetProduct(ctx context.Context, req *GetProductReq) (*GetProductResp, error) {
	var resp GetProductResp
	if err := c.doPost(ctx, "/shop/spu/get", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListProductsReq is the request to list products.
type ListProductsReq struct {
	Status   int `json:"status,omitempty"`
	PageSize int `json:"page_size,omitempty"`
	Page     int `json:"page,omitempty"`
}

// ListProductsResp is the response from ListProducts.
type ListProductsResp struct {
	SPUs     []*Product `json:"spus"`
	TotalNum int        `json:"total_num"`
}

// ListProducts returns a paginated list of products.
func (c *Client) ListProducts(ctx context.Context, req *ListProductsReq) (*ListProductsResp, error) {
	var resp ListProductsResp
	if err := c.doPost(ctx, "/shop/spu/get_list", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateProductStatusReq is the request to change product listing status.
type UpdateProductStatusReq struct {
	ProductID string `json:"product_id"`
	Status    int    `json:"status"` // 0=off-shelf, 1=on-shelf
}

// UpdateProductStatus changes the on-shelf/off-shelf status of a product.
func (c *Client) UpdateProductStatus(ctx context.Context, req *UpdateProductStatusReq) error {
	return c.doPost(ctx, "/shop/spu/update_without_audit", req, nil)
}

// SubmitProductAuditReq is the request to submit a product for audit.
type SubmitProductAuditReq struct {
	ProductID string `json:"product_id"`
}

// SubmitProductAudit submits a product for platform audit before listing.
func (c *Client) SubmitProductAudit(ctx context.Context, req *SubmitProductAuditReq) error {
	return c.doPost(ctx, "/shop/audit/audit_spu", req, nil)
}

// CancelProductAuditReq is the request to withdraw a pending audit.
type CancelProductAuditReq struct {
	ProductID string `json:"product_id"`
}

// CancelProductAudit withdraws a product audit submission.
func (c *Client) CancelProductAudit(ctx context.Context, req *CancelProductAuditReq) error {
	return c.doPost(ctx, "/shop/audit/cancel_audit_spu", req, nil)
}
