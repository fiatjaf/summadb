package handle

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fiatjaf/summadb/database"
)

func Get(w http.ResponseWriter, r *http.Request) {
	var data []byte
	tree, err := database.GetTreeAt(r.URL.Path)

	if err == nil {
		data, err = json.Marshal(tree)
	}

	if err != nil {
		log.Print("error on fetching", err)
		http.Error(w, "error on fetching", 404)
		return
	}

	w.Write(data)
}

func Put(w http.ResponseWriter, r *http.Request) {
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

		database.SaveTree(r.URL.Path, input)
	} else {
		// otherwise just save it as a string
		database.SaveValue(r.URL.Path, body)
	}

	w.WriteHeader(http.StatusOK)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	database.Delete(r.URL.Path)
	w.WriteHeader(http.StatusOK)
}
