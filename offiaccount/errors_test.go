package offiaccount_test

import (
	"errors"
	"testing"

	"github.com/godrealms/go-wechat-sdk/offiaccount"
	"github.com/godrealms/go-wechat-sdk/utils"
)

// Compile-time assertion: *offiaccount.WeixinError satisfies utils.WechatAPIError.
var _ utils.WechatAPIError = (*offiaccount.WeixinError)(nil)

func TestWeixinError_Code_Message(t *testing.T) {
	tests := []struct {
		code int
		msg  string
	}{
		{40001, "invalid credential"},
		{42001, "access_token expired"},
		{48001, "api unauthorized"},
	}
	for _, tt := range tests {
		e := &offiaccount.WeixinError{ErrCode: tt.code, ErrMsg: tt.msg}
		if e.Code() != tt.code {
			t.Errorf("Code() = %d, want %d", e.Code(), tt.code)
		}
		if e.Message() != tt.msg {
			t.Errorf("Message() = %q, want %q", e.Message(), tt.msg)
		}
	}
}

func TestCheckResp_ZeroErrcode(t *testing.T) {
	if err := offiaccount.CheckResp(&offiaccount.Resp{ErrCode: 0}); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestCheckResp_NonZeroErrcode(t *testing.T) {
	err := offiaccount.CheckResp(&offiaccount.Resp{ErrCode: 40001, ErrMsg: "invalid credential"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var werr *offiaccount.WeixinError
	if !errors.As(err, &werr) {
		t.Fatalf("expected *WeixinError, got %T", err)
	}
	if werr.Code() != 40001 {
		t.Errorf("Code() = %d, want 40001", werr.Code())
	}
	if werr.Message() != "invalid credential" {
		t.Errorf("Message() = %q, want \"invalid credential\"", werr.Message())
	}
	var apiErr utils.WechatAPIError
	if !errors.As(err, &apiErr) {
		t.Fatal("expected utils.WechatAPIError")
	}
}
