package pay

import (
	"context"
	"strings"
	"testing"
)

func TestQueryTransferBatch_EmptyBatchId(t *testing.T) {
	c, _, srv := newClientWithFakeServer(t)
	defer srv.Close()

	_, err := c.QueryTransferBatch(context.Background(), "", false, 0, 10)
	if err == nil {
		t.Fatal("expected error for empty batchId")
	}
	if !strings.Contains(err.Error(), "batchId is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestQueryTransferDetail_EmptyIds(t *testing.T) {
	c, _, srv := newClientWithFakeServer(t)
	defer srv.Close()

	// empty batchId
	_, err := c.QueryTransferDetail(context.Background(), "", "detail-001")
	if err == nil {
		t.Fatal("expected error for empty batchId")
	}
	if !strings.Contains(err.Error(), "batchId and detailId are required") {
		t.Errorf("unexpected error message: %v", err)
	}

	// empty detailId
	_, err = c.QueryTransferDetail(context.Background(), "batch-001", "")
	if err == nil {
		t.Fatal("expected error for empty detailId")
	}
	if !strings.Contains(err.Error(), "batchId and detailId are required") {
		t.Errorf("unexpected error message: %v", err)
	}
}
