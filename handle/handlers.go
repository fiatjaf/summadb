package handle

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	db "github.com/fiatjaf/summadb/database"
	responses "github.com/fiatjaf/summadb/handle/responses"
)

func Get(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)

	/* here we handle special endpoints that, in CouchDB, refer to a database,
	   but here are treated as documents (or better, subtrees) methods */
	if ctx.wantsDatabaseInfo {
		DatabaseInfo(w, r)
		return
	} else if ctx.lastKey == "_changes" {
		Changes(w, r)
		return
	} else if ctx.lastKey == "_all_docs" {
		AllDocs(w, r)
		return
	} else if ctx.lastKey == "_revs_diff" {
		return
	} else if ctx.lastKey == "_missing_revs" {
		return
	}

	var response []byte
	var err error

	if !ctx.exists {
		res := responses.NotFound()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	if ctx.wantsTree {
		var tree map[string]interface{}
		tree, err = db.GetTreeAt(ctx.path)
		if err == nil {
			tree["_id"] = ctx.lastKey
			tree["_rev"] = ctx.actualRev

			if flag(r, "revs") {
				var revs []string
				revs, err = db.RevsAt(ctx.path)
				if err == nil {
					var start int
					start, err = strconv.Atoi(strings.Split(revs[0], "-")[0])
					revids := make([]string, len(revs))
					for i, rev := range revs {
						revids[i] = strings.Split(rev, "-")[1]
					}
					tree["_revisions"] = responses.Revisions{
						Start: start,
						Ids:   revids,
					}
				}
			}

			if flag(r, "revs_info") {
				var revs []string
				revs, err = db.RevsAt(ctx.path)
				if err == nil {
					revsInfo := make([]responses.RevInfo, len(revs))
					i := 0
					for r := len(revs) - 1; r >= 0; r-- {
						revsInfo[i] = responses.RevInfo{
							Rev:    revs[r],
							Status: "missing",
						}
						i++
					}
					revsInfo[0].Status = "available"
					tree["_revs_info"] = revsInfo
				}
			}

			response, err = json.Marshal(tree)
		}
	} else {
		if ctx.lastKey == "_rev" {
			response = []byte(ctx.actualRev)
		} else {
			response, err = db.GetValueAt(ctx.path)
		}
	}

	if err != nil {
		log.Error("unknown error: ", err)
		res := responses.UnknownError()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
}

func Post(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)
	var err error
	var rev string

	if ctx.lastKey == "_bulk_get" {
		BulkGet(w, r)
		return
	}

	if ctx.lastKey[0] == '_' {
		res := responses.BadRequest("you can't post to special keys.")
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	// POST is supposed to generate a new doc, with a new random _id:
	id := db.Random(7)
	path := ctx.path + "/" + id

	if ctx.jsonBody != nil {
		// if it is JSON we must save it as its structure demands
		rev, err = db.SaveTreeAt(path, ctx.jsonBody)
	} else {
		// otherwise it's an error
		res := responses.BadRequest("you need to send a JSON body for this request.")
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	if err != nil {
		log.Error("couldn't save value: ", err)
		res := responses.UnknownError()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	w.WriteHeader(201)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses.Success{id, true, rev})
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

	if ctx.exists && ctx.actualRev != ctx.providedRev {
		log.WithFields(log.Fields{
			"given":  ctx.providedRev,
			"actual": ctx.actualRev,
		}).Debug("rev mismatch.")
		res := responses.ConflictError()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	if ctx.lastKey == "_val" || ctx.jsonBody == nil {
		// updating the value at only this specific path, as raw bytes
		rev, err = db.SaveValueAt(ctx.path, db.ToLevel(ctx.body))
	} else if ctx.jsonBody != nil {
		// saving a JSON tree, as its structure demands, expanding to subpaths and so on
		if ctx.lastKey[0] == '_' {
			log.WithFields(log.Fields{
				"path":    ctx.path,
				"lastkey": ctx.lastKey,
			}).Debug("couldn't update special key.")
			res := responses.BadRequest("you can't update special keys")
			w.WriteHeader(res.Code)
			json.NewEncoder(w).Encode(res)
			return
		}

		rev, err = db.ReplaceTreeAt(ctx.path, ctx.jsonBody)
	} else {
		err = errors.New("value received at " + ctx.path + " is not JSON: " + string(ctx.body))
	}

	if err != nil {
		log.Debug("couldn't save value: ", err)
		res := responses.UnknownError()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	w.WriteHeader(201)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses.Success{ctx.lastKey, true, rev})
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
		res := responses.BadRequest("you can't update special keys")
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	if !ctx.exists {
		res := responses.NotFound()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	if ctx.actualRev != ctx.providedRev {
		log.WithFields(log.Fields{
			"given":  ctx.providedRev,
			"actual": ctx.actualRev,
		}).Debug("rev mismatch.")
		res := responses.ConflictError()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	// update the tree as the JSON structure demands
	rev, err = db.SaveTreeAt(ctx.path, ctx.jsonBody)

	if err != nil {
		log.Debug("couldn't save value: ", err)
		res := responses.UnknownError()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses.Success{ctx.lastKey, true, rev})
}

func Delete(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)
	var err error
	var rev string

	if !ctx.exists {
		res := responses.NotFound()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	if ctx.lastKey[0] == '_' {
		res := responses.BadRequest("you can't delete special keys")
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	if ctx.actualRev != "" && ctx.actualRev != ctx.providedRev {
		log.WithFields(log.Fields{
			"given":  ctx.providedRev,
			"actual": ctx.actualRev,
		}).Debug("rev mismatch.")
		res := responses.ConflictError()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	rev, err = db.DeleteAt(ctx.path)
	if err != nil {
		log.Error("couldn't delete key at ", ctx.path, ": ", err)
		res := responses.NotFound()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses.Success{ctx.lastKey, true, rev})
}
