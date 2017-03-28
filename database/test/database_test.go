package test

import (
	"testing"

	"github.com/fiatjaf/summadb/database"
	"github.com/fiatjaf/summadb/types"
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
	defer db.Erase()

	err = db.Set(
		types.Path{"fruits"},
		types.Tree{
			Leaf: types.StringLeaf("can be eaten"),
			Branches: map[string]types.Tree{
				"banana": types.Tree{
					Branches: map[string]types.Tree{
						"color": types.Tree{Leaf: types.StringLeaf("blue")},
						"size":  types.Tree{Leaf: types.NumberLeaf(15)},
					},
				},
			},
		},
	)
	c.Assert(err, IsNil)
}
