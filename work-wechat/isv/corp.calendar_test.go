package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateCalendar(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/oa/calendar/add" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		cal := body["calendar"].(map[string]any)
		if cal["organizer"] != "user1" {
			t.Errorf("organizer: %v", cal["organizer"])
		}
		if cal["summary"] != "Team Meeting" {
			t.Errorf("summary: %v", cal["summary"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"cal_id":  "cal-xxx",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.CreateCalendar(context.Background(), &CreateCalendarReq{
		Calendar: Calendar{
			Organizer: "user1",
			Summary:   "Team Meeting",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.CalID != "cal-xxx" {
		t.Errorf("cal_id: %q", resp.CalID)
	}
}

func TestUpdateCalendar(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/oa/calendar/update" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		cal := body["calendar"].(map[string]any)
		if cal["cal_id"] != "cal-001" {
			t.Errorf("cal_id: %v", cal["cal_id"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.UpdateCalendar(context.Background(), &UpdateCalendarReq{
		Calendar: Calendar{
			CalID:     "cal-001",
			Organizer: "user1",
			Summary:   "Updated Meeting",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetCalendar(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/oa/calendar/get" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		ids := body["cal_id_list"].([]any)
		if len(ids) != 1 || ids[0] != "cal-001" {
			t.Errorf("cal_id_list: %v", ids)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"calendar_list": []map[string]any{
				{
					"cal_id":    "cal-001",
					"organizer": "user1",
					"summary":   "Team Meeting",
				},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetCalendar(context.Background(), []string{"cal-001"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.CalendarList) != 1 {
		t.Fatalf("calendar_list count: %d", len(resp.CalendarList))
	}
	cal := resp.CalendarList[0]
	if cal.Organizer != "user1" {
		t.Errorf("organizer: %q", cal.Organizer)
	}
	if cal.Summary != "Team Meeting" {
		t.Errorf("summary: %q", cal.Summary)
	}
}

func TestDeleteCalendar(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/oa/calendar/del" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["cal_id"] != "cal-001" {
			t.Errorf("cal_id: %v", body["cal_id"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.DeleteCalendar(context.Background(), "cal-001")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateSchedule(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/oa/schedule/add" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		sch := body["schedule"].(map[string]any)
		if sch["organizer"] != "user1" {
			t.Errorf("organizer: %v", sch["organizer"])
		}
		if sch["summary"] != "Sprint Review" {
			t.Errorf("summary: %v", sch["summary"])
		}
		if int64(sch["start_time"].(float64)) != 1700000000 {
			t.Errorf("start_time: %v", sch["start_time"])
		}
		if int64(sch["end_time"].(float64)) != 1700003600 {
			t.Errorf("end_time: %v", sch["end_time"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode":     0,
			"errmsg":      "ok",
			"schedule_id": "sch-xxx",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.CreateSchedule(context.Background(), &CreateScheduleReq{
		Schedule: Schedule{
			Organizer: "user1",
			Summary:   "Sprint Review",
			StartTime: 1700000000,
			EndTime:   1700003600,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ScheduleID != "sch-xxx" {
		t.Errorf("schedule_id: %q", resp.ScheduleID)
	}
}

func TestUpdateSchedule(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/oa/schedule/update" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		sch := body["schedule"].(map[string]any)
		if sch["schedule_id"] != "sch-001" {
			t.Errorf("schedule_id: %v", sch["schedule_id"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.UpdateSchedule(context.Background(), &UpdateScheduleReq{
		Schedule: Schedule{
			ScheduleID: "sch-001",
			Organizer:  "user1",
			Summary:    "Updated Review",
			StartTime:  1700000000,
			EndTime:    1700003600,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetSchedule(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/oa/schedule/get" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		ids := body["schedule_id_list"].([]any)
		if len(ids) != 1 || ids[0] != "sch-001" {
			t.Errorf("schedule_id_list: %v", ids)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"schedule_list": []map[string]any{
				{
					"schedule_id": "sch-001",
					"organizer":   "user1",
					"summary":     "Sprint Review",
					"start_time":  1700000000,
					"end_time":    1700003600,
				},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetSchedule(context.Background(), []string{"sch-001"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ScheduleList) != 1 {
		t.Fatalf("schedule_list count: %d", len(resp.ScheduleList))
	}
	sch := resp.ScheduleList[0]
	if sch.Organizer != "user1" {
		t.Errorf("organizer: %q", sch.Organizer)
	}
	if sch.StartTime != 1700000000 {
		t.Errorf("start_time: %d", sch.StartTime)
	}
}

func TestDeleteSchedule(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/oa/schedule/del" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["schedule_id"] != "sch-001" {
			t.Errorf("schedule_id: %v", body["schedule_id"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.DeleteSchedule(context.Background(), "sch-001")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetScheduleByCalendar(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if r.URL.Path != "/cgi-bin/oa/schedule/get_by_calendar" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["cal_id"] != "cal-001" {
			t.Errorf("cal_id: %v", body["cal_id"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0,
			"errmsg":  "ok",
			"schedule_list": []map[string]any{
				{
					"schedule_id": "sch-001",
					"organizer":   "user1",
					"summary":     "Sprint Review",
					"start_time":  1700000000,
					"end_time":    1700003600,
				},
				{
					"schedule_id": "sch-002",
					"organizer":   "user2",
					"summary":     "Planning",
					"start_time":  1700010000,
					"end_time":    1700013600,
				},
			},
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetScheduleByCalendar(context.Background(), &GetScheduleByCalendarReq{
		CalID: "cal-001",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ScheduleList) != 2 {
		t.Fatalf("schedule_list count: %d", len(resp.ScheduleList))
	}
}
