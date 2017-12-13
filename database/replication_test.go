package database

import (
	"github.com/summadb/summadb/types"
	. "github.com/summadb/summadb/utils"
	. "gopkg.in/check.v1"
)

func (s *DatabaseSuite) TestAllRevs(c *C) {
	db := Open("/tmp/summadb-test-allrevs")
	defer db.Erase()

	rpl := Replicator{db, types.Path{"subdb"}}

	pathrevs, err := rpl.AllRevs()
	c.Assert(err, IsNil)
	c.Assert(pathrevs, HasLen, 0)

	db.Set(types.Path{}, types.Tree{
		Leaf: types.NumberLeaf(12),
		Branches: types.Branches{
			"subdb": &types.Tree{
				Leaf: types.StringLeaf("uwuwu"),
				Branches: types.Branches{
					"doc": &types.Tree{
						Leaf: types.StringLeaf("a doc."),
						Branches: types.Branches{
							"size": &types.Tree{Leaf: types.StringLeaf("a value.")},
						},
					},
					"other doc": &types.Tree{
						Leaf: types.StringLeaf("another doc."),
						Branches: types.Branches{
							"size": &types.Tree{Leaf: types.StringLeaf("another value.")},
						},
					},
				},
			},
			"otehr subdb": &types.Tree{
				Leaf: types.StringLeaf("ytytyt"),
				Branches: types.Branches{
					"doc": &types.Tree{
						Leaf: types.StringLeaf("a doc."),
						Branches: types.Branches{
							"size": &types.Tree{Leaf: types.StringLeaf("a value.")},
						},
					},
					"other doc": &types.Tree{
						Leaf: types.StringLeaf("another doc."),
						Branches: types.Branches{
							"size": &types.Tree{Leaf: types.StringLeaf("another value.")},
						},
					},
				},
			},
		},
	})

	pathrevs, err = rpl.AllRevs()
	c.Assert(err, IsNil)
	c.Assert(pathrevs, HasLen, 4)
	c.Assert(pathrevs[0].Path, Equals, "doc")
	c.Assert(pathrevs[1].Path, Equals, "doc/size")
	c.Assert(pathrevs[2].Path, Equals, "other doc")
	c.Assert(pathrevs[3].Path, Equals, "other doc/size")
	c.Assert(pathrevs[0].Rev, StartsWith, "1-")
	c.Assert(pathrevs[1].Rev, StartsWith, "1-")
	c.Assert(pathrevs[2].Rev, StartsWith, "1-")
	c.Assert(pathrevs[3].Rev, StartsWith, "1-")

	rev, _ := db.Rev(types.Path{})
	err = db.Merge(types.Path{}, types.TreeFromJSON(`{
        "_rev": "`+rev+`",
        "subdb": {
            "doc": {"size": 12},
            "third doc": {"_val": "a third doc.", "size": 23}
        }
    }`))
	c.Assert(err, IsNil)

	pathrevs, err = rpl.AllRevs()
	c.Assert(err, IsNil)
	c.Assert(pathrevs, HasLen, 6)
	c.Assert(pathrevs[0].Path, Equals, "doc")
	c.Assert(pathrevs[1].Path, Equals, "doc/size")
	c.Assert(pathrevs[2].Path, Equals, "other doc")
	c.Assert(pathrevs[3].Path, Equals, "other doc/size")
	c.Assert(pathrevs[4].Path, Equals, "third doc")
	c.Assert(pathrevs[5].Path, Equals, "third doc/size")
	c.Assert(pathrevs[0].Rev, StartsWith, "2-")
	c.Assert(pathrevs[1].Rev, StartsWith, "2-")
	c.Assert(pathrevs[2].Rev, StartsWith, "1-")
	c.Assert(pathrevs[3].Rev, StartsWith, "1-")
	c.Assert(pathrevs[2].Rev, StartsWith, "1-")
	c.Assert(pathrevs[3].Rev, StartsWith, "1-")
}

func (s *DatabaseSuite) TestRevsDiff(c *C) {
	db := Open("/tmp/summadb-test-revsdiff")
	defer db.Erase()

	rpl1 := Replicator{db, types.Path{"sub1"}}
	rpl2 := Replicator{db, types.Path{"sub2"}}

	err := db.Set(types.Path{"sub1"}, types.TreeFromJSON(`{
        "1": {"ok": true},
        "2": {"ok": true},
    }`))
	c.Assert(err, IsNil)

	err = db.Set(types.Path{"sub2"}, types.TreeFromJSON(`{
        "4": {"ok": false},
        "3": {"ok": true},
    }`))
	c.Assert(err, IsNil)

	allrevs, err := rpl1.AllRevs()
	c.Assert(err, IsNil)

	diff, err := rpl2.RevsDiff(allrevs)
	c.Assert(err, IsNil)
	c.Assert(diff, DeepEquals, []string{
		"1", "1/ok",
		"2", "2/ok",
		"3", "3/ok",
		"4", "4/ok",
	})
}
