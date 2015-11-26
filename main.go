package main

import (
	"log"
	"net/http"

	db "github.com/fiatjaf/summadb/database"
	handle "github.com/fiatjaf/summadb/handle"
	settings "github.com/fiatjaf/summadb/settings"
)

func main() {
	middle := handle.BuildHTTPHandler()

	log.Print("listening at port " + settings.PORT + " and saving db at " + db.GetDBFile())
	http.ListenAndServe(":"+settings.PORT, middle)
}
