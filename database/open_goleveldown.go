// +build goleveldown

package database

import (
	"github.com/fiatjaf/goleveldown"
	slu "github.com/fiatjaf/levelup/stringlevelup"
)

func Open(dbpath string) *SummaDB {
	db := slu.StringDB(goleveldown.NewDatabase(dbpath))
	local := slu.StringDB(goleveldown.NewDatabase(dbpath + "_local"))
	return &SummaDB{db, local}
}
