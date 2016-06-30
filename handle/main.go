package handle

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/carbocation/interpose/adaptors"
	"github.com/fiatjaf/summadb/graphql"
	"github.com/graphql-go/handler"
	"github.com/justinas/alice"
	"github.com/rs/cors"
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

	// graphql endpoint
	schema, err := graphql.MakeSchema()
	if err != nil {
		log.Warn("error creating graphql schema: ", err.Error())
	}

	// create, update, delete, view values
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			if r.URL.Path == "/_users" {
				chain.ThenFunc(CreateUser).ServeHTTP(w, r)
			} else if r.URL.Path == "/_graphql" {
				gql := handler.New(&handler.Config{
					Schema: &schema,
					Pretty: true,
				})
				gql.ServeHTTP(w, r)
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
