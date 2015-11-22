package handle_test

import (
	"encoding/json"
	"net/http"

	db "github.com/fiatjaf/summadb/database"
	handle "github.com/fiatjaf/summadb/handle"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("changes feed", func() {
	Context("changes HTTP", func() {
		It("should erase the db and prepopulate", func() {
			Expect(db.Erase()).To(Succeed())
			populateDB()
		})

		It("should hit the changes endpoint", func() {
			r, _ = http.NewRequest("GET", "/blabla/_changes", nil)
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
		})

		It("should return a list of changes for a sub db", func() {
			r, _ = http.NewRequest("GET", "/vehicles/_changes", nil)
			server.ServeHTTP(rec, r)

			var res handle.ChangesResponse
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

			var res handle.ChangesResponse
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

			var res handle.ChangesResponse
			json.Unmarshal(rec.Body.Bytes(), &res)
			Expect(res.LastSeq).To(BeNumerically(">", 0))
			Expect(res.LastSeq).To(BeNumerically("<=", 21))
			Expect(res.Results).To(HaveLen(3))
			Expect(res.Results[0].Changes[0].Rev).To(HavePrefix("1-"))
			Expect(res.Results[2].Seq).To(Equal(res.LastSeq))
		})
	})
})
