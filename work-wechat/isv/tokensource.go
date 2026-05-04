package isv

import "github.com/godrealms/go-wechat-sdk/utils"

// TokenSource 是下游"代企业调用"子项目的注入点。
// CorpClient 会实现此接口。
//
// Aliased to utils.TokenSource so a single implementation works across
// every WeChat-product Client.
type TokenSource = utils.TokenSource
