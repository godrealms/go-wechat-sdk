package offiaccount

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------- Invoice auth ----------

func TestSetInvoiceBizAttr_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if r.URL.Query().Get("action") != "set_contact" {
			t.Errorf("expected action=set_contact, got %q", r.URL.Query().Get("action"))
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.SetInvoiceBizAttr(context.Background(), "set_contact", &SetBizAttrRequest{
		Contact: &Contact{Phone: "13800138000", Timeout: 300},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestGetAuthData_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/getauthdata") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","invoice_status":"auth success","auth_time":1609459200}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetAuthData(context.Background(), &GetAuthDataRequest{
		OrderID: "order1", SPAppID: "sp1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.InvoiceStatus != "auth success" {
		t.Errorf("unexpected invoice_status: %q", resp.InvoiceStatus)
	}
}

func TestGetInvoiceTicket_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","ticket":"TICKET123","expires_in":7200}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetInvoiceTicket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Ticket != "TICKET123" {
		t.Errorf("expected TICKET123, got %q", resp.Ticket)
	}
}

// ---------- Invoice reimburser ----------

func TestGetInvoiceInfo_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/reimburse/getinvoiceinfo") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["card_id"] != "card1" {
			t.Errorf("unexpected card_id: %v", body["card_id"])
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","card_id":"card1","begin_time":1609459200,"end_time":1640995200}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetInvoiceInfo(context.Background(), &GetInvoiceInfoRequest{
		CardID: "card1", EncryptCode: "enc1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.CardID != "card1" {
		t.Errorf("expected card1, got %q", resp.CardID)
	}
}

func TestUpdateInvoiceReimburseStatus_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.UpdateInvoiceReimburseStatus(context.Background(), &UpdateInvoiceReimburseStatusRequest{
		CardID: "card1", EncryptCode: "enc1", ReimburseStatus: "INVOICE_REIMBURSE_INIT",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestGetInvoiceBatch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","item_list":[{"card_id":"card1","encrypt_code":"enc1"}]}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetInvoiceBatch(context.Background(), &GetInvoiceBatchRequest{
		ItemList: []*InvoiceListItem{{CardID: "card1", EncryptCode: "enc1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ItemList) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.ItemList))
	}
}

func TestUpdateInvoiceReimburseStatusBatch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.UpdateInvoiceReimburseStatusBatch(context.Background(), &UpdateInvoiceReimburseStatusBatchRequest{
		OpenID:          "o1",
		ReimburseStatus: "INVOICE_REIMBURSE_INIT",
		InvoiceList:     []*InvoiceListItem{{CardID: "card1", EncryptCode: "enc1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}
