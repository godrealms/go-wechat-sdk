# work-wechat ISV 子项目 4:应用消息发送 + 工作台自定义展示

**Date:** 2026-04-12
**Status:** Draft
**Scope:** 企业微信 work-wechat/isv 包 —— CorpClient HTTP 基础设施 + 11 种消息 Send 方法 + 3 个工作台方法
**Depends on:** 子项目 1(ISV 认证底座)

---

## 1. 目标

在已完成的 `work-wechat/isv` 包中:

1. 为 `*CorpClient` 增加 **2 个私有 HTTP helper**(`doPost` / `doGet`)
2. 增加 **11 个消息发送方法**(`SendText` ~ `SendTemplateCard`)
3. 增加 **3 个工作台方法**(`SetWorkbenchTemplate` / `GetWorkbenchTemplate` / `SetWorkbenchData`)

追加后 `*CorpClient` 公开 API 从 3 个方法增加到 **17 个方法**。

### 1.1 必须交付的 14 个公开方法

#### 消息发送(11 个)

| # | 签名 | msgtype | Token |
|---|---|---|---|
| 1 | `(cc *CorpClient) SendText(ctx, req *SendTextReq) (*SendMessageResp, error)` | text | corp |
| 2 | `(cc *CorpClient) SendImage(ctx, req *SendImageReq) (*SendMessageResp, error)` | image | corp |
| 3 | `(cc *CorpClient) SendVoice(ctx, req *SendVoiceReq) (*SendMessageResp, error)` | voice | corp |
| 4 | `(cc *CorpClient) SendVideo(ctx, req *SendVideoReq) (*SendMessageResp, error)` | video | corp |
| 5 | `(cc *CorpClient) SendFile(ctx, req *SendFileReq) (*SendMessageResp, error)` | file | corp |
| 6 | `(cc *CorpClient) SendTextCard(ctx, req *SendTextCardReq) (*SendMessageResp, error)` | textcard | corp |
| 7 | `(cc *CorpClient) SendNews(ctx, req *SendNewsReq) (*SendMessageResp, error)` | news | corp |
| 8 | `(cc *CorpClient) SendMpNews(ctx, req *SendMpNewsReq) (*SendMessageResp, error)` | mpnews | corp |
| 9 | `(cc *CorpClient) SendMarkdown(ctx, req *SendMarkdownReq) (*SendMessageResp, error)` | markdown | corp |
| 10 | `(cc *CorpClient) SendMiniProgramNotice(ctx, req *SendMiniProgramNoticeReq) (*SendMessageResp, error)` | miniprogram_notice | corp |
| 11 | `(cc *CorpClient) SendTemplateCard(ctx, req *SendTemplateCardReq) (*SendMessageResp, error)` | template_card | corp |

#### 工作台(3 个)

| # | 签名 | 接口 | Token |
|---|---|---|---|
| 12 | `(cc *CorpClient) SetWorkbenchTemplate(ctx, req *WorkbenchTemplateReq) error` | POST `/cgi-bin/agent/set_workbench_template` | corp |
| 13 | `(cc *CorpClient) GetWorkbenchTemplate(ctx, agentID int) (*WorkbenchTemplateResp, error)` | POST `/cgi-bin/agent/get_workbench_template` | corp |
| 14 | `(cc *CorpClient) SetWorkbenchData(ctx, req *WorkbenchDataReq) error` | POST `/cgi-bin/agent/set_workbench_data` | corp |

### 1.2 非目标

- **媒体文件上传**(`media/upload`):消息发送所需的 `media_id` 由调用方预先获取,本轮不实现 upload。
- **消息撤回**(`message/recall`):留到后续子项目。
- **互动消息回调**(`update_template_card`):事件回调属于子项目 6。
- **CorpClient.doGet**:本轮所有接口都是 POST,但 helper 一并实现以供后续子项目使用。

## 2. 架构决策

### 2.1 CorpClient HTTP helper

`CorpClient` 目前仅实现 `TokenSource`(3 个方法:`CorpID` / `AccessToken` / `Refresh`)。消息和工作台 API 都使用 `corp_access_token`(企业微信 query key 为 `access_token`)。

新增 2 个私有 helper,挂在 `*CorpClient` 上,复用 `parent.doPostRaw` / `parent.doRequestRaw`:

```go
// corp.http.go

func (cc *CorpClient) doPost(ctx context.Context, path string, body, out interface{}) error {
    tok, err := cc.AccessToken(ctx)
    if err != nil {
        return err
    }
    q := url.Values{"access_token": {tok}}
    return cc.parent.doPostRaw(ctx, path, q, body, out)
}

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

**注意:** 企业微信 corp 接口的 query key 是 `access_token`(不是 `corp_access_token`),与 suite/provider 不同。

### 2.2 消息发送模式

所有 11 种消息都发往 `POST /cgi-bin/message/send`,差异仅在 `msgtype` 字段和对应的 content 块。每个 Send 方法用一个内部 wire 结构体自动注入 `msgtype`:

```go
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
```

调用方不需要设置 `msgtype`,只需填充类型专属字段。

### 2.3 文件拆分

```
work-wechat/isv/
├── corp.http.go              # NEW — 2 个私有 HTTP helper
├── corp.http_test.go         # NEW — 2 个基础测试
├── struct.message.go         # NEW — MessageHeader + 11 种消息 DTO + SendMessageResp
├── corp.message.go           # NEW — 11 个 Send* 方法
├── corp.message_test.go      # NEW — ~6 个测试
├── struct.workbench.go       # NEW — 工作台 DTO
├── corp.workbench.go         # NEW — 3 个工作台方法
└── corp.workbench_test.go    # NEW — 3 个测试
```

## 3. DTO 设计

### 3.1 消息公共字段

```go
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
```

### 3.2 消息响应

```go
// SendMessageResp 是 message/send 的统一响应。
type SendMessageResp struct {
    InvalidUser    string `json:"invaliduser"`
    InvalidParty   string `json:"invalidparty"`
    InvalidTag     string `json:"invalidtag"`
    UnlicensedUser string `json:"unlicenseduser"`
    MsgID          string `json:"msgid"`
    ResponseCode   string `json:"response_code"`
}
```

### 3.3 简单消息类型

```go
// --- text ---
type TextContent struct {
    Content string `json:"content"`
}
type SendTextReq struct {
    MessageHeader
    Text TextContent `json:"text"`
}

// --- image ---
type ImageContent struct {
    MediaID string `json:"media_id"`
}
type SendImageReq struct {
    MessageHeader
    Image ImageContent `json:"image"`
}

// --- voice ---
type VoiceContent struct {
    MediaID string `json:"media_id"`
}
type SendVoiceReq struct {
    MessageHeader
    Voice VoiceContent `json:"voice"`
}

// --- video ---
type VideoContent struct {
    MediaID     string `json:"media_id"`
    Title       string `json:"title,omitempty"`
    Description string `json:"description,omitempty"`
}
type SendVideoReq struct {
    MessageHeader
    Video VideoContent `json:"video"`
}

// --- file ---
type FileContent struct {
    MediaID string `json:"media_id"`
}
type SendFileReq struct {
    MessageHeader
    File FileContent `json:"file"`
}

// --- textcard ---
type TextCardContent struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    URL         string `json:"url"`
    BtnTxt      string `json:"btntxt,omitempty"`
}
type SendTextCardReq struct {
    MessageHeader
    TextCard TextCardContent `json:"textcard"`
}

// --- news ---
type NewsArticle struct {
    Title       string `json:"title"`
    Description string `json:"description,omitempty"`
    URL         string `json:"url,omitempty"`
    PicURL      string `json:"picurl,omitempty"`
    AppID       string `json:"appid,omitempty"`
    PagePath    string `json:"pagepath,omitempty"`
}
type NewsContent struct {
    Articles []NewsArticle `json:"articles"`
}
type SendNewsReq struct {
    MessageHeader
    News NewsContent `json:"news"`
}

// --- mpnews ---
type MpNewsArticle struct {
    Title            string `json:"title"`
    ThumbMediaID     string `json:"thumb_media_id"`
    Author           string `json:"author,omitempty"`
    ContentSourceURL string `json:"content_source_url,omitempty"`
    Content          string `json:"content"`
    Digest           string `json:"digest,omitempty"`
}
type MpNewsContent struct {
    Articles []MpNewsArticle `json:"articles"`
}
type SendMpNewsReq struct {
    MessageHeader
    MpNews MpNewsContent `json:"mpnews"`
}

// --- markdown ---
type MarkdownContent struct {
    Content string `json:"content"`
}
type SendMarkdownReq struct {
    MessageHeader
    Markdown MarkdownContent `json:"markdown"`
}

// --- miniprogram_notice ---
type MiniProgramContentItem struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}
type MiniProgramNoticeContent struct {
    AppID             string                   `json:"appid"`
    Page              string                   `json:"page,omitempty"`
    Title             string                   `json:"title"`
    Description       string                   `json:"description,omitempty"`
    EmphasisFirstItem bool                     `json:"emphasis_first_item,omitempty"`
    ContentItem       []MiniProgramContentItem `json:"content_item,omitempty"`
}
type SendMiniProgramNoticeReq struct {
    MessageHeader
    MiniProgramNotice MiniProgramNoticeContent `json:"miniprogram_notice"`
}
```

### 3.4 TemplateCard 类型

`template_card` 有 5 种子类型,嵌套结构较深。所有嵌套类型统一用 `TC` 前缀避免与其他 DTO 冲突。

```go
// --- template_card ---

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

// TemplateCardContent 是 template_card 消息的 content 结构。
// card_type 决定哪些字段有效:
//   - text_notice / news_notice: 基本字段
//   - button_interaction: + ButtonSelection + ButtonList
//   - vote_interaction: + ButtonSelection + ButtonList
//   - multiple_interaction: + SelectList + Checkbox + SubmitButton
type TemplateCardContent struct {
    CardType                string                `json:"card_type"`
    Source                  *TCSource             `json:"source,omitempty"`
    ActionMenu              *TCActionMenu         `json:"action_menu,omitempty"`
    TaskID                  string                `json:"task_id,omitempty"`
    MainTitle               TCMainTitle           `json:"main_title"`
    EmphasisContent         *TCEmphasisContent    `json:"emphasis_content,omitempty"`
    QuoteArea               *TCQuoteArea          `json:"quote_area,omitempty"`
    SubTitleText            string                `json:"sub_title_text,omitempty"`
    HorizontalContentList   []TCHorizontalContent `json:"horizontal_content_list,omitempty"`
    JumpList                []TCJumpItem          `json:"jump_list,omitempty"`
    CardAction              TCCardAction          `json:"card_action"`
    ButtonSelection         *TCButtonSelection    `json:"button_selection,omitempty"`
    ButtonList              []TCButton            `json:"button_list,omitempty"`
    SelectList              []TCSelectItem        `json:"select_list,omitempty"`
    Checkbox                *TCCheckbox           `json:"checkbox,omitempty"`
    SubmitButton            *TCSubmitButton       `json:"submit_button,omitempty"`
    CardImage               *TCCardImage          `json:"card_image,omitempty"`
    ImageTextArea           *TCImageTextArea      `json:"image_text_area,omitempty"`
    VerticalContentList     []TCVerticalContent   `json:"vertical_content_list,omitempty"`
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

type SendTemplateCardReq struct {
    MessageHeader
    TemplateCard TemplateCardContent `json:"template_card"`
}
```

### 3.5 工作台 DTO

```go
// --- workbench ---

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

// WBImageItem 图片条目。
type WBImageItem struct {
    URL      string `json:"url"`
    JumpURL  string `json:"jump_url,omitempty"`
    PagePath string `json:"pagepath,omitempty"`
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

## 4. 实现细节

### 4.1 Send 方法模板

所有 11 个 Send 方法结构一致:

```go
func (cc *CorpClient) Send<Type>(ctx context.Context, req *Send<Type>Req) (*SendMessageResp, error) {
    var wire struct {
        MsgType string `json:"msgtype"`
        Send<Type>Req
    }
    wire.MsgType = "<msgtype>"
    wire.Send<Type>Req = *req
    var resp SendMessageResp
    if err := cc.doPost(ctx, "/cgi-bin/message/send", wire, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.2 工作台方法

```go
func (cc *CorpClient) SetWorkbenchTemplate(ctx context.Context, req *WorkbenchTemplateReq) error {
    return cc.doPost(ctx, "/cgi-bin/agent/set_workbench_template", req, nil)
}

func (cc *CorpClient) GetWorkbenchTemplate(ctx context.Context, agentID int) (*WorkbenchTemplateResp, error) {
    body := map[string]int{"agentid": agentID}
    var resp WorkbenchTemplateResp
    if err := cc.doPost(ctx, "/cgi-bin/agent/get_workbench_template", body, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

func (cc *CorpClient) SetWorkbenchData(ctx context.Context, req *WorkbenchDataReq) error {
    return cc.doPost(ctx, "/cgi-bin/agent/set_workbench_data", req, nil)
}
```

## 5. 测试策略

### 5.1 测试辅助函数

需要新建一个 `newTestCorpClient(t, baseURL)` helper,它:
1. 创建 ISV Client (with test config + baseURL)
2. 往 Store 写入一条假的 authorizer token
3. 返回 `*CorpClient`

这避免每个测试都重复 token 换取的 httptest mock。

### 5.2 测试矩阵

| 文件 | # | Case | 断言 |
|---|---|---|---|
| `corp.http_test.go` | 1 | `TestCorpClient_DoPost_TokenInjection` | 验证 query 含 `access_token=CTOK`,body 正确传递 |
| `corp.http_test.go` | 2 | `TestCorpClient_DoGet_TokenInjection` | 验证 query 含 `access_token=CTOK` + extra params |
| `corp.message_test.go` | 3 | `TestSendText_HappyPath` | text content 映射,resp.MsgID 非空 |
| `corp.message_test.go` | 4 | `TestSendTextCard_HappyPath` | textcard 四字段映射 |
| `corp.message_test.go` | 5 | `TestSendMarkdown_HappyPath` | markdown content 映射 |
| `corp.message_test.go` | 6 | `TestSendNews_MultiArticle` | articles 数组正确序列化 |
| `corp.message_test.go` | 7 | `TestSendTemplateCard_TextNotice` | template_card card_type=text_notice,主要嵌套字段映射 |
| `corp.message_test.go` | 8 | `TestSendMessage_WeixinError` | errcode 非零返回 *WeixinError |
| `corp.workbench_test.go` | 9 | `TestSetWorkbenchTemplate` | key_data 模板设置,无 error |
| `corp.workbench_test.go` | 10 | `TestGetWorkbenchTemplate` | 返回 type + key_data 字段映射 |
| `corp.workbench_test.go` | 11 | `TestSetWorkbenchData` | userid + key_data 设置,无 error |

覆盖率目标 ≥85%。不需要 11 个 Send 方法各一个测试 —— 结构完全一致,选代表性类型覆盖即可。

## 6. 错误处理

- HTTP 方法沿用 `doPostRaw` → `decodeRaw` → `WeixinError` 两阶段解码。
- `SetWorkbenchTemplate` / `SetWorkbenchData` 返回 `error`(成功时企业微信返回 `errcode=0`),`out` 传 `nil` 即可。
- `CorpClient.AccessToken` 失败时透传 `ErrAuthorizerRevoked` / Store 错误。
- 未新增哨兵错误。

## 7. 交付规模估计

| 文件 | 生产行数 | 测试行数 |
|---|---|---|
| `corp.http.go` | ~35 | — |
| `struct.message.go` | ~300 | — |
| `corp.message.go` | ~150 | — |
| `struct.workbench.go` | ~80 | — |
| `corp.workbench.go` | ~40 | — |
| `corp.http_test.go` | — | ~80 |
| `corp.message_test.go` | — | ~220 |
| `corp.workbench_test.go` | — | ~100 |
| **合计** | **~605** | **~400** |

## 8. Commit 节奏

~6 个原子 commit:

1. CorpClient HTTP helpers(`corp.http.go` + `corp.http_test.go`)
2. 消息 DTO(`struct.message.go`,不含 TemplateCard)
3. TemplateCard DTO(追加到 `struct.message.go`)
4. Send 方法 + 消息测试(`corp.message.go` + `corp.message_test.go`)
5. 工作台 DTO + 方法 + 测试(`struct.workbench.go` + `corp.workbench.go` + `corp.workbench_test.go`)
6. 全量 `-race` / `-cover` / `./...` 确认(最后一步合并到 commit 5)

## 9. Self-Review Checklist

- [ ] 2 个私有 HTTP helper + 14 个公开方法全部实现
- [ ] 企业微信 corp 接口 query key 是 `access_token`(不是 `corp_access_token`)
- [ ] 每个 Send 方法内部自动注入 `msgtype`,调用方无需关心
- [ ] MessageHeader 嵌入正确展开为 JSON 顶层字段
- [ ] TemplateCard 嵌套类型用 `TC` 前缀,工作台用 `WB` 前缀
- [ ] 未新增哨兵错误
- [ ] 未修改 Config / Store / Client 结构
- [ ] `go test -race ./work-wechat/isv/...` 通过
- [ ] 覆盖率 ≥85%

---

**spec 完成。**
