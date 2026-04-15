package pay

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestClient_ProfitSharingOrder_HappyPath(t *testing.T) {
	client, fs, srv := newClientWithFakeServer(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"order_id":"3008450740201411110007820472","state":"PROCESSING"}`)
	}

	resp, err := client.ProfitSharingOrder(context.Background(), map[string]any{
		"appid":          "wxtest",
		"transaction_id": "4200000000000000000000000001",
		"out_order_no":   "P00001",
	})
	if err != nil {
		t.Fatalf("ProfitSharingOrder failed: %v", err)
	}
	if resp["state"] != "PROCESSING" {
		t.Errorf("unexpected state: %v", resp["state"])
	}
	if !strings.Contains(fs.requests[0].Path, "/v3/profitsharing/orders") {
		t.Errorf("unexpected path: %s", fs.requests[0].Path)
	}
	if fs.requests[0].Method != http.MethodPost {
		t.Errorf("unexpected method: %s", fs.requests[0].Method)
	}
}

func TestClient_ProfitSharingQueryOrder_EmptyArgs(t *testing.T) {
	client, _, srv := newClientWithFakeServer(t)
	defer srv.Close()
	if _, err := client.ProfitSharingQueryOrder(context.Background(), "", ""); err == nil {
		t.Fatal("expected error for empty outOrderNo/transactionId")
	}
}
