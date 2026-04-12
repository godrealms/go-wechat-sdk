package mini_program

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func TestGetPhoneNumber(t *testing.T) {
	srv := httptest.NewServer(tokenHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wxa/business/getuserphonenumber" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Query().Get("access_token") != "TOK" {
			t.Errorf("missing or wrong access_token: %q", r.URL.Query().Get("access_token"))
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("invalid JSON body: %v", err)
		}
		if _, ok := req["code"]; !ok {
			t.Error("body missing 'code' field")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"errcode": 0,
			"errmsg": "ok",
			"phone_info": {
				"phoneNumber": "+8613800138000",
				"purePhoneNumber": "13800138000",
				"countryCode": "86",
				"watermark": {
					"appid": "wx123456",
					"timestamp": 1680000000
				}
			}
		}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{AppId: "wx", AppSecret: "sec"},
		WithHTTP(utils.NewHTTP(srv.URL, utils.WithTimeout(time.Second*3))))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.GetPhoneNumber(context.Background(), "test-code-abc")
	if err != nil {
		t.Fatal(err)
	}
	if resp.PhoneInfo.PhoneNumber != "+8613800138000" {
		t.Errorf("unexpected phoneNumber: %q", resp.PhoneInfo.PhoneNumber)
	}
	if resp.PhoneInfo.PurePhoneNumber != "13800138000" {
		t.Errorf("unexpected purePhoneNumber: %q", resp.PhoneInfo.PurePhoneNumber)
	}
	if resp.PhoneInfo.CountryCode != "86" {
		t.Errorf("unexpected countryCode: %q", resp.PhoneInfo.CountryCode)
	}
	if resp.PhoneInfo.Watermark.AppID != "wx123456" {
		t.Errorf("unexpected watermark appid: %q", resp.PhoneInfo.Watermark.AppID)
	}
	if resp.PhoneInfo.Watermark.Timestamp != 1680000000 {
		t.Errorf("unexpected watermark timestamp: %d", resp.PhoneInfo.Watermark.Timestamp)
	}
}
