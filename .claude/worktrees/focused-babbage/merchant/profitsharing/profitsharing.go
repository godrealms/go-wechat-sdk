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
