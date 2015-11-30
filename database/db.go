package database

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/fiatjaf/sublevel"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"

	settings "github.com/fiatjaf/summadb/settings"
)

const (
	DOC_STORE = "d"
	REV_STORE = "r"
	BY_SEQ    = "s"

	UPDATE_SEQ_KEY = "_update_seq_key"
)

func Open() *sublevel.AbstractLevel {
	dbfile := settings.DBFILE
	db, err := sublevel.Open(dbfile, &opt.Options{
		Filter: filter.NewBloomFilter(10),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"DBFILE": settings.DBFILE,
		}).Fatal("couldn't open database file.")
	}
	return db
}

func Erase() error {
	dbfile := settings.DBFILE
	return os.RemoveAll(dbfile)
}
