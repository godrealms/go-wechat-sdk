package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// 服务商侧的分账「动态接口」— 即与具体订单 / 回退单 / 剩余金额相关的接口。
//
// 这一组接口与直连商户共享相同的 REST 路径（例如 /v3/profitsharing/orders），
// 服务商模式的差别只在：
//  1. 请求体需要 appid + sub_mchid；
//  2. 部分 GET 接口需要通过 query 带上 sub_mchid。
//
// 因此本文件的实现策略：
//   - POST 类：复用 profitSharingFillDefaults 自动注入 appid / sub_mchid；
//   - GET 类：在 url.Values 里注入 sub_mchid，调用方可显式传参覆盖默认值；
//   - 仍然复用 merchant/developed 包的签名 / 验签核心逻辑。
//
// 静态接口（添加/删除分账接收方、查询最大分账比例）仍由 profit_sharing.go
// 提供，本文件不重复实现。

// ProfitSharingOrder 创建分账单（服务商版）。
//
// body 通常需要包含 transaction_id / out_order_no / receivers / unfreeze_unsplit
// 等字段。appid / sub_mchid 若缺省则使用 Client 初始化时配置的默认值；
// 若 receivers 中含有加密过的 name 字段，请改用 ProfitSharingOrderWithSerial
// 并传入平台证书序列号。
func (c *Client) ProfitSharingOrder(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.ProfitSharingOrderWithSerial(ctx, body, "")
}

// ProfitSharingOrderWithSerial 与 ProfitSharingOrder 相同，但允许通过
// platformSerial 指定 Wechatpay-Serial 头（当 receivers 中含有已加密字段时必填）。
func (c *Client) ProfitSharingOrderWithSerial(
	ctx context.Context,
	body map[string]any,
	platformSerial string,
) (map[string]any, error) {
	if body == nil {
		return nil, errors.New("service: profit sharing order body is required")
	}
	merged := c.profitSharingFillDefaults(body)
	result := map[string]any{}
	if err := c.inner.DoV3(
		ctx, http.MethodPost, "/v3/profitsharing/orders", nil, merged,
		serialHeader(platformSerial), &result,
	); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingQueryOrder 查询分账单（服务商版）。
//
// subMchid 为空时使用默认子商户号；outOrderNo、transactionId 均为必填。
func (c *Client) ProfitSharingQueryOrder(
	ctx context.Context,
	subMchid, outOrderNo, transactionId string,
) (map[string]any, error) {
	if outOrderNo == "" || transactionId == "" {
		return nil, errors.New("service: outOrderNo and transactionId are required")
	}
	if subMchid == "" {
		subMchid = c.subMchid
	}
	if subMchid == "" {
		return nil, errors.New("service: subMchid is required")
	}
	path := fmt.Sprintf("/v3/profitsharing/orders/%s", outOrderNo)
	query := url.Values{
		"sub_mchid":      {subMchid},
		"transaction_id": {transactionId},
	}
	result := map[string]any{}
	if err := c.inner.GetV3Raw(ctx, path, query, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingReturn 请求分账回退（服务商版）。sub_mchid / appid 若缺省
// 则使用 Client 初始化时配置的默认值。
func (c *Client) ProfitSharingReturn(ctx context.Context, body map[string]any) (map[string]any, error) {
	if body == nil {
		return nil, errors.New("service: profit sharing return body is required")
	}
	merged := c.profitSharingFillDefaults(body)
	result := map[string]any{}
	if err := c.inner.PostV3Raw(ctx, "/v3/profitsharing/return-orders", merged, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingQueryReturn 查询分账回退（服务商版）。
//
// subMchid 为空时使用默认子商户号；outReturnNo、outOrderNo 均为必填。
func (c *Client) ProfitSharingQueryReturn(
	ctx context.Context,
	subMchid, outReturnNo, outOrderNo string,
) (map[string]any, error) {
	if outReturnNo == "" || outOrderNo == "" {
		return nil, errors.New("service: outReturnNo and outOrderNo are required")
	}
	if subMchid == "" {
		subMchid = c.subMchid
	}
	if subMchid == "" {
		return nil, errors.New("service: subMchid is required")
	}
	path := fmt.Sprintf("/v3/profitsharing/return-orders/%s", outReturnNo)
	query := url.Values{
		"sub_mchid":    {subMchid},
		"out_order_no": {outOrderNo},
	}
	result := map[string]any{}
	if err := c.inner.GetV3Raw(ctx, path, query, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingUnfreeze 解冻分账剩余金额（服务商版）。
// sub_mchid / appid 若缺省则使用 Client 初始化时配置的默认值。
func (c *Client) ProfitSharingUnfreeze(ctx context.Context, body map[string]any) (map[string]any, error) {
	if body == nil {
		return nil, errors.New("service: profit sharing unfreeze body is required")
	}
	merged := c.profitSharingFillDefaults(body)
	result := map[string]any{}
	if err := c.inner.PostV3Raw(ctx, "/v3/profitsharing/orders/unfreeze", merged, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingMerchantAmounts 查询某笔交易的剩余待分金额（服务商版）。
// 路径里的 transactionId 为必填；subMchid 为空时使用默认子商户号。
func (c *Client) ProfitSharingMerchantAmounts(
	ctx context.Context,
	subMchid, transactionId string,
) (map[string]any, error) {
	if transactionId == "" {
		return nil, errors.New("service: transactionId is required")
	}
	if subMchid == "" {
		subMchid = c.subMchid
	}
	if subMchid == "" {
		return nil, errors.New("service: subMchid is required")
	}
	path := fmt.Sprintf("/v3/profitsharing/transactions/%s/amounts", transactionId)
	query := url.Values{"sub_mchid": {subMchid}}
	result := map[string]any{}
	if err := c.inner.GetV3Raw(ctx, path, query, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingBills 申请分账账单（按日，服务商版）。
//
// 参数：
//   - billDate：账单日期 YYYY-MM-DD，必填。
//   - subMchid：指定子商户号。为空时使用默认子商户号；
//     如果想拉服务商全量账单，请传一个特殊占位符或改用 Inner().ProfitSharingBills
//     直接调用直连商户版。
//   - tarType：压缩类型，可选，传空不启用。
func (c *Client) ProfitSharingBills(
	ctx context.Context,
	billDate, subMchid, tarType string,
) (map[string]any, error) {
	if billDate == "" {
		return nil, errors.New("service: billDate is required")
	}
	if subMchid == "" {
		subMchid = c.subMchid
	}
	query := url.Values{"bill_date": {billDate}}
	if subMchid != "" {
		query.Set("sub_mchid", subMchid)
	}
	if tarType != "" {
		query.Set("tar_type", tarType)
	}
	result := map[string]any{}
	if err := c.inner.GetV3Raw(ctx, "/v3/profitsharing/bills", query, &result); err != nil {
		return nil, err
	}
	return result, nil
}
