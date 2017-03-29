package runlua

import (
	"github.com/fiatjaf/summadb/types"
	"github.com/yuin/gopher-lua"
)

func View(code string, t types.Tree) (map[string]interface{}, error) {
	L := lua.NewState()
	defer L.Close()

	// the 'doc'
	doc := L.CreateTable(32, 32)
	treeToLTable(L, doc, t)
	L.SetGlobal("doc", doc)

	// the 'emit' function
	emitted := map[string]interface{}{}
	L.SetGlobal("emit", L.NewFunction(func(L *lua.LState) int {
		key := L.ToString(1)
		val := L.Get(2)
		emitted[key] = lvalueToInterface(val)
		return 0
	}))

	err := L.DoString(code)
	return emitted, err
}

func lvalueToInterface(lvalue lua.LValue) interface{} {
	switch value := lvalue.(type) {
	case *lua.LTable:
		m := make(map[string]interface{}, value.Len())
		value.ForEach(func(k lua.LValue, v lua.LValue) {
			m[lua.LVAsString(k)] = lvalueToInterface(v)
		})
		return m
	case lua.LNumber:
		return float64(value)
	case lua.LString:
		return string(value)
	default:
		switch lvalue {
		case lua.LTrue:
			return true
		case lua.LFalse:
			return false
		case lua.LNil:
			return nil
		}
	}
	return nil
}

func treeToLTable(L *lua.LState, table *lua.LTable, t types.Tree) {
	for key, subtree := range t.Branches {
		subtable := L.CreateTable(32, 32)
		treeToLTable(L, subtable, *subtree)
		table.RawSetString(key, subtable)
	}

	var leafvalue lua.LValue
	switch t.Leaf.Kind {
	case types.STRING:
		leafvalue = lua.LString(t.Leaf.String())
	case types.NUMBER:
		leafvalue = lua.LNumber(t.Leaf.Number())
	case types.BOOL:
		leafvalue = lua.LBool(t.Leaf.Bool())
	case types.NULL:
		leafvalue = lua.LNil
	default:
		return
	}
	table.RawSetString("_val", leafvalue)
}
