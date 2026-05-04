package offiaccount

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestDoGet_PropagatesNon40001Errcode verifies that any non-token-expired
// errcode flows through doGet → checkEmbeddedResp → *WeixinError without
// being silently dropped. This guards the 'audit 2026-04-14 fix' note in
// check.go that was prompted by ~85% of methods historically dropping
// business errors.
func TestDoGet_PropagatesNon40001Errcode(t *testing.T) {
	tests := []struct {
		name    string
		errcode int
		errmsg  string
	}{
		{"40013 invalid appid", 40013, "invalid appid"},
		{"45009 reach api limit", 45009, "reach max api daily quota limit"},
		{"40029 invalid code", 40029, "invalid code"},
		{"48001 api unauthorized", 48001, "api unauthorized"},
		{"45011 api freq out of limit", 45011, "api freq out of limit"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/cgi-bin/token") {
					_, _ = w.Write([]byte(`{"access_token":"T","expires_in":7200}`))
					return
				}
				w.Write([]byte(`{"errcode":` + itoa(tt.errcode) + `,"errmsg":"` + tt.errmsg + `"}`))
			}))
			defer srv.Close()

			c := newClientWithBaseURL(srv.URL, &Config{AppId: "a", AppSecret: "b"})
			result := &Resp{}
			err := c.doGet(context.Background(), "/cgi-bin/test?access_token=T", nil, result)
			if err == nil {
				t.Fatal("expected error")
			}
			var werr *WeixinError
			if !errors.As(err, &werr) {
				t.Fatalf("expected *WeixinError, got %T: %v", err, err)
			}
			if werr.ErrCode != tt.errcode {
				t.Errorf("ErrCode = %d, want %d", werr.ErrCode, tt.errcode)
			}
			if !strings.Contains(werr.ErrMsg, tt.errmsg) {
				t.Errorf("ErrMsg = %q, want contain %q", werr.ErrMsg, tt.errmsg)
			}
		})
	}
}

// TestDoPost_PropagatesNon40001Errcode mirrors the GET test for POST.
func TestDoPost_PropagatesNon40001Errcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cgi-bin/token") {
			_, _ = w.Write([]byte(`{"access_token":"T","expires_in":7200}`))
			return
		}
		_, _ = w.Write([]byte(`{"errcode":45011,"errmsg":"api freq out of limit"}`))
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "a", AppSecret: "b"})
	result := &Resp{}
	err := c.doPost(context.Background(), "/cgi-bin/test?access_token=T", map[string]any{"x": 1}, result)
	if err == nil {
		t.Fatal("expected error")
	}
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 45011 {
		t.Errorf("expected WeixinError 45011, got %v", err)
	}
}

// TestCheckEmbeddedResp_HandlesEmbeddedRespOnStruct verifies the reflection
// path for result types that embed Resp (the dominant pattern for response
// structs in this package).
func TestCheckEmbeddedResp_HandlesEmbeddedRespOnStruct(t *testing.T) {
	type WrapperWithEmbeddedResp struct {
		Resp
		Data string `json:"data"`
	}
	r := &WrapperWithEmbeddedResp{Resp: Resp{ErrCode: 40029, ErrMsg: "invalid code"}}
	err := checkEmbeddedResp(r)
	if err == nil {
		t.Fatal("expected error from embedded non-zero Resp")
	}
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 40029 {
		t.Errorf("expected WeixinError 40029, got %v", err)
	}
}

func TestCheckEmbeddedResp_NilInput(t *testing.T) {
	if err := checkEmbeddedResp(nil); err != nil {
		t.Errorf("nil input must produce nil error, got %v", err)
	}
}

func TestCheckEmbeddedResp_NoEmbeddedRespIsNoOp(t *testing.T) {
	type StructWithoutResp struct {
		Foo string
	}
	if err := checkEmbeddedResp(&StructWithoutResp{Foo: "bar"}); err != nil {
		t.Errorf("struct without embedded Resp must be tolerated, got %v", err)
	}
}

func TestCheckEmbeddedResp_NonStructIsNoOp(t *testing.T) {
	s := "just a string"
	if err := checkEmbeddedResp(&s); err != nil {
		t.Errorf("non-struct must be tolerated, got %v", err)
	}
}

// itoa is a tiny local helper to avoid pulling strconv just for one int→string.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
