package offiaccount

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type withResp struct {
	Resp
	Data string
}

type withoutResp struct {
	Data string
}

func TestCheckEmbeddedResp_NilInputs(t *testing.T) {
	if err := checkEmbeddedResp(nil); err != nil {
		t.Errorf("nil any: %v", err)
	}
	var p *withResp
	if err := checkEmbeddedResp(p); err != nil {
		t.Errorf("typed nil pointer: %v", err)
	}
}

func TestCheckEmbeddedResp_OkStruct(t *testing.T) {
	r := &withResp{Resp: Resp{ErrCode: 0, ErrMsg: "ok"}}
	if err := checkEmbeddedResp(r); err != nil {
		t.Errorf("zero errcode: %v", err)
	}
}

func TestCheckEmbeddedResp_ErrcodeSurfaced(t *testing.T) {
	r := &withResp{Resp: Resp{ErrCode: 40013, ErrMsg: "invalid appid"}}
	err := checkEmbeddedResp(r)
	if err == nil {
		t.Fatal("expected errcode error")
	}
	var wx *WeixinError
	if !errors.As(err, &wx) {
		t.Fatalf("expected *WeixinError, got %T: %v", err, err)
	}
	if wx.ErrCode != 40013 {
		t.Errorf("ErrCode = %d, want 40013", wx.ErrCode)
	}
}

func TestCheckEmbeddedResp_NoRespField(t *testing.T) {
	r := &withoutResp{Data: "x"}
	if err := checkEmbeddedResp(r); err != nil {
		t.Errorf("non-Resp struct should pass through: %v", err)
	}
}

// TestDoPostRaw_SendsBodyVerbatim verifies that doPostRaw writes the raw body
// bytes (no JSON wrapping) with the caller-supplied Content-Type.
func TestDoPostRaw_SendsBodyVerbatim(t *testing.T) {
	const wantBody = "hello world"
	const wantCT = "text/plain; charset=utf-8"

	var gotBody string
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		gotCT = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "a", AppSecret: "s"})
	var resp Resp
	if err := c.doPostRaw(context.Background(), "/x", []byte(wantBody), wantCT, &resp); err != nil {
		t.Fatalf("doPostRaw: %v", err)
	}
	if gotBody != wantBody {
		t.Errorf("body = %q, want %q", gotBody, wantBody)
	}
	if gotCT != wantCT {
		t.Errorf("content-type = %q, want %q", gotCT, wantCT)
	}
}

// TestDoPostRaw_EmptyBody verifies the img_url-only path: nil body and no
// Content-Type, which is what /cv/ocr/* callers rely on when a URL is supplied.
func TestDoPostRaw_EmptyBody(t *testing.T) {
	var gotMethod string
	var gotCL int64
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotCL = r.ContentLength
		gotCT = r.Header.Get("Content-Type")
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "a", AppSecret: "s"})
	var resp Resp
	if err := c.doPostRaw(context.Background(), "/x", nil, "", &resp); err != nil {
		t.Fatalf("doPostRaw nil body: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotCL != 0 {
		t.Errorf("content-length = %d, want 0", gotCL)
	}
	if gotCT != "" {
		// WeChat doesn't care when there's no body, but we promise in the
		// ocrRequest docs that we don't lie about the body type.
		t.Errorf("unexpected content-type on empty body: %q", gotCT)
	}
}

// TestDoPostRaw_SurfacesErrcode verifies the embedded-Resp check fires.
func TestDoPostRaw_SurfacesErrcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"bad token"}`))
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "a", AppSecret: "s"})
	var resp Resp
	err := c.doPostRaw(context.Background(), "/x", []byte("x"), "application/octet-stream", &resp)
	if err == nil {
		t.Fatal("expected *WeixinError, got nil")
	}
	var wx *WeixinError
	if !errors.As(err, &wx) {
		t.Fatalf("expected *WeixinError, got %T: %v", err, err)
	}
	if wx.ErrCode != 40001 {
		t.Errorf("ErrCode = %d, want 40001", wx.ErrCode)
	}
}

// TestDoPostMultipartFile verifies that doPostMultipartFile produces a valid
// multipart/form-data request with the expected field name, filename, and
// file body.
func TestDoPostMultipartFile(t *testing.T) {
	const wantField = "img"
	const wantFilename = "image"
	wantData := []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x01, 0x02, 0x03} // fake PNG header + payload

	var gotField, gotFilename string
	var gotData []byte
	var gotCT string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(gotCT)
		if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
			t.Errorf("content-type = %q, want multipart/form-data", gotCT)
			w.Write([]byte(`{"errcode":-1}`))
			return
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("next part: %v", err)
			}
			gotField = part.FormName()
			gotFilename = part.FileName()
			gotData, _ = io.ReadAll(part)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "a", AppSecret: "s"})
	var resp Resp
	if err := c.doPostMultipartFile(context.Background(), "/x", wantField, wantFilename, wantData, &resp); err != nil {
		t.Fatalf("doPostMultipartFile: %v", err)
	}
	if gotField != wantField {
		t.Errorf("field = %q, want %q", gotField, wantField)
	}
	if gotFilename != wantFilename {
		t.Errorf("filename = %q, want %q", gotFilename, wantFilename)
	}
	if string(gotData) != string(wantData) {
		t.Errorf("data mismatch: got %x, want %x", gotData, wantData)
	}
}

// TestDoPostMultipartFile_SurfacesErrcode verifies embedded-Resp errcode
// propagation through the multipart helper.
func TestDoPostMultipartFile_SurfacesErrcode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40007,"errmsg":"invalid media_id"}`))
	}))
	defer srv.Close()

	c := newClientWithBaseURL(srv.URL, &Config{AppId: "a", AppSecret: "s"})
	var resp Resp
	err := c.doPostMultipartFile(context.Background(), "/x", "img", "x.jpg", []byte{1, 2, 3}, &resp)
	if err == nil {
		t.Fatal("expected *WeixinError, got nil")
	}
	var wx *WeixinError
	if !errors.As(err, &wx) {
		t.Fatalf("expected *WeixinError, got %T: %v", err, err)
	}
	if wx.ErrCode != 40007 {
		t.Errorf("ErrCode = %d, want 40007", wx.ErrCode)
	}
}
