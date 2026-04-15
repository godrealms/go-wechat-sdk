package isv

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// safetyMargin 是 token 过期前的安全窗口,避免临界时刻使用即将过期的 token。
const safetyMargin = 5 * time.Minute

// clampTokenTTL floors the raw `expires_in` returned by WeChat before we
// subtract safetyMargin. WeChat normally returns 7200s; a hostile or malformed
// upstream could return a small or negative value, which after subtracting
// safetyMargin would place expiresAt in the past or inside the safetyMargin
// window, causing a refresh storm on every call. We floor raw to
// 3*safetyMargin so the effective TTL (raw - safetyMargin) is at least
// 2*safetyMargin — strictly greater than safetyMargin, the threshold at which
// readValidSuiteToken/readValidCorpToken treats a token as stale.
func clampTokenTTL(expiresInSeconds int) time.Duration {
	raw := time.Duration(expiresInSeconds) * time.Second
	min := 3 * safetyMargin
	if raw < min {
		raw = min
	}
	return raw - safetyMargin
}

// GetSuiteAccessToken 返回 suite_access_token。
// 策略:lazy + 双检锁。Store 命中且未到安全窗口直接返回;否则加锁再查一次,仍过期则刷新。
func (c *Client) GetSuiteAccessToken(ctx context.Context) (string, error) {
	if tok, ok, err := c.readValidSuiteToken(ctx); err != nil {
		return "", err
	} else if ok {
		return tok, nil
	}

	c.suiteMu.Lock()
	defer c.suiteMu.Unlock()

	if tok, ok, err := c.readValidSuiteToken(ctx); err != nil {
		return "", err
	} else if ok {
		return tok, nil
	}

	return c.fetchSuiteTokenLocked(ctx)
}

// RefreshSuiteToken 强制刷新 suite_access_token(忽略缓存)。
func (c *Client) RefreshSuiteToken(ctx context.Context) error {
	c.suiteMu.Lock()
	defer c.suiteMu.Unlock()
	_, err := c.fetchSuiteTokenLocked(ctx)
	return err
}

func (c *Client) readValidSuiteToken(ctx context.Context) (string, bool, error) {
	tok, exp, err := c.store.GetSuiteToken(ctx, c.cfg.SuiteID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", false, nil
		}
		return "", false, err
	}
	if time.Until(exp) <= safetyMargin {
		return "", false, nil
	}
	return tok, true, nil
}

// fetchSuiteTokenLocked 发起一次 HTTP,写回 Store,返回新 token。
// 调用方必须已持有 c.suiteMu。
func (c *Client) fetchSuiteTokenLocked(ctx context.Context) (string, error) {
	ticket, err := c.store.GetSuiteTicket(ctx, c.cfg.SuiteID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", ErrSuiteTicketMissing
		}
		return "", fmt.Errorf("isv: read suite_ticket: %w", err)
	}

	body := map[string]string{
		"suite_id":     c.cfg.SuiteID,
		"suite_secret": c.cfg.SuiteSecret,
		"suite_ticket": ticket,
	}
	var resp SuiteAccessTokenResp
	// get_suite_token 不需要附带 access_token query
	if err := c.doPostRaw(ctx, "/cgi-bin/service/get_suite_token", url.Values{}, body, &resp); err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(clampTokenTTL(resp.ExpiresIn))
	if err := c.store.PutSuiteToken(ctx, c.cfg.SuiteID, resp.SuiteAccessToken, expiresAt); err != nil {
		return "", fmt.Errorf("isv: persist suite_token: %w", err)
	}
	return resp.SuiteAccessToken, nil
}
