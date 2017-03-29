package test

import (
	"testing"

	"github.com/fiatjaf/summadb/database"
	"github.com/fiatjaf/summadb/types"
	. "github.com/fiatjaf/summadb/utils"
	. "gopkg.in/check.v1"
)

var err error

func TestDatabase(t *testing.T) {
	TestingT(t)
}

type DatabaseSuite struct{}

var _ = Suite(&DatabaseSuite{})

func (s *DatabaseSuite) TestBasic(c *C) {
	db := database.Open("/tmp/summadb-test-database")
	// defer db.Erase()

	fruitsbanana := types.Tree{
		Leaf: types.StringLeaf("can be eaten"),
		Branches: types.Branches{
			"banana": &types.Tree{
				Branches: types.Branches{
					"color": &types.Tree{Leaf: types.StringLeaf("blue")},
					"size":  &types.Tree{Leaf: types.NumberLeaf(15)},
				},
			},
		},
	}

	err = db.Set(
		types.Path{"fruits"},
		fruitsbanana,
	)
	c.Assert(err, IsNil)

	treeread, err := db.Read(types.Path{"fruits"})
	c.Assert(err, IsNil)
	c.Assert(treeread, DeeplyEquals, fruitsbanana)
}
