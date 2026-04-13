package xiaowei

import "context"

// StoreInfo contains Xiaowei merchant store information.
type StoreInfo struct {
	StoreName   string `json:"store_name,omitempty"`
	StoreHead   string `json:"store_head_img,omitempty"` // logo URL
	StoreStatus int    `json:"store_status,omitempty"`  // 1=active, 2=suspended
}

// GetStoreInfoResp is the response from GetStoreInfo.
type GetStoreInfoResp struct {
	StoreInfo *StoreInfo `json:"store_info"`
}

// GetStoreInfo returns the merchant's Xiaowei store information.
func (c *Client) GetStoreInfo(ctx context.Context) (*GetStoreInfoResp, error) {
	var resp GetStoreInfoResp
	if err := c.doPost(ctx, "/wxaapi/wxamicrostore/get_store_info", struct{}{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateStoreInfoReq is the request to update store information.
type UpdateStoreInfoReq struct {
	StoreName string `json:"store_name,omitempty"`
	StoreHead string `json:"store_head_img,omitempty"`
}

// UpdateStoreInfo updates the merchant's store name and/or logo.
func (c *Client) UpdateStoreInfo(ctx context.Context, req *UpdateStoreInfoReq) error {
	return c.doPost(ctx, "/wxaapi/wxamicrostore/update_store_info", req, nil)
}

// GetKYCStatusResp is the response from GetKYCStatus.
type GetKYCStatusResp struct {
	Status int    `json:"kyc_status"` // 0=not submitted, 1=pending, 2=approved, 3=rejected
	Reason string `json:"reject_reason,omitempty"`
}

// GetKYCStatus returns the KYC (real-name verification) status of the merchant.
func (c *Client) GetKYCStatus(ctx context.Context) (*GetKYCStatusResp, error) {
	var resp GetKYCStatusResp
	if err := c.doPost(ctx, "/wxaapi/wxamicrostore/get_kyc_status", struct{}{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SubmitKYCReq is the request to submit KYC (real-name) verification.
type SubmitKYCReq struct {
	RealName    string `json:"real_name"`          // legal name of the individual merchant
	IDCardNo    string `json:"id_card_no"`         // national ID card number
	IDCardFront string `json:"id_card_front_img"`  // media_id of front photo
	IDCardBack  string `json:"id_card_back_img"`   // media_id of back photo
}

// SubmitKYC submits the merchant's real-name verification (KYC) documents.
func (c *Client) SubmitKYC(ctx context.Context, req *SubmitKYCReq) error {
	return c.doPost(ctx, "/wxaapi/wxamicrostore/submit_kyc", req, nil)
}
