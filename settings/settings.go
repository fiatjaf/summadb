package settings

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
)

var VERSION = "0.1"
var STARTTIME time.Time
var DBFILE string
var PORT string
var CORS_ORIGINS string
var LOGLEVEL string

func init() {
	DBFILE = os.Getenv("LEVELDB_PATH")
	if DBFILE == "" {
		DBFILE = "/tmp/summa.db"
	}
	PORT = os.Getenv("PORT")
	if PORT == "" {
		PORT = "5000"
	}

	if os.Getenv("DEBUG") == "false" {
		log.SetLevel(log.InfoLevel)
		LOGLEVEL = "info"
	} else if os.Getenv("ENVIRONMENT") != "production" {
		log.SetLevel(log.DebugLevel)
		LOGLEVEL = "debug"
	} else if os.Getenv("ENVIRONMENT") == "production" {
		log.SetLevel(log.ErrorLevel)
		LOGLEVEL = "error"
	}

	CORS_ORIGINS = "*"
	STARTTIME = time.Now()

	log.WithFields(log.Fields{
		"DBFILE":       DBFILE,
		"PORT":         PORT,
		"STARTTIME":    STARTTIME,
		"CORS_ORIGINS": CORS_ORIGINS,
		"LOGLEVEL":     LOGLEVEL,
	}).Debug("coming up with settings for the database and server.")
}
