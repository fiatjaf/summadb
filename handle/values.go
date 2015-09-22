package handle

import (
	//"github.com/syndtr/goleveldb/leveldb/util"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fiatjaf/summadb/context"
)

func Values(w http.ResponseWriter, r *http.Request) {
	db := context.LoadDB(r)

	w.WriteHeader(http.StatusOK)

	if r.Method == "GET" || r.Method == "" {
		//var dat map[string]string
		//db.NewIterator(util.BytesPrefix([]byte(r.URL.Path)), nil)
		data, err := db.Get([]byte(r.URL.Path), nil)
		if err != nil {
			log.Print("error on db.Get", err)
			http.Error(w, "error on db.Get", 404)
			return
		}
		log.Print("content for ", r.URL.Path, ": ", data)

		w.Write(data)
	} else if r.Method == "PUT" {
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			log.Print("couldn't read request body:", err)
			http.Error(w, "couldn't read request body.", 400)
			return
		}

		db.Put([]byte(r.URL.Path), body, nil)
	} else if r.Method == "DELETE" {
		db.Delete([]byte(r.URL.Path), nil)
	}

	context.StoreDB(r, db)
}
