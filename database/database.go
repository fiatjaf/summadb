package database

import (
	"github.com/fiatjaf/goleveldown"
	slu "github.com/fiatjaf/levelup/stringlevelup"
)

type SummaDB struct {
	slu.DB
	local slu.DB
}

func Open(dbpath string) *SummaDB {
	db := slu.StringDB(goleveldown.NewDatabase(dbpath))
	local := slu.StringDB(goleveldown.NewDatabase(dbpath + "_local"))
	return &SummaDB{db, local}
}

func (db *SummaDB) Erase() {
	db.DB.Erase()
	db.local.Erase()
}

func (db *SummaDB) Close() {
	db.DB.Close()
	db.local.Close()
}
