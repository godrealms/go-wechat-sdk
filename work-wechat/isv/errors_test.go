package isv_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/godrealms/go-wechat-sdk/utils"
	isv "github.com/godrealms/go-wechat-sdk/work-wechat/isv"
)

// Compile-time assertion: *isv.WeixinError satisfies utils.WechatAPIError.
var _ utils.WechatAPIError = (*isv.WeixinError)(nil)

func TestWeixinError_Implements_WechatAPIError(t *testing.T) {
	tests := []struct {
		name    string
		errcode int
		errmsg  string
	}{
		{"suite token expired", 42001, "suite_access_token expired"},
		{"no permission", 60011, "no privilege to access/modify contact"},
		{"user not exist", 60111, "userid not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &isv.WeixinError{ErrCode: tt.errcode, ErrMsg: tt.errmsg}

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
	e := &isv.WeixinError{ErrCode: 42001, ErrMsg: "suite_access_token expired"}
	got := e.Error()
	want := "isv: weixin error 42001: suite_access_token expired"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
