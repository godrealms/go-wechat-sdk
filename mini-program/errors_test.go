package mini_program_test

import (
	"errors"
	"testing"

	mini_program "github.com/godrealms/go-wechat-sdk/mini-program"
	"github.com/godrealms/go-wechat-sdk/utils"
)

var _ utils.WechatAPIError = (*mini_program.APIError)(nil)

func TestAPIError_Code_Message(t *testing.T) {
	tests := []struct {
		code int
		msg  string
		path string
	}{
		{40001, "invalid credential", "/wxa/get_wxacode"},
		{45009, "api freq out of limit", "/wxa/msg_sec_check"},
	}
	for _, tt := range tests {
		e := &mini_program.APIError{ErrCode: tt.code, ErrMsg: tt.msg, Path: tt.path}
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
	e := &mini_program.APIError{ErrCode: 40001, ErrMsg: "invalid credential", Path: "/wxa/get_wxacode"}
	got := e.Error()
	want := "mini_program: /wxa/get_wxacode errcode=40001 errmsg=invalid credential"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
