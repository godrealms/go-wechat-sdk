package core

import "testing"

func TestResp_GetError_ReturnsNilWhenErrCodeIsZero(t *testing.T) {
	r := &Resp{ErrCode: 0, ErrMsg: "ok"}
	if err := r.GetError(); err != nil {
		t.Errorf("expected nil error, got: %v", err)
	}
}

func TestResp_GetError_ReturnsErrorWhenErrCodeIsNonZero(t *testing.T) {
	r := &Resp{ErrCode: 40001, ErrMsg: "invalid credential"}
	err := r.GetError()
	if err == nil {
		t.Fatal("expected non-nil error, got nil")
	}
	expected := "wechat api error 40001: invalid credential"
	if err.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func TestResp_GetError_FormatsMessageCorrectly(t *testing.T) {
	r := &Resp{ErrCode: 40013, ErrMsg: "invalid appid"}
	err := r.GetError()
	if err == nil {
		t.Fatal("expected non-nil error, got nil")
	}
	expected := "wechat api error 40013: invalid appid"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestResp_GetError_HandlesNilReceiver(t *testing.T) {
	var r *Resp
	err := r.GetError()
	if err != nil {
		t.Errorf("expected nil for nil receiver, got %v", err)
	}
}
