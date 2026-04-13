package mini_store

import "context"

// MerchantInfo holds registration and settlement details for a merchant.
type MerchantInfo struct {
	Name             string `json:"name,omitempty"`
	MerchantID       string `json:"merchant_id,omitempty"`
	Status           int    `json:"status,omitempty"`
	SettlementBankNo string `json:"settlement_bank_no,omitempty"`
}

// GetMerchantInfoResp is the response from GetMerchantInfo.
type GetMerchantInfoResp struct {
	MerchantInfo *MerchantInfo `json:"merchant_info"`
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

// Settlement holds a single settlement record.
type Settlement struct {
	ID           string `json:"id,omitempty"`
	SettleTime   int64  `json:"settle_time,omitempty"`
	SettleAmount int64  `json:"settle_amount,omitempty"`
	Status       int    `json:"status,omitempty"`
}

// GetSettlementResp is the response from GetSettlement.
type GetSettlementResp struct {
	SettlementList []Settlement `json:"settlement_list"`
}

// GetSettlement returns the settlement records for the specified time range.
func (c *Client) GetSettlement(ctx context.Context, req *GetSettlementReq) (*GetSettlementResp, error) {
	var resp GetSettlementResp
	if err := c.doPost(ctx, "/shop/pay/get_pay_list", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Brand represents a brand entry returned by GetBrandList.
type Brand struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// GetBrandListResp is the response from GetBrandList.
type GetBrandListResp struct {
	BrandList []Brand `json:"brand_list"`
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

// Category represents a product category returned by GetCategoryList.
type Category struct {
	CatID int    `json:"f_cat_id,omitempty"`
	Name  string `json:"name,omitempty"`
}

// GetCategoryListResp is the response from GetCategoryList.
type GetCategoryListResp struct {
	CategoryList []Category `json:"cat_list"`
}

// GetCategoryList returns product categories, optionally filtered by parent.
func (c *Client) GetCategoryList(ctx context.Context, req *GetCategoryListReq) (*GetCategoryListResp, error) {
	var resp GetCategoryListResp
	if err := c.doPost(ctx, "/shop/cat/get", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
