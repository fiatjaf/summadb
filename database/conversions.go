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
	var bs []byte
	switch v := val.(type) {
	case []byte:
		bs, _ = json.Marshal(string(v))
	default:
		bs, _ = json.Marshal(v)
	}
	return bs
}
