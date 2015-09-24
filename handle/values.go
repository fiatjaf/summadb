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
	var err error

	qs := r.URL.Query()
	if qs.Get("children") == "true" {
		tree, err := database.GetTreeAt(r.URL.Path)
		if err == nil {
			data, err = json.Marshal(tree)
		}
	} else {
		data, err = database.GetValueAt(r.URL.Path)
	}

	if err != nil {
		log.Print("error on fetching", err)
		http.Error(w, "Value not here: "+r.URL.Path, 404)
		return
	}

	w.Write(data)
}

/* Should accept PUT requests with raw string bodies:
     - curl -X PUT http://db/path/to/key -d 'some value'
   and complete JSON objects, when specified with the Content-Type header:
     - curl -X PUT http://db/path -d '{"to": {"_val": "nothing here", "key": "some value"}}' -H 'content-type: application/json'

   The "_val" key is optional when setting, but can be used to set values right to the key to which they refer. It is sometimes needed, like in this example, here "path/to" had some children values to be set, but also needed a value of its own.
*/
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

		database.SaveTreeAt(r.URL.Path, input)
	} else {
		// otherwise just save it as a string
		database.SaveValueAt(r.URL.Path, database.ToLevel(body))
	}

	w.WriteHeader(http.StatusOK)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	database.DeleteAt(r.URL.Path)
	w.WriteHeader(http.StatusOK)
}
