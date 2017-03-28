package database

import (
	"encoding/json"
	"testing"

	. "gopkg.in/check.v1"
)

var err error

func Test(t *testing.T) {
	TestingT(t)
}

type TypesSuite struct{}

var _ = Suite(&TypesSuite{})

func (s *TypesSuite) TestUnmarshalJSON(c *C) {
	c.Assert(Tree(`{"a": "b"}`), DeepEquals, tree{
		subtrees: map[string]tree{
			"a": tree{
				leaf: leaf{kind: STRING, string: "b"},
			},
		},
	})
	c.Assert(Tree(`92`), DeepEquals, tree{
		leaf: leaf{kind: NUMBER, float64: 92},
	})
	c.Assert(Tree(`{"a": {"f": false, "n": null, "m": {"t": true}}}`), DeepEquals, tree{
		subtrees: map[string]tree{
			"a": tree{
				subtrees: map[string]tree{
					"f": tree{
						leaf: leaf{kind: BOOL, bool: false},
					},
					"n": tree{
						leaf: leaf{kind: NULL},
					},
					"m": tree{
						subtrees: map[string]tree{
							"t": tree{
								leaf: leaf{kind: BOOL, bool: true},
							},
						},
					},
				},
			},
		},
	})
}

func (s *TypesSuite) TestMarshalJSON(c *C) {
	j, _ := (tree{
		subtrees: map[string]tree{
			"a": tree{
				leaf: leaf{kind: STRING, string: "b"},
			},
		},
	}).MarshalJSON()
	c.Assert(j, DeepEquals, []byte(`{"a":{"_val":"b"}}`))

	j, _ = (tree{
		leaf: leaf{kind: NUMBER, float64: 92},
	}).MarshalJSON()
	c.Assert(j, DeepEquals, []byte(`{"_val":92}`))

	j, _ = (tree{
		leaf: leaf{kind: STRING, string: "www"},
		subtrees: map[string]tree{
			"a": tree{
				subtrees: map[string]tree{
					"f": tree{
						leaf: leaf{kind: BOOL, bool: false},
					},
					"n": tree{
						leaf: leaf{kind: NULL},
					},
					"m": tree{
						subtrees: map[string]tree{
							"t": tree{
								leaf: leaf{kind: BOOL, bool: true},
							},
						},
					},
				},
			},
		},
	}).MarshalJSON()

	// since keys are out of order here, let's do this comparison in a bizarre way:
	var obtained interface{}
	json.Unmarshal(j, &obtained)
	var expected interface{}
	json.Unmarshal(
		[]byte(`{"a":{"f": {"_val":false},"n":{"_val":null},"m":{"t":{"_val":true}}},"_val":"www"}`),
		&expected,
	)
	c.Assert(obtained, DeepEquals, expected)
}
