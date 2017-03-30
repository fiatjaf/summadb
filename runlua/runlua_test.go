package runlua

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

func (s *RunLuaSuite) TestRunView(c *C) {
	emitted, err := View(`
emit("x", {b="name"})
emit("y", {a=3, 4, 3, b=false, 12, "ss", {xx="xx"}})
emit("z", 23)
emit({"w", "m"}, "dabliuême")
emit({"boolean", true, 0}, false)
emit("r", null)
    `, types.Tree{})

	c.Assert(err, IsNil)
	c.Assert(emitted, DeeplyEquals, []EmittedRow{
		EmittedRow{ToIndexable("x"), map[string]interface{}{"b": "name"}},
		EmittedRow{ToIndexable("y"), []interface{}{
			float64(4),
			float64(3),
			float64(12),
			"ss",
			map[string]interface{}{"xx": "xx"},
		}},
		EmittedRow{ToIndexable("z"), float64(23)},
		EmittedRow{ToIndexable([]interface{}{"w", "m"}), "dabliuême"},
		EmittedRow{ToIndexable([]interface{}{"boolean", true, float64(0)}), false},
		EmittedRow{ToIndexable("r"), nil},
	})

	emitted, err = View(`
emit(doc.name._val, string.len(doc.name._val))
    `, types.Tree{
		Branches: types.Branches{
			"name": &types.Tree{
				Leaf: types.StringLeaf("mariazinha"),
			},
		},
	})

	c.Assert(err, IsNil)
	c.Assert(emitted, DeeplyEquals, []EmittedRow{
		EmittedRow{ToIndexable("mariazinha"), float64(10)},
	})
}
