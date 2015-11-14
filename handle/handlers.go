package handle

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	db "github.com/fiatjaf/summadb/database"
)

func Get(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)
	var response []byte
	var err error

	if !ctx.exists {
		res := notFound()
		w.WriteHeader(res.code)
		json.NewEncoder(w).Encode(res)
		return
	}

	if ctx.wantsTree {
		var tree map[string]interface{}
		tree, err = db.GetTreeAt(ctx.path)
		tree["_id"] = ctx.lastKey
		tree["_rev"] = ctx.actualRev
		if err == nil {
			response, err = json.Marshal(tree)
		}
	} else {
		if ctx.lastKey == "_rev" {
			response = []byte(ctx.actualRev)
		} else {
			response, err = db.GetValueAt(ctx.path)
		}
	}

	/* special cases defined by query parameters */
	if ctx.opts.revs {

	}

	if err != nil {
		log.Print("unknown error: ", err)
		res := unknownError()
		w.WriteHeader(res.code)
		json.NewEncoder(w).Encode(res)
		return
	}

	w.WriteHeader(200)
	w.Write(response)
}

/* Should accept PUT requests with raw string bodies:
     - curl -X PUT http://db/path/to/key -d 'some value'
   and complete JSON objects, when specified with the Content-Type header:
     - curl -X PUT http://db/path -d '{"to": {"_val": "nothing here", "key": "some value"}}' -H 'content-type: application/json'

   The "_val" key is optional when setting, but can be used to set values right to the key to which they refer. It is sometimes needed, like in this example, here "path/to" had some children values to be set, but also needed a value of its own.
*/
func Put(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)
	var err error
	var rev string

	if ctx.lastKey[0] == '_' {
		res := badRequest("you can't update special keys")
		w.WriteHeader(res.code)
		json.NewEncoder(w).Encode(res)
		return
	}

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Print("couldn't read request body: ", err)
		http.Error(w, "couldn't read request body.", 400)
		return
	}

	if ctx.isTree {
		// if it is JSON we must save it as its structure demands
		var input map[string]interface{}
		err = json.Unmarshal(body, &input)
		if err != nil {
			log.Print("Invalid JSON sent as JSON. ", err)
			http.Error(w, "Invalid JSON sent as JSON.", 400)
			return
		}

		rev, err = db.SaveTreeAt(r.URL.Path, input)
	} else {
		// otherwise just save it as a string
		rev, err = db.SaveValueAt(r.URL.Path, db.ToLevel(body))
	}

	if err != nil {
		log.Print("couldn't save value: ", err)
		res := unknownError()
		w.WriteHeader(res.code)
		json.NewEncoder(w).Encode(res)
		return
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(Success{ctx.lastKey, true, rev})
}

func Delete(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)
	var err error
	var rev string

	if !ctx.exists {
		res := notFound()
		w.WriteHeader(res.code)
		json.NewEncoder(w).Encode(res)
		return
	}

	if ctx.lastKey[0] == '_' {
		res := badRequest("you can't delete special keys")
		w.WriteHeader(res.code)
		json.NewEncoder(w).Encode(res)
		return
	}

	rev, err = db.DeleteAt(ctx.path)
	if err != nil {
		log.Print("couldn't delete key at ", ctx.path, ": ", err)
		res := notFound()
		w.WriteHeader(res.code)
		json.NewEncoder(w).Encode(res)
		return
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(Success{ctx.lastKey, true, rev})
}
