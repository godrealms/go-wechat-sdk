package mini_program

import "context"

// TokenSource is an injectable source of access tokens.
// When a Client is configured with a TokenSource, AccessToken() delegates to it
// instead of calling /cgi-bin/token directly. The typical use case is delegated
// calls via the Open Platform (oplatform.AuthorizerClient implements this interface).
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}
