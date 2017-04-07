// +build goleveldown rocksdown

package database

import (
	"github.com/fiatjaf/summadb/types"
	"github.com/fiatjaf/summadb/views"
)

func runMap(mapf string, tree types.Tree, key string) []types.EmittedRow {
	emittedrows, err := views.Map(mapf, tree, key)
	if err != nil {
		log.Error("views.Map returned error.",
			"err", err,
			"mapf", mapf,
			"docid", key)
	}
	return emittedrows
}
