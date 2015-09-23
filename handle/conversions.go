package handle

import (
	"encoding/json"
)

func fromLevel(bs []byte) interface{} {
	var val interface{}
	json.Unmarshal(bs, &val)
	return val
}

func toLevel(val interface{}) []byte {
	bs, _ := json.Marshal(val)
	return bs
}
