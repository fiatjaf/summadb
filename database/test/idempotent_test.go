package db_test

import (
	"testing"

	db "github.com/fiatjaf/summadb/database"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestIdempotent(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("idempotent put", func() {
		g.Before(func() {
			db.Erase()
			db.Start()
		})

		g.After(func() {
			db.End()
		})

		g.It("should put a tree", func() {
			Expect(db.ReplaceTreeAt("/fruits/banana", map[string]interface{}{
				"colour":   "yellow",
				"hardness": "low",
				"_val":     "a fruit.",
			})).To(HavePrefix("1-"))
			Expect(db.GetValueAt("/fruits/banana")).To(BeEquivalentTo(`"a fruit."`))
			Expect(db.GetValueAt("/fruits/banana/colour")).To(BeEquivalentTo(`"yellow"`))
		})

		g.It("should replace it with a totally different object with arrays", func() {
			rev, err := db.ReplaceTreeAt("", map[string]interface{}{
				"what":    "numbers",
				"numbers": []interface{}{"zero", "one", "two", "three"},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(rev).To(HavePrefix("2-"))
			Expect(db.GetValueAt("/numbers/3")).To(BeEquivalentTo(`"three"`))
			_, err = db.GetValueAt("/numbers")
			Expect(err).To(HaveOccurred())
			Expect(db.GetTreeAt("")).To(Equal(map[string]interface{}{
				"what": value("numbers"),
				"numbers": map[string]interface{}{
					"0": value("zero"),
					"1": value("one"),
					"2": value("two"),
					"3": value("three"),
				},
			}))
			_, err = db.GetValueAt("/fruits")
			Expect(err).To(HaveOccurred())
		})

		g.It("should replace it again with a totally different object", func() {
			Expect(db.ReplaceTreeAt("/fruits/orange", map[string]interface{}{
				"_val":   nil,
				"colour": "orange",
			})).To(HavePrefix("1-"))
			Expect(db.GetTreeAt("")).To(Equal(map[string]interface{}{
				"what": value("numbers"),
				"numbers": map[string]interface{}{
					"0": value("zero"),
					"1": value("one"),
					"2": value("two"),
					"3": value("three"),
				},
				"fruits": map[string]interface{}{
					"orange": map[string]interface{}{
						"_val":   nil,
						"colour": value("orange"),
					},
				},
			}))
		})
	})
}
