package mini_store

import (
	"testing"

	"github.com/godrealms/go-wechat-sdk/utils"
)

// Compile-time assertion: *APIError satisfies utils.WechatAPIError. This keeps
// mini-store in lockstep with every other package-specific WeChat error type.
var _ utils.WechatAPIError = (*APIError)(nil)

func TestAPIError_Code_Message(t *testing.T) {
	e := &APIError{ErrCode: 9101000, ErrMsg: "no permission", Path: "/some/path"}
	if e.Code() != 9101000 {
		t.Errorf("Code() = %d, want 9101000", e.Code())
	}
	if e.Message() != "no permission" {
		t.Errorf("Message() = %q, want %q", e.Message(), "no permission")
	}
}
