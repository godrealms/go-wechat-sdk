package pay

import (
	"context"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
)

// TradeBill 申请交易账单。
func (c *Client) TradeBill(ctx context.Context, quest *types.TradeBillQuest) (*types.BillResp, error) {
	resp := &types.BillResp{}
	if err := c.getV3(ctx, "/v3/bill/tradebill", quest.ToUrlValues(), resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// FundFlowBill 申请资金账单。
func (c *Client) FundFlowBill(ctx context.Context, quest *types.FundsBillQuest) (*types.BillResp, error) {
	resp := &types.BillResp{}
	if err := c.getV3(ctx, "/v3/bill/fundflowbill", quest.ToUrlValues(), resp); err != nil {
		return nil, err
	}
	return resp, nil
}
