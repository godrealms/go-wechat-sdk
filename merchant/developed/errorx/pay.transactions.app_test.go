package errorx

import (
	"errors"
	"strings"
	"testing"
)

func TestError_ImplementsErrorInterface(t *testing.T) {
	var _ error = (*Error)(nil)
}

func TestError_Error(t *testing.T) {
	e := &Error{Code: 400, Message: "param error", Solution: "fix it"}
	if e.Error() != "param error" {
		t.Errorf("Error() = %q, want %q", e.Error(), "param error")
	}
}

func TestNewError(t *testing.T) {
	e := NewError(429, "rate limited", "back off")
	if e == nil {
		t.Fatal("NewError returned nil")
	}
	if e.Code != 429 || e.Message != "rate limited" || e.Solution != "back off" {
		t.Errorf("unexpected fields: %+v", e)
	}
}

func TestLookupTransactionsApp_Known(t *testing.T) {
	tests := []struct {
		code        string
		expectMatch string
	}{
		{"PARAM_ERROR", "参数"},
		{"SIGN_ERROR", "签名"},
		{"OUT_TRADE_NO_USED", "订单号"},
		{"FREQUENCY_LIMITED", "频率"},
		{"MCH_NOT_EXISTS", "商户号"},
		{"NO_AUTH", "权限"},
		{"INVALID_REQUEST", "请参阅"},
		{"SYSTEM_ERROR", "重试"},
		{"APPID_MCHID_NOT_MATCH", "AppID"},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			sol, ok := LookupTransactionsApp(tt.code)
			if !ok {
				t.Fatalf("expected %q to be in lookup", tt.code)
			}
			if !strings.Contains(sol, tt.expectMatch) {
				t.Errorf("solution for %q = %q, expected contain %q", tt.code, sol, tt.expectMatch)
			}
		})
	}
}

func TestLookupTransactionsApp_Unknown(t *testing.T) {
	if _, ok := LookupTransactionsApp("DOES_NOT_EXIST"); ok {
		t.Error("expected ok=false for unknown code")
	}
}

func TestLookupTransactionsApp_EmptyCode(t *testing.T) {
	if _, ok := LookupTransactionsApp(""); ok {
		t.Error("expected ok=false for empty string")
	}
}

func TestErrors_AllEntriesNonNilAndPopulated(t *testing.T) {
	if len(Errors) == 0 {
		t.Fatal("Errors map is empty — package is broken")
	}
	for code, e := range Errors {
		if e == nil {
			t.Errorf("nil entry for %q", code)
			continue
		}
		if e.Message == "" {
			t.Errorf("empty Message for %q", code)
		}
		if e.Solution == "" {
			t.Errorf("empty Solution for %q", code)
		}
		if e.Code == 0 {
			t.Errorf("zero Code for %q", code)
		}
	}
}

// TestError_AsErrorChain ensures *Error works with errors.As, the standard
// way callers in the merchant package would inspect a returned error.
func TestError_AsErrorChain(t *testing.T) {
	original := NewError(400, "bad", "fix")
	wrapped := wrap(original)
	var got *Error
	if !errors.As(wrapped, &got) {
		t.Fatal("errors.As failed to find *Error in wrap chain")
	}
	if got.Code != 400 {
		t.Errorf("unwrapped Code = %d, want 400", got.Code)
	}
}

type wrappedErr struct{ inner error }

func (w *wrappedErr) Error() string { return "outer: " + w.inner.Error() }
func (w *wrappedErr) Unwrap() error { return w.inner }
func wrap(e error) error            { return &wrappedErr{inner: e} }
