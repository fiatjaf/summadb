package handle

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	db "github.com/fiatjaf/summadb/database"

	"github.com/gorilla/context"
)

const k int = iota

func getContext(r *http.Request) common {
	return context.Get(r, k).(common)
}

type options struct {
	include_docs bool
	descending   bool
	startkey     bool
	endkey       bool
	revs         bool
	open_revs    bool
}
type common struct {
	opts options

	body      []byte
	jsonBody  map[string]interface{}
	path      string
	pathKeys  []string
	nkeys     int
	lastKey   string
	wantsTree bool

	actualRev   string
	deleted     bool
	exists      bool
	providedRev string
}

func setCommonVariables(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/* query parameters and options (with defaults) */
		opts := options{
			revs:         r.URL.Query().Get("revs") == "true",
			include_docs: r.URL.Query().Get("include_docs") == "true",
			descending:   r.URL.Query().Get("descending") == "true",
			startkey:     r.URL.Query().Get("startkey") == "true",
			endkey:       r.URL.Query().Get("endkey") == "true",
			open_revs:    r.URL.Query().Get("openrevs") == "all",
		}

		/* when the path is in the format /nanana/nanana/_val
		   we reply with the single value for that path, otherwise
		   assume the whole tree is being requested. */
		path := r.URL.Path
		wantsTree := true
		entityPath := path
		pathKeys := db.SplitKeys(path)
		nkeys := len(pathKeys)
		var lastKey string
		if path == "/" {
			path = ""
			lastKey = ""
			entityPath = ""
		} else {
			lastKey = pathKeys[nkeys-1]
			if lastKey[0] == '_' {
				wantsTree = false
				entityPath = db.JoinKeys(pathKeys[:nkeys-1])
				if lastKey == "_val" {
					path = db.JoinKeys(pathKeys[:nkeys-1])
					lastKey = pathKeys[nkeys-1]
				}
			}
		}

		actualRev := ""
		deleted := false
		specialKeys, err := db.GetSpecialKeysAt(entityPath)
		if err == nil {
			actualRev = specialKeys.Rev
			deleted = specialKeys.Deleted
		}

		revfail := false
		qrev := r.URL.Query().Get("rev")
		hrev := r.Header.Get("If-Match")
		providedRev := hrev // will be "" if there's no header rev
		if qrev != "" {
			if hrev != "" && qrev != hrev {
				revfail = true
			} else {
				providedRev = qrev
			}
		}

		var jsonBody map[string]interface{}
		var body []byte

		if r.Method == "PUT" || r.Method == "POST" {
			/* filter body size */
			body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
			if err != nil {
				log.Print("couldn't read request body: ", err)
				res := badRequest("request body too large")
				w.WriteHeader(res.code)
				json.NewEncoder(w).Encode(res)
				return
			}

			isTree := false
			bodyType := r.Header.Get("Content-Type")
			if lastKey != "_val" &&
				(bodyType == "application/json" || bodyType == "text/json") {
				isTree = true
			}

			if isTree {
				err = json.Unmarshal(body, &jsonBody)
				if err != nil {
					log.Print("invalid JSON sent as JSON: ", err)
					res := badRequest("invalid JSON sent as JSON")
					w.WriteHeader(res.code)
					json.NewEncoder(w).Encode(res)
				}
			}

			drev := ""
			if drevi, ok := jsonBody["_rev"]; ok {
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
					providedRev = drev
				}
			}
		}

		if revfail {
			log.Print("multiple revs mismatching.")
			res := badRequest("different rev values were sent")
			w.WriteHeader(res.code)
			json.NewEncoder(w).Encode(res)
			return
		}

		context.Set(r, k, common{
			opts: opts,

			body:      body,
			jsonBody:  jsonBody,
			path:      path,
			pathKeys:  pathKeys,
			nkeys:     nkeys,
			lastKey:   lastKey,
			wantsTree: wantsTree,

			actualRev:   actualRev,
			deleted:     deleted,
			exists:      actualRev != "" && !deleted,
			providedRev: providedRev,
		})

		next.ServeHTTP(w, r)
		context.Clear(r) // clears after handling everything.
	})
}

//func LoadDB(r *http.Request) *leveldb.DB {
//	if val := context.Get(r, k); val != nil {
//		return val.(*leveldb.DB)
//	}
//	return database.OpenDB()
//}
//
//func StoreDB(r *http.Request, db *leveldb.DB) {
//	context.Set(r, k, db)
//	db.Close()
//}
