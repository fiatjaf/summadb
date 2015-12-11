package handle

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	db "github.com/fiatjaf/summadb/database"
	"github.com/fiatjaf/summadb/handle/responses"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := getContext(r)
		user := getUser(r)

		var ok bool
		if r.Method[0] == 'P' { // PUT, POST, PATCH
			ok = db.WriteAllowedAt(ctx.path, user)
		} else { // otherwise
			ok = db.ReadAllowedAt(ctx.path, user)
		}

		if !ok {
			res := responses.Unauthorized()
			w.WriteHeader(res.Code)
			json.NewEncoder(w).Encode(res)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getUser(r *http.Request) string {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 || s[0] != "Basic" {
		return ""
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return ""
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return ""
	}

	name, password := pair[0], pair[1]
	if db.ValidUser(name, password) {
		return name
	}
	return ""
}
