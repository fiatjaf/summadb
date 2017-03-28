package database

import (
	"github.com/fiatjaf/goleveldown"
	"github.com/fiatjaf/levelup"
	"github.com/fiatjaf/summadb/types"
)

type SummaDB struct {
	levelup.DB
}

func Open(dbpath string) *SummaDB {
	db := goleveldown.NewDatabase(dbpath)
	return &SummaDB{db}
}

func (db *SummaDB) Set(p types.Path, t types.Tree) error {
	ops := db.setOperations(p, t)
	return db.Batch(ops)
}

func (db *SummaDB) setOperations(p types.Path, t types.Tree) (ops []levelup.Operation) {
	t.Recurse(p, func(p types.Path, l types.Leaf) {
		jsonvalue, _ := l.MarshalJSON()
		ops = append(ops, levelup.Put(p.Join(), string(jsonvalue)))
	})
	return ops
}

func (db *SummaDB) Drop(p types.Path) error {
	ops := db.dropOperations(p)
	return db.Batch(ops)
}

func (db *SummaDB) dropOperations(p types.Path) (ops []levelup.Operation) {
	iter := db.ReadRange(&levelup.RangeOpts{
		Start: p.Join(),
		End:   p.Join() + "~~~",
	})
	for ; iter.Valid(); iter.Next() {
		ops = append(ops, levelup.Del(iter.Key()))
	}
	return ops
}

func (db *SummaDB) Replace(p types.Path, t types.Tree) error {
	ops := append(db.dropOperations(p), db.setOperations(p, t)...)
	return db.Batch(ops)
}

// func Get(p types.Path) (types.Tree, error) {
// 	iter := db.ReadRange(&levelup.RangeOpts{
// 		Start: p.Join(),
// 		End:   p.Join() + "~~~",
// 	})
// }
