package types

import "encoding/json"

type MchID struct {
	Mchid string `json:"mchid"`
}

func (t *MchID) ToString() string {
	marshal, _ := json.Marshal(t)
	return string(marshal)
}
