package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// 服务商侧的分账接口封装。
//
// 服务商和直连商户使用同一套分账 REST 路径，区别只在请求体里要求带上
// sub_mchid/appid。本文件的方法：
//   - 复用 merchant/developed 的核心签名/验签逻辑；
//   - 自动把 Client 配置的默认 sub_mchid / sp_appid 填入请求体（调用方
//     显式提供的字段优先）；
//   - 对需要敏感字段加密的接口（例如 ReceiverType 为 PERSONAL_OPENID 时
//     的 name 字段）提供 *WithSerial 变体，支持携带 Wechatpay-Serial 头。
//
// 文档入口：https://pay.weixin.qq.com/docs/partner/apis/profitsharing/receivers/add-receivers.html

// ProfitSharingAddReceiver 添加分账接收方。
//
// body 至少需要包含 receiver 相关字段；appid / sub_mchid 若缺省则使用 Client
// 初始化时配置的 sp_appid / sub_mchid。如果 receiver.name 需要加密，请改用
// ProfitSharingAddReceiverWithSerial 并一并传入平台证书序列号。
func (c *Client) ProfitSharingAddReceiver(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.ProfitSharingAddReceiverWithSerial(ctx, body, "")
}

// ProfitSharingAddReceiverWithSerial 在 ProfitSharingAddReceiver 的基础上允许
// 指定 Wechatpay-Serial 头，用于接收方姓名（name 字段）已加密的场景。
func (c *Client) ProfitSharingAddReceiverWithSerial(ctx context.Context, body map[string]any, platformSerial string) (map[string]any, error) {
	if body == nil {
		return nil, errors.New("service: profit sharing add-receiver body is required")
	}
	merged := c.profitSharingFillDefaults(body)
	headers := serialHeader(platformSerial)
	result := map[string]any{}
	if err := c.inner.DoV3(ctx, http.MethodPost, "/v3/profitsharing/receivers/add", nil, merged, headers, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingDeleteReceiver 删除分账接收方。
// 同样会自动填充 appid / sub_mchid；delete 接口不涉及敏感字段，因此没有
// *WithSerial 变体。
func (c *Client) ProfitSharingDeleteReceiver(ctx context.Context, body map[string]any) (map[string]any, error) {
	if body == nil {
		return nil, errors.New("service: profit sharing delete-receiver body is required")
	}
	merged := c.profitSharingFillDefaults(body)
	result := map[string]any{}
	if err := c.inner.PostV3Raw(ctx, "/v3/profitsharing/receivers/delete", merged, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProfitSharingMerchantConfig 查询子商户的最大分账比例配置。
//
// subMchid 为空时使用 Client 初始化时配置的默认 sub_mchid。
// 返回体中的关键字段包括 max_ratio（单位：万分比，例如 2000 表示 20%）。
func (c *Client) ProfitSharingMerchantConfig(ctx context.Context, subMchid string) (map[string]any, error) {
	if subMchid == "" {
		subMchid = c.subMchid
	}
	if subMchid == "" {
		return nil, errors.New("service: subMchid is required")
	}
	path := fmt.Sprintf("/v3/profitsharing/merchant-configs/%s", subMchid)
	result := map[string]any{}
	if err := c.inner.GetV3Raw(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// profitSharingFillDefaults 在调用方未显式提供时，把 appid / sub_mchid 填成
// Client 配置的默认值。该方法不会修改入参，返回新的 map。
//
// 注意：profitsharing 接口只接受 appid / sub_mchid 字段（不接受 sp_appid /
// sp_mchid），因此不能直接复用 injectSubFields。
func (c *Client) profitSharingFillDefaults(body map[string]any) map[string]any {
	out := make(map[string]any, len(body)+2)
	for k, v := range body {
		out[k] = v
	}
	if _, ok := out["appid"]; !ok && c.inner.Appid() != "" {
		out["appid"] = c.inner.Appid()
	}
	if _, ok := out["sub_mchid"]; !ok && c.subMchid != "" {
		out["sub_mchid"] = c.subMchid
	}
	return out
}

// serialHeader 构造仅包含 Wechatpay-Serial 的 http.Header；serial 为空时返回 nil。
func serialHeader(serial string) http.Header {
	if serial == "" {
		return nil
	}
	return http.Header{"Wechatpay-Serial": []string{serial}}
}
