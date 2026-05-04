package types

// ProfitsharingReceiver represents one profit-sharing receiver
type ProfitsharingReceiver struct {
	Type        string `json:"type"`                  // MERCHANT_ID / PERSONAL_OPENID / PERSONAL_SUB_OPENID
	Account     string `json:"account"`
	Amount      int64  `json:"amount"`                // 分账金额，单位分
	Description string `json:"description"`
	Name        string `json:"name,omitempty"`
	RoleType    string `json:"role_type,omitempty"`   // STORE / STAFF / STORE_OWNER / PARTNER / HEADQUARTER / BRAND / DISTRIBUTOR / USER / SUPPLIER
}

// ProfitsharingRequest is the request for CreateProfitsharing
type ProfitsharingRequest struct {
	Appid           string                   `json:"appid"`
	TransactionId   string                   `json:"transaction_id"`
	OutOrderNo      string                   `json:"out_order_no"`
	Receivers       []*ProfitsharingReceiver `json:"receivers"`
	UnfreezeUnsplit bool                     `json:"unfreeze_unsplit"`
}

// ProfitsharingResult is the result of CreateProfitsharing
type ProfitsharingResult struct {
	Appid         string                   `json:"appid"`
	TransactionId string                   `json:"transaction_id"`
	OutOrderNo    string                   `json:"out_order_no"`
	OrderId       string                   `json:"order_id"`
	State         string                   `json:"state"` // PROCESSING / FINISHED
	Receivers     []*ProfitsharingReceiver `json:"receivers"`
}

// QueryProfitsharingResult is the result of QueryProfitsharing
type QueryProfitsharingResult struct {
	ProfitsharingResult
	FinishAmount int64  `json:"finish_amount"`
	FinishDesc   string `json:"finish_description"`
}

// ReturnProfitsharingRequest is the request for ReturnProfitsharing
type ReturnProfitsharingRequest struct {
	OrderId     string `json:"order_id"`
	OutReturnNo string `json:"out_return_no"`
	ReturnMchid string `json:"return_mchid"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

// ReturnProfitsharingResult is the result of ReturnProfitsharing
type ReturnProfitsharingResult struct {
	OrderId     string `json:"order_id"`
	OutReturnNo string `json:"out_return_no"`
	ReturnId    string `json:"return_id"`
	ReturnMchid string `json:"return_mchid"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	Result      string `json:"result"`               // PROCESSING / SUCCESS / FAILED
	FailReason  string `json:"fail_reason,omitempty"`
	FinishTime  string `json:"finish_time,omitempty"`
}

// UnfreezeRequest is the request for UnfreezeProfitsharing
type UnfreezeRequest struct {
	TransactionId string                   `json:"transaction_id"`
	OutOrderNo    string                   `json:"out_order_no"`
	Receivers     []*ProfitsharingReceiver `json:"receivers"`
	Description   string                   `json:"description"`
}
