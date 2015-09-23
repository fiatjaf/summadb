package views

import (
	"github.com/robertkrimen/otto"
	"log"

	"github.com/fiatjaf/summadb/database"
)

func Map(viewName string, key string, val map[string]interface{}) {
	// get javascript defined function as a string
	js, err := database.GetValue("/_map/" + viewName)
	if err != nil {
		log.Print("couldn't find map function " + viewName)
		return
	}

	// run it
	vm := otto.New()
	vm.Set("emit", func(call otto.FunctionCall) otto.Value {
		key := call.Argument(0)
		value := call.Argument(1)

		// indexing happens here

		return otto.Value{}
	})
	_, err = vm.Call(string(js), map[string]interface{}, map[string]interface{}{
		"_id":  key,
		"_val": val,
	})
}
