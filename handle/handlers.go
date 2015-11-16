package handle

import (
	"encoding/json"
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

/* requests with raw string bodies:
   - curl -X PUT http://db/path/to/key/_val -d 'some value'
 and complete JSON objects
   - curl -X PUT http://db/path -d '{"to": {"_val": "nothing here", "key": "some value"}}' -H 'content-type: application/json'

While setting the raw string body of a path will only update that path and do not change others, a full JSON request will replace all keys under the specified path. PUT is idempotent.

The "_val" key is optional when setting, but can be used to set values right to the key to which they refer. It is sometimes needed, like in this example, here "path/to" had some children values to be set, but also needed a value of its own.
Other use of the "_val" key is to set a value to null strictly, because setting the nude key to null will delete it instead of setting it.
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

	if ctx.exists && ctx.actualRev != ctx.providedRev {
		res := conflictError()
		w.WriteHeader(res.code)
		json.NewEncoder(w).Encode(res)
		return
	}

	if ctx.jsonBody != nil {
		// if it is JSON we must save it as its structure demands
		rev, err = db.ReplaceTreeAt(ctx.path, ctx.jsonBody)
	} else {
		// otherwise just save it as a string
		rev, err = db.SaveValueAt(ctx.path, db.ToLevel(ctx.body))
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

/* Should accept PATCH requests with JSON objects:
   - curl -X PATCH http://db/path -d '{"to": {"key": "some value"}}' -H 'content-type: application/json'
There will not replace all values under /path, but only modify the values which the JSON object refers to.
*/
func Patch(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)
	var err error
	var rev string

	if ctx.lastKey[0] == '_' {
		res := badRequest("you can't update special keys")
		w.WriteHeader(res.code)
		json.NewEncoder(w).Encode(res)
		return
	}

	if !ctx.exists {
		res := notFound()
		w.WriteHeader(res.code)
		json.NewEncoder(w).Encode(res)
		return
	}

	if ctx.actualRev != ctx.providedRev {
		res := conflictError()
		w.WriteHeader(res.code)
		json.NewEncoder(w).Encode(res)
		return
	}

	// update the tree as the JSON structure demands
	rev, err = db.SaveTreeAt(ctx.path, ctx.jsonBody)

	if err != nil {
		log.Print("couldn't save value: ", err)
		res := unknownError()
		w.WriteHeader(res.code)
		json.NewEncoder(w).Encode(res)
		return
	}

	w.WriteHeader(200)
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

	if ctx.actualRev != "" && ctx.actualRev != ctx.providedRev {
		res := conflictError()
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
