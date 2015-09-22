package handle

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
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
				tree["_val"] = val
			} else {
				baseTree := tree // temporarily save reference to the base tree here
				for _, subkey := range strings.Split(key[1:], "/") {
					var subtree map[string]interface{}
					var ok bool
					if subtree, ok = tree[subkey].(map[string]interface{}); !ok {
						subtree = make(map[string]interface{})
						tree[subkey] = subtree
					}
					subtree["_val"] = val
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

		bodyType := r.Header.Get("Content-Type")
		if bodyType == "application/json" || bodyType == "text/json" {
			// if it is JSON we must save it as its structure demands
			var input map[string]interface{}
			err := json.Unmarshal(body, &input)
			if err != nil {
				log.Print("Invalid JSON sent as JSON. ", err)
				http.Error(w, "Invalid JSON sent as JSON.", 400)
				return
			}

			saveObjectAt(db, baseKey, input)
		} else {
			// otherwise just save it as a bytes
			db.Put([]byte(baseKey), body, nil)
		}

		w.WriteHeader(http.StatusOK)
	} else if r.Method == "DELETE" {
		db.Delete([]byte(baseKey), nil)
		w.WriteHeader(http.StatusOK)
	}
}

func saveObjectAt(db *leveldb.DB, base string, o map[string]interface{}) {
	for k, v := range o {
		if v == nil || reflect.TypeOf(v).Kind() != reflect.Map {
			k = string(k)
			log.Print(k, " : ", v)
			var path []byte
			if k == "_val" {
				path = []byte(base)
			} else {
				path = []byte(base + "/" + k)
			}

			if v == nil {
				// setting a value to null should delete it
				db.Delete(path, nil)
			} else {
				// we are accepting anything as a value, so we must do
				// proper conversions.
				var bs []byte
				switch value := v.(type) {
				case string:
					bs = []byte(value)
				case float64:
					buf := new(bytes.Buffer)
					err := binary.Write(buf, binary.LittleEndian, value)
					if err != nil {
						log.Print("error encoding float to bytes. ", err)
					}
					bs = buf.Bytes()
				case []interface{}:
					var err error
					bs, err = json.Marshal(value)
					if err != nil {
						log.Print("error encoding array to json. ", err)
					}
				}
				db.Put(path, bs, nil)
			}
		} else {
			saveObjectAt(db, base+"/"+k, v.(map[string]interface{}))
		}
	}
}
