package handle

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/streamrail/concurrent-map"

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
	// case "application/graphql":
	default:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			json.NewEncoder(w).Encode(GraphQLResponse{
				Errors: []GraphQLError{GraphQLError{"couldn't read request body."}},
			})
			return
		}
		gql = string(body)
	}

	doc, err := parser.Parse(parser.ParseParams{
		Source:  gql,
		Options: parser.ParseOptions{true, true},
	})
	if err != nil || len(doc.Definitions) != 1 {
		message := "your graphql query must describe a 'query' operation."
		if err != nil {
			message = err.Error()
		}
		json.NewEncoder(w).Encode(GraphQLResponse{
			Errors: []GraphQLError{GraphQLError{"failed to parse gql query: " + message}},
		})
		return
	}

	// we're just ignoring Args for now -- maybe we'll find an utility for them in the future
	topfields := doc.Definitions[0].(*ast.OperationDefinition).SelectionSet.Selections
	response := cmap.New()

	godeepAsyncMultiple(topfields, startPath, &response)

	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GraphQLResponse{
		Data: response.Items(),
	})
}

func godeep(field *ast.Field, path string, base *cmap.ConcurrentMap) {
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
		base.Set(alias, db.FromLevel(val))
	} else {
		// will continue with the following selections
		next := cmap.New()
		base.Set(alias, next)
		godeepAsyncMultiple(
			field.SelectionSet.Selections,
			path+"/"+field.Name.Value,
			&next,
		)
	}
}

func godeepAsyncMultiple(selections []ast.Selection, nextPath string, next *cmap.ConcurrentMap) {
	nfields := len(selections)
	wait := make(chan int, nfields)

	for _, nextSel := range selections {
		nextField := nextSel.(*ast.Field)
		godeep(nextField, nextPath, next)
		wait <- 1
	}

	for i := 0; i < nfields; i++ {
		<-wait
	}
}

type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data,omitempty"`
	Errors []GraphQLError         `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}
