package handle

import (
	"net/http"
	"strings"

	"github.com/justinas/alice"
	"github.com/rs/cors"
)

var corsMiddleware = cors.New(cors.Options{
	AllowedOrigins:   []string{"*"},
	AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	AllowedHeaders:   []string{"Content-Type", "Accept", "If-Match"},
	AllowCredentials: true,
}).Handler

func BuildHandler() http.Handler {
	// middleware for non-graphql endpoints
	chain := alice.New(
		corsMiddleware,
		createContext,
		setCommonVariables,
		setUserVariable,
		authMiddleware,
	)

	// create, update, delete, view values
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			if r.URL.Path == "/_users" {
				chain.ThenFunc(CreateUser).ServeHTTP(w, r)
			} else if strings.HasSuffix(r.URL.Path, "/_graphql") {
				alice.New(
					createContext,
					setUserVariable,
					corsMiddleware,
				).ThenFunc(HandleGraphQL).ServeHTTP(w, r)
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
