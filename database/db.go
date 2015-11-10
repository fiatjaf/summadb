package database

import (
	"os"

	"github.com/fiatjaf/sublevel"
)

func getDBFile() string {
	dbfile := os.Getenv("LEVELDB_PATH")
	if dbfile == "" {
		dbfile = "/tmp/summa.db"
	}
	return dbfile
}

func Open() sublevel.AbstractLevel {
	dbfile := getDBFile()
	return sublevel.OpenFile(dbfile, nil)
}

func Erase() error {
	dbfile := getDBFile()
	return os.RemoveAll(dbfile)
}
