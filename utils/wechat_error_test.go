package utils_test

import (
	"testing"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// fakeAPIError is a local type used to verify the interface contract.
type fakeAPIError struct {
	code int
	msg  string
}

func (f *fakeAPIError) Error() string   { return f.msg }
func (f *fakeAPIError) Code() int       { return f.code }
func (f *fakeAPIError) Message() string { return f.msg }

// Compile-time assertion: fakeAPIError satisfies WechatAPIError.
var _ utils.WechatAPIError = (*fakeAPIError)(nil)

func TestWechatAPIError_Interface(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
	}{
		{"zero errcode", 0, "ok"},
		{"common auth error", 40001, "invalid credential"},
		{"rate limit", 45009, "api freq out of limit"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e utils.WechatAPIError = &fakeAPIError{code: tt.code, msg: tt.message}
			if got := e.Code(); got != tt.code {
				t.Errorf("Code() = %d, want %d", got, tt.code)
			}
			if got := e.Message(); got != tt.message {
				t.Errorf("Message() = %q, want %q", got, tt.message)
			}
			if got := e.Error(); got == "" && tt.message != "" {
				t.Errorf("Error() returned empty string, want %q", tt.message)
			}
		})
	}
}

func TestWechatAPIError_NilSafety(t *testing.T) {
	var f *fakeAPIError
	var e utils.WechatAPIError = f
	if e == nil {
		t.Error("interface value should not be nil when set to a typed nil")
	}
}
