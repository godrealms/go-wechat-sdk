package offiaccount

import "context"

// TokenSource is an injectable access_token provider.
// When set on a Client via WithTokenSource, AccessTokenE delegates to it instead of calling
// /cgi-bin/token directly. Implement this interface to support open-platform
// component-on-behalf-of flows.
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}

// Invalidator is an optional capability for TokenSource implementations that
// support explicit cache eviction. When a TokenSource implements this
// interface, Client.Invalidate() will delegate to it; otherwise the Client
// can only clear its own internal cache.
//
// utils.TokenCache implements this interface, so any TokenSource backed by
// utils.TokenCache (oplatform AuthorizerClient, mini-program, channels, etc.)
// supports 40001 self-heal out of the box.
type Invalidator interface {
	Invalidate()
}
