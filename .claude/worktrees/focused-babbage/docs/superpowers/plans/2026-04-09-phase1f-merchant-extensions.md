# Phase 1F: Merchant Sub-packages (certificate, profitsharing, transfer, bill)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add four new merchant sub-packages wrapping the existing developed.Client: certificate (platform cert download), profitsharing (分账), transfer (商家转账), and bill (账单/资金账单).

**Architecture:** Each sub-package has its own `client.go` (thin wrapper over `*developed.Client`) plus `types/` directory for request/response structs, and one or more API files. All use the existing WECHATPAY2-SHA256-RSA2048 signing pattern from the developed package.

**Tech Stack:** Go 1.23.1, standard library only. Depends on `merchant/developed/` package being intact.

**PREREQUISITE:** `merchant/developed/` package must compile cleanly (no prerequisite plan needed — it already exists).

---

## Task 1: Add `GetPrivateKey()` accessor to `merchant/developed/client.go`

**Files:**
- MODIFY `merchant/developed/client.go`

The `privateKey` field on `developed.Client` is unexported. Sub-packages that embed `*developed.Client` cannot access it directly. Add a public accessor method.

- [ ] Open `merchant/developed/client.go` and add the following method to the `Client` type:

```go
// GetPrivateKey returns the merchant private key (used by sub-packages)
func (c *Client) GetPrivateKey() *rsa.PrivateKey {
    return c.privateKey
}
```

Make sure the `crypto/rsa` import is present at the top of the file (it almost certainly already is, since `privateKey` is of type `*rsa.PrivateKey`).

- [ ] Verify the package still compiles:

```bash
go build ./merchant/developed/...
```

- [ ] Commit:

```bash
git add merchant/developed/client.go
git commit -m "feat(merchant/developed): add GetPrivateKey() accessor for sub-package access"
```

---

## Task 2: Create `merchant/certificate/`

**Files:**
- CREATE `merchant/certificate/client.go`
- CREATE `merchant/certificate/types/certificate.go`
- CREATE `merchant/certificate/certificates.go`

This package downloads the WeChat Pay platform certificate list. The platform certificate is required to verify WeChat Pay response signatures.

- [ ] Create directory structure:

```bash
mkdir -p merchant/certificate/types
```

- [ ] Create `merchant/certificate/client.go`:

```go
package certificate

import wechat "github.com/godrealms/go-wechat-sdk/merchant/developed"

// Client is the WeChat Pay certificate client
type Client struct {
	*wechat.Client
}

// NewClient creates a certificate client wrapping an existing developed Client
func NewClient(c *wechat.Client) *Client {
	return &Client{Client: c}
}
```

- [ ] Create `merchant/certificate/types/certificate.go`:

```go
package types

// PlatformCertificate represents one platform certificate
type PlatformCertificate struct {
	SerialNo           string              `json:"serial_no"`
	EffectiveTime      string              `json:"effective_time"`
	ExpireTime         string              `json:"expire_time"`
	EncryptCertificate *EncryptCertificate `json:"encrypt_certificate"`
}

// EncryptCertificate holds the encrypted certificate data
type EncryptCertificate struct {
	Algorithm      string `json:"algorithm"`      // AEAD_AES_256_GCM
	Nonce          string `json:"nonce"`
	AssociatedData string `json:"associated_data"`
	CipherText     string `json:"ciphertext"`
}

// GetCertificatesResult is the result of GetCertificates
type GetCertificatesResult struct {
	Data []*PlatformCertificate `json:"data"`
}
```

- [ ] Create `merchant/certificate/certificates.go`:

```go
package certificate

import (
	"context"
	"fmt"
	"time"

	"github.com/godrealms/go-wechat-sdk/merchant/certificate/types"
	"github.com/godrealms/go-wechat-sdk/utils"
)

// GetCertificates 获取平台证书列表
// GET /v3/certificates
func (c *Client) GetCertificates() (*types.GetCertificatesResult, error) {
	path := "/v3/certificates"
	nonceStr := utils.GenerateHashBasedString(32)
	timestamp := time.Now().Unix()
	sign := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", "GET", path, timestamp, nonceStr, "")
	signature, err := utils.SignSHA256WithRSA(sign, c.GetPrivateKey())
	if err != nil {
		return nil, err
	}
	headers := map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%d",serial_no="%s"`, c.Mchid, nonceStr, signature, timestamp, c.CertificateNumber),
	}
	result := &types.GetCertificatesResult{}
	err = c.Http.GetWithHeaders(context.Background(), path, headers, nil, result)
	return result, err
}
```

- [ ] Verify the package compiles:

```bash
go build ./merchant/certificate/...
```

- [ ] Commit:

```bash
git add merchant/certificate/
git commit -m "feat(merchant/certificate): add platform certificate download sub-package"
```

---

## Task 3: Create `merchant/profitsharing/`

**Files:**
- CREATE `merchant/profitsharing/client.go`
- CREATE `merchant/profitsharing/types/profitsharing.go`
- CREATE `merchant/profitsharing/profitsharing.go`

This package wraps the WeChat Pay v3 profit-sharing (分账) APIs: create order, query order, return order, unfreeze remaining funds, and query unsplit amount.

- [ ] Create directory structure:

```bash
mkdir -p merchant/profitsharing/types
```

- [ ] Create `merchant/profitsharing/client.go`:

```go
package profitsharing

import wechat "github.com/godrealms/go-wechat-sdk/merchant/developed"

// Client is the WeChat Pay profit-sharing client
type Client struct {
	*wechat.Client
}

// NewClient creates a profitsharing client wrapping an existing developed Client
func NewClient(c *wechat.Client) *Client {
	return &Client{Client: c}
}
```

- [ ] Create `merchant/profitsharing/types/profitsharing.go`:

```go
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
```

- [ ] Create `merchant/profitsharing/profitsharing.go`:

```go
package profitsharing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/godrealms/go-wechat-sdk/merchant/profitsharing/types"
	"github.com/godrealms/go-wechat-sdk/utils"
)

func (c *Client) buildHeaders(method, path, body string) (map[string]string, error) {
	nonceStr := utils.GenerateHashBasedString(32)
	timestamp := time.Now().Unix()
	sign := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", method, path, timestamp, nonceStr, body)
	signature, err := utils.SignSHA256WithRSA(sign, c.GetPrivateKey())
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%d",serial_no="%s"`, c.Mchid, nonceStr, signature, timestamp, c.CertificateNumber),
	}, nil
}

// CreateProfitsharing 请求分账
// POST /v3/profitsharing/orders
func (c *Client) CreateProfitsharing(req *types.ProfitsharingRequest) (*types.ProfitsharingResult, error) {
	path := "/v3/profitsharing/orders"
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	headers, err := c.buildHeaders("POST", path, string(bodyBytes))
	if err != nil {
		return nil, err
	}
	result := &types.ProfitsharingResult{}
	err = c.Http.PostWithHeaders(context.Background(), path, req, headers, result)
	return result, err
}

// QueryProfitsharing 查询分账结果
// GET /v3/profitsharing/orders/{out_order_no}?transaction_id={transaction_id}
func (c *Client) QueryProfitsharing(transactionId, outOrderNo string) (*types.QueryProfitsharingResult, error) {
	path := fmt.Sprintf("/v3/profitsharing/orders/%s?transaction_id=%s", outOrderNo, transactionId)
	headers, err := c.buildHeaders("GET", path, "")
	if err != nil {
		return nil, err
	}
	result := &types.QueryProfitsharingResult{}
	err = c.Http.GetWithHeaders(context.Background(), path, headers, nil, result)
	return result, err
}

// ReturnProfitsharing 请求分账回退
// POST /v3/profitsharing/return-orders
func (c *Client) ReturnProfitsharing(req *types.ReturnProfitsharingRequest) (*types.ReturnProfitsharingResult, error) {
	path := "/v3/profitsharing/return-orders"
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	headers, err := c.buildHeaders("POST", path, string(bodyBytes))
	if err != nil {
		return nil, err
	}
	result := &types.ReturnProfitsharingResult{}
	err = c.Http.PostWithHeaders(context.Background(), path, req, headers, result)
	return result, err
}

// UnfreezeProfitsharing 解冻剩余资金
// POST /v3/profitsharing/orders/unfreeze
func (c *Client) UnfreezeProfitsharing(req *types.UnfreezeRequest) (*types.ProfitsharingResult, error) {
	path := "/v3/profitsharing/orders/unfreeze"
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	headers, err := c.buildHeaders("POST", path, string(bodyBytes))
	if err != nil {
		return nil, err
	}
	result := &types.ProfitsharingResult{}
	err = c.Http.PostWithHeaders(context.Background(), path, req, headers, result)
	return result, err
}

// GetProfitsharingUnsplitAmount 查询订单待分账金额
// GET /v3/profitsharing/transactions-amount/{transaction_id}
func (c *Client) GetProfitsharingUnsplitAmount(transactionId string) (int64, error) {
	path := fmt.Sprintf("/v3/profitsharing/transactions-amount/%s", transactionId)
	headers, err := c.buildHeaders("GET", path, "")
	if err != nil {
		return 0, err
	}
	result := &struct {
		TransactionId string `json:"transaction_id"`
		UnsplitAmount int64  `json:"unsplit_amount"`
	}{}
	err = c.Http.GetWithHeaders(context.Background(), path, headers, nil, result)
	if err != nil {
		return 0, err
	}
	return result.UnsplitAmount, nil
}
```

- [ ] Verify the package compiles:

```bash
go build ./merchant/profitsharing/...
```

- [ ] Commit:

```bash
git add merchant/profitsharing/
git commit -m "feat(merchant/profitsharing): add profit-sharing sub-package (分账)"
```

---

## Task 4: Create `merchant/transfer/`

**Files:**
- CREATE `merchant/transfer/client.go`
- CREATE `merchant/transfer/types/transfer.go`
- CREATE `merchant/transfer/transfer.go`

This package wraps the WeChat Pay v3 merchant transfer (商家转账) batch APIs: create batch transfer, query batch by batch ID, and query batch by out-batch number.

- [ ] Create directory structure:

```bash
mkdir -p merchant/transfer/types
```

- [ ] Create `merchant/transfer/client.go`:

```go
package transfer

import wechat "github.com/godrealms/go-wechat-sdk/merchant/developed"

// Client is the WeChat Pay merchant transfer client
type Client struct {
	*wechat.Client
}

// NewClient creates a transfer client wrapping an existing developed Client
func NewClient(c *wechat.Client) *Client {
	return &Client{Client: c}
}
```

- [ ] Create `merchant/transfer/types/transfer.go`:

```go
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
```

- [ ] Create `merchant/transfer/transfer.go`:

```go
package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/godrealms/go-wechat-sdk/merchant/transfer/types"
	"github.com/godrealms/go-wechat-sdk/utils"
)

func (c *Client) buildHeaders(method, path, body string) (map[string]string, error) {
	nonceStr := utils.GenerateHashBasedString(32)
	timestamp := time.Now().Unix()
	sign := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", method, path, timestamp, nonceStr, body)
	signature, err := utils.SignSHA256WithRSA(sign, c.GetPrivateKey())
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%d",serial_no="%s"`, c.Mchid, nonceStr, signature, timestamp, c.CertificateNumber),
	}, nil
}

// CreateTransferBatch 发起商家转账
// POST /v3/transfer/batches
func (c *Client) CreateTransferBatch(req *types.TransferBatchRequest) (*types.TransferBatchResult, error) {
	path := "/v3/transfer/batches"
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	headers, err := c.buildHeaders("POST", path, string(bodyBytes))
	if err != nil {
		return nil, err
	}
	result := &types.TransferBatchResult{}
	err = c.Http.PostWithHeaders(context.Background(), path, req, headers, result)
	return result, err
}

// QueryTransferBatchByBatchId 通过微信批次单号查询批次单
// GET /v3/transfer/batches/batch-id/{batch_id}?need_query_detail={bool}&offset={n}&limit={n}
func (c *Client) QueryTransferBatchByBatchId(batchId string, needQueryDetail bool, offset, limit int) (*types.QueryTransferBatchResult, error) {
	path := fmt.Sprintf("/v3/transfer/batches/batch-id/%s?need_query_detail=%v&offset=%d&limit=%d", batchId, needQueryDetail, offset, limit)
	headers, err := c.buildHeaders("GET", path, "")
	if err != nil {
		return nil, err
	}
	result := &types.QueryTransferBatchResult{}
	err = c.Http.GetWithHeaders(context.Background(), path, headers, nil, result)
	return result, err
}

// QueryTransferBatchByOutBatchNo 通过商家批次单号查询批次单
// GET /v3/transfer/batches/out-batch-no/{out_batch_no}?need_query_detail={bool}&offset={n}&limit={n}
func (c *Client) QueryTransferBatchByOutBatchNo(outBatchNo string, needQueryDetail bool, offset, limit int) (*types.QueryTransferBatchResult, error) {
	path := fmt.Sprintf("/v3/transfer/batches/out-batch-no/%s?need_query_detail=%v&offset=%d&limit=%d", outBatchNo, needQueryDetail, offset, limit)
	headers, err := c.buildHeaders("GET", path, "")
	if err != nil {
		return nil, err
	}
	result := &types.QueryTransferBatchResult{}
	err = c.Http.GetWithHeaders(context.Background(), path, headers, nil, result)
	return result, err
}
```

- [ ] Verify the package compiles:

```bash
go build ./merchant/transfer/...
```

- [ ] Commit:

```bash
git add merchant/transfer/
git commit -m "feat(merchant/transfer): add merchant transfer batch sub-package (商家转账)"
```

---

## Task 5: Create `merchant/bill/` and run full build verification

**Files:**
- CREATE `merchant/bill/client.go`
- CREATE `merchant/bill/types/bill.go`
- CREATE `merchant/bill/bill.go`

This package adds the fund flow bill (`/v3/bill/fundflowbill`) and sub-merchant fund flow bill (`/v3/bill/sub-merchant-fundflowbill`) endpoints. Note: the existing `merchant/developed/` package already handles trade bills (`/v3/bill/tradebill`) and basic fund flow via `pay.transactions.bill.go`. This new package extends coverage to the remaining bill API endpoints not present in the developed package.

- [ ] Create directory structure:

```bash
mkdir -p merchant/bill/types
```

- [ ] Create `merchant/bill/client.go`:

```go
package bill

import wechat "github.com/godrealms/go-wechat-sdk/merchant/developed"

// Client is the WeChat Pay bill client
type Client struct {
	*wechat.Client
}

// NewClient creates a bill client wrapping an existing developed Client
func NewClient(c *wechat.Client) *Client {
	return &Client{Client: c}
}
```

- [ ] Create `merchant/bill/types/bill.go`:

```go
package types

// FundFlowBillRequest is the request for GetFundFlowBill
type FundFlowBillRequest struct {
	BillDate    string `json:"bill_date"`            // format: 2019-06-11
	AccountType string `json:"account_type"`         // BASIC/OPERATION/FEES
	TarType     string `json:"tar_type,omitempty"`   // GZIP
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
	AccountType string `json:"account_type"`         // BASIC/OPERATION/FEES
	Algorithm   string `json:"algorithm"`            // AEAD_AES_256_GCM / SM4_GCM
	TarType     string `json:"tar_type,omitempty"`
}
```

- [ ] Create `merchant/bill/bill.go`:

```go
package bill

import (
	"context"
	"fmt"
	"time"

	"github.com/godrealms/go-wechat-sdk/merchant/bill/types"
	"github.com/godrealms/go-wechat-sdk/utils"
)

func (c *Client) buildHeaders(method, path, body string) (map[string]string, error) {
	nonceStr := utils.GenerateHashBasedString(32)
	timestamp := time.Now().Unix()
	sign := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", method, path, timestamp, nonceStr, body)
	signature, err := utils.SignSHA256WithRSA(sign, c.GetPrivateKey())
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",signature="%s",timestamp="%d",serial_no="%s"`, c.Mchid, nonceStr, signature, timestamp, c.CertificateNumber),
	}, nil
}

// GetFundFlowBill 申请资金账单
// GET /v3/bill/fundflowbill?bill_date={date}&account_type={type}[&tar_type=GZIP]
func (c *Client) GetFundFlowBill(req *types.FundFlowBillRequest) (*types.BillDownloadResult, error) {
	path := fmt.Sprintf("/v3/bill/fundflowbill?bill_date=%s&account_type=%s", req.BillDate, req.AccountType)
	if req.TarType != "" {
		path += "&tar_type=" + req.TarType
	}
	headers, err := c.buildHeaders("GET", path, "")
	if err != nil {
		return nil, err
	}
	result := &types.BillDownloadResult{}
	err = c.Http.GetWithHeaders(context.Background(), path, headers, nil, result)
	return result, err
}

// GetSubMerchantFundFlowBill 申请单个子商户资金账单
// GET /v3/bill/sub-merchant-fundflowbill?sub_mchid={id}&bill_date={date}&account_type={type}&algorithm={alg}[&tar_type=GZIP]
func (c *Client) GetSubMerchantFundFlowBill(req *types.SubMerchantFundFlowBillRequest) (*types.BillDownloadResult, error) {
	path := fmt.Sprintf("/v3/bill/sub-merchant-fundflowbill?sub_mchid=%s&bill_date=%s&account_type=%s&algorithm=%s", req.SubMchid, req.BillDate, req.AccountType, req.Algorithm)
	if req.TarType != "" {
		path += "&tar_type=" + req.TarType
	}
	headers, err := c.buildHeaders("GET", path, "")
	if err != nil {
		return nil, err
	}
	result := &types.BillDownloadResult{}
	err = c.Http.GetWithHeaders(context.Background(), path, headers, nil, result)
	return result, err
}
```

- [ ] Verify the package compiles:

```bash
go build ./merchant/bill/...
```

- [ ] Run full merchant build verification to confirm all sub-packages compile together without errors:

```bash
go build ./merchant/...
```

- [ ] Commit:

```bash
git add merchant/bill/
git commit -m "feat(merchant/bill): add fund flow bill sub-package (资金账单/子商户账单)"
```

---

## Summary

After all 5 tasks complete, the following new packages will exist under `merchant/`:

| Package | API Endpoints |
|---|---|
| `merchant/certificate` | GET /v3/certificates |
| `merchant/profitsharing` | POST /v3/profitsharing/orders, GET /v3/profitsharing/orders/{no}, POST /v3/profitsharing/return-orders, POST /v3/profitsharing/orders/unfreeze, GET /v3/profitsharing/transactions-amount/{id} |
| `merchant/transfer` | POST /v3/transfer/batches, GET /v3/transfer/batches/batch-id/{id}, GET /v3/transfer/batches/out-batch-no/{no} |
| `merchant/bill` | GET /v3/bill/fundflowbill, GET /v3/bill/sub-merchant-fundflowbill |

All packages follow the same pattern: thin `Client` struct embedding `*developed.Client`, a `buildHeaders` helper using the WECHATPAY2-SHA256-RSA2048 signing scheme, and a `types/` sub-package for request/response structs. Zero external dependencies.
