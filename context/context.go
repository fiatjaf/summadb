package context

import (
	"github.com/gorilla/context"
	//"github.com/syndtr/goleveldb/leveldb"
	"net/http"

	//"github.com/fiatjaf/summadb/database"
)

type key int

const k key = 23

//func LoadDB(r *http.Request) *leveldb.DB {
//	if val := context.Get(r, k); val != nil {
//		return val.(*leveldb.DB)
//	}
//	return database.OpenDB()
//}
//
//func StoreDB(r *http.Request, db *leveldb.DB) {
//	context.Set(r, k, db)
//	db.Close()
//}

func ClearContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		context.Clear(r) // clears after handling everything.
	})
}
