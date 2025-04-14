package wechat

import (
	"context"
	"fmt"
	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
	"github.com/godrealms/go-wechat-sdk/utils"
	"time"
)

// TransactionsClose 关闭订单
func (c *Client) TransactionsClose(outTradeNo string) error {
	request := &types.MchID{Mchid: c.Mchid}
	body := request.ToString()
	path := fmt.Sprintf("/v3/pay/transactions/out-trade-no/%s/close", outTradeNo)
	nonceStr := utils.GenerateHashBasedString(32)
	timestamp := time.Now().Unix()
	sign := fmt.Sprintf("%s\n%s\n%d\n%s\n%s\n", "POST", path, timestamp, nonceStr, body) // 签名体
	signature, err := utils.SignSHA256WithRSA(sign, c.privateKey)
	if err != nil {
		return err
	}
	c.Http.Headers = map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("WECHATPAY2-SHA256-RSA2048 mchid=\"%s\",nonce_str=\"%s\",signature=\"%s\",timestamp=\"%d\",serial_no=\"%s\"", c.Mchid, nonceStr, signature, timestamp, c.CertificateNumber),
	}
	return c.Http.Post(context.Background(), path, body, nil)
}
