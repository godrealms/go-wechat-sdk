package types

import "encoding/json"

type MchID struct {
	Mchid string `json:"mchid"`
}

func (t *MchID) ToString() string {
	marshal, err := json.Marshal(t)
	if err != nil {
		return "<marshal error: " + err.Error() + ">"
	}
	return string(marshal)
}
