package db_test

import (
	"testing"

	db "github.com/fiatjaf/summadb/database"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestSeqs(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("seqs getting bumped", func() {

		g.Before(func() {
			db.Erase()
			db.Start()
		})

		g.After(func() {
			db.End()
		})

		/*
		   we currently don't care a lot to the exact seqs here,
		   we only care if the values are being increased and there
		   are no conflicts.

		   later maybe we'll have to optimize all this (and conseqüently
		   change all these tests) so we don't get new the seq
		   bumped for operations that do not result in an actual _rev
		   being written (in DELETE and UNDELETE operations there are
		   temporary _revs being created and then replaced by others
		   with the same number prefix).
		*/

		g.It("should increase seqs when adding a new tree", func() {
			db.SaveTreeAt("", map[string]interface{}{
				"x": "oiwaeuriasburis",
			})
			Expect(db.GlobalUpdateSeq()).To(BeEquivalentTo(2))
			Expect(db.LastSeqAt("")).To(BeNumerically(">", uint64(0)))
			Expect(db.LastSeqAt("/x")).To(BeNumerically("==", uint64(0)))
		})

		g.It("should increase seqs when adding a new value", func() {
			db.SaveValueAt("/z", []byte("ihfiuewrhewoiruh"))
			Expect(db.GlobalUpdateSeq()).To(BeEquivalentTo(4))
			Expect(db.LastSeqAt("")).To(BeNumerically(">", uint64(2)))
			Expect(db.LastSeqAt("/x")).To(BeNumerically("==", uint64(0)))
		})

		g.It("should increase seqs when deleting a value", func() {
			db.DeleteAt("/x")
			Expect(db.GlobalUpdateSeq()).To(BeEquivalentTo(7))
			Expect(db.LastSeqAt("")).To(BeNumerically(">", uint64(5)))
			Expect(db.LastSeqAt("/x")).To(BeNumerically("==", uint64(0)))
			Expect(db.LastSeqAt("/z")).To(BeNumerically("==", uint64(0)))
		})

		g.It("should increase seqs when undeleting a value", func() {
			db.SaveValueAt("/x/xchild", []byte("skjfbslkfbskdf"))
			Expect(db.GlobalUpdateSeq()).To(BeEquivalentTo(10))
			Expect(db.LastSeqAt("")).To(BeNumerically(">", uint64(7)))
			Expect(db.LastSeqAt("/x")).To(BeNumerically(">", uint64(7)))
		})

		g.It("should increase seqs when making bizarre things", func() {
			db.ReplaceTreeAt("/x", map[string]interface{}{
				"xchild": nil,
				"other":  "saldkasndlksad",
				"_val":   "askjdasnkdjasd",
			})
			db.ReplaceTreeAt("/x/xchild", map[string]interface{}{
				"ham":  "sadljkasndlksad",
				"_val": "askjdasnkdjasd",
			})
			Expect(db.GlobalUpdateSeq()).To(BeEquivalentTo(22))
		})
	})
}
