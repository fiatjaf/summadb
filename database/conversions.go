package database

import (
	"encoding/json"
)

func FromLevel(bs []byte) interface{} {
	var val interface{}
	json.Unmarshal(bs, &val)
	return val
}

func ToLevel(val interface{}) []byte {
	bs, _ := json.Marshal(val)
	return bs
}
