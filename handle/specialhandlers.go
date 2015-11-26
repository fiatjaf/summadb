package handle

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

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

// this method only exists for compatibility with PouchDB and should not be used.
//func BulkDocs(w http.ResponseWriter, r *http.Request) {
//	ctx := getContext(r)
//
//
//}
