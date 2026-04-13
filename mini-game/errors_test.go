package mini_game_test

import (
	"errors"
	"testing"

	mini_game "github.com/godrealms/go-wechat-sdk/mini-game"
	"github.com/godrealms/go-wechat-sdk/utils"
)

var _ utils.WechatAPIError = (*mini_game.APIError)(nil)

func TestAPIError_Code_Message(t *testing.T) {
	tests := []struct {
		code int
		msg  string
		path string
	}{
		{40001, "invalid credential", "/wxa/game/getaccessinfo"},
		{45009, "api freq out of limit", "/wxa/game/getframesync"},
	}
	for _, tt := range tests {
		e := &mini_game.APIError{ErrCode: tt.code, ErrMsg: tt.msg, Path: tt.path}
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
	e := &mini_game.APIError{ErrCode: 40001, ErrMsg: "invalid credential", Path: "/wxa/game/getaccessinfo"}
	got := e.Error()
	want := "mini_game: /wxa/game/getaccessinfo errcode=40001 errmsg=invalid credential"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
