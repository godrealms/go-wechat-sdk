package offiaccount

import "encoding/json"

// RedirectPage 跳转页面信息
type RedirectPage struct {
	PageType string `json:"page_type"` // 页面类型
	URL      string `json:"url"`       // 跳转链接
}

// BusinessInfo 业务字段信息
type BusinessInfo struct {
	PatName         string        `json:"pat_name,omitempty"`         // 患者姓名
	DocName         string        `json:"doc_name,omitempty"`         // 医生姓名
	PatHospitalID   string        `json:"pat_hospital_id,omitempty"`  // 患者医院ID
	DepartmentName  string        `json:"department_name,omitempty"`  // 科室名称
	AppointmentTime string        `json:"appointment_time,omitempty"` // 预约时间
	RedirectPage    *RedirectPage `json:"redirect_page,omitempty"`    // 跳转页面信息
}

// SendChannelMsgRequest 消息推送请求参数
type SendChannelMsgRequest struct {
	Status       int64         `json:"status"`                  // 消息子状态(如：1501001-预约挂号成功通知)
	OpenID       string        `json:"open_id"`                 // 用户openid(公众号/小程序)
	OrderID      string        `json:"order_id"`                // 业务方生成的唯一订单ID
	MsgID        string        `json:"msg_id"`                  // 消息唯一标识(需保证同用户同order_id下唯一)
	AppID        string        `json:"app_id"`                  // 公众号appid(需开通就医助手)
	BusinessID   int64         `json:"business_id"`             // 固定值150
	BusinessInfo *BusinessInfo `json:"business_info,omitempty"` // 业务字段(不同status对应不同结构)
}

// SendChannelMsgResult 消息推送结果
type SendChannelMsgResult struct {
	Resp
}

// SendChannelMsg 消息推送接口
// 用于下发就医助手消息，结合通用参数和不同子状态status参数组合实现各类业务消息推送
// https://developers.weixin.qq.com/doc/service/api/medicalassistant/api_cityservice_sendchannelmsg.html
func (c *Client) SendChannelMsg(request *SendChannelMsgRequest) (*SendChannelMsgResult, error) {
	result := &SendChannelMsgResult{}

	// 将business_info转换为json.RawMessage以保持原始结构
	var body map[string]interface{}
	bodyBytes, _ := json.Marshal(request)
	json.Unmarshal(bodyBytes, &body)

	// 特别处理business_info字段
	if request.BusinessInfo != nil {
		businessInfoBytes, _ := json.Marshal(request.BusinessInfo)
		body["business_info"] = json.RawMessage(businessInfoBytes)
	}

	err := c.Https.Post(c.ctx, "/cityservice/sendchannelmsg", body, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
