package utils

import "errors"

// IsTokenExpired reports whether err carries a WeChat legacy-API errcode that
// means the access_token used in the request is no longer usable and the caller
// should fetch a fresh one and retry:
//
//   - 40001: access_token invalid (most common — admin reset, or another
//     instance refreshed the token out from under this one)
//   - 40014: access_token malformed (corrupted / wrong token)
//   - 42001: access_token expired (early-expiry edge case)
//   - 42007: ticket / token expired
//
// It matches against the WechatAPIError interface via errors.As, so it works
// for every package-specific error type in this SDK (channels.APIError,
// mini_program.APIError, isv.WeixinError, …) without importing the concrete
// package. WeChat Pay v3 errors use string codes (WechatAPIV3Error) and are
// intentionally not matched.
//
// When this returns true the caller should evict the cached token
// (Invalidator.Invalidate) and retry once; DoWithTokenRetry does exactly that.
func IsTokenExpired(err error) bool {
	var apiErr WechatAPIError
	if !errors.As(err, &apiErr) {
		return false
	}
	switch apiErr.Code() {
	case 40001, 40014, 42001, 42007:
		return true
	default:
		return false
	}
}

// DoWithTokenRetry runs fn, and if fn fails with a token-expired error (see
// IsTokenExpired) it evicts the cached token via inv.Invalidate() and runs fn
// exactly once more. Any other error — or a second token-expired error — is
// returned to the caller unchanged.
//
// fn MUST re-read the access_token from its (now-invalidated) source on every
// invocation for the retry to have any effect. The standard package helpers
// call c.AccessToken(ctx) inside fn, which re-fetches after Invalidate, so a
// stale-token request self-heals transparently. A nil inv disables the retry:
// fn is run exactly once.
//
// This consolidates the 40001 self-heal that previously lived only in
// offiaccount, extending it to every product line. Without it, a token that
// expires early — the common case under multi-instance deployment, where one
// instance can refresh the shared token while another is mid-request — leaves
// the second instance failing every call until the token naturally expires
// (up to ~2 hours), wasting quota the whole time.
func DoWithTokenRetry(inv Invalidator, fn func() error) error {
	err := fn()
	if inv == nil || !IsTokenExpired(err) {
		return err
	}
	inv.Invalidate()
	return fn()
}
