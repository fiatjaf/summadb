package database

import (
	"errors"

	"github.com/fiatjaf/levelup"
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/summadb/summadb/types"
)

func (db *SummaDB) Set(p types.Path, t types.Tree) error {
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

	iter := db.ReadRange(&slu.RangeOpts{
		Start: p.Join(),
		End:   p.Join() + "~~~",
	})
	defer iter.Release()
	for ; iter.Valid(); iter.Next() {
		path := types.ParsePath(iter.Key())

		switch path.Last() {
		case "_rev":
			revsToBump[path.Parent().Join()] = iter.Value()
		default:
			// drop the value at this path (it doesn't matter,
			// we're deleting everything besides _rev and _del)
			ops = append(ops, slu.Del(path.Join()))

			if path.IsLeaf() {
				// mark it as deleted (will unmark later if needed)
				ops = append(ops, slu.Put(path.Child("_del").Join(), "1"))
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

	// store all updated map functions so we can trigger computations
	var mapfUpdated []fupdated

	t.Recurse(p, func(path types.Path, leaf types.Leaf, t types.Tree) (proceed bool) {
		if t.Deleted {
			// ignore
			proceed = true
			return
		} else {
			// the leaf value
			if leaf.Kind != types.NULL {
				jsonvalue, _ := leaf.MarshalJSON()
				ops = append(ops, slu.Put(path.Join(), string(jsonvalue)))
			}

			// mark the rev to bump (from 0, if this path wasn't already on the database)
			if _, exists := revsToBump[path.Join()]; !exists {
				revsToBump[path.Join()] = "0-"
			}

			// undelete
			ops = append(ops, slu.Del(path.Child("_del").Join()))

			// save the map function if provided
			if t.Map != "" {
				ops = append(ops, slu.Put(path.Child("!map").Join(), t.Map))

				// trigger map computations for all direct children of this key
				mapfUpdated = append(mapfUpdated, fupdated{path, t.Map})
			}

			// save the reduce function if provided
			if t.Reduce != "" {
				ops = append(ops, slu.Put(path.Child("!reduce").Join(), t.Reduce))

				// trigger map computations for all direct children of this key
				mapfUpdated = append(mapfUpdated, fupdated{path.Parent(), t.Map}) // yes, map.
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
	err := db.Batch(ops)

	if err == nil {
		go func() {
			for _, update := range mapfUpdated {
				db.triggerChildrenMapUpdates(update.code, update.path)
			}
			db.triggerAncestorMapFunctions(p)
		}()
	}

	return err
}
