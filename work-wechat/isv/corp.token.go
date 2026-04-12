package isv

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// GetCorpToken 调用 service/get_corp_token 换取企业 corp_access_token(底层方法)。
// 不写 Store —— 调用方通常应该用 CorpClient.AccessToken。
func (c *Client) GetCorpToken(ctx context.Context, corpID, permanentCode string) (*CorpTokenResp, error) {
	body := map[string]string{
		"auth_corpid":    corpID,
		"permanent_code": permanentCode,
	}
	var resp CorpTokenResp
	if err := c.doPost(ctx, "/cgi-bin/service/get_corp_token", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CorpClient 代表"代某企业调用"的会话句柄,实现 TokenSource 接口。
type CorpClient struct {
	parent *Client
	corpID string
}

// CorpClient 工厂。
func (c *Client) CorpClient(corpID string) *CorpClient {
	return &CorpClient{parent: c, corpID: corpID}
}

// CorpID 返回当前 CorpClient 代理的企业 corpid。
func (cc *CorpClient) CorpID() string { return cc.corpID }

// AccessToken 返回企业 corp_access_token(lazy + 双检锁 + 单飞)。
func (cc *CorpClient) AccessToken(ctx context.Context) (string, error) {
	if tok, ok, err := cc.readValidCorpToken(ctx); err != nil {
		return "", err
	} else if ok {
		return tok, nil
	}

	lock := cc.parent.lockFor(cc.corpID)
	lock.Lock()
	defer lock.Unlock()

	if tok, ok, err := cc.readValidCorpToken(ctx); err != nil {
		return "", err
	} else if ok {
		return tok, nil
	}

	return cc.refreshLocked(ctx)
}

// Refresh 强制刷新单个企业的 corp_token(忽略缓存)。
func (cc *CorpClient) Refresh(ctx context.Context) error {
	lock := cc.parent.lockFor(cc.corpID)
	lock.Lock()
	defer lock.Unlock()
	_, err := cc.refreshLocked(ctx)
	return err
}

func (cc *CorpClient) readValidCorpToken(ctx context.Context) (string, bool, error) {
	tokens, err := cc.parent.store.GetAuthorizer(ctx, cc.parent.cfg.SuiteID, cc.corpID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", false, ErrAuthorizerRevoked
		}
		return "", false, err
	}
	if tokens.CorpAccessToken == "" || time.Until(tokens.CorpTokenExpireAt) <= safetyMargin {
		return "", false, nil
	}
	return tokens.CorpAccessToken, true, nil
}

// refreshLocked 在持有 lockFor(corpID) 的前提下,通过 permanent_code 换取新 corp_token 并写回 Store。
func (cc *CorpClient) refreshLocked(ctx context.Context) (string, error) {
	tokens, err := cc.parent.store.GetAuthorizer(ctx, cc.parent.cfg.SuiteID, cc.corpID)
	if err != nil {
		return "", err
	}
	resp, err := cc.parent.GetCorpToken(ctx, cc.corpID, tokens.PermanentCode)
	if err != nil {
		return "", err
	}
	tokens.CorpAccessToken = resp.AccessToken
	tokens.CorpTokenExpireAt = time.Now().Add(time.Duration(resp.ExpiresIn)*time.Second - safetyMargin)
	if err := cc.parent.store.PutAuthorizer(ctx, cc.parent.cfg.SuiteID, cc.corpID, tokens); err != nil {
		return "", fmt.Errorf("isv: persist corp token: %w", err)
	}
	return tokens.CorpAccessToken, nil
}

// lockFor 返回 corpid 专属的 mutex(从 sync.Map 取,首次创建)。
func (c *Client) lockFor(corpID string) *sync.Mutex {
	if v, ok := c.corpMu.Load(corpID); ok {
		return v.(*sync.Mutex)
	}
	v, _ := c.corpMu.LoadOrStore(corpID, &sync.Mutex{})
	return v.(*sync.Mutex)
}

// RefreshAll 遍历 Store 中所有已授权企业,刷新它们的 corp_token。
// 任一失败继续下一个,最后聚合错误用 errors.Join。
func (c *Client) RefreshAll(ctx context.Context) error {
	list, err := c.store.ListAuthorizers(ctx, c.cfg.SuiteID)
	if err != nil {
		return err
	}
	var errs []error
	for _, corpID := range list {
		if err := c.CorpClient(corpID).Refresh(ctx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", corpID, err))
		}
	}
	return errors.Join(errs...)
}
