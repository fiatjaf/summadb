package handle_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/fiatjaf/summadb/database"
	handle "github.com/fiatjaf/summadb/handle"
	responses "github.com/fiatjaf/summadb/handle/responses"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestChanges(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("changes feed", func() {
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

		var intermediary uint64
		var last uint64

		g.It("should return a list of changes for a sub db", func() {
			r, _ = http.NewRequest("GET", "/vehicles/_changes", nil)
			server.ServeHTTP(rec, r)

			var res responses.Changes
			json.Unmarshal(rec.Body.Bytes(), &res)
			Expect(res.LastSeq).To(BeNumerically(">", 0))
			Expect(res.LastSeq).To(BeNumerically("<=", 21))
			Expect(res.Results).To(HaveLen(3))
			Expect(res.Results[0].Changes[0].Rev).To(HavePrefix("1-"))
			Expect(res.Results[2].Seq).To(Equal(res.LastSeq))
		})

		g.It("should return a list of changes for the global db", func() {
			r, _ = http.NewRequest("GET", "/_changes", nil)
			server.ServeHTTP(rec, r)

			var res responses.Changes
			json.Unmarshal(rec.Body.Bytes(), &res)
			Expect(res.LastSeq).To(BeNumerically(">", 0))
			Expect(res.LastSeq).To(BeNumerically("<=", 21))
			Expect(res.Results).To(HaveLen(3))
			Expect(res.Results[0].Changes[0].Rev).To(HavePrefix("1-"))
			Expect(res.Results[2].Seq).To(Equal(res.LastSeq))
		})

		g.It("should return a list of changes for another sub db", func() {
			r, _ = http.NewRequest("GET", "/vehicles/boat/_changes", nil)
			server.ServeHTTP(rec, r)

			var res responses.Changes
			json.Unmarshal(rec.Body.Bytes(), &res)
			Expect(res.LastSeq).To(BeNumerically(">", 0))
			Expect(res.LastSeq).To(BeNumerically("<=", 21))
			Expect(res.Results).To(HaveLen(3))
			Expect(res.Results[0].Changes[0].Rev).To(HavePrefix("1-"))
			Expect(res.Results[2].Seq).To(Equal(res.LastSeq))

			intermediary = res.Results[1].Seq
			last = res.Results[2].Seq
		})

		g.It("should honour 'since'", func() {
			r, _ = http.NewRequest("GET", fmt.Sprintf("/vehicles/boat/_changes?since=%d", intermediary), nil)
			server.ServeHTTP(rec, r)

			var res responses.Changes
			json.Unmarshal(rec.Body.Bytes(), &res)
			Expect(res.Results).To(HaveLen(1))
			Expect(res.Results[0].Seq).To(Equal(last))
			Expect(res.Results[0].Seq).To(Equal(res.LastSeq))
		})
	})
}
