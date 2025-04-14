package wechat

import (
	"context"
	"fmt"
	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
	"github.com/godrealms/go-wechat-sdk/utils"
	"time"
)

// TradeBill 申请交易账单
func (c *Client) TradeBill(quest *types.TradeBillQuest) (*types.BillResp, error) {
	values := quest.ToUrlValues()
	path := fmt.Sprintf("/v3/bill/tradebill?%s", values.Encode())
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

	path = "/v3/bill/tradebill"
	response := &types.BillResp{}
	err = c.Http.Get(context.Background(), path, values, response)
	return response, err
}

func (c *Client) FundFlowBill(quest *types.FundsBillQuest) (*types.BillResp, error) {
	values := quest.ToUrlValues()
	path := fmt.Sprintf("/v3/bill/fundflowbill?%s", values.Encode())
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

	path = "/v3/bill/fundflowbill"
	response := &types.BillResp{}
	err = c.Http.Get(context.Background(), path, values, response)
	return response, err
}
