package isv

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// msgSendServer returns an httptest server that validates message/send requests.
// checkBody is called with the decoded JSON body for type-specific assertions.
func msgSendServer(t *testing.T, checkBody func(m map[string]interface{})) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/message/send" {
			// Not the target path — might be a token endpoint, ignore.
			return
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if checkBody != nil {
			checkBody(body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"msgid": "MSG001",
		})
	}))
}

func TestSendText_HappyPath(t *testing.T) {
	srv := msgSendServer(t, func(m map[string]interface{}) {
		if m["msgtype"] != "text" {
			t.Errorf("msgtype: %v", m["msgtype"])
		}
		if m["touser"] != "u1|u2" {
			t.Errorf("touser: %v", m["touser"])
		}
		if int(m["agentid"].(float64)) != 1000001 {
			t.Errorf("agentid: %v", m["agentid"])
		}
		text := m["text"].(map[string]interface{})
		if text["content"] != "Hello World" {
			t.Errorf("text.content: %v", text["content"])
		}
	})
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.SendText(context.Background(), &SendTextReq{
		MessageHeader: MessageHeader{
			ToUser:  "u1|u2",
			AgentID: 1000001,
		},
		Text: TextContent{Content: "Hello World"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.MsgID != "MSG001" {
		t.Errorf("msgid: %q", resp.MsgID)
	}
}

func TestSendTextCard_HappyPath(t *testing.T) {
	srv := msgSendServer(t, func(m map[string]interface{}) {
		if m["msgtype"] != "textcard" {
			t.Errorf("msgtype: %v", m["msgtype"])
		}
		tc := m["textcard"].(map[string]interface{})
		if tc["title"] != "Title1" || tc["description"] != "Desc1" || tc["url"] != "https://example.com" {
			t.Errorf("textcard: %+v", tc)
		}
		if tc["btntxt"] != "More" {
			t.Errorf("btntxt: %v", tc["btntxt"])
		}
	})
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.SendTextCard(context.Background(), &SendTextCardReq{
		MessageHeader: MessageHeader{ToUser: "u1", AgentID: 1000001},
		TextCard: TextCardContent{
			Title:       "Title1",
			Description: "Desc1",
			URL:         "https://example.com",
			BtnTxt:      "More",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.MsgID != "MSG001" {
		t.Errorf("msgid: %q", resp.MsgID)
	}
}

func TestSendMarkdown_HappyPath(t *testing.T) {
	srv := msgSendServer(t, func(m map[string]interface{}) {
		if m["msgtype"] != "markdown" {
			t.Errorf("msgtype: %v", m["msgtype"])
		}
		md := m["markdown"].(map[string]interface{})
		if md["content"] != "# Title\n> quote" {
			t.Errorf("markdown.content: %v", md["content"])
		}
	})
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.SendMarkdown(context.Background(), &SendMarkdownReq{
		MessageHeader: MessageHeader{ToParty: "1", AgentID: 1000001},
		Markdown:      MarkdownContent{Content: "# Title\n> quote"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.MsgID != "MSG001" {
		t.Errorf("msgid: %q", resp.MsgID)
	}
}

func TestSendNews_MultiArticle(t *testing.T) {
	srv := msgSendServer(t, func(m map[string]interface{}) {
		if m["msgtype"] != "news" {
			t.Errorf("msgtype: %v", m["msgtype"])
		}
		news := m["news"].(map[string]interface{})
		articles := news["articles"].([]interface{})
		if len(articles) != 2 {
			t.Errorf("articles count: %d", len(articles))
		}
		a0 := articles[0].(map[string]interface{})
		if a0["title"] != "Art1" {
			t.Errorf("articles[0].title: %v", a0["title"])
		}
		a1 := articles[1].(map[string]interface{})
		if a1["title"] != "Art2" || a1["picurl"] != "https://img/2.png" {
			t.Errorf("articles[1]: %+v", a1)
		}
	})
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.SendNews(context.Background(), &SendNewsReq{
		MessageHeader: MessageHeader{ToUser: "u1", AgentID: 1000001},
		News: NewsContent{
			Articles: []NewsArticle{
				{Title: "Art1", URL: "https://example.com/1"},
				{Title: "Art2", PicURL: "https://img/2.png"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.MsgID != "MSG001" {
		t.Errorf("msgid: %q", resp.MsgID)
	}
}

func TestSendTemplateCard_TextNotice(t *testing.T) {
	srv := msgSendServer(t, func(m map[string]interface{}) {
		if m["msgtype"] != "template_card" {
			t.Errorf("msgtype: %v", m["msgtype"])
		}
		tc := m["template_card"].(map[string]interface{})
		if tc["card_type"] != "text_notice" {
			t.Errorf("card_type: %v", tc["card_type"])
		}
		mt := tc["main_title"].(map[string]interface{})
		if mt["title"] != "Urgent" {
			t.Errorf("main_title.title: %v", mt["title"])
		}
		ca := tc["card_action"].(map[string]interface{})
		if ca["url"] != "https://example.com" {
			t.Errorf("card_action.url: %v", ca["url"])
		}
	})
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.SendTemplateCard(context.Background(), &SendTemplateCardReq{
		MessageHeader: MessageHeader{ToUser: "u1", AgentID: 1000001},
		TemplateCard: TemplateCardContent{
			CardType:  "text_notice",
			MainTitle: TCMainTitle{Title: "Urgent", Desc: "Please review"},
			CardAction: TCCardAction{
				Type: 1,
				URL:  "https://example.com",
			},
			EmphasisContent: &TCEmphasisContent{Title: "100", Desc: "items"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.MsgID != "MSG001" {
		t.Errorf("msgid: %q", resp.MsgID)
	}
}

func TestSendRemainingTypes(t *testing.T) {
	srv := msgSendServer(t, func(m map[string]interface{}) {
		// Just verify msgtype is set — all methods share the same wire pattern.
		if m["msgtype"] == nil || m["msgtype"] == "" {
			t.Errorf("msgtype missing")
		}
	})
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	ctx := context.Background()
	hdr := MessageHeader{ToUser: "u1", AgentID: 1000001}

	if _, err := cc.SendImage(ctx, &SendImageReq{MessageHeader: hdr, Image: ImageContent{MediaID: "M1"}}); err != nil {
		t.Errorf("SendImage: %v", err)
	}
	if _, err := cc.SendVoice(ctx, &SendVoiceReq{MessageHeader: hdr, Voice: VoiceContent{MediaID: "M2"}}); err != nil {
		t.Errorf("SendVoice: %v", err)
	}
	if _, err := cc.SendVideo(ctx, &SendVideoReq{MessageHeader: hdr, Video: VideoContent{MediaID: "M3", Title: "T"}}); err != nil {
		t.Errorf("SendVideo: %v", err)
	}
	if _, err := cc.SendFile(ctx, &SendFileReq{MessageHeader: hdr, File: FileContent{MediaID: "M4"}}); err != nil {
		t.Errorf("SendFile: %v", err)
	}
	if _, err := cc.SendMpNews(ctx, &SendMpNewsReq{MessageHeader: hdr, MpNews: MpNewsContent{Articles: []MpNewsArticle{{Title: "T", ThumbMediaID: "TM", Content: "C"}}}}); err != nil {
		t.Errorf("SendMpNews: %v", err)
	}
	if _, err := cc.SendMiniProgramNotice(ctx, &SendMiniProgramNoticeReq{MessageHeader: hdr, MiniProgramNotice: MiniProgramNoticeContent{AppID: "wx123", Title: "T"}}); err != nil {
		t.Errorf("SendMiniProgramNotice: %v", err)
	}
}

func TestSendMessage_WeixinError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errcode": 40014,
			"errmsg":  "invalid access_token",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	_, err := cc.SendText(context.Background(), &SendTextReq{
		MessageHeader: MessageHeader{ToUser: "u1", AgentID: 1000001},
		Text:          TextContent{Content: "test"},
	})
	if err == nil {
		t.Fatal("want error, got nil")
	}
	var we *WeixinError
	if !errors.As(err, &we) || we.ErrCode != 40014 {
		t.Errorf("want *WeixinError errcode=40014, got %v", err)
	}
}
