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
