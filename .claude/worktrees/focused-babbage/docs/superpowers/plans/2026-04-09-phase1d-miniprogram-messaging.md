# Phase 1D: mini-program Subscribe Messaging, Customer Service, Security

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add subscribe message push, customer service (kefu) message routing, and content security check APIs to the mini-program package.

**Architecture:** All files follow the pattern from Plan C — use c.Ctx, c.GetAccessToken(), c.TokenQuery(). Subscribe message send uses POST with access_token in URL. Security check APIs (imgSecCheck, msgSecCheck) POST binary/text content.

**Tech Stack:** Go 1.23.1, standard library only. Depends on Plan C being complete.

**PREREQUISITE:** `2026-04-09-phase1c-miniprogram-client-auth.md` must be completed first.

---

## Task 1: Subscribe Message Structs and API

**Files:**
- `mini-program/struct.subscribe.go`
- `mini-program/api.subscribe.go`

- [ ] **Step 1:** Create `mini-program/struct.subscribe.go` with the following complete content:

```go
package mini_program

import "github.com/godrealms/go-wechat-sdk/core"

// SubscribeMessageValue holds the template keyword value
type SubscribeMessageValue struct {
	Value string `json:"value"`
}

// SendSubscribeMessageRequest is the request body for SendSubscribeMessage
type SendSubscribeMessageRequest struct {
	ToUser           string                            `json:"touser"`
	TemplateId       string                            `json:"template_id"`
	Page             string                            `json:"page,omitempty"`
	MiniProgramState string                            `json:"miniprogram_state,omitempty"` // developer/trial/formal
	Lang             string                            `json:"lang,omitempty"`              // zh_CN/en_US/zh_HK/zh_TW
	Data             map[string]*SubscribeMessageValue `json:"data"`
}

// SendSubscribeMessageResult is the result of SendSubscribeMessage
type SendSubscribeMessageResult struct {
	core.Resp
}
```

- [ ] **Step 2:** Create `mini-program/api.subscribe.go` with the following complete content:

```go
package mini_program

import "fmt"

// SendSubscribeMessage 发送订阅消息
// POST /cgi-bin/message/subscribe/send (access_token in URL)
func (c *Client) SendSubscribeMessage(req *SendSubscribeMessageRequest) error {
	path := fmt.Sprintf("/cgi-bin/message/subscribe/send?access_token=%s", c.GetAccessToken())
	result := &SendSubscribeMessageResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return err
	}
	return result.GetError()
}
```

- [ ] **Step 3:** Commit:

```bash
git add mini-program/struct.subscribe.go mini-program/api.subscribe.go
git commit -m "feat(mini-program): add subscribe message API"
```

---

## Task 2: Customer Service Structs and API

**Files:**
- `mini-program/struct.customer.go`
- `mini-program/api.customer.go`

- [ ] **Step 1:** Create `mini-program/struct.customer.go` with the following complete content:

```go
package mini_program

import "github.com/godrealms/go-wechat-sdk/core"

// KfAccount represents a customer service account
type KfAccount struct {
	KfAccount    string `json:"kf_account"`
	NickName     string `json:"nickname"`
	KfHeadImgUrl string `json:"kf_headimgurl"`
}

// KfAccountListResult is the result of GetKfAccountList
type KfAccountListResult struct {
	core.Resp
	KfList []*KfAccount `json:"kf_list"`
}

// CustomerMsgText holds text message content
type CustomerMsgText struct {
	Content string `json:"content"`
}

// CustomerMsgImage holds image message content
type CustomerMsgImage struct {
	MediaId string `json:"media_id"`
}

// CustomerMsgLink holds link message content
type CustomerMsgLink struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	ThumbUrl    string `json:"thumb_url"`
}

// CustomerMsgMiniProgram holds mini-program card message content
type CustomerMsgMiniProgram struct {
	Title        string `json:"title"`
	Pagepath     string `json:"pagepath"`
	ThumbMediaId string `json:"thumb_media_id"`
}

// SendCustomerMessageRequest is the request for SendCustomerMessage
type SendCustomerMessageRequest struct {
	ToUser      string                  `json:"touser"`
	MsgType     string                  `json:"msgtype"` // text/image/link/miniprogrampage
	Text        *CustomerMsgText        `json:"text,omitempty"`
	Image       *CustomerMsgImage       `json:"image,omitempty"`
	Link        *CustomerMsgLink        `json:"link,omitempty"`
	MiniProgram *CustomerMsgMiniProgram `json:"miniprogrampage,omitempty"`
}

// SendCustomerMessageResult is the result of SendCustomerMessage
type SendCustomerMessageResult struct {
	core.Resp
}

// TypingStatus represents typing status command
type TypingStatus string

const (
	TypingStatusTyping       TypingStatus = "Typing"
	TypingStatusCancelTyping TypingStatus = "CancelTyping"
)

// SetTypingRequest is the request for SetTyping
type SetTypingRequest struct {
	ToUser  string       `json:"touser"`
	Command TypingStatus `json:"command"`
}
```

- [ ] **Step 2:** Create `mini-program/api.customer.go` with the following complete content:

```go
package mini_program

import (
	"fmt"
	"net/url"
)

// SendCustomerMessage 发送客服消息给用户
// POST /cgi-bin/message/custom/send (access_token in URL)
func (c *Client) SendCustomerMessage(req *SendCustomerMessageRequest) error {
	path := fmt.Sprintf("/cgi-bin/message/custom/send?access_token=%s", c.GetAccessToken())
	result := &SendCustomerMessageResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return err
	}
	return result.GetError()
}

// SetTyping 下发客服当前输入状态给用户
// POST /cgi-bin/message/custom/typing (access_token in URL)
func (c *Client) SetTyping(toUser string, command TypingStatus) error {
	path := fmt.Sprintf("/cgi-bin/message/custom/typing?access_token=%s", c.GetAccessToken())
	req := &SetTypingRequest{ToUser: toUser, Command: command}
	result := &SendCustomerMessageResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return err
	}
	return result.GetError()
}

// GetKfAccountList 获取客服账号列表
// GET /cgi-bin/customservice/getkfaccountlist
func (c *Client) GetKfAccountList() (*KfAccountListResult, error) {
	query := c.TokenQuery(url.Values{})
	result := &KfAccountListResult{}
	err := c.Https.Get(c.Ctx, "/cgi-bin/customservice/getkfaccountlist", query, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}
```

- [ ] **Step 3:** Commit:

```bash
git add mini-program/struct.customer.go mini-program/api.customer.go
git commit -m "feat(mini-program): add customer service APIs"
```

---

## Task 3: Security Check Structs and API

**Files:**
- `mini-program/struct.security.go`
- `mini-program/api.security.go`

- [ ] **Step 1:** Create `mini-program/struct.security.go` with the following complete content:

```go
package mini_program

import "github.com/godrealms/go-wechat-sdk/core"

// MediaCheckResult is the result of ImgSecCheck and MediaCheckAsync
type MediaCheckResult struct {
	core.Resp
	TraceId string `json:"trace_id"`
}

// MsgSecCheckDetail contains detail for a text segment
type MsgSecCheckDetail struct {
	Strategy string `json:"strategy"`
	ErrCode  int    `json:"errcode"`
	Suggest  string `json:"suggest"` // pass/review/risky
	Label    int    `json:"label"`
	Level    int    `json:"level"`
	Prob     int    `json:"prob"`
	KeyWord  string `json:"keyword"`
}

// MsgSecCheckResult is the result of MsgSecCheck
type MsgSecCheckResult struct {
	core.Resp
	TraceId string               `json:"trace_id"`
	Result  *MsgSecCheckSummary  `json:"result"`
	Detail  []*MsgSecCheckDetail `json:"detail"`
}

// MsgSecCheckSummary is the overall result summary
type MsgSecCheckSummary struct {
	Suggest string `json:"suggest"` // pass/review/risky
	Label   int    `json:"label"`
}

// MsgSecCheckRequest is the request for MsgSecCheck
type MsgSecCheckRequest struct {
	Content   string `json:"content"`
	Version   int    `json:"version"`           // 1 or 2 (v2 provides more detail)
	Scene     int    `json:"scene"`             // 1=资料 2=评论 3=论坛 4=社交日志
	Openid    string `json:"openid"`
	Title     string `json:"title,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// MediaCheckAsyncRequest is the request for MediaCheckAsync
type MediaCheckAsyncRequest struct {
	MediaUrl  string `json:"media_url"`
	MediaType int    `json:"media_type"` // 1=音频 2=图片
	Version   int    `json:"version"`
	Scene     int    `json:"scene"`
	Openid    string `json:"openid"`
}

// MediaCheckAsyncResult is the result of MediaCheckAsync
type MediaCheckAsyncResult struct {
	core.Resp
	TraceId string `json:"trace_id"`
}
```

- [ ] **Step 2:** Create `mini-program/api.security.go` with the following complete content:

```go
package mini_program

import "fmt"

// ImgSecCheck 校验一张图片是否含有违法违规内容 (同步检测，图片<1MB)
// POST /wxa/img_sec_check (access_token in URL, multipart form — posts raw image bytes)
// Returns (traceId, error). The imageData should be raw PNG/JPG bytes.
func (c *Client) ImgSecCheck(imageData []byte) (*MediaCheckResult, error) {
	path := fmt.Sprintf("/wxa/img_sec_check?access_token=%s", c.GetAccessToken())
	result := &MediaCheckResult{}
	err := c.Https.PostForm(c.Ctx, path, "media", "image.jpg", imageData, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// MsgSecCheck 检查一段文本是否含有违法违规内容 (v2 同步检测)
// POST /wxa/msg_sec_check (access_token in URL)
func (c *Client) MsgSecCheck(req *MsgSecCheckRequest) (*MsgSecCheckResult, error) {
	path := fmt.Sprintf("/wxa/msg_sec_check?access_token=%s", c.GetAccessToken())
	result := &MsgSecCheckResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// MediaCheckAsync 异步校验图片/音频是否含有违法违规内容
// POST /wxa/media_check_async (access_token in URL)
func (c *Client) MediaCheckAsync(req *MediaCheckAsyncRequest) (*MediaCheckAsyncResult, error) {
	path := fmt.Sprintf("/wxa/media_check_async?access_token=%s", c.GetAccessToken())
	result := &MediaCheckAsyncResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}
```

Note: `ImgSecCheck` uses `c.Https.PostForm` which already exists in utils/http.go with signature:
`PostForm(ctx context.Context, path string, fieldName string, fileName string, fileData []byte, result interface{}) error`

- [ ] **Step 3:** Commit:

```bash
git add mini-program/struct.security.go mini-program/api.security.go
git commit -m "feat(mini-program): add security check APIs"
```

---

## Task 4: Tests

**Files:**
- `mini-program/api_subscribe_test.go`

- [ ] **Step 1:** Create `mini-program/api_subscribe_test.go` with the following complete content:

```go
package mini_program

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/godrealms/go-wechat-sdk/core"
)

func TestSendSubscribeMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/message/subscribe/send" {
			json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 0, "errmsg": "ok"})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "tok", "expires_in": 7200})
		}
	}))
	defer srv.Close()

	cfg := &Config{BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"}}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "POST")
	c := &Client{BaseClient: base}

	req := &SendSubscribeMessageRequest{
		ToUser:     "openid123",
		TemplateId: "tmpl001",
		Data: map[string]*SubscribeMessageValue{
			"thing1": {Value: "test value"},
		},
	}
	err := c.SendSubscribeMessage(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMsgSecCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/wxa/msg_sec_check" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode":  0,
				"errmsg":   "ok",
				"trace_id": "trace001",
				"result":   map[string]interface{}{"suggest": "pass", "label": 100},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "tok", "expires_in": 7200})
		}
	}))
	defer srv.Close()

	cfg := &Config{BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"}}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "POST")
	c := &Client{BaseClient: base}

	req := &MsgSecCheckRequest{
		Content: "hello world",
		Version: 2,
		Scene:   2,
		Openid:  "openid123",
	}
	result, err := c.MsgSecCheck(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TraceId != "trace001" {
		t.Errorf("expected trace001, got %s", result.TraceId)
	}
}
```

- [ ] **Step 2:** Run build and tests to verify:

```bash
cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
go build ./...
go test ./mini-program/ -run TestSendSubscribeMessage -v
go test ./mini-program/ -run TestMsgSecCheck -v
```

- [ ] **Step 3:** Commit:

```bash
git add mini-program/api_subscribe_test.go
git commit -m "test(mini-program): add subscribe and security tests"
```
