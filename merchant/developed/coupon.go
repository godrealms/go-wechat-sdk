package pay

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// 代金券（商家券 / Favor）接口：
// https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter9_5_1.shtml

// FavorCreateStock 创建代金券批次。
func (c *Client) FavorCreateStock(ctx context.Context, body any) (map[string]any, error) {
	result := map[string]any{}
	if err := c.postV3(ctx, "/v3/marketing/favor/coupon-stocks", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// FavorStartStock 激活代金券批次。
func (c *Client) FavorStartStock(ctx context.Context, stockId string, body any) (map[string]any, error) {
	if stockId == "" {
		return nil, fmt.Errorf("pay: stockId is required")
	}
	urlPath := fmt.Sprintf("/v3/marketing/favor/stocks/%s/start", stockId)
	result := map[string]any{}
	if err := c.postV3(ctx, urlPath, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// FavorPauseStock 暂停代金券批次。
func (c *Client) FavorPauseStock(ctx context.Context, stockId string, body any) error {
	if stockId == "" {
		return fmt.Errorf("pay: stockId is required")
	}
	urlPath := fmt.Sprintf("/v3/marketing/favor/stocks/%s/pause", stockId)
	return c.postV3(ctx, urlPath, body, nil)
}

// FavorRestartStock 重启代金券批次。
func (c *Client) FavorRestartStock(ctx context.Context, stockId string, body any) error {
	if stockId == "" {
		return fmt.Errorf("pay: stockId is required")
	}
	urlPath := fmt.Sprintf("/v3/marketing/favor/stocks/%s/restart", stockId)
	return c.postV3(ctx, urlPath, body, nil)
}

// FavorSendCoupon 发放指定批次代金券。
func (c *Client) FavorSendCoupon(ctx context.Context, openid string, body any) (map[string]any, error) {
	if openid == "" {
		return nil, fmt.Errorf("pay: openid is required")
	}
	urlPath := fmt.Sprintf("/v3/marketing/favor/users/%s/coupons", openid)
	result := map[string]any{}
	if err := c.postV3(ctx, urlPath, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// FavorQueryStock 查询批次详情。
func (c *Client) FavorQueryStock(ctx context.Context, stockId, stockCreatorMchid string) (map[string]any, error) {
	if stockId == "" || stockCreatorMchid == "" {
		return nil, fmt.Errorf("pay: stockId and stockCreatorMchid are required")
	}
	urlPath := fmt.Sprintf("/v3/marketing/favor/stocks/%s", stockId)
	query := url.Values{"stock_creator_mchid": {stockCreatorMchid}}
	result := map[string]any{}
	if err := c.doV3(ctx, http.MethodGet, urlPath, query, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// FavorQueryUserCoupon 查询用户单张代金券。
func (c *Client) FavorQueryUserCoupon(ctx context.Context, openid, couponId, appid string) (map[string]any, error) {
	if openid == "" || couponId == "" || appid == "" {
		return nil, fmt.Errorf("pay: openid/couponId/appid are required")
	}
	urlPath := fmt.Sprintf("/v3/marketing/favor/users/%s/coupons/%s", openid, couponId)
	query := url.Values{"appid": {appid}}
	result := map[string]any{}
	if err := c.doV3(ctx, http.MethodGet, urlPath, query, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
