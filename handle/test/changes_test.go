package handle_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	db "github.com/fiatjaf/summadb/database"
	responses "github.com/fiatjaf/summadb/handle/responses"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("changes feed", func() {
	Context("changes HTTP", func() {
		var intermediary uint64
		var last uint64

		It("should erase the db and prepopulate", func() {
			Expect(db.Erase()).To(Succeed())
			populateDB()
		})

		It("should return a list of changes for a sub db", func() {
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

		It("should return a list of changes for the global db", func() {
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

		It("should return a list of changes for another sub db", func() {
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

		It("should honour 'since'", func() {
			r, _ = http.NewRequest("GET", fmt.Sprintf("/vehicles/boat/_changes?since=%d", intermediary), nil)
			server.ServeHTTP(rec, r)

			var res responses.Changes
			json.Unmarshal(rec.Body.Bytes(), &res)
			Expect(res.Results).To(HaveLen(1))
			Expect(res.Results[0].Seq).To(Equal(last))
			Expect(res.Results[0].Seq).To(Equal(res.LastSeq))
		})
	})
})
