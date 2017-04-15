package server

import (
	"net/http"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/spf13/viper"
	"github.com/summadb/summadb/database"
)

var log = log15.New()

func Start(db *database.SummaDB, addr string) {
	h := Handler{db}

	spl := strings.Split(addr, "://")
	log.Info("server started.", "addr", addr)
	switch spl[0] {
	case "http":
		log.Info("server stopped", "err", http.ListenAndServe(spl[1], h))
	case "https":
		crt := viper.GetString("crt")
		key := viper.GetString("key")
		log.Info("server stopped", "err", http.ListenAndServeTLS(spl[1], crt, key, h))
	default:
		log.Error("please specify a valid address.", "addr", addr)
	}
}

type Handler struct {
	db *database.SummaDB
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") == "websocket" {
		handlewebsocket(h.db, w, r)
	} else {
		handlehttp(h.db, w, r)
	}
}
