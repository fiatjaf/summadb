package server

import (
	"net/http"

	"github.com/summadb/summadb/database"
)

func handlehttp(db *database.SummaDB, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
}
