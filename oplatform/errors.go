package oplatform

import (
	"errors"
	"fmt"
)

// WeixinError wraps a non-zero WeChat errcode from the Open Platform API.
type WeixinError struct {
	ErrCode int
	ErrMsg  string
}

func (e *WeixinError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("oplatform: errcode=%d errmsg=%s", e.ErrCode, e.ErrMsg)
}

// Code returns the numeric errcode. Implements utils.WechatAPIError.
func (e *WeixinError) Code() int { return e.ErrCode }

// Message returns the human-readable errmsg. Implements utils.WechatAPIError.
func (e *WeixinError) Message() string { return e.ErrMsg }

// Sentinel errors returned by the oplatform package.
var (
	// ErrNotFound is returned by Store implementations when a key does not exist (not an I/O error).
	ErrNotFound = errors.New("oplatform: not found")

	// ErrAuthorizerRevoked is returned when an authorizer's refresh_token has been revoked (errcode=61023).
	// Callers should remove the authorizer record from the Store and re-initiate the authorization flow.
	ErrAuthorizerRevoked = errors.New("oplatform: authorizer refresh_token revoked")

	// ErrVerifyTicketMissing is returned when the component_access_token needs refreshing
	// but the Store has not yet received the component_verify_ticket pushed by WeChat.
	ErrVerifyTicketMissing = errors.New("oplatform: component_verify_ticket not yet received")
)
