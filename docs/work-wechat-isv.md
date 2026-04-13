# work-wechat/isv — 企业微信 ISV 服务商

> 企业微信第三方应用（ISV）服务商模式 SDK：管理套件凭证、代多个企业调用通讯录/审批/日历/打卡/外部联系人/消息/JSSDK 等接口。

## 适用场景

此包适用于以下场景：

- 你是企业微信第三方应用服务商（ISV），开发了一个套件（suite），需要代多个企业（授权方）调用企业微信 API。
- 套件层面：接收微信回调（`suite_ticket`、授权变更事件）、换取 `suite_access_token`、获取预授权码、完成企业授权流程。
- 企业层面：代企业调用通讯录（成员/部门/标签）、应用管理、消息发送、审批、日历、打卡、外部联系人、JSSDK 等接口。
- 服务商层面（Provider）：服务商登录、企业注册、成员 ID 转换（`open_userid`）。

**核心模式：**

```
服务商（ISV）
    └── Client（持有 suite 凭证，管理所有企业的 token）
            └── CorpClient（per-企业句柄，代该企业发起 API 调用）
```

## 初始化 / Initialization

```go
func NewClient(cfg Config, opts ...Option) (*Client, error)
```

创建 ISV 客户端。`SuiteID`、`SuiteSecret`、`Token`、`EncodingAESKey`（43 字符）为必填项。`ProviderCorpID` 与 `ProviderSecret` 要么同时填写，要么同时留空（仅 Provider 级接口需要）。

```go
cfg := isv.Config{
    SuiteID:        "ww...",                              // 第三方应用 suite_id
    SuiteSecret:    "xxx",                                // 第三方应用 suite_secret
    ProviderCorpID: "ww...",                              // 服务商企业 corpid（可选）
    ProviderSecret: "yyy",                                // 服务商 provider_secret（可选）
    Token:          "callback_token",                     // 回调验签 token
    EncodingAESKey: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ", // 43 字符
}

client, err := isv.NewClient(cfg,
    isv.WithStore(isv.NewMemoryStore()), // 生产环境请换成持久化 Store
)
```

### CorpClient —— 企业级操作句柄

所有企业级接口均通过 `CorpClient` 调用：

```go
func (c *Client) CorpClient(corpID string) *CorpClient
```

`CorpClient` 代表对指定企业的访问会话，内部自动管理 `corp_access_token` 的获取与刷新（lazy + 双检锁）。

```go
corpClient := client.CorpClient("wx_corp123")
// 之后所有企业级 API 通过 corpClient 调用
users, err := corpClient.ListUserDetail(ctx, 1, true)
```

### Options

| Option | 说明 |
|--------|------|
| `WithStore(s Store)` | 注入自定义持久化 Store（生产环境必须替换默认的内存 Store） |
| `WithHTTPClient(h *http.Client)` | 注入自定义 HTTP 客户端（用于测试或自定义超时） |
| `WithBaseURL(u string)` | 覆盖 API 基础 URL（测试用） |

### Store 接口

`Store` 负责持久化套件票据（`suite_ticket`）、各类 token 及企业授权码。默认实现 `MemoryStore` 仅适合单进程测试；生产环境应实现 `Store` 接口并持久化到 Redis 或数据库。

```go
type Store interface {
    GetSuiteTicket(ctx context.Context, suiteID string) (string, error)
    PutSuiteTicket(ctx context.Context, suiteID, ticket string) error

    GetSuiteToken(ctx context.Context, suiteID string) (token string, expiresAt time.Time, err error)
    PutSuiteToken(ctx context.Context, suiteID, token string, expiresAt time.Time) error

    GetProviderToken(ctx context.Context, suiteID string) (token string, expiresAt time.Time, err error)
    PutProviderToken(ctx context.Context, suiteID, token string, expiresAt time.Time) error

    GetAuthorizer(ctx context.Context, suiteID, corpID string) (*AuthorizerTokens, error)
    PutAuthorizer(ctx context.Context, suiteID, corpID string, tokens *AuthorizerTokens) error
    DeleteAuthorizer(ctx context.Context, suiteID, corpID string) error
    ListAuthorizers(ctx context.Context, suiteID string) ([]string, error)
}
```

## 错误处理 / Error Handling

微信 API 返回的非零 errcode 会被包装为 `*WeixinError`，使用 `errors.As` 解包：

```go
var wxErr *isv.WeixinError
if errors.As(err, &wxErr) {
    fmt.Printf("errcode=%d errmsg=%s\n", wxErr.ErrCode, wxErr.ErrMsg)
}
```

哨兵错误（可用 `errors.Is` 判断）：

| 错误 | 含义 |
|------|------|
| `isv.ErrNotFound` | Store 中未找到对应记录 |
| `isv.ErrSuiteTicketMissing` | Store 中无 `suite_ticket`（需等待微信推送） |
| `isv.ErrProviderCorpIDMissing` | Config 未配置 `ProviderCorpID` |
| `isv.ErrProviderSecretMissing` | Config 未配置 `ProviderSecret` |
| `isv.ErrAuthorizerRevoked` | 企业已撤销授权或尚未完成授权流程 |

---

## API Reference

### ISV Client 级别

#### ParseNotify

```go
func (c *Client) ParseNotify(r *http.Request) (Event, error)
```

验签、解密并解析企业微信 ISV 回调请求，返回强类型 `Event`。`suite_ticket` 事件会自动写入 Store。

| 参数 | 类型 | 说明 |
|------|------|------|
| `r` | `*http.Request` | 来自微信服务器的 HTTP 回调请求 |

返回类型包括 `*SuiteTicketEvent`、`*CreateAuthEvent`、`*ChangeAuthEvent`、`*CancelAuthEvent`、`*ResetPermanentCodeEvent`、各类通讯录变更事件、外部联系人变更事件，以及未知类型的 `*RawEvent`。

#### ParseDataNotify

```go
func (c *Client) ParseDataNotify(r *http.Request) (DataEvent, error)
```

验签、解密并解析企业微信数据回调（企业应用消息回调），返回强类型 `DataEvent`。

| 参数 | 类型 | 说明 |
|------|------|------|
| `r` | `*http.Request` | 来自微信服务器的 HTTP 数据回调请求 |

#### GetSuiteAccessToken

```go
func (c *Client) GetSuiteAccessToken(ctx context.Context) (string, error)
```

获取 `suite_access_token`（lazy + 双检锁 + 自动缓存），过期前 5 分钟自动刷新。

#### RefreshSuiteToken

```go
func (c *Client) RefreshSuiteToken(ctx context.Context) error
```

强制忽略缓存，刷新 `suite_access_token`。

#### RefreshAll

```go
func (c *Client) RefreshAll(ctx context.Context) error
```

遍历 Store 中所有已授权企业，刷新它们的 `corp_access_token`。任一失败继续下一个，最后聚合错误。

---

### 套件授权流程 / Suite Authorization

#### GetPreAuthCode

```go
func (c *Client) GetPreAuthCode(ctx context.Context) (*PreAuthCodeResp, error)
```

拉取预授权码，用于构造企业管理员授权 URL。

#### SetSessionInfo

```go
func (c *Client) SetSessionInfo(ctx context.Context, preAuthCode string, info *SessionInfo) error
```

为预授权码绑定授权会话配置（限定可授权的应用或授权类型）。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `preAuthCode` | `string` | 预授权码 |
| `info` | `*SessionInfo` | 会话配置（AppID 列表、AuthType） |

#### AuthorizeURL

```go
func (c *Client) AuthorizeURL(preAuthCode, redirectURI, state string) string
```

拼接企业管理员扫码授权的跳转 URL（纯计算，不发起 HTTP 请求）。

#### GetPermanentCode

```go
func (c *Client) GetPermanentCode(ctx context.Context, authCode string) (*PermanentCodeResp, error)
```

用授权回调中的 `auth_code` 换取企业永久授权码，并自动将授权信息（`permanent_code` + 初始 `corp_access_token`）写入 Store。**授权流程完成后必须调用此方法。**

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `authCode` | `string` | `CreateAuthEvent.AuthCode` |

返回 `*PermanentCodeResp`，包含 `AuthCorpInfo`（企业信息）、`PermanentCode`、`AuthInfo`（授权应用信息）。

#### GetAuthInfo

```go
func (c *Client) GetAuthInfo(ctx context.Context, corpID, permanentCode string) (*AuthInfoResp, error)
```

查询企业授权信息（不缓存）。

#### GetAdminList

```go
func (c *Client) GetAdminList(ctx context.Context, corpID, agentID string) (*AdminListResp, error)
```

获取指定企业授权应用的管理员列表。

#### GetCorpToken

```go
func (c *Client) GetCorpToken(ctx context.Context, corpID, permanentCode string) (*CorpTokenResp, error)
```

直接调用 `service/get_corp_token` 换取企业 `corp_access_token`（底层方法，通常应通过 `CorpClient.AccessToken` 使用）。

---

### Provider 级别接口

以下接口使用 `provider_access_token`，需要在 Config 中配置 `ProviderCorpID` 和 `ProviderSecret`。

#### GetProviderAccessToken

```go
func (c *Client) GetProviderAccessToken(ctx context.Context) (string, error)
```

获取 `provider_access_token`（lazy 获取 + 自动缓存）。

#### GetLoginInfo

```go
func (c *Client) GetLoginInfo(ctx context.Context, authCode string) (*LoginInfoResp, error)
```

用服务商管理端 OAuth 回跳的 `auth_code` 换取登录身份信息。

#### GetRegisterCode

```go
func (c *Client) GetRegisterCode(ctx context.Context, req *GetRegisterCodeReq) (*RegisterCodeResp, error)
```

生成注册企业微信的 `register_code`（邀请链接核心参数）。

#### GetRegistrationInfo

```go
func (c *Client) GetRegistrationInfo(ctx context.Context, registerCode string) (*RegistrationInfoResp, error)
```

查询 `register_code` 对应的注册进度，成功后返回已注册企业的 corpid、管理员、永久授权码等信息。

#### CorpIDToOpenCorpID

```go
func (c *Client) CorpIDToOpenCorpID(ctx context.Context, corpID string) (string, error)
```

把企业 corpid 转换为跨服务商匿名的 `open_corpid`。

#### UserIDToOpenUserID

```go
func (c *Client) UserIDToOpenUserID(ctx context.Context, corpID string, userIDs []string) (*UserIDConvertResp, error)
```

批量将企业内部 `userid` 转换为跨服务商匿名的 `open_userid`。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `corpID` | `string` | 授权企业的 corpid |
| `userIDs` | `[]string` | 待转换的 userid 列表 |

#### OAuth2URL

```go
func (c *Client) OAuth2URL(redirectURI, state string, opts ...OAuth2Option) string
```

构造企业微信第三方网页授权 URL（纯计算）。可选参数 `WithOAuth2Scope(scope)`、`WithOAuth2AgentID(agentID)`。

#### GetUserInfo3rd

```go
func (c *Client) GetUserInfo3rd(ctx context.Context, authCode string) (*UserInfo3rdResp, error)
```

用网页 OAuth2 回调的 `code` 换取成员身份（`userid` / `user_ticket` / `open_userid`）。

#### GetUserDetail3rd

```go
func (c *Client) GetUserDetail3rd(ctx context.Context, userTicket string) (*UserDetail3rdResp, error)
```

用 `user_ticket` 换取成员敏感详情（姓名/邮箱/头像/手机号）。调用前请确认合规备案。

---

### CorpClient — 成员管理

#### CreateUser

```go
func (cc *CorpClient) CreateUser(ctx context.Context, req *CreateUserReq) error
```

在授权企业中创建成员。

#### UpdateUser

```go
func (cc *CorpClient) UpdateUser(ctx context.Context, req *UpdateUserReq) error
```

更新授权企业中的成员信息。

#### DeleteUser

```go
func (cc *CorpClient) DeleteUser(ctx context.Context, userID string) error
```

删除授权企业中指定 `userid` 的成员。

#### GetUser

```go
func (cc *CorpClient) GetUser(ctx context.Context, userID string) (*UserDetail, error)
```

获取指定成员的详细信息。

#### ListUserSimple

```go
func (cc *CorpClient) ListUserSimple(ctx context.Context, deptID int, fetchChild bool) (*UserSimpleListResp, error)
```

列出部门成员（仅基本信息：userid + 姓名）。`fetchChild=true` 时递归返回子部门成员。

#### ListUserDetail

```go
func (cc *CorpClient) ListUserDetail(ctx context.Context, deptID int, fetchChild bool) (*UserDetailListResp, error)
```

列出部门成员（全量字段）。

---

### CorpClient — 部门管理

#### CreateDepartment

```go
func (cc *CorpClient) CreateDepartment(ctx context.Context, req *CreateDeptReq) (*CreateDeptResp, error)
```

创建部门，返回新部门 ID。

#### UpdateDepartment

```go
func (cc *CorpClient) UpdateDepartment(ctx context.Context, req *UpdateDeptReq) error
```

更新部门信息。

#### DeleteDepartment

```go
func (cc *CorpClient) DeleteDepartment(ctx context.Context, id int) error
```

删除指定 ID 的部门。

#### ListDepartment

```go
func (cc *CorpClient) ListDepartment(ctx context.Context, id int) ([]Department, error)
```

列出指定部门 ID 下的子部门（`id=0` 为根部门）。

---

### CorpClient — 标签管理

#### CreateTag

```go
func (cc *CorpClient) CreateTag(ctx context.Context, req *CreateTagReq) (*CreateTagResp, error)
```

创建标签。

#### UpdateTag

```go
func (cc *CorpClient) UpdateTag(ctx context.Context, req *UpdateTagReq) error
```

更新标签。

#### DeleteTag

```go
func (cc *CorpClient) DeleteTag(ctx context.Context, tagID int) error
```

删除标签。

#### ListTag

```go
func (cc *CorpClient) ListTag(ctx context.Context) ([]Tag, error)
```

列出企业所有标签。

#### GetTagUsers

```go
func (cc *CorpClient) GetTagUsers(ctx context.Context, tagID int) (*TagUsersResp, error)
```

获取指定标签下的成员与部门列表。

#### AddTagUsers

```go
func (cc *CorpClient) AddTagUsers(ctx context.Context, req *TagUsersModifyReq) (*TagUsersModifyResp, error)
```

向标签添加成员或部门。

#### DelTagUsers

```go
func (cc *CorpClient) DelTagUsers(ctx context.Context, req *TagUsersModifyReq) (*TagUsersModifyResp, error)
```

从标签移除成员或部门。

---

### CorpClient — 通讯录邀请

#### InviteUser

```go
func (cc *CorpClient) InviteUser(ctx context.Context, req *InviteReq) (*InviteResp, error)
```

批量邀请成员、部门或标签中的成员加入企业。

---

### CorpClient — 应用管理

#### GetAgent

```go
func (cc *CorpClient) GetAgent(ctx context.Context, agentID int) (*AgentDetail, error)
```

获取指定应用（agent）的详情。

#### SetAgent

```go
func (cc *CorpClient) SetAgent(ctx context.Context, req *SetAgentReq) error
```

更新应用属性（名称、描述、主页等）。

---

### CorpClient — 菜单管理

#### CreateMenu

```go
func (cc *CorpClient) CreateMenu(ctx context.Context, agentID int, req *CreateMenuReq) error
```

为指定应用创建自定义菜单。

#### GetMenu

```go
func (cc *CorpClient) GetMenu(ctx context.Context, agentID int) (*MenuResp, error)
```

获取指定应用的当前自定义菜单。

#### DeleteMenu

```go
func (cc *CorpClient) DeleteMenu(ctx context.Context, agentID int) error
```

删除指定应用的自定义菜单。

---

### CorpClient — 消息发送

所有消息发送方法均调用 `/cgi-bin/message/send` 接口，自动设置 `msgtype` 字段。

#### SendText

```go
func (cc *CorpClient) SendText(ctx context.Context, req *SendTextReq) (*SendMessageResp, error)
```

发送文本消息。

#### SendImage

```go
func (cc *CorpClient) SendImage(ctx context.Context, req *SendImageReq) (*SendMessageResp, error)
```

发送图片消息。

#### SendVoice

```go
func (cc *CorpClient) SendVoice(ctx context.Context, req *SendVoiceReq) (*SendMessageResp, error)
```

发送语音消息。

#### SendVideo

```go
func (cc *CorpClient) SendVideo(ctx context.Context, req *SendVideoReq) (*SendMessageResp, error)
```

发送视频消息。

#### SendFile

```go
func (cc *CorpClient) SendFile(ctx context.Context, req *SendFileReq) (*SendMessageResp, error)
```

发送文件消息。

#### SendTextCard

```go
func (cc *CorpClient) SendTextCard(ctx context.Context, req *SendTextCardReq) (*SendMessageResp, error)
```

发送文本卡片消息。

#### SendNews

```go
func (cc *CorpClient) SendNews(ctx context.Context, req *SendNewsReq) (*SendMessageResp, error)
```

发送图文消息（news 类型）。

#### SendMpNews

```go
func (cc *CorpClient) SendMpNews(ctx context.Context, req *SendMpNewsReq) (*SendMessageResp, error)
```

发送图文消息（mpnews 类型，图文内容存储在企业微信）。

#### SendMarkdown

```go
func (cc *CorpClient) SendMarkdown(ctx context.Context, req *SendMarkdownReq) (*SendMessageResp, error)
```

发送 Markdown 消息。

#### SendMiniProgramNotice

```go
func (cc *CorpClient) SendMiniProgramNotice(ctx context.Context, req *SendMiniProgramNoticeReq) (*SendMessageResp, error)
```

发送小程序通知消息。

#### SendTemplateCard

```go
func (cc *CorpClient) SendTemplateCard(ctx context.Context, req *SendTemplateCardReq) (*SendMessageResp, error)
```

发送模板卡片消息。

---

### CorpClient — 工作台

#### SetWorkbenchTemplate

```go
func (cc *CorpClient) SetWorkbenchTemplate(ctx context.Context, req *WorkbenchTemplateReq) error
```

设置应用在工作台的展示模板。

#### GetWorkbenchTemplate

```go
func (cc *CorpClient) GetWorkbenchTemplate(ctx context.Context, agentID int) (*WorkbenchTemplateResp, error)
```

获取应用在工作台的当前展示模板。

#### SetWorkbenchData

```go
func (cc *CorpClient) SetWorkbenchData(ctx context.Context, req *WorkbenchDataReq) error
```

设置指定用户在工作台上的个性化展示数据。

---

### CorpClient — 媒体文件

#### UploadMedia

```go
func (cc *CorpClient) UploadMedia(ctx context.Context, mediaType, fileName string, fileData io.Reader) (*UploadMediaResp, error)
```

上传临时素材（图片/语音/视频/文件），返回 `media_id`，有效期 3 天。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `mediaType` | `string` | 素材类型：`"image"` / `"voice"` / `"video"` / `"file"` |
| `fileName` | `string` | 文件名（含扩展名） |
| `fileData` | `io.Reader` | 文件内容 |

返回 `*UploadMediaResp`，其中 `MediaID` 可用于消息发送。

---

### CorpClient — 审批

#### GetApprovalTemplate

```go
func (cc *CorpClient) GetApprovalTemplate(ctx context.Context, templateID string) (*ApprovalTemplateResp, error)
```

根据模板 ID 获取审批模板详情（控件定义、多语言名称等）。

#### ApplyEvent

```go
func (cc *CorpClient) ApplyEvent(ctx context.Context, req *ApplyEventReq) (*ApplyEventResp, error)
```

提交审批申请，返回审批单号 `sp_no`。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req` | `*ApplyEventReq` | 申请人 userid、模板 ID、控件值列表等 |

#### GetApprovalDetail

```go
func (cc *CorpClient) GetApprovalDetail(ctx context.Context, spNo string) (*ApprovalDetailResp, error)
```

根据审批单号查询审批单详情（状态、审批记录、申请数据）。

#### GetApprovalData

```go
func (cc *CorpClient) GetApprovalData(ctx context.Context, req *GetApprovalDataReq) (*GetApprovalDataResp, error)
```

批量拉取满足条件的审批单号列表，支持分页（cursor）和过滤条件。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req` | `*GetApprovalDataReq` | 时间范围（Unix 时间戳）、分页 cursor、过滤条件 |

---

### CorpClient — 日历

#### CreateCalendar

```go
func (cc *CorpClient) CreateCalendar(ctx context.Context, req *CreateCalendarReq) (*CreateCalendarResp, error)
```

创建日历，返回 `cal_id`。

#### UpdateCalendar

```go
func (cc *CorpClient) UpdateCalendar(ctx context.Context, req *UpdateCalendarReq) error
```

更新已有日历的信息。

#### GetCalendar

```go
func (cc *CorpClient) GetCalendar(ctx context.Context, calIDs []string) (*GetCalendarResp, error)
```

批量查询日历详情。

#### DeleteCalendar

```go
func (cc *CorpClient) DeleteCalendar(ctx context.Context, calID string) error
```

删除指定日历。

#### CreateSchedule

```go
func (cc *CorpClient) CreateSchedule(ctx context.Context, req *CreateScheduleReq) (*CreateScheduleResp, error)
```

创建日程事件，返回 `schedule_id`。

#### UpdateSchedule

```go
func (cc *CorpClient) UpdateSchedule(ctx context.Context, req *UpdateScheduleReq) error
```

更新日程事件。

#### GetSchedule

```go
func (cc *CorpClient) GetSchedule(ctx context.Context, scheduleIDs []string) (*GetScheduleResp, error)
```

批量查询日程详情。

#### DeleteSchedule

```go
func (cc *CorpClient) DeleteSchedule(ctx context.Context, scheduleID string) error
```

删除指定日程事件。

#### GetScheduleByCalendar

```go
func (cc *CorpClient) GetScheduleByCalendar(ctx context.Context, req *GetScheduleByCalendarReq) (*GetScheduleByCalendarResp, error)
```

查询指定日历下的日程列表，支持分页。

---

### CorpClient — 打卡

#### GetCheckinData

```go
func (cc *CorpClient) GetCheckinData(ctx context.Context, req *GetCheckinDataReq) (*GetCheckinDataResp, error)
```

获取指定成员在时间范围内的打卡记录。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ctx` | `context.Context` | 请求上下文 |
| `req` | `*GetCheckinDataReq` | 打卡类型（1=上下班/2=外出/3=全部）、时间范围、成员 userid 列表（≤100） |

#### GetCheckinOption

```go
func (cc *CorpClient) GetCheckinOption(ctx context.Context, req *GetCheckinOptionReq) (*GetCheckinOptionResp, error)
```

获取指定成员的打卡规则设置。

#### GetCheckinDayData

```go
func (cc *CorpClient) GetCheckinDayData(ctx context.Context, req *GetCheckinDayDataReq) (*GetCheckinDayDataResp, error)
```

获取指定成员的每日打卡统计报表。

#### GetCheckinMonthData

```go
func (cc *CorpClient) GetCheckinMonthData(ctx context.Context, req *GetCheckinMonthDataReq) (*GetCheckinMonthDataResp, error)
```

获取指定成员的月度打卡统计报表。

---

### CorpClient — 外部联系人

#### GetExternalContact

```go
func (cc *CorpClient) GetExternalContact(ctx context.Context, externalUserID string) (*GetExternalContactResp, error)
```

根据外部联系人 ID 查询其详情及关注该联系人的内部成员信息。

#### ListExternalContact

```go
func (cc *CorpClient) ListExternalContact(ctx context.Context, userID string) (*ListExternalContactResp, error)
```

获取指定内部成员的外部联系人 ID 列表。

#### BatchGetExternalContactByUser

```go
func (cc *CorpClient) BatchGetExternalContactByUser(ctx context.Context, req *BatchGetExternalContactReq) (*BatchGetExternalContactResp, error)
```

批量拉取多个内部成员的外部联系人详情，支持游标分页（`cursor` + `limit`）。

#### RemarkExternalContact

```go
func (cc *CorpClient) RemarkExternalContact(ctx context.Context, req *RemarkExternalContactReq) error
```

更新外部联系人的备注信息（备注、描述、公司、手机、备注图片等）。

#### GetCorpTagList

```go
func (cc *CorpClient) GetCorpTagList(ctx context.Context, req *GetCorpTagListReq) (*GetCorpTagListResp, error)
```

获取企业客户标签库（按标签 ID 或标签组 ID 过滤）。

#### AddCorpTag

```go
func (cc *CorpClient) AddCorpTag(ctx context.Context, req *AddCorpTagReq) (*AddCorpTagResp, error)
```

新建企业客户标签，`req.Tag` 为 `[]CorpTagInput`（`CorpTagInput` 包含 `Name` 和 `Order` 字段）。

#### EditCorpTag

```go
func (cc *CorpClient) EditCorpTag(ctx context.Context, req *EditCorpTagReq) error
```

修改企业客户标签的名称或排序。

#### DelCorpTag

```go
func (cc *CorpClient) DelCorpTag(ctx context.Context, req *DelCorpTagReq) error
```

删除企业客户标签（按标签 ID 或标签组 ID）。

#### MarkTag

```go
func (cc *CorpClient) MarkTag(ctx context.Context, req *MarkTagReq) error
```

为外部联系人打标签或取消标签。

#### GetFollowUserList

```go
func (cc *CorpClient) GetFollowUserList(ctx context.Context) (*FollowUserListResp, error)
```

获取已配置客户联系功能的内部成员列表。

---

### CorpClient — JSSDK

#### GetJSAPITicket

```go
func (cc *CorpClient) GetJSAPITicket(ctx context.Context) (*JSAPITicketResp, error)
```

获取企业 `jsapi_ticket`，用于签名 `wx.config` 调用。返回 `Ticket` 和 `ExpiresIn`。

#### GetAgentConfigTicket

```go
func (cc *CorpClient) GetAgentConfigTicket(ctx context.Context) (*JSAPITicketResp, error)
```

获取应用 `jsapi_ticket`（`type=agent_config`），用于签名 `wx.agentConfig` 调用。

#### SignJSAPI

```go
func SignJSAPI(ticket, nonceStr, timestamp, pageURL string) string
```

计算 JS-SDK 签名（SHA1 哈希，纯计算，不发起网络请求，无需 `ctx`）。

| 参数 | 类型 | 说明 |
|------|------|------|
| `ticket` | `string` | `GetJSAPITicket` 或 `GetAgentConfigTicket` 返回的 ticket |
| `nonceStr` | `string` | 随机字符串 |
| `timestamp` | `string` | Unix 时间戳字符串 |
| `pageURL` | `string` | 当前网页完整 URL（不含 `#` 及其后内容） |

返回签名字符串，直接填入 `wx.config` 的 `signature` 字段。

---

## 完整示例 / Complete Example

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"
    "net/http"

    "github.com/godrealms/go-wechat-sdk/work-wechat/isv"
)

func main() {
    // 1. 初始化 ISV Client
    cfg := isv.Config{
        SuiteID:        "ww_your_suite_id",
        SuiteSecret:    "your_suite_secret",
        ProviderCorpID: "ww_your_provider_corpid",
        ProviderSecret: "your_provider_secret",
        Token:          "your_callback_token",
        EncodingAESKey: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQ",
    }

    client, err := isv.NewClient(cfg,
        isv.WithStore(isv.NewMemoryStore()), // 生产环境换成 Redis/DB Store
    )
    if err != nil {
        log.Fatal(err)
    }

    // 2. 注册回调处理器：接收 suite_ticket 和授权事件
    http.HandleFunc("/wecom/callback", func(w http.ResponseWriter, r *http.Request) {
        ev, err := client.ParseNotify(r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        switch e := ev.(type) {
        case *isv.SuiteTicketEvent:
            // suite_ticket 已自动写入 Store
            log.Printf("suite_ticket updated: %s", e.SuiteTicket)

        case *isv.CreateAuthEvent:
            // 企业完成授权，用 auth_code 换取永久授权码
            resp, err := client.GetPermanentCode(r.Context(), e.AuthCode)
            if err != nil {
                log.Printf("GetPermanentCode error: %v", err)
                return
            }
            log.Printf("corp %s (%s) authorized", resp.AuthCorpInfo.CorpName, resp.AuthCorpInfo.CorpID)

        case *isv.CancelAuthEvent:
            log.Printf("corp %s cancelled authorization", e.AuthCorpID)
        }

        _, _ = w.Write([]byte("success"))
    })

    // 3. 构造预授权 URL，引导企业管理员授权
    ctx := context.Background()
    preAuth, err := client.GetPreAuthCode(ctx)
    if err != nil {
        log.Fatal(err)
    }
    authURL := client.AuthorizeURL(preAuth.PreAuthCode, "https://your.domain/auth/callback", "csrf_state")
    fmt.Println("授权链接:", authURL)

    // 4. 假设企业 "wx_corp_123" 已完成授权，通过 CorpClient 调用企业级接口
    corpClient := client.CorpClient("wx_corp_123")

    // 4a. 获取部门成员列表
    users, err := corpClient.ListUserDetail(ctx, 1, true)
    if err != nil {
        var wxErr *isv.WeixinError
        if errors.As(err, &wxErr) {
            log.Printf("API error %d: %s", wxErr.ErrCode, wxErr.ErrMsg)
        } else {
            log.Fatal(err)
        }
    } else {
        log.Printf("dept 1 has %d users", len(users.UserList))
    }

    // 4b. 上传临时素材
    // fileData, _ := os.Open("photo.jpg")
    // media, _ := corpClient.UploadMedia(ctx, "image", "photo.jpg", fileData)
    // log.Printf("media_id: %s", media.MediaID)

    // 4c. 获取 JSSDK 票据并签名
    ticket, err := corpClient.GetJSAPITicket(ctx)
    if err == nil {
        sig := isv.SignJSAPI(ticket.Ticket, "random123", "1713000000", "https://your.domain/page")
        fmt.Println("JS-SDK signature:", sig)
    }

    // 4d. 查询外部联系人
    extContacts, err := corpClient.ListExternalContact(ctx, "zhangsan")
    if err == nil {
        log.Printf("zhangsan has %d external contacts", len(extContacts.ExternalUserID))
    }

    // 4e. 批量刷新所有企业的 corp_access_token（定时任务中调用）
    if err := client.RefreshAll(ctx); err != nil {
        log.Printf("RefreshAll partial error: %v", err)
    }

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```
