package settings

import (
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

var VERSION = "0.1"
var STARTTIME time.Time
var DBFILE string = "/tmp/summa.db"
var PORT string = "7896"
var CORS_ORIGINS []string = []string{"*"}
var LOGLEVEL string
var DEBUG bool

func init() {
	if os.Getenv("DEBUG") == "false" {
		log.SetLevel(log.InfoLevel)
		LOGLEVEL = "info"
		DEBUG = false
	} else if os.Getenv("ENVIRONMENT") != "production" {
		log.SetLevel(log.DebugLevel)
		LOGLEVEL = "debug"
		DEBUG = true
	} else if os.Getenv("ENVIRONMENT") == "production" {
		log.SetLevel(log.ErrorLevel)
		LOGLEVEL = "error"
		DEBUG = false
	}

	STARTTIME = time.Now()
}

func HandleArgs(arguments map[string]interface{}) {
	if port, _ := arguments["--port"]; port != nil {
		PORT = port.(string)
	}
	if dbfile, _ := arguments["--db"]; dbfile != nil {
		DBFILE = dbfile.(string)
	}
	if cors_origins, _ := arguments["--cors"]; cors_origins != nil {
		CORS_ORIGINS = strings.Split(cors_origins.(string), ",")
	}
	if debug, _ := arguments["--debug"]; debug != nil && debug.(bool) {
		log.SetLevel(log.DebugLevel)
		LOGLEVEL = "debug"
		DEBUG = true
	}
}
