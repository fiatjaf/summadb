package database

import (
	"time"

	"github.com/summadb/summadb/types"
	. "github.com/summadb/summadb/utils"
	. "gopkg.in/check.v1"
)

func (s *DatabaseSuite) TestRows(c *C) {
	db := Open("/tmp/summadb-test-rows")
	db.Erase()
	db = Open("/tmp/summadb-test-rows")

	// insert a tree
	err = db.Set(types.Path{"eatables"}, types.Tree{
		Leaf: types.StringLeaf("can be eaten"),
		Branches: types.Branches{
			"banana": &types.Tree{
				Branches: types.Branches{
					"color": &types.Tree{Leaf: types.StringLeaf("yellow")},
					"size":  &types.Tree{Leaf: types.NumberLeaf(15)},
				},
			},
			"chia": &types.Tree{
				Branches: types.Branches{
					"color": &types.Tree{Leaf: types.StringLeaf("blue")},
					"size":  &types.Tree{Leaf: types.NumberLeaf(0.7)},
				},
			},
			"alfajor": &types.Tree{
				Branches: types.Branches{
					"color": &types.Tree{Leaf: types.StringLeaf("brown")},
					"size":  &types.Tree{Leaf: types.NumberLeaf(7)},
				},
			},
			"alfarroba": &types.Tree{
				Branches: types.Branches{
					"color": &types.Tree{Leaf: types.StringLeaf("brown")},
					"size":  &types.Tree{Leaf: types.NumberLeaf(0.001)},
				},
			},
			"bancha": &types.Tree{
				Branches: types.Branches{
					"color": &types.Tree{Leaf: types.StringLeaf("green")},
					"size":  &types.Tree{Leaf: types.NumberLeaf(1)},
				},
			},
		},
	})
	c.Assert(err, IsNil)

	rows, err := db.Rows(types.Path{"eatables"}, RowsParams{KeyStart: "b", KeyEnd: "ba~"})
	c.Assert(err, IsNil)

	c.Assert(rows, HasLen, 2)
	c.Assert(rows[0].Branches["color"].Leaf, DeepEquals, types.StringLeaf("yellow"))
	c.Assert(rows[1].Branches["color"].Leaf, DeepEquals, types.StringLeaf("green"))

	// make an autocomplete index
	rev, _ := db.Rev(types.Path{"eatables"})
	err = db.Merge(types.Path{"eatables"}, types.Tree{
		Rev: rev,
		Map: `
for i=0, string.len(doc._key) do
  local part = string.sub(doc._key, 0, i)
  emit('search', part .. ":" .. doc._key, doc._key)
end
        `,
	})
	c.Assert(err, IsNil)
	time.Sleep(time.Millisecond * 200)

	rows, err = db.Rows(types.Path{"eatables", "@map", "search"}, RowsParams{
		KeyStart:   "alf:",
		KeyEnd:     "alf:~",
		Descending: true,
	})
	c.Assert(err, IsNil)
	c.Assert(rows, HasLen, 2)
	c.Assert(rows[0].Leaf, DeepEquals, types.StringLeaf("alfarroba"))
	c.Assert(rows[1].Leaf, DeepEquals, types.StringLeaf("alfajor"))

	// limit
	rows, err = db.Rows(types.Path{"eatables"}, RowsParams{Limit: 2})
	c.Assert(err, IsNil)
	c.Assert(rows, HasLen, 2)
	c.Assert(rows[0].Rev, StartsWith, "1-") // do fetched _rev

	rows, err = db.Rows(types.Path{"eatables"}, RowsParams{Limit: 64})
	c.Assert(err, IsNil)
	c.Assert(rows, HasLen, 5) // don't fetched more than available, nor @map

	// don't fetched deleted
	err = db.Delete(types.Path{"eatables", rows[0].Key}, rows[0].Rev)
	c.Assert(err, IsNil)

	rows, err = db.Rows(types.Path{"eatables"}, RowsParams{})
	c.Assert(err, IsNil)
	c.Assert(rows, HasLen, 4) // don't fetched deleted rows

	// don't fetch @map results
	rows, err = db.Rows(types.Path{}, RowsParams{})
	c.Assert(err, IsNil)
	c.Assert(rows, HasLen, 1)
	c.Assert(rows[0].Key, Equals, "eatables")
	c.Assert(rows[0].Leaf, DeepEquals, types.StringLeaf("can be eaten"))
	_, hasmapbranch := rows[0].Branches["@map"]
	c.Assert(hasmapbranch, Equals, false)
}
