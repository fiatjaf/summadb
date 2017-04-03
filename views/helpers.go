package views

import (
	"github.com/fiatjaf/summadb/types"
	"github.com/yuin/gopher-lua"
)

func lvalueToInterface(lvalue lua.LValue) interface{} {
	switch value := lvalue.(type) {
	case *lua.LTable:
		array := toArray(value)
		if array != nil {
			return array
		} else {
			m := make(map[string]interface{}, value.Len())
			value.ForEach(func(k lua.LValue, v lua.LValue) {
				m[lua.LVAsString(k)] = lvalueToInterface(v)
			})
			return m
		}
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

// toArray returns nil if the table is not a proper array.
func toArray(table *lua.LTable) []interface{} {
	// check if it is an array: http://stackoverflow.com/a/7527013/973380
	if table.RawGetInt(1) == lua.LNil {
		// this is the only check we'll do. don't blame us.
		return nil
	}

	// now we can safely treat it as an array and ignore non-array keys.
	array := make([]interface{}, table.Len())
	for i := 0; i < len(array); i++ {
		lvalue := table.RawGetInt(i + 1 /* lua tables are 1-indexed */)
		array[i] = lvalueToInterface(lvalue)
	}
	return array
}
