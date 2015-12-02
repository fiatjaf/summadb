package main

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/tylerb/graceful.v1"

	db "github.com/fiatjaf/summadb/database"
	handle "github.com/fiatjaf/summadb/handle"
	settings "github.com/fiatjaf/summadb/settings"
)

func main() {
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
