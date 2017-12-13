package views

import (
	"log"

	"github.com/summadb/summadb/types"
	"github.com/summadb/summadb/utils"
	"github.com/yuin/gopher-lua"
)

func Map(code string, t types.Tree, key string) ([]types.EmittedRow, error) {
	L := lua.NewState()
	defer L.Close()

	// the 'doc'
	L.SetGlobal("doc", treeToLTable(L, t))

	// the "_key"
	L.SetGlobal("_key", lua.LString(key))

	// the 'emit' function
	var emitted []types.EmittedRow
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

		emitted = append(emitted, types.EmittedRow{RelativePath: path, Value: value})
		return 0
	}))

	// the 'indexify' function
	L.SetGlobal("indexify", createIndexify(L))

	err := L.DoString(code)
	return emitted, err
}

func Reduce(
	code string,
	directive string,
	acc types.Tree,
	row types.EmittedRow,
	key string,
) (types.Tree, error) {
	L := lua.NewState()
	defer L.Close()

	// the '_key' of the original record being mapped
	L.SetGlobal("_key", lua.LString(key))

	// the 'path' emitted by the mapf
	lpath := L.CreateTable(32, 32)
	for _, k := range row.RelativePath {
		lpath.Append(lua.LString(k))
	}
	L.SetGlobal("path", lpath)

	// the 'value' emitted by the mapf
	L.SetGlobal("value", treeToLTable(L, row.Value))

	// the 'directive': "add" or "remove"
	L.SetGlobal("directive", lua.LString(directive))

	// the previous value, what is being reduced
	L.SetGlobal("acc", treeToLTable(L, acc))

	// the 'indexify' function
	L.SetGlobal("indexify", createIndexify(L))

	log.Print("running reducef: path=", row.RelativePath, " value=", row.Value, " directive=", directive, " acc=", acc)

	err := L.DoString(code)
	if err != nil {
		return types.Tree{}, err
	}

	output := L.GetGlobal("acc")
	return types.TreeFromInterface(lvalueToInterface(output)), nil
}

func createIndexify(L *lua.LState) lua.LValue {
	return L.NewFunction(func(L *lua.LState) int {
		/* return after converting to string from ToIndexable []byte the first argument */
		L.Push(lua.LString(string(utils.ToIndexable(lvalueToInterface(L.Get(1))))))
		return 1
	})
}
