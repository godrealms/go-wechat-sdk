package wechat

import (
	"context"
	"fmt"
	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
	"github.com/godrealms/go-wechat-sdk/utils"
	"time"
)

// QueryTransactionId 微信支付订单号查询订单
func (c *Client) QueryTransactionId(transactionId string) (*types.QueryResponse, error) {
	path := fmt.Sprintf("/v3/pay/transactions/id/%s?mchid=%s", transactionId, c.Mchid)
	nonceStr := utils.GenerateHashBasedString(32)
	timestamp := time.Now().Unix()
	sign := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", "GET", path, timestamp, nonceStr, "")
	signature, err := utils.SignSHA256WithRSA(sign, c.privateKey)
	if err != nil {
		return nil, err
	}
	c.Http.Headers = map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",signature=\"%s\",timestamp=\"%d\",serial_no=\"%s\"", c.Mchid, nonceStr, signature, timestamp, c.CertificateNumber),
	}

	response := &types.QueryResponse{}
	err = c.Http.Get(context.Background(), path, nil, response)
	return response, nil
}
func (c *Client) QueryOutTradeNo(outTradeNo string) (*types.QueryResponse, error) {
	path := fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s?mchid=%s", outTradeNo, c.Mchid)
	nonceStr := utils.GenerateHashBasedString(32)
	timestamp := time.Now().Unix()
	sign := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", "GET", path, timestamp, nonceStr, "")
	signature, err := utils.SignSHA256WithRSA(sign, c.privateKey)
	if err != nil {
		return nil, err
	}
	c.Http.Headers = map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",signature=\"%s\",timestamp=\"%d\",serial_no=\"%s\"", c.Mchid, nonceStr, signature, timestamp, c.CertificateNumber),
	}

	response := &types.QueryResponse{}
	err = c.Http.Get(context.Background(), path, nil, response)
	return response, err
}
