package channels

import (
	"context"
	"fmt"
)

// ProductInfo contains the metadata of a single Channels e-commerce product.
type ProductInfo struct {
	ProductID  string   `json:"product_id,omitempty"`
	Title      string   `json:"title"`
	SubTitle   string   `json:"sub_title,omitempty"`
	HeadImgs   []string `json:"head_imgs,omitempty"`
	Status     *int     `json:"status,omitempty"`
	CreateTime int64    `json:"create_time,omitempty"`
}

// AddProductReq holds the product information for creating a new Channels product.
type AddProductReq struct {
	Product ProductInfo `json:"product"`
}

// AddProductResp is the response returned by AddProduct.
type AddProductResp struct {
	ProductID string `json:"product_id"`
}

// UpdateProductReq holds the product information for updating an existing Channels product.
type UpdateProductReq struct {
	Product ProductInfo `json:"product"`
}

// GetProductReq holds the product ID for querying a single Channels product.
type GetProductReq struct {
	ProductID string `json:"product_id"`
}

// GetProductResp is the response returned by GetProduct.
type GetProductResp struct {
	Product ProductInfo `json:"product"`
}

// ListProductReq holds the filter and pagination parameters for listing Channels products.
type ListProductReq struct {
	Status *int `json:"status,omitempty"`
	Offset *int `json:"offset,omitempty"`
	Limit  *int `json:"limit,omitempty"`
}

// ListProductResp is the response returned by ListProduct.
type ListProductResp struct {
	Products []ProductInfo `json:"products"`
	Total    int           `json:"total"`
}

// DeleteProductReq holds the product ID for deleting a Channels product.
type DeleteProductReq struct {
	ProductID string `json:"product_id"`
}

// AddProduct creates a new Channels e-commerce product and returns its assigned product ID.
func (c *Client) AddProduct(ctx context.Context, req *AddProductReq) (*AddProductResp, error) {
	if req == nil {
		return nil, fmt.Errorf("channels: AddProduct: req is required")
	}
	if req.Product.Title == "" {
		return nil, fmt.Errorf("channels: AddProduct: req.Product.Title is required")
	}
	var resp AddProductResp
	if err := c.doPost(ctx, "/channels/ec/product/add", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateProduct updates the details of an existing Channels e-commerce product.
func (c *Client) UpdateProduct(ctx context.Context, req *UpdateProductReq) error {
	if req == nil {
		return fmt.Errorf("channels: UpdateProduct: req is required")
	}
	if req.Product.ProductID == "" {
		return fmt.Errorf("channels: UpdateProduct: req.Product.ProductID is required")
	}
	return c.doPost(ctx, "/channels/ec/product/update", req, nil)
}

// GetProduct retrieves the details of the specified Channels e-commerce product.
func (c *Client) GetProduct(ctx context.Context, req *GetProductReq) (*GetProductResp, error) {
	if req == nil || req.ProductID == "" {
		return nil, fmt.Errorf("channels: GetProduct: req.ProductID is required")
	}
	var resp GetProductResp
	if err := c.doPost(ctx, "/channels/ec/product/get", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListProduct retrieves a paginated list of Channels e-commerce products.
func (c *Client) ListProduct(ctx context.Context, req *ListProductReq) (*ListProductResp, error) {
	if req == nil {
		req = &ListProductReq{} // null-safe — list with default pagination
	}
	var resp ListProductResp
	if err := c.doPost(ctx, "/channels/ec/product/list", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteProduct removes the specified Channels e-commerce product.
func (c *Client) DeleteProduct(ctx context.Context, req *DeleteProductReq) error {
	if req == nil || req.ProductID == "" {
		return fmt.Errorf("channels: DeleteProduct: req.ProductID is required")
	}
	return c.doPost(ctx, "/channels/ec/product/delete", req, nil)
}
