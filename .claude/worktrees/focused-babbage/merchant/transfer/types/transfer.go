package types

// TransferBatchRequest is the request for CreateTransferBatch
type TransferBatchRequest struct {
	Appid              string            `json:"appid"`
	OutBatchNo         string            `json:"out_batch_no"`
	BatchName          string            `json:"batch_name"`
	BatchRemark        string            `json:"batch_remark"`
	TotalAmount        int64             `json:"total_amount"`
	TotalNum           int               `json:"total_num"`
	TransferDetailList []*TransferDetail `json:"transfer_detail_list"`
	TransferScene      string            `json:"transfer_scene,omitempty"` // 现金营销 等
}

// TransferDetail represents one transfer recipient
type TransferDetail struct {
	OutDetailNo    string `json:"out_detail_no"`
	TransferAmount int64  `json:"transfer_amount"` // 单位分
	TransferRemark string `json:"transfer_remark"`
	OpenId         string `json:"openid"`
	UserName       string `json:"user_name,omitempty"` // 收款方真实姓名(加密)
}

// TransferBatchResult is the result of CreateTransferBatch
type TransferBatchResult struct {
	OutBatchNo string `json:"out_batch_no"`
	BatchId    string `json:"batch_id"`
	CreateTime string `json:"create_time"`
}

// QueryTransferBatchResult is the result of QueryTransferBatch
type QueryTransferBatchResult struct {
	TransferBatch      *TransferBatchInfo    `json:"transfer_batch"`
	TransferDetailList []*TransferDetailInfo `json:"transfer_detail_list,omitempty"`
}

// TransferBatchInfo contains batch-level information
type TransferBatchInfo struct {
	Mchid         string `json:"mchid"`
	OutBatchNo    string `json:"out_batch_no"`
	BatchId       string `json:"batch_id"`
	Appid         string `json:"appid"`
	BatchStatus   string `json:"batch_status"` // ACCEPTED/PROCESSING/FINISHED/CLOSED
	BatchType     string `json:"batch_type"`
	BatchName     string `json:"batch_name"`
	BatchRemark   string `json:"batch_remark"`
	CloseReason   string `json:"close_reason,omitempty"`
	TotalAmount   int64  `json:"total_amount"`
	TotalNum      int    `json:"total_num"`
	SendNum       int    `json:"send_num,omitempty"`
	SuccessAmount int64  `json:"success_amount,omitempty"`
	SuccessNum    int    `json:"success_num,omitempty"`
	FailAmount    int64  `json:"fail_amount,omitempty"`
	FailNum       int    `json:"fail_num,omitempty"`
	UpdateTime    string `json:"update_time,omitempty"`
	CreateTime    string `json:"create_time"`
	AuthDueTo     string `json:"auth_due_to,omitempty"`
	TransferScene string `json:"transfer_scene,omitempty"`
}

// TransferDetailInfo contains detail-level information
type TransferDetailInfo struct {
	DetailId       string `json:"detail_id"`
	OutDetailNo    string `json:"out_detail_no"`
	DetailStatus   string `json:"detail_status"` // PROCESSING/SUCCESS/FAIL
	TransferAmount int64  `json:"transfer_amount"`
	TransferRemark string `json:"transfer_remark"`
	FailReason     string `json:"fail_reason,omitempty"`
	OpenId         string `json:"openid"`
	UserName       string `json:"user_name,omitempty"`
	InitiateTime   string `json:"initiate_time,omitempty"`
	UpdateTime     string `json:"update_time,omitempty"`
}
