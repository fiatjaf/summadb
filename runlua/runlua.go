package runlua

import (
	"github.com/fiatjaf/summadb/types"
	"github.com/fiatjaf/summadb/utils"
	"github.com/yuin/gopher-lua"
)

func View(code string, t types.Tree) ([]EmittedRow, error) {
	L := lua.NewState()
	defer L.Close()

	// the 'doc'
	doc := L.CreateTable(32, 32)
	treeToLTable(L, doc, t)
	L.SetGlobal("doc", doc)

	// the 'emit' function
	var emitted []EmittedRow
	L.SetGlobal("emit", L.NewFunction(func(L *lua.LState) int {
		key := L.Get(1)
		val := L.Get(2)
		emitted = append(emitted, EmittedRow{
			Key:   utils.ToIndexable(lvalueToInterface(key)),
			Value: lvalueToInterface(val),
		})
		return 0
	}))

	err := L.DoString(code)
	return emitted, err
}

type EmittedRow struct {
	Key   []byte
	Value interface{}
}
