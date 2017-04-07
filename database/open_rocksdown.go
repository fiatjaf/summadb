// +build rocksdown

package database

import (
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/fiatjaf/rocksdown"
)

func Open(dbpath string) *SummaDB {
	db := slu.StringDB(rocksdown.NewDatabase(dbpath))
	local := slu.StringDB(rocksdown.NewDatabase(dbpath + "_local"))
	return &SummaDB{db, local}
}
