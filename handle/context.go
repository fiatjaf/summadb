package handle

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/context"

	db "github.com/fiatjaf/summadb/database"
	responses "github.com/fiatjaf/summadb/handle/responses"
)

const k int = iota

func getContext(r *http.Request) common {
	return context.Get(r, k).(common)
}

type common struct {
	body              []byte
	jsonBody          map[string]interface{}
	path              string
	pathKeys          []string
	nkeys             int
	lastKey           string
	wantsTree         bool
	localDoc          bool
	wantsDatabaseInfo bool

	currentRev  string
	deleted     bool
	exists      bool
	providedRev string
}

func setCommonVariables(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/* when the path is in the format /nanana/nanana/_val
		   we reply with the single value for that path, otherwise
		   assume the whole tree is being requested. */
		path := r.URL.Path
		wantsTree := true
		localDoc := false
		wantsDatabaseInfo := false
		entityPath := path
		pathKeys := db.SplitKeys(path)
		nkeys := len(pathKeys)
		lastKey := pathKeys[nkeys-1]

		if nkeys > 1 && pathKeys[nkeys-2] == "_local" {
			// workaround for couchdb-like _local/blablabla docs
			// we use a different sublevel for local docs so we must
			// acknowledge that somehow.
			lastKey = db.UnescapeKey(db.JoinKeys(pathKeys[nkeys-2:]))
			localDoc = true
		} else if lastKey == "" {
			// this means the request was made with an ending slash (for example:
			// https://my.summadb.com/path/to/here/), so it wants couchdb-like information
			// for the referred sub-database, and not the document at the referred path.
			// to get information on the document at the referred path the request must be
			// made without the trailing slash (or with a special header, for the root document
			// and other scenarios in which the user does not have control over the presence
			// of the ending slash).
			wantsDatabaseInfo = true

			if path == "/" {
				path = ""
				lastKey = ""
				entityPath = ""
			}
		} else {
			if lastKey[0] == '_' {
				wantsTree = false
				entityPath = db.JoinKeys(pathKeys[:nkeys-1])
				if lastKey == "_val" {
					path = db.JoinKeys(pathKeys[:nkeys-1])
					lastKey = pathKeys[nkeys-1]
				}
			}
		}

		currentRev := ""
		deleted := false
		qrev := r.URL.Query().Get("rev")
		hrev := r.Header.Get("If-Match")
		drev := ""
		providedRev := hrev // will be "" if there's no header rev

		// fetching current rev
		if !localDoc {
			// procedure for normal documents

			specialKeys, err := db.GetSpecialKeysAt(entityPath)
			if err == nil {
				currentRev = specialKeys.Rev
				deleted = specialKeys.Deleted
			}

		} else {
			// procedure for local documents
			currentRev = db.GetLocalDocRev(path)
		}

		var jsonBody map[string]interface{}
		var body []byte
		// body parsing
		if r.Method[0] == 'P' { // PUT, PATCH, POST
			/* filter body size */
			body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
			if err != nil {
				log.Error("couldn't read request body: ", err)
				res := responses.BadRequest("request body too large")
				w.WriteHeader(res.Code)
				json.NewEncoder(w).Encode(res)
				return
			}

			if lastKey != "_val" {
				err = json.Unmarshal(body, &jsonBody)
				if err != nil {
					log.Error("invalid JSON sent as JSON: ", err, " || ", string(body))
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
				providedRev = qrev
			}
		}
		drev = ""
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

		if revfail {
			res := responses.BadRequest("different rev values were sent")
			w.WriteHeader(res.Code)
			json.NewEncoder(w).Encode(res)
			return
		}

		context.Set(r, k, common{
			body:              body,
			jsonBody:          jsonBody,
			path:              path,
			pathKeys:          pathKeys,
			nkeys:             nkeys,
			lastKey:           lastKey,
			wantsTree:         wantsTree,
			localDoc:          localDoc,
			wantsDatabaseInfo: wantsDatabaseInfo,

			currentRev:  currentRev,
			deleted:     deleted,
			exists:      currentRev != "" && !deleted,
			providedRev: providedRev,
		})

		next.ServeHTTP(w, r)
		context.Clear(r) // clears after handling everything.
	})
}
