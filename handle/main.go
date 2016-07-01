package handle

import (
	"net/http"

	"github.com/carbocation/interpose/adaptors"
	"github.com/justinas/alice"
	"github.com/rs/cors"

	graphql "github.com/fiatjaf/summadb/handle/graphql"
)

func BuildHandler() http.Handler {
	// middleware for non-graphql endpoints
	chain := alice.New(
		setCommonVariables,
		authMiddleware,
		adaptors.FromNegroni(cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
			AllowedHeaders:   []string{"Content-Type", "Accept", "If-Match"},
			AllowCredentials: true,
		})),
	)

	// create, update, delete, view values
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			if r.URL.Path == "/_users" {
				chain.ThenFunc(CreateUser).ServeHTTP(w, r)
			} else if r.URL.Path == "/_graphql" {
				// graphql endpoint
				graphql.HandleFunc(w, r)
			} else {
				chain.ThenFunc(Post).ServeHTTP(w, r)
			}
		case "GET":
			chain.ThenFunc(Get).ServeHTTP(w, r)
		case "PUT":
			chain.ThenFunc(Put).ServeHTTP(w, r)
		case "DELETE":
			chain.ThenFunc(Delete).ServeHTTP(w, r)
		case "PATCH":
			chain.ThenFunc(Patch).ServeHTTP(w, r)
		}
	})

	return mux
}
