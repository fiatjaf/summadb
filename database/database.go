package database

import (
	"github.com/fiatjaf/goleveldown"
	"github.com/fiatjaf/levelup"
)

var db levelup.DB

func Start(path string) {
	db = goleveldown.NewDatabase(path)
}

func Set(p path, t tree) error {
	ops := setOperations(p, t)
	return db.Batch(ops)
}

func setOperations(p path, t tree) (ops []levelup.Operation) {
	t.recurse(p, func(p path, l leaf) {
		jsonvalue, _ := l.MarshalJSON()
		ops = append(ops, levelup.Put(p.join(), string(jsonvalue)))
	})
	return ops
}

func Drop(p path) error {
	ops := dropOperations(p)
	return db.Batch(ops)
}

func dropOperations(p path) (ops []levelup.Operation) {
	iter := db.ReadRange(&levelup.RangeOpts{
		Start: p.join(),
		End:   p.join() + "~~~",
	})
	for ; iter.Valid(); iter.Next() {
		ops = append(ops, levelup.Del(iter.Key()))
	}
	return ops
}

func Replace(p path, t tree) error {
	ops := append(dropOperations(p), setOperations(p, t)...)
	return db.Batch(ops)
}

// func Get(p path) (tree, error) {
// 	iter := db.ReadRange(&levelup.RangeOpts{
// 		Start: p.join(),
// 		End:   p.join() + "~~~",
// 	})
// }
