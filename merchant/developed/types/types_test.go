package types_test

import (
	"testing"

	"github.com/godrealms/go-wechat-sdk/merchant/developed/types"
)

// --- TradeBillQuest.ToUrlValues ---

func TestTradeBillQuest_ToUrlValues_AllFields(t *testing.T) {
	q := &types.TradeBillQuest{
		BillDate: "2024-01-15",
		BillType: "SUCCESS",
		TarType:  "GZIP",
	}
	vals := q.ToUrlValues()
	if vals.Get("bill_date") != "2024-01-15" {
		t.Errorf("expected bill_date=2024-01-15, got %q", vals.Get("bill_date"))
	}
	if vals.Get("bill_type") != "SUCCESS" {
		t.Errorf("expected bill_type=SUCCESS, got %q", vals.Get("bill_type"))
	}
	if vals.Get("tar_type") != "GZIP" {
		t.Errorf("expected tar_type=GZIP, got %q", vals.Get("tar_type"))
	}
}

func TestTradeBillQuest_ToUrlValues_OptionalFieldsOmitted(t *testing.T) {
	q := &types.TradeBillQuest{BillDate: "2024-01-15"}
	vals := q.ToUrlValues()
	if vals.Get("bill_date") != "2024-01-15" {
		t.Errorf("expected bill_date=2024-01-15, got %q", vals.Get("bill_date"))
	}
	if vals.Get("bill_type") != "" {
		t.Errorf("expected empty bill_type, got %q", vals.Get("bill_type"))
	}
	if vals.Get("tar_type") != "" {
		t.Errorf("expected empty tar_type, got %q", vals.Get("tar_type"))
	}
}

func TestTradeBillQuest_ToUrlValues_EmptyBillDateUsesToday(t *testing.T) {
	q := &types.TradeBillQuest{}
	vals := q.ToUrlValues()
	if vals.Get("bill_date") == "" {
		t.Error("expected bill_date to default to today when empty")
	}
}

// --- FundsBillQuest.ToUrlValues ---

func TestFundsBillQuest_ToUrlValues_AllFields(t *testing.T) {
	q := &types.FundsBillQuest{
		BillDate:    "2024-01-15",
		AccountType: "BASIC",
		TarType:     "GZIP",
	}
	vals := q.ToUrlValues()
	if vals.Get("bill_date") != "2024-01-15" {
		t.Errorf("expected bill_date=2024-01-15, got %q", vals.Get("bill_date"))
	}
	if vals.Get("account_type") != "BASIC" {
		t.Errorf("expected account_type=BASIC, got %q", vals.Get("account_type"))
	}
	if vals.Get("tar_type") != "GZIP" {
		t.Errorf("expected tar_type=GZIP, got %q", vals.Get("tar_type"))
	}
}

func TestFundsBillQuest_ToUrlValues_OptionalFieldsOmitted(t *testing.T) {
	q := &types.FundsBillQuest{BillDate: "2024-01-15"}
	vals := q.ToUrlValues()
	if vals.Get("account_type") != "" {
		t.Errorf("expected empty account_type, got %q", vals.Get("account_type"))
	}
	if vals.Get("tar_type") != "" {
		t.Errorf("expected empty tar_type, got %q", vals.Get("tar_type"))
	}
}

func TestFundsBillQuest_ToUrlValues_EmptyDateUsesToday(t *testing.T) {
	q := &types.FundsBillQuest{}
	vals := q.ToUrlValues()
	if vals.Get("bill_date") == "" {
		t.Error("expected bill_date to default to today when empty")
	}
}

// --- Notify.IsPaymentSuccess ---

func TestNotify_IsPaymentSuccess_True(t *testing.T) {
	n := &types.Notify{EventType: "TRANSACTION.SUCCESS"}
	if !n.IsPaymentSuccess() {
		t.Error("expected IsPaymentSuccess to return true for TRANSACTION.SUCCESS")
	}
}

func TestNotify_IsPaymentSuccess_False(t *testing.T) {
	cases := []string{"", "TRANSACTION.REFUND", "REFUND.SUCCESS"}
	for _, et := range cases {
		n := &types.Notify{EventType: et}
		if n.IsPaymentSuccess() {
			t.Errorf("expected IsPaymentSuccess=false for EventType=%q", et)
		}
	}
}
