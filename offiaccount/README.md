# 微信公众号 SDK

微信公众号 Go SDK，提供完整的公众号 API 支持。

## 功能特性

- 🎯 完整的公众号 API 实现
- 🔒 自动 access_token 管理
- 📦 模块化设计，易于扩展
- 🛠️ 详细的错误处理机制
- 📘 丰富的示例代码

## 安装

```bash
go get github.com/godrealms/go-wechat-sdk/offiaccount
```

## 快速开始

### 基础配置

```go
package main

import (
    "context"
    "log"
    
    "github.com/godrealms/go-wechat-sdk/offiaccount"
)

func main() {
    // 初始化配置
    config := &offiaccount.Config{
        AppId:     "your-app-id",
        AppSecret: "your-app-secret",
    }
    
    // 创建客户端实例
    client := offiaccount.NewClient(context.Background(), config)
    
    // 现在可以使用 client 调用各种 API
}
```

## 核心功能

### 自定义菜单管理

```go
// 创建自定义菜单
menu := &offiaccount.CreateMenuRequest{
    Button: []*offiaccount.Button{
        {
            Type: "click",
            Name: "今日歌曲",
            Key:  "V1001_TODAY_MUSIC",
        },
        {
            Name: "菜单",
            SubButton: []*offiaccount.Button{
                {
                    Type: "view",
                    Name: "搜索",
                    URL:  "http://www.soso.com/",
                },
                {
                    Type: "click",
                    Name: "赞一下我们",
                    Key:  "V1001_GOOD",
                },
            },
        },
    },
}

err := client.CreateCustomMenu(menu)
if err != nil {
    log.Fatal(err)
}
```

### 消息管理

```go
// 发送文本消息
err := client.SendTextMessage(&offiaccount.TextMessageRequest{
    ToUser:  "OPENID",
    Content: "Hello World",
})
if err != nil {
    log.Fatal(err)
}

// 发送模板消息
template := &offiaccount.SubscribeMessageRequest{
    ToUser:     "OPENID",
    TemplateID: "TEMPLATE_ID",
    Data: map[string]interface{}{
        "first": map[string]string{
            "value": "恭喜你购买成功！",
            "color": "#173177",
        },
        "product": map[string]string{
            "value": "巧克力",
            "color": "#173177",
        },
        "amount": map[string]string{
            "value": "39.8元",
            "color": "#173177",
        },
        "time": map[string]string{
            "value": "2014年9月22日",
            "color": "#173177",
        },
    },
    URL: "http://weixin.qq.com/download",
}

result, err := client.SendTemplateMessage(template)
if err != nil {
    log.Fatal(err)
} else if result.ErrCode != 0 {
    log.Printf("发送模板消息失败: %s (错误码: %d)", result.ErrMsg, result.ErrCode)
}
```

### 用户管理

```go
// 获取用户列表
users, err := client.GetFansList("")
if err != nil {
    log.Fatal(err)
}

// 批量获取用户信息
openids := []string{"OPENID1", "OPENID2"}
userInfos, err := client.BatchGetUserInfo(&offiaccount.BatchGetUserInfoRequest{
    UserList: []*offiaccount.UserListItem{
        {Openid: openids[0], Language: "zh_CN"},
        {Openid: openids[1], Language: "zh_CN"},
    },
})
if err != nil {
    log.Fatal(err)
}
```

### 素材管理

```go
// 上传临时素材
media, err := client.UploadTemporaryMedia("image", "./test.jpg")
if err != nil {
    log.Fatal(err)
}

// 获取临时素材
err = client.GetTemporaryMedia("MEDIA_ID", "./download.jpg")
if err != nil {
    log.Fatal(err)
}
```

## 高级功能

### 发票功能

```go
// 获取用户发票抬头
titleURL, err := client.GetUserTitleUrl(&offiaccount.GetUserTitleUrlRequest{
    Attach: "自定义字段",
})
if err != nil {
    log.Fatal(err)
}

// 查询发票信息
invoiceInfo, err := client.GetInvoiceInfo(&offiaccount.GetInvoiceInfoRequest{
    CardID: "CARD_ID",
    EncryptCode: "ENCRYPT_CODE",
})
if err != nil {
    log.Fatal(err)
}

// 更新发票状态
err = client.UpdateInvoiceStatus(&offiaccount.UpdateInvoiceStatusRequest{
    CardID:          "CARD_ID",
    Code:            "CODE",
    ReimburseStatus: "INVOICE_REIMBURSE_INIT",
})
if err != nil {
    log.Fatal(err)
}
```

### 非税支付

```go
// 查询应收信息
feeInfo, err := client.QueryFee(&offiaccount.QueryFeeRequest{
    AppID:           "APP_ID",
    ServiceID:       123,
    PaymentNoticeNo: "PAYMENT_NOTICE_NO",
    DepartmentCode:  "DEPARTMENT_CODE",
    RegionCode:      "REGION_CODE",
})
if err != nil {
    log.Fatal(err)
}

// 缴费支付下单
order, err := client.UnifiedOrder(&offiaccount.UnifiedOrderRequest{
    AppID:           "APP_ID",
    Desc:            "缴费描述",
    Fee:             100, // 单位：分
    IP:              "127.0.0.1",
    PaymentNoticeNo: "PAYMENT_NOTICE_NO",
    DepartmentCode:  "DEPARTMENT_CODE",
    DepartmentName:  "DEPARTMENT_NAME",
    RegionCode:      "REGION_CODE",
    Items: []offiaccount.NonTaxItem{
        {
            No:      1,
            ItemID:  "ITEM_ID",
            ItemName: "项目名称",
            Fee:     100,
        },
    },
    PaymentNoticeCreateTime: time.Now().Unix(),
    Scene: "biz",
})
if err != nil {
    log.Fatal(err)
}
```

### 一码通

```go
// 申请二维码
code, err := client.ApplyCode(&offiaccount.ApplyCodeRequest{
    CodeCount: 10000,
    IsvApplicationID: "OUT_REQUEST_NO",
})
if err != nil {
    log.Fatal(err)
}

// 激活二维码
err = client.CodeActive(&offiaccount.CodeActiveRequest{
    ApplicationID: 123456,
    ActivityName: "活动名称",
    ProductBrand: "商品品牌",
    ProductTitle: "商品标题",
    ProductCode:  "商品条码",
    WxaAppid:     "小程序的appid",
    WxaPath:      "小程序的path",
    CodeStart:    0,
    CodeEnd:      9999,
})
if err != nil {
    log.Fatal(err)
}
```

### 医疗助手

```go
// 发送城市服务消息
result, err := client.SendChannelMsg(&offiaccount.SendChannelMsgRequest{
    Status:     1501001, // 预约挂号成功通知
    OpenID:     "USER_OPENID",
    OrderID:    "ORDER_ID",
    MsgID:      "MSG_ID",
    AppID:      "APP_ID",
    BusinessID: 150,
    BusinessInfo: &offiaccount.BusinessInfo{
        PatName:         "患者姓名",
        DocName:         "医生姓名",
        DepartmentName:  "科室名称",
        AppointmentTime: "2023-06-07 10:30-11:00",
        RedirectPage: &offiaccount.RedirectPage{
            PageType: "web",
            URL:      "https://example.com/order/detail",
        },
    },
})
if err != nil {
    log.Fatal(err)
} else if result.ErrCode != 0 {
    log.Printf("发送城市服务消息失败: %s (错误码: %d)", result.ErrMsg, result.ErrCode)
}
```

### 财政电子票据

```go
// 查询财政电子票据授权信息
authData, err := client.GetFiscalAuthData(&offiaccount.GetFiscalAuthDataRequest{
    OrderID: "ORDER_ID",
    SPAppID: "SP_APP_ID",
})
if err != nil {
    log.Fatal(err)
}

// 获取sdk临时票据
ticket, err := client.GetTicket()
if err != nil {
    log.Fatal(err)
}

// 拒绝开票
rejectResult, err := client.RejectInsertFiscal(&offiaccount.RejectInsertFiscalRequest{
    SPAppID: "SP_APP_ID",
    OrderID: "ORDER_ID",
    Reason:  "拒绝开票原因",
})
if err != nil {
    log.Fatal(err)
} else if rejectResult.ErrCode != 0 {
    log.Printf("拒绝开票失败: %s (错误码: %d)", rejectResult.ErrMsg, rejectResult.ErrCode)
}
```

## 错误处理

SDK 提供了统一的错误处理机制：

```go
result, err := client.SomeAPI()
if err != nil {
    // 网络错误或HTTP错误
    log.Printf("网络错误: %v", err)
    return
}

if result.ErrCode != 0 {
    // 微信API返回错误
    log.Printf("API错误: %s (错误码: %d)", result.ErrMsg, result.ErrCode)
    return
}

// 处理成功的结果
```

## Access Token 管理

SDK 内置了 Access Token 的自动管理机制，无需手动获取和刷新：

```go
// SDK 会自动处理 access_token 的获取和刷新
result, err := client.AnyAPI()
```

如果需要手动获取 access_token：

```go
token := client.GetAccessToken()
```

## 最佳实践

### 1. 客户端复用

```go
// 推荐：全局复用一个客户端实例
var client *offiaccount.Client

func init() {
    config := &offiaccount.Config{
        AppId:     "your-app-id",
        AppSecret: "your-app-secret",
    }
    client = offiaccount.NewClient(context.Background(), config)
}

func handler() {
    // 使用全局客户端实例
    client.GetUserInfo("openid")
}
```

### 2. 上下文管理

```go
// 为每个请求创建带超时的上下文
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

client := offiaccount.NewClient(ctx, config)
```

### 3. 日志记录

```go
// 启用调试日志
import "github.com/godrealms/go-wechat-sdk/utils"

// 在初始化时启用调试
client := offiaccount.NewClient(context.Background(), config)
// utils 包中会自动记录请求和响应日志
```

## 支持的 API

### 基础接口
- [x] 获取 Access Token
- [x] 获取稳定 Access Token
- [x] 网络检测
- [x] 获取微信服务器 IP
- [x] 获取 API 域名 IP

### 自定义菜单
- [x] 创建菜单
- [x] 查询菜单
- [x] 删除菜单
- [x] 个性化菜单
- [x] 查询自定义菜单信息

### 消息管理
- [x] 发送客服消息
- [x] 群发消息
- [x] 模板消息
- [x] 客服管理
- [x] 客服会话控制

### 用户管理
- [x] 用户标签管理
- [x] 用户信息管理
- [x] 黑名单管理
- [x] 备注名设置

### 账号管理
- [x] 二维码管理
- [x] 短 Key URL
- [x] 数据统计
- [x] 二维码跳转规则

### 素材管理
- [x] 临时素材
- [x] 永久素材
- [x] 草稿箱管理
- [x] 图文消息管理

### 发票功能
- [x] 发票信息查询
- [x] 发票状态管理
- [x] 发票抬头管理
- [x] 获取授权页链接
- [x] 设置授权页字段
- [x] 设置支付商户信息
- [x] 设置联系方式
- [x] 拒绝领取发票
- [x] 开票平台接口
- [x] 财政电子票据

### 非税支付
- [x] 查询应收信息
- [x] 缴费支付下单
- [x] 下载对账单
- [x] 订单管理
- [x] 退款处理
- [x] 模拟支付通知
- [x] 模拟查询缴费信息

### 一码通
- [x] 二维码申请
- [x] 二维码激活
- [x] 状态查询
- [x] 下载二维码包
- [x] ticket转码

### 医疗助手
- [x] 消息推送

### 留言功能
- [x] 开启/关闭留言功能
- [x] 删除留言
- [x] 标记留言处理状态
- [x] 获取图文统计

### 数据分析
- [x] 用户分析
- [x] 图文分析
- [x] 消息分析
- [x] 接口分析

### 模板消息
- [x] 发送模板消息
- [x] 模板管理
- [x] 行业设置

### 自动回复
- [x] 获取自动回复规则

### 微信门店
- [x] 门店小程序管理

### 微信AI开放平台
- [x] 语音识别
- [x] 通用印刷体识别
- [x] 行驶证识别
- [x] 银行卡识别
- [x] 营业执照识别
- [x] 驾驶证识别
- [x] 身份证识别
- [x] 图片智能裁剪
- [x] 二维码/条码识别

### JS-SDK
- [x] JS-SDK签名

### 网页授权
- [x] 获取授权链接
- [x] 获取用户信息

### 评论管理
- [x] 打开/关闭评论
- [x] 查看评论
- [x] 精选/取消精选评论
- [x] 删除评论
- [x] 回复评论
- [x] 删除回复

## 贡献指南

欢迎提交 Pull Request 或创建 Issue。

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个 Pull Request

## 许可证

本项目采用 Apache License 2.0 - 查看 [LICENSE](../LICENSE) 文件了解详情
