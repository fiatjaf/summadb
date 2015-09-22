package db

import (
	"github.com/syndtr/goleveldb/leveldb"
	"log"
)

func OpenDB() *leveldb.DB {
	db, err := leveldb.OpenFile("db.example", nil)
	if err != nil {
		log.Print("couldn't open database file.")
	}
	return db
}
