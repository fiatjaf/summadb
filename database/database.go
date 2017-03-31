package database

import (
	"github.com/fiatjaf/goleveldown"
	slu "github.com/fiatjaf/levelup/stringlevelup"
)

type SummaDB struct {
	slu.DB
}

func Open(dbpath string) *SummaDB {
	db := slu.StringDB(goleveldown.NewDatabase(dbpath))
	return &SummaDB{db}
}
