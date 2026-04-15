package isv

import (
	"context"
	"errors"
	"time"
)

// GetPermanentCode 用 auth_code 换取企业永久授权码并自动持久化 AuthorizerTokens。
//
// Concurrency note (audit Batch 3): the PutAuthorizer write is serialized
// through the per-corpID mutex that CorpClient.refreshLocked also holds, so
// GetPermanentCode landing concurrently with an in-flight corp_token refresh
// for the same corpID cannot tear the stored record. The HTTP call itself
// is intentionally kept outside the lock — callers waiting on the mutex
// shouldn't be blocked by network round-trips from an unrelated operation.
func (c *Client) GetPermanentCode(ctx context.Context, authCode string) (*PermanentCodeResp, error) {
	body := map[string]string{"auth_code": authCode}
	var resp PermanentCodeResp
	if err := c.doPost(ctx, "/cgi-bin/service/get_permanent_code", body, &resp); err != nil {
		return nil, err
	}

	// 首次同时拿到 corp_token,写入 Store。
	// 缺字段视为 API 返回不完整,直接报错,避免后续调用拿到 ErrAuthorizerRevoked。
	if resp.AuthCorpInfo.CorpID == "" || resp.PermanentCode == "" || resp.AccessToken == "" {
		return nil, errors.New("isv: get_permanent_code response missing required fields (auth_corp_info.corpid / permanent_code / access_token)")
	}
	expiresAt := time.Now().Add(clampTokenTTL(resp.ExpiresIn))
	tokens := &AuthorizerTokens{
		CorpID:            resp.AuthCorpInfo.CorpID,
		PermanentCode:     resp.PermanentCode,
		CorpAccessToken:   resp.AccessToken,
		CorpTokenExpireAt: expiresAt,
	}
	// Serialize the store write with CorpClient.refreshLocked for the same
	// corpID so that an in-flight refresh cannot clobber our fresh tokens
	// (or vice versa).
	lock := c.lockFor(resp.AuthCorpInfo.CorpID)
	lock.Lock()
	defer lock.Unlock()
	if err := c.store.PutAuthorizer(ctx, c.cfg.SuiteID, resp.AuthCorpInfo.CorpID, tokens); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAuthInfo 查询企业授权信息(不缓存)。
func (c *Client) GetAuthInfo(ctx context.Context, corpID, permanentCode string) (*AuthInfoResp, error) {
	body := map[string]string{
		"auth_corpid":    corpID,
		"permanent_code": permanentCode,
	}
	var resp AuthInfoResp
	if err := c.doPost(ctx, "/cgi-bin/service/get_auth_info", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAdminList 获取授权应用的管理员列表。
func (c *Client) GetAdminList(ctx context.Context, corpID, agentID string) (*AdminListResp, error) {
	body := map[string]string{
		"auth_corpid": corpID,
		"agentid":     agentID,
	}
	var resp AdminListResp
	if err := c.doPost(ctx, "/cgi-bin/service/get_admin_list", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
