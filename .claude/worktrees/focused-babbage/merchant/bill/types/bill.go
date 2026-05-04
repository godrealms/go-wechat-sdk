package types

// FundFlowBillRequest is the request for GetFundFlowBill
type FundFlowBillRequest struct {
	BillDate    string `json:"bill_date"`           // format: 2019-06-11
	AccountType string `json:"account_type"`        // BASIC/OPERATION/FEES
	TarType     string `json:"tar_type,omitempty"`  // GZIP
}

// BillDownloadResult is the result of any bill download API
type BillDownloadResult struct {
	HashType    string `json:"hash_type"`
	HashValue   string `json:"hash_value"`
	DownloadUrl string `json:"download_url"`
}

// SubMerchantFundFlowBillRequest is the request for GetSubMerchantFundFlowBill
type SubMerchantFundFlowBillRequest struct {
	SubMchid    string `json:"sub_mchid"`
	BillDate    string `json:"bill_date"`
	AccountType string `json:"account_type"`        // BASIC/OPERATION/FEES
	Algorithm   string `json:"algorithm"`           // AEAD_AES_256_GCM / SM4_GCM
	TarType     string `json:"tar_type,omitempty"`
}
