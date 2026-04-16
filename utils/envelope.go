package utils

import (
	"encoding/json"
	"fmt"
)

// BaseResp holds the common WeChat error envelope fields (errcode + errmsg)
// present in virtually every WeChat API JSON response. It is used as the
// first-pass probe in DecodeEnvelope to detect API-level errors before
// unmarshalling the full response body.
type BaseResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// ErrFactory builds a package-specific error from errcode, errmsg, and the
// API path. Each package passes its own factory to DecodeEnvelope so the
// returned error has the correct concrete type (e.g. *channels.APIError)
// for callers using errors.As.
type ErrFactory func(code int, msg, path string) error

// DecodeEnvelope performs the two-stage JSON decode that every WeChat legacy
// API response requires:
//
//  1. Unmarshal into BaseResp to check errcode.
//  2. If errcode != 0, return the error produced by newErr.
//  3. Otherwise, unmarshal the full body into out (if non-nil).
//
// pkg is the human-readable package prefix used in wrap messages
// (e.g. "channels", "mini-program"). path is the API endpoint, used both
// in error messages and passed through to newErr.
//
// This consolidates the identical decodeEnvelope logic previously duplicated
// across six packages.
func DecodeEnvelope(pkg, path string, body []byte, out any, newErr ErrFactory) error {
	var base BaseResp
	if err := json.Unmarshal(body, &base); err != nil {
		return fmt.Errorf("%s: %s: decode envelope: %w (body snippet: %s)",
			pkg, path, err, Snippet(body))
	}
	if base.ErrCode != 0 {
		return newErr(base.ErrCode, base.ErrMsg, path)
	}
	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("%s: %s: decode result: %w", pkg, path, err)
		}
	}
	return nil
}

// Snippet returns at most the first 200 bytes of b as a string, for use in
// error messages. This avoids dumping multi-KB error pages into log lines.
func Snippet(b []byte) string {
	const max = 200
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "...(truncated)"
}
