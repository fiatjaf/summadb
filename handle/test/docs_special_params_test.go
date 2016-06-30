package handle_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/fiatjaf/summadb/database"
	handle "github.com/fiatjaf/summadb/handle"
	responses "github.com/fiatjaf/summadb/handle/responses"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestCouchDBDocsSpecial(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("couchdb documents special endpoints", func() {
		g.BeforeEach(func() {
			rec = httptest.NewRecorder()
			server = handle.BuildHandler()
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

		g.It("should change some values to generate revs", func() {
			brev, _ := db.GetValueAt("/vehicles/boat/air/_rev")
			r, _ = http.NewRequest("PUT", "/vehicles/boat/air/_val", bytes.NewReader([]byte("true")))
			r.Header.Add("If-Match", string(brev))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(201))

			var res responses.Success
			json.Unmarshal(rec.Body.Bytes(), &res)
			rev = res.Rev
		})

		g.It("once more:", func() {
			r, _ = http.NewRequest("PATCH", "/vehicles/boat/air", bytes.NewReader([]byte(`{
                "_rev": "`+rev+`",
                "really?": false
            }`)))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
		})

		g.It("should fetch some key with the _revisions special field", func() {
			r, _ = http.NewRequest("GET", "/vehicles?revs=true", nil)
			server.ServeHTTP(rec, r)

			var res map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &res)

			irevisions, ok := res["_revisions"]
			Expect(ok).To(Equal(true))
			revisions := irevisions.(map[string]interface{})
			start, _ := revisions["start"]
			ids, _ := revisions["ids"]
			Expect(start).To(BeEquivalentTo(1))
			Expect(ids).To(HaveLen(3))
		})

		g.It("should fetch some key with the _revs_info special field", func() {
			r, _ = http.NewRequest("GET", "/vehicles/boat?revs_info=true", nil)
			server.ServeHTTP(rec, r)

			var res map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &res)

			irevsinfo, ok := res["_revs_info"]
			Expect(ok).To(Equal(true))
			revsinfo := irevsinfo.([]interface{})
			Expect(revsinfo).To(HaveLen(3))

			first := revsinfo[0].(map[string]interface{})
			second := revsinfo[1].(map[string]interface{})
			status, _ := first["status"]
			Expect(status).To(BeEquivalentTo("available"))
			status, _ = second["status"]
			Expect(status).To(BeEquivalentTo("missing"))
		})
	})
}
