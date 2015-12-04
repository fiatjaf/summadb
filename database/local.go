package database

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func SaveLocalDocAt(path string, doc map[string]interface{}) (newrev string, err error) {
	locals := db.Sub(LOCAL_STORE)

	var jsondoc []byte
	if irev, ok := doc["_rev"]; ok {
		newrev = newLocalRev(irev.(string))
	} else {
		newrev = "0-1"
	}
	doc["_rev"] = newrev
	doc["_id"] = "_local/" + strings.Split(path, "/_local/")[1]
	jsondoc, err = json.Marshal(doc)
	if err != nil {
		return "", err
	}
	err = locals.Put([]byte(path), jsondoc, nil)
	if err != nil {
		return "", err
	}
	return newrev, nil
}

func GetLocalDocJsonAt(path string) (json []byte, err error) {
	locals := db.Sub(LOCAL_STORE)
	return locals.Get([]byte(path), nil) // the error here should be handled the next
}

func GetLocalDocRev(path string) string {
	locals := db.Sub(LOCAL_STORE)

	var doc map[string]interface{}
	bs, _ := locals.Get([]byte(path), nil) // the error here should be handled the next
	err := json.Unmarshal(bs, &doc)
	if err != nil {
		return "0-0"
	}

	if irev, ok := doc["_rev"]; ok {
		return irev.(string)
	}
	return ""
}

func newLocalRev(rev string) (newrev string) {
	n, err := strconv.Atoi(strings.Split(rev, "-")[1])
	if err != nil {
		return "0-0"
	}
	return fmt.Sprintf("0-%d", n+1)
}
