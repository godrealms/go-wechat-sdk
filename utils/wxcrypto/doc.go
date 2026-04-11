// Package wxcrypto 实现微信公众号/开放平台"消息加解密"(Biz Msg Crypt)。
//
// 参考文档:
//
//	https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Message_encryption_and_decryption.html
//
// 算法流程:
//  1. 校验 msg_signature = sha1(sort([token, timestamp, nonce, encrypted])).
//  2. 解密 encrypted: base64 -> AES-256-CBC(iv=aesKey[:16]) -> PKCS#7 unpad
//     -> 16 字节随机前缀 + 4 字节网络序长度 + 明文 + 发送方 appid.
//  3. 加密反之。
//
// 本包被 offiaccount 和 oplatform 共同引用。offiaccount 保留了
// 原有 MsgCrypto/ParseNotify 导出符号的薄别名，不破坏外部调用点。
package wxcrypto
