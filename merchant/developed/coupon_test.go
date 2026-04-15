package pay

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestClient_FavorCreateStock_HappyPath(t *testing.T) {
	client, fs, srv := newClientWithFakeServer(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"stock_id":"9856888"}`)
	}

	resp, err := client.FavorCreateStock(context.Background(), map[string]any{
		"stock_name":          "test stock",
		"belong_merchant":     "1900000001",
		"available_begin_time": "2026-04-14T00:00:00+08:00",
	})
	if err != nil {
		t.Fatalf("FavorCreateStock failed: %v", err)
	}
	if resp["stock_id"] != "9856888" {
		t.Errorf("unexpected stock_id: %v", resp["stock_id"])
	}
	if !strings.Contains(fs.requests[0].Path, "/v3/marketing/favor/coupon-stocks") {
		t.Errorf("unexpected path: %s", fs.requests[0].Path)
	}
	if fs.requests[0].Method != http.MethodPost {
		t.Errorf("unexpected method: %s", fs.requests[0].Method)
	}
}

func TestClient_FavorStartStock_EmptyStockId(t *testing.T) {
	client, _, srv := newClientWithFakeServer(t)
	defer srv.Close()
	if _, err := client.FavorStartStock(context.Background(), "", nil); err == nil {
		t.Fatal("expected error for empty stockId")
	}
}
