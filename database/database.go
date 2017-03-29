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
	defer iter.Release()

	for ; iter.Valid(); iter.Next() {
		ops = append(ops, levelup.Del(iter.Key()))
	}
	return ops
}

func (db *SummaDB) Replace(p types.Path, t types.Tree) error {
	ops := append(db.dropOperations(p), db.setOperations(p, t)...)
	return db.Batch(ops)
}

func (db *SummaDB) Read(p types.Path) (t types.Tree, err error) {
	iter := db.ReadRange(&levelup.RangeOpts{
		Start: p.Join(),
		End:   p.Join() + "~~~",
	})
	defer iter.Release()

	tree := types.NewTree()
	for ; iter.Valid(); iter.Next() {
		if err = iter.Error(); err != nil {
			return
		}

		fullpath := types.ParsePath(iter.Key())
		relpath := fullpath.RelativeTo(p)

		value := iter.Value()
		if value == "" {
			continue
		}

		leaf := &types.Leaf{}
		if err = leaf.UnmarshalJSON([]byte(value)); err != nil {
			return
		}
		//log.Print("row ", fullpath, " ", relpath, " ", iter.Key(), " ", iter.Value(), " leaf: ", *leaf)

		currentbranch := tree
		for i := 0; i <= len(relpath); i++ {
			if i == len(relpath) {
				// last key of the path
				//log.Print("   key ", i, " (last) ", value, " leaf: ", *leaf)
				currentbranch.Leaf = *leaf
			} else {
				key := relpath[i]
				// create a subbranch at this key
				subbranch, exists := currentbranch.Branches[key]
				if !exists {
					subbranch = types.NewTree()
					currentbranch.Branches[key] = subbranch
				}
				//log.Print("   key ", i, " ", key)

				currentbranch = subbranch
				// proceed to the next, deeper, branch
			}
		}
	}

	return *tree, nil
}
