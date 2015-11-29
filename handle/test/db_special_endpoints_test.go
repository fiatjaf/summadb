package handle_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sort"

	db "github.com/fiatjaf/summadb/database"
	responses "github.com/fiatjaf/summadb/handle/responses"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("all docs", func() {
	Context("_all_docs HTTP", func() {
		var rev string

		It("should erase the db and prepopulate", func() {
			Expect(db.Erase()).To(Succeed())
			populateDB()
		})

		It("should return _all_docs for a sub db", func() {
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

		It("should return all_docs with include_docs for another sub db", func() {
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

		It("should _bulk_get", func() {
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
			Expect(res.Responses[0].Docs[0].Ok).To(BeNil())
			Expect(res.Responses[0].Docs[0].Error).ToNot(BeNil())
			Expect(res.Responses[1].Docs[0].Ok).ToNot(BeNil())
			Expect(res.Responses[1].Docs[0].Error).To(BeNil())
			doc := *res.Responses[1].Docs[0].Ok
			id, _ := doc["_id"]
			irev, _ := doc["_rev"]
			rev = irev.(string)
			water, _ := doc["water"]
			Expect(id.(string)).To(Equal("car"))
			Expect(res.Responses[1].Id).To(Equal(id))
			Expect(water).To(BeEquivalentTo(value(false)))
			Expect(res.Responses[2].Docs[0].Ok).To(BeNil())
			Expect(res.Responses[0].Docs[0].Error).ToNot(BeNil())
		})
	})
})
