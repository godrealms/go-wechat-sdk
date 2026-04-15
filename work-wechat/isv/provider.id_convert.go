package isv

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// getProviderAccessToken 返回 provider_access_token(lazy + 双检锁)。
// 校验 ProviderCorpID / ProviderSecret 是否配置,未配置返回哨兵错误。
func (c *Client) getProviderAccessToken(ctx context.Context) (string, error) {
	if c.cfg.ProviderCorpID == "" {
		return "", ErrProviderCorpIDMissing
	}
	if c.cfg.ProviderSecret == "" {
		return "", ErrProviderSecretMissing
	}

	if tok, ok, err := c.readValidProviderToken(ctx); err != nil {
		return "", err
	} else if ok {
		return tok, nil
	}

	c.providerMu.Lock()
	defer c.providerMu.Unlock()

	if tok, ok, err := c.readValidProviderToken(ctx); err != nil {
		return "", err
	} else if ok {
		return tok, nil
	}

	return c.fetchProviderTokenLocked(ctx)
}

func (c *Client) readValidProviderToken(ctx context.Context) (string, bool, error) {
	tok, exp, err := c.store.GetProviderToken(ctx, c.cfg.SuiteID)
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

func (c *Client) fetchProviderTokenLocked(ctx context.Context) (string, error) {
	body := map[string]string{
		"corpid":          c.cfg.ProviderCorpID,
		"provider_secret": c.cfg.ProviderSecret,
	}
	var resp struct {
		ProviderAccessToken string `json:"provider_access_token"`
		ExpiresIn           int    `json:"expires_in"`
	}
	if err := c.doPostRaw(ctx, "/cgi-bin/service/get_provider_token", url.Values{}, body, &resp); err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(clampTokenTTL(resp.ExpiresIn))
	if err := c.store.PutProviderToken(ctx, c.cfg.SuiteID, resp.ProviderAccessToken, expiresAt); err != nil {
		return "", fmt.Errorf("isv: persist provider_token: %w", err)
	}
	return resp.ProviderAccessToken, nil
}

// providerDoPost 和 doPost 类似,只是注入的 token 是 provider_access_token。
func (c *Client) providerDoPost(ctx context.Context, path string, body, out interface{}) error {
	tok, err := c.getProviderAccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"provider_access_token": {tok}}
	return c.doPostRaw(ctx, path, q, body, out)
}

// providerDoGet 和 doGet 类似,只是注入的 token 是 provider_access_token。
// 不能复用 c.doGet —— 后者会自动注入 suite_access_token。
func (c *Client) providerDoGet(ctx context.Context, path string, extra url.Values, out interface{}) error {
	tok, err := c.getProviderAccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"provider_access_token": {tok}}
	for k, vs := range extra {
		q[k] = vs
	}
	return c.doRequestRaw(ctx, http.MethodGet, path, q, nil, out)
}

// CorpIDToOpenCorpID 把企业 corpid 转换成跨服务商匿名的 open_corpid。
func (c *Client) CorpIDToOpenCorpID(ctx context.Context, corpID string) (string, error) {
	body := map[string]string{"corpid": corpID}
	var resp struct {
		OpenCorpID string `json:"open_corpid"`
	}
	if err := c.providerDoPost(ctx, "/cgi-bin/service/corpid_to_opencorpid", body, &resp); err != nil {
		return "", err
	}
	return resp.OpenCorpID, nil
}

// UserIDToOpenUserID 批量把 userid 转换为跨服务商匿名的 open_userid。
func (c *Client) UserIDToOpenUserID(ctx context.Context, corpID string, userIDs []string) (*UserIDConvertResp, error) {
	body := map[string]interface{}{
		"auth_corpid": corpID,
		"userid_list": userIDs,
	}
	var resp UserIDConvertResp
	if err := c.providerDoPost(ctx, "/cgi-bin/service/batch/userid_to_openuserid", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetProviderAccessToken 返回当前 provider_access_token(lazy 获取 + 自动缓存)。
// 未在 Config 里配置 ProviderCorpID / ProviderSecret 时返回对应哨兵错误。
func (c *Client) GetProviderAccessToken(ctx context.Context) (string, error) {
	return c.getProviderAccessToken(ctx)
}
