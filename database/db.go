package database

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/fiatjaf/sublevel"

	settings "github.com/fiatjaf/summadb/settings"
)

const (
	DOC_STORE = "doc-store"
	BY_SEQ    = "by-seq"

	UPDATE_SEQ_KEY = "_update_seq_key"
)

func Open() *sublevel.AbstractLevel {
	dbfile := settings.DBFILE
	db, err := sublevel.Open(dbfile, nil)
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
