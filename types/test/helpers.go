package test

import "github.com/fiatjaf/summadb/types"

func treeFromJSON(j string) types.Tree {
	t := &types.Tree{}
	t.UnmarshalJSON([]byte(j))
	return *t
}
