package database

import (
	"os"

	"github.com/fiatjaf/sublevel"
)

func GetDBFile() string {
	dbfile := os.Getenv("LEVELDB_PATH")
	if dbfile == "" {
		dbfile = "/tmp/summa.db"
	}
	return dbfile
}

func Open() sublevel.AbstractLevel {
	dbfile := GetDBFile()
	return sublevel.OpenFile(dbfile, nil)
}

func Erase() error {
	dbfile := GetDBFile()
	return os.RemoveAll(dbfile)
}
