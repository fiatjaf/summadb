package database

import (
	"errors"

	"github.com/fiatjaf/levelup"
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/fiatjaf/summadb/types"
)

func (db *SummaDB) Set(p types.Path, t types.Tree) error {
	var ops []levelup.Operation

	// check if the toplevel rev matches and cancel everything if it doesn't
	toplevelrev, err := db.Get(p.Child("_rev").Join())
	if (toplevelrev != t.Rev) && (err != levelup.NotFound && t.Rev != "") {
		return errors.New("mismatching revs: " + toplevelrev + " and " + t.Rev)
	}

	iter := db.ReadRange(&slu.RangeOpts{
		Start: p.Join(),
		End:   p.Join() + "~~~",
	})
	defer iter.Release()

	// store all revs to bump in a map and bump the all at once
	revsToBump := make(map[string]string)
	for ; iter.Valid(); iter.Next() {
		path := types.ParsePath(iter.Key())

		switch path.Last() {
		case "_rev":
			revsToBump[path.Parent().Join()] = iter.Value()
		default:
			// drop the value at this path (it doesn't matter,
			// we're deleting everything besides _rev and _deleted)
			ops = append(ops, slu.Del(path.Join()))

			if path.Leaf() {
				// mark it as deleted (will unmark later if needed)
				ops = append(ops, slu.Put(path.Child("_deleted").Join(), "1"))
			}
		}
	}

	// gather all revs in ancestors, all them should be bumped
	son := p.Copy()
	for parent := son.Parent(); !parent.Equals(son); parent = son.Parent() {
		rev, _ := db.Get(parent.Child("_rev").Join())
		revsToBump[parent.Join()] = rev
		son = parent
	}

	mapfUpdated := make(map[string]types.Path)
	t.Recurse(p, func(p types.Path, leaf types.Leaf, t types.Tree) (proceed bool) {
		if t.Deleted {
			proceed = true
			return
		} else {
			// the leaf value
			jsonvalue, _ := leaf.MarshalJSON()
			ops = append(ops, slu.Put(p.Join(), string(jsonvalue)))

			// mark the rev to bump (from 0, if this path wasn't already on the database)
			if _, exists := revsToBump[p.Join()]; !exists {
				revsToBump[p.Join()] = "0-"
			}

			// undelete
			ops = append(ops, slu.Del(p.Child("_deleted").Join()))

			// save the map function if provided
			if t.Map != "" {
				ops = append(ops, slu.Put(p.Child("_map").Join(), t.Map))

				// trigger map computations for all direct children in this key
				mapfUpdated[t.Map] = p
			}

			proceed = true
			return
		}
	})

	// bump revs
	for leafpath, oldrev := range revsToBump {
		p := types.ParsePath(leafpath)
		newrev := bumpRev(oldrev)
		ops = append(ops, slu.Put(p.Child("_rev").Join(), newrev))
	}

	// write
	err = db.Batch(ops)

	if err == nil {
		for mapf, p := range mapfUpdated {
			go db.triggerChildrenMapUpdates(mapf, p)
		}
		go db.triggerAncestorMapFunctions(p)
	}

	return err
}
