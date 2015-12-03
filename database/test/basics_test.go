package db_test

import (
	"testing"

	db "github.com/fiatjaf/summadb/database"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestBasics(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("basics test", func() {

		g.Before(func() {
			db.Erase()
			db.Start()
		})

		g.After(func() {
			db.End()
		})

		g.It("should save a tree", func() {
			db.SaveTreeAt("/fruits/banana", map[string]interface{}{
				"colour":   "yellow",
				"hardness": "low",
				"_val":     "a fruit.",
			})
			Expect(db.GetValueAt("/fruits/banana")).To(BeEquivalentTo(`"a fruit."`))
			Expect(db.GetValueAt("/fruits/banana/colour")).To(BeEquivalentTo(`"yellow"`))
			Expect(db.GetValueAt("/fruits/banana/hardness")).To(BeEquivalentTo(`"low"`))
			_, err := db.GetValueAt("/fruits")
			Expect(err).To(HaveOccurred())
			Expect(db.GetTreeAt("/fruits")).To(Equal(map[string]interface{}{
				"banana": map[string]interface{}{
					"colour":   value("yellow"),
					"hardness": value("low"),
					"_val":     "a fruit.",
				},
			}))
		})

		g.It("should modify a subvalue of the tree", func() {
			db.SaveValueAt("/fruits/banana/colour", []byte(`"black-and-yellow"`))
			Expect(db.GetValueAt("/fruits/banana/colour")).To(BeEquivalentTo(`"black-and-yellow"`))
			Expect(db.GetValueAt("/fruits/banana/hardness")).To(BeEquivalentTo(`"low"`))
		})

		g.It("should add a value deeply nested in a tree that doesn't exists", func() {
			db.SaveValueAt("/fruits/mellon/season", []byte(`"spring"`))
			Expect(db.GetValueAt("/fruits/mellon/season")).To(BeEquivalentTo(`"spring"`))
			Expect(db.GetTreeAt("/fruits")).To(Equal(map[string]interface{}{
				"banana": map[string]interface{}{
					"colour":   value("black-and-yellow"),
					"hardness": value("low"),
					"_val":     "a fruit.",
				},
				"mellon": map[string]interface{}{
					"season": value("spring"),
				},
			}))
		})

		g.It("should add a tree deeply nested like the previous", func() {
			db.SaveTreeAt("/fruits/orange", map[string]interface{}{
				"colour":   "orange",
				"hardness": "medium",
				"_val":     "name == colour",
			})
			Expect(db.GetValueAt("/fruits/orange/colour")).To(BeEquivalentTo(`"orange"`))
			Expect(db.GetValueAt("/fruits/orange")).To(BeEquivalentTo(`"name == colour"`))
			Expect(db.GetTreeAt("/fruits/orange")).To(Equal(map[string]interface{}{
				"_val":     "name == colour",
				"colour":   value("orange"),
				"hardness": value("medium"),
			}))
		})

		g.It("should delete a key", func() {
			db.DeleteAt("/fruits/banana/colour")
			Expect(db.GetValueAt("/fruits/orange/colour")).To(BeEquivalentTo(`"orange"`))
			_, err := db.GetValueAt("/fruits/banana/colour")
			Expect(db.GetValueAt("/fruits/banana/colour/_deleted")).To(BeEquivalentTo(""))
			Expect(err).To(HaveOccurred())
		})

		g.It("should delete a value when setting it to null with a tree", func() {
			db.SaveTreeAt("/fruits/mellon", map[string]interface{}{
				"colour": "orange",
				"season": nil,
			})
			Expect(db.GetValueAt("/fruits/mellon/colour")).To(BeEquivalentTo(`"orange"`))
			_, err := db.GetValueAt("/fruits/mellon/season")
			Expect(err).To(HaveOccurred())

			db.SaveTreeAt("/fruits", map[string]interface{}{
				"mellon": nil,
			})
			_, err = db.GetValueAt("/fruits/mellon/colour")
			Expect(err).To(HaveOccurred())
			_, err = db.GetValueAt("/fruits/mellon")
			Expect(err).To(HaveOccurred())
		})

		g.It("should delete a tree", func() {
			db.DeleteAt("/fruits/banana")
			Expect(db.GetValueAt("/fruits/orange/colour")).To(BeEquivalentTo(`"orange"`))
			_, err := db.GetValueAt("/fruits/banana/hardness")
			Expect(err).To(HaveOccurred())

			rev, err := db.DeleteAt("/fruits")
			Expect(err).ToNot(HaveOccurred())
			Expect(rev).To(HavePrefix("9-"))
			_, err = db.GetValueAt("/fruits")
			Expect(err).To(HaveOccurred())
			Expect(db.GetValueAt("/fruits/orange/_deleted")).To(BeEquivalentTo(""))
			_, err = db.GetValueAt("/fruits/orange/colour")
			Expect(err).To(HaveOccurred())
			_, err = db.GetValueAt("/fruits/banana/hardness")
			Expect(err).To(HaveOccurred())
		})

		g.It("should error when fetching an untouched tree path", func() {
			_, err := db.GetTreeAt("/nowhere")
			Expect(err).To(HaveOccurred())
		})

		g.It("should return when fetching a deleted tree path", func() {
			tree, err := db.GetTreeAt("/fruits/banana")
			Expect(err).ToNot(HaveOccurred())
			empty := make(map[string]interface{})
			Expect(tree).To(BeEquivalentTo(empty))
		})
	})
}
