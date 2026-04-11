package offiaccount

import "github.com/godrealms/go-wechat-sdk/utils/wxcrypto"

// MsgCrypto is an alias for wxcrypto.MsgCrypto.
//
// 本类型是微信消息加解密的主入口。原实现已迁移到 utils/wxcrypto；
// 为了不破坏已经在生产使用的 offiaccount.NewMsgCrypto/*MsgCrypto 调用点，
// 这里保留类型别名和构造器转发。
type MsgCrypto = wxcrypto.MsgCrypto

// NewMsgCrypto 构造一个消息加解密器。等价于 wxcrypto.New。
func NewMsgCrypto(token, encodingAESKey, appid string) (*MsgCrypto, error) {
	return wxcrypto.New(token, encodingAESKey, appid)
}

// subtleConstEq 历史符号，保留以兼容同包内引用。新代码请用 wxcrypto.SubtleConstEq。
func subtleConstEq(a, b string) bool {
	return wxcrypto.SubtleConstEq(a, b)
}
