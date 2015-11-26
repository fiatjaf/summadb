package settings

import (
	"os"
	"time"
)

var STARTTIME time.Time
var PORT string

func init() {
	PORT = os.Getenv("PORT")
	if PORT == "" {
		PORT = "5000"
	}

	STARTTIME = time.Now()
}
