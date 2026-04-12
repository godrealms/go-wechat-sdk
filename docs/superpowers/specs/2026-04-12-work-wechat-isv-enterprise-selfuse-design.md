# Sub-project 8: 自建企业自用 — Design Spec

**Package:** `github.com/godrealms/go-wechat-sdk/work-wechat/isv`
**Date:** 2026-04-12
**Depends on:** Sub-project 1 (认证底座), Sub-project 4 (CorpClient HTTP helpers)

---

## 1. Scope

Five areas of functionality, all on `*CorpClient` (except `SignJSAPI` which is a package-level function):

1. **JS-SDK 签名** — GetJSAPITicket + GetAgentConfigTicket + SignJSAPI
2. **外部联系人管理** — 10 methods for external contact CRUD and tagging
3. **审批流** — GetApprovalTemplate + ApplyEvent + GetApprovalDetail + GetApprovalData
4. **打卡/考勤** — GetCheckinData + GetCheckinOption + GetCheckinDayData + GetCheckinMonthData
5. **日历/日程** — 4 calendar CRUD + 4 schedule CRUD + GetScheduleByCalendar

### 1.1 Method Summary

| # | Method | HTTP | Path | Notes |
|---|---|---|---|---|
| 1 | GetJSAPITicket | GET | `/cgi-bin/get_jsapi_ticket` | |
| 2 | GetAgentConfigTicket | GET | `/cgi-bin/ticket/get?type=agent_config` | |
| 3 | SignJSAPI | — | — | Pure computation, package-level func |
| 4 | GetExternalContact | GET | `/cgi-bin/externalcontact/get?external_userid=xxx` | |
| 5 | ListExternalContact | GET | `/cgi-bin/externalcontact/list?userid=xxx` | |
| 6 | BatchGetExternalContactByUser | POST | `/cgi-bin/externalcontact/batch/get_by_user` | Paginated |
| 7 | RemarkExternalContact | POST | `/cgi-bin/externalcontact/remark` | |
| 8 | GetCorpTagList | POST | `/cgi-bin/externalcontact/get_corp_tag_list` | |
| 9 | AddCorpTag | POST | `/cgi-bin/externalcontact/add_corp_tag` | |
| 10 | EditCorpTag | POST | `/cgi-bin/externalcontact/edit_corp_tag` | |
| 11 | DelCorpTag | POST | `/cgi-bin/externalcontact/del_corp_tag` | |
| 12 | MarkTag | POST | `/cgi-bin/externalcontact/mark_tag` | |
| 13 | GetFollowUserList | GET | `/cgi-bin/externalcontact/get_follow_user_list` | |
| 14 | GetApprovalTemplate | POST | `/cgi-bin/oa/gettemplatedetail` | |
| 15 | ApplyEvent | POST | `/cgi-bin/oa/applyevent` | |
| 16 | GetApprovalDetail | POST | `/cgi-bin/oa/getapprovaldetail` | |
| 17 | GetApprovalData | POST | `/cgi-bin/oa/getapprovalinfo` | |
| 18 | GetCheckinData | POST | `/cgi-bin/checkin/getcheckindata` | |
| 19 | GetCheckinOption | POST | `/cgi-bin/checkin/getcheckinoption` | |
| 20 | GetCheckinDayData | POST | `/cgi-bin/checkin/getcheckin_daydata` | |
| 21 | GetCheckinMonthData | POST | `/cgi-bin/checkin/getcheckin_monthdata` | |
| 22 | CreateCalendar | POST | `/cgi-bin/oa/calendar/add` | |
| 23 | UpdateCalendar | POST | `/cgi-bin/oa/calendar/update` | |
| 24 | GetCalendar | POST | `/cgi-bin/oa/calendar/get` | |
| 25 | DeleteCalendar | POST | `/cgi-bin/oa/calendar/del` | |
| 26 | CreateSchedule | POST | `/cgi-bin/oa/schedule/add` | |
| 27 | UpdateSchedule | POST | `/cgi-bin/oa/schedule/update` | |
| 28 | GetSchedule | POST | `/cgi-bin/oa/schedule/get` | |
| 29 | DeleteSchedule | POST | `/cgi-bin/oa/schedule/del` | |
| 30 | GetScheduleByCalendar | POST | `/cgi-bin/oa/schedule/get_by_calendar` | |

### 1.2 File Summary

- 5 new DTO files
- 5 new implementation files
- 5 new test files
- 30 public methods (29 on `*CorpClient` + 1 package-level)
- No new private helpers needed (all use existing `doGet`/`doPost`)

---

## 2. Architecture

### 2.1 Method Placement

All 29 methods are on `*CorpClient`, using `doGet` for GET requests and `doPost` for POST requests. `SignJSAPI` is a package-level function (no receiver) since it's pure computation.

### 2.2 No New Helpers

All JSON APIs fit existing `doGet(ctx, path, extra, out)` and `doPost(ctx, path, body, out)`. No `doPostExtra`, `doUpload`, or other helpers needed.

### 2.3 File Layout

| File | Content |
|---|---|
| `struct.jssdk.go` (new) | JSAPITicketResp |
| `corp.jssdk.go` (new) | GetJSAPITicket, GetAgentConfigTicket, SignJSAPI |
| `corp.jssdk_test.go` (new) | 3 tests |
| `struct.external_contact.go` (new) | ~15 DTOs for external contact and corp tags |
| `corp.external_contact.go` (new) | 10 external contact methods |
| `corp.external_contact_test.go` (new) | 10 tests |
| `struct.approval.go` (new) | ~20 DTOs for approval templates, events, details |
| `corp.approval.go` (new) | 4 approval methods |
| `corp.approval_test.go` (new) | 4 tests |
| `struct.checkin.go` (new) | ~12 DTOs for checkin data, options, reports |
| `corp.checkin.go` (new) | 4 checkin methods |
| `corp.checkin_test.go` (new) | 4 tests |
| `struct.calendar.go` (new) | ~15 DTOs for calendar and schedule |
| `corp.calendar.go` (new) | 9 calendar/schedule methods |
| `corp.calendar_test.go` (new) | 9 tests |

---

## 3. DTOs

### 3.1 JS-SDK

```go
// JSAPITicketResp is the response from GetJSAPITicket / GetAgentConfigTicket.
type JSAPITicketResp struct {
    Ticket    string `json:"ticket"`
    ExpiresIn int    `json:"expires_in"`
}
```

### 3.2 External Contact

```go
// ExternalContact 客户详情。
type ExternalContact struct {
    ExternalUserID string `json:"external_userid"`
    Name           string `json:"name"`
    Position       string `json:"position"`
    Avatar         string `json:"avatar"`
    CorpName       string `json:"corp_name"`
    CorpFullName   string `json:"corp_full_name"`
    Type           int    `json:"type"`
    Gender         int    `json:"gender"`
    UnionID        string `json:"unionid"`
}

// FollowUser 跟进人信息。
type FollowUser struct {
    UserID      string      `json:"userid"`
    Remark      string      `json:"remark"`
    Description string      `json:"description"`
    CreateTime  int64       `json:"createtime"`
    State       string      `json:"state"`
    Tags        []FollowTag `json:"tags"`
}

// FollowTag 跟进人给客户打的标签。
type FollowTag struct {
    GroupName string `json:"group_name"`
    TagName   string `json:"tag_name"`
    Type      int    `json:"type"`
}

// GetExternalContactResp 获取客户详情响应。
type GetExternalContactResp struct {
    ExternalContact ExternalContact `json:"external_contact"`
    FollowUser      []FollowUser   `json:"follow_user"`
}

// ListExternalContactResp 获取客户列表响应。
type ListExternalContactResp struct {
    ExternalUserID []string `json:"external_userid"`
}

// BatchGetExternalContactReq 批量获取客户详情请求。
type BatchGetExternalContactReq struct {
    UserIDList []string `json:"userid_list"`
    Cursor     string   `json:"cursor,omitempty"`
    Limit      int      `json:"limit,omitempty"`
}

// BatchGetExternalContactResp 批量获取客户详情响应。
type BatchGetExternalContactResp struct {
    ExternalContactList []GetExternalContactResp `json:"external_contact_list"`
    NextCursor          string                   `json:"next_cursor"`
}

// RemarkExternalContactReq 修改客户备注请求。
type RemarkExternalContactReq struct {
    UserID           string   `json:"userid"`
    ExternalUserID   string   `json:"external_userid"`
    Remark           string   `json:"remark,omitempty"`
    Description      string   `json:"description,omitempty"`
    RemarkCompany    string   `json:"remark_company,omitempty"`
    RemarkMobiles    []string `json:"remark_mobiles,omitempty"`
    RemarkPicMediaID string   `json:"remark_pic_mediaid,omitempty"`
}

// CorpTag 企业客户标签。
type CorpTag struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Order int    `json:"order"`
}

// CorpTagGroup 企业客户标签组。
type CorpTagGroup struct {
    GroupID   string    `json:"group_id"`
    GroupName string    `json:"group_name"`
    Tag       []CorpTag `json:"tag"`
}

// GetCorpTagListReq 获取标签库请求。
type GetCorpTagListReq struct {
    TagID   []string `json:"tag_id,omitempty"`
    GroupID []string `json:"group_id,omitempty"`
}

// GetCorpTagListResp 获取标签库响应。
type GetCorpTagListResp struct {
    TagGroup []CorpTagGroup `json:"tag_group"`
}

// AddCorpTagReq 添加企业客户标签请求。
type AddCorpTagReq struct {
    GroupID   string `json:"group_id,omitempty"`
    GroupName string `json:"group_name,omitempty"`
    Tag       []struct {
        Name  string `json:"name"`
        Order int    `json:"order,omitempty"`
    } `json:"tag"`
}

// AddCorpTagResp 添加企业客户标签响应。
type AddCorpTagResp struct {
    TagGroup CorpTagGroup `json:"tag_group"`
}

// EditCorpTagReq 编辑企业客户标签请求。
type EditCorpTagReq struct {
    ID    string `json:"id"`
    Name  string `json:"name,omitempty"`
    Order *int   `json:"order,omitempty"`
}

// DelCorpTagReq 删除企业客户标签请求。
type DelCorpTagReq struct {
    TagID   []string `json:"tag_id,omitempty"`
    GroupID []string `json:"group_id,omitempty"`
}

// MarkTagReq 编辑客户企业标签请求。
type MarkTagReq struct {
    UserID         string   `json:"userid"`
    ExternalUserID string   `json:"external_userid"`
    AddTag         []string `json:"add_tag,omitempty"`
    RemoveTag      []string `json:"remove_tag,omitempty"`
}

// FollowUserListResp 获取配置了客户联系功能的成员列表响应。
type FollowUserListResp struct {
    FollowUser []string `json:"follow_user"`
}
```

### 3.3 Approval

```go
// GetApprovalTemplateReq 获取审批模板详情请求。
type GetApprovalTemplateReq struct {
    TemplateID string `json:"template_id"`
}

// ApprovalTemplateResp 审批模板详情响应。
type ApprovalTemplateResp struct {
    TemplateNames   []ApprovalText  `json:"template_names"`
    TemplateContent ApprovalContent `json:"template_content"`
}

// ApprovalText 多语言文本。
type ApprovalText struct {
    Text string `json:"text"`
    Lang string `json:"lang"`
}

// ApprovalContent 审批模板内容。
type ApprovalContent struct {
    Controls []ApprovalControl `json:"controls"`
}

// ApprovalControl 审批模板控件。
type ApprovalControl struct {
    Property ApprovalControlProperty `json:"property"`
    Config   ApprovalControlConfig   `json:"config,omitempty"`
}

// ApprovalControlProperty 控件属性。
type ApprovalControlProperty struct {
    Control string         `json:"control"`
    ID      string         `json:"id"`
    Title   []ApprovalText `json:"title"`
}

// ApprovalControlConfig 控件配置。
type ApprovalControlConfig struct {
    Date     *ApprovalDateConfig     `json:"date,omitempty"`
    Selector *ApprovalSelectorConfig `json:"selector,omitempty"`
}

// ApprovalDateConfig 日期控件配置。
type ApprovalDateConfig struct {
    Type string `json:"type"`
}

// ApprovalSelectorConfig 选择控件配置。
type ApprovalSelectorConfig struct {
    Type    string           `json:"type"`
    Options []ApprovalOption `json:"options"`
}

// ApprovalOption 选择控件选项。
type ApprovalOption struct {
    Key   string         `json:"key"`
    Value []ApprovalText `json:"value"`
}

// ApplyEventReq 提交审批申请请求。
type ApplyEventReq struct {
    CreatorUserID       string         `json:"creator_userid"`
    TemplateID          string         `json:"template_id"`
    UseTemplateApprover int            `json:"use_template_approver"`
    ApplyData           ApplyData      `json:"apply_data"`
    SummaryList         []ApplySummary `json:"summary_list"`
}

// ApplyData 审批申请数据。
type ApplyData struct {
    Contents []ApplyContent `json:"contents"`
}

// ApplyContent 审批申请控件值。
type ApplyContent struct {
    Control string     `json:"control"`
    ID      string     `json:"id"`
    Value   ApplyValue `json:"value"`
}

// ApplyValue 控件值（各类型共用，按需填充）。
type ApplyValue struct {
    Text     string              `json:"text,omitempty"`
    Date     *ApplyDateValue     `json:"date,omitempty"`
    Selector *ApplySelectorValue `json:"selector,omitempty"`
}

// ApplyDateValue 日期控件值。
type ApplyDateValue struct {
    Type      string `json:"type"`
    Timestamp string `json:"s_timestamp"`
}

// ApplySelectorValue 选择控件值。
type ApplySelectorValue struct {
    Type    string             `json:"type"`
    Options []ApplySelectorOpt `json:"options"`
}

// ApplySelectorOpt 选择控件选中项。
type ApplySelectorOpt struct {
    Key string `json:"key"`
}

// ApplySummary 审批摘要。
type ApplySummary struct {
    SummaryInfo []ApprovalText `json:"summary_info"`
}

// ApplyEventResp 提交审批申请响应。
type ApplyEventResp struct {
    SpNo string `json:"sp_no"`
}

// GetApprovalDetailReq 获取审批申请详情请求。
type GetApprovalDetailReq struct {
    SpNo string `json:"sp_no"`
}

// ApprovalDetailResp 审批申请详情响应。
type ApprovalDetailResp struct {
    Info ApprovalInfoDetail `json:"info"`
}

// ApprovalInfoDetail 审批单详情。
type ApprovalInfoDetail struct {
    SpNo       string           `json:"sp_no"`
    SpName     string           `json:"sp_name"`
    SpStatus   int              `json:"sp_status"`
    TemplateID string           `json:"template_id"`
    ApplyTime  int64            `json:"apply_time"`
    Applyer    ApprovalApplyer  `json:"applyer"`
    SpRecord   []ApprovalRecord `json:"sp_record"`
    ApplyData  ApplyData        `json:"apply_data"`
}

// ApprovalApplyer 申请人信息。
type ApprovalApplyer struct {
    UserID  string `json:"userid"`
    PartyID string `json:"partyid"`
}

// ApprovalRecord 审批节点记录。
type ApprovalRecord struct {
    SpStatus     int              `json:"sp_status"`
    ApproverAttr int              `json:"approverattr"`
    Details      []ApprovalDetail `json:"details"`
}

// ApprovalDetail 审批人详情。
type ApprovalDetail struct {
    Approver ApprovalApprover `json:"approver"`
    Speech   string           `json:"speech"`
    SpStatus int              `json:"sp_status"`
    SpTime   int64            `json:"sptime"`
}

// ApprovalApprover 审批人。
type ApprovalApprover struct {
    UserID string `json:"userid"`
}

// GetApprovalDataReq 批量获取审批单号请求。
type GetApprovalDataReq struct {
    StartTime int64            `json:"starttime"`
    EndTime   int64            `json:"endtime"`
    Cursor    int              `json:"cursor,omitempty"`
    Size      int              `json:"size,omitempty"`
    Filters   []ApprovalFilter `json:"filters,omitempty"`
}

// ApprovalFilter 审批单号筛选条件。
type ApprovalFilter struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

// GetApprovalDataResp 批量获取审批单号响应。
type GetApprovalDataResp struct {
    SpNoList      []string `json:"sp_no_list"`
    NewNextCursor int      `json:"new_next_cursor"`
}
```

### 3.4 Checkin

```go
// GetCheckinDataReq 获取打卡记录请求。
type GetCheckinDataReq struct {
    OpenCheckinDataType int      `json:"opencheckindatatype"`
    StartTime           int64    `json:"starttime"`
    EndTime             int64    `json:"endtime"`
    UserIDList          []string `json:"useridlist"`
}

// CheckinData 打卡记录。
type CheckinData struct {
    UserID         string `json:"userid"`
    GroupName      string `json:"groupname"`
    CheckinType    string `json:"checkin_type"`
    CheckinTime    int64  `json:"checkin_time"`
    ExceptionType  string `json:"exception_type"`
    LocationTitle  string `json:"location_title"`
    LocationDetail string `json:"location_detail"`
    Notes          string `json:"notes"`
}

// GetCheckinDataResp 获取打卡记录响应。
type GetCheckinDataResp struct {
    CheckinData []CheckinData `json:"checkindata"`
}

// GetCheckinOptionReq 获取打卡规则请求。
type GetCheckinOptionReq struct {
    DateTime   int64    `json:"datetime"`
    UserIDList []string `json:"useridlist"`
}

// CheckinOption 打卡规则。
type CheckinOption struct {
    UserID string       `json:"userid"`
    Group  CheckinGroup `json:"group"`
}

// CheckinGroup 打卡规则组。
type CheckinGroup struct {
    GroupID   int    `json:"groupid"`
    GroupName string `json:"groupname"`
    GroupType int    `json:"grouptype"`
}

// GetCheckinOptionResp 获取打卡规则响应。
type GetCheckinOptionResp struct {
    Info []CheckinOption `json:"info"`
}

// GetCheckinDayDataReq 获取打卡日报请求。
type GetCheckinDayDataReq struct {
    StartTime  int64    `json:"starttime"`
    EndTime    int64    `json:"endtime"`
    UserIDList []string `json:"useridlist"`
}

// CheckinDayData 打卡日报数据。
type CheckinDayData struct {
    BaseInfo    CheckinDayBase    `json:"base_info"`
    SummaryInfo CheckinDaySummary `json:"summary_info"`
}

// CheckinDayBase 日报基础信息。
type CheckinDayBase struct {
    Date   int64  `json:"date"`
    Name   string `json:"name"`
    NameEx string `json:"name_ex"`
    AcctID string `json:"acctid"`
}

// CheckinDaySummary 日报统计。
type CheckinDaySummary struct {
    CheckinCount    int `json:"checkin_count"`
    RegularWorkSec  int `json:"regular_work_sec"`
    StandardWorkSec int `json:"standard_work_sec"`
}

// GetCheckinDayDataResp 获取打卡日报响应。
type GetCheckinDayDataResp struct {
    Datas []CheckinDayData `json:"datas"`
}

// GetCheckinMonthDataReq 获取打卡月报请求。
type GetCheckinMonthDataReq struct {
    StartTime  int64    `json:"starttime"`
    EndTime    int64    `json:"endtime"`
    UserIDList []string `json:"useridlist"`
}

// CheckinMonthData 打卡月报数据。
type CheckinMonthData struct {
    BaseInfo    CheckinDayBase      `json:"base_info"`
    SummaryInfo CheckinMonthSummary `json:"summary_info"`
}

// CheckinMonthSummary 月报统计。
type CheckinMonthSummary struct {
    WorkDays       int `json:"work_days"`
    RegularWorkSec int `json:"regular_work_sec"`
    ExceptDays     int `json:"except_days"`
}

// GetCheckinMonthDataResp 获取打卡月报响应。
type GetCheckinMonthDataResp struct {
    Datas []CheckinMonthData `json:"datas"`
}
```

### 3.5 Calendar & Schedule

```go
// Calendar 日历对象。
type Calendar struct {
    CalID       string          `json:"cal_id,omitempty"`
    Organizer   string          `json:"organizer"`
    Summary     string          `json:"summary"`
    Color       string          `json:"color,omitempty"`
    Description string          `json:"description,omitempty"`
    Shares      []CalendarShare `json:"shares,omitempty"`
}

// CalendarShare 日历共享对象。
type CalendarShare struct {
    UserID string `json:"userid"`
}

// CreateCalendarReq 创建日历请求。
type CreateCalendarReq struct {
    Calendar Calendar `json:"calendar"`
}

// CreateCalendarResp 创建日历响应。
type CreateCalendarResp struct {
    CalID string `json:"cal_id"`
}

// UpdateCalendarReq 更新日历请求。
type UpdateCalendarReq struct {
    Calendar Calendar `json:"calendar"`
}

// GetCalendarReq 获取日历详情请求。
type GetCalendarReq struct {
    CalIDList []string `json:"cal_id_list"`
}

// GetCalendarResp 获取日历详情响应。
type GetCalendarResp struct {
    CalendarList []Calendar `json:"calendar_list"`
}

// DeleteCalendarReq 删除日历请求。
type DeleteCalendarReq struct {
    CalID string `json:"cal_id"`
}

// Schedule 日程对象。
type Schedule struct {
    ScheduleID  string             `json:"schedule_id,omitempty"`
    Organizer   string             `json:"organizer"`
    Summary     string             `json:"summary"`
    Description string             `json:"description,omitempty"`
    StartTime   int64              `json:"start_time"`
    EndTime     int64              `json:"end_time"`
    Location    string             `json:"location,omitempty"`
    CalID       string             `json:"cal_id,omitempty"`
    Attendees   []ScheduleAttendee `json:"attendees,omitempty"`
    Reminders   *ScheduleReminder  `json:"reminders,omitempty"`
}

// ScheduleAttendee 日程参与人。
type ScheduleAttendee struct {
    UserID string `json:"userid"`
}

// ScheduleReminder 日程提醒。
type ScheduleReminder struct {
    IsRemind     int `json:"is_remind"`
    RemindBefore int `json:"remind_before_event_secs,omitempty"`
}

// CreateScheduleReq 创建日程请求。
type CreateScheduleReq struct {
    Schedule Schedule `json:"schedule"`
}

// CreateScheduleResp 创建日程响应。
type CreateScheduleResp struct {
    ScheduleID string `json:"schedule_id"`
}

// UpdateScheduleReq 更新日程请求。
type UpdateScheduleReq struct {
    Schedule Schedule `json:"schedule"`
}

// GetScheduleReq 获取日程详情请求。
type GetScheduleReq struct {
    ScheduleIDList []string `json:"schedule_id_list"`
}

// GetScheduleResp 获取日程详情响应。
type GetScheduleResp struct {
    ScheduleList []Schedule `json:"schedule_list"`
}

// DeleteScheduleReq 删除日程请求。
type DeleteScheduleReq struct {
    ScheduleID string `json:"schedule_id"`
}

// GetScheduleByCalendarReq 获取日历下日程列表请求。
type GetScheduleByCalendarReq struct {
    CalID  string `json:"cal_id"`
    Offset int    `json:"offset,omitempty"`
    Limit  int    `json:"limit,omitempty"`
}

// GetScheduleByCalendarResp 获取日历下日程列表响应。
type GetScheduleByCalendarResp struct {
    ScheduleList []Schedule `json:"schedule_list"`
}
```

---

## 4. Method Implementations

### 4.1 JS-SDK (`corp.jssdk.go`)

```go
package isv

import (
    "context"
    "crypto/sha1"
    "encoding/hex"
    "fmt"
    "net/url"
)

// GetJSAPITicket 获取企业的 jsapi_ticket（用于 wx.config 签名）。
func (cc *CorpClient) GetJSAPITicket(ctx context.Context) (*JSAPITicketResp, error) {
    var resp JSAPITicketResp
    if err := cc.doGet(ctx, "/cgi-bin/get_jsapi_ticket", nil, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// GetAgentConfigTicket 获取应用的 jsapi_ticket（用于 wx.agentConfig 签名）。
func (cc *CorpClient) GetAgentConfigTicket(ctx context.Context) (*JSAPITicketResp, error) {
    extra := url.Values{"type": {"agent_config"}}
    var resp JSAPITicketResp
    if err := cc.doGet(ctx, "/cgi-bin/ticket/get", extra, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// SignJSAPI 计算 JS-SDK 签名。纯计算函数，不发网络请求。
func SignJSAPI(ticket, nonceStr, timestamp, url string) string {
    s := fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%s&url=%s",
        ticket, nonceStr, timestamp, url)
    h := sha1.New()
    h.Write([]byte(s))
    return hex.EncodeToString(h.Sum(nil))
}
```

### 4.2 External Contact (`corp.external_contact.go`)

```go
package isv

import (
    "context"
    "net/url"
)

// GetExternalContact 获取客户详情。
func (cc *CorpClient) GetExternalContact(ctx context.Context, externalUserID string) (*GetExternalContactResp, error) {
    extra := url.Values{"external_userid": {externalUserID}}
    var resp GetExternalContactResp
    if err := cc.doGet(ctx, "/cgi-bin/externalcontact/get", extra, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// ListExternalContact 获取客户列表（按跟进人）。
func (cc *CorpClient) ListExternalContact(ctx context.Context, userID string) (*ListExternalContactResp, error) {
    extra := url.Values{"userid": {userID}}
    var resp ListExternalContactResp
    if err := cc.doGet(ctx, "/cgi-bin/externalcontact/list", extra, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// BatchGetExternalContactByUser 批量获取客户详情。
func (cc *CorpClient) BatchGetExternalContactByUser(ctx context.Context, req *BatchGetExternalContactReq) (*BatchGetExternalContactResp, error) {
    var resp BatchGetExternalContactResp
    if err := cc.doPost(ctx, "/cgi-bin/externalcontact/batch/get_by_user", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// RemarkExternalContact 修改客户备注信息。
func (cc *CorpClient) RemarkExternalContact(ctx context.Context, req *RemarkExternalContactReq) error {
    return cc.doPost(ctx, "/cgi-bin/externalcontact/remark", req, nil)
}

// GetCorpTagList 获取企业标签库。
func (cc *CorpClient) GetCorpTagList(ctx context.Context, req *GetCorpTagListReq) (*GetCorpTagListResp, error) {
    var resp GetCorpTagListResp
    if err := cc.doPost(ctx, "/cgi-bin/externalcontact/get_corp_tag_list", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// AddCorpTag 添加企业客户标签。
func (cc *CorpClient) AddCorpTag(ctx context.Context, req *AddCorpTagReq) (*AddCorpTagResp, error) {
    var resp AddCorpTagResp
    if err := cc.doPost(ctx, "/cgi-bin/externalcontact/add_corp_tag", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// EditCorpTag 编辑企业客户标签。
func (cc *CorpClient) EditCorpTag(ctx context.Context, req *EditCorpTagReq) error {
    return cc.doPost(ctx, "/cgi-bin/externalcontact/edit_corp_tag", req, nil)
}

// DelCorpTag 删除企业客户标签。
func (cc *CorpClient) DelCorpTag(ctx context.Context, req *DelCorpTagReq) error {
    return cc.doPost(ctx, "/cgi-bin/externalcontact/del_corp_tag", req, nil)
}

// MarkTag 编辑客户企业标签（给客户打/取消标签）。
func (cc *CorpClient) MarkTag(ctx context.Context, req *MarkTagReq) error {
    return cc.doPost(ctx, "/cgi-bin/externalcontact/mark_tag", req, nil)
}

// GetFollowUserList 获取配置了客户联系功能的成员列表。
func (cc *CorpClient) GetFollowUserList(ctx context.Context) (*FollowUserListResp, error) {
    var resp FollowUserListResp
    if err := cc.doGet(ctx, "/cgi-bin/externalcontact/get_follow_user_list", nil, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.3 Approval (`corp.approval.go`)

```go
package isv

import "context"

// GetApprovalTemplate 获取审批模板详情。
func (cc *CorpClient) GetApprovalTemplate(ctx context.Context, templateID string) (*ApprovalTemplateResp, error) {
    body := &GetApprovalTemplateReq{TemplateID: templateID}
    var resp ApprovalTemplateResp
    if err := cc.doPost(ctx, "/cgi-bin/oa/gettemplatedetail", body, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// ApplyEvent 提交审批申请。
func (cc *CorpClient) ApplyEvent(ctx context.Context, req *ApplyEventReq) (*ApplyEventResp, error) {
    var resp ApplyEventResp
    if err := cc.doPost(ctx, "/cgi-bin/oa/applyevent", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// GetApprovalDetail 获取审批申请详情。
func (cc *CorpClient) GetApprovalDetail(ctx context.Context, spNo string) (*ApprovalDetailResp, error) {
    body := &GetApprovalDetailReq{SpNo: spNo}
    var resp ApprovalDetailResp
    if err := cc.doPost(ctx, "/cgi-bin/oa/getapprovaldetail", body, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// GetApprovalData 批量获取审批单号。
func (cc *CorpClient) GetApprovalData(ctx context.Context, req *GetApprovalDataReq) (*GetApprovalDataResp, error) {
    var resp GetApprovalDataResp
    if err := cc.doPost(ctx, "/cgi-bin/oa/getapprovalinfo", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.4 Checkin (`corp.checkin.go`)

```go
package isv

import "context"

// GetCheckinData 获取打卡记录数据。
func (cc *CorpClient) GetCheckinData(ctx context.Context, req *GetCheckinDataReq) (*GetCheckinDataResp, error) {
    var resp GetCheckinDataResp
    if err := cc.doPost(ctx, "/cgi-bin/checkin/getcheckindata", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// GetCheckinOption 获取打卡规则。
func (cc *CorpClient) GetCheckinOption(ctx context.Context, req *GetCheckinOptionReq) (*GetCheckinOptionResp, error) {
    var resp GetCheckinOptionResp
    if err := cc.doPost(ctx, "/cgi-bin/checkin/getcheckinoption", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// GetCheckinDayData 获取打卡日报数据。
func (cc *CorpClient) GetCheckinDayData(ctx context.Context, req *GetCheckinDayDataReq) (*GetCheckinDayDataResp, error) {
    var resp GetCheckinDayDataResp
    if err := cc.doPost(ctx, "/cgi-bin/checkin/getcheckin_daydata", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// GetCheckinMonthData 获取打卡月报数据。
func (cc *CorpClient) GetCheckinMonthData(ctx context.Context, req *GetCheckinMonthDataReq) (*GetCheckinMonthDataResp, error) {
    var resp GetCheckinMonthDataResp
    if err := cc.doPost(ctx, "/cgi-bin/checkin/getcheckin_monthdata", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.5 Calendar & Schedule (`corp.calendar.go`)

```go
package isv

import "context"

// CreateCalendar 创建日历。
func (cc *CorpClient) CreateCalendar(ctx context.Context, req *CreateCalendarReq) (*CreateCalendarResp, error) {
    var resp CreateCalendarResp
    if err := cc.doPost(ctx, "/cgi-bin/oa/calendar/add", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// UpdateCalendar 更新日历。
func (cc *CorpClient) UpdateCalendar(ctx context.Context, req *UpdateCalendarReq) error {
    return cc.doPost(ctx, "/cgi-bin/oa/calendar/update", req, nil)
}

// GetCalendar 获取日历详情。
func (cc *CorpClient) GetCalendar(ctx context.Context, calIDs []string) (*GetCalendarResp, error) {
    body := &GetCalendarReq{CalIDList: calIDs}
    var resp GetCalendarResp
    if err := cc.doPost(ctx, "/cgi-bin/oa/calendar/get", body, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// DeleteCalendar 删除日历。
func (cc *CorpClient) DeleteCalendar(ctx context.Context, calID string) error {
    body := &DeleteCalendarReq{CalID: calID}
    return cc.doPost(ctx, "/cgi-bin/oa/calendar/del", body, nil)
}

// CreateSchedule 创建日程。
func (cc *CorpClient) CreateSchedule(ctx context.Context, req *CreateScheduleReq) (*CreateScheduleResp, error) {
    var resp CreateScheduleResp
    if err := cc.doPost(ctx, "/cgi-bin/oa/schedule/add", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// UpdateSchedule 更新日程。
func (cc *CorpClient) UpdateSchedule(ctx context.Context, req *UpdateScheduleReq) error {
    return cc.doPost(ctx, "/cgi-bin/oa/schedule/update", req, nil)
}

// GetSchedule 获取日程详情。
func (cc *CorpClient) GetSchedule(ctx context.Context, scheduleIDs []string) (*GetScheduleResp, error) {
    body := &GetScheduleReq{ScheduleIDList: scheduleIDs}
    var resp GetScheduleResp
    if err := cc.doPost(ctx, "/cgi-bin/oa/schedule/get", body, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

// DeleteSchedule 删除日程。
func (cc *CorpClient) DeleteSchedule(ctx context.Context, scheduleID string) error {
    body := &DeleteScheduleReq{ScheduleID: scheduleID}
    return cc.doPost(ctx, "/cgi-bin/oa/schedule/del", body, nil)
}

// GetScheduleByCalendar 获取日历下的日程列表。
func (cc *CorpClient) GetScheduleByCalendar(ctx context.Context, req *GetScheduleByCalendarReq) (*GetScheduleByCalendarResp, error) {
    var resp GetScheduleByCalendarResp
    if err := cc.doPost(ctx, "/cgi-bin/oa/schedule/get_by_calendar", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

---

## 5. Testing

### 5.1 Test Matrix

| # | Test Function | File | Validates |
|---|---|---|---|
| 1 | TestGetJSAPITicket | corp.jssdk_test.go | GET path, access_token, response parsing |
| 2 | TestGetAgentConfigTicket | corp.jssdk_test.go | GET path, type=agent_config query, response parsing |
| 3 | TestSignJSAPI | corp.jssdk_test.go | Known input/output pair, SHA1 correctness |
| 4 | TestGetExternalContact | corp.external_contact_test.go | GET path, external_userid query, response parsing |
| 5 | TestListExternalContact | corp.external_contact_test.go | GET path, userid query, response parsing |
| 6 | TestBatchGetExternalContactByUser | corp.external_contact_test.go | POST path, request body, paginated response |
| 7 | TestRemarkExternalContact | corp.external_contact_test.go | POST path, request body, no-error success |
| 8 | TestGetCorpTagList | corp.external_contact_test.go | POST path, tag group response parsing |
| 9 | TestAddCorpTag | corp.external_contact_test.go | POST path, request body, tag group response |
| 10 | TestEditCorpTag | corp.external_contact_test.go | POST path, *int order serialization |
| 11 | TestDelCorpTag | corp.external_contact_test.go | POST path, request body, no-error success |
| 12 | TestMarkTag | corp.external_contact_test.go | POST path, add/remove tag arrays |
| 13 | TestGetFollowUserList | corp.external_contact_test.go | GET path, string slice response |
| 14 | TestGetApprovalTemplate | corp.approval_test.go | POST path, nested template response |
| 15 | TestApplyEvent | corp.approval_test.go | POST path, deep nested request body, sp_no response |
| 16 | TestGetApprovalDetail | corp.approval_test.go | POST path, approval detail response parsing |
| 17 | TestGetApprovalData | corp.approval_test.go | POST path, sp_no_list response |
| 18 | TestGetCheckinData | corp.checkin_test.go | POST path, checkin records response |
| 19 | TestGetCheckinOption | corp.checkin_test.go | POST path, checkin rules response |
| 20 | TestGetCheckinDayData | corp.checkin_test.go | POST path, day report response |
| 21 | TestGetCheckinMonthData | corp.checkin_test.go | POST path, month report response |
| 22 | TestCreateCalendar | corp.calendar_test.go | POST path, cal_id response |
| 23 | TestUpdateCalendar | corp.calendar_test.go | POST path, no-error success |
| 24 | TestGetCalendar | corp.calendar_test.go | POST path, calendar_list response |
| 25 | TestDeleteCalendar | corp.calendar_test.go | POST path, no-error success |
| 26 | TestCreateSchedule | corp.calendar_test.go | POST path, schedule_id response |
| 27 | TestUpdateSchedule | corp.calendar_test.go | POST path, no-error success |
| 28 | TestGetSchedule | corp.calendar_test.go | POST path, schedule_list response |
| 29 | TestDeleteSchedule | corp.calendar_test.go | POST path, no-error success |
| 30 | TestGetScheduleByCalendar | corp.calendar_test.go | POST path, paginated schedule_list response |

### 5.2 Test Patterns

- All tests use `newTestCorpClient(t, srv.URL)` with pre-seeded token "CTOK"
- GET tests: verify path + query params, return JSON fixture, assert response fields
- POST tests: verify path + access_token query, decode request body, return JSON fixture, assert response fields
- SignJSAPI: pure function test with known WeChat documentation example values
- No-response methods (Remark, Edit, Del, Update, Delete): verify no error on `{"errcode":0,"errmsg":"ok"}`

### 5.3 Coverage Target

≥ 85% for the package after changes.

---

## 6. Implementation Order

1. Create `struct.jssdk.go` + `corp.jssdk.go` + tests (3 methods, warmup)
2. Create `struct.external_contact.go` + `corp.external_contact.go` + tests (10 methods, core)
3. Create `struct.approval.go` + `corp.approval.go` + tests (4 methods)
4. Create `struct.checkin.go` + `corp.checkin.go` + tests (4 methods)
5. Create `struct.calendar.go` + `corp.calendar.go` + tests (9 methods, finish)
