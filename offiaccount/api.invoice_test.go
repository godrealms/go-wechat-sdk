package offiaccount

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------- Invoice auth ----------

func TestSetInvoiceBizAttr_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/setbizattr") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("action") != "set_contact" {
			t.Errorf("expected action=set_contact, got %q", r.URL.Query().Get("action"))
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		contact, ok := body["contact"].(map[string]any)
		if !ok || contact["phone"] != "13800138000" {
			t.Errorf("unexpected body: %+v", body)
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
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["order_id"] != "order1" || body["s_pappid"] != "sp1" {
			t.Errorf("unexpected body: %+v", body)
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
		if !strings.HasSuffix(r.URL.Path, "/cgi-bin/ticket/getticket") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "wx_card" {
			t.Errorf("expected type=wx_card, got %q", r.URL.Query().Get("type"))
		}
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

func TestGetAuthUrl_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/getauthurl") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["s_pappid"] != "sp1" || body["order_id"] != "order1" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","auth_url":"https://example.com/auth","appid":"wx123"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetAuthUrl(context.Background(), &GetAuthUrlRequest{
		SPAppID: "sp1", OrderID: "order1", Money: 100, Timestamp: 1609459200, Source: "web", Ticket: "tk1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.AuthUrl != "https://example.com/auth" {
		t.Errorf("unexpected auth_url: %q", resp.AuthUrl)
	}
}

func TestRejectInsert_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/rejectinsert") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["order_id"] != "order1" || body["reason"] != "duplicate" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.RejectInsert(context.Background(), &RejectInsertRequest{
		SPAppID: "sp1", OrderID: "order1", Reason: "duplicate",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

// ---------- Invoice reimburser ----------

func TestGetInvoiceInfo_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/reimburse/getinvoiceinfo") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
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
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/reimburse/updateinvoicestatus") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["card_id"] != "card1" || body["reimburse_status"] != "INVOICE_REIMBURSE_INIT" {
			t.Errorf("unexpected body: %+v", body)
		}
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
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/reimburse/getinvoicebatch") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		items, ok := body["item_list"].([]any)
		if !ok || len(items) != 1 {
			t.Errorf("unexpected body: %+v", body)
		}
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
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/reimburse/updatestatusbatch") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["openid"] != "o1" || body["reimburse_status"] != "INVOICE_REIMBURSE_INIT" {
			t.Errorf("unexpected body: %+v", body)
		}
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

// ---------- Invoice platform ----------

func TestSetInvoiceUrl_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/seturl") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","invoice_url":"https://example.com/invoice"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.SetInvoiceUrl(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.InvoiceUrl != "https://example.com/invoice" {
		t.Errorf("unexpected invoice_url: %q", resp.InvoiceUrl)
	}
}

func TestGetPdf_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/platform/getpdf") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["action"] != "get_url" || body["s_media_id"] != "media1" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","pdf_url":"https://example.com/pdf","pdf_url_expire_time":7200}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetPdf(context.Background(), &GetPdfRequest{
		Action: "get_url", SMediaID: "media1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.PdfUrl != "https://example.com/pdf" {
		t.Errorf("unexpected pdf_url: %q", resp.PdfUrl)
	}
}

func TestUpdateInvoiceStatus_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/platform/updatestatus") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["card_id"] != "card1" || body["reimburse_status"] != "INVOICE_REIMBURSE_INIT" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.UpdateInvoiceStatus(context.Background(), &UpdateInvoiceStatusRequest{
		CardID: "card1", Code: "code1", ReimburseStatus: "INVOICE_REIMBURSE_INIT",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestSetPdf_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/platform/setpdf") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "multipart/form-data") {
			t.Errorf("expected multipart, got %q", ct)
		}
		_ = r.ParseMultipartForm(1 << 20)
		f, fh, err := r.FormFile("pdf")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if fh.Filename != "invoice.pdf" {
			t.Errorf("expected filename invoice.pdf, got %q", fh.Filename)
		}
		data, _ := io.ReadAll(f)
		if string(data) != "fake pdf" {
			t.Errorf("unexpected data: %q", data)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","s_media_id":"smedia123"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.SetPdf(context.Background(), "invoice.pdf", strings.NewReader("fake pdf"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.SMediaID != "smedia123" {
		t.Errorf("expected smedia123, got %q", resp.SMediaID)
	}
}

func TestCreateCard_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/platform/createcard") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		info, ok := body["invoice_info"].(map[string]any)
		if !ok || info["payee"] != "Test Corp" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","card_id":"card_new"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.CreateCard(context.Background(), &CreateCardRequest{
		InvoiceInfo: &CreateCardInvoiceInfo{
			Payee: "Test Corp",
			Type:  "增值税电子普通发票",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.CardID != "card_new" {
		t.Errorf("expected card_new, got %q", resp.CardID)
	}
}

func TestInsertInvoice_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/insert") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["order_id"] != "order1" || body["card_id"] != "card1" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","code":"inv_code","openid":"o1"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.InsertInvoice(context.Background(), &InsertInvoiceRequest{
		OrderID: "order1", CardID: "card1", AppID: "wx123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Code != "inv_code" {
		t.Errorf("expected inv_code, got %q", resp.Code)
	}
}

// ---------- Invoice fiscal receipt ----------

func TestGetFiscalAuthData_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/getauthdata") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["order_id"] != "fiscal_order1" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","invoice_status":"auth ok","auth_time":1609459200}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetFiscalAuthData(context.Background(), &GetFiscalAuthDataRequest{
		OrderID: "fiscal_order1", SPAppID: "sp_fiscal",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.InvoiceStatus != "auth ok" {
		t.Errorf("unexpected invoice_status: %q", resp.InvoiceStatus)
	}
}

func TestGetTicket_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/cgi-bin/ticket/getticket") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "wx_card" {
			t.Errorf("expected type=wx_card, got %q", r.URL.Query().Get("type"))
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","ticket":"FISCAL_TICKET","expires_in":7200}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetTicket(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Ticket != "FISCAL_TICKET" {
		t.Errorf("expected FISCAL_TICKET, got %q", resp.Ticket)
	}
}

func TestRejectInsertFiscal_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/rejectinsert") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["order_id"] != "fiscal_order1" || body["reason"] != "invalid" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.RejectInsertFiscal(context.Background(), &RejectInsertFiscalRequest{
		SPAppID: "sp1", OrderID: "fiscal_order1", Reason: "invalid",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestSetFiscalInvoiceUrl_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/seturl") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","invoice_url":"https://example.com/fiscal"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.SetFiscalInvoiceUrl(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.InvoiceUrl != "https://example.com/fiscal" {
		t.Errorf("unexpected invoice_url: %q", resp.InvoiceUrl)
	}
}

func TestGetPlatformPdf_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/platform/getpdf") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["s_media_id"] != "fiscal_media1" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","pdf_url":"https://example.com/fiscal.pdf","pdf_url_expire_time":7200}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetPlatformPdf(context.Background(), &GetPlatformPdfRequest{
		Action: "get_url", SMediaID: "fiscal_media1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.PdfUrl != "https://example.com/fiscal.pdf" {
		t.Errorf("unexpected pdf_url: %q", resp.PdfUrl)
	}
}

func TestUpdateInvoicePlatformStatus_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/platform/updatestatus") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["card_id"] != "card_fiscal" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.UpdateInvoicePlatformStatus(context.Background(), &UpdateInvoicePlatformStatusRequest{
		CardID: "card_fiscal", Code: "code1", ReimburseStatus: "INVOICE_REIMBURSE_INIT",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestGetFiscalAuthUrl_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/nontax/getbillauthurl") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["s_pappid"] != "fiscal_sp" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","auth_url":"https://example.com/fiscal_auth","expire_time":3600}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetFiscalAuthUrl(context.Background(), &GetFiscalAuthUrlRequest{
		SPAppID: "fiscal_sp", OrderID: "order1", Money: 100, Timestamp: 1609459200, Source: "web", Ticket: "tk1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.AuthUrl != "https://example.com/fiscal_auth" {
		t.Errorf("unexpected auth_url: %q", resp.AuthUrl)
	}
}

func TestCreateFiscalCard_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/nontax/createbillcard") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["payee"] != "Fiscal Bureau" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","card_id":"fiscal_card1"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.CreateFiscalCard(context.Background(), &CreateFiscalCardRequest{
		BaseInfo: &CreateFiscalCardBaseInfo{LogoUrl: "https://example.com/logo.png"},
		Payee:    "Fiscal Bureau",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.CardID != "fiscal_card1" {
		t.Errorf("expected fiscal_card1, got %q", resp.CardID)
	}
}

func TestInsertFiscalInvoice_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/nontax/insertbill") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["order_id"] != "fiscal_order1" || body["card_id"] != "fiscal_card1" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","code":"fiscal_code","openid":"o1"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.InsertFiscalInvoice(context.Background(), &InsertFiscalInvoiceRequest{
		OrderID: "fiscal_order1", CardID: "fiscal_card1", AppID: "wx123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Code != "fiscal_code" {
		t.Errorf("expected fiscal_code, got %q", resp.Code)
	}
}

// ---------- Invoice name ----------

func TestGetUserTitleUrl_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/biz/getusertitleurl") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["title"] != "Test Corp" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","url":"https://example.com/title"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetUserTitleUrl(context.Background(), &GetUserTitleUrlRequest{
		Title: "Test Corp", TaxNo: "123456789012345",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Url != "https://example.com/title" {
		t.Errorf("unexpected url: %q", resp.Url)
	}
}

func TestGetSelectTitleUrl_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/biz/getselecttitleurl") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["biz_name"] != "My Shop" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","url":"https://example.com/select"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.GetSelectTitleUrl(context.Background(), &GetSelectTitleUrlRequest{
		BizName: "My Shop",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Url != "https://example.com/select" {
		t.Errorf("unexpected url: %q", resp.Url)
	}
}

func TestScanTitle_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		if !strings.HasSuffix(r.URL.Path, "/card/invoice/scantitle") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["scan_text"] != "qrcode_data" {
			t.Errorf("unexpected body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok","title_type":0,"title":"Test Corp","tax_no":"123456789012345"}`))
	}))
	defer srv.Close()

	c := newBaseTestClient(t, srv)
	resp, err := c.ScanTitle(context.Background(), &ScanTitleRequest{
		ScanText: "qrcode_data",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Title != "Test Corp" {
		t.Errorf("expected Test Corp, got %q", resp.Title)
	}
}
