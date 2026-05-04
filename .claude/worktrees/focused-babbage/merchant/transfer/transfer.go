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
