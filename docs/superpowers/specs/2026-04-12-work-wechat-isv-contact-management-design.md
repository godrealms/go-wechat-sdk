# work-wechat ISV 子项目 3:代企业通讯录管理

**Date:** 2026-04-12
**Status:** Draft
**Scope:** 企业微信 work-wechat/isv 包 —— 部门 CRUD + 成员 CRUD + 标签管理 + 邀请成员
**Depends on:** 子项目 1(ISV 认证底座)、子项目 4(CorpClient HTTP helpers)

---

## 1. 目标

为 `*CorpClient` 增加 **18 个公开方法**,覆盖企业微信通讯录管理全量 API:

- 部门管理 4 个
- 成员管理 6 个
- 标签管理 6 个
- 邀请成员 1 个
- 辅助函数 1 个(`boolToStr` 内部)

### 1.1 必须交付的 18 个公开方法

#### 部门管理(4 个)

| # | 签名 | 接口 | HTTP |
|---|---|---|---|
| 1 | `(cc *CorpClient) CreateDepartment(ctx, req *CreateDeptReq) (*CreateDeptResp, error)` | `/cgi-bin/department/create` | POST |
| 2 | `(cc *CorpClient) UpdateDepartment(ctx, req *UpdateDeptReq) error` | `/cgi-bin/department/update` | POST |
| 3 | `(cc *CorpClient) DeleteDepartment(ctx, id int) error` | `/cgi-bin/department/delete?id=` | GET |
| 4 | `(cc *CorpClient) ListDepartment(ctx, id int) ([]Department, error)` | `/cgi-bin/department/list?id=` | GET |

#### 成员管理(6 个)

| # | 签名 | 接口 | HTTP |
|---|---|---|---|
| 5 | `(cc *CorpClient) CreateUser(ctx, req *CreateUserReq) error` | `/cgi-bin/user/create` | POST |
| 6 | `(cc *CorpClient) UpdateUser(ctx, req *UpdateUserReq) error` | `/cgi-bin/user/update` | POST |
| 7 | `(cc *CorpClient) DeleteUser(ctx, userID string) error` | `/cgi-bin/user/delete?userid=` | GET |
| 8 | `(cc *CorpClient) GetUser(ctx, userID string) (*UserDetail, error)` | `/cgi-bin/user/get?userid=` | GET |
| 9 | `(cc *CorpClient) ListUserSimple(ctx, deptID int, fetchChild bool) (*UserSimpleListResp, error)` | `/cgi-bin/user/simplelist?department_id=&fetch_child=` | GET |
| 10 | `(cc *CorpClient) ListUserDetail(ctx, deptID int, fetchChild bool) (*UserDetailListResp, error)` | `/cgi-bin/user/list?department_id=&fetch_child=` | GET |

#### 标签管理(6 个)

| # | 签名 | 接口 | HTTP |
|---|---|---|---|
| 11 | `(cc *CorpClient) CreateTag(ctx, req *CreateTagReq) (*CreateTagResp, error)` | `/cgi-bin/tag/create` | POST |
| 12 | `(cc *CorpClient) UpdateTag(ctx, req *UpdateTagReq) error` | `/cgi-bin/tag/update` | POST |
| 13 | `(cc *CorpClient) DeleteTag(ctx, tagID int) error` | `/cgi-bin/tag/delete?tagid=` | GET |
| 14 | `(cc *CorpClient) ListTag(ctx) ([]Tag, error)` | `/cgi-bin/tag/list` | GET |
| 15 | `(cc *CorpClient) GetTagUsers(ctx, tagID int) (*TagUsersResp, error)` | `/cgi-bin/tag/get?tagid=` | GET |
| 16 | `(cc *CorpClient) AddTagUsers(ctx, req *TagUsersModifyReq) (*TagUsersModifyResp, error)` | `/cgi-bin/tag/addtagusers` | POST |
| 17 | `(cc *CorpClient) DelTagUsers(ctx, req *TagUsersModifyReq) (*TagUsersModifyResp, error)` | `/cgi-bin/tag/deltagusers` | POST |

#### 邀请(1 个)

| # | 签名 | 接口 | HTTP |
|---|---|---|---|
| 18 | `(cc *CorpClient) InviteUser(ctx, req *InviteReq) (*InviteResp, error)` | `/cgi-bin/batch/invite` | POST |

### 1.2 非目标

- 异步批量导入(`batch/syncuser` / `batch/replaceuser` / `batch/replaceparty`):接口涉及异步 jobid 轮询,留到后续。
- 部门 ID 转 open_departmentid:企业微信无独立接口,不实现。
- 已离职成员列表(`leaved`):不在核心通讯录 API 范围。

## 2. 架构决策

### 2.1 复用 CorpClient HTTP helpers

子项目 4 已在 `corp.http.go` 中实现了 `cc.doPost` 和 `cc.doGet`,自动注入 `access_token`。本子项目所有方法直接复用,无需新增 HTTP 基础设施。

### 2.2 GET 方法参数传递

部门删除、成员删除等 GET 接口通过 query 参数传递。使用 `url.Values` + `cc.doGet`:

```go
func (cc *CorpClient) DeleteDepartment(ctx context.Context, id int) error {
    extra := url.Values{"id": {strconv.Itoa(id)}}
    return cc.doGet(ctx, "/cgi-bin/department/delete", extra, nil)
}
```

### 2.3 bool 参数编码

`ListUserSimple` / `ListUserDetail` 的 `fetch_child` 参数需要编码为 `"1"` / `"0"`。新增一个包级私有函数:

```go
func boolToStr(b bool) string {
    if b {
        return "1"
    }
    return "0"
}
```

放在 `corp.user.go` 文件中。

### 2.4 文件拆分

```
work-wechat/isv/
├── struct.contact.go         # NEW — 所有通讯录 DTO
├── corp.department.go        # NEW — 4 个部门方法
├── corp.department_test.go   # NEW — 3 个测试
├── corp.user.go              # NEW — 6 个成员方法 + boolToStr
├── corp.user_test.go         # NEW — 4 个测试
├── corp.tag.go               # NEW — 6 个标签方法
├── corp.tag_test.go          # NEW — 3 个测试
├── corp.invite.go            # NEW — 1 个邀请方法
└── corp.invite_test.go       # NEW — 1 个测试
```

## 3. DTO 设计(`struct.contact.go`)

### 3.1 部门

```go
// CreateDeptReq 创建部门请求。
type CreateDeptReq struct {
    Name     string `json:"name"`
    NameEn   string `json:"name_en,omitempty"`
    ParentID int    `json:"parentid"`
    Order    int    `json:"order,omitempty"`
    ID       int    `json:"id,omitempty"`
}

// CreateDeptResp 创建部门响应。
type CreateDeptResp struct {
    ID int `json:"id"`
}

// UpdateDeptReq 更新部门请求。
type UpdateDeptReq struct {
    ID       int    `json:"id"`
    Name     string `json:"name,omitempty"`
    NameEn   string `json:"name_en,omitempty"`
    ParentID int    `json:"parentid,omitempty"`
    Order    int    `json:"order,omitempty"`
}

// Department 部门信息(列表返回)。
type Department struct {
    ID               int    `json:"id"`
    Name             string `json:"name"`
    NameEn           string `json:"name_en"`
    DepartmentLeader []string `json:"department_leader"`
    ParentID         int    `json:"parentid"`
    Order            int    `json:"order"`
}
```

### 3.2 成员

```go
// CreateUserReq 创建成员请求。
type CreateUserReq struct {
    UserID         string   `json:"userid"`
    Name           string   `json:"name"`
    Alias          string   `json:"alias,omitempty"`
    Mobile         string   `json:"mobile,omitempty"`
    Department     []int    `json:"department"`
    Order          []int    `json:"order,omitempty"`
    Position       string   `json:"position,omitempty"`
    Gender         string   `json:"gender,omitempty"`
    Email          string   `json:"email,omitempty"`
    BizMail        string   `json:"biz_mail,omitempty"`
    IsLeaderInDept []int    `json:"is_leader_in_dept,omitempty"`
    DirectLeader   []string `json:"direct_leader,omitempty"`
    Enable         int      `json:"enable,omitempty"`
    Telephone      string   `json:"telephone,omitempty"`
    Address        string   `json:"address,omitempty"`
    MainDepartment int      `json:"main_department,omitempty"`
    ToInvite       bool     `json:"to_invite,omitempty"`
}

// UpdateUserReq 更新成员请求。
type UpdateUserReq struct {
    UserID         string   `json:"userid"`
    Name           string   `json:"name,omitempty"`
    Alias          string   `json:"alias,omitempty"`
    Mobile         string   `json:"mobile,omitempty"`
    Department     []int    `json:"department,omitempty"`
    Order          []int    `json:"order,omitempty"`
    Position       string   `json:"position,omitempty"`
    Gender         string   `json:"gender,omitempty"`
    Email          string   `json:"email,omitempty"`
    BizMail        string   `json:"biz_mail,omitempty"`
    IsLeaderInDept []int    `json:"is_leader_in_dept,omitempty"`
    DirectLeader   []string `json:"direct_leader,omitempty"`
    Enable         int      `json:"enable,omitempty"`
    Telephone      string   `json:"telephone,omitempty"`
    Address        string   `json:"address,omitempty"`
    MainDepartment int      `json:"main_department,omitempty"`
}

// UserSimple 简单成员信息。
type UserSimple struct {
    UserID     string `json:"userid"`
    Name       string `json:"name"`
    Department []int  `json:"department"`
    OpenUserID string `json:"open_userid"`
}

// UserSimpleListResp simplelist 响应。
type UserSimpleListResp struct {
    UserList []UserSimple `json:"userlist"`
}

// UserDetail 详细成员信息(GetUser / ListUserDetail 共用)。
type UserDetail struct {
    UserID         string   `json:"userid"`
    Name           string   `json:"name"`
    Department     []int    `json:"department"`
    Order          []int    `json:"order"`
    Position       string   `json:"position"`
    Mobile         string   `json:"mobile"`
    Gender         string   `json:"gender"`
    Email          string   `json:"email"`
    BizMail        string   `json:"biz_mail"`
    IsLeaderInDept []int    `json:"is_leader_in_dept"`
    DirectLeader   []string `json:"direct_leader"`
    Avatar         string   `json:"avatar"`
    ThumbAvatar    string   `json:"thumb_avatar"`
    Telephone      string   `json:"telephone"`
    Alias          string   `json:"alias"`
    Address        string   `json:"address"`
    OpenUserID     string   `json:"open_userid"`
    MainDepartment int      `json:"main_department"`
    Status         int      `json:"status"`
    QRCode         string   `json:"qr_code"`
}

// UserDetailListResp user/list 响应。
type UserDetailListResp struct {
    UserList []UserDetail `json:"userlist"`
}
```

### 3.3 标签

```go
// CreateTagReq 创建标签请求。
type CreateTagReq struct {
    TagName string `json:"tagname"`
    TagID   int    `json:"tagid,omitempty"`
}

// CreateTagResp 创建标签响应。
type CreateTagResp struct {
    TagID int `json:"tagid"`
}

// UpdateTagReq 更新标签请求。
type UpdateTagReq struct {
    TagID   int    `json:"tagid"`
    TagName string `json:"tagname"`
}

// Tag 标签信息。
type Tag struct {
    TagID   int    `json:"tagid"`
    TagName string `json:"tagname"`
}

// TagUser 标签成员。
type TagUser struct {
    UserID string `json:"userid"`
    Name   string `json:"name"`
}

// TagUsersResp tag/get 响应。
type TagUsersResp struct {
    TagName   string    `json:"tagname"`
    UserList  []TagUser `json:"userlist"`
    PartyList []int     `json:"partylist"`
}

// TagUsersModifyReq addtagusers / deltagusers 请求。
type TagUsersModifyReq struct {
    TagID     int      `json:"tagid"`
    UserList  []string `json:"userlist,omitempty"`
    PartyList []int    `json:"partylist,omitempty"`
}

// TagUsersModifyResp addtagusers / deltagusers 响应。
type TagUsersModifyResp struct {
    InvalidList  string `json:"invalidlist"`
    InvalidParty []int  `json:"invalidparty"`
}
```

### 3.4 邀请

```go
// InviteReq batch/invite 请求。
type InviteReq struct {
    User  []string `json:"user,omitempty"`
    Party []int    `json:"party,omitempty"`
    Tag   []int    `json:"tag,omitempty"`
}

// InviteResp batch/invite 响应。
type InviteResp struct {
    InvalidUser  []string `json:"invaliduser"`
    InvalidParty []int    `json:"invalidparty"`
    InvalidTag   []int    `json:"invalidtag"`
}
```

## 4. 实现细节

### 4.1 部门方法

```go
func (cc *CorpClient) CreateDepartment(ctx context.Context, req *CreateDeptReq) (*CreateDeptResp, error) {
    var resp CreateDeptResp
    if err := cc.doPost(ctx, "/cgi-bin/department/create", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

func (cc *CorpClient) UpdateDepartment(ctx context.Context, req *UpdateDeptReq) error {
    return cc.doPost(ctx, "/cgi-bin/department/update", req, nil)
}

func (cc *CorpClient) DeleteDepartment(ctx context.Context, id int) error {
    extra := url.Values{"id": {strconv.Itoa(id)}}
    return cc.doGet(ctx, "/cgi-bin/department/delete", extra, nil)
}

func (cc *CorpClient) ListDepartment(ctx context.Context, id int) ([]Department, error) {
    extra := url.Values{"id": {strconv.Itoa(id)}}
    var resp struct {
        Department []Department `json:"department"`
    }
    if err := cc.doGet(ctx, "/cgi-bin/department/list", extra, &resp); err != nil {
        return nil, err
    }
    return resp.Department, nil
}
```

### 4.2 成员方法

```go
func boolToStr(b bool) string {
    if b {
        return "1"
    }
    return "0"
}

func (cc *CorpClient) CreateUser(ctx context.Context, req *CreateUserReq) error {
    return cc.doPost(ctx, "/cgi-bin/user/create", req, nil)
}

func (cc *CorpClient) UpdateUser(ctx context.Context, req *UpdateUserReq) error {
    return cc.doPost(ctx, "/cgi-bin/user/update", req, nil)
}

func (cc *CorpClient) DeleteUser(ctx context.Context, userID string) error {
    extra := url.Values{"userid": {userID}}
    return cc.doGet(ctx, "/cgi-bin/user/delete", extra, nil)
}

func (cc *CorpClient) GetUser(ctx context.Context, userID string) (*UserDetail, error) {
    extra := url.Values{"userid": {userID}}
    var resp UserDetail
    if err := cc.doGet(ctx, "/cgi-bin/user/get", extra, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

func (cc *CorpClient) ListUserSimple(ctx context.Context, deptID int, fetchChild bool) (*UserSimpleListResp, error) {
    extra := url.Values{
        "department_id": {strconv.Itoa(deptID)},
        "fetch_child":   {boolToStr(fetchChild)},
    }
    var resp UserSimpleListResp
    if err := cc.doGet(ctx, "/cgi-bin/user/simplelist", extra, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

func (cc *CorpClient) ListUserDetail(ctx context.Context, deptID int, fetchChild bool) (*UserDetailListResp, error) {
    extra := url.Values{
        "department_id": {strconv.Itoa(deptID)},
        "fetch_child":   {boolToStr(fetchChild)},
    }
    var resp UserDetailListResp
    if err := cc.doGet(ctx, "/cgi-bin/user/list", extra, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.3 标签方法

```go
func (cc *CorpClient) CreateTag(ctx context.Context, req *CreateTagReq) (*CreateTagResp, error) {
    var resp CreateTagResp
    if err := cc.doPost(ctx, "/cgi-bin/tag/create", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

func (cc *CorpClient) UpdateTag(ctx context.Context, req *UpdateTagReq) error {
    return cc.doPost(ctx, "/cgi-bin/tag/update", req, nil)
}

func (cc *CorpClient) DeleteTag(ctx context.Context, tagID int) error {
    extra := url.Values{"tagid": {strconv.Itoa(tagID)}}
    return cc.doGet(ctx, "/cgi-bin/tag/delete", extra, nil)
}

func (cc *CorpClient) ListTag(ctx context.Context) ([]Tag, error) {
    var resp struct {
        TagList []Tag `json:"taglist"`
    }
    if err := cc.doGet(ctx, "/cgi-bin/tag/list", nil, &resp); err != nil {
        return nil, err
    }
    return resp.TagList, nil
}

func (cc *CorpClient) GetTagUsers(ctx context.Context, tagID int) (*TagUsersResp, error) {
    extra := url.Values{"tagid": {strconv.Itoa(tagID)}}
    var resp TagUsersResp
    if err := cc.doGet(ctx, "/cgi-bin/tag/get", extra, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

func (cc *CorpClient) AddTagUsers(ctx context.Context, req *TagUsersModifyReq) (*TagUsersModifyResp, error) {
    var resp TagUsersModifyResp
    if err := cc.doPost(ctx, "/cgi-bin/tag/addtagusers", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}

func (cc *CorpClient) DelTagUsers(ctx context.Context, req *TagUsersModifyReq) (*TagUsersModifyResp, error) {
    var resp TagUsersModifyResp
    if err := cc.doPost(ctx, "/cgi-bin/tag/deltagusers", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

### 4.4 邀请方法

```go
func (cc *CorpClient) InviteUser(ctx context.Context, req *InviteReq) (*InviteResp, error) {
    var resp InviteResp
    if err := cc.doPost(ctx, "/cgi-bin/batch/invite", req, &resp); err != nil {
        return nil, err
    }
    return &resp, nil
}
```

## 5. 测试策略

### 5.1 测试矩阵

| 文件 | # | Case | 断言 |
|---|---|---|---|
| `corp.department_test.go` | 1 | `TestCreateDepartment` | body.name 映射,resp.ID 非零 |
| `corp.department_test.go` | 2 | `TestListDepartment` | query id= 参数,返回 []Department |
| `corp.department_test.go` | 3 | `TestDeleteDepartment` | query id= 参数,GET 方法,无 error |
| `corp.user_test.go` | 4 | `TestCreateUser` | body.userid + body.name 映射 |
| `corp.user_test.go` | 5 | `TestGetUser` | query userid= 参数,返回 UserDetail 字段 |
| `corp.user_test.go` | 6 | `TestListUserSimple` | query department_id + fetch_child=1,返回 []UserSimple |
| `corp.user_test.go` | 7 | `TestListUserDetail` | query department_id + fetch_child=0,返回 []UserDetail |
| `corp.tag_test.go` | 8 | `TestCreateTag` | body.tagname 映射,resp.TagID 非零 |
| `corp.tag_test.go` | 9 | `TestGetTagUsers` | query tagid=,返回 userlist + partylist |
| `corp.tag_test.go` | 10 | `TestAddTagUsers` | body.tagid + body.userlist 映射 |
| `corp.invite_test.go` | 11 | `TestInviteUser` | body.user 映射,resp.InvalidUser |

所有测试复用 `newTestCorpClient(t, baseURL)`(已存在于 `corp.http_test.go`),token 预设 `CTOK`。

覆盖率目标 ≥85%。

## 6. 错误处理

- 沿用 `doPost` / `doGet` → `decodeRaw` → `WeixinError` 两阶段解码。
- 删除/更新类方法返回 `error`(`out` 传 `nil`)。
- 未新增哨兵错误。

## 7. 交付规模估计

| 文件 | 生产行数 | 测试行数 |
|---|---|---|
| `struct.contact.go` | ~200 | — |
| `corp.department.go` | ~50 | — |
| `corp.user.go` | ~70 | — |
| `corp.tag.go` | ~80 | — |
| `corp.invite.go` | ~15 | — |
| `corp.department_test.go` | — | ~90 |
| `corp.user_test.go` | — | ~130 |
| `corp.tag_test.go` | — | ~90 |
| `corp.invite_test.go` | — | ~40 |
| **合计** | **~415** | **~350** |

## 8. Commit 节奏

~5 个原子 commit:

1. 通讯录 DTO(`struct.contact.go`)
2. 部门方法 + 测试(`corp.department.go` + `corp.department_test.go`)
3. 成员方法 + 测试(`corp.user.go` + `corp.user_test.go`)
4. 标签方法 + 测试(`corp.tag.go` + `corp.tag_test.go`)
5. 邀请方法 + 测试 + 全量验证(`corp.invite.go` + `corp.invite_test.go` + race/cover/vet)

## 9. Self-Review Checklist

- [ ] 18 个公开方法全部实现
- [ ] GET 方法使用 `cc.doGet` + `url.Values` 参数传递
- [ ] POST 方法使用 `cc.doPost`
- [ ] `boolToStr` 编码 fetch_child 为 "1" / "0"
- [ ] `ListDepartment` 用内联 struct 解包 `department` 数组
- [ ] `ListTag` 用内联 struct 解包 `taglist` 数组
- [ ] 未新增哨兵错误
- [ ] 未修改 Config / Store / Client 结构
- [ ] `go test -race ./work-wechat/isv/...` 通过
- [ ] 覆盖率 ≥85%
