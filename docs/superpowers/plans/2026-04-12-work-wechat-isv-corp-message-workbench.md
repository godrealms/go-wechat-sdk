# work-wechat ISV Sub-Project 4 — Corp Message + Workbench Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Append 14 public methods to `*CorpClient` covering 11 message types (`SendText` ~ `SendTemplateCard`) and 3 workbench APIs, plus 2 private HTTP helpers.

**Architecture:** All methods hang off `*CorpClient` (not `*Client`), using a new `cc.doPost` helper that injects `access_token` (corp_access_token) into the query string via `cc.parent.doPostRaw`. Each Send method uses an inline wire struct to auto-inject `msgtype`. Workbench methods are thin wrappers around `cc.doPost`. DTO-heavy but logic-light.

**Tech Stack:** Go 1.23+, `encoding/json`, `net/http/httptest`, existing `work-wechat/isv` package. Spec: `docs/superpowers/specs/2026-04-12-work-wechat-isv-corp-message-workbench-design.md`.

---

## File Layout

```
work-wechat/isv/
├── corp.http.go              # NEW — 2 private HTTP helpers (doPost/doGet)
├── corp.http_test.go         # NEW — 2 tests + newTestCorpClient helper
├── struct.message.go         # NEW — MessageHeader + 10 simple msg DTOs + TemplateCard DTOs + SendMessageResp
├── corp.message.go           # NEW — 11 Send* methods
├── corp.message_test.go      # NEW — 6 tests
├── struct.workbench.go       # NEW — workbench DTOs
├── corp.workbench.go         # NEW — 3 workbench methods
└── corp.workbench_test.go    # NEW — 3 tests
```

**Assumptions about existing code (verified):**
- `CorpClient` struct: `{ parent *Client; corpID string }` in `corp.token.go:26-28`.
- `cc.AccessToken(ctx)` returns corp_access_token (lazy + double-check lock).
- `cc.parent.doPostRaw(ctx, path, query, body, out)` — low-level POST, caller controls query.
- `cc.parent.doRequestRaw(ctx, method, path, query, body, out)` — low-level raw HTTP.
- `newTestISVClient(t, baseURL)` in `suite.token_test.go` — seeds suite_ticket, SuiteID=`suite1`.
- `testConfig()` in `client_test.go` — returns Config with SuiteID=`suite1`.
- `store.PutAuthorizer(ctx, suiteID, corpID, &AuthorizerTokens{...})` — pre-seed corp token.
- `store.PutSuiteToken(ctx, suiteID, token, expiresAt)` — pre-seed suite token.
- Enterprise WeChat corp APIs use query key `access_token` (not `corp_access_token`).

---

## Task 1: CorpClient HTTP helpers + tests

**Files:**
- Create: `work-wechat/isv/corp.http.go`
- Create: `work-wechat/isv/corp.http_test.go`

- [ ] **Step 1.1: Write the test helper and failing tests**

Create `work-wechat/isv/corp.http_test.go`:

```go
package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestCorpClient creates a CorpClient with a pre-seeded valid corp_access_token.
// No httptest mock needed for token fetch — the token is directly in the Store.
func newTestCorpClient(t *testing.T, baseURL string) *CorpClient {
	t.Helper()
	c := newTestISVClient(t, baseURL)
	ctx := context.Background()
	_ = c.store.PutSuiteToken(ctx, "suite1", "STOK", time.Now().Add(time.Hour))
	_ = c.store.PutAuthorizer(ctx, "suite1", "wxcorp1", &AuthorizerTokens{
		CorpID:            "wxcorp1",
		PermanentCode:     "PERM",
		CorpAccessToken:   "CTOK",
		CorpTokenExpireAt: time.Now().Add(time.Hour),
	})
	return c.CorpClient("wxcorp1")
}

func TestCorpClient_DoPost_TokenInjection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["key"] != "val" {
			t.Errorf("body: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"result": "ok",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	var resp struct {
		Result string `json:"result"`
	}
	err := cc.doPost(context.Background(), "/test/post", map[string]string{"key": "val"}, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result != "ok" {
		t.Errorf("resp: %+v", resp)
	}
}

func TestCorpClient_DoGet_TokenInjection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: %s", r.Method)
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		if got := r.URL.Query().Get("extra"); got != "123" {
			t.Errorf("extra: %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": "hello",
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	var resp struct {
		Data string `json:"data"`
	}
	extra := map[string][]string{"extra": {"123"}}
	err := cc.doGet(context.Background(), "/test/get", extra, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Data != "hello" {
		t.Errorf("resp: %+v", resp)
	}
}
```

- [ ] **Step 1.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run "TestCorpClient_Do" -v`
Expected: FAIL — `cc.doPost undefined` / `cc.doGet undefined`.

- [ ] **Step 1.3: Implement HTTP helpers**

Create `work-wechat/isv/corp.http.go`:

```go
package isv

import (
	"context"
	"net/http"
	"net/url"
)

// doPost 发送 JSON POST 到 parent.baseURL + path,query 自动注入 corp_access_token。
// 企业微信 corp 接口的 query key 是 access_token(不是 corp_access_token)。
func (cc *CorpClient) doPost(ctx context.Context, path string, body, out interface{}) error {
	tok, err := cc.AccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"access_token": {tok}}
	return cc.parent.doPostRaw(ctx, path, q, body, out)
}

// doGet 发送 GET 到 parent.baseURL + path,query 自动注入 corp_access_token。
func (cc *CorpClient) doGet(ctx context.Context, path string, extra url.Values, out interface{}) error {
	tok, err := cc.AccessToken(ctx)
	if err != nil {
		return err
	}
	q := url.Values{"access_token": {tok}}
	for k, vs := range extra {
		q[k] = vs
	}
	return cc.parent.doRequestRaw(ctx, http.MethodGet, path, q, nil, out)
}
```

- [ ] **Step 1.4: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run "TestCorpClient_Do" -v`
Expected: both PASS.

- [ ] **Step 1.5: Commit**

```bash
git add work-wechat/isv/corp.http.go work-wechat/isv/corp.http_test.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add CorpClient HTTP helpers (doPost/doGet)

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Message DTOs (simple types)

**Files:**
- Create: `work-wechat/isv/struct.message.go`

- [ ] **Step 2.1: Write the DTO file (simple types only, no TemplateCard yet)**

Create `work-wechat/isv/struct.message.go`:

```go
package isv

// ---- message/send common ----

// MessageHeader 是所有消息发送请求共用的头部字段。
type MessageHeader struct {
	ToUser                 string `json:"touser,omitempty"`
	ToParty                string `json:"toparty,omitempty"`
	ToTag                  string `json:"totag,omitempty"`
	AgentID                int    `json:"agentid"`
	Safe                   int    `json:"safe,omitempty"`
	EnableIDTrans          int    `json:"enable_id_trans,omitempty"`
	EnableDuplicateCheck   int    `json:"enable_duplicate_check,omitempty"`
	DuplicateCheckInterval int    `json:"duplicate_check_interval,omitempty"`
}

// SendMessageResp 是 message/send 的统一响应。
type SendMessageResp struct {
	InvalidUser    string `json:"invaliduser"`
	InvalidParty   string `json:"invalidparty"`
	InvalidTag     string `json:"invalidtag"`
	UnlicensedUser string `json:"unlicenseduser"`
	MsgID          string `json:"msgid"`
	ResponseCode   string `json:"response_code"`
}

// ---- text ----

// TextContent 文本消息内容。
type TextContent struct {
	Content string `json:"content"`
}

// SendTextReq 文本消息请求。
type SendTextReq struct {
	MessageHeader
	Text TextContent `json:"text"`
}

// ---- image ----

// ImageContent 图片消息内容。
type ImageContent struct {
	MediaID string `json:"media_id"`
}

// SendImageReq 图片消息请求。
type SendImageReq struct {
	MessageHeader
	Image ImageContent `json:"image"`
}

// ---- voice ----

// VoiceContent 语音消息内容。
type VoiceContent struct {
	MediaID string `json:"media_id"`
}

// SendVoiceReq 语音消息请求。
type SendVoiceReq struct {
	MessageHeader
	Voice VoiceContent `json:"voice"`
}

// ---- video ----

// VideoContent 视频消息内容。
type VideoContent struct {
	MediaID     string `json:"media_id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// SendVideoReq 视频消息请求。
type SendVideoReq struct {
	MessageHeader
	Video VideoContent `json:"video"`
}

// ---- file ----

// FileContent 文件消息内容。
type FileContent struct {
	MediaID string `json:"media_id"`
}

// SendFileReq 文件消息请求。
type SendFileReq struct {
	MessageHeader
	File FileContent `json:"file"`
}

// ---- textcard ----

// TextCardContent 文本卡片消息内容。
type TextCardContent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	BtnTxt      string `json:"btntxt,omitempty"`
}

// SendTextCardReq 文本卡片消息请求。
type SendTextCardReq struct {
	MessageHeader
	TextCard TextCardContent `json:"textcard"`
}

// ---- news ----

// NewsArticle 图文消息条目。
type NewsArticle struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	PicURL      string `json:"picurl,omitempty"`
	AppID       string `json:"appid,omitempty"`
	PagePath    string `json:"pagepath,omitempty"`
}

// NewsContent 图文消息内容。
type NewsContent struct {
	Articles []NewsArticle `json:"articles"`
}

// SendNewsReq 图文消息请求。
type SendNewsReq struct {
	MessageHeader
	News NewsContent `json:"news"`
}

// ---- mpnews ----

// MpNewsArticle 图文消息(mpnews)条目。
type MpNewsArticle struct {
	Title            string `json:"title"`
	ThumbMediaID     string `json:"thumb_media_id"`
	Author           string `json:"author,omitempty"`
	ContentSourceURL string `json:"content_source_url,omitempty"`
	Content          string `json:"content"`
	Digest           string `json:"digest,omitempty"`
}

// MpNewsContent 图文消息(mpnews)内容。
type MpNewsContent struct {
	Articles []MpNewsArticle `json:"articles"`
}

// SendMpNewsReq 图文消息(mpnews)请求。
type SendMpNewsReq struct {
	MessageHeader
	MpNews MpNewsContent `json:"mpnews"`
}

// ---- markdown ----

// MarkdownContent Markdown 消息内容。
type MarkdownContent struct {
	Content string `json:"content"`
}

// SendMarkdownReq Markdown 消息请求。
type SendMarkdownReq struct {
	MessageHeader
	Markdown MarkdownContent `json:"markdown"`
}

// ---- miniprogram_notice ----

// MiniProgramContentItem 小程序通知的 content_item 条目。
type MiniProgramContentItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// MiniProgramNoticeContent 小程序通知消息内容。
type MiniProgramNoticeContent struct {
	AppID             string                   `json:"appid"`
	Page              string                   `json:"page,omitempty"`
	Title             string                   `json:"title"`
	Description       string                   `json:"description,omitempty"`
	EmphasisFirstItem bool                     `json:"emphasis_first_item,omitempty"`
	ContentItem       []MiniProgramContentItem `json:"content_item,omitempty"`
}

// SendMiniProgramNoticeReq 小程序通知消息请求。
type SendMiniProgramNoticeReq struct {
	MessageHeader
	MiniProgramNotice MiniProgramNoticeContent `json:"miniprogram_notice"`
}
```

- [ ] **Step 2.2: Verify compilation**

Run: `go build ./work-wechat/isv/...`
Expected: clean build (no output).

- [ ] **Step 2.3: Commit**

```bash
git add work-wechat/isv/struct.message.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add message DTOs (10 simple types)

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: TemplateCard DTOs

**Files:**
- Modify: `work-wechat/isv/struct.message.go` — append TemplateCard types at the bottom

- [ ] **Step 3.1: Append TemplateCard types**

Append to the bottom of `work-wechat/isv/struct.message.go`:

```go
// ---- template_card ----

// TCSource 卡片来源样式。
type TCSource struct {
	IconURL   string `json:"icon_url,omitempty"`
	Desc      string `json:"desc,omitempty"`
	DescColor int    `json:"desc_color,omitempty"` // 0 灰 / 1 黑 / 2 红 / 3 绿
}

// TCActionMenuItem 右上角菜单项。
type TCActionMenuItem struct {
	Text string `json:"text"`
	Key  string `json:"key"`
}

// TCActionMenu 右上角菜单。
type TCActionMenu struct {
	Desc       string             `json:"desc,omitempty"`
	ActionList []TCActionMenuItem `json:"action_list"`
}

// TCMainTitle 一级标题。
type TCMainTitle struct {
	Title string `json:"title,omitempty"`
	Desc  string `json:"desc,omitempty"`
}

// TCEmphasisContent 关键数据。
type TCEmphasisContent struct {
	Title string `json:"title,omitempty"`
	Desc  string `json:"desc,omitempty"`
}

// TCQuoteArea 引用。
type TCQuoteArea struct {
	Type      int    `json:"type,omitempty"` // 0 文本 / 1 链接
	URL       string `json:"url,omitempty"`
	Title     string `json:"title,omitempty"`
	QuoteText string `json:"quote_text,omitempty"`
}

// TCHorizontalContent 二级标题 + 文本列表。
type TCHorizontalContent struct {
	KeyName string `json:"keyname"`
	Value   string `json:"value,omitempty"`
	Type    int    `json:"type,omitempty"` // 0 文本 / 1 链接 / 2 附件 / 3 @人
	URL     string `json:"url,omitempty"`
	MediaID string `json:"media_id,omitempty"`
	UserID  string `json:"userid,omitempty"`
}

// TCJumpItem 跳转列表项。
type TCJumpItem struct {
	Type     int    `json:"type,omitempty"` // 0 链接 / 1 小程序
	Title    string `json:"title"`
	URL      string `json:"url,omitempty"`
	AppID    string `json:"appid,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// TCCardAction 整体卡片跳转。
type TCCardAction struct {
	Type     int    `json:"type"` // 1 链接 / 2 小程序
	URL      string `json:"url,omitempty"`
	AppID    string `json:"appid,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// TCOption 选项(投票/多选共用)。
type TCOption struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsChecked bool   `json:"is_checked,omitempty"`
}

// TCButton 按钮。
type TCButton struct {
	Text  string `json:"text"`
	Style int    `json:"style,omitempty"` // 1 常规 / 2 强调
	Key   string `json:"key"`
}

// TCButtonSelection 下拉按钮。
type TCButtonSelection struct {
	QuestionKey string     `json:"question_key"`
	Title       string     `json:"title,omitempty"`
	OptionList  []TCOption `json:"option_list"`
	SelectedID  string     `json:"selected_id,omitempty"`
}

// TCSelectItem 多选列表单项。
type TCSelectItem struct {
	QuestionKey string     `json:"question_key"`
	Title       string     `json:"title,omitempty"`
	SelectedID  string     `json:"selected_id,omitempty"`
	OptionList  []TCOption `json:"option_list"`
}

// TCCheckbox 多选框。
type TCCheckbox struct {
	QuestionKey string     `json:"question_key"`
	OptionList  []TCOption `json:"option_list"`
	Mode        int        `json:"mode,omitempty"` // 0 多选 / 1 单选
}

// TCSubmitButton 提交按钮。
type TCSubmitButton struct {
	Text string `json:"text"`
	Key  string `json:"key"`
}

// TCCardImage 卡片图片(news_notice 子类型)。
type TCCardImage struct {
	URL         string  `json:"url"`
	AspectRatio float64 `json:"aspect_ratio,omitempty"`
}

// TCImageTextArea 左图右文(news_notice 子类型)。
type TCImageTextArea struct {
	Type     int    `json:"type,omitempty"`
	URL      string `json:"url,omitempty"`
	Title    string `json:"title,omitempty"`
	Desc     string `json:"desc,omitempty"`
	ImageURL string `json:"image_url"`
}

// TCVerticalContent 竖向内容。
type TCVerticalContent struct {
	Title string `json:"title"`
	Desc  string `json:"desc,omitempty"`
}

// TemplateCardContent 是 template_card 消息的 content 结构。
// card_type 决定哪些字段有效:
//   - text_notice / news_notice: 基本字段
//   - button_interaction: + ButtonSelection + ButtonList
//   - vote_interaction: + ButtonSelection + ButtonList
//   - multiple_interaction: + SelectList + Checkbox + SubmitButton
type TemplateCardContent struct {
	CardType              string                `json:"card_type"`
	Source                *TCSource             `json:"source,omitempty"`
	ActionMenu            *TCActionMenu         `json:"action_menu,omitempty"`
	TaskID                string                `json:"task_id,omitempty"`
	MainTitle             TCMainTitle           `json:"main_title"`
	EmphasisContent       *TCEmphasisContent    `json:"emphasis_content,omitempty"`
	QuoteArea             *TCQuoteArea          `json:"quote_area,omitempty"`
	SubTitleText          string                `json:"sub_title_text,omitempty"`
	HorizontalContentList []TCHorizontalContent `json:"horizontal_content_list,omitempty"`
	JumpList              []TCJumpItem          `json:"jump_list,omitempty"`
	CardAction            TCCardAction          `json:"card_action"`
	ButtonSelection       *TCButtonSelection    `json:"button_selection,omitempty"`
	ButtonList            []TCButton            `json:"button_list,omitempty"`
	SelectList            []TCSelectItem        `json:"select_list,omitempty"`
	Checkbox              *TCCheckbox           `json:"checkbox,omitempty"`
	SubmitButton          *TCSubmitButton       `json:"submit_button,omitempty"`
	CardImage             *TCCardImage          `json:"card_image,omitempty"`
	ImageTextArea         *TCImageTextArea      `json:"image_text_area,omitempty"`
	VerticalContentList   []TCVerticalContent   `json:"vertical_content_list,omitempty"`
}

// SendTemplateCardReq 模板卡片消息请求。
type SendTemplateCardReq struct {
	MessageHeader
	TemplateCard TemplateCardContent `json:"template_card"`
}
```

- [ ] **Step 3.2: Verify compilation**

Run: `go build ./work-wechat/isv/...`
Expected: clean build.

- [ ] **Step 3.3: Commit**

```bash
git add work-wechat/isv/struct.message.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add TemplateCard DTOs (15 nested types)

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: Send methods + message tests

**Files:**
- Create: `work-wechat/isv/corp.message.go`
- Create: `work-wechat/isv/corp.message_test.go`

- [ ] **Step 4.1: Write the failing tests**

Create `work-wechat/isv/corp.message_test.go`:

```go
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
```

- [ ] **Step 4.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run "TestSend" -v`
Expected: FAIL — `cc.SendText undefined` etc.

- [ ] **Step 4.3: Implement all 11 Send methods**

Create `work-wechat/isv/corp.message.go`:

```go
package isv

import "context"

// SendText 发送文本消息。
func (cc *CorpClient) SendText(ctx context.Context, req *SendTextReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendTextReq
	}
	wire.MsgType = "text"
	wire.SendTextReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendImage 发送图片消息。
func (cc *CorpClient) SendImage(ctx context.Context, req *SendImageReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendImageReq
	}
	wire.MsgType = "image"
	wire.SendImageReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendVoice 发送语音消息。
func (cc *CorpClient) SendVoice(ctx context.Context, req *SendVoiceReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendVoiceReq
	}
	wire.MsgType = "voice"
	wire.SendVoiceReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendVideo 发送视频消息。
func (cc *CorpClient) SendVideo(ctx context.Context, req *SendVideoReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendVideoReq
	}
	wire.MsgType = "video"
	wire.SendVideoReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendFile 发送文件消息。
func (cc *CorpClient) SendFile(ctx context.Context, req *SendFileReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendFileReq
	}
	wire.MsgType = "file"
	wire.SendFileReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendTextCard 发送文本卡片消息。
func (cc *CorpClient) SendTextCard(ctx context.Context, req *SendTextCardReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendTextCardReq
	}
	wire.MsgType = "textcard"
	wire.SendTextCardReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendNews 发送图文消息。
func (cc *CorpClient) SendNews(ctx context.Context, req *SendNewsReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendNewsReq
	}
	wire.MsgType = "news"
	wire.SendNewsReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendMpNews 发送图文消息(mpnews)。
func (cc *CorpClient) SendMpNews(ctx context.Context, req *SendMpNewsReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendMpNewsReq
	}
	wire.MsgType = "mpnews"
	wire.SendMpNewsReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendMarkdown 发送 Markdown 消息。
func (cc *CorpClient) SendMarkdown(ctx context.Context, req *SendMarkdownReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendMarkdownReq
	}
	wire.MsgType = "markdown"
	wire.SendMarkdownReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendMiniProgramNotice 发送小程序通知消息。
func (cc *CorpClient) SendMiniProgramNotice(ctx context.Context, req *SendMiniProgramNoticeReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendMiniProgramNoticeReq
	}
	wire.MsgType = "miniprogram_notice"
	wire.SendMiniProgramNoticeReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SendTemplateCard 发送模板卡片消息。
func (cc *CorpClient) SendTemplateCard(ctx context.Context, req *SendTemplateCardReq) (*SendMessageResp, error) {
	var wire struct {
		MsgType string `json:"msgtype"`
		SendTemplateCardReq
	}
	wire.MsgType = "template_card"
	wire.SendTemplateCardReq = *req
	var resp SendMessageResp
	if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 4.4: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run "TestSend" -v`
Expected: all 6 PASS.

- [ ] **Step 4.5: Commit**

```bash
git add work-wechat/isv/corp.message.go work-wechat/isv/corp.message_test.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add 11 Send* methods for message/send

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Workbench DTOs + methods + tests

**Files:**
- Create: `work-wechat/isv/struct.workbench.go`
- Create: `work-wechat/isv/corp.workbench.go`
- Create: `work-wechat/isv/corp.workbench_test.go`

- [ ] **Step 5.1: Write the failing tests**

Create `work-wechat/isv/corp.workbench_test.go`:

```go
package isv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetWorkbenchTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/agent/set_workbench_template" {
			return
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if int(body["agentid"].(float64)) != 1000001 {
			t.Errorf("agentid: %v", body["agentid"])
		}
		if body["type"] != "key_data" {
			t.Errorf("type: %v", body["type"])
		}
		kd := body["key_data"].(map[string]interface{})
		items := kd["items"].([]interface{})
		if len(items) != 2 {
			t.Errorf("items count: %d", len(items))
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 0, "errmsg": "ok"})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.SetWorkbenchTemplate(context.Background(), &WorkbenchTemplateReq{
		AgentID: 1000001,
		Type:    "key_data",
		KeyData: &WBKeyData{
			Items: []WBKeyDataItem{
				{Key: "待审批", Data: "2", JumpURL: "https://example.com/1"},
				{Key: "已通过", Data: "100"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetWorkbenchTemplate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/agent/get_workbench_template" {
			return
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if int(body["agentid"].(float64)) != 1000001 {
			t.Errorf("agentid: %v", body["agentid"])
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"agentid": 1000001,
			"type":    "key_data",
			"key_data": map[string]interface{}{
				"items": []map[string]interface{}{
					{"key": "待审批", "data": "2", "jump_url": "https://example.com/1"},
				},
			},
			"replace_user_data": true,
		})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	resp, err := cc.GetWorkbenchTemplate(context.Background(), 1000001)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Type != "key_data" || resp.AgentID != 1000001 {
		t.Errorf("resp: %+v", resp)
	}
	if resp.KeyData == nil || len(resp.KeyData.Items) != 1 || resp.KeyData.Items[0].Key != "待审批" {
		t.Errorf("key_data: %+v", resp.KeyData)
	}
	if !resp.ReplaceUserData {
		t.Errorf("replace_user_data: %v", resp.ReplaceUserData)
	}
}

func TestSetWorkbenchData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/agent/set_workbench_data" {
			return
		}
		if got := r.URL.Query().Get("access_token"); got != "CTOK" {
			t.Errorf("access_token: %q", got)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["userid"] != "u1" {
			t.Errorf("userid: %v", body["userid"])
		}
		if body["type"] != "key_data" {
			t.Errorf("type: %v", body["type"])
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 0, "errmsg": "ok"})
	}))
	defer srv.Close()

	cc := newTestCorpClient(t, srv.URL)
	err := cc.SetWorkbenchData(context.Background(), &WorkbenchDataReq{
		AgentID: 1000001,
		UserID:  "u1",
		Type:    "key_data",
		KeyData: &WBKeyData{
			Items: []WBKeyDataItem{
				{Key: "待审批", Data: "5"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 5.2: Run tests — expect fail**

Run: `go test ./work-wechat/isv/... -run "TestSetWorkbench|TestGetWorkbench" -v`
Expected: FAIL — `WorkbenchTemplateReq` / `cc.SetWorkbenchTemplate` undefined.

- [ ] **Step 5.3: Write workbench DTOs**

Create `work-wechat/isv/struct.workbench.go`:

```go
package isv

// ---- workbench ----

// WBKeyDataItem 关键数据条目。
type WBKeyDataItem struct {
	Key      string `json:"key"`
	Data     string `json:"data"`
	JumpURL  string `json:"jump_url,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// WBKeyData 关键数据型工作台。
type WBKeyData struct {
	Items []WBKeyDataItem `json:"items"`
}

// WBImage 图片型工作台。
type WBImage struct {
	URL      string `json:"url"`
	JumpURL  string `json:"jump_url,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// WBListItem 列表条目。
type WBListItem struct {
	Title    string `json:"title"`
	JumpURL  string `json:"jump_url,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// WBList 列表型工作台。
type WBList struct {
	Items []WBListItem `json:"items"`
}

// WBWebview 网页型工作台。
type WBWebview struct {
	URL      string `json:"url"`
	JumpURL  string `json:"jump_url,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// WorkbenchTemplateReq 是 agent/set_workbench_template 的请求体。
type WorkbenchTemplateReq struct {
	AgentID         int        `json:"agentid"`
	Type            string     `json:"type"` // key_data / image / list / webview / normal
	KeyData         *WBKeyData `json:"key_data,omitempty"`
	Image           *WBImage   `json:"image,omitempty"`
	List            *WBList    `json:"list,omitempty"`
	Webview         *WBWebview `json:"webview,omitempty"`
	ReplaceUserData bool       `json:"replace_user_data,omitempty"`
}

// WorkbenchTemplateResp 是 agent/get_workbench_template 的响应。
type WorkbenchTemplateResp struct {
	AgentID         int        `json:"agentid"`
	Type            string     `json:"type"`
	KeyData         *WBKeyData `json:"key_data,omitempty"`
	Image           *WBImage   `json:"image,omitempty"`
	List            *WBList    `json:"list,omitempty"`
	Webview         *WBWebview `json:"webview,omitempty"`
	ReplaceUserData bool       `json:"replace_user_data"`
}

// WorkbenchDataReq 是 agent/set_workbench_data 的请求体。
type WorkbenchDataReq struct {
	AgentID int        `json:"agentid"`
	UserID  string     `json:"userid"`
	Type    string     `json:"type"`
	KeyData *WBKeyData `json:"key_data,omitempty"`
	Image   *WBImage   `json:"image,omitempty"`
	List    *WBList    `json:"list,omitempty"`
	Webview *WBWebview `json:"webview,omitempty"`
}
```

- [ ] **Step 5.4: Write workbench methods**

Create `work-wechat/isv/corp.workbench.go`:

```go
package isv

import "context"

// SetWorkbenchTemplate 设置应用在工作台的展示模板。
func (cc *CorpClient) SetWorkbenchTemplate(ctx context.Context, req *WorkbenchTemplateReq) error {
	return cc.doPost(ctx, "/cgi-bin/agent/set_workbench_template", req, nil)
}

// GetWorkbenchTemplate 获取应用在工作台的展示模板。
// 注意:企业微信此接口也是 POST(不是 GET)。
func (cc *CorpClient) GetWorkbenchTemplate(ctx context.Context, agentID int) (*WorkbenchTemplateResp, error) {
	body := map[string]int{"agentid": agentID}
	var resp WorkbenchTemplateResp
	if err := cc.doPost(ctx, "/cgi-bin/agent/get_workbench_template", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetWorkbenchData 设置指定用户在工作台上的个性化展示数据。
func (cc *CorpClient) SetWorkbenchData(ctx context.Context, req *WorkbenchDataReq) error {
	return cc.doPost(ctx, "/cgi-bin/agent/set_workbench_data", req, nil)
}
```

- [ ] **Step 5.5: Run tests — expect pass**

Run: `go test ./work-wechat/isv/... -run "TestSetWorkbench|TestGetWorkbench" -v`
Expected: all 3 PASS.

- [ ] **Step 5.6: Full isv suite + race + coverage + vet**

Run (serially):

```bash
go vet ./work-wechat/isv/...
go test -race ./work-wechat/isv/... -count=1
go test -cover ./work-wechat/isv/...
```

Expected:
- vet: clean
- race: all green
- coverage: ≥85%

- [ ] **Step 5.7: Full repo regression**

Run: `go test ./... -count=1`
Expected: every package green.

- [ ] **Step 5.8: Commit**

```bash
git add work-wechat/isv/struct.workbench.go work-wechat/isv/corp.workbench.go work-wechat/isv/corp.workbench_test.go
git commit -m "$(cat <<'EOF'
feat(work-wechat/isv): add workbench template + data methods

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 5.9: Tag final state and report**

Run the following and report the output:

```bash
git log --oneline -8
go test -cover ./work-wechat/isv/...
```

Expected: 5 new `feat(work-wechat/isv)` commits on top of the sub-project 4 spec commit (`48cf0dd`), plus the coverage line.

---

## Self-Review

### Spec coverage

| Spec requirement | Task |
|---|---|
| §2.1 CorpClient `doPost` / `doGet` | Task 1 |
| §2.2 Send method wire struct pattern | Task 4 |
| §3.1 `MessageHeader` | Task 2 |
| §3.2 `SendMessageResp` | Task 2 |
| §3.3 10 simple message types | Task 2 |
| §3.4 TemplateCard DTOs (15+ nested types) | Task 3 |
| §3.5 Workbench DTOs (WBKeyData, WBImage, WBList, WBWebview, req/resp) | Task 5 |
| §1.1 #1-#11 Send methods | Task 4 |
| §1.1 #12 `SetWorkbenchTemplate` | Task 5 |
| §1.1 #13 `GetWorkbenchTemplate` | Task 5 |
| §1.1 #14 `SetWorkbenchData` | Task 5 |
| §5.2 11 test cases | Task 1 (2) + Task 4 (6) + Task 5 (3) = **11** ✔ |
| §6 no new sentinel errors | ✔ |
| §8 ~6 commits | 5 task commits ✔ (merged DTO commits 2+3 + method commits save one vs. spec's 6) |
| §9 corp query key = `access_token` | ✔ Task 1 code uses `"access_token"` |

### Placeholder scan
- No TBD / TODO / "implement later".
- Every code step contains complete Go code.
- Every test step contains complete Go test code.
- Every command step has the exact command + expected output.

### Type / name consistency
- `MessageHeader` (Task 2) embedded in all Send*Req types (Task 2, 3) — field names match test assertions in Task 4. ✔
- `SendMessageResp.MsgID` (`json:"msgid"`): asserted as `resp.MsgID` in Task 4 tests. ✔
- `CorpClient.doPost` (Task 1): used by all Send methods (Task 4) and workbench methods (Task 5). ✔
- `newTestCorpClient` (Task 1): used in Tasks 4 and 5. Seeds token `CTOK`, tests assert `access_token=CTOK`. ✔
- `WBKeyData.Items` / `WBKeyDataItem.Key` / `.Data` / `.JumpURL`: defined in Task 5, used in Task 5 tests. ✔
- `WorkbenchTemplateResp.ReplaceUserData` (`json:"replace_user_data"`): asserted in `TestGetWorkbenchTemplate`. ✔
- `TemplateCardContent.MainTitle` / `.CardAction` / `.EmphasisContent`: defined in Task 3, used in Task 4's `TestSendTemplateCard_TextNotice`. ✔
- `WeixinError`: pre-existing from sub-project 1, used in Task 4's `TestSendMessage_WeixinError`. ✔

**Plan is self-consistent and fully covers the spec.**
