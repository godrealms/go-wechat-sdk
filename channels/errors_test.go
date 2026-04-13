package channels_test

import (
	"errors"
	"testing"

	"github.com/godrealms/go-wechat-sdk/channels"
	"github.com/godrealms/go-wechat-sdk/utils"
)

var _ utils.WechatAPIError = (*channels.APIError)(nil)

func TestAPIError_Code_Message(t *testing.T) {
	tests := []struct {
		code int
		msg  string
		path string
	}{
		{40001, "invalid credential", "/channels/ec/order/list"},
		{45009, "api freq out of limit", "/channels/ec/product/list"},
	}
	for _, tt := range tests {
		e := &channels.APIError{ErrCode: tt.code, ErrMsg: tt.msg, Path: tt.path}
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
	e := &channels.APIError{ErrCode: 40001, ErrMsg: "invalid credential", Path: "/channels/ec/order/list"}
	got := e.Error()
	want := "channels: /channels/ec/order/list errcode=40001 errmsg=invalid credential"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
