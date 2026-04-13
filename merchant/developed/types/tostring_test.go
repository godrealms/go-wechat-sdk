package types_test

import (
	"strings"
	"testing"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
)

func TestMchID_ToString_ValidJSON(t *testing.T) {
	m := &types.MchID{Mchid: "1234567890"}
	got := m.ToString()
	if !strings.Contains(got, `"mchid"`) {
		t.Errorf("ToString() = %q, want JSON containing mchid key", got)
	}
	if !strings.Contains(got, "1234567890") {
		t.Errorf("ToString() = %q, want JSON containing 1234567890", got)
	}
	if strings.HasPrefix(got, "<marshal error") {
		t.Errorf("ToString() returned error string: %s", got)
	}
}

func TestTransactions_ToString_ValidJSON(t *testing.T) {
	tx := &types.Transactions{
		Appid:       "wx123",
		Mchid:       "1234567890",
		Description: "test",
		OutTradeNo:  "order-001",
		NotifyUrl:   "https://example.com/notify",
		Amount:      &types.Amount{Total: 100, Currency: "CNY"},
	}
	got := tx.ToString()
	if !strings.Contains(got, `"appid"`) {
		t.Errorf("ToString() = %q, want JSON with appid key", got)
	}
	if strings.HasPrefix(got, "<marshal error") {
		t.Errorf("ToString() returned error string: %s", got)
	}
}

func TestRefunds_ToString_ValidJSON(t *testing.T) {
	r := &types.Refunds{
		OutTradeNo:  "order-001",
		OutRefundNo: "refund-001",
		Reason:      "test refund",
		NotifyUrl:   "https://example.com/notify",
	}
	got := r.ToString()
	if !strings.Contains(got, `"out_trade_no"`) {
		t.Errorf("ToString() = %q, want JSON with out_trade_no key", got)
	}
	if strings.HasPrefix(got, "<marshal error") {
		t.Errorf("ToString() returned error string: %s", got)
	}
}

func TestAbnormalRefund_ToString_ValidJSON(t *testing.T) {
	r := &types.AbnormalRefund{
		OutRefundNo: "refund-001",
		Type:        "USER_BANK_CARD",
	}
	got := r.ToString()
	if !strings.Contains(got, `"out_refund_no"`) {
		t.Errorf("ToString() = %q, want JSON with out_refund_no key", got)
	}
	if strings.HasPrefix(got, "<marshal error") {
		t.Errorf("ToString() returned error string: %s", got)
	}
}

func TestToString_NeverEmpty(t *testing.T) {
	m := &types.MchID{Mchid: "test123"}
	got := m.ToString()
	if got == "" {
		t.Error("ToString() returned empty string; expected JSON or error marker")
	}
}
