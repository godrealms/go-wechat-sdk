package errorx

type Error struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	Solution string `json:"solution"`
}

func (e *Error) Error() string {
	return e.Message
}
func NewError(code int, message, solution string) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Solution: solution,
	}
}

var Errors = map[string]*Error{
	"PARAM_ERROR":     NewError(400, "参数错误", "请根据错误提示正确传入参数"),
	"INVALID_REQUEST": NewError(400, "HTTP 请求不符合微信支付 APIv3 接口规则", "请参阅接口规则检查传入的参数"),
	"SIGN_ERROR":      NewError(401, "签名验证不通过", "请参阅签名常见问题排查"),
	"SYSTEM_ERROR":    NewError(500, "系统异常，请稍后重试", "请稍后重试"),

	"APPID_MCHID_NOT_MATCH": NewError(400, "AppID和mch_id不匹配", "请确认AppID和mch_id是否匹配，查询指引参考：https://pay.weixin.qq.com/doc/v3/merchant/4013289251"),
	"MCH_NOT_EXISTS":        NewError(400, "商户号不存在", "请检查商户号是否正确，商户号获取方式请参考：https://pay.weixin.qq.com/doc/v3/merchant/4013070756"),
	"NO_AUTH":               NewError(403, "商户无权限", "请商户前往商户平台申请此接口相关权限，参考：https://pay.weixin.qq.com/doc/v3/merchant/4013070174"),
	"OUT_TRADE_NO_USED":     NewError(403, "商户订单号重复", "请核实商户订单号是否重复提交"),
	"FREQUENCY_LIMITED":     NewError(429, "频率超限", "请求频率超限，请降低请求接口频率"),
}
