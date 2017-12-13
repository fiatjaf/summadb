package database

import (
	"time"

	"github.com/summadb/summadb/types"
	. "github.com/summadb/summadb/utils"
	. "gopkg.in/check.v1"
)

func (s *DatabaseSuite) TestMapFunctions(c *C) {
	db := Open("/tmp/summadb-test-mapf")
	defer db.Erase()

	// insert a tree with a map function
	mapf := `
local food = doc
emit('by-kind', food.kind._val, _key, food.name._val)
    `
	err = db.Set(types.Path{"food"}, types.Tree{
		Map: mapf,
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

	treeread, err := db.Read(types.Path{"food", "!map", "by-kind"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Branches, HasLen, 2 /* 'tuber' and 'fruit' */)
	c.Assert(treeread.Branches["fruit"].Branches, HasLen, 1)
	c.Assert(treeread.Branches["tuber"], DeeplyEquals, &types.Tree{
		Branches: types.Branches{
			"3": &types.Tree{Leaf: types.StringLeaf("carrot")},
			"2": &types.Tree{Leaf: types.StringLeaf("potato")},
		},
	})

	// can't read !map directly
	_, err = db.Read(types.Path{"food", "!map"})
	c.Assert(err, Not(IsNil))

	// !map results are not in the tree
	treeread, err = db.Read(types.Path{"food"})
	c.Assert(err, IsNil)
	_, is := treeread.Branches["!map"]
	c.Assert(is, Equals, false)
	c.Assert(treeread.Map, Equals, mapf) // correct mapf value is returned on Map

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

	treeread, err = db.Read(types.Path{"food", "!map", "by-kind"})
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

	treeread, err = db.Read(types.Path{"food", "!map", "by-kind"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Branches, HasLen, 0)

	treeread, err = db.Read(types.Path{"food", "!map", "by-size"})
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

	treeread, err = db.Read(types.Path{"food", "!map", "by-size"})
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

	treeread, err = db.Read(types.Path{"food", "!map", "by-kind"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Branches, HasLen, 0)

	treeread, err = db.Read(types.Path{"food", "!map", "by-size"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Branches, HasLen, 0)

	// insert multiple map functions at different levels
	rev, _ = db.Rev(types.Path{})

	// check rev first
	c.Assert(rev, StartsWith, "5-")

	err = db.Merge(types.Path{}, types.Tree{
		Rev: rev,
		Map: `emit('categories', _key, true)`,
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

	treeread, err = db.Read(types.Path{"!map", "categories", "food"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Leaf, DeepEquals, types.BoolLeaf(true))

	treeread, err = db.Read(types.Path{"food", "!map", "categories", "tuber"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Leaf, DeepEquals, types.BoolLeaf(true))
}

func (s *DatabaseSuite) TestReduceFunctions(c *C) {
	db := Open("/tmp/summadb-test-reducef")
	db.Erase()
	db = Open("/tmp/summadb-test-reducef")

	mapf := `
for word in string.gmatch(doc._val, "%S+") do
  if prev ~= "" then
    emit("next-word", prev, word)
  end
  prev = word
end
    `
	reducef := `
local thisword = path[2] -- path is {'next-word', <word>}
local all = acc[thisword] or {}

-- all is a table of {<indexify({<count>, <nextword>})>={count=<count>, next=<nextword>}}

-- value is the value emitted by the mapf
local emittednext = value._val

for idx, data in pairs(all) do
  if idx == indexify({data.count, emittednext}) then
    -- this is the record we're looking for, it matches word and nextword.

    if directive == "add" then
      count = count + 1
    elseif directive == "remove" then
      count = count - 1
    end

    -- remove this item from the index
    acc[thisword][idx] = nil

    -- add a new
    if count > 0 then
      acc[thisword][indexify(count, emittednext)] = {count=count, next=nextword}
    end
  end
end
    `

	t := types.Tree{
		Map:    mapf,
		Reduce: reducef,
		Branches: types.Branches{
			"82eu7": &types.Tree{Leaf: types.StringLeaf("amanhã você vai lá ontem?")},
			"52rt9": &types.Tree{Leaf: types.StringLeaf("amanhã vai ser outro dia")},
			"4w2vx": &types.Tree{Leaf: types.StringLeaf("quando for amanhã será")},
			"s9wo2": &types.Tree{Leaf: types.StringLeaf("amanhã vai chover")},
		},
	}

	err := db.Set(types.Path{}, t)
	c.Assert(err, IsNil)
	time.Sleep(time.Millisecond * 200)

	records, err := db.Query(types.Path{"!reduce", "amanhã"}, QueryParams{
		Limit:      3,
		Descending: true,
	})
	c.Assert(err, IsNil)

	c.Assert(records, HasLen, 3)
	c.Assert(records[0].Branches["next"].Leaf, DeepEquals, types.StringLeaf("vai"))
	c.Assert(records[0].Branches["count"].Leaf, DeepEquals, types.NumberLeaf(2))
	c.Assert(records[1].Branches["next"].Leaf, DeepEquals, types.StringLeaf("for"))
	c.Assert(records[1].Branches["count"].Leaf, DeepEquals, types.NumberLeaf(1))
	c.Assert(records[2].Branches["next"].Leaf, DeepEquals, types.StringLeaf("você"))
	c.Assert(records[2].Branches["count"].Leaf, DeepEquals, types.NumberLeaf(1))
}
