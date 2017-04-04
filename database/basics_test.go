package database

import (
	"testing"

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
	db := Open("/tmp/summadb-test-basics")
	db.Erase()
	db = Open("/tmp/summadb-test-basics")

	// insert a tree
	err = db.Set(types.Path{"fruits"}, types.Tree{
		Leaf: types.StringLeaf("can be eaten"),
		Branches: types.Branches{
			"banana": &types.Tree{
				Branches: types.Branches{
					"color": &types.Tree{Leaf: types.StringLeaf("blue")},
					"size":  &types.Tree{Leaf: types.NumberLeaf(15)},
				},
			},
		},
	})
	c.Assert(err, IsNil)

	treeread, err := db.Read(types.Path{"fruits"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["banana"].Branches["color"].Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["banana"].Branches["size"].Rev, StartsWith, "1-")
	c.Assert(treeread.Leaf, DeepEquals, types.StringLeaf("can be eaten"))
	c.Assert(treeread.Branches["banana"].Branches["color"].Leaf, DeepEquals, types.StringLeaf("blue"))

	// change a property inside a tree
	err = db.Set(types.Path{"fruits", "banana", "color"}, types.Tree{
		Rev:  treeread.Branches["banana"].Branches["color"].Rev,
		Leaf: types.StringLeaf("yellow"),
	})
	c.Assert(err, IsNil)

	treeread, err = db.Read(types.Path{"fruits"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Rev, StartsWith, "2-")
	c.Assert(treeread.Branches["banana"].Branches["color"].Rev, StartsWith, "2-")
	c.Assert(treeread.Branches["banana"].Branches["size"].Rev, StartsWith, "1-")
	c.Assert(treeread.Leaf, DeepEquals, types.StringLeaf("can be eaten"))
	c.Assert(treeread.Branches["banana"].Branches["color"].Leaf, DeepEquals, types.StringLeaf("yellow"))

	// insert a sibling tree
	err = db.Set(types.Path{"fruits", "tangerine"}, types.Tree{
		Leaf: types.StringLeaf("juice can be made of"),
		Branches: types.Branches{
			"color": &types.Tree{Leaf: types.StringLeaf("orange")},
			"size":  &types.Tree{Leaf: types.NumberLeaf(11)},
		},
	})
	c.Assert(err, IsNil)

	treeread, err = db.Read(types.Path{"fruits"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Rev, StartsWith, "3-")
	c.Assert(treeread.Branches["banana"].Branches["color"].Rev, StartsWith, "2-")
	c.Assert(treeread.Branches["banana"].Branches["size"].Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["banana"].Branches["color"].Leaf, DeepEquals, types.StringLeaf("yellow"))
	c.Assert(treeread.Branches["tangerine"].Leaf, DeepEquals, types.StringLeaf("juice can be made of"))
	c.Assert(treeread.Branches["tangerine"].Branches["color"].Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["tangerine"].Branches["size"].Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["tangerine"].Branches["color"].Leaf, DeepEquals, types.StringLeaf("orange"))

	// insert a property at a previously unknown path, without parents
	err = db.Set(types.Path{"fruits", "lemon", "color"}, types.Tree{
		Leaf: types.StringLeaf("green"),
	})
	c.Assert(err, IsNil)

	treeread, err = db.Read(types.Path{})
	c.Assert(err, IsNil)
	c.Assert(treeread.Rev, StartsWith, "4-")
	c.Assert(treeread.Branches["fruits"].Rev, StartsWith, "4-")
	c.Assert(treeread.Branches["fruits"].Branches["lemon"].Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["fruits"].Branches["lemon"].Branches["color"].Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["fruits"].Branches["lemon"].Branches["color"].Leaf, DeepEquals, types.StringLeaf("green"))

	// mark some paths as deleted
	err = db.Set(types.Path{"fruits", "tangerine"}, types.Tree{
		Rev:     treeread.Branches["fruits"].Branches["tangerine"].Rev,
		Deleted: true,
		Branches: types.Branches{
			"color":   &types.Tree{Deleted: true},
			"size":    &types.Tree{Deleted: true},
			"comment": &types.Tree{Leaf: types.StringLeaf("juice can be made of")},
			"tasty":   &types.Tree{Leaf: types.BoolLeaf(true)},
		},
	})
	c.Assert(err, IsNil)

	treeread, err = db.Read(types.Path{"fruits"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Rev, StartsWith, "5-")
	c.Assert(treeread.Branches["tangerine"].Leaf, DeepEquals, types.Leaf{})
	c.Assert(treeread.Branches["tangerine"].Deleted, Equals, true)
	c.Assert(treeread.Branches["tangerine"].Rev, StartsWith, "2-")
	c.Assert(treeread.Branches["tangerine"].Branches["color"].Rev, StartsWith, "2-")
	c.Assert(treeread.Branches["tangerine"].Branches["color"].Deleted, Equals, true)
	c.Assert(treeread.Branches["tangerine"].Branches["color"].Leaf, DeepEquals, types.Leaf{})
	c.Assert(treeread.Branches["tangerine"].Branches["size"].Rev, StartsWith, "2-")
	c.Assert(treeread.Branches["tangerine"].Branches["size"].Deleted, Equals, true)
	c.Assert(treeread.Branches["tangerine"].Branches["size"].Leaf, DeepEquals, types.Leaf{})
	c.Assert(treeread.Branches["tangerine"].Branches["comment"].Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["tangerine"].Branches["tasty"].Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["tangerine"].Branches["comment"].Leaf, DeepEquals, types.StringLeaf("juice can be made of"))
	c.Assert(treeread.Branches["tangerine"].Branches["tasty"].Leaf, DeepEquals, types.BoolLeaf(true))

	// delete entire subtrees
	err = db.Delete(types.Path{"fruits", "tangerine"}, treeread.Branches["tangerine"].Rev)
	c.Assert(err, IsNil)
	treeread, err = db.Read(types.Path{"fruits"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Rev, StartsWith, "6-")
	c.Assert(treeread.Branches["tangerine"].Rev, StartsWith, "3-")
	c.Assert(treeread.Branches["tangerine"].Deleted, Equals, true)
	c.Assert(treeread.Branches["tangerine"].Leaf, DeepEquals, types.Leaf{})
	c.Assert(treeread.Branches["tangerine"].Branches["color"].Rev, StartsWith, "2-")
	c.Assert(treeread.Branches["tangerine"].Branches["color"].Deleted, Equals, true)
	c.Assert(treeread.Branches["tangerine"].Branches["size"].Rev, StartsWith, "2-")
	c.Assert(treeread.Branches["tangerine"].Branches["size"].Deleted, Equals, true)
	c.Assert(treeread.Branches["tangerine"].Branches["comment"].Rev, StartsWith, "2-")
	c.Assert(treeread.Branches["tangerine"].Branches["tasty"].Rev, StartsWith, "2-")
	c.Assert(treeread.Branches["tangerine"].Branches["comment"].Deleted, Equals, true)
	c.Assert(treeread.Branches["tangerine"].Branches["tasty"].Deleted, Equals, true)
	c.Assert(treeread.Branches["tangerine"].Branches["comment"].Leaf, DeepEquals, types.Leaf{})
	c.Assert(treeread.Branches["tangerine"].Branches["tasty"].Leaf, DeepEquals, types.Leaf{})
}

func (s *DatabaseSuite) TestPatch(c *C) {
	db := Open("/tmp/summadb-test-patch")
	db.Erase()
	db = Open("/tmp/summadb-test-patch")

	// insert things by patching
	err = db.Patch([]PatchOp{
		PatchOp{types.Path{"gods", "1", "name"}, "", types.StringLeaf("zeus")},
		PatchOp{types.Path{"gods", "1", "son"}, "", types.StringLeaf("heracles")},
		PatchOp{types.Path{"gods", "2", "name"}, "", types.StringLeaf("odin")},
		PatchOp{types.Path{"gods", "2", "son"}, "", types.StringLeaf("thor")},
	})
	c.Assert(err, IsNil)

	treeread, err := db.Read(types.Path{"gods"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["1"].Branches["name"].Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["2"].Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["2"].Branches["name"].Leaf, DeepEquals, types.StringLeaf("odin"))

	// fail to patch with wrong rev
	err = db.Patch([]PatchOp{
		PatchOp{types.Path{"gods", "1", "name"}, "1-wrong", types.StringLeaf("hades")},
	})
	c.Assert(err, Not(IsNil))

	// patch some properties
	err = db.Patch([]PatchOp{
		PatchOp{types.Path{"gods", "1", "power"}, "", types.StringLeaf("thunder")},
		PatchOp{types.Path{"gods", "2", "power"}, "", types.StringLeaf("battle")},
	})
	c.Assert(err, IsNil)

	treeread, err = db.Read(types.Path{})
	c.Assert(err, IsNil)
	c.Assert(treeread.Rev, StartsWith, "2-")
	c.Assert(treeread.Branches["gods"].Rev, StartsWith, "2-")
	c.Assert(treeread.Branches["gods"].Branches["1"].Branches["power"].Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["gods"].Branches["2"].Branches["power"].Leaf, DeepEquals, types.StringLeaf("battle"))

	// patch with null to delete
	err = db.Patch([]PatchOp{
		PatchOp{types.Path{"gods", "1", "son"}, "", types.NullLeaf()},
		PatchOp{types.Path{"gods", "1", "son", "name"}, "", types.StringLeaf("thor")},
	})
	c.Assert(err, Not(IsNil) /* fails because of the rev */)
	err = db.Patch([]PatchOp{ /* try again */
		PatchOp{types.Path{"gods", "1", "son"}, treeread.Branches["gods"].Branches["1"].Branches["son"].Rev, types.NullLeaf()},
		PatchOp{types.Path{"gods", "1", "son", "name"}, "", types.StringLeaf("thor")},
	})
	c.Assert(err, IsNil)

	treeread, err = db.Read(types.Path{"gods"})
	c.Assert(err, IsNil)
	c.Assert(treeread.Rev, StartsWith, "3-")
	c.Assert(treeread.Branches["1"].Branches["power"].Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["2"].Branches["power"].Leaf, DeepEquals, types.StringLeaf("battle"))
	c.Assert(treeread.Branches["1"].Branches["son"].Rev, StartsWith, "2-")
	c.Assert(treeread.Branches["1"].Branches["son"].Leaf, DeepEquals, types.NullLeaf())
	c.Assert(treeread.Branches["1"].Branches["son"].Deleted, Equals, true)
	c.Assert(treeread.Branches["1"].Branches["son"].Branches["name"].Rev, StartsWith, "1-")
	c.Assert(treeread.Branches["1"].Branches["son"].Branches["name"].Leaf, DeepEquals, types.StringLeaf("thor"))
}
