package database

import (
	"github.com/fiatjaf/sublevel"
	"os"
)

func Open() sublevel.AbstractLevel {
	dbfile := os.Getenv("LEVELDB_PATH")
	if dbfile == "" {
		dbfile = "/tmp/summa.db"
	}

	return sublevel.OpenFile(dbfile, nil)
}
