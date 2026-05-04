package mini_program

import "github.com/godrealms/go-wechat-sdk/utils"

// TokenSource is an injectable source of access tokens.
// When a Client is configured with a TokenSource, AccessToken() delegates to it
// instead of calling /cgi-bin/token directly. The typical use case is delegated
// calls via the Open Platform (oplatform.AuthorizerClient implements this interface).
//
// Aliased to utils.TokenSource so a single implementation works across every
// WeChat-product Client.
type TokenSource = utils.TokenSource
