package database

import (
	"time"

	"github.com/fiatjaf/summadb/types"
	. "github.com/fiatjaf/summadb/utils"
	. "gopkg.in/check.v1"
)

func (s *DatabaseSuite) TestViews(c *C) {
	db := Open("/tmp/summadb-test-views")
	db.Erase()
	db = Open("/tmp/summadb-test-views")

	// insert a tree with a map function
	err = db.Set(types.Path{"food"}, types.Tree{
		Map: `
local food = doc
emit('by-kind', food.kind._val, food._id, food.name._val)
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
	time.Sleep(time.Millisecond * 200)

	treeread, err := db.Read(types.Path{"food", "@map", "by-kind"})
	c.Assert(err, IsNil)
	c.Assert(len(treeread.Branches), Equals, 2 /* 'tuber' and 'fruit' */)
	c.Assert(len(treeread.Branches["fruit"].Branches), Equals, 1)
	c.Assert(treeread.Branches["tuber"], DeeplyEquals, &types.Tree{
		Branches: types.Branches{
			"3": &types.Tree{Leaf: types.StringLeaf("carrot")},
			"2": &types.Tree{Leaf: types.StringLeaf("potato")},
		},
	})

	// modify the tree
	err = db.Set(types.Path{"food", "4"}, types.Tree{
		Branches: types.Branches{
			"kind": &types.Tree{Leaf: types.StringLeaf("tuber")},
			"name": &types.Tree{Leaf: types.StringLeaf("yam")},
			"size": &types.Tree{Leaf: types.NumberLeaf(9)},
		},
	})
	c.Assert(err, IsNil)
	time.Sleep(time.Millisecond * 200)

	treeread, err = db.Read(types.Path{"food", "@map", "by-kind"})
	c.Assert(err, IsNil)
	c.Assert(len(treeread.Branches), Equals, 2 /* 'tuber' and 'fruit' */)
	c.Assert(treeread.Branches["tuber"], DeeplyEquals, &types.Tree{
		Branches: types.Branches{
			"3": &types.Tree{Leaf: types.StringLeaf("carrot")},
			"2": &types.Tree{Leaf: types.StringLeaf("potato")},
			"4": &types.Tree{Leaf: types.StringLeaf("yam")},
		},
	})
}
