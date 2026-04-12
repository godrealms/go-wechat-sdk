package isv

import "context"

// TokenSource 是下游"代企业调用"子项目的注入点。
// CorpClient 会实现此接口。
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}
