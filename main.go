package main

import (
	"net/http"

	handle "github.com/fiatjaf/summadb/handle"
	settings "github.com/fiatjaf/summadb/settings"
)

func main() {
	middle := handle.BuildHTTPHandler()

	http.ListenAndServe(":"+settings.PORT, middle)
}
