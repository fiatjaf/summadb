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
	doc.RawSetString("_id", lua.LString(docid))
	L.SetGlobal("doc", doc)

	// the 'emit' function
	var emitted []EmittedRow
	var i = 0
	L.SetGlobal("emit", L.NewFunction(func(L *lua.LState) int {
		key := L.Get(1)
		val := L.Get(2)
		emitted = append(emitted, EmittedRow{
			Key: string(utils.ToIndexable([]interface{}{
				lvalueToInterface(key),
				docid, /* the doc _id */
				i,     /* to make it unique */
			})),
			Value: types.TreeFromInterface(lvalueToInterface(val)),
		})
		i++
		return 0
	}))

	err := L.DoString(code)
	return emitted, err
}

type EmittedRow struct {
	Key   string
	Value types.Tree
}
