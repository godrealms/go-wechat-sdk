package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCheckinData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/checkin/getcheckindata" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if int(body["opencheckindatatype"].(float64)) != 3 {
			t.Errorf("opencheckindatatype: %v", body["opencheckindatatype"])
		}
		if int64(body["starttime"].(float64)) != 1609430400 {
			t.Errorf("starttime: %v", body["starttime"])
		}
		if int64(body["endtime"].(float64)) != 1609516800 {
			t.Errorf("endtime: %v", body["endtime"])
		}
		users := body["useridlist"].([]any)
		if len(users) != 1 || users[0] != "user1" {
			t.Errorf("useridlist: %v", users)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"checkindata": []map[string]any{
				{
					"userid":       "user1",
					"groupname":    "标准打卡",
					"checkin_type": "上班打卡",
					"checkin_time": 1609430401,
				},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetCheckinData(context.Background(), &GetCheckinDataReq{
		OpenCheckinDataType: 3,
		StartTime:           1609430400,
		EndTime:             1609516800,
		UserIDList:          []string{"user1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.CheckinData) != 1 {
		t.Fatalf("checkindata count: %d", len(resp.CheckinData))
	}
	d := resp.CheckinData[0]
	if d.UserID != "user1" {
		t.Errorf("userid: %q", d.UserID)
	}
	if d.GroupName != "标准打卡" {
		t.Errorf("groupname: %q", d.GroupName)
	}
	if d.CheckinType != "上班打卡" {
		t.Errorf("checkin_type: %q", d.CheckinType)
	}
	if d.CheckinTime != 1609430401 {
		t.Errorf("checkin_time: %d", d.CheckinTime)
	}
}

func TestGetCheckinOption(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/checkin/getcheckinoption" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if int64(body["datetime"].(float64)) != 1609430400 {
			t.Errorf("datetime: %v", body["datetime"])
		}
		users := body["useridlist"].([]any)
		if len(users) != 1 || users[0] != "user1" {
			t.Errorf("useridlist: %v", users)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"info": []map[string]any{
				{
					"userid": "user1",
					"group": map[string]any{
						"groupid":   101,
						"groupname": "研发组",
						"grouptype": 1,
					},
				},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetCheckinOption(context.Background(), &GetCheckinOptionReq{
		DateTime:   1609430400,
		UserIDList: []string{"user1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Info) != 1 {
		t.Fatalf("info count: %d", len(resp.Info))
	}
	opt := resp.Info[0]
	if opt.UserID != "user1" {
		t.Errorf("userid: %q", opt.UserID)
	}
	if opt.Group.GroupID != 101 {
		t.Errorf("groupid: %d", opt.Group.GroupID)
	}
	if opt.Group.GroupName != "研发组" {
		t.Errorf("groupname: %q", opt.Group.GroupName)
	}
	if opt.Group.GroupType != 1 {
		t.Errorf("grouptype: %d", opt.Group.GroupType)
	}
}

func TestGetCheckinDayData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/checkin/getcheckin_daydata" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if int64(body["starttime"].(float64)) != 1609430400 {
			t.Errorf("starttime: %v", body["starttime"])
		}
		if int64(body["endtime"].(float64)) != 1609516800 {
			t.Errorf("endtime: %v", body["endtime"])
		}
		users := body["useridlist"].([]any)
		if len(users) != 1 || users[0] != "user1" {
			t.Errorf("useridlist: %v", users)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"datas": []map[string]any{
				{
					"base_info": map[string]any{
						"date":   1609430400,
						"name":   "张三",
						"acctid": "user1",
					},
					"summary_info": map[string]any{
						"checkin_count":     2,
						"regular_work_sec":  28800,
						"standard_work_sec": 28800,
					},
				},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetCheckinDayData(context.Background(), &GetCheckinDayDataReq{
		StartTime:  1609430400,
		EndTime:    1609516800,
		UserIDList: []string{"user1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Datas) != 1 {
		t.Fatalf("datas count: %d", len(resp.Datas))
	}
	d := resp.Datas[0]
	if d.BaseInfo.Name != "张三" {
		t.Errorf("base_info.name: %q", d.BaseInfo.Name)
	}
	if d.BaseInfo.AcctID != "user1" {
		t.Errorf("base_info.acctid: %q", d.BaseInfo.AcctID)
	}
	if d.SummaryInfo.CheckinCount != 2 {
		t.Errorf("summary_info.checkin_count: %d", d.SummaryInfo.CheckinCount)
	}
	if d.SummaryInfo.RegularWorkSec != 28800 {
		t.Errorf("summary_info.regular_work_sec: %d", d.SummaryInfo.RegularWorkSec)
	}
	if d.SummaryInfo.StandardWorkSec != 28800 {
		t.Errorf("summary_info.standard_work_sec: %d", d.SummaryInfo.StandardWorkSec)
	}
}

func TestGetCheckinMonthData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/checkin/getcheckin_monthdata" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if int64(body["starttime"].(float64)) != 1609430400 {
			t.Errorf("starttime: %v", body["starttime"])
		}
		if int64(body["endtime"].(float64)) != 1612108800 {
			t.Errorf("endtime: %v", body["endtime"])
		}
		users := body["useridlist"].([]any)
		if len(users) != 1 || users[0] != "user1" {
			t.Errorf("useridlist: %v", users)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"datas": []map[string]any{
				{
					"base_info": map[string]any{
						"date":   1609430400,
						"name":   "张三",
						"acctid": "user1",
					},
					"summary_info": map[string]any{
						"work_days":        22,
						"regular_work_sec": 633600,
						"except_days":      1,
					},
				},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetCheckinMonthData(context.Background(), &GetCheckinMonthDataReq{
		StartTime:  1609430400,
		EndTime:    1612108800,
		UserIDList: []string{"user1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Datas) != 1 {
		t.Fatalf("datas count: %d", len(resp.Datas))
	}
	d := resp.Datas[0]
	if d.BaseInfo.Name != "张三" {
		t.Errorf("base_info.name: %q", d.BaseInfo.Name)
	}
	if d.BaseInfo.AcctID != "user1" {
		t.Errorf("base_info.acctid: %q", d.BaseInfo.AcctID)
	}
	if d.SummaryInfo.WorkDays != 22 {
		t.Errorf("summary_info.work_days: %d", d.SummaryInfo.WorkDays)
	}
	if d.SummaryInfo.RegularWorkSec != 633600 {
		t.Errorf("summary_info.regular_work_sec: %d", d.SummaryInfo.RegularWorkSec)
	}
	if d.SummaryInfo.ExceptDays != 1 {
		t.Errorf("summary_info.except_days: %d", d.SummaryInfo.ExceptDays)
	}
}
