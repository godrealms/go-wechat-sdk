package oplatform

import (
	"context"
	"errors"
	"fmt"
	"strings"

	mini_program "github.com/godrealms/go-wechat-sdk/mini-program"
	"github.com/godrealms/go-wechat-sdk/offiaccount"
)

// Compile-time assertions that AuthorizerClient satisfies both TokenSource shapes.
var (
	_ offiaccount.TokenSource  = (*AuthorizerClient)(nil)
	_ mini_program.TokenSource = (*AuthorizerClient)(nil)
)

// OffiaccountClient 返回一个预先注入了 AuthorizerClient 作为 TokenSource
// 的 offiaccount.Client。之后调用 off.AccessTokenE / 菜单 / 模板消息 / 素材
// 等任意 offiaccount 方法时，底层 token 自动来自开放平台。
func (a *AuthorizerClient) OffiaccountClient(opts ...offiaccount.Option) *offiaccount.Client {
	allOpts := append([]offiaccount.Option{offiaccount.WithTokenSource(a)}, opts...)
	return offiaccount.NewClient(context.Background(), &offiaccount.Config{AppId: a.appID}, allOpts...)
}

// MiniProgramClient 返回一个预先注入了 AuthorizerClient 作为 TokenSource
// 的 mini_program.Client。
func (a *AuthorizerClient) MiniProgramClient(opts ...mini_program.Option) (*mini_program.Client, error) {
	allOpts := append([]mini_program.Option{mini_program.WithTokenSource(a)}, opts...)
	return mini_program.NewClient(mini_program.Config{AppId: a.appID}, allOpts...)
}

// RefreshAll 对 Store 中所有已登记的 authorizer 调用 Refresh。
// 用于启动预热或外部 cron 触发。单个 appid 失败不中断循环，
// 所有错误汇总后以多行错误字符串返回。
func (c *Client) RefreshAll(ctx context.Context) error {
	ctx = touchContext(ctx)
	ids, err := c.store.ListAuthorizerAppIDs(ctx)
	if err != nil {
		return fmt.Errorf("oplatform: list authorizers: %w", err)
	}
	var errs []string
	for _, id := range ids {
		auth := c.Authorizer(id)
		if err := auth.Refresh(ctx); err != nil {
			if errors.Is(err, ErrAuthorizerRevoked) {
				errs = append(errs, fmt.Sprintf("%s: revoked", id))
				continue
			}
			errs = append(errs, fmt.Sprintf("%s: %v", id, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("oplatform: RefreshAll had %d failures:\n%s",
			len(errs), strings.Join(errs, "\n"))
	}
	return nil
}
