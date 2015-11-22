package handle_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/carbocation/interpose"

	db "github.com/fiatjaf/summadb/database"
	handle "github.com/fiatjaf/summadb/handle"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHTTP(t *testing.T) {
	RegisterFailHandler(Fail)
	log.Print("using db at " + db.GetDBFile())
	RunSpecs(t, "HTTP Suite")
}

var r *http.Request
var server *interpose.Middleware
var rec *httptest.ResponseRecorder

var _ = BeforeEach(func() {
	rec = httptest.NewRecorder()
	server = handle.BuildHTTPHandler()
})

func value(v string) map[string]interface{} {
	return map[string]interface{}{"_val": v}
}

func populateDB() {
	db.SaveTreeAt("", map[string]interface{}{
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
}

func StripRandom(JSON *bytes.Buffer) string {
	re := regexp.MustCompile("(\\d)-[\\w\\d]+")
	return re.ReplaceAllString(JSON.String(), "$1")
}
