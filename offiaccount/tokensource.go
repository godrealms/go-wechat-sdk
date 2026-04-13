package offiaccount

import "context"

// TokenSource is an injectable access_token provider.
// When set on a Client via WithTokenSource, AccessTokenE delegates to it instead of calling
// /cgi-bin/token directly. Implement this interface to support open-platform
// component-on-behalf-of flows.
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}
