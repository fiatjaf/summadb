package main

import (
	"github.com/carbocation/interpose"
	"github.com/carbocation/interpose/adaptors"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"

	"github.com/fiatjaf/summadb/context"
	"github.com/fiatjaf/summadb/handle"
)

func main() {
	// middleware
	middle := interpose.New()
	middle.Use(context.ClearContextMiddleware)
	middle.Use(adaptors.FromNegroni(cors.New(cors.Options{
		// CORS
		AllowedOrigins: []string{"*"},
	})))

	// router
	router := mux.NewRouter()
	middle.UseHandler(router)

	// create, update, delete, view values
	router.HandleFunc("/{path:.*}", handle.Values)

	log.Print("listening at port " + "5000")
	http.ListenAndServe(":"+"5000", middle)
}
