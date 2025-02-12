# Go WeChat SDK

一个功能完整的 Go 语言微信开发工具包，支持微信生态的主要服务接口。

[![Go Reference](https://pkg.go.dev/badge/github.com/godrealms/go-wechat-sdk.svg)](https://pkg.go.dev/github.com/godrealms/go-wechat-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/godrealms/go-wechat-sdk)](https://goreportcard.com/report/github.com/godrealms/go-wechat-sdk)
[![License](https://img.shields.io/github/license/godrealms/go-wechat-sdk)](LICENSE)

## 功能特性

- 支持微信生态全系列服务
- 简洁易用的 API 设计
- 完善的错误处理机制
- 详细的使用文档
- 丰富的示例代码

## 支持的服务

- 📱 小程序 (`mini-program`): 完整的小程序服务端 API
- 🎮 小游戏 (`mini-game`): 小游戏开发相关接口
- 📢 公众号 (`offiaccount`): 公众号全套功能支持
- 🤖 智能对话 (`aispeech`): AI 对话能力
- 🌐 开放平台 (`oplatform`): 第三方平台开发支持
- 💼 企业微信 (`work-wechat`): 企业微信应用开发
- 💰 微信支付 (`merchant`): 支付相关接口
- 📺 视频号 (`channels`): 视频号开放能力
- 🏪 微信小店 (`mini-store`): 小店商城功能
- 🤖 微信小微 (`xiaowei`): IoT 设备接入能力

## 快速开始

### 安装

```bash
go get github.com/godrealms/go-wechat-sdk
```

### 基础配置

```go
package main

import (
    "github.com/godrealms/go-wechat-sdk/offiaccount"
    "github.com/godrealms/go-wechat-sdk/merchant"
)

func main() {
    // 初始化公众号配置
    officialConfig := &offiaccount.Config{
        AppID:     "your-app-id",
        AppSecret: "your-app-secret",
    }
    
    // 创建公众号实例
    official := offiaccount.NewClient(officialConfig)
    
    // 初始化支付配置
    payConfig := &merchant.Config{
        MchID:    "your-merchant-id",
        ApiKey:   "your-api-key",
        CertFile: "path/to/cert.pem",
    }
    
    // 创建支付实例
    pay := merchant.NewClient(payConfig)
}
```

## 示例

### 小程序登录

```go
import "github.com/godrealms/go-wechat-sdk/mini-program"

func login(code string) (*miniprogram.LoginResponse, error) {
    client := miniprogram.NewClient(&miniprogram.Config{
        AppID:     "your-app-id",
        AppSecret: "your-app-secret",
    })
    
    return client.Login(code)
}
```

### 发送模板消息

```go
import "github.com/godrealms/go-wechat-sdk/offiaccount"

func sendTemplate(openID, templateID string, data interface{}) error {
    client := offiaccount.NewClient(config)
    
    msg := &offiaccount.TemplateMessage{
        ToUser:     openID,
        TemplateID: templateID,
        Data:       data,
    }
    
    return client.SendTemplate(msg)
}
```

## 文档

详细文档请访问我们的 [Wiki](https://github.com/godrealms/go-wechat-sdk/wiki)

## 贡献指南

欢迎提交 Pull Request 或创建 Issue。

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 联系我们

- 项目地址：[https://github.com/godrealms/go-wechat-sdk](https://github.com/godrealms/go-wechat-sdk)
- 问题反馈：[Issues](https://github.com/godrealms/go-wechat-sdk/issues)

## 致谢

感谢所有贡献者对本项目的支持！
