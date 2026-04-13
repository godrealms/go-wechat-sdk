package mini_store

import "context"

// GetMerchantInfoResp is the response from GetMerchantInfo.
type GetMerchantInfoResp struct {
	MerchantInfo map[string]interface{} `json:"merchant_info"`
}

// GetMerchantInfo returns the merchant registration and settlement information.
func (c *Client) GetMerchantInfo(ctx context.Context) (*GetMerchantInfoResp, error) {
	var resp GetMerchantInfoResp
	if err := c.doPost(ctx, "/shop/merchant/get_merchant_info", struct{}{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetSettlementReq is the request to get settlement details for a period.
type GetSettlementReq struct {
	StartTime int64 `json:"start_time"` // Unix timestamp
	EndTime   int64 `json:"end_time"`
}

// GetSettlementResp is the response from GetSettlement.
type GetSettlementResp struct {
	SettlementList []map[string]interface{} `json:"settlement_list"`
}

// GetSettlement returns the settlement records for the specified time range.
func (c *Client) GetSettlement(ctx context.Context, req *GetSettlementReq) (*GetSettlementResp, error) {
	var resp GetSettlementResp
	if err := c.doPost(ctx, "/shop/pay/get_pay_list", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetBrandListResp is the response from GetBrandList.
type GetBrandListResp struct {
	BrandList []map[string]interface{} `json:"brand_list"`
}

// GetBrandList returns the list of available brand categories for product listing.
func (c *Client) GetBrandList(ctx context.Context) (*GetBrandListResp, error) {
	var resp GetBrandListResp
	if err := c.doPost(ctx, "/shop/account/get_brand_list", struct{}{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetCategoryListReq is the request to get category list.
type GetCategoryListReq struct {
	ParentCatID int `json:"f_cat_id,omitempty"` // 0 for root categories
}

// GetCategoryListResp is the response from GetCategoryList.
type GetCategoryListResp struct {
	CategoryList []map[string]interface{} `json:"cat_list"`
}

// GetCategoryList returns product categories, optionally filtered by parent.
func (c *Client) GetCategoryList(ctx context.Context, req *GetCategoryListReq) (*GetCategoryListResp, error) {
	var resp GetCategoryListResp
	if err := c.doPost(ctx, "/shop/cat/get", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
