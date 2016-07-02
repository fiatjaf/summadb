package handle_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	db "github.com/fiatjaf/summadb/database"
	handle "github.com/fiatjaf/summadb/handle"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestGraphQL(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("simple graphql queries", func() {
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

		g.It("should query with type application/graphql", func() {
			r, _ = http.NewRequest("POST", "/_graphql", bytes.NewReader([]byte(`query {
              _val
              vehicles {
                _val
                car { land, water }
                airplane { land, air }
              }
            }`)))
			r.Header.Set("Content-Type", "application/graphql")
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "data": {
                "_val": "root",
                "vehicles": {
                  "_val": "things that move",
                  "car": {
                    "land": true,
                    "water": false
                  },
                  "airplane": {
                    "land": true,
                    "air": true
                  }
                }
              }
            }`))
			Expect(rec.Code).To(Equal(200))
		})

		g.It("should query with type application/json", func() {
			r, _ = http.NewRequest("POST", "/_graphql", bytes.NewReader([]byte(`{
              "query": "query { vehicles { runs:car { land, water }, flies:airplane { land, air } }, rootValue:_val }"
            }`)))
			r.Header.Set("Content-Type", "application/json")
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "data": {
                "rootValue": "root",
                "vehicles": {
                  "runs": {
                    "land": true,
                    "water": false
                  },
                  "flies": {
                    "land": true,
                    "air": true
                  }
                }
              }
            }`))
			Expect(rec.Code).To(Equal(200))
		})

		g.It("should query with type application/x-www-form-urlencoded", func() {
			form := url.Values{}
			form.Add("query", `
              query {
                v : vehicles {
                  rocket:flyingtorpedo { land }
                  car { land, air, water }
                  boat {
                    land {
                      _val
                      w:wot
                    }
                  }
                }
              }
            `)

			r, _ = http.NewRequest("POST", "/_graphql", strings.NewReader(form.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "data": {
                "v": {
                  "rocket": {
                    "land": null
                  },
                  "car": {
                    "land": true,
                    "air": false,
                    "water": false
                  },
                  "boat": {
                    "land": {
                      "_val": false,
                      "w": null
                    }
                  }
                }
              }
            }`))
			Expect(rec.Code).To(Equal(200))
		})

		g.It("graphql query on a subdb", func() {
			r, _ = http.NewRequest("POST", "/vehicles/_graphql", strings.NewReader(`query {
              desc:_val
              car { land, water }
              airplane { land, air }
            }`))
			r.Header.Set("Content-Type", "application/graphql")
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "data": {
                "desc": "things that move",
                "car": {
                  "land": true,
                  "water": false
                },
                "airplane": {
                  "land": true,
                  "air": true
                }
              }
            }`))
			Expect(rec.Code).To(Equal(200))
		})
	})

	g.Describe("authorized and restricted graphql queries", func() {
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

		g.It("should create some users and rules", func() {
			r, _ = http.NewRequest("POST", "/_users", bytes.NewReader([]byte(`{
				"name": "vehicles_user",
				"password": "12345678"
			}`)))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(201))
			rec = httptest.NewRecorder()

			r, _ = http.NewRequest("POST", "/_users", bytes.NewReader([]byte(`{
				"name": "boat_user",
				"password": "12345678"
			}`)))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(201))
			rec = httptest.NewRecorder()

			r, _ = http.NewRequest("PUT", "/vehicles/_security", bytes.NewReader([]byte(`{
				"_read": "vehicles_user"
			}`)))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
			rec = httptest.NewRecorder()

			r, _ = http.NewRequest("PUT", "/vehicles/boat/_security", bytes.NewReader([]byte(`{
				"_read": "boat_user"
			}`)))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
			rec = httptest.NewRecorder()

			r, _ = http.NewRequest("PUT", "/_security", bytes.NewReader([]byte(`{
				"_read": "no-one",
				"_write": "no-one",
				"_admin": "no-one"
			}`)))
			server.ServeHTTP(rec, r)
			Expect(rec.Code).To(Equal(200))
		})

		g.It("should be authorized for the paths where the user has immediate access", func() {
			r, _ = http.NewRequest("POST", "/vehicles/_graphql", strings.NewReader(`query {
              car { land, water }
            }`))
			r.Header.Set("Content-Type", "application/graphql")
			r.SetBasicAuth("vehicles_user", "12345678")
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "data": {
                "car": {
                  "land": true,
                  "water": false
                }
              }
            }`))
			Expect(rec.Code).To(Equal(200))
			rec = httptest.NewRecorder()

			r, _ = http.NewRequest("POST", "/vehicles/boat/_graphql", strings.NewReader(`query {
              air
            }`))
			r.Header.Set("Content-Type", "application/graphql")
			r.SetBasicAuth("boat_user", "12345678")
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "data": {
                "air": false
              }
            }`))
			Expect(rec.Code).To(Equal(200))
		})

		g.It("should be authorized for all paths below", func() {
			r, _ = http.NewRequest("POST", "/vehicles/boat/_graphql", strings.NewReader(`query {
              land
              water
            }`))
			r.Header.Set("Content-Type", "application/graphql")
			r.SetBasicAuth("vehicles_user", "12345678")
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "data": {
                "land": false,
                "water": true
              }
            }`))
			Expect(rec.Code).To(Equal(200))
			rec = httptest.NewRecorder()

			r, _ = http.NewRequest("POST", "/vehicles/boat/air/_graphql", strings.NewReader(`query {
              flies: _val
            }`))
			r.Header.Set("Content-Type", "application/graphql")
			r.SetBasicAuth("boat_user", "12345678")
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "data": {
                "flies": false
              }
            }`))
			Expect(rec.Code).To(Equal(200))
		})

		g.It("should be unauthorized for all paths above", func() {
			r, _ = http.NewRequest("POST", "/_graphql", strings.NewReader(`query {
              root: _val
              vehicles {
                car { land }
              }
            }`))
			r.Header.Set("Content-Type", "application/graphql")
			r.SetBasicAuth("vehicles_user", "12345678")
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "errors": [{
                "message": "_read permission for this path needed."
              }]
            }`))
			rec = httptest.NewRecorder()

			r, _ = http.NewRequest("POST", "/vehicles/_graphql", strings.NewReader(`query {
              boat { land, air }
            }`))
			r.Header.Set("Content-Type", "application/graphql")
			r.SetBasicAuth("boat_user", "12345678")
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "errors": [{
                "message": "_read permission for this path needed."
              }]
            }`))
		})

		g.It("should be unauthorized for neighbour paths also", func() {
			r, _ = http.NewRequest("POST", "/vehicles/car/_graphql", strings.NewReader(`query {
              land
              water
            }`))
			r.Header.Set("Content-Type", "application/graphql")
			r.SetBasicAuth("boat_user", "12345678")
			server.ServeHTTP(rec, r)
			Expect(rec.Body.String()).To(MatchJSON(`{
              "errors": [{
                "message": "_read permission for this path needed."
              }]
            }`))
		})
	})
}

func BenchmarkGraphQLQuery(b *testing.B) {
	rec = httptest.NewRecorder()
	server = handle.BuildHandler()

	db.Erase()
	db.Start()
	populateDB()

	for i := 0; i < b.N; i++ {
		/* BENCHMARK CODE */
		r, _ = http.NewRequest("POST", "/_graphql", bytes.NewReader([]byte(`query {
              root:_val
              vehicles {
                desc:_val
                car { land, water }
                airplane { land, air }
              }
            }`)))
		r.Header.Set("Content-Type", "application/graphql")
		server.ServeHTTP(rec, r)
		/* END BENCHMARK CODE */
	}

	db.End()
}
