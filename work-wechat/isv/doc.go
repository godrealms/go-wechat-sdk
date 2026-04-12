// Package isv 提供企业微信第三方应用服务商(ISV)的认证底座。
//
// 主要能力:
//   - 维护 suite_access_token / provider_access_token 的生命周期
//   - 处理企业管理员扫码授权流程(pre_auth_code → permanent_code → corp_token)
//   - 为下游"代企业调用"子项目提供 TokenSource 注入点
//   - 解密企业微信回调事件(suite_ticket / 授权变更 / 通讯录变更等)
//
// 本包对标 oplatform 代公众号/代小程序的认证底座,架构范式保持一致。
package isv
