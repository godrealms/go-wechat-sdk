package wechat

import (
	"context"
	"fmt"
	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
	"github.com/godrealms/go-wechat-sdk/utils"
	"time"
)

func (c *Client) Refunds(refund *types.Refunds) (*types.RefundResp, error) {
	body := refund.ToString()
	path := "/v3/refund/domestic/refunds"
	nonceStr := utils.GenerateHashBasedString(32)
	timestamp := time.Now().Unix()
	sign := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", "POST", path, timestamp, nonceStr, body)
	signature, err := utils.SignSHA256WithRSA(sign, c.privateKey)
	if err != nil {
		return nil, err
	}
	c.Http.Headers = map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",signature=\"%s\",timestamp=\"%d\",serial_no=\"%s\"", c.Mchid, nonceStr, signature, timestamp, c.CertificateNumber),
	}

	response := &types.RefundResp{}
	err = c.Http.Post(context.Background(), path, refund, response)
	return response, err
}

// QueryRefunds 查询单笔退款（通过商户退款单号）
func (c *Client) QueryRefunds(outRefundNo string) (*types.RefundResp, error) {
	path := fmt.Sprintf("/v3/refund/domestic/refunds/%s", outRefundNo)
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

	response := &types.RefundResp{}
	err = c.Http.Get(context.Background(), path, nil, response)
	return response, err
}

// ApplyAbnormalRefund 发起异常退款
func (c *Client) ApplyAbnormalRefund(refundId string, refund *types.AbnormalRefund) (*types.RefundResp, error) {
	body := refund.ToString()
	path := fmt.Sprintf("/v3/refund/domestic/refunds/%s/apply-abnormal-refund", refundId)
	nonceStr := utils.GenerateHashBasedString(32)
	timestamp := time.Now().Unix()
	sign := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", "POST", path, timestamp, nonceStr, body)
	signature, err := utils.SignSHA256WithRSA(sign, c.privateKey)
	if err != nil {
		return nil, err
	}
	c.Http.Headers = map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",signature=\"%s\",timestamp=\"%d\",serial_no=\"%s\"", c.Mchid, nonceStr, signature, timestamp, c.CertificateNumber),
	}

	response := &types.RefundResp{}
	err = c.Http.Post(context.Background(), path, refund, response)
	return response, err
}
