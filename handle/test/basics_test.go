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

func TestBasics(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("basics", func() {
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

		g.It("should get an empty doc", func() {
			r, _ = http.NewRequest("GET", "/nothing/here", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "error": "not_found",
              "reason": "missing"
            }`))
			Expect(rec.Code).To(Equal(404))
		})

		g.It("should create a new doc", func() {
			body := `{"a": "one", "dfg": {"many": 3, "which": ["d", "f", "g"]}}`
			jsonbody := []byte(body)
			r, _ = http.NewRequest("PUT", "/something/here", bytes.NewReader(jsonbody))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(201))
			var resp responses.Success
			Expect(json.Unmarshal(rec.Body.Bytes(), &resp)).To(Succeed())
			Expect(resp.Ok).To(BeTrue())
			Expect(resp.Id).To(Equal("here"))
			Expect(resp.Rev).To(HavePrefix("1-"))
			rev = resp.Rev
		})

		g.It("should fetch a subfield", func() {
			r, _ = http.NewRequest("GET", "/something/here/a/_val", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
			Expect(rec.Body.String()).To(Equal(`"one"`))
		})

		g.It("should fetch a subrev", func() {
			r, _ = http.NewRequest("GET", "/something/here/dfg/_rev", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
			rev = rec.Body.String()
			Expect(rev).To(HavePrefix(`1-`))
		})

		g.It("should fetch a subtree", func() {
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

		g.It("should get the newest _rev for a path", func() {
			r, _ = http.NewRequest("GET", "/something/here/dfg/many/_rev", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
			rev = rec.Body.String()
		})

		g.It("should delete a key (providing rev)", func() {
			r, _ = http.NewRequest("DELETE", "/something/here/dfg/many", nil)
			r.Header.Set("If-Match", rev)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
			var resp responses.Success
			Expect(json.Unmarshal(rec.Body.Bytes(), &resp)).To(Succeed())
			Expect(resp.Ok).To(BeTrue())
			Expect(resp.Id).To(Equal("many"))
			Expect(resp.Rev).To(HavePrefix("2-"))
		})

		g.It("should fail to fetch deleted key", func() {
			r, _ = http.NewRequest("GET", "/something/here/dfg/many", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(404))
		})

		g.It("should fail to delete a special key", func() {
			r, _ = http.NewRequest("DELETE", "/something/here/_rev", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(400))
		})

		g.It("should fail to update a special key", func() {
			body := `{"a": "one", "dfg": {"many": 3, "which": ["d", "f", "g"]}}`
			jsonbody := []byte(body)
			r, _ = http.NewRequest("PATCH", "/something/_rev", bytes.NewReader(jsonbody))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(400))
		})

		g.It("should fail to update a key without providing a _rev", func() {
			body := `{"was": "another thing"}`
			jsonbody := []byte(body)
			r, _ = http.NewRequest("PATCH", "/something", bytes.NewReader(jsonbody))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(409))
		})

		g.It("should fail to update a key providing a wrong _rev", func() {
			body := `{"_rev": "2-389247isdbf", "was": "another thing"}`
			jsonbody := []byte(body)
			r, _ = http.NewRequest("PATCH", "/something/here/dfg", bytes.NewReader(jsonbody))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(409))
		})

		g.It("should fail to update a deleted key when providing a mismatching revs", func() {
			body := `{"_rev": "3-1asd623a5", "was": "another thing"}`
			jsonbody := []byte(body)
			r, _ = http.NewRequest("PATCH", "/something/here/dfg?rev=7-sdf98h435", bytes.NewReader(jsonbody))
			r.Header.Set("If-Match", rev)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(400))
		})

		g.It("should fail to patch an untouched path", func() {
			jsonbody := []byte(`{"1": 2}`)
			r, _ = http.NewRequest("PATCH", "/nowhere", bytes.NewReader(jsonbody))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(404))
		})

		g.It("should get the newest _rev for a path", func() {
			r, _ = http.NewRequest("GET", "/something/here/_rev", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
			rev = rec.Body.String()
		})

		g.It("should delete a path providing the correct rev", func() {
			log.Print("sending rev ", rev)
			r, _ = http.NewRequest("DELETE", "/something/here?rev="+rev, nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
		})

		g.It("should fail to patch a deleted path", func() {
			jsonbody := []byte(`{"1": 2}`)
			r, _ = http.NewRequest("PATCH", "/something/here", bytes.NewReader(jsonbody))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(404))
		})

		g.It("should put a tree on a deleted path without providing any rev", func() {
			body := `{"was": {"before": "another thing", "long_before": "a different thing"}}`
			jsonbody := []byte(body)
			r, _ = http.NewRequest("PUT", "/something/here", bytes.NewReader(jsonbody))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(201))

			var resp responses.Success
			json.Unmarshal(rec.Body.Bytes(), &resp)
			rev = resp.Rev
		})

		g.It("should update a subpath with the rev of a parent", func() {
			body := `{"was": {"before": "still another thing"}}`
			jsonbody := []byte(body)
			r, _ = http.NewRequest("PATCH", "/something/here?rev="+rev, bytes.NewReader(jsonbody))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))

			var resp responses.Success
			json.Unmarshal(rec.Body.Bytes(), &resp)
			rev = resp.Rev
		})

		g.It("should delete a subpath with the rev of a parent (using a tree)", func() {
			body := `{"was": {"long_before": null}, "_rev": "` + rev + `"}`
			jsonbody := []byte(body)
			r, _ = http.NewRequest("PATCH", "/something/here", bytes.NewReader(jsonbody))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
		})

		g.It("should get the rev of the something path", func() {
			r, _ = http.NewRequest("GET", "/something/_rev", nil)
			server.ServeHTTP(rec, r)
			rev = rec.Body.String()
		})

		g.It("should have the correct tree in the end", func() {
			r, _ = http.NewRequest("GET", "/something", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
			Expect(rec.Body.String()).To(MatchJSON(`{
              "_id": "something",
              "_rev": "` + rev + `",
              "here": {
                "was": {
                  "before": {
                    "_val": "still another thing"
                  }
                }
              }
            }`))
		})
	})
}
