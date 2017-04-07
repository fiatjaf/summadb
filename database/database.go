package database

import (
	slu "github.com/fiatjaf/levelup/stringlevelup"
)

type Logger interface {
	Info(string, ...interface{})
	Error(string, ...interface{})
	Warn(string, ...interface{})
	Debug(string, ...interface{})
}

var log Logger

type SummaDB struct {
	slu.DB
	local slu.DB
}

func (db *SummaDB) Erase() {
	db.DB.Erase()
	db.local.Erase()
}

func (db *SummaDB) Close() {
	db.DB.Close()
	db.local.Close()
}
