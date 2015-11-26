package handle

import (
	"github.com/carbocation/interpose"
	"github.com/carbocation/interpose/adaptors"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func BuildHTTPHandler() *interpose.Middleware {
	// middleware
	middle := interpose.New()
	middle.Use(setCommonVariables)
	middle.Use(adaptors.FromNegroni(cors.New(cors.Options{
		// CORS
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Accept", "If-Match"},
		AllowCredentials: true,
	})))

	// router
	router := mux.NewRouter()
	middle.UseHandler(router)

	// create, update, delete, view values
	router.HandleFunc("/{path:.*}", Get).Methods("GET")
	router.HandleFunc("/{path:.*}", Put).Methods("PUT")
	router.HandleFunc("/{path:.*}", Patch).Methods("PATCH")
	router.HandleFunc("/{path:.*}", Delete).Methods("DELETE")

	return middle
}
