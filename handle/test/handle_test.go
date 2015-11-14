package handle_test

import (
	"log"
	"net/http"
	"net/http/httptest"
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
