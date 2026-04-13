# aispeech — 微信 AI 语音

> 提供语音识别（ASR）、语音合成（TTS）、自然语言理解（NLU）和多轮对话功能，基于微信 AI 语音开放接口。

**Base URL：** `https://openai.weixin.qq.com`

## 适用场景

- 语音转文字（短音频实时同步识别 / 长音频异步识别，最长 300 秒）
- 文字转语音（TTS 合成，返回 Base64 编码 MP3）
- 文本语义理解与命名实体抽取
- 意图分类与置信度评分
- 多轮对话会话管理

## 初始化 / Initialization

```go
import "github.com/godrealms/go-wechat-sdk/aispeech"

c, err := aispeech.NewClient(aispeech.Config{
    AppId:     "wx_your_appid",
    AppSecret: "your_app_secret",
})
if err != nil {
    log.Fatal(err)
}
```

### Config 字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `AppId` | `string` | ✅ | 微信 AppID |
| `AppSecret` | `string` | 与 `TokenSource` 二选一 | 微信 AppSecret，用于自动获取并刷新 access_token |

> **Token 刷新机制：** 客户端在 access_token 到期前 60 秒自动刷新。若接口返回的 `expires_in` ≤ 60，则按 120 秒处理，避免立即再次刷新。

### Options

| Option | 说明 |
|--------|------|
| `WithHTTP(h *utils.HTTP)` | 注入自定义 HTTP 客户端，用于测试或代理场景 |
| `WithTokenSource(ts TokenSource)` | 注入外部 token 来源（如开放平台代调用），设置后不再调用 `/cgi-bin/token` |

## 错误处理 / Error Handling

所有方法在网络错误或微信 AI 接口返回非零 `errcode` 时返回非 nil `error`。错误信息格式为 `"aispeech: ..."` 并包含 `errcode` 和 `errmsg`，可直接通过 `err.Error()` 获取描述，或使用 `errors.Is` / `errors.As` 解包底层错误。

```go
resp, err := c.ASRShort(ctx, req)
if err != nil {
    // err.Error() 示例: "aispeech: token errcode=40001 errmsg=invalid credential"
    log.Fatal(err)
}
```

## API Reference

---

### ASRLong — 长音频异步识别

```go
func (c *Client) ASRLong(ctx context.Context, req *ASRLongReq) (*ASRLongResp, error)
```

提交长音频（最长 300 秒）识别任务，接口异步处理，立即返回任务 ID。

**ASRLongReq 参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `VoiceID` | `string` | ✅ | 音频唯一标识，由调用方自定义 |
| `VoiceURL` | `string` | ✅ | 音频文件的可访问 URL |
| `Format` | `string` | ✅ | 音频格式，如 `"mp3"`、`"wav"`、`"pcm"` |
| `Lang` | `string` | — | 语言代码，如 `"zh"`；为空时默认普通话 |
| `CallbackURL` | `string` | — | 识别完成后的回调地址；不填则需轮询结果 |

**返回值**

`*ASRLongResp`，其中 `TaskID` 为异步任务 ID，可用于后续查询识别结果。

---

### ASRShort — 短音频同步识别

```go
func (c *Client) ASRShort(ctx context.Context, req *ASRShortReq) (*ASRShortResp, error)
```

对 60 秒以内的音频进行同步识别，直接返回识别文本。

**ASRShortReq 参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `VoiceID` | `string` | ✅ | 音频唯一标识，由调用方自定义 |
| `VoiceData` | `string` | ✅ | Base64 编码的音频二进制数据 |
| `Format` | `string` | ✅ | 音频格式，如 `"mp3"`、`"wav"`、`"pcm"` |
| `Rate` | `int` | ✅ | 音频采样率，如 `16000` |
| `Bits` | `int` | ✅ | 音频位深，如 `16` |
| `Lang` | `string` | — | 语言代码，如 `"zh"`；为空时默认普通话 |

**返回值**

`*ASRShortResp`，其中 `Result` 为识别出的文本字符串。

---

### TextToSpeech — 文字转语音

```go
func (c *Client) TextToSpeech(ctx context.Context, req *TextToSpeechReq) (*TextToSpeechResp, error)
```

将文本合成为语音，返回 Base64 编码的 MP3 音频数据。

**TextToSpeechReq 参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `Text` | `string` | ✅ | 待合成的文本内容 |
| `Speed` | `int` | — | 语速，取值范围 `[-500, 500]`，`0` 为正常速度 |
| `Volume` | `int` | — | 音量，取值范围 `[0, 100]`，`0` 为默认音量 |
| `Pitch` | `int` | — | 音调，取值范围 `[-500, 500]`，`0` 为正常音调 |
| `VoiceType` | `int` | — | 音色类型，不同值对应不同发音人；`0` 为默认音色 |

**返回值**

`*TextToSpeechResp`：
- `AudioData`：Base64 编码的 MP3 音频数据
- `AudioSize`：音频数据字节数
- `SessionID`：本次合成的会话 ID

---

### NLUUnderstand — 语义理解

```go
func (c *Client) NLUUnderstand(ctx context.Context, req *NLUUnderstandReq) (*NLUUnderstandResp, error)
```

对输入文本进行语义分析，提取意图和命名实体（槽位）。

**NLUUnderstandReq 参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `Query` | `string` | ✅ | 待理解的文本内容 |
| `SessionID` | `string` | — | 会话 ID，用于多轮上下文关联 |
| `Lang` | `string` | — | 语言代码，如 `"zh"` |

**返回值**

`*NLUUnderstandResp`：
- `Intent`：识别出的意图名称
- `Slots`：命名实体列表（`[]NLUEntity`），每个元素包含 `Type`（实体类型）、`Value`（实体值）、`Begin`（起始位置）、`End`（结束位置）
- `SessionID`：本次请求的会话 ID

---

### NLUIntentRecognize — 意图识别

```go
func (c *Client) NLUIntentRecognize(ctx context.Context, req *NLUIntentRecognizeReq) (*NLUIntentRecognizeResp, error)
```

将输入文本与已配置的意图集合进行匹配，返回最佳匹配意图及置信度。

**NLUIntentRecognizeReq 参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `Query` | `string` | ✅ | 待识别的文本内容 |
| `IntentIDs` | `[]string` | — | 候选意图 ID 列表；为空则在全部意图中匹配 |
| `SessionID` | `string` | — | 会话 ID，用于多轮上下文关联 |

**返回值**

`*NLUIntentRecognizeResp`：
- `IntentID`：匹配到的意图 ID
- `IntentName`：匹配到的意图名称
- `Confidence`：置信度，取值范围 `[0, 1]`，越高越可信
- `SessionID`：本次请求的会话 ID

---

### DialogQuery — 多轮对话

```go
func (c *Client) DialogQuery(ctx context.Context, req *DialogQueryReq) (*DialogQueryResp, error)
```

向对话系统发送一条用户话语，返回系统回复。通过保持相同的 `SessionID` 维持多轮上下文。

**DialogQueryReq 参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `Query` | `string` | ✅ | 用户输入的文本 |
| `SessionID` | `string` | ✅ | 会话 ID；同一对话中保持不变，首次可传空字符串并使用返回值中的 `SessionID` |
| `Lang` | `string` | — | 语言代码，如 `"zh"` |

**返回值**

`*DialogQueryResp`：
- `Answer`：系统生成的回复文本
- `SessionID`：本轮对话的会话 ID，供后续轮次使用
- `EndFlag`：为 `true` 时表示对话已结束，无需继续发送消息

---

### DialogReset — 重置对话

```go
func (c *Client) DialogReset(ctx context.Context, req *DialogResetReq) error
```

终止并重置指定会话，清除所有上下文状态。重置后如需继续对话需开启新会话。

**DialogResetReq 参数**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `SessionID` | `string` | ✅ | 要重置的会话 ID |

**返回值**

仅返回 `error`；成功时为 `nil`。

---

## 完整示例 / Complete Example

```go
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/godrealms/go-wechat-sdk/aispeech"
)

func main() {
	ctx := context.Background()

	// 初始化客户端
	c, err := aispeech.NewClient(aispeech.Config{
		AppId:     os.Getenv("WX_APPID"),
		AppSecret: os.Getenv("WX_APPSECRET"),
	})
	if err != nil {
		log.Fatal(err)
	}

	// 短音频同步识别
	audioBytes, err := os.ReadFile("hello.mp3")
	if err != nil {
		log.Fatal(err)
	}
	asrResp, err := c.ASRShort(ctx, &aispeech.ASRShortReq{
		VoiceID:   "voice_001",
		VoiceData: base64.StdEncoding.EncodeToString(audioBytes),
		Format:    "mp3",
		Rate:      16000,
		Bits:      16,
		Lang:      "zh",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("识别结果:", asrResp.Result)

	// 多轮对话：使用识别文本作为第一轮输入
	dResp, err := c.DialogQuery(ctx, &aispeech.DialogQueryReq{
		Query:     asrResp.Result,
		SessionID: "", // 首次传空，使用返回的 SessionID 继续后续轮次
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("对话回复:", dResp.Answer)
	fmt.Println("会话 ID:", dResp.SessionID)
	fmt.Println("对话已结束:", dResp.EndFlag)

	// 对话结束后重置会话
	if err := c.DialogReset(ctx, &aispeech.DialogResetReq{
		SessionID: dResp.SessionID,
	}); err != nil {
		log.Fatal(err)
	}
	fmt.Println("会话已重置")
}
```
