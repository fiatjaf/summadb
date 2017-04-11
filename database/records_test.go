package database

import (
	"time"

	"github.com/summadb/summadb/types"
	. "github.com/summadb/summadb/utils"
	. "gopkg.in/check.v1"
)

func (s *DatabaseSuite) TestRecords(c *C) {
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

	rows, err := db.Records(types.Path{"eatables"}, RecordsParams{KeyStart: "b", KeyEnd: "ba~"})
	c.Assert(err, IsNil)

	c.Assert(rows, HasLen, 2)
	c.Assert(rows[0].Branches["color"].Leaf, DeepEquals, types.StringLeaf("yellow"))
	c.Assert(rows[1].Branches["color"].Leaf, DeepEquals, types.StringLeaf("green"))

	// make an autocomplete index
	mapf := `
for i=0, string.len(_key) do
  local part = string.sub(_key, 0, i)
  emit('search', part .. ":" .. _key, _key)
end
    `
	rev, _ := db.Rev(types.Path{"eatables"})
	err = db.Merge(types.Path{"eatables"}, types.Tree{
		Rev: rev,
		Map: mapf,
	})
	c.Assert(err, IsNil)
	time.Sleep(time.Millisecond * 200)

	rows, err = db.Records(types.Path{"eatables", "!map", "search"}, RecordsParams{
		KeyStart:   "alf:",
		KeyEnd:     "alf:~",
		Descending: true,
	})
	c.Assert(err, IsNil)
	c.Assert(rows, HasLen, 2)
	c.Assert(rows[0].Leaf, DeepEquals, types.StringLeaf("alfarroba"))
	c.Assert(rows[1].Leaf, DeepEquals, types.StringLeaf("alfajor"))

	// limit
	rows, err = db.Records(types.Path{"eatables"}, RecordsParams{Limit: 2})
	c.Assert(err, IsNil)
	c.Assert(rows, HasLen, 2)
	c.Assert(rows[0].Rev, StartsWith, "1-") // do fetched _rev

	rows, err = db.Records(types.Path{"eatables"}, RecordsParams{Limit: 64})
	c.Assert(err, IsNil)
	c.Assert(rows, HasLen, 5) // don't fetched more than available, nor !map

	// don't fetched deleted
	err = db.Delete(types.Path{"eatables", rows[0].Key}, rows[0].Rev)
	c.Assert(err, IsNil)

	rows, err = db.Records(types.Path{"eatables"}, RecordsParams{})
	c.Assert(err, IsNil)
	c.Assert(rows, HasLen, 4) // don't fetched deleted rows

	// don't fetch !map results
	rows, err = db.Records(types.Path{}, RecordsParams{})
	c.Assert(err, IsNil)
	c.Assert(rows, HasLen, 1)
	c.Assert(rows[0].Key, Equals, "eatables")
	c.Assert(rows[0].Leaf, DeepEquals, types.StringLeaf("can be eaten"))
	_, hasmapbranch := rows[0].Branches["!map"]
	c.Assert(hasmapbranch, Equals, false)
	c.Assert(rows[0].Map, Equals, mapf) // .Map should be the code of the mapf
}
