package offiaccount

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/godrealms/go-wechat-sdk/utils"
)

func newUserTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	h := utils.NewHTTP(srv.URL, utils.WithTimeout(3*time.Second))
	return NewClient(context.Background(), &Config{AppId: "test", AppSecret: "secret"},
		WithHTTPClient(h),
		WithTokenSource(fixedToken{"FAKE_TOKEN"}),
	)
}

func userOkServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(body))
	}))
}

func userClosedServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertAccessToken(t, r)
	}))
	srv.Close()
	return srv
}

func TestGetUserInfo_Success(t *testing.T) {
	body := `{"subscribe":1,"openid":"oUser123","language":"zh_CN"}`
	srv := userOkServer(t, body)
	defer srv.Close()

	c := newUserTestClient(t, srv)
	result, err := c.GetUserInfo(context.Background(), "oUser123", "zh_CN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Openid != "oUser123" {
		t.Errorf("expected openid oUser123, got %q", result.Openid)
	}
}

func TestGetUserInfo_NetworkError(t *testing.T) {
	c := newUserTestClient(t, userClosedServer(t))
	_, err := c.GetUserInfo(context.Background(), "oUser123", "")
	if err == nil {
		t.Error("expected network error")
	}
}

func TestGetFans_Success(t *testing.T) {
	body := `{"total":2,"count":2,"data":{"openid":["oUser1","oUser2"]},"next_openid":""}`
	srv := userOkServer(t, body)
	defer srv.Close()

	c := newUserTestClient(t, srv)
	result, err := c.GetFans(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected total=2, got %d", result.Total)
	}
}

func TestGetFans_NetworkError(t *testing.T) {
	c := newUserTestClient(t, userClosedServer(t))
	_, err := c.GetFans(context.Background(), "")
	if err == nil {
		t.Error("expected network error")
	}
}

func TestGetBlacklist_Success(t *testing.T) {
	body := `{"total":1,"count":1,"data":{"openid":["oBlack1"]},"next_openid":""}`
	srv := userOkServer(t, body)
	defer srv.Close()

	c := newUserTestClient(t, srv)
	result, err := c.GetBlacklist(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 1 {
		t.Errorf("expected count=1, got %d", result.Count)
	}
}

func TestBatchBlacklist_Success(t *testing.T) {
	srv := userOkServer(t, `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	c := newUserTestClient(t, srv)
	_, err := c.BatchBlacklist(context.Background(), []string{"oUser1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBatchUnblacklist_Success(t *testing.T) {
	srv := userOkServer(t, `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	c := newUserTestClient(t, srv)
	_, err := c.BatchUnblacklist(context.Background(), []string{"oUser1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateRemark_Success(t *testing.T) {
	srv := userOkServer(t, `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	c := newUserTestClient(t, srv)
	_, err := c.UpdateRemark(context.Background(), "oUser1", "my remark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBatchGetUserInfo_Success(t *testing.T) {
	body := `{"user_info_list":[{"openid":"oUser1"},{"openid":"oUser2"}]}`
	srv := userOkServer(t, body)
	defer srv.Close()
	c := newUserTestClient(t, srv)
	req := &BatchGetUserInfoRequest{
		UserList: []*UserListItem{
			{Openid: "oUser1", Language: "zh_CN"},
			{Openid: "oUser2", Language: "zh_CN"},
		},
	}
	result, err := c.BatchGetUserInfo(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.UserInfoList) != 2 {
		t.Errorf("expected 2 users, got %d", len(result.UserInfoList))
	}
}

func TestGetTags_Success(t *testing.T) {
	srv := userOkServer(t, `{"tags":[{"id":1,"name":"VIP","count":100}]}`)
	defer srv.Close()
	c := newUserTestClient(t, srv)
	result, err := c.GetTags(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(result.Tags))
	}
}

func TestCreateTag_Success(t *testing.T) {
	srv := userOkServer(t, `{"tag":{"id":100,"name":"VIP","count":0}}`)
	defer srv.Close()
	c := newUserTestClient(t, srv)
	result, err := c.CreateTag(context.Background(), "VIP")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Tag.Name != "VIP" {
		t.Errorf("expected tag name VIP, got %q", result.Tag.Name)
	}
}

func TestCreateTag_NetworkError(t *testing.T) {
	c := newUserTestClient(t, userClosedServer(t))
	_, err := c.CreateTag(context.Background(), "VIP")
	if err == nil {
		t.Error("expected network error")
	}
}

func TestUpdateTag_Success(t *testing.T) {
	srv := userOkServer(t, `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	c := newUserTestClient(t, srv)
	_, err := c.UpdateTag(context.Background(), 100, "Premium")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteTag_Success(t *testing.T) {
	srv := userOkServer(t, `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	c := newUserTestClient(t, srv)
	_, err := c.DeleteTag(context.Background(), 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetTagFans_Success(t *testing.T) {
	body := `{"count":1,"data":{"openid":["oUser1"]},"next_openid":""}`
	srv := userOkServer(t, body)
	defer srv.Close()
	c := newUserTestClient(t, srv)
	result, err := c.GetTagFans(context.Background(), &GetTagFansRequest{TagId: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 1 {
		t.Errorf("expected count=1, got %d", result.Count)
	}
}

func TestBatchTagging_Success(t *testing.T) {
	srv := userOkServer(t, `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	c := newUserTestClient(t, srv)
	_, err := c.BatchTagging(context.Background(), []string{"oUser1", "oUser2"}, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBatchUntagging_Success(t *testing.T) {
	srv := userOkServer(t, `{"errcode":0,"errmsg":"ok"}`)
	defer srv.Close()
	c := newUserTestClient(t, srv)
	_, err := c.BatchUntagging(context.Background(), []string{"oUser1"}, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetTagidList_Success(t *testing.T) {
	srv := userOkServer(t, `{"tagid_list":[100,200]}`)
	defer srv.Close()
	c := newUserTestClient(t, srv)
	result, err := c.GetTagidList(context.Background(), "oUser1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.TagidList) != 2 {
		t.Errorf("expected 2 tagids, got %d", len(result.TagidList))
	}
}
