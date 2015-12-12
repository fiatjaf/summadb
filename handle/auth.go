package handle

import (
	"encoding/base64"
	"encoding/json"
	"log"
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

// HTTP handler for reading security metadata
func ReadSecurity(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)

	path := db.CleanPath(ctx.path)
	res := responses.Security{
		Read:  db.GetReadRuleAt(path),
		Write: db.GetWriteRuleAt(path),
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}

// HTTP handler for writing security metadata
func WriteSecurity(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)
	path := db.CleanPath(ctx.path)

	err := db.SetRulesAt(path, ctx.jsonBody)
	if err != nil {
		log.Print("unknown error: ", err)
		res := responses.UnknownError()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}
