package graphql

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"

	db "github.com/fiatjaf/summadb/database"
	responses "github.com/fiatjaf/summadb/handle/responses"
)

func HandleFunc(w http.ResponseWriter, r *http.Request) {
	var gql string
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		res := responses.BadRequest("couldn't read request body.")
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}
	switch r.Header.Get("Content-Type") {
	case "application/json":
		jsonBody := struct {
			Query string `json:"query"`
		}{}
		err = json.Unmarshal(body, &jsonBody)
		if err != nil {
			res := responses.BadRequest("couldn't parse json.")
			w.WriteHeader(res.Code)
			json.NewEncoder(w).Encode(res)
			return
		}
		gql = jsonBody.Query
		break
	case "application/x-www-form-urlencoded":
		break
	case "application/graphql":
		gql = string(body)
		break
	default:
		res := responses.BadRequest("invalid content-type.")
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	doc, err := parser.Parse(parser.ParseParams{
		Source:  gql,
		Options: parser.ParseOptions{true, true},
	})
	if err != nil {
		res := responses.BadRequest("failed to parse gql query.")
		w.WriteHeader(res.Code)
		json.NewEncoder(w).Encode(res)
		return
	}

	// we're just ignoring Args for now -- maybe we'll find an utility for them in the future
	response := make(map[string]interface{})

	for _, field := range doc.Definitions[0].(*ast.OperationDefinition).SelectionSet.Selections {
		err = godeep(field.(*ast.Field), "", response)
		if err != nil {
			res := responses.UnknownError(err.Error())
			w.WriteHeader(res.Code)
			json.NewEncoder(w).Encode(res)
			return
		}
	}

	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func godeep(field *ast.Field, path string, base map[string]interface{}) (err error) {
	alias := field.Name.Value
	if field.Alias != nil {
		alias = field.Alias.Value
	}

	if field.SelectionSet == nil {
		// end of line
		log.Print("getting value at ", path+"/"+field.Name.Value)
		val, err := db.GetValueAt(path + "/" + field.Name.Value)
		log.Print(val, err)
		base[alias] = val
		return nil
	} else {
		// will continue with the following selections
		next := make(map[string]interface{})
		base[alias] = next
		for _, nextField := range field.SelectionSet.Selections {
			err = godeep(nextField.(*ast.Field), path+"/"+field.Name.Value, next)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
