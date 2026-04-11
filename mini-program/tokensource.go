package mini_program

import "context"

// TokenSource 是 access_token 的可注入来源。
// 当 Client 配置了 TokenSource 时，AccessToken() 会直接委托给它，
// 不再调用 /cgi-bin/token。典型场景：开放平台代调用
// (oplatform.AuthorizerClient 实现本接口)。
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}
