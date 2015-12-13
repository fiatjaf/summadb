package handle_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/fiatjaf/summadb/database"
	handle "github.com/fiatjaf/summadb/handle"
	"github.com/fiatjaf/summadb/handle/responses"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestAuthUsersACL(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("auth", func() {
		g.BeforeEach(func() {
			rec = httptest.NewRecorder()
			server = handle.BuildHTTPMux()
		})

		g.Before(func() {
			db.Erase()
			db.Start()
		})

		g.After(func() {
			db.End()
		})

		g.It("database should be in admin party", func() {
			r, _ = http.NewRequest("GET", "/_security", nil)
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(200))

			var res responses.Security
			json.Unmarshal(rec.Body.Bytes(), &res)
			Expect(res).To(BeEquivalentTo(responses.Security{"*", "*", "*"}))
		})

		g.It("should set some rules for database", func() {
			r, _ = http.NewRequest("PUT", "/_security", bytes.NewReader([]byte(`{
				"_read": "*",
				"_write": "myself",
				"_admin": "myself"
			}`)))
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(200))
			Expect(db.GetReadRuleAt("/")).To(Equal("*"))
			Expect(db.GetWriteRuleAt("/")).To(Equal("myself"))
			Expect(db.GetAdminRuleAt("/")).To(Equal("myself"))
		})

		g.It("should fail to set other rules (since _admin is a user)", func() {
			r, _ = http.NewRequest("PUT", "/_security", bytes.NewReader([]byte(`{
				"_read": "*",
				"_write": "myself, others",
				"_admin": "myself"
			}`)))
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(401))
			Expect(db.GetReadRuleAt("/")).To(Equal("*"))
			Expect(db.GetWriteRuleAt("/")).To(Equal("myself"))
			Expect(db.GetAdminRuleAt("/")).To(Equal("myself"))
		})

		g.It("should fail to create accounts due to the security policy", func() {
			r, _ = http.NewRequest("POST", "/_users", bytes.NewReader([]byte(`{
				"name": "myself",
				"password": "12345678"
			}`)))
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(401))
		})

		g.It("should change the security policy by cheating, then create accounts", func() {
			db.SetRulesAt("/", map[string]interface{}{"_write": "*"})

			r, _ = http.NewRequest("POST", "/_users", bytes.NewReader([]byte(`{
				"name": "myself",
				"password": "12345678"
			}`)))
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(201))
			db.SetRulesAt("/", map[string]interface{}{"_write": "myself, others"})
		})

		g.It("should create accounts, now using the recently created and authorized user", func() {
			r, _ = http.NewRequest("POST", "/_users", bytes.NewReader([]byte(`{
				"name": "others",
				"password": "qwerty"
			}`)))
			r.SetBasicAuth("myself", "12345678")
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(201))
		})

		g.It("another account created", func() {
			r, _ = http.NewRequest("POST", "/_users", bytes.NewReader([]byte(`{
				"name": "bob",
				"password": "gki48dh3w"
			}`)))
			r.SetBasicAuth("myself", "12345678")
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(201))
		})

		g.It("should fail to create document without a user", func() {
			r, _ = http.NewRequest("PUT", "/a/b/c", bytes.NewReader([]byte(`{
				"doc": "iuiwebsd",
				"val": "woiernhoq234"
			}`)))
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(401))
		})

		g.It("should fail to create document with an unallowed user", func() {
			r, _ = http.NewRequest("PUT", "/a/b/c", bytes.NewReader([]byte(`{
				"doc": "iuiwebsd",
				"val": "woiernhoq234"
			}`)))
			r.SetBasicAuth("bob", "gki48dh3w")
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(401))
		})

		g.It("should succeed to created document with a correct user", func() {
			r, _ = http.NewRequest("PUT", "/a/b/c", bytes.NewReader([]byte(`{
				"doc": "iuiwebsd",
				"val": "woiernhoq234"
			}`)))
			r.SetBasicAuth("others", "qwerty")
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(201))
		})

		g.It("should change only read permission for a subpath", func() {
			r, _ = http.NewRequest("PUT", "/a/b/_security", bytes.NewReader([]byte(`{
				"_read": "bob"
			}`)))
			r.SetBasicAuth("myself", "12345678")
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(200))
		})

		g.It("should be able to read with wrong user because of upper rule", func() {
			r, _ = http.NewRequest("GET", "/a/b/c", nil)
			r.SetBasicAuth("myself", "12345678")
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(200))
		})

		g.It("remove upper rule that allowed everybody", func() {
			r, _ = http.NewRequest("PUT", "/_security", bytes.NewReader([]byte(`{
				"_read": ""
			}`)))
			r.SetBasicAuth("myself", "12345678")
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(200))
		})

		g.It("shouldn't be able to read with wrong user", func() {
			r, _ = http.NewRequest("GET", "/a/b/c", nil)
			r.SetBasicAuth("myself", "12345678")
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(401))
		})

		g.It("should be able to read with correct user", func() {
			r, _ = http.NewRequest("GET", "/a/b/c", nil)
			r.SetBasicAuth("bob", "gki48dh3w")
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(200))
		})

		g.It("change the rule again, now a lower rule will allow a different user", func() {
			r, _ = http.NewRequest("PUT", "/a/b/c/_security", bytes.NewReader([]byte(`{
				"_read": "others"
			}`)))
			r.SetBasicAuth("myself", "12345678")
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(200))
		})

		g.It("should be able to read with this different user", func() {
			r, _ = http.NewRequest("GET", "/a/b/c", nil)
			r.SetBasicAuth("others", "qwerty")
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(200))
		})

		g.It("and again with the same user that was allowed first", func() {
			r, _ = http.NewRequest("GET", "/a/b/c", nil)
			r.SetBasicAuth("bob", "gki48dh3w")
			server.ServeHTTP(rec, r)

			Expect(rec.Code).To(Equal(200))
		})
	})
}
