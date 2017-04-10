package database

import (
	"errors"

	"github.com/fiatjaf/levelup"
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/summadb/summadb/types"
)

// Merge is similar to Set, the difference being that it merges
// the given tree with the current stored tree at the path, leaving
// untouched all the paths unmentioned in the given tree.
func (db *SummaDB) Merge(p types.Path, t types.Tree) error {
	var ops []levelup.Operation

	// check if the path is valid for mutating
	if !p.WriteValid() {
		return errors.New("cannot set on invalid path: " + p.Join())
	}

	// check if the toplevel rev matches and cancel everything if it doesn't
	if err := db.checkRev(t.Rev, p); err != nil {
		return err
	}

	// store all revs to bump in a map and bump them all at once
	revsToBump := make(map[string]string)

	// store all updated map functions so we can trigger computations
	var mapfUpdated []mapfupdated

	// visit all branches of the given tree
	t.Recurse(p, func(path types.Path, leaf types.Leaf, t types.Tree) (proceed bool) {
		rev, _ := db.Get(path.Child("_rev").Join())
		revsToBump[path.Join()] = rev

		mapf, _ := db.Get(path.Child("@map").Join())

		if t.Deleted {
			// delete this leaf
			ops = append(ops, slu.Del(path.Join()))
			ops = append(ops, slu.Put(path.Child("_del").Join(), "1"))

			// trigger removal of @map results
			if mapf != "" {
				mapfUpdated = append(mapfUpdated, mapfupdated{path, ""})
			}
		} else {
			// undelete
			ops = append(ops, slu.Del(path.Child("_del").Join()))

			if leaf.Kind != types.NULL {
				// modify this leaf
				jsonvalue, _ := leaf.MarshalJSON()
				ops = append(ops, slu.Put(path.Join(), string(jsonvalue)))
			}

			if mapf != t.Map {
				ops = append(ops, slu.Put(path.Child("@map").Join(), t.Map))

				// trigger map computations for all direct children of this key
				mapfUpdated = append(mapfUpdated, mapfupdated{path, t.Map})
			}
		}
		proceed = true
		return
	})

	// gather all revs in ancestors, all them should be bumped
	son := p.Copy()
	for parent := son.Parent(); !parent.Equals(son); parent = son.Parent() {
		rev, _ := db.Get(parent.Child("_rev").Join())
		revsToBump[parent.Join()] = rev
		son = parent
	}

	// bump revs
	for leafpath, oldrev := range revsToBump {
		path := types.ParsePath(leafpath)
		newrev := bumpRev(oldrev)
		ops = append(ops, slu.Put(path.Child("_rev").Join(), newrev))
	}

	// write
	err := db.Batch(ops)

	if err == nil {
		for _, update := range mapfUpdated {
			go db.triggerChildrenMapUpdates(update.mapf, update.path)
		}
		t.Recurse(p, func(p types.Path, _ types.Leaf, _ types.Tree) (proceed bool) {
			go db.triggerAncestorMapFunctions(p)
			proceed = true
			return
		})
	}

	return err
}
