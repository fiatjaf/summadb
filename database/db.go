package database

import (
	"os"

	"github.com/fiatjaf/sublevel"
)

const (
	DOC_STORE = "doc-store"
	BY_SEQ    = "by-seq"
)

func GetDBFile() string {
	dbfile := os.Getenv("LEVELDB_PATH")
	if dbfile == "" {
		dbfile = "/tmp/summa.db"
	}
	return dbfile
}

func Open() *sublevel.AbstractLevel {
	dbfile := GetDBFile()
	return sublevel.MustOpen(dbfile, nil)
}

func Erase() error {
	dbfile := GetDBFile()
	return os.RemoveAll(dbfile)
}
