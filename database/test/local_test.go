package db_test

import (
	"testing"

	db "github.com/fiatjaf/summadb/database"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestLocalDocs(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("local docs", func() {
		g.Before(func() {
			db.Erase()
			db.Start()
		})

		g.After(func() {
			db.End()
		})

		g.It("should save a _local doc", func() {
			Expect(db.SaveLocalDocAt("/banana/_local/w3hnks8hr", map[string]interface{}{
				"j": 23,
				"history": []map[string]interface{}{
					map[string]interface{}{"a": "b"},
				},
			})).To(Equal("0-1"))
			Expect(db.GetLocalDocJsonAt("/banana/_local/w3hnks8hr")).To(MatchJSON(`{
                "_id": "_local/w3hnks8hr",
                "_rev": "0-1",
                "j": 23,
                "history": [{"a": "b"}]
            }`))
		})

		g.It("should update a _local doc", func() {
			Expect(db.SaveLocalDocAt("/banana/_local/w3hnks8hr", map[string]interface{}{
				"j": 77,
				"history": []map[string]interface{}{
					map[string]interface{}{"a": "b"},
					map[string]interface{}{"c": "d"},
				},
				"_rev": "0-1",
				"_id":  "_local/w3hnks8hr",
			})).To(Equal("0-2"))
			Expect(db.GetLocalDocJsonAt("/banana/_local/w3hnks8hr")).To(MatchJSON(`{
                "_id": "_local/w3hnks8hr",
                "_rev": "0-2",
                "j": 77,
                "history": [{"a": "b"}, {"c": "d"}]
            }`))
		})

		g.It("should get the _local doc rev", func() {
			Expect(db.GetLocalDocRev("/banana/_local/w3hnks8hr")).To(BeEquivalentTo("0-2"))
		})
	})
}
