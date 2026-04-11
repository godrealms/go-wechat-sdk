package oplatform

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// componentTokenSafetyWindow 提前过期 60 秒，避免临界抖动。
const componentTokenSafetyWindow = 60 * time.Second

// ComponentAccessToken 读取 Store 中的 component_access_token；
// 过期则调用微信接口刷新并写回 Store。
func (c *Client) ComponentAccessToken(ctx context.Context) (string, error) {
	ctx = touchContext(ctx)

	// 先乐观读一次（无锁）
	if tok, expireAt, err := c.store.GetComponentToken(ctx); err == nil {
		if time.Now().Add(componentTokenSafetyWindow).Before(expireAt) {
			return tok, nil
		}
	} else if !errors.Is(err, ErrNotFound) {
		return "", fmt.Errorf("oplatform: store get component token: %w", err)
	}

	c.componentMu.Lock()
	defer c.componentMu.Unlock()

	// 双重检查
	if tok, expireAt, err := c.store.GetComponentToken(ctx); err == nil {
		if time.Now().Add(componentTokenSafetyWindow).Before(expireAt) {
			return tok, nil
		}
	} else if !errors.Is(err, ErrNotFound) {
		return "", fmt.Errorf("oplatform: store get component token: %w", err)
	}

	return c.fetchComponentTokenLocked(ctx)
}

// RefreshComponentToken 无视缓存强制刷新。
func (c *Client) RefreshComponentToken(ctx context.Context) error {
	ctx = touchContext(ctx)
	c.componentMu.Lock()
	defer c.componentMu.Unlock()
	_, err := c.fetchComponentTokenLocked(ctx)
	return err
}

// fetchComponentTokenLocked 必须在 componentMu 持锁时调用。
func (c *Client) fetchComponentTokenLocked(ctx context.Context) (string, error) {
	ticket, err := c.store.GetVerifyTicket(ctx)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", ErrVerifyTicketMissing
		}
		return "", fmt.Errorf("oplatform: store get verify ticket: %w", err)
	}

	body := map[string]string{
		"component_appid":         c.cfg.ComponentAppID,
		"component_appsecret":     c.cfg.ComponentAppSecret,
		"component_verify_ticket": ticket,
	}
	var resp componentTokenResp
	if err := c.http.Post(ctx, "/cgi-bin/component/api_component_token", body, &resp); err != nil {
		return "", fmt.Errorf("oplatform: api_component_token: %w", err)
	}
	if err := checkWeixinErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return "", err
	}
	if resp.ComponentAccessToken == "" {
		return "", fmt.Errorf("oplatform: empty component_access_token")
	}

	expireAt := time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)
	if err := c.store.SetComponentToken(ctx, resp.ComponentAccessToken, expireAt); err != nil {
		return "", fmt.Errorf("oplatform: store set component token: %w", err)
	}
	return resp.ComponentAccessToken, nil
}
