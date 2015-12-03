package main

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docopt/docopt-go"
	"gopkg.in/tylerb/graceful.v1"

	db "github.com/fiatjaf/summadb/database"
	handle "github.com/fiatjaf/summadb/handle"
	settings "github.com/fiatjaf/summadb/settings"
)

func main() {
	usage := `SummaDB ` + settings.VERSION + `

Usage:
  summadb reset
  summadb
    `
	arguments, _ := docopt.Parse(usage, nil, true, settings.VERSION, false)
	reset, _ := arguments["reset"]
	if reset.(bool) {
		db.Erase()
	}

	db.Start()

	mux := handle.BuildHTTPMux()
	server := &graceful.Server{
		Timeout: 2 * time.Second,
		Server: &http.Server{
			Addr:    ":" + settings.PORT,
			Handler: mux,
		},
	}
	stop := server.StopChan()
	server.ListenAndServe()

	<-stop
	log.Info("Exiting...")
	db.End()
}
