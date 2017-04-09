package views

import (
	"github.com/summadb/summadb/types"
	"github.com/yuin/gopher-lua"
	. "gopkg.in/check.v1"
)

type HelpersSuite struct{}

var _ = Suite(&HelpersSuite{})

func (s *HelpersSuite) TestTypeConversion(c *C) {
	L := lua.NewState()
	defer L.Close()

	table := L.CreateTable(32, 32)
	table.RawSetString("a key", lua.LString("a value"))
	table.RawSetString("another key", lua.LString("another value"))
	c.Assert(toArray(table), IsNil)
	c.Assert(lvalueToInterface(table), DeepEquals, map[string]interface{}{
		"a key":       "a value",
		"another key": "another value",
	})

	table.Append(lua.LNumber(1))
	table.Append(lua.LNumber(2))
	table.Append(lua.LNumber(3))
	c.Assert(toArray(table), Not(IsNil))
	c.Assert(lvalueToInterface(table), DeepEquals, []interface{}{
		float64(1),
		float64(2),
		float64(3),
	})

	c.Assert(lvalueToInterface(lua.LString("banana")), DeepEquals, "banana")
	c.Assert(lvalueToInterface(lua.LNumber(64)), DeepEquals, float64(64))

	table2 := L.CreateTable(32, 32)
	table2.Append(lua.LNumber(117))
	table2.Append(lua.LString("bala"))
	table2.Append(lua.LBool(false))

	c.Assert(toArray(table2), Not(IsNil))
	c.Assert(lvalueToInterface(table2), DeepEquals, []interface{}{
		float64(117),
		"bala",
		false,
	})

	table.Append(table2)
	c.Assert(lvalueToInterface(table), DeepEquals, []interface{}{
		float64(1),
		float64(2),
		float64(3),
		[]interface{}{
			float64(117),
			"bala",
			false,
		},
	})

	table3 := L.CreateTable(32, 32)
	tree := types.Tree{
		Leaf: types.StringLeaf("banana"),
		Branches: types.Branches{
			"kind": &types.Tree{Leaf: types.StringLeaf("fruit")},
			"size": &types.Tree{Leaf: types.NumberLeaf(12)},
			"places": &types.Tree{
				Branches: types.Branches{
					"0": &types.Tree{Leaf: types.StringLeaf("equador")},
					"1": &types.Tree{Leaf: types.StringLeaf("brazil")},
				},
			},
		},
	}
	treeToLTable(L, table3, tree)

	table3expected := L.CreateTable(32, 32)
	table3expected.RawSetString("_val", lua.LString("banana"))
	tkind := L.CreateTable(32, 32)
	tkind.RawSetString("_val", lua.LString("fruit"))
	table3expected.RawSetString("kind", tkind)
	tsize := L.CreateTable(32, 32)
	tsize.RawSetString("_val", lua.LNumber(12))
	table3expected.RawSetString("size", tsize)
	tplaces := L.CreateTable(32, 32)
	tequador := L.CreateTable(32, 32)
	tbrazil := L.CreateTable(32, 32)
	tequador.RawSetString("_val", lua.LString("equador"))
	tbrazil.RawSetString("_val", lua.LString("brazil"))
	tplaces.RawSetString("0", tequador)
	tplaces.RawSetString("1", tbrazil)
	table3expected.RawSetString("places", tplaces)

	c.Assert(table3, DeepEquals, table3expected)
}
