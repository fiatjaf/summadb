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
}
