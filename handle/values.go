package handle

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/util"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/fiatjaf/summadb/context"
)

func Values(w http.ResponseWriter, r *http.Request) {
	db := context.LoadDB(r)
	defer context.StoreDB(r, db)

	baseKey := r.URL.Path

	if r.Method == "GET" || r.Method == "" {
		var data []byte
		baseLength := len(baseKey)

		tree := make(map[string]interface{})
		iter := db.NewIterator(util.BytesPrefix([]byte(baseKey)), nil)
		for iter.Next() {
			key := string(iter.Key())[baseLength:]
			val := string(iter.Value())

			if key == "" {
				tree["value"] = val
			} else {
				baseTree := tree // temporarily save reference to the base tree here
				for _, subkey := range strings.Split(key[1:], "/") {
					var subtree map[string]interface{}
					var ok bool
					if subtree, ok = tree[subkey].(map[string]interface{}); !ok {
						subtree = make(map[string]interface{})
						tree[subkey] = subtree
					}
					subtree["value"] = val
					tree = subtree
				}
				tree = baseTree // get reference to base tree back before next iteration step
			}
		}
		iter.Release()
		err := iter.Error()

		if err == nil {
			data, err = json.Marshal(tree)
		}

		if err != nil {
			log.Print("error on fetching", err)
			http.Error(w, "error on fetching", 404)
			return
		}

		w.Write(data)
	} else if r.Method == "PUT" {
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			log.Print("couldn't read request body:", err)
			http.Error(w, "couldn't read request body.", 400)
			return
		}

		db.Put([]byte(baseKey), body, nil)
		w.WriteHeader(http.StatusOK)
	} else if r.Method == "DELETE" {
		db.Delete([]byte(baseKey), nil)
		w.WriteHeader(http.StatusOK)
	}
}
