package handle_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"

	db "github.com/fiatjaf/summadb/database"
	handle "github.com/fiatjaf/summadb/handle"
	responses "github.com/fiatjaf/summadb/handle/responses"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestCouchDBSpecialEndpoints(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("couchdb db special endpoints", func() {
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
		var oldrev string
		var id string

		g.It("should return _all_docs for a sub db", func() {
			r, _ = http.NewRequest("GET", "/vehicles/_all_docs", nil)
			server.ServeHTTP(rec, r)

			var res responses.AllDocs
			json.Unmarshal(rec.Body.Bytes(), &res)
			Expect(res.Rows).To(HaveLen(3))
			Expect(res.Rows[2].Id).To(Equal(res.Rows[2].Key))
			rev, _ := res.Rows[2].Value.(map[string]interface{})["rev"]
			Expect(rev).To(HavePrefix("1-"))

			keys := []string{res.Rows[0].Id, res.Rows[1].Key, res.Rows[2].Id}
			sort.Strings(keys)
			Expect(keys).To(Equal([]string{"airplane", "boat", "car"}))
		})

		g.It("should return all_docs with include_docs for another sub db", func() {
			r, _ = http.NewRequest("GET", "/vehicles/airplane/_all_docs?include_docs=true", nil)
			server.ServeHTTP(rec, r)

			var res responses.AllDocs
			json.Unmarshal(rec.Body.Bytes(), &res)
			Expect(res.Rows).To(HaveLen(3))

			docid, _ := res.Rows[1].Doc["_id"]
			Expect(res.Rows[1].Key).To(Equal(docid))

			rev, _ := res.Rows[1].Value.(map[string]interface{})["rev"]
			docrev, _ := res.Rows[1].Doc["_rev"]
			Expect(rev).To(Equal(docrev))

			keys := []string{res.Rows[0].Id, res.Rows[1].Key, res.Rows[2].Id}
			sort.Strings(keys)
			Expect(keys).To(Equal([]string{"air", "land", "water"}))

			docs := map[string]interface{}{
				res.Rows[0].Key: res.Rows[0].Doc,
				res.Rows[1].Key: res.Rows[1].Doc,
				res.Rows[2].Key: res.Rows[2].Doc,
			}
			Expect(docs).To(HaveKey("air"))
			Expect(res.Rows[0].Doc).To(HaveKey("_rev"))
			Expect(res.Rows[0].Doc).To(HaveKey("_val"))
			Expect(res.Rows[0].Doc).To(HaveKeyWithValue("_id", res.Rows[0].Id))
		})

		g.It("should _bulk_get", func() {
			r, _ = http.NewRequest("POST", "/vehicles/_bulk_get", bytes.NewReader([]byte(`{
                "docs": [
                    {"id": "nonexisting-doc"},
                    {"id": "car"},
                    {"_id": "airplane"}
                ]
            }`)))
			server.ServeHTTP(rec, r)

			var res responses.BulkGet
			json.Unmarshal(rec.Body.Bytes(), &res)
			Expect(res.Results[0].Docs[0].Ok).To(BeNil())
			Expect(res.Results[0].Docs[0].Error).ToNot(BeNil())
			Expect(res.Results[1].Docs[0].Ok).ToNot(BeNil())
			Expect(res.Results[1].Docs[0].Error).To(BeNil())
			doc := *res.Results[1].Docs[0].Ok
			id, _ := doc["_id"]
			irev, _ := doc["_rev"]
			rev = irev.(string)
			water, _ := doc["water"]
			Expect(id.(string)).To(Equal("car"))
			Expect(res.Results[1].Id).To(Equal(id))
			Expect(water).To(BeEquivalentTo(value(false)))
			Expect(res.Results[2].Docs[0].Ok).To(BeNil())
			Expect(res.Results[0].Docs[0].Error).ToNot(BeNil())
		})

		g.It("should post docs with _bulk_docs", func() {
			r, _ = http.NewRequest("POST", "/vehicles/_bulk_docs", bytes.NewReader([]byte(`{
                "docs": [
                    {"everywhere": true},
                    {"_id": "car", "_rev": "`+rev+`", "space": false, "land": true},
                    {"_id": "airplane", "nowhere": false},
                    {"_id": "_local/.abchtru", "replication+data": "k"}
                ]
            }`)))
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(201))
			var res []responses.BulkDocsResult
			json.Unmarshal(rec.Body.Bytes(), &res)

			Expect(res).To(HaveLen(4))
			Expect(res[0].Error).To(Equal(""))
			Expect(res[0].Ok).To(Equal(true))
			Expect(res[0].Rev).To(HavePrefix("1-"))
			id = res[0].Id
			Expect(res[1].Id).To(Equal("car"))
			prevn, _ := strconv.Atoi(strings.Split(rev, "-")[0])
			Expect(res[1].Rev).To(HavePrefix(fmt.Sprintf("%d-", prevn+1)))
			oldrev = rev
			rev = res[1].Rev
			cfe := responses.ConflictError()
			Expect(res[2].Error).To(Equal(cfe.Error))
			Expect(res[3].Id).To(Equal("_local/.abchtru"))
			Expect(res[3].Ok).To(Equal(true))
		})

		g.It("should have the correct docs saved", func() {
			Expect(db.GetValueAt("/vehicles/" + id + "/everywhere")).To(BeEquivalentTo("true"))
			Expect(db.GetValueAt("/vehicles/_local%2F.abchtru/replication%2Bdata")).To(BeEquivalentTo(`"k"`))
		})

		g.It("shouldn't show _local docs on _all_docs", func() {
			r, _ = http.NewRequest("GET", "/vehicles/_all_docs", nil)
			server.ServeHTTP(rec, r)
			var res responses.AllDocs
			json.Unmarshal(rec.Body.Bytes(), &res)
			Expect(res.Rows).To(HaveLen(4))
		})

		g.It("should return the _revs_diff", func() {
			r, _ = http.NewRequest("POST", "/vehicles/_revs_diff", bytes.NewReader([]byte(`{
                "everywhere": ["2-invalidrev"],
                "car": ["`+oldrev+`", "`+rev+`", "1-invalidrev"],
                "airplane": ["1-nonexisting"]
            }`)))
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(200))
			var res map[string]responses.RevsDiffResult
			json.Unmarshal(rec.Body.Bytes(), &res)

			everywhere, _ := res["everywhere"]
			car, _ := res["car"]
			airplane, _ := res["airplane"]
			Expect(everywhere.Missing).To(Equal([]string{"2-invalidrev"}))
			Expect(car.Missing).To(Equal([]string{oldrev, "1-invalidrev"}))
			Expect(airplane.Missing).To(Equal([]string{"1-nonexisting"}))
		})
	})
}
