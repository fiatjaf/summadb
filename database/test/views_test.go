package test

import (
	"time"

	"github.com/fiatjaf/summadb/database"
	"github.com/fiatjaf/summadb/types"
	. "gopkg.in/check.v1"
)

func (s *DatabaseSuite) TestViews(c *C) {
	db := database.Open("/tmp/summadb-test-views")
	db.Erase()
	db = database.Open("/tmp/summadb-test-views")

	// insert a tree with a map function
	err = db.Set(types.Path{"food"}, types.Tree{
		Map: `
local food = doc
emit(
  {food.kind._val, food.size._val},
  {name=food.name._val, kind=food.kind._val}
)
        `,
		Branches: types.Branches{
			"1": &types.Tree{
				Branches: types.Branches{
					"kind": &types.Tree{Leaf: types.StringLeaf("fruit")},
					"name": &types.Tree{Leaf: types.StringLeaf("apple")},
					"size": &types.Tree{Leaf: types.NumberLeaf(10)},
				},
			},
			"2": &types.Tree{
				Branches: types.Branches{
					"kind": &types.Tree{Leaf: types.StringLeaf("tuber")},
					"name": &types.Tree{Leaf: types.StringLeaf("potato")},
					"size": &types.Tree{Leaf: types.NumberLeaf(12)},
				},
			},
			"3": &types.Tree{
				Branches: types.Branches{
					"kind": &types.Tree{Leaf: types.StringLeaf("tuber")},
					"name": &types.Tree{Leaf: types.StringLeaf("carrot")},
					"size": &types.Tree{Leaf: types.NumberLeaf(17)},
				},
			},
		},
	})
	c.Assert(err, IsNil)
	time.Sleep(time.Second * 1)

	treeread, err := db.Read(types.Path{"food", "@map"})
	c.Assert(err, IsNil)
	c.Assert(len(treeread.Branches), Equals, 3)
}
