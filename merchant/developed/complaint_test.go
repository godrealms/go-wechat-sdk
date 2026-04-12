package pay

import (
	"context"
	"strings"
	"testing"
)

func TestListComplaints_EmptyDates(t *testing.T) {
	c, _, srv := newClientWithFakeServer(t)
	defer srv.Close()

	// empty beginDate
	_, err := c.ListComplaints(context.Background(), "", "2024-01-31", 0, 10)
	if err == nil {
		t.Fatal("expected error for empty beginDate")
	}
	if !strings.Contains(err.Error(), "beginDate and endDate are required") {
		t.Errorf("unexpected error message: %v", err)
	}

	// empty endDate
	_, err = c.ListComplaints(context.Background(), "2024-01-01", "", 0, 10)
	if err == nil {
		t.Fatal("expected error for empty endDate")
	}
	if !strings.Contains(err.Error(), "beginDate and endDate are required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGetComplaint_EmptyId(t *testing.T) {
	c, _, srv := newClientWithFakeServer(t)
	defer srv.Close()

	_, err := c.GetComplaint(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty complaintId")
	}
	if !strings.Contains(err.Error(), "complaintId is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestResponseComplaint_EmptyId(t *testing.T) {
	c, _, srv := newClientWithFakeServer(t)
	defer srv.Close()

	err := c.ResponseComplaint(context.Background(), "", map[string]string{"response_content": "已处理"})
	if err == nil {
		t.Fatal("expected error for empty complaintId")
	}
	if !strings.Contains(err.Error(), "complaintId is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCompleteComplaint_EmptyId(t *testing.T) {
	c, _, srv := newClientWithFakeServer(t)
	defer srv.Close()

	err := c.CompleteComplaint(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty complaintId")
	}
	if !strings.Contains(err.Error(), "complaintId is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}
