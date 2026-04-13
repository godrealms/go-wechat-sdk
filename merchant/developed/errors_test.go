package pay_test

import (
	"errors"
	"testing"

	developed "github.com/godrealms/go-wechat-sdk/merchant/developed"
	"github.com/godrealms/go-wechat-sdk/utils"
)

var _ utils.WechatAPIError = (*developed.APIError)(nil)

func TestAPIError_Code_Message(t *testing.T) {
	tests := []struct {
		code int
		msg  string
		path string
	}{
		{400, "PARAM_ERROR", "/v3/pay/transactions/app"},
		{401, "SIGN_ERROR", "/v3/pay/transactions/jsapi"},
		{500, "SYSTEM_ERROR", "/v3/pay/transactions/native"},
	}
	for _, tt := range tests {
		e := &developed.APIError{ErrCode: tt.code, ErrMsg: tt.msg, Path: tt.path}
		if e.Code() != tt.code {
			t.Errorf("Code() = %d, want %d", e.Code(), tt.code)
		}
		if e.Message() != tt.msg {
			t.Errorf("Message() = %q, want %q", e.Message(), tt.msg)
		}
		var apiErr utils.WechatAPIError
		if !errors.As(e, &apiErr) {
			t.Fatalf("errors.As: expected WechatAPIError")
		}
		if apiErr.Code() != tt.code {
			t.Errorf("via interface: Code() = %d, want %d", apiErr.Code(), tt.code)
		}
	}
}

func TestAPIError_ErrorString(t *testing.T) {
	e := &developed.APIError{ErrCode: 400, ErrMsg: "PARAM_ERROR", Path: "/v3/pay/transactions/app"}
	got := e.Error()
	want := "merchant/developed: /v3/pay/transactions/app errcode=400 errmsg=PARAM_ERROR"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
