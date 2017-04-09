package database

import (
	"time"

	"github.com/summadb/summadb/types"
	. "github.com/summadb/summadb/utils"
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
emit('by-kind', food.kind._val, doc._key, food.name._val)
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
	c.Assert(treeread.Branches, HasLen, 2 /* 'tuber' and 'fruit' */)
	c.Assert(treeread.Branches["fruit"].Branches, HasLen, 1)
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
	c.Assert(treeread.Branches, HasLen, 2 /* 'tuber' and 'fruit' */)
	c.Assert(treeread.Branches["tuber"], DeeplyEquals, &types.Tree{
		Branches: types.Branches{
			"3": &types.Tree{Leaf: types.StringLeaf("carrot")},
			"2": &types.Tree{Leaf: types.StringLeaf("potato")},
			"4": &types.Tree{Leaf: types.StringLeaf("yam")},
		},
	})

	// modify the map function
	rev, _ := db.Rev(types.Path{"food"})
	err = db.Merge(types.Path{"food"}, types.Tree{
		Rev: rev,
		Map: `
local food = doc
emit('by-size', food.size._val, food)
        `,
	})
	time.Sleep(time.Millisecond * 200)
	c.Assert(err, IsNil)

	treeread, err = db.Read(types.Path{"food", "@map", "by-kind"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Branches, HasLen, 0)

	treeread, err = db.Read(types.Path{"food", "@map", "by-size"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Branches, HasLen, 4 /* each food has a different size */)
	c.Assert(treeread.Branches["9"].Branches["name"].Leaf, DeepEquals, types.StringLeaf("yam"))
	c.Assert(treeread.Branches["17"].Branches["size"].Leaf, DeepEquals, types.NumberLeaf(17))
	c.Assert(treeread.Branches["12"].Branches["kind"].Leaf, DeepEquals, types.StringLeaf("tuber"))
	c.Assert(treeread.Branches["10"].Branches["name"].Leaf, DeepEquals, types.StringLeaf("apple"))

	// delete a subtree using delete()
	rev, _ = db.Rev(types.Path{"food", "1"})
	err = db.Delete(types.Path{"food", "1"}, rev)
	time.Sleep(time.Millisecond * 200)
	c.Assert(err, IsNil)

	treeread, err = db.Read(types.Path{"food", "@map", "by-size"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Branches, HasLen, 3)

	// delete the map function with merge()
	rev, _ = db.Rev(types.Path{"food"})
	err = db.Merge(types.Path{"food"}, types.Tree{
		Rev:     rev,
		Deleted: true,
	})
	time.Sleep(time.Millisecond * 200)
	c.Assert(err, IsNil)

	treeread, err = db.Read(types.Path{"food", "@map", "by-kind"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Branches, HasLen, 0)

	treeread, err = db.Read(types.Path{"food", "@map", "by-size"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Branches, HasLen, 0)

	// insert multiple map functions at different levels
	rev, _ = db.Rev(types.Path{})

	// check rev first
	c.Assert(rev, StartsWith, "5-")

	err = db.Merge(types.Path{}, types.Tree{
		Rev: rev,
		Map: `emit('categories', doc._key, true)`,
		Branches: types.Branches{
			"food": &types.Tree{
				Map: `
                if doc.kind ~= nil and doc.kind._del ~= true then
                  emit('categories', doc.kind._val,  true)
                end
                `,
			},
		},
	})
	time.Sleep(time.Millisecond * 200)
	c.Assert(err, IsNil)

	treeread, err = db.Read(types.Path{"@map", "categories", "food"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Leaf, DeepEquals, types.BoolLeaf(true))

	treeread, err = db.Read(types.Path{"food", "@map", "categories", "tuber"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Leaf, DeepEquals, types.BoolLeaf(true))
}
