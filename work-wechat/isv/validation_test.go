package isv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newValidationCorpClient returns a CorpClient whose underlying server
// fails the test if any HTTP request reaches it. This proves that
// validation short-circuits before any network I/O.
func newValidationCorpClient(t *testing.T) *CorpClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("validation should have short-circuited before HTTP: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	return newTestCorpClient(t, srv.URL)
}

// newValidationClient returns an ISV Client whose underlying server
// fails the test if any HTTP request reaches it.
func newValidationClient(t *testing.T) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("validation should have short-circuited before HTTP: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	return newTestISVClient(t, srv.URL)
}

func mustContain(t *testing.T, err error, sub string) {
	t.Helper()
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if !strings.Contains(err.Error(), sub) {
		t.Errorf("error %q should contain %q", err.Error(), sub)
	}
}

// ---------- Message validation ----------

func TestSendText_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.SendText(context.Background(), nil)
	mustContain(t, err, "SendText")
}

func TestSendText_ZeroAgentID(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.SendText(context.Background(), &SendTextReq{
		MessageHeader: MessageHeader{ToUser: "u1", AgentID: 0},
	})
	mustContain(t, err, "AgentID must be > 0")
}

func TestSendImage_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.SendImage(context.Background(), nil)
	mustContain(t, err, "SendImage")
}

func TestSendVoice_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.SendVoice(context.Background(), nil)
	mustContain(t, err, "SendVoice")
}

func TestSendVideo_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.SendVideo(context.Background(), nil)
	mustContain(t, err, "SendVideo")
}

func TestSendFile_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.SendFile(context.Background(), nil)
	mustContain(t, err, "SendFile")
}

func TestSendTextCard_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.SendTextCard(context.Background(), nil)
	mustContain(t, err, "SendTextCard")
}

func TestSendNews_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.SendNews(context.Background(), nil)
	mustContain(t, err, "SendNews")
}

func TestSendMpNews_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.SendMpNews(context.Background(), nil)
	mustContain(t, err, "SendMpNews")
}

func TestSendMarkdown_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.SendMarkdown(context.Background(), nil)
	mustContain(t, err, "SendMarkdown")
}

func TestSendMiniProgramNotice_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.SendMiniProgramNotice(context.Background(), nil)
	mustContain(t, err, "SendMiniProgramNotice")
}

func TestSendTemplateCard_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.SendTemplateCard(context.Background(), nil)
	mustContain(t, err, "SendTemplateCard")
}

// ---------- User validation ----------

func TestCreateUser_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.CreateUser(context.Background(), nil)
	mustContain(t, err, "CreateUser")
}

func TestUpdateUser_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.UpdateUser(context.Background(), nil)
	mustContain(t, err, "UpdateUser")
}

func TestDeleteUser_EmptyID(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.DeleteUser(context.Background(), "")
	mustContain(t, err, "userID")
}

func TestGetUser_EmptyID(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.GetUser(context.Background(), "")
	mustContain(t, err, "userID")
}

// ---------- Agent validation ----------

func TestGetAgent_ZeroID(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.GetAgent(context.Background(), 0)
	mustContain(t, err, "agentID must be > 0")
}

func TestSetAgent_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.SetAgent(context.Background(), nil)
	mustContain(t, err, "SetAgent")
}

func TestSetAgent_ZeroAgentID(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.SetAgent(context.Background(), &SetAgentReq{AgentID: 0})
	mustContain(t, err, "AgentID must be > 0")
}

// ---------- Tag validation ----------

func TestCreateTag_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.CreateTag(context.Background(), nil)
	mustContain(t, err, "CreateTag")
}

func TestUpdateTag_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.UpdateTag(context.Background(), nil)
	mustContain(t, err, "UpdateTag")
}

func TestDeleteTag_ZeroID(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.DeleteTag(context.Background(), 0)
	mustContain(t, err, "tagID must be > 0")
}

func TestGetTagUsers_ZeroID(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.GetTagUsers(context.Background(), 0)
	mustContain(t, err, "tagID must be > 0")
}

func TestAddTagUsers_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.AddTagUsers(context.Background(), nil)
	mustContain(t, err, "AddTagUsers")
}

func TestAddTagUsers_ZeroTagID(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.AddTagUsers(context.Background(), &TagUsersModifyReq{TagID: 0})
	mustContain(t, err, "TagID must be > 0")
}

func TestDelTagUsers_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.DelTagUsers(context.Background(), nil)
	mustContain(t, err, "DelTagUsers")
}

func TestDelTagUsers_ZeroTagID(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.DelTagUsers(context.Background(), &TagUsersModifyReq{TagID: 0})
	mustContain(t, err, "TagID must be > 0")
}

// ---------- Approval validation ----------

func TestGetApprovalTemplate_EmptyID(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.GetApprovalTemplate(context.Background(), "")
	mustContain(t, err, "templateID")
}

func TestApplyEvent_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.ApplyEvent(context.Background(), nil)
	mustContain(t, err, "ApplyEvent")
}

func TestGetApprovalDetail_EmptySpNo(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.GetApprovalDetail(context.Background(), "")
	mustContain(t, err, "spNo")
}

func TestGetApprovalData_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.GetApprovalData(context.Background(), nil)
	mustContain(t, err, "GetApprovalData")
}

// ---------- Menu validation ----------

func TestCreateMenu_ZeroAgentID(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.CreateMenu(context.Background(), 0, &CreateMenuReq{})
	mustContain(t, err, "agentID must be > 0")
}

func TestCreateMenu_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.CreateMenu(context.Background(), 1, nil)
	mustContain(t, err, "CreateMenu")
}

func TestGetMenu_ZeroAgentID(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.GetMenu(context.Background(), 0)
	mustContain(t, err, "agentID must be > 0")
}

func TestDeleteMenu_ZeroAgentID(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.DeleteMenu(context.Background(), 0)
	mustContain(t, err, "agentID must be > 0")
}

// ---------- External Contact validation ----------

func TestGetExternalContact_EmptyID(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.GetExternalContact(context.Background(), "")
	mustContain(t, err, "externalUserID")
}

func TestListExternalContact_EmptyID(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.ListExternalContact(context.Background(), "")
	mustContain(t, err, "userID")
}

func TestBatchGetExternalContactByUser_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.BatchGetExternalContactByUser(context.Background(), nil)
	mustContain(t, err, "BatchGetExternalContactByUser")
}

func TestRemarkExternalContact_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.RemarkExternalContact(context.Background(), nil)
	mustContain(t, err, "RemarkExternalContact")
}

func TestAddCorpTag_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.AddCorpTag(context.Background(), nil)
	mustContain(t, err, "AddCorpTag")
}

func TestEditCorpTag_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.EditCorpTag(context.Background(), nil)
	mustContain(t, err, "EditCorpTag")
}

func TestDelCorpTag_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.DelCorpTag(context.Background(), nil)
	mustContain(t, err, "DelCorpTag")
}

func TestMarkTag_NilReq(t *testing.T) {
	cc := newValidationCorpClient(t)
	err := cc.MarkTag(context.Background(), nil)
	mustContain(t, err, "MarkTag")
}

// ---------- Media validation ----------

func TestUploadMedia_BadType(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.UploadMedia(context.Background(), "thumb", "f.jpg", strings.NewReader("data"))
	mustContain(t, err, "mediaType must be one of")
}

func TestUploadMedia_EmptyFileName(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.UploadMedia(context.Background(), "image", "", strings.NewReader("data"))
	mustContain(t, err, "fileName is required")
}

func TestUploadMedia_NilData(t *testing.T) {
	cc := newValidationCorpClient(t)
	_, err := cc.UploadMedia(context.Background(), "image", "f.jpg", nil)
	mustContain(t, err, "fileData is required")
}

// ---------- OAuth2 validation ----------

func TestGetUserInfo3rd_EmptyCode(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GetUserInfo3rd(context.Background(), "")
	mustContain(t, err, "authCode")
}

func TestGetUserDetail3rd_EmptyTicket(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GetUserDetail3rd(context.Background(), "")
	mustContain(t, err, "userTicket")
}

// ---------- Provider Login validation ----------

func TestGetLoginInfo_EmptyCode(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GetLoginInfo(context.Background(), "")
	mustContain(t, err, "authCode")
}

func TestGetRegistrationInfo_EmptyCode(t *testing.T) {
	c := newValidationClient(t)
	_, err := c.GetRegistrationInfo(context.Background(), "")
	mustContain(t, err, "registerCode")
}

// ---------- Suite PreAuth validation ----------

func TestSetSessionInfo_EmptyPreAuthCode(t *testing.T) {
	c := newValidationClient(t)
	err := c.SetSessionInfo(context.Background(), "", &SessionInfo{})
	mustContain(t, err, "preAuthCode")
}

func TestSetSessionInfo_NilInfo(t *testing.T) {
	c := newValidationClient(t)
	err := c.SetSessionInfo(context.Background(), "preauthcode123", nil)
	mustContain(t, err, "info is required")
}
