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
