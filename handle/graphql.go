package handle

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"

	db "github.com/fiatjaf/summadb/database"
)

func HandleGraphQL(w http.ResponseWriter, r *http.Request) {
	ctx := getContext(r)

	startPath := r.URL.Path[:len(r.URL.Path)-9]
	allow := db.ReadAllowedAt(startPath, ctx.user)
	if !allow {
		json.NewEncoder(w).Encode(GraphQLResponse{
			Errors: []GraphQLError{GraphQLError{"_read permission for this path needed."}},
		})
		return
	}

	var gql string
	switch r.Header.Get("Content-Type") {
	case "application/json":
		jsonBody := struct {
			Query string `json:"query"`
		}{}
		err := json.NewDecoder(r.Body).Decode(&jsonBody)
		if err != nil {
			json.NewEncoder(w).Encode(GraphQLResponse{
				Errors: []GraphQLError{GraphQLError{"failed to parse json: " + err.Error()}},
			})
			return
		}
		gql = jsonBody.Query
		break
	case "application/x-www-form-urlencoded":
		r.ParseForm()
		gql = r.FormValue("query")
		break
	case "application/graphql":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			json.NewEncoder(w).Encode(GraphQLResponse{
				Errors: []GraphQLError{GraphQLError{"couldn't read request body."}},
			})
		}
		gql = string(body)
		break
	default:
		json.NewEncoder(w).Encode(GraphQLResponse{
			Errors: []GraphQLError{GraphQLError{"invalid content-type"}},
		})
		return
	}

	doc, err := parser.Parse(parser.ParseParams{
		Source:  gql,
		Options: parser.ParseOptions{true, true},
	})
	if err != nil {
		json.NewEncoder(w).Encode(GraphQLResponse{
			Errors: []GraphQLError{GraphQLError{"failed to parse gql query."}},
		})
		return
	}

	// we're just ignoring Args for now -- maybe we'll find an utility for them in the future
	var response = make(map[string]interface{})
	var errors []GraphQLError

	for _, field := range doc.Definitions[0].(*ast.OperationDefinition).SelectionSet.Selections {
		err = godeep(field.(*ast.Field), startPath, response)
		if err != nil {
			errors = append(errors, GraphQLError{err.Error()})
		}
	}

	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GraphQLResponse{
		Data:   response,
		Errors: errors,
	})
}

func godeep(field *ast.Field, path string, base map[string]interface{}) (err error) {
	alias := field.Name.Value
	if field.Alias != nil {
		alias = field.Alias.Value
	}

	if field.SelectionSet == nil {
		// end of line
		var val []byte
		if field.Name.Value == "_val" {
			val, _ = db.GetValueAt(path)
		} else {
			val, _ = db.GetValueAt(path + "/" + field.Name.Value)
		}
		base[alias] = db.FromLevel(val)
		return nil
	} else {
		// will continue with the following selections
		next := make(map[string]interface{})
		base[alias] = next
		for _, nextSel := range field.SelectionSet.Selections {
			nextField := nextSel.(*ast.Field)
			err = godeep(nextField, path+"/"+field.Name.Value, next)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data"`
	Errors []GraphQLError         `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}
