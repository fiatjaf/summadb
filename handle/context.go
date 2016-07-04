package handle

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/context"

	db "github.com/fiatjaf/summadb/database"
	responses "github.com/fiatjaf/summadb/handle/responses"
)

const k int = iota

func getContext(r *http.Request) Context {
	return context.Get(r, k).(Context)
}

type Context struct {
	user string

	body              []byte
	jsonBody          map[string]interface{}
	path              string
	lastKey           string
	wantsTree         bool
	localDoc          bool
	wantsDatabaseInfo bool

	currentRev  string
	deleted     bool
	exists      bool
	providedRev string
}

func createContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context.Set(r, k, Context{})
		next.ServeHTTP(w, r)
		context.Clear(r) // clears after handling everything.
	})
}

func setCommonVariables(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := getContext(r)

		/* when the path is in the format /nanana/nanana/_val
		   we reply with the single value for that path, otherwise
		   assume the whole tree is being requested. */
		ctx.path = r.URL.Path
		ctx.wantsTree = true
		ctx.localDoc = false
		ctx.wantsDatabaseInfo = false

		entityPath := ctx.path
		pathKeys := db.SplitKeys(ctx.path)
		nkeys := len(pathKeys)
		ctx.lastKey = pathKeys[nkeys-1]

		if nkeys > 1 && pathKeys[nkeys-2] == "_local" {
			// workaround for couchdb-like _local/blablabla docs
			// we use a different sublevel for local docs so we must
			// acknowledge that somehow.
			ctx.lastKey = db.UnescapeKey(db.JoinKeys(pathKeys[nkeys-2:]))
			ctx.localDoc = true
		} else if ctx.lastKey == "" {
			// this means the request was made with an ending slash (for example:
			// https://my.summadb.com/path/to/here/), so it wants couchdb-like information
			// for the referred sub-database, and not the document at the referred path.
			// to get information on the document at the referred path the request must be
			// made without the trailing slash (or with a special header, for the root document
			// and other scenarios in which the user does not have control over the presence
			// of the ending slash).
			ctx.wantsDatabaseInfo = true

			if ctx.path == "/" {
				ctx.path = ""
				ctx.lastKey = ""
				entityPath = ""
			}
		} else {
			if ctx.lastKey[0] == '_' {
				ctx.wantsTree = false
				entityPath = db.JoinKeys(pathKeys[:nkeys-1])
				if ctx.lastKey == "_val" {
					ctx.path = db.JoinKeys(pathKeys[:nkeys-1])
					ctx.lastKey = pathKeys[nkeys-1]
				}
			}
		}

		ctx.currentRev = ""
		ctx.deleted = false
		qrev := r.URL.Query().Get("rev")
		hrev := r.Header.Get("If-Match")
		drev := ""
		ctx.providedRev = hrev // will be "" if there's no header rev

		// fetching current rev
		if !ctx.localDoc {
			// procedure for normal documents

			specialKeys, err := db.GetSpecialKeysAt(entityPath)
			if err == nil {
				ctx.currentRev = specialKeys.Rev
				ctx.deleted = specialKeys.Deleted
			}

		} else {
			// procedure for local documents
			ctx.currentRev = db.GetLocalDocRev(ctx.path)
		}

		// body parsing
		if r.Method[0] == 'P' { // PUT, PATCH, POST
			/* filter body size */
			b, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
			ctx.body = b
			if err != nil {
				res := responses.BadRequest("request body too large")
				w.WriteHeader(res.Code)
				json.NewEncoder(w).Encode(res)
				return
			}

			if ctx.lastKey != "_val" {
				err = json.Unmarshal(ctx.body, &ctx.jsonBody)
				if err != nil {
					res := responses.BadRequest("invalid JSON sent as JSON")
					w.WriteHeader(res.Code)
					json.NewEncoder(w).Encode(res)
				}
			}
		}

		revfail := false
		// rev checking
		if qrev != "" {
			if hrev != "" && qrev != hrev {
				revfail = true
			} else {
				ctx.providedRev = qrev
			}
		}
		drev = ""
		if drevi, ok := ctx.jsonBody["_rev"]; ok {
			drev = drevi.(string)
		}
		if drev != "" {
			if qrev != "" && qrev != drev {
				revfail = true
			} else if hrev != "" && hrev != drev {
				revfail = true
			} else if qrev != "" && hrev != "" && qrev != hrev {
				revfail = true
			} else {
				ctx.providedRev = drev
			}
		}

		if revfail {
			res := responses.BadRequest("different rev values were sent")
			w.WriteHeader(res.Code)
			json.NewEncoder(w).Encode(res)
			return
		}

		ctx.exists = ctx.currentRev != "" && !ctx.deleted
		context.Set(r, k, ctx)

		next.ServeHTTP(w, r)
	})
}

func setUserVariable(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := getContext(r)
		ctx.user = getUser(r)
		context.Set(r, k, ctx)
		next.ServeHTTP(w, r)
	})
}
