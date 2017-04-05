package views

import (
	"github.com/fiatjaf/summadb/types"
	"github.com/fiatjaf/summadb/utils"
	"github.com/yuin/gopher-lua"
)

func Map(code string, t types.Tree, docid string) ([]EmittedRow, error) {
	L := lua.NewState()
	defer L.Close()

	// the 'doc'
	doc := L.CreateTable(32, 32)
	treeToLTable(L, doc, t)
	doc.RawSetString("_key", lua.LString(docid))
	L.SetGlobal("doc", doc)

	// the 'emit' function
	var emitted []EmittedRow
	L.SetGlobal("emit", L.NewFunction(func(L *lua.LState) int {
		path := types.Path{}

		// all string arguments (except the last, if we reach it) will constitute the path
		narg := 1
		for ; ; narg++ {
			lvarg := L.Get(narg)
			if lvarg == lua.LNil {
				// null, means we reached the end. the previous argument must be the value.
				path = path[:len(path)-1]
				narg--
				break
			}
			arg := lua.LVAsString(lvarg)
			if arg == "" {
				// not a string or number (which can be cast to string), so this is the value.
				break
			}

			// a valid string, it will be part of the full path of this emitted item
			// if the user wants to use arrays as part of the path (i.e. keys) he can
			// use the provided function 'indexify' on them.
			path = append(path, arg)
		}

		var value types.Tree
		if narg == 0 {
			// wrong. ignore
			return 0
		} else if narg > 1 {
			// ok, expected.
			value = types.TreeFromInterface(lvalueToInterface(L.Get(narg)))
		} else {
			// the user has only passed 1 argument to 'emit',
			// so use it as key and set the value to a dummy 1.
			path = append(path, L.ToString(narg))
			value = types.Tree{Leaf: types.NumberLeaf(1)}
		}

		emitted = append(emitted, EmittedRow{RelativePath: path, Value: value})
		return 0
	}))

	// the 'indexify' function
	L.SetGlobal("indexify", L.NewFunction(func(L *lua.LState) int {
		/* return after converting to string from ToIndexable []byte the first argument */
		L.Push(lua.LString(string(utils.ToIndexable(lvalueToInterface(L.Get(1))))))
		return 1
	}))

	err := L.DoString(code)
	return emitted, err
}

type EmittedRow struct {
	RelativePath types.Path
	Value        types.Tree
}
