package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetApprovalTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/oa/gettemplatedetail" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["template_id"] != "tpl001" {
			t.Errorf("body.template_id: %v", body["template_id"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"template_names": []map[string]any{
				{"text": "请假申请", "lang": "zh_CN"},
			},
			"template_content": map[string]any{
				"controls": []map[string]any{
					{
						"property": map[string]any{
							"control": "Text",
							"id":      "ctrl_1",
							"title": []map[string]any{
								{"text": "事由", "lang": "zh_CN"},
							},
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetApprovalTemplate(context.Background(), "tpl001")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.TemplateNames) != 1 {
		t.Fatalf("len(TemplateNames): %d", len(resp.TemplateNames))
	}
	if resp.TemplateNames[0].Text != "请假申请" {
		t.Errorf("TemplateNames[0].Text: %q", resp.TemplateNames[0].Text)
	}
	if len(resp.TemplateContent.Controls) != 1 {
		t.Fatalf("len(Controls): %d", len(resp.TemplateContent.Controls))
	}
	if resp.TemplateContent.Controls[0].Property.Control != "Text" {
		t.Errorf("Controls[0].Property.Control: %q", resp.TemplateContent.Controls[0].Property.Control)
	}
	if resp.TemplateContent.Controls[0].Property.ID != "ctrl_1" {
		t.Errorf("Controls[0].Property.ID: %q", resp.TemplateContent.Controls[0].Property.ID)
	}
}

func TestApplyEvent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/oa/applyevent" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["creator_userid"] != "zhangsan" {
			t.Errorf("body.creator_userid: %v", body["creator_userid"])
		}
		if body["template_id"] != "tpl001" {
			t.Errorf("body.template_id: %v", body["template_id"])
		}
		applyData, ok := body["apply_data"].(map[string]any)
		if !ok {
			t.Errorf("body.apply_data missing or wrong type")
		} else {
			contents, ok2 := applyData["contents"].([]any)
			if !ok2 || len(contents) == 0 {
				t.Errorf("apply_data.contents: %v", applyData["contents"])
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"sp_no":   "202001010001",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	req := &ApplyEventReq{
		CreatorUserID:       "zhangsan",
		TemplateID:          "tpl001",
		UseTemplateApprover: 1,
		ApplyData: ApplyData{
			Contents: []ApplyContent{
				{
					Control: "Text",
					ID:      "ctrl_1",
					Value:   ApplyValue{Text: "年假"},
				},
			},
		},
	}
	resp, err := cc.ApplyEvent(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.SpNo != "202001010001" {
		t.Errorf("SpNo: %q", resp.SpNo)
	}
}

func TestGetApprovalDetail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/oa/getapprovaldetail" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["sp_no"] != "202001010001" {
			t.Errorf("body.sp_no: %v", body["sp_no"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"info": map[string]any{
				"sp_no":       "202001010001",
				"sp_name":     "请假申请",
				"sp_status":   2,
				"template_id": "tpl001",
				"apply_time":  int64(1577836800),
				"applyer": map[string]any{
					"userid":  "zhangsan",
					"partyid": "1",
				},
				"sp_record": []map[string]any{
					{
						"sp_status":    2,
						"approverattr": 1,
						"details": []map[string]any{
							{
								"approver": map[string]any{
									"userid": "lisi",
								},
								"speech":    "同意",
								"sp_status": 2,
								"sptime":    int64(1577836900),
							},
						},
					},
				},
				"apply_data": map[string]any{
					"contents": []any{},
				},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetApprovalDetail(context.Background(), "202001010001")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Info.SpNo != "202001010001" {
		t.Errorf("Info.SpNo: %q", resp.Info.SpNo)
	}
	if resp.Info.SpName != "请假申请" {
		t.Errorf("Info.SpName: %q", resp.Info.SpName)
	}
	if resp.Info.SpStatus != 2 {
		t.Errorf("Info.SpStatus: %d", resp.Info.SpStatus)
	}
	if resp.Info.Applyer.UserID != "zhangsan" {
		t.Errorf("Info.Applyer.UserID: %q", resp.Info.Applyer.UserID)
	}
	if len(resp.Info.SpRecord) != 1 {
		t.Fatalf("len(SpRecord): %d", len(resp.Info.SpRecord))
	}
	if len(resp.Info.SpRecord[0].Details) != 1 {
		t.Fatalf("len(SpRecord[0].Details): %d", len(resp.Info.SpRecord[0].Details))
	}
	if resp.Info.SpRecord[0].Details[0].Approver.UserID != "lisi" {
		t.Errorf("SpRecord[0].Details[0].Approver.UserID: %q", resp.Info.SpRecord[0].Details[0].Approver.UserID)
	}
}

func TestGetApprovalData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/oa/getapprovalinfo" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["starttime"] == nil {
			t.Errorf("body.starttime missing")
		}
		if body["endtime"] == nil {
			t.Errorf("body.endtime missing")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode":         0,
			"errmsg":          "ok",
			"sp_no_list":      []string{"202001010001", "202001010002"},
			"new_next_cursor": 2,
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	req := &GetApprovalDataReq{
		StartTime: 1577836800,
		EndTime:   1577923200,
	}
	resp, err := cc.GetApprovalData(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.SpNoList) != 2 {
		t.Fatalf("len(SpNoList): %d", len(resp.SpNoList))
	}
	if resp.SpNoList[0] != "202001010001" {
		t.Errorf("SpNoList[0]: %q", resp.SpNoList[0])
	}
	if resp.SpNoList[1] != "202001010002" {
		t.Errorf("SpNoList[1]: %q", resp.SpNoList[1])
	}
	if resp.NewNextCursor != 2 {
		t.Errorf("NewNextCursor: %d", resp.NewNextCursor)
	}
}
