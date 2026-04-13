package pay

import (
	"context"
	"fmt"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
)

// Refunds submits a domestic refund request.
// 申请退款。文档：https://pay.weixin.qq.com/doc/v3/merchant/4012556589
func (c *Client) Refunds(ctx context.Context, refund *types.Refunds) (*types.RefundResp, error) {
	resp := &types.RefundResp{}
	if err := c.postV3(ctx, "/v3/refund/domestic/refunds", refund, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// QueryRefunds queries a single refund by the merchant out_refund_no.
// 通过商户退款单号查询单笔退款。
func (c *Client) QueryRefunds(ctx context.Context, outRefundNo string) (*types.RefundResp, error) {
	if outRefundNo == "" {
		return nil, fmt.Errorf("pay: outRefundNo is required")
	}
	urlPath := fmt.Sprintf("/v3/refund/domestic/refunds/%s", outRefundNo)
	resp := &types.RefundResp{}
	if err := c.getV3(ctx, urlPath, nil, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// ApplyAbnormalRefund 发起异常退款。
func (c *Client) ApplyAbnormalRefund(ctx context.Context, refundId string, refund *types.AbnormalRefund) (*types.RefundResp, error) {
	if refundId == "" {
		return nil, fmt.Errorf("pay: refundId is required")
	}
	urlPath := fmt.Sprintf("/v3/refund/domestic/refunds/%s/apply-abnormal-refund", refundId)
	resp := &types.RefundResp{}
	if err := c.postV3(ctx, urlPath, refund, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
