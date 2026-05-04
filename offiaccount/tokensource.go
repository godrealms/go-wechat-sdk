package offiaccount

import "github.com/godrealms/go-wechat-sdk/utils"

// TokenSource is an injectable access_token provider.
// When set on a Client via WithTokenSource, AccessTokenE delegates to it instead of calling
// /cgi-bin/token directly. Implement this interface to support open-platform
// component-on-behalf-of flows.
//
// Aliased to utils.TokenSource so a single implementation satisfies every
// WeChat-product Client (mini-program, channels, mini-game, mini-store,
// xiaowei, aispeech, isv).
type TokenSource = utils.TokenSource

// Invalidator is an optional capability for TokenSource implementations that
// support explicit cache eviction. When a TokenSource implements this
// interface, Client.Invalidate() will delegate to it; otherwise the Client
// can only clear its own internal cache.
//
// utils.TokenCache implements this interface, so any TokenSource backed by
// utils.TokenCache (oplatform AuthorizerClient, mini-program, channels, etc.)
// supports 40001 self-heal out of the box.
type Invalidator = utils.Invalidator
