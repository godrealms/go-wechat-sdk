# Phase 1G: offiaccount Security and Card APIs

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add content security check APIs and WeChat Card (微信卡券) APIs to the offiaccount package.

**Architecture:** New files follow existing offiaccount patterns: use c.Ctx for context, c.TokenQuery() for GET params, fmt.Sprintf path with access_token for POST. Response structs embed offiaccount.Resp (which is core.Resp after the Plan B refactor — or the existing Resp type).

**Tech Stack:** Go 1.23.1, standard library only. Depends on Plan B (offiaccount refactor) being complete.

**PREREQUISITE:** `2026-04-09-phase1b-offiaccount-refactor.md` must be completed first.

---

### Task 1: offiaccount/api.security.go + offiaccount/struct.security.go

**Files:**
- Create: `offiaccount/struct.security.go`
- Create: `offiaccount/api.security.go`

- [ ] Create `offiaccount/struct.security.go` with the following content:

```go
package offiaccount

// MsgSecCheckResult is the result of MsgSecCheck
type MsgSecCheckResult struct {
	Resp
	TraceId string               `json:"trace_id"`
	Result  *MsgSecCheckSummary  `json:"result"`
	Detail  []*MsgSecCheckDetail `json:"detail"`
}

// MsgSecCheckSummary is the overall judgment
type MsgSecCheckSummary struct {
	Suggest string `json:"suggest"` // pass/review/risky
	Label   int    `json:"label"`
}

// MsgSecCheckDetail contains one label's detailed judgment
type MsgSecCheckDetail struct {
	Strategy string `json:"strategy"`
	ErrCode  int    `json:"errcode"`
	Suggest  string `json:"suggest"` // pass/review/risky
	Label    int    `json:"label"`
	Level    int    `json:"level"`
	Prob     int    `json:"prob"`
	KeyWord  string `json:"keyword"`
}

// MsgSecCheckRequest is the request for MsgSecCheck
type MsgSecCheckRequest struct {
	Content   string `json:"content"`
	Version   int    `json:"version"`  // 1 or 2
	Scene     int    `json:"scene"`    // 1=资料 2=评论 3=论坛 4=社交日志
	Openid    string `json:"openid"`
	Title     string `json:"title,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// MediaCheckAsyncResult is the result of MediaCheckAsync
type MediaCheckAsyncResult struct {
	Resp
	TraceId string `json:"trace_id"`
}

// MediaCheckAsyncRequest is the request for MediaCheckAsync
type MediaCheckAsyncRequest struct {
	MediaUrl  string `json:"media_url"`
	MediaType int    `json:"media_type"` // 1=音频 2=图片
	Version   int    `json:"version"`
	Scene     int    `json:"scene"`
	Openid    string `json:"openid"`
}
```

- [ ] Create `offiaccount/api.security.go` with the following content:

```go
package offiaccount

import "fmt"

// MsgSecCheck 检查一段文本是否含有违法违规内容
// POST /wxa/msg_sec_check (access_token in URL)
func (c *Client) MsgSecCheck(req *MsgSecCheckRequest) (*MsgSecCheckResult, error) {
	path := fmt.Sprintf("/wxa/msg_sec_check?access_token=%s", c.GetAccessToken())
	result := &MsgSecCheckResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

// ImgSecCheck 校验一张图片是否含有违法违规内容 (synchronous, image < 1MB)
// POST /wxa/img_sec_check (multipart form, access_token in URL)
func (c *Client) ImgSecCheck(imageData []byte) (*MediaCheckAsyncResult, error) {
	path := fmt.Sprintf("/wxa/img_sec_check?access_token=%s", c.GetAccessToken())
	result := &MediaCheckAsyncResult{}
	err := c.Https.PostForm(c.Ctx, path, "media", "image.jpg", imageData, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
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
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}
```

- [ ] Commit:

```bash
git add offiaccount/struct.security.go offiaccount/api.security.go
git commit -m "feat(offiaccount): add content security check APIs"
```

---

### Task 2: offiaccount/struct.card.go

**Files:**
- Create: `offiaccount/struct.card.go`

- [ ] Create `offiaccount/struct.card.go` with the following content:

```go
package offiaccount

// CardBaseInfo is the common base info for all card types
type CardBaseInfo struct {
	LogoUrl           string       `json:"logo_url"`
	BrandName         string       `json:"brand_name"`
	CodeType          string       `json:"code_type"`          // CODE_TYPE_TEXT/CODE_TYPE_BARCODE/CODE_TYPE_QRCODE/CODE_TYPE_ONLY_QRCODE/CODE_TYPE_ONLY_BARCODE/CODE_TYPE_NONE
	Title             string       `json:"title"`
	SubTitle          string       `json:"sub_title,omitempty"`
	Color             string       `json:"color"`              // Color010 ~ Color100
	Notice            string       `json:"notice"`
	ServicePhone      string       `json:"service_phone,omitempty"`
	Description       string       `json:"description"`
	UseCondition      *UseCondition    `json:"use_condition,omitempty"`
	Abstract          *Abstract        `json:"abstract,omitempty"`
	TextImageList     []*TextImage     `json:"text_image_list,omitempty"`
	TimeInfo          *TimeInfo        `json:"time_info"`
	Sku               *Sku             `json:"sku"`
	LocationIdList    []int64          `json:"location_id_list,omitempty"`
	CenterTitle       string           `json:"center_title,omitempty"`
	CenterSubTitle    string           `json:"center_sub_title,omitempty"`
	CenterUrl         string           `json:"center_url,omitempty"`
	CustomUrl         string           `json:"custom_url,omitempty"`
	CustomUrlName     string           `json:"custom_url_name,omitempty"`
	CustomUrlSubTitle string           `json:"custom_url_sub_title,omitempty"`
	PromotionUrl      string           `json:"promotion_url,omitempty"`
	PromotionUrlName  string           `json:"promotion_url_name,omitempty"`
	GetLimit          int              `json:"get_limit,omitempty"`
	CanShare          bool             `json:"can_share,omitempty"`
	CanGiveFriend     bool             `json:"can_give_friend,omitempty"`
}

// UseCondition specifies card usage conditions
type UseCondition struct {
	AcceptCategory         string `json:"accept_category,omitempty"`
	RejectCategory         string `json:"reject_category,omitempty"`
	CanUseWithOtherDiscount bool   `json:"can_use_with_other_discount,omitempty"`
}

// Abstract is a short description with icon
type Abstract struct {
	Abstract    string   `json:"abstract"`
	IconUrlList []string `json:"icon_url_list,omitempty"`
}

// TextImage is one line in the details section
type TextImage struct {
	ImageUrl string `json:"image_url"`
	Text     string `json:"text"`
}

// TimeInfo specifies the validity period
type TimeInfo struct {
	Type           string `json:"type"`            // DATE_TYPE_FIX_TIME_RANGE / DATE_TYPE_FIX_TERM
	BeginTimestamp int64  `json:"begin_timestamp,omitempty"`
	EndTimestamp   int64  `json:"end_timestamp,omitempty"`
	FixedTerm      int    `json:"fixed_term,omitempty"`
	FixedBeginTerm int    `json:"fixed_begin_term,omitempty"`
}

// Sku specifies quantity info
type Sku struct {
	Quantity int64 `json:"quantity"`
}

// DiscountCard is a discount card (打折券)
type DiscountCard struct {
	BaseInfo *CardBaseInfo `json:"base_info"`
	Discount int           `json:"discount"` // 折扣值，例如填写70表示7折
}

// CashCard is a cash voucher (代金券)
type CashCard struct {
	BaseInfo   *CardBaseInfo `json:"base_info"`
	LeastCost  int64         `json:"least_cost"`  // 使用门槛，单位分，0=无门槛
	ReduceCost int64         `json:"reduce_cost"` // 优惠金额，单位分
}

// MemberCard is a membership card (会员卡)
type MemberCard struct {
	BaseInfo                 *CardBaseInfo `json:"base_info"`
	SupplyBonus              bool          `json:"supply_bonus"`
	SupplyBalance            bool          `json:"supply_balance"`
	BonusCleaned             bool          `json:"bonus_cleared,omitempty"`
	BonusRules               string        `json:"bonus_rules,omitempty"`
	BalanceRules             string        `json:"balance_rules,omitempty"`
	Prerogative              string        `json:"prerogative"`
	AutoActivate             bool          `json:"auto_activate,omitempty"`
	WxActivate               bool          `json:"wx_activate,omitempty"`
	ActivateUrl              string        `json:"activate_url,omitempty"`
	ActivateAppBrandUserName string        `json:"activate_app_brand_user_name,omitempty"`
	ActivateAppBrandPass     string        `json:"activate_app_brand_pass,omitempty"`
}

// CreateCardRequest is the request for CreateCard
type CreateCardRequest struct {
	Card *CardSpec `json:"card"`
}

// CardSpec specifies which card type and its data
type CardSpec struct {
	CardType   string        `json:"card_type"`             // DISCOUNT / CASH / MEMBER_CARD / GROUPON / GIFT
	Discount   *DiscountCard `json:"discount,omitempty"`
	Cash       *CashCard     `json:"cash,omitempty"`
	MemberCard *MemberCard   `json:"member_card,omitempty"`
}

// CreateCardResult is the result of CreateCard
type CreateCardResult struct {
	Resp
	CardId string `json:"card_id"`
}

// GetCardResult is the result of GetCard
type GetCardResult struct {
	Resp
	Card *CardSpec `json:"card"`
}

// UpdateCardRequest is the request for UpdateCard
type UpdateCardRequest struct {
	CardId string    `json:"card_id"`
	Card   *CardSpec `json:"card"`
}

// CardQRCodeRequest is the request for GetCardQRCode
type CardQRCodeRequest struct {
	ActionName    string            `json:"action_name"` // QR_CARD / QR_MULTIPLE_CARD / QR_SCENE
	ActionInfo    *CardQRCodeAction `json:"action_info"`
	ExpireSeconds int               `json:"expire_seconds,omitempty"`
}

// CardQRCodeAction specifies which cards go in the QR code
type CardQRCodeAction struct {
	Card  *CardQRItem  `json:"card,omitempty"`
	Scene *CardQRScene `json:"scene,omitempty"`
}

// CardQRItem is one card in the QR code
type CardQRItem struct {
	CardId       string `json:"card_id"`
	Code         string `json:"code,omitempty"`
	OpenId       string `json:"openid,omitempty"`
	IsUniqueCode bool   `json:"is_unique_code,omitempty"`
}

// CardQRScene is scene info for the QR code
type CardQRScene struct {
	SceneStr string `json:"scene_str,omitempty"`
	SceneId  int    `json:"scene_id,omitempty"`
}

// CardQRCodeResult is the result of GetCardQRCode
type CardQRCodeResult struct {
	Resp
	Ticket        string `json:"ticket"`
	ExpireSeconds int    `json:"expire_seconds"`
	Url           string `json:"url"`
	ShowQrcodeUrl string `json:"show_qrcode_url"`
}

// CardCodeRequest is the request for GetCardCode / ConsumeCardCode
type CardCodeRequest struct {
	Code   string `json:"code"`
	CardId string `json:"card_id,omitempty"`
}

// CardCodeResult is the result of GetCardCode
type CardCodeResult struct {
	Resp
	Card       *CardCodeInfo `json:"card"`
	OpenId     string        `json:"openid"`
	CanConsume bool          `json:"can_consume"`
}

// CardCodeInfo is the card info returned with a code
type CardCodeInfo struct {
	CardId    string `json:"card_id"`
	BeginTime int64  `json:"begin_time"`
	EndTime   int64  `json:"end_time"`
}

// ConsumeCardCodeResult is the result of ConsumeCardCode
type ConsumeCardCodeResult struct {
	Resp
	Card   *CardCodeInfo `json:"card"`
	OpenId string        `json:"openid"`
}

// CardListResult is the result of GetCardList.
// Note: GetCardList uses an inline anonymous struct as return type in api.card.go.
// If the inline form feels awkward, replace it with this named struct instead.
type CardListResult struct {
	Resp
	CardIdList []string `json:"card_id_list"`
	TotalNum   int      `json:"total_num"`
}
```

- [ ] Commit:

```bash
git add offiaccount/struct.card.go
git commit -m "feat(offiaccount): add card structs"
```

---

### Task 3: offiaccount/api.card.go

**Files:**
- Create: `offiaccount/api.card.go`

- [ ] Create `offiaccount/api.card.go` with the following content:

```go
package offiaccount

import (
	"fmt"
	"net/url"
)

// CreateCard 创建卡券
// POST /card/create (access_token in URL)
func (c *Client) CreateCard(req *CreateCardRequest) (*CreateCardResult, error) {
	path := fmt.Sprintf("/card/create?access_token=%s", c.GetAccessToken())
	result := &CreateCardResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

// GetCard 查看卡券详情
// POST /card/get (access_token in URL)
func (c *Client) GetCard(cardId string) (*GetCardResult, error) {
	path := fmt.Sprintf("/card/get?access_token=%s", c.GetAccessToken())
	body := map[string]string{"card_id": cardId}
	result := &GetCardResult{}
	err := c.Https.Post(c.Ctx, path, body, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

// UpdateCard 更改卡券信息
// POST /card/update (access_token in URL)
func (c *Client) UpdateCard(req *UpdateCardRequest) error {
	path := fmt.Sprintf("/card/update?access_token=%s", c.GetAccessToken())
	result := &Resp{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

// DeleteCard 删除卡券
// POST /card/delete (access_token in URL)
func (c *Client) DeleteCard(cardId string) error {
	path := fmt.Sprintf("/card/delete?access_token=%s", c.GetAccessToken())
	body := map[string]string{"card_id": cardId}
	result := &Resp{}
	err := c.Https.Post(c.Ctx, path, body, result)
	if err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

// GetCardQRCode 生成卡券二维码
// POST /card/qrcode/create (access_token in URL)
func (c *Client) GetCardQRCode(req *CardQRCodeRequest) (*CardQRCodeResult, error) {
	path := fmt.Sprintf("/card/qrcode/create?access_token=%s", c.GetAccessToken())
	result := &CardQRCodeResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

// GetCardCode 查询 code 信息
// POST /card/code/get (access_token in URL)
func (c *Client) GetCardCode(req *CardCodeRequest) (*CardCodeResult, error) {
	path := fmt.Sprintf("/card/code/get?access_token=%s", c.GetAccessToken())
	result := &CardCodeResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

// ConsumeCardCode 核销 code
// POST /card/code/consume (access_token in URL)
func (c *Client) ConsumeCardCode(req *CardCodeRequest) (*ConsumeCardCodeResult, error) {
	path := fmt.Sprintf("/card/code/consume?access_token=%s", c.GetAccessToken())
	result := &ConsumeCardCodeResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

// GetCardList 拉取卡券概况数据 (list of all cards)
// POST /card/batchget (access_token in URL)
// Note: the return type uses an inline anonymous struct. If this feels awkward,
// define CardListResult as a named struct in struct.card.go and return *CardListResult instead.
func (c *Client) GetCardList(offset, count int, statusList []string) (*struct {
	Resp
	CardIdList []string `json:"card_id_list"`
	TotalNum   int      `json:"total_num"`
}, error) {
	query := c.TokenQuery(url.Values{
		"offset": {fmt.Sprintf("%d", offset)},
		"count":  {fmt.Sprintf("%d", count)},
	})
	body := map[string]interface{}{
		"offset":      offset,
		"count":       count,
		"status_list": statusList,
	}
	result := &struct {
		Resp
		CardIdList []string `json:"card_id_list"`
		TotalNum   int      `json:"total_num"`
	}{}
	// Note: /card/batchget uses POST with access_token in URL
	path := fmt.Sprintf("/card/batchget?access_token=%s", c.GetAccessToken())
	_ = query // not used for POST
	err := c.Https.Post(c.Ctx, path, body, result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}
```

- [ ] Commit:

```bash
git add offiaccount/api.card.go
git commit -m "feat(offiaccount): add card API methods"
```

---

### Task 4: Tests + build verification

**Files:**
- Create: `offiaccount/api_security_test.go`

- [ ] Create `offiaccount/api_security_test.go` with the following content:

```go
package offiaccount

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/godrealms/go-wechat-sdk/core"
)

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

	cfg := &Config{
		BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"},
	}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "GET")
	c := &Client{BaseClient: base}

	result, err := c.MsgSecCheck(&MsgSecCheckRequest{
		Content: "test content",
		Version: 2,
		Scene:   2,
		Openid:  "openid123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TraceId != "trace001" {
		t.Errorf("expected trace001, got %s", result.TraceId)
	}
	if result.Result.Suggest != "pass" {
		t.Errorf("expected pass, got %s", result.Result.Suggest)
	}
}

func TestCreateCard(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/card/create" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode": 0,
				"errmsg":  "ok",
				"card_id": "pFS7Fjg8kV1IdDz01r4jqycQZtVk",
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "tok", "expires_in": 7200})
		}
	}))
	defer srv.Close()

	cfg := &Config{
		BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"},
	}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "GET")
	c := &Client{BaseClient: base}

	result, err := c.CreateCard(&CreateCardRequest{
		Card: &CardSpec{
			CardType: "DISCOUNT",
			Discount: &DiscountCard{
				BaseInfo: &CardBaseInfo{
					LogoUrl:     "http://example.com/logo.png",
					BrandName:   "TestBrand",
					CodeType:    "CODE_TYPE_TEXT",
					Title:       "测试折扣券",
					Color:       "Color010",
					Notice:      "出示此码享受折扣",
					Description: "测试描述",
					TimeInfo:    &TimeInfo{Type: "DATE_TYPE_FIX_TIME_RANGE", BeginTimestamp: 1700000000, EndTimestamp: 1800000000},
					Sku:         &Sku{Quantity: 100},
				},
				Discount: 70,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CardId != "pFS7Fjg8kV1IdDz01r4jqycQZtVk" {
		t.Errorf("expected card id, got %s", result.CardId)
	}
}
```

- [ ] Run build and tests:

```bash
cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
go build ./...
go test ./offiaccount/ -run TestMsgSecCheck -v
go test ./offiaccount/ -run TestCreateCard -v
```

- [ ] Commit:

```bash
git add offiaccount/api_security_test.go
git commit -m "test(offiaccount): add security and card API tests"
```
