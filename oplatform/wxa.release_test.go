package oplatform

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWxaAdmin_SubmitAudit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/submit_audit") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"auditid":1234567890}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.SubmitAudit(context.Background(), &WxaSubmitAuditReq{
		VersionDesc: "bugfix",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.AuditID != 1234567890 {
		t.Errorf("audit_id: %d", resp.AuditID)
	}
}

func TestWxaAdmin_SubmitAudit_InProgress(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":85013,"errmsg":"invalid version"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	_, err := w.SubmitAudit(context.Background(), &WxaSubmitAuditReq{})
	var werr *WeixinError
	if !errors.As(err, &werr) || werr.ErrCode != 85013 {
		t.Errorf("expected 85013, got %v", err)
	}
}

func TestWxaAdmin_GetAuditStatus(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/get_auditstatus") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"status":2,"reason":"违反规则"}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	status, err := w.GetAuditStatus(context.Background(), 9999)
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != 2 || status.Reason != "违反规则" {
		t.Errorf("unexpected: %+v", status)
	}
	if int(body["auditid"].(float64)) != 9999 {
		t.Errorf("body auditid: %+v", body)
	}
}

func TestWxaAdmin_GetLatestAuditStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/get_latest_auditstatus") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"auditid":111,"status":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	status, err := w.GetLatestAuditStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.AuditID != 111 || status.Status != 0 {
		t.Errorf("unexpected: %+v", status)
	}
}

func TestWxaAdmin_UndoCodeAudit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/undocodeaudit") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.UndoCodeAudit(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_SpeedupAudit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/speedupaudit") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.SpeedupAudit(context.Background(), 1234); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_Release(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/release") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.Release(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_RevertCodeRelease(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/revertcoderelease") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.RevertCodeRelease(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestWxaAdmin_ChangeVisitStatus(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/wxa/change_visitstatus") {
			t.Errorf("path: %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.ChangeVisitStatus(context.Background(), "close"); err != nil {
		t.Fatal(err)
	}
	if body["action"] != "close" {
		t.Errorf("body: %+v", body)
	}
}

func TestWxaAdmin_GetSupportVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/getweappsupportversion") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"now_version":"2.10.0","uv_info":{"items":[{"percentage":95.5,"version":"2.10.0"}]}}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	resp, err := w.GetSupportVersion(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.NowVersion != "2.10.0" || len(resp.UVInfo.Items) != 1 {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestWxaAdmin_SetSupportVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/wxopen/setweappsupportversion") {
			t.Errorf("path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	w := newTestWxaAdmin(t, srv.URL)

	if err := w.SetSupportVersion(context.Background(), "2.10.0"); err != nil {
		t.Fatal(err)
	}
}
