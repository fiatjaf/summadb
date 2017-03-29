package runlua

import (
	"testing"

	"github.com/fiatjaf/summadb/types"
	. "github.com/fiatjaf/summadb/utils"
	. "gopkg.in/check.v1"
)

var err error

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
emit("w", "xuyxuxu")
emit("i", true)
emit("r", null)
    `, types.Tree{})

	c.Assert(err, IsNil)
	c.Assert(emitted, DeepEquals, map[string]interface{}{
		"w": "xuyxuxu",
		"i": bool(true),
		"r": nil,
		"x": map[string]interface{}{
			"b": "name",
		},
		"y": map[string]interface{}{
			"4": "ss",
			"5": map[string]interface{}{
				"xx": "xx",
			},
			"a": float64(3),
			"b": bool(false),
			"1": float64(4),
			"2": float64(3),
			"3": float64(12),
		},
		"z": float64(23),
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
	c.Assert(emitted, DeeplyEquals, map[string]interface{}{
		"mariazinha": float64(10),
	})
}
