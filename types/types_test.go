package types

import (
	"testing"

	. "github.com/fiatjaf/summadb/utils"
	. "gopkg.in/check.v1"
)

var err error

func TestTypes(t *testing.T) {
	TestingT(t)
}

type TypesSuite struct{}

var _ = Suite(&TypesSuite{})

func (s *TypesSuite) TestTreeFromInterface(c *C) {
	var x interface{}
	x = map[string]interface{}{"hello": "world"}
	c.Assert(TreeFromInterface(x), DeepEquals, Tree{
		Branches: Branches{
			"hello": &Tree{Leaf: StringLeaf("world")},
		},
	})
}

func (s *TypesSuite) TestUnmarshalJSON(c *C) {
	c.Assert(TreeFromJSON(`{"a": "b"}`), DeepEquals, Tree{
		Branches: Branches{
			"a": &Tree{
				Leaf: StringLeaf("b"),
			},
		},
	})
	c.Assert(TreeFromJSON(`92`), DeepEquals, Tree{
		Leaf: NumberLeaf(92),
	})
	c.Assert(TreeFromJSON(`{"a": {"f": false, "n": null, "m": {"t": true}}}`), DeepEquals, Tree{
		Branches: Branches{
			"a": &Tree{
				Branches: Branches{
					"f": &Tree{
						Leaf: BoolLeaf(false),
					},
					"n": &Tree{
						Leaf: NullLeaf(),
					},
					"m": &Tree{
						Branches: Branches{
							"t": &Tree{
								Leaf: BoolLeaf(true),
							},
						},
					},
				},
			},
		},
	})

	c.Assert(
		TreeFromJSON(`{"_val": 12, "_rev": "2-oweqwe", "@map": "emit(1, 2)", "_del": false}`),
		DeepEquals,
		Tree{
			Leaf:     NumberLeaf(12),
			Rev:      "2-oweqwe",
			Map:      "emit(1, 2)",
			Deleted:  false,
			Branches: Branches{},
		},
	)

	c.Assert(
		TreeFromJSON(`{"subt": {"_rev": "2-oweqwe", "@map": "emit(1, 2)", "_del": true}, "_rev": "3-s5w"}`),
		DeepEquals,
		Tree{
			Rev: "3-s5w",
			Branches: Branches{
				"subt": &Tree{
					Deleted:  true,
					Rev:      "2-oweqwe",
					Map:      "emit(1, 2)",
					Branches: Branches{},
				},
			},
		},
	)
}

func (s *TypesSuite) TestMarshalJSON(c *C) {
	j, _ := (Tree{
		Branches: Branches{
			"a": &Tree{
				Leaf: StringLeaf("b"),
			},
		},
	}).MarshalJSON()
	c.Assert(j, DeepEquals, []byte(`{"a":{"_val":"b"}}`))

	j, _ = (Tree{
		Leaf: NumberLeaf(92),
	}).MarshalJSON()
	c.Assert(j, DeepEquals, []byte(`{"_val":92}`))

	j, _ = (Tree{
		Leaf: StringLeaf("www"),
		Branches: Branches{
			"a": &Tree{
				Branches: Branches{
					"f": &Tree{
						Leaf: BoolLeaf(false),
					},
					"n": &Tree{
						Leaf: NullLeaf(),
					},
					"m": &Tree{
						Branches: Branches{
							"t": &Tree{
								Leaf: BoolLeaf(true),
							},
						},
					},
				},
			},
		},
	}).MarshalJSON()
	c.Assert(
		j,
		JSONEquals,
		`{"a":{"f": {"_val":false},"n":{"_val":null},"m":{"t":{"_val":true}}},"_val":"www"}`,
	)

	j, _ = (Tree{
		Leaf:     NumberLeaf(12),
		Rev:      "2-oweqwe",
		Map:      "emit(1, 2)",
		Deleted:  false,
		Branches: Branches{},
	}).MarshalJSON()
	c.Assert(
		j,
		JSONEquals,
		`{"_val": 12, "_rev": "2-oweqwe", "@map": "emit(1, 2)", "_del": false}`,
	)

	j, _ = (Tree{
		Rev: "3-s5w",
		Branches: Branches{
			"subt": &Tree{
				Deleted:  true,
				Rev:      "2-oweqwe",
				Map:      "emit(1, 2)",
				Branches: Branches{},
			},
		},
	}).MarshalJSON()
	c.Assert(
		j,
		JSONEquals,
		`{"subt": {"_rev": "2-oweqwe", "@map": "emit(1, 2)", "_del": true}, "_rev": "3-s5w"}`,
	)
}

func (s *TypesSuite) TestPath(c *C) {
	c.Assert(ParsePath("fruits/banana"), DeepEquals, Path{"fruits", "banana"})
	c.Assert(ParsePath("fruits/banana/color").RelativeTo(ParsePath("fruits")), DeepEquals, Path{"banana", "color"})
}
