package handle_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/fiatjaf/summadb/database"
	handle "github.com/fiatjaf/summadb/handle"
	responses "github.com/fiatjaf/summadb/handle/responses"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestAuth(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("basics", func() {
		g.BeforeEach(func() {
			rec = httptest.NewRecorder()
			server = handle.BuildHTTPMux()
		})

		g.Before(func() {
			db.Erase()
			db.Start()
			populateDB()
		})

		g.After(func() {
			db.End()
		})

		var rev string

		g.It("should get an empty doc", func() {
			r, _ = http.NewRequest("POST", "/nothing/here", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "error": "not_found",
              "reason": "missing"
            }`))
			Expect(rec.Code).To(Equal(404))
		})
		})
