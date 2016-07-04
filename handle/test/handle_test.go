package handle_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"

	db "github.com/fiatjaf/summadb/database"
)

var r *http.Request
var server http.Handler
var rec *httptest.ResponseRecorder

func value(v interface{}) map[string]interface{} {
	return map[string]interface{}{"_val": v}
}

func populateDB() (err error) {
	_, err = db.SaveTreeAt("", map[string]interface{}{
		"_val": "root",
		"vehicles": map[string]interface{}{
			"_val": "things that move",
			"car": map[string]interface{}{
				"land":  true,
				"air":   false,
				"water": false,
			},
			"airplane": map[string]interface{}{
				"land":  true,
				"air":   true,
				"water": false,
			},
			"boat": map[string]interface{}{
				"land":  false,
				"air":   false,
				"water": true,
			},
		},
		"animals": []map[string]interface{}{
			map[string]interface{}{
				"name": "bird",
			},
			map[string]interface{}{
				"name": "dog",
			},
			map[string]interface{}{
				"name": "cow",
			},
		},
	})
	return err
}

func StripRandom(JSON *bytes.Buffer) string {
	re := regexp.MustCompile("(\\d)-[\\w\\d]+")
	return re.ReplaceAllString(JSON.String(), "$1")
}
