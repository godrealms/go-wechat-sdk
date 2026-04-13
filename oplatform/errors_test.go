package oplatform_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/godrealms/go-wechat-sdk/oplatform"
	"github.com/godrealms/go-wechat-sdk/utils"
)

// Compile-time assertion: *oplatform.WeixinError satisfies utils.WechatAPIError.
var _ utils.WechatAPIError = (*oplatform.WeixinError)(nil)

func TestWeixinError_Implements_WechatAPIError(t *testing.T) {
	tests := []struct {
		name    string
		errcode int
		errmsg  string
	}{
		{"auth failure", 40001, "invalid credential"},
		{"access token expired", 42001, "access_token expired"},
		{"api unauthorized", 48001, "api unauthorized"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &oplatform.WeixinError{ErrCode: tt.errcode, ErrMsg: tt.errmsg}

			// Interface satisfaction via errors.As.
			var apiErr utils.WechatAPIError
			wrapped := fmt.Errorf("wrapped: %w", e)
			if !errors.As(wrapped, &apiErr) {
				t.Fatal("errors.As: expected WechatAPIError")
			}
			if apiErr.Code() != tt.errcode {
				t.Errorf("Code() = %d, want %d", apiErr.Code(), tt.errcode)
			}
			if apiErr.Message() != tt.errmsg {
				t.Errorf("Message() = %q, want %q", apiErr.Message(), tt.errmsg)
			}

			// Direct method calls.
			if e.Code() != tt.errcode {
				t.Errorf("e.Code() = %d, want %d", e.Code(), tt.errcode)
			}
			if e.Message() != tt.errmsg {
				t.Errorf("e.Message() = %q, want %q", e.Message(), tt.errmsg)
			}
		})
	}
}

func TestWeixinError_ErrorString(t *testing.T) {
	e := &oplatform.WeixinError{ErrCode: 40001, ErrMsg: "invalid credential"}
	got := e.Error()
	if got != "oplatform: errcode=40001 errmsg=invalid credential" {
		t.Errorf("Error() = %q, unexpected format", got)
	}
}
