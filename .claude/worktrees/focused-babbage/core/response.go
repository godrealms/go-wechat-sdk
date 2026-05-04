package core

import "fmt"

type Resp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (r *Resp) GetError() error {
	if r == nil || r.ErrCode == 0 {
		return nil
	}
	return fmt.Errorf("wechat api error %d: %s", r.ErrCode, r.ErrMsg)
}
