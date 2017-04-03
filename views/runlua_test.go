package views

import (
	"testing"

	"github.com/fiatjaf/summadb/types"
	. "github.com/fiatjaf/summadb/utils"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type RunLuaSuite struct{}

var _ = Suite(&RunLuaSuite{})

func (s *RunLuaSuite) TestRunMap(c *C) {
	emitted, err := Map(`
emit("x", {b="name"})
emit("y", {a=3, l={xx="xx"}})
emit("z", 23, 18)
emit("w", "m", "dabliuême")
emit("r", null)
    `, types.Tree{}, "")

	c.Assert(err, IsNil)
	c.Assert(emitted, DeeplyEquals, []EmittedRow{
		EmittedRow{types.Path{"x"}, types.TreeFromJSON(`{"b": "name"}`)},
		EmittedRow{types.Path{"y"}, types.TreeFromJSON(`{
            "a": 3,
            "l": {
              "xx": "xx"
            }
		}`)},
		EmittedRow{types.Path{"z", "23"}, types.TreeFromJSON(`18`)},
		EmittedRow{types.Path{"w", "m"}, types.TreeFromJSON(`"dabliuême"`)},
		EmittedRow{types.Path{"r"}, types.TreeFromJSON(`1`)},
	})

	emitted, err = Map(`
emit('name-lengths', doc.name._val, string.len(doc.name._val))
    `, types.Tree{
		Branches: types.Branches{
			"name": &types.Tree{
				Leaf: types.StringLeaf("mariazinha"),
			},
		},
	}, "")

	c.Assert(err, IsNil)
	c.Assert(emitted, DeeplyEquals, []EmittedRow{
		EmittedRow{types.Path{"name-lengths", "mariazinha"}, types.TreeFromJSON(`10`)},
	})
}
