package handle

import (
	"encoding/json"
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"

	db "github.com/fiatjaf/summadb/database"
	responses "github.com/fiatjaf/summadb/handle/responses"
	settings "github.com/fiatjaf/summadb/settings"
)

func DatabaseInfo(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)

	lastSeq, err := db.LastSeqAt(ctx.path)
	if err != nil {
		log.Print("responses.Unknown error: ", err)
		res := responses.UnknownError()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := responses.DatabaseInfo{
		DBName:            ctx.path,
		UpdateSeq:         lastSeq,
		InstanceStartTime: settings.STARTTIME.Unix(),
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}

func Changes(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)

	/* options */
	// always true temporarily: all_docs := flag(r, "style", "all_docs")
	sincep := param(r, "since")
	var since uint64
	if sincep == "now" {
		since = db.GlobalUpdateSeq()
	} else {
		var err error
		since, err = strconv.ParseUint(sincep, 10, 64)
		if err != nil {
			since = 0
		}
	}

	path := db.CleanPath(ctx.path)
	changes, err := db.ListChangesAt(path, since)
	if err != nil {
		log.Print("responses.Unknown error: ", err)
		res := responses.UnknownError()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	var lastSeq uint64 = 0
	nchanges := len(changes)
	if nchanges > 0 {
		lastSeq = changes[nchanges-1].Seq
	}

	res := responses.Changes{
		LastSeq: lastSeq,
		Results: changes,
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}

/*
   Currently _all_docs does not guarantee key order and should not be used for
   querying or anything. It is here just to provide PouchDB replication support.
*/
func AllDocs(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)

	/* options */
	include_docs := flag(r, "include_docs")
	// startkey := param(r, "startkey")
	// endkey := param(r, "endkey")
	// descending := flag(r, "descending")
	// skip := param(r, "skip")
	// limit := param(r, "limit")

	path := db.CleanPath(ctx.path)
	tree, err := db.GetTreeAt(path)
	if err != nil {
		log.Print("unknown error: ", err)
		res := responses.UnknownError()
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := responses.AllDocs{
		TotalRows: 0,
		Offset:    0, // temporary
		Rows:      make([]responses.Row, 0),
	}

	for id, doc := range tree {
		if id[0] == '_' {
			continue
		}

		specialKeys, err := db.GetSpecialKeysAt(path + "/" + id)
		if err != nil {
			log.Print("unknown error: ", err)
			res := responses.UnknownError()
			w.WriteHeader(res.Code)
			json.NewEncoder(w).Encode(res)
			return
		}

		row := responses.Row{
			Id:    id,
			Key:   id,
			Value: map[string]interface{}{"rev": specialKeys.Rev},
		}
		if include_docs {
			row.Doc = doc.(map[string]interface{})
			row.Doc["_id"] = id
			row.Doc["_rev"] = specialKeys.Rev
		}
		res.Rows = append(res.Rows, row)
		res.TotalRows += 1
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}

// this method only exists for compatibility with PouchDB and should not be used elsewhere.
func BulkGet(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)
	var ireqs interface{}

	/* options */
	revs := flag(r, "revs")

	var ok bool
	if ireqs, ok = ctx.jsonBody["docs"]; !ok {
		res := responses.BadRequest("You were supposed to request some docs specified by their ids, and you didn't.")
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	reqs := ireqs.([]interface{})
	res := responses.BulkGet{
		Results: make([]responses.BulkGetResult, len(reqs)),
	}
	for i, ireq := range reqs {
		req := ireq.(map[string]interface{})
		res.Results[i] = responses.BulkGetResult{
			Docs: make([]responses.DocOrError, 1),
		}

		iid, ok := req["id"]
		if !ok {
			err := responses.BadRequest("missing id")
			res.Results[i].Docs[0].Error = &err
			continue
		}
		id := iid.(string)
		res.Results[i].Id = id

		path := db.CleanPath(ctx.path) + "/" + id
		doc, err1 := db.GetTreeAt(path)
		specialKeys, err2 := db.GetSpecialKeysAt(path)
		if err1 != nil || err2 != nil {
			err := responses.NotFound()
			res.Results[i].Docs[0].Error = &err
			continue
		}

		doc["_id"] = id
		doc["_rev"] = specialKeys.Rev

		if revs {
			// magic method to fetch _revisions
			// docs["_revisions"] = ...
		}

		res.Results[i].Docs[0].Ok = &doc
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}

// this method only exists for compatibility with PouchDB and should not be used elsewhere.
func BulkDocs(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)
	var idocs interface{}
	var ok bool

	if idocs, ok = ctx.jsonBody["docs"]; !ok {
		res := responses.BadRequest("You're supposed to send an array of docs to input on the database, and you didn't.")
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	/* options */
	// inewedits, _ := ctx.jsonBody["new_edits"]
	// newedits := inewedits.(bool) // for now this is always true.

	path := db.CleanPath(ctx.path)
	docs := idocs.([]interface{})
	res := make([]responses.BulkDocsResult, len(docs))
	for i, idoc := range docs {
		doc := idoc.(map[string]interface{})
		var iid interface{}
		var id string
		var irev interface{}
		var rev string
		if iid, ok = doc["_id"]; ok {
			id = iid.(string)
			if irev, ok = doc["_rev"]; ok {
				rev = irev.(string)
			}
		} else {
			id = db.Random(5)
		}
		delete(doc, "_rev")
		delete(doc, "_id")

		// check rev matching:
		if iid != nil /* when iid is nil that means the doc had no _id, so we don't have to check. */ {
			actualRev, err := db.GetValueAt(path + "/" + id + "/_rev")
			/* err!=nil means there's no _rev, so ok */
			if err == nil && string(actualRev) != rev {
				e := responses.ConflictError()
				res[i] = responses.BulkDocsResult{
					Id:     id,
					Error:  e.Error,
					Reason: e.Reason,
				}
				continue
			}
		}

		// proceed to write.
		newrev, err := db.ReplaceTreeAt(path+"/"+db.EscapeKey(id), doc)
		if err != nil {
			e := responses.UnknownError()
			res[i] = responses.BulkDocsResult{
				Id:     id,
				Error:  e.Error,
				Reason: e.Reason,
			}
			continue
		}
		res[i] = responses.BulkDocsResult{
			Id:  id,
			Ok:  true,
			Rev: newrev,
		}
	}
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(res)
}

// this method only exists for compatibility with PouchDB and should not be used elsewhere.
func RevsDiff(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)

	path := db.CleanPath(ctx.path)

	res := make(map[string]responses.RevsDiffResult)
	for id, irevs := range ctx.jsonBody {
		missing := make([]string, 0)

		currentRevb, err := db.GetValueAt(path + "/" + id + "/_rev")
		if err != nil {
			/* no _rev for this id, means it has never been inserted in this database.
			   let's say we miss all the revs. */
			for _, irev := range irevs.([]interface{}) {
				rev := irev.(string)
				missing = append(missing, rev)
			}
		} else {
			/* otherwise we will say we have the current rev and none of the others,
			   because that's the truth. */
			currentRev := string(currentRevb)
			for _, irev := range irevs.([]interface{}) {
				rev := irev.(string)
				if rev != currentRev {
					missing = append(missing, rev)
				}
			}
		}

		res[id] = responses.RevsDiffResult{Missing: missing}
	}

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(res)
}
