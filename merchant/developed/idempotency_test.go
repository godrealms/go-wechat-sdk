package pay

import (
	"strings"
	"testing"
)

func TestRequireIdempotencyKey_NilBody(t *testing.T) {
	if err := requireIdempotencyKey(nil, "out_batch_no"); err == nil {
		t.Error("expected error for nil body")
	}
}

func TestRequireIdempotencyKey_MapAny_Present(t *testing.T) {
	body := map[string]any{"out_batch_no": "BATCH001", "total_amount": 100}
	if err := requireIdempotencyKey(body, "out_batch_no"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRequireIdempotencyKey_MapAny_Empty(t *testing.T) {
	body := map[string]any{"out_batch_no": "", "total_amount": 100}
	err := requireIdempotencyKey(body, "out_batch_no")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
	if !strings.Contains(err.Error(), "non-empty") {
		t.Errorf("error should mention non-empty: %v", err)
	}
}

func TestRequireIdempotencyKey_MapAny_Missing(t *testing.T) {
	body := map[string]any{"total_amount": 100}
	err := requireIdempotencyKey(body, "out_batch_no")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("error should say required: %v", err)
	}
}

func TestRequireIdempotencyKey_MapAny_Nil(t *testing.T) {
	body := map[string]any{"out_batch_no": nil}
	if err := requireIdempotencyKey(body, "out_batch_no"); err == nil {
		t.Error("expected error for nil value")
	}
}

func TestRequireIdempotencyKey_MapString(t *testing.T) {
	if err := requireIdempotencyKey(map[string]string{"out_order_no": "X"}, "out_order_no"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := requireIdempotencyKey(map[string]string{"out_order_no": ""}, "out_order_no"); err == nil {
		t.Error("expected error for empty value")
	}
	if err := requireIdempotencyKey(map[string]string{}, "out_order_no"); err == nil {
		t.Error("expected error for missing key")
	}
}

type transferReq struct {
	OutBatchNo string `json:"out_batch_no"`
	Total      int    `json:"total_amount,omitempty"`
}

func TestRequireIdempotencyKey_StructPresent(t *testing.T) {
	body := &transferReq{OutBatchNo: "B1", Total: 100}
	if err := requireIdempotencyKey(body, "out_batch_no"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRequireIdempotencyKey_StructEmpty(t *testing.T) {
	body := &transferReq{}
	err := requireIdempotencyKey(body, "out_batch_no")
	if err == nil {
		t.Fatal("expected error for empty struct field")
	}
	if !strings.Contains(err.Error(), "non-empty") {
		t.Errorf("error should mention non-empty: %v", err)
	}
}

func TestRequireIdempotencyKey_StructByValue(t *testing.T) {
	body := transferReq{OutBatchNo: "B1"}
	if err := requireIdempotencyKey(body, "out_batch_no"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRequireIdempotencyKey_StructPointerNil(t *testing.T) {
	var body *transferReq
	err := requireIdempotencyKey(body, "out_batch_no")
	if err == nil {
		t.Fatal("expected error for nil pointer")
	}
}

// TestRequireIdempotencyKey_StructFieldNotInJSONTag verifies that if the
// struct doesn't expose a json tag matching the key, we fall back to the
// marshal-probe path. The struct here has the field with a different json
// name; the helper should NOT find an "out_batch_no" via the tag scan and
// should fall through to the probe.
type oddTagReq struct {
	WrongName string `json:"out_order_no"` // not what we'll search for
}

func TestRequireIdempotencyKey_FallbackToMarshalProbe(t *testing.T) {
	body := oddTagReq{WrongName: "X"}
	// Looking for "out_batch_no" — neither the tag nor the marshalled JSON
	// contains it, so we expect "is required".
	err := requireIdempotencyKey(body, "out_batch_no")
	if err == nil {
		t.Fatal("expected error when neither tag nor marshalled JSON contains the key")
	}
}

// TestRequireIdempotencyKey_NonObjectBodyAllowsThrough verifies the conservative
// "if we can't validate, allow through" behaviour. A slice marshals to a JSON
// array, not an object, so we cannot look up keys.
func TestRequireIdempotencyKey_NonObjectBodyAllowsThrough(t *testing.T) {
	body := []int{1, 2, 3}
	if err := requireIdempotencyKey(body, "out_batch_no"); err != nil {
		t.Errorf("expected nil (conservative allow-through), got: %v", err)
	}
}
