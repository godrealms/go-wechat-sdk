package pay

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestClient_CombineTransactionsJsapi_HappyPath(t *testing.T) {
	client, fs, srv := newClientWithFakeServer(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"prepay_id":"wx_combine_123"}`)
	}

	resp, err := client.CombineTransactionsJsapi(context.Background(), map[string]any{
		"combine_appid":      "wxtest",
		"combine_mchid":      "1900000001",
		"combine_out_trade_no": "OUT_COMBINE_123",
	})
	if err != nil {
		t.Fatalf("CombineTransactionsJsapi failed: %v", err)
	}
	if resp["prepay_id"] != "wx_combine_123" {
		t.Errorf("unexpected prepay_id: %v", resp["prepay_id"])
	}
	if len(fs.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(fs.requests))
	}
	if !strings.Contains(fs.requests[0].Path, "/v3/combine-transactions/jsapi") {
		t.Errorf("unexpected path: %s", fs.requests[0].Path)
	}
	if fs.requests[0].Method != http.MethodPost {
		t.Errorf("unexpected method: %s", fs.requests[0].Method)
	}
}

func TestClient_QueryCombineOrder_EmptyArg(t *testing.T) {
	client, _, srv := newClientWithFakeServer(t)
	defer srv.Close()
	if _, err := client.QueryCombineOrder(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty combineOutTradeNo")
	}
}
