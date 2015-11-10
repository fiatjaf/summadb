package db_test

import (
	"testing"

	db "github.com/fiatjaf/summadb/database"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCRUD(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CRUD Suite")
}

var _ = BeforeSuite(func() {
	Expect(db.Erase()).To(Succeed())
})

var _ = Describe("CRUD", func() {
	Describe("basic operations", func() {
		Context("save operations", func() {
			It("should save a tree", func() {
				Expect(db.SaveTreeAt("/fruits/banana", map[string]interface{}{
					"colour":   "yellow",
					"hardness": "low",
					"_val":     "a fruit.",
				})).To(Succeed())
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

			It("should modify a subvalue of the tree", func() {
				Expect(db.SaveValueAt("/fruits/banana/colour", []byte(`"black-and-yellow"`))).To(Succeed())
				Expect(db.GetValueAt("/fruits/banana/colour")).To(BeEquivalentTo(`"black-and-yellow"`))
				Expect(db.GetValueAt("/fruits/banana/hardness")).To(BeEquivalentTo(`"low"`))
			})

			It("should add a value deeply nested in a tree that doesn't exists", func() {
				Expect(db.SaveValueAt("/fruits/mellon/season", []byte(`"spring"`))).To(Succeed())
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

			It("should add a tree deeply nested like the previous", func() {
				Expect(db.SaveTreeAt("/fruits/orange", map[string]interface{}{
					"colour":   "orange",
					"hardness": "medium",
					"_val":     "name == colour",
				})).To(Succeed())
				Expect(db.GetValueAt("/fruits/orange/colour")).To(BeEquivalentTo(`"orange"`))
				Expect(db.GetValueAt("/fruits/orange")).To(BeEquivalentTo(`"name == colour"`))
				Expect(db.GetTreeAt("/fruits/orange")).To(Equal(map[string]interface{}{
					"_val":     "name == colour",
					"colour":   value("orange"),
					"hardness": value("medium"),
				}))
			})
		})

		Context("delete operations", func() {
			It("should delete a key", func() {
				Expect(db.DeleteAt("/fruits/banana/colour")).To(Succeed())
				Expect(db.GetValueAt("/fruits/orange/colour")).To(BeEquivalentTo(`"orange"`))
				_, err := db.GetValueAt("/fruits/banana/colour")
				Expect(db.GetValueAt("/fruits/banana/colour/_deleted")).To(BeEquivalentTo(""))
				Expect(err).To(HaveOccurred())
			})

			It("should delete a value when setting it to null with a tree", func() {
				Expect(db.SaveTreeAt("/fruits/mellon", map[string]interface{}{
					"colour": "orange",
					"season": nil,
				})).To(Succeed())
				Expect(db.GetValueAt("/fruits/mellon/colour")).To(BeEquivalentTo(`"orange"`))
				_, err := db.GetValueAt("/fruits/mellon/season")
				Expect(err).To(HaveOccurred())

				Expect(db.SaveTreeAt("/fruits", map[string]interface{}{
					"mellon": nil,
				})).To(Succeed())
				_, err = db.GetValueAt("/fruits/mellon/colour")
				Expect(err).To(HaveOccurred())
				_, err = db.GetValueAt("/fruits/mellon")
				Expect(err).To(HaveOccurred())
			})

			It("should delete a tree", func() {
				Expect(db.DeleteAt("/fruits/banana")).To(Succeed())
				Expect(db.GetValueAt("/fruits/orange/colour")).To(BeEquivalentTo(`"orange"`))
				_, err := db.GetValueAt("/fruits/banana/hardness")
				Expect(err).To(HaveOccurred())

				Expect(db.DeleteAt("/fruits")).To(Succeed())
				_, err = db.GetValueAt("/fruits")
				Expect(err).To(HaveOccurred())
				Expect(db.GetValueAt("/fruits/orange/_deleted")).To(BeEquivalentTo(""))
				_, err = db.GetValueAt("/fruits/orange/colour")
				Expect(err).To(HaveOccurred())
				_, err = db.GetValueAt("/fruits/banana/hardness")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("_rev", func() {
		Context("_rev generation and bumping for single operations", func() {
			It("should generate _rev for a single key", func() {
				db.SaveValueAt("/name", []byte(`"database of vehicles"`))
				Expect(db.GetValueAt("/name/_rev")).To(HavePrefix("1-"))
			})

			It("should generate _rev for parent keys", func() {
				db.SaveValueAt("/vehicles/car/land", []byte("true"))
				db.SaveValueAt("/vehicles/carriage/land", []byte("true"))
				db.SaveValueAt("/vehicles/carriage/air", []byte("false"))
				Expect(db.GetValueAt("/vehicles/car/land/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/car/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/_rev")).To(HavePrefix("3-"))
				Expect(db.GetValueAt("/vehicles/carriage/land/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/carriage/_rev")).To(HavePrefix("2-"))
			})

			It("should bump _rev for single keys", func() {
				db.SaveValueAt("/name", []byte(`"just a database of vehicles"`))
				Expect(db.GetValueAt("/name/_rev")).To(HavePrefix("2-"))
			})

			It("should bump _rev for parent keys", func() {
				db.SaveValueAt("/vehicles/car/water", []byte("false"))
				Expect(db.GetValueAt("/vehicles/car/land/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/car/water/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/car/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/_rev")).To(HavePrefix("4-"))
				Expect(db.GetValueAt("/vehicles/carriage/land/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/carriage/_rev")).To(HavePrefix("2-"))

				db.SaveValueAt("/vehicles/boat/water", []byte("true"))
				Expect(db.GetValueAt("/vehicles/car/land/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/car/water/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/boat/water/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/car/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/boat/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/_rev")).To(HavePrefix("5-"))
			})

			It("on delete, should bump _rev for parents and sons", func() {
				db.DeleteAt("/vehicles/car")
				Expect(db.GetValueAt("/vehicles/car/land/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/car/water/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/car/_rev")).To(HavePrefix("3-"))
				Expect(db.GetValueAt("/vehicles/boat/water/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/boat/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/_rev")).To(HavePrefix("6-"))
				Expect(db.GetValueAt("/vehicles/carriage/land/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/carriage/_rev")).To(HavePrefix("2-"))
			})
		})

		Context("_rev generation and bumping for tree operations", func() {
			It("should bump rev of all parents of affected keys", func() {
				db.SaveTreeAt("/vehicles/boat", map[string]interface{}{
					"water": true,
					"land":  false,
					"air":   false,
				})
				Expect(db.GetValueAt("/vehicles/car/land/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/car/water/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/car/_rev")).To(HavePrefix("3-"))
				Expect(db.GetValueAt("/vehicles/boat/water/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/boat/land/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/boat/air/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/boat/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/_rev")).To(HavePrefix("7-"))
			})

			It("doing it again to make sure", func() {
				db.SaveTreeAt("/vehicles", map[string]interface{}{
					"car": map[string]interface{}{
						"water": true,
					},
					"boat": map[string]interface{}{
						"air": true,
					},
				})
				Expect(db.GetValueAt("/vehicles/car/land/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/car/water/_rev")).To(HavePrefix("3-"))
				Expect(db.GetValueAt("/vehicles/car/_rev")).To(HavePrefix("4-"))
				Expect(db.GetValueAt("/vehicles/boat/water/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/boat/land/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/boat/air/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/boat/_rev")).To(HavePrefix("3-"))
				Expect(db.GetValueAt("/vehicles/_rev")).To(HavePrefix("8-"))
				Expect(db.GetValueAt("/vehicles/carriage/land/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/carriage/_rev")).To(HavePrefix("2-"))
			})

			It("should bump the revs correctly when a tree operation involves deleting", func() {
				db.SaveTreeAt("/vehicles", map[string]interface{}{
					"carriage": map[string]interface{}{
						"space": false,
						"land":  nil,
					},
					"boat": nil,
				})
				Expect(db.GetValueAt("/vehicles/car/_rev")).To(HavePrefix("4-"))
				Expect(db.GetValueAt("/vehicles/boat/water/_rev")).To(HavePrefix("3-"))
				Expect(db.GetValueAt("/vehicles/boat/land/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/boat/air/_rev")).To(HavePrefix("3-"))
				Expect(db.GetValueAt("/vehicles/boat/_rev")).To(HavePrefix("4-"))
				Expect(db.GetValueAt("/vehicles/_rev")).To(HavePrefix("9-"))
				Expect(db.GetValueAt("/vehicles/carriage/land/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/carriage/_rev")).To(HavePrefix("3-"))
			})

			It("should bump revs of intermediate paths when modifying a deep field", func() {
				db.SaveValueAt("/vehicles/train/land/rail", []byte("true"))
				Expect(db.GetValueAt("/vehicles/_rev")).To(HavePrefix("10-"))
				Expect(db.GetValueAt("/vehicles/train/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/train/land/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/train/land/rail/_rev")).To(HavePrefix("1-"))

				db.DeleteAt("/vehicles/train")
				Expect(db.GetValueAt("/vehicles/_rev")).To(HavePrefix("11-"))
				Expect(db.GetValueAt("/vehicles/train/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/train/land/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/train/land/rail/_rev")).To(HavePrefix("2-"))

				db.SaveTreeAt("", map[string]interface{}{
					"vehicles": map[string]interface{}{
						"skate": map[string]interface{}{
							"air": map[string]interface{}{
								"carried": map[string]interface{}{
									"_val": true,
								},
							},
						},
					},
				})
				Expect(db.GetValueAt("/vehicles/_rev")).To(HavePrefix("12-"))
				Expect(db.GetValueAt("/vehicles/skate/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/skate/air/_rev")).To(HavePrefix("1-"))
				Expect(db.GetValueAt("/vehicles/skate/air/carried/_rev")).To(HavePrefix("1-"))

				db.SaveTreeAt("", map[string]interface{}{
					"vehicles": map[string]interface{}{
						"skate": map[string]interface{}{
							"air": map[string]interface{}{
								"carried": nil,
							},
						},
					},
				})
				Expect(db.GetValueAt("/vehicles/_rev")).To(HavePrefix("13-"))
				Expect(db.GetValueAt("/vehicles/skate/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/skate/air/_rev")).To(HavePrefix("2-"))
				Expect(db.GetValueAt("/vehicles/skate/air/carried/_rev")).To(HavePrefix("2-"))
			})
		})
	})
})

func value(v string) map[string]interface{} {
	return map[string]interface{}{"_val": v}
}
