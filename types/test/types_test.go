package test

import (
	"testing"

	"github.com/fiatjaf/summadb/types"
	. "github.com/fiatjaf/summadb/utils"
	. "gopkg.in/check.v1"
)

var err error

func TestTypes(t *testing.T) {
	TestingT(t)
}

type TypesSuite struct{}

var _ = Suite(&TypesSuite{})

func (s *TypesSuite) TestUnmarshalJSON(c *C) {
	c.Assert(treeFromJSON(`{"a": "b"}`), DeepEquals, types.Tree{
		Branches: types.Branches{
			"a": &types.Tree{
				Leaf: types.StringLeaf("b"),
			},
		},
	})
	c.Assert(treeFromJSON(`92`), DeepEquals, types.Tree{
		Leaf: types.NumberLeaf(92),
	})
	c.Assert(treeFromJSON(`{"a": {"f": false, "n": null, "m": {"t": true}}}`), DeepEquals, types.Tree{
		Branches: types.Branches{
			"a": &types.Tree{
				Branches: types.Branches{
					"f": &types.Tree{
						Leaf: types.BoolLeaf(false),
					},
					"n": &types.Tree{
						Leaf: types.NullLeaf(),
					},
					"m": &types.Tree{
						Branches: types.Branches{
							"t": &types.Tree{
								Leaf: types.BoolLeaf(true),
							},
						},
					},
				},
			},
		},
	})
}

func (s *TypesSuite) TestMarshalJSON(c *C) {
	j, _ := (types.Tree{
		Branches: types.Branches{
			"a": &types.Tree{
				Leaf: types.StringLeaf("b"),
			},
		},
	}).MarshalJSON()
	c.Assert(j, DeepEquals, []byte(`{"a":{"_val":"b"}}`))

	j, _ = (types.Tree{
		Leaf: types.NumberLeaf(92),
	}).MarshalJSON()
	c.Assert(j, DeepEquals, []byte(`{"_val":92}`))

	j, _ = (types.Tree{
		Leaf: types.StringLeaf("www"),
		Branches: types.Branches{
			"a": &types.Tree{
				Branches: types.Branches{
					"f": &types.Tree{
						Leaf: types.BoolLeaf(false),
					},
					"n": &types.Tree{
						Leaf: types.NullLeaf(),
					},
					"m": &types.Tree{
						Branches: types.Branches{
							"t": &types.Tree{
								Leaf: types.BoolLeaf(true),
							},
						},
					},
				},
			},
		},
	}).MarshalJSON()

	c.Assert(j, JSONEquals, `{"a":{"f": {"_val":false},"n":{"_val":null},"m":{"t":{"_val":true}}},"_val":"www"}`)
}
