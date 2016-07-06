package handle

import (
	"net/http"
	"strings"

	"github.com/justinas/alice"
	"github.com/rs/cors"

	settings "github.com/fiatjaf/summadb/settings"
)

func BuildHandler() http.Handler {
	// middleware for non-graphql endpoints
	chain := alice.New(
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

	return cors.New(cors.Options{
		AllowedOrigins:   settings.CORS_ORIGINS,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Accept", "If-Match"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler(mux)
}
