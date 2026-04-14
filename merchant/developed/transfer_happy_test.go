package pay

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestClient_CreateTransferBatch_HappyPath(t *testing.T) {
	client, fs, srv := newClientWithFakeServer(t)
	defer srv.Close()
	fs.respond = func(r *http.Request) (int, []byte) {
		return 200, []byte(`{"out_batch_no":"B001","batch_id":"131000000000000000000000000001","create_time":"2026-04-14T12:00:00+08:00"}`)
	}

	resp, err := client.CreateTransferBatch(context.Background(), map[string]any{
		"appid":        "wxtest",
		"out_batch_no": "B001",
		"batch_name":   "test transfer",
	})
	if err != nil {
		t.Fatalf("CreateTransferBatch failed: %v", err)
	}
	if resp["out_batch_no"] != "B001" {
		t.Errorf("unexpected out_batch_no: %v", resp["out_batch_no"])
	}
	if !strings.Contains(fs.requests[0].Path, "/v3/transfer/batches") {
		t.Errorf("unexpected path: %s", fs.requests[0].Path)
	}
	if fs.requests[0].Method != http.MethodPost {
		t.Errorf("unexpected method: %s", fs.requests[0].Method)
	}
}
