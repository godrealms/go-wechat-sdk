# Phase 1E: mini-program Data Analysis, Live Streaming, Logistics

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add data analysis (visit trends, user portrait, performance), live streaming room management, and logistics/delivery APIs to the mini-program package.

**Architecture:** Same pattern as Plans C/D — use c.Ctx, c.GetAccessToken(), c.TokenQuery(). POST endpoints put access_token in URL. GET endpoints use c.TokenQuery(). All response types embed core.Resp.

**Tech Stack:** Go 1.23.1, standard library only. Depends on Plan C being complete.

**PREREQUISITE:** `2026-04-09-phase1c-miniprogram-client-auth.md` must be completed first.

---

### Task 1: mini-program/struct.analysis.go + mini-program/api.analysis.go

**Files:**
- Create: `mini-program/struct.analysis.go`
- Create: `mini-program/api.analysis.go`

- [ ] **Step 1: Create `mini-program/struct.analysis.go`**

```go
package mini_program

import "github.com/godrealms/go-wechat-sdk/core"

// DailyVisitTrendItem represents one day of visit trend data
type DailyVisitTrendItem struct {
	RefDate            string  `json:"ref_date"`
	SessionCnt         int64   `json:"session_cnt"`
	VisitPv            int64   `json:"visit_pv"`
	VisitUv            int64   `json:"visit_uv"`
	VisitUvNew         int64   `json:"visit_uv_new"`
	StayTimeSession    float64 `json:"stay_time_session"`
	VisitDepth         float64 `json:"visit_depth"`
}

// VisitTrendResult is the result of GetDailyVisitTrend / GetWeeklyVisitTrend / GetMonthlyVisitTrend
type VisitTrendResult struct {
	core.Resp
	List []*DailyVisitTrendItem `json:"list"`
}

// DailyRetainInfo contains new and active user retain info
type DailyRetainInfo struct {
	RefDate    string `json:"ref_date"`
	VisitUvNew int64  `json:"visit_uv_new"`
	VisitUv    int64  `json:"visit_uv"`
}

// UserRetainResult is the result of GetDailyRetain / GetWeeklyRetain / GetMonthlyRetain
type UserRetainResult struct {
	core.Resp
	RefDate   string             `json:"ref_date"`
	VisitUvNew []*DailyRetainInfo `json:"visit_uv_new"`
	VisitUv    []*DailyRetainInfo `json:"visit_uv"`
}

// VisitPageItem represents one page's visit statistics
type VisitPageItem struct {
	PagePath        string  `json:"page_path"`
	PageVisitPv     int64   `json:"page_visit_pv"`
	PageVisitUv     int64   `json:"page_visit_uv"`
	PageStayTimeUv  float64 `json:"page_staytime_uv"`
	EntrypagePv     int64   `json:"entrypage_pv"`
	ExitpagePv      int64   `json:"exitpage_pv"`
	PageSharePv     int64   `json:"page_share_pv"`
	PageShareUv     int64   `json:"page_share_uv"`
}

// VisitPageResult is the result of GetVisitPage
type VisitPageResult struct {
	core.Resp
	List []*VisitPageItem `json:"list"`
}

// AnalysisDateRequest is the common date range request for analysis APIs
type AnalysisDateRequest struct {
	BeginDate string `json:"begin_date"` // format: 20170313
	EndDate   string `json:"end_date"`   // format: 20170313
}

// UserPortraitItem represents user attribute distribution
type UserPortraitItem struct {
	Id    int64   `json:"id"`
	Name  string  `json:"name"`
	Count int64   `json:"count"`
}

// UserPortraitResult is the result of GetUserPortrait
type UserPortraitResult struct {
	core.Resp
	RefDate  string              `json:"ref_date"`
	VisitUv  *UserPortraitDetail `json:"visit_uv"`
	ShareUv  *UserPortraitDetail `json:"share_uv"`
}

// UserPortraitDetail contains demographic breakdown
type UserPortraitDetail struct {
	Province []*UserPortraitItem `json:"province"`
	City     []*UserPortraitItem `json:"city"`
	Genders  []*UserPortraitItem `json:"genders"`
	Platforms []*UserPortraitItem `json:"platforms"`
	Devices  []*UserPortraitItem `json:"devices"`
	Ages     []*UserPortraitItem `json:"ages"`
}

// PerformanceQueryRequest is the request for GetPerformanceData
type PerformanceQueryRequest struct {
	CommonQuery *PerformanceCommonQuery `json:"commonQuery"`
	Queries     []*PerformanceQuery     `json:"queries"`
}

// PerformanceCommonQuery contains common query parameters
type PerformanceCommonQuery struct {
	AppId string `json:"appid"`
}

// PerformanceQuery specifies a single metric to query
type PerformanceQuery struct {
	Metric    string `json:"metric"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}

// PerformanceDataResult is the result of GetPerformanceData
type PerformanceDataResult struct {
	core.Resp
	Data []*PerformanceDataItem `json:"data"`
}

// PerformanceDataItem represents one metric's data
type PerformanceDataItem struct {
	Metric string                   `json:"metric"`
	Data   []*PerformanceDataPoint  `json:"data"`
}

// PerformanceDataPoint is a single time-series data point
type PerformanceDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}
```

- [ ] **Step 2: Create `mini-program/api.analysis.go`**

```go
package mini_program

import "fmt"

// GetDailyVisitTrend 获取用户访问小程序日趋势
// POST /datacube/getweanalysisappiddailyvisittrend
func (c *Client) GetDailyVisitTrend(req *AnalysisDateRequest) (*VisitTrendResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappiddailyvisittrend?access_token=%s", c.GetAccessToken())
	result := &VisitTrendResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetWeeklyVisitTrend 获取用户访问小程序周趋势
// POST /datacube/getweanalysisappidweeklyvisittrend
func (c *Client) GetWeeklyVisitTrend(req *AnalysisDateRequest) (*VisitTrendResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappidweeklyvisittrend?access_token=%s", c.GetAccessToken())
	result := &VisitTrendResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetMonthlyVisitTrend 获取用户访问小程序月趋势
// POST /datacube/getweanalysisappidmonthlyvisittrend
func (c *Client) GetMonthlyVisitTrend(req *AnalysisDateRequest) (*VisitTrendResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappidmonthlyvisittrend?access_token=%s", c.GetAccessToken())
	result := &VisitTrendResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetDailyRetain 获取用户小程序访问日留存
// POST /datacube/getweanalysisappiddailyretaininfo
func (c *Client) GetDailyRetain(req *AnalysisDateRequest) (*UserRetainResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappiddailyretaininfo?access_token=%s", c.GetAccessToken())
	result := &UserRetainResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetWeeklyRetain 获取用户小程序访问周留存
// POST /datacube/getweanalysisappidweeklyretaininfo
func (c *Client) GetWeeklyRetain(req *AnalysisDateRequest) (*UserRetainResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappidweeklyretaininfo?access_token=%s", c.GetAccessToken())
	result := &UserRetainResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetMonthlyRetain 获取用户小程序访问月留存
// POST /datacube/getweanalysisappidmonthlyretaininfo
func (c *Client) GetMonthlyRetain(req *AnalysisDateRequest) (*UserRetainResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappidmonthlyretaininfo?access_token=%s", c.GetAccessToken())
	result := &UserRetainResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetVisitPage 获取小程序访问页面数据
// POST /datacube/getweanalysisappidvisitpage
func (c *Client) GetVisitPage(req *AnalysisDateRequest) (*VisitPageResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappidvisitpage?access_token=%s", c.GetAccessToken())
	result := &VisitPageResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetUserPortrait 获取小程序新增或活跃用户的画像分布数据
// POST /datacube/getweanalysisappiduserportrait
func (c *Client) GetUserPortrait(req *AnalysisDateRequest) (*UserPortraitResult, error) {
	path := fmt.Sprintf("/datacube/getweanalysisappiduserportrait?access_token=%s", c.GetAccessToken())
	result := &UserPortraitResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetPerformanceData 小程序性能监控数据
// POST /wxaapi/log/get_performance
func (c *Client) GetPerformanceData(req *PerformanceQueryRequest) (*PerformanceDataResult, error) {
	path := fmt.Sprintf("/wxaapi/log/get_performance?access_token=%s", c.GetAccessToken())
	result := &PerformanceDataResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}
```

- [ ] **Step 3: Commit**

```bash
cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
git add mini-program/struct.analysis.go mini-program/api.analysis.go
git commit -m "feat(mini-program): add data analysis APIs"
```

---

### Task 2: mini-program/struct.live.go + mini-program/api.live.go

**Files:**
- Create: `mini-program/struct.live.go`
- Create: `mini-program/api.live.go`

- [ ] **Step 1: Create `mini-program/struct.live.go`**

```go
package mini_program

import "github.com/godrealms/go-wechat-sdk/core"

// LiveRoom represents a live streaming room
type LiveRoom struct {
	Name         string `json:"name"`
	Roomid       int64  `json:"roomid"`
	CoverImg     string `json:"cover_img"`
	ShareImg     string `json:"share_img"`
	LiveStatus   int    `json:"live_status"` // 101=直播中 102=未开始 103=已结束 104=禁播 105=暂停
	StartTime    int64  `json:"start_time"`
	EndTime      int64  `json:"end_time"`
	AnchorName   string `json:"anchor_name"`
	AnchorImgUrl string `json:"anchor_img_url"`
	Goods        []*LiveGoods `json:"goods"`
	RoomType     int    `json:"room_type"` // 0=普通直播 1=小店直播
	SubRoomType  int    `json:"sub_room_type"`
}

// LiveGoods represents a product in a live room
type LiveGoods struct {
	GoodsId   int64  `json:"goods_id"`
	CoverImg  string `json:"cover_img"`
	Url       string `json:"url"`
	Price     int64  `json:"price"`
	Name      string `json:"name"`
	Price2    int64  `json:"price2"`
	PriceType int    `json:"price_type"` // 1=一口价 2=价格区间 3=折扣价
	ThirdPartyAppid string `json:"third_party_appid,omitempty"`
}

// GetLiveRoomsResult is the result of GetLiveRooms
type GetLiveRoomsResult struct {
	core.Resp
	RoomInfo []*LiveRoom `json:"room_info"`
	Total    int64       `json:"total"`
}

// GetLiveRoomsRequest is the request for GetLiveRooms
type GetLiveRoomsRequest struct {
	Start int `json:"start"` // offset, starts from 0
	Limit int `json:"limit"` // max 10
}

// LiveGoodsListResult is the result of GetLiveGoods
type LiveGoodsListResult struct {
	core.Resp
	Goods       []*LiveGoods `json:"goods"`
	TotalNum    int64        `json:"total_num"`
}
```

- [ ] **Step 2: Create `mini-program/api.live.go`**

```go
package mini_program

import (
	"fmt"
	"net/url"
)

// GetLiveRooms 获取直播间列表及直播间信息
// GET /wxa/business/getliveinfo
func (c *Client) GetLiveRooms(start, limit int) (*GetLiveRoomsResult, error) {
	query := c.TokenQuery(url.Values{
		"start": {fmt.Sprintf("%d", start)},
		"limit": {fmt.Sprintf("%d", limit)},
	})
	result := &GetLiveRoomsResult{}
	err := c.Https.Get(c.Ctx, "/wxa/business/getliveinfo", query, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetLiveRoomInfo 获取指定直播间信息
// GET /wxa/business/getliveinfo?room_id=ROOMID
func (c *Client) GetLiveRoomInfo(roomId int64) (*GetLiveRoomsResult, error) {
	query := c.TokenQuery(url.Values{
		"room_id": {fmt.Sprintf("%d", roomId)},
	})
	result := &GetLiveRoomsResult{}
	err := c.Https.Get(c.Ctx, "/wxa/business/getliveinfo", query, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetLiveGoods 获取商品列表
// GET /wxa/business/getgoodswarehouse
func (c *Client) GetLiveGoods(start, limit int) (*LiveGoodsListResult, error) {
	query := c.TokenQuery(url.Values{
		"offset": {fmt.Sprintf("%d", start)},
		"limit":  {fmt.Sprintf("%d", limit)},
	})
	result := &LiveGoodsListResult{}
	err := c.Https.Get(c.Ctx, "/wxa/business/getgoodswarehouse", query, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// AddLiveGoods 直播间导入商品
// POST /wxa/business/add_goods
func (c *Client) AddLiveGoods(roomId int64, goodsIds []int64) error {
	path := fmt.Sprintf("/wxa/business/add_goods?access_token=%s", c.GetAccessToken())
	body := map[string]interface{}{
		"roomId":  roomId,
		"goodsId": goodsIds,
	}
	result := &struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}{}
	err := c.Https.Post(c.Ctx, path, body, result)
	if err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("wechat api error %d: %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}
```

- [ ] **Step 3: Commit**

```bash
cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
git add mini-program/struct.live.go mini-program/api.live.go
git commit -m "feat(mini-program): add live streaming APIs"
```

---

### Task 3: mini-program/struct.logistics.go + mini-program/api.logistics.go

**Files:**
- Create: `mini-program/struct.logistics.go`
- Create: `mini-program/api.logistics.go`

- [ ] **Step 1: Create `mini-program/struct.logistics.go`**

```go
package mini_program

import "github.com/godrealms/go-wechat-sdk/core"

// AddExpressOrderRequest is the request to create an express waybill
type AddExpressOrderRequest struct {
	AddSource   int                    `json:"add_source"` // 0=微信侧 2=自定义
	WxAppId     string                 `json:"wx_appid,omitempty"`
	OrderId     string                 `json:"order_id"`
	OpenId      string                 `json:"openid,omitempty"`
	DeliveryId  string                 `json:"delivery_id"`
	BizId       string                 `json:"biz_id"`
	CustomRemark string                `json:"custom_remark,omitempty"`
	Tagid       int                    `json:"tagid,omitempty"`
	Sender      *ExpressContact        `json:"sender"`
	Receiver    *ExpressContact        `json:"receiver"`
	Cargo       *ExpressCargo          `json:"cargo"`
	Shop        *ExpressShop           `json:"shop,omitempty"`
	SubBizId    string                 `json:"sub_biz_id,omitempty"`
}

// ExpressContact represents sender or receiver info
type ExpressContact struct {
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Company  string `json:"company,omitempty"`
	PostCode string `json:"post_code,omitempty"`
	Country  string `json:"country,omitempty"`
	Province string `json:"province"`
	City     string `json:"city"`
	Area     string `json:"area"`
	Address  string `json:"address"`
}

// ExpressCargo describes the parcel contents
type ExpressCargo struct {
	Count      int     `json:"count"`
	Weight     float64 `json:"weight"`
	SpaceX     float64 `json:"space_x"`
	SpaceY     float64 `json:"space_y"`
	SpaceZ     float64 `json:"space_z"`
	DetailList []*ExpressCargoDetail `json:"detail_list"`
}

// ExpressCargoDetail is one item in the parcel
type ExpressCargoDetail struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ExpressShop holds shop info for print
type ExpressShop struct {
	WxaPath  string `json:"wxa_path"`
	ImgUrl   string `json:"img_url,omitempty"`
	GoodsName string `json:"goods_name,omitempty"`
	GoodsCount int  `json:"goods_count,omitempty"`
}

// AddExpressOrderResult is the result of AddExpressOrder
type AddExpressOrderResult struct {
	core.Resp
	OrderId    string `json:"order_id"`
	WaybillId  string `json:"waybill_id"`
	WaybillData []*WaybillData `json:"waybill_data"`
}

// WaybillData is one field on the printed waybill
type WaybillData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetExpressOrderResult is the result of GetExpressOrder
type GetExpressOrderResult struct {
	core.Resp
	OrderId       string `json:"order_id"`
	OpenId        string `json:"openid,omitempty"`
	DeliveryId    string `json:"delivery_id"`
	WaybillId     string `json:"waybill_id"`
	WaybillData   []*WaybillData `json:"waybill_data"`
	PathItemNum   int    `json:"path_item_num"`
	PathItemList  []*ExpressPathItem `json:"path_item_list"`
}

// ExpressPathItem is one tracking update
type ExpressPathItem struct {
	ActionTime int64  `json:"action_time"`
	ActionType int    `json:"action_type"`
	ActionMsg  string `json:"action_msg"`
}

// CancelExpressOrderRequest is the request to cancel a waybill
type CancelExpressOrderRequest struct {
	OrderId    string `json:"order_id"`
	OpenId     string `json:"openid,omitempty"`
	DeliveryId string `json:"delivery_id"`
	WaybillId  string `json:"waybill_id"`
}

// CancelExpressOrderResult is the result of CancelExpressOrder
type CancelExpressOrderResult struct {
	core.Resp
	Count int `json:"count"` // 0=失败 1=成功
}

// ExpressDelivery represents an express delivery company
type ExpressDelivery struct {
	DeliveryId   string `json:"delivery_id"`
	DeliveryName string `json:"delivery_name"`
}

// GetAllDeliveryResult is the result of GetAllDelivery
type GetAllDeliveryResult struct {
	core.Resp
	Count    int                `json:"count"`
	Data     []*ExpressDelivery `json:"data"`
}
```

- [ ] **Step 2: Create `mini-program/api.logistics.go`**

```go
package mini_program

import (
	"fmt"
	"net/url"
)

// AddExpressOrder 生成运单
// POST /cgi-bin/express/business/order/add
func (c *Client) AddExpressOrder(req *AddExpressOrderRequest) (*AddExpressOrderResult, error) {
	path := fmt.Sprintf("/cgi-bin/express/business/order/add?access_token=%s", c.GetAccessToken())
	result := &AddExpressOrderResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetExpressOrder 查询运单
// POST /cgi-bin/express/business/order/get
func (c *Client) GetExpressOrder(orderId, openId, deliveryId, waybillId string) (*GetExpressOrderResult, error) {
	path := fmt.Sprintf("/cgi-bin/express/business/order/get?access_token=%s", c.GetAccessToken())
	body := map[string]string{
		"order_id":    orderId,
		"openid":      openId,
		"delivery_id": deliveryId,
		"waybill_id":  waybillId,
	}
	result := &GetExpressOrderResult{}
	err := c.Https.Post(c.Ctx, path, body, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// CancelExpressOrder 取消运单
// POST /cgi-bin/express/business/order/cancel
func (c *Client) CancelExpressOrder(req *CancelExpressOrderRequest) (*CancelExpressOrderResult, error) {
	path := fmt.Sprintf("/cgi-bin/express/business/order/cancel?access_token=%s", c.GetAccessToken())
	result := &CancelExpressOrderResult{}
	err := c.Https.Post(c.Ctx, path, req, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}

// GetAllDelivery 获取支持的快递公司列表
// GET /cgi-bin/express/business/delivery/getall
func (c *Client) GetAllDelivery() (*GetAllDeliveryResult, error) {
	query := c.TokenQuery(url.Values{})
	result := &GetAllDeliveryResult{}
	err := c.Https.Get(c.Ctx, "/cgi-bin/express/business/delivery/getall", query, result)
	if err != nil {
		return nil, err
	}
	return result, result.GetError()
}
```

- [ ] **Step 3: Commit**

```bash
cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
git add mini-program/struct.logistics.go mini-program/api.logistics.go
git commit -m "feat(mini-program): add logistics APIs"
```

---

### Task 4: Tests + build verification

**Files:**
- Create: `mini-program/api_analysis_test.go`

- [ ] **Step 1: Create `mini-program/api_analysis_test.go`**

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

func TestGetDailyVisitTrend(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/datacube/getweanalysisappiddailyvisittrend" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode": 0,
				"errmsg":  "ok",
				"list": []map[string]interface{}{
					{"ref_date": "20240101", "session_cnt": 100, "visit_pv": 200, "visit_uv": 50},
				},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "tok", "expires_in": 7200})
		}
	}))
	defer srv.Close()

	cfg := &Config{BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"}}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "POST")
	c := &Client{BaseClient: base}

	result, err := c.GetDailyVisitTrend(&AnalysisDateRequest{
		BeginDate: "20240101",
		EndDate:   "20240101",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.List) != 1 {
		t.Errorf("expected 1 item, got %d", len(result.List))
	}
	if result.List[0].RefDate != "20240101" {
		t.Errorf("expected 20240101, got %s", result.List[0].RefDate)
	}
}

func TestGetAllDelivery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/express/business/delivery/getall" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode": 0,
				"errmsg":  "ok",
				"count":   2,
				"data": []map[string]interface{}{
					{"delivery_id": "SF", "delivery_name": "顺丰速递"},
					{"delivery_id": "ZTO", "delivery_name": "中通快递"},
				},
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"access_token": "tok", "expires_in": 7200})
		}
	}))
	defer srv.Close()

	cfg := &Config{BaseConfig: core.BaseConfig{AppId: "app1", AppSecret: "sec1"}}
	base := core.NewBaseClient(context.Background(), &cfg.BaseConfig, srv.URL, "/token", "POST")
	c := &Client{BaseClient: base}

	result, err := c.GetAllDelivery()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 2 {
		t.Errorf("expected 2 deliveries, got %d", result.Count)
	}
	if result.Data[0].DeliveryId != "SF" {
		t.Errorf("expected SF, got %s", result.Data[0].DeliveryId)
	}
}
```

- [ ] **Step 2: Run build and tests**

```bash
cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
go build ./...
go test ./mini-program/ -v
```

- [ ] **Step 3: Commit**

```bash
cd /Volumes/Fanxiang-S790-1TB-Media/www.godrealms.cn/go-wechat-sdk/.claude/worktrees/focused-babbage
git add mini-program/api_analysis_test.go
git commit -m "test(mini-program): add analysis and logistics tests"
```
