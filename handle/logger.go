package handle

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := getContext(r)

		var body string
		if ctx.jsonBody != nil {
			b, _ := json.Marshal(ctx.jsonBody)
			body = string(b)
		}

		log.WithFields(log.Fields{
			"body": body,
		}).Info(r.Method + " " + r.RequestURI)

		next.ServeHTTP(w, r)
	})
}
