// +build levelupjs

package database

import (
	levelupjs "github.com/fiatjaf/levelup-js"
	slu "github.com/fiatjaf/levelup/stringlevelup"
)

func Open(dbpath string, adapterName string) *SummaDB {
	db := slu.StringDB(levelupjs.NewDatabase(dbpath, adapterName))
	local := slu.StringDB(levelupjs.NewDatabase(dbpath+"_local", adapterName))
	return &SummaDB{db, local}
}
