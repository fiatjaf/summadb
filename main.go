package summadb

import (
	"log"
	"net/http"

	db "github.com/fiatjaf/summadb/database"
	handle "github.com/fiatjaf/summadb/handle"
)

func main() {
	middle := handle.BuildHTTPHandler()

	log.Print("listening at port " + "5000" + " and saving db at " + db.GetDBFile())
	http.ListenAndServe(":"+"5000", middle)
}
