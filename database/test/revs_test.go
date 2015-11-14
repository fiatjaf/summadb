package db_test

import (
	db "github.com/fiatjaf/summadb/database"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("_rev", func() {
	Context("_rev generation and bumping for single operations", func() {
		It("should erase the db", func() {
			Expect(db.Erase()).To(Succeed())
		})

		It("should generate _rev for a single key", func() {
			savedrev, _ := db.SaveValueAt("/name", []byte(`"database of vehicles"`))
			gottenrev, _ := db.GetValueAt("/name/_rev")
			Expect(savedrev).To(BeEquivalentTo(gottenrev))
			Expect(gottenrev).To(HavePrefix("1-"))
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

			db.SaveTreeAt("/", map[string]interface{}{
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

			db.SaveTreeAt("/", map[string]interface{}{
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

	Context("GetSpecialKeysAt", func() {
		It("should return rev", func() {
			sk, err := db.GetSpecialKeysAt("/vehicles/skate")
			Expect(err).ToNot(HaveOccurred())
			Expect(sk.Rev).To(HavePrefix("2-"))
		})
	})
})
