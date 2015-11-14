package handle_test

import (
	"bytes"
	"encoding/json"
	"net/http"

	db "github.com/fiatjaf/summadb/database"
	handle "github.com/fiatjaf/summadb/handle"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("server", func() {
	Context("open", func() {
		var rev string

		It("should erase the db", func() {
			Expect(db.Erase()).To(Succeed())
		})

		It("gets an empty doc", func() {
			r, _ = http.NewRequest("GET", "/nothing/here", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "error": "not_found",
              "reason": "missing"
            }`))
			Expect(rec.Code).To(Equal(404))
		})

		It("creates a new doc", func() {
			body := `{"a": "one", "dfg": {"many": 3, "which": ["d", "f", "g"]}}`
			jsonbody := []byte(body)
			r, _ = http.NewRequest("PUT", "/something/here", bytes.NewReader(jsonbody))
			r.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(201))
			var resp handle.Success
			Expect(json.Unmarshal(rec.Body.Bytes(), &resp)).To(Succeed())
			Expect(resp.Ok).To(BeTrue())
			Expect(resp.Id).To(Equal("here"))
			Expect(resp.Rev).To(HavePrefix("1-"))
			rev = resp.Rev
		})

		It("fetches a subfield", func() {
			r, _ = http.NewRequest("GET", "/something/here/a/_val", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
			Expect(rec.Body.String()).To(Equal(`"one"`))
		})

		It("fetches a subrev", func() {
			r, _ = http.NewRequest("GET", "/something/here/dfg/_rev", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
			rev = rec.Body.String()
			Expect(rev).To(HavePrefix(`1-`))
		})

		It("fetches a subtree", func() {
			r, _ = http.NewRequest("GET", "/something/here/dfg", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
			Expect(rec.Body.String()).To(MatchJSON(`{
              "_id": "dfg",
              "_rev": "` + rev + `",
              "many": {"_val": 3},
              "which": {
                "0": {"_val": "d"},
                "1": {"_val": "f"},
                "2": {"_val": "g"}
              }
            }`))
		})

		It("deletes a key", func() {
			r, _ = http.NewRequest("DELETE", "/something/here/dfg/many", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
			var resp handle.Success
			Expect(json.Unmarshal(rec.Body.Bytes(), &resp)).To(Succeed())
			Expect(resp.Ok).To(BeTrue())
			Expect(resp.Id).To(Equal("many"))
			Expect(resp.Rev).To(HavePrefix("2-"))
		})

		It("fails to fetch deleted key", func() {
			r, _ = http.NewRequest("GET", "/something/here/dfg/many", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(404))
		})

		It("fails to update or delete a special key", func() {
			r, _ = http.NewRequest("DELETE", "/something/here/dfg/many", nil)

		})
	})
})
