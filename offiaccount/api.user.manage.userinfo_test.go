package offiaccount

import (
	"context"
	"errors"
	"testing"
)

// TestGetUserInfo_PropagatesErrcode locks the fix for the silent-error-swallow
// bug (audit C1 / S-H3): UserInfo now embeds Resp, so checkEmbeddedResp catches
// a WeChat business error (e.g. 40003 invalid openid) instead of returning
// (&UserInfo{全零}, nil) and leaking an empty openid to the caller's auth logic.
//
// The sibling TestGetUserInfo_Success covers the happy path (and, post-fix, that
// embedding Resp did not break normal decoding); this test covers the error
// path that previously went unchecked.
func TestGetUserInfo_PropagatesErrcode(t *testing.T) {
	srv := userOkServer(t, `{"errcode":40003,"errmsg":"invalid openid"}`)
	defer srv.Close()

	c := newUserTestClient(t, srv)
	info, err := c.GetUserInfo(context.Background(), "bad-openid", "")
	if err == nil {
		t.Fatalf("GetUserInfo must return an error when WeChat returns errcode; got info=%+v, err=nil", info)
	}
	var werr *WeixinError
	if !errors.As(err, &werr) {
		t.Fatalf("want *WeixinError, got %T: %v", err, err)
	}
	if werr.ErrCode != 40003 {
		t.Errorf("want errcode 40003, got %d", werr.ErrCode)
	}
}
