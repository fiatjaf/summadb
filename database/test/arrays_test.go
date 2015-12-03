package db_test

import (
	"testing"

	db "github.com/fiatjaf/summadb/database"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestArrays(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("array values", func() {

		g.Before(func() {
			db.Erase()
			db.Start()
		})

		g.After(func() {
			db.End()
		})

		g.It("should save a tree with a simple array", func() {
			rev, err := db.SaveTreeAt("", map[string]interface{}{
				"numbers": []interface{}{"zero", "one", "two", "three"},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(rev).To(HavePrefix("1-"))
			Expect(db.GetValueAt("/numbers/0")).To(BeEquivalentTo(`"zero"`))
			Expect(db.GetValueAt("/numbers/3")).To(BeEquivalentTo(`"three"`))
			_, err = db.GetValueAt("/numbers")
			Expect(err).To(HaveOccurred())
			Expect(db.GetTreeAt("/numbers")).To(Equal(map[string]interface{}{
				"0": value("zero"),
				"1": value("one"),
				"2": value("two"),
				"3": value("three"),
			}))
		})

		g.It("should save a tree with a complex array", func() {
			rev, err := db.SaveTreeAt("", map[string]interface{}{
				"letters": []interface{}{
					map[string]interface{}{
						"name":       "á",
						"variations": []interface{}{"a", "A"},
					},
					map[string]interface{}{
						"name":       "bê",
						"variations": []interface{}{"b", "B"},
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(rev).To(HavePrefix("2-"))
			Expect(db.GetValueAt("/letters/0/name")).To(BeEquivalentTo(`"á"`))
			Expect(db.GetValueAt("/letters/1/variations/1")).To(BeEquivalentTo(`"B"`))
			_, err = db.GetValueAt("/letters/0/variations")
			Expect(err).To(HaveOccurred())
			Expect(db.GetTreeAt("/letters")).To(Equal(map[string]interface{}{
				"0": map[string]interface{}{
					"name": value("á"),
					"variations": map[string]interface{}{
						"0": value("a"),
						"1": value("A"),
					},
				},
				"1": map[string]interface{}{
					"name": value("bê"),
					"variations": map[string]interface{}{
						"0": value("b"),
						"1": value("B"),
					},
				},
			}))
		})
	})
}
