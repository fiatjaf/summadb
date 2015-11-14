package handle

import (
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

	path      string
	pathKeys  []string
	nkeys     int
	lastKey   string
	wantsTree bool
	isTree    bool

	actualRev string
	deleted   bool
	exists    bool
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
		entityPath := path
		pathKeys := db.SplitKeys(path)
		nkeys := len(pathKeys)
		lastKey := pathKeys[nkeys-1]
		wantsTree := true
		if lastKey == "_val" {
			path = db.JoinKeys(pathKeys[:nkeys-1])
			lastKey = pathKeys[nkeys-1]
		}
		if lastKey[0] == '_' {
			wantsTree = false
			entityPath = db.JoinKeys(pathKeys[:nkeys-1])
		}

		isTree := false
		bodyType := r.Header.Get("Content-Type")
		if lastKey != "_val" && (bodyType == "application/json" || bodyType == "text/json") {
			isTree = true
		}

		actualRev := ""
		deleted := false
		specialKeys, err := db.GetSpecialKeysAt(entityPath)
		if err == nil {
			actualRev = specialKeys.Rev
			deleted = specialKeys.Deleted
		}

		context.Set(r, k, common{
			opts: opts,

			path:      path,
			pathKeys:  pathKeys,
			nkeys:     nkeys,
			lastKey:   lastKey,
			wantsTree: wantsTree,
			isTree:    isTree,

			actualRev: actualRev,
			deleted:   deleted,
			exists:    actualRev != "" && !deleted,
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
