package settings

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
)

var VERSION = "0.1"
var STARTTIME time.Time
var DBFILE string = "/tmp/summa.db"
var PORT string = "5000"
var CORS_ORIGINS string
var LOGLEVEL string

func init() {
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
}
