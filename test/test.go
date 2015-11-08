package test

import (
	. "github.com/franela/goblin"
	"testing"

	"github.com/fiatjaf/summadb/database"
)

func Test(t *testing.T) {
	g := Goblin(t)
	g.Describe("CRUD", func() {
		g.It("should save some values, get them and delete them", func() {
			database.SaveTreeAt("/fruits/banana", map[string]interface{}{
				"colour":   "yellow",
				"hardness": "low",
			})
			g.Assert(database.GetValueAt("/fruits/banana/colour")).Equals("yellow")
		})
	})
}
