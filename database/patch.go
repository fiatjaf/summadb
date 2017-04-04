package database

import (
	"errors"

	"github.com/fiatjaf/levelup"
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/fiatjaf/summadb/types"
)

func (db *SummaDB) Patch(patches []PatchOp) error {
	var ops []levelup.Operation

	// store all revs to bump in a map and bump them all at once
	revsToBump := make(map[string]string)

	// store all updated map functions so we can trigger computations
	mapfUpdated := make(map[string]types.Path)

	for _, patchop := range patches {
		// any of the paths don't match cancel everything
		if !patchop.Path.Valid() {
			return errors.New("cannot patch invalid path: " + patchop.Path.Join())
		}

		// if any rev doesn't match cancel everything
		if err := db.checkRev(patchop.Rev, patchop.Path); err != nil {
			return err
		}

		if patchop.Leaf.Kind == types.NULL {
			// deleting a value
			ops = append(ops, slu.Del(patchop.Path.Join()))

			// mark as deleted
			ops = append(ops, slu.Put(patchop.Path.Child("_deleted").Join(), "1"))
		} else {
			// setting a value
			// unmark as deleted
			ops = append(ops, slu.Del(patchop.Path.Child("_deleted").Join()))

			if patchop.Path.Last() == "@map" {
				mapfUpdated[patchop.Leaf.String()] = patchop.Path
				ops = append(ops, slu.Put(patchop.Path.Join(), patchop.Leaf.String()))
			} else {
				jsonvalue, _ := patchop.Leaf.MarshalJSON()
				ops = append(ops, slu.Put(patchop.Path.Join(), string(jsonvalue)))

			}
		}

		// mark revs to bump
		rev, err := db.Get(patchop.Path.Child("_rev").Join())
		if err == nil || err == levelup.NotFound {
			revsToBump[patchop.Path.Join()] = rev
		}

		// gather all revs in ancestors, all them should be bumped
		son := patchop.Path.Copy()
		for parent := son.Parent(); !parent.Equals(son); parent = son.Parent() {
			rev, _ := db.Get(parent.Child("_rev").Join())
			revsToBump[parent.Join()] = rev
			son = parent
		}
	}

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
		for _, patchop := range patches {
			go db.triggerAncestorMapFunctions(patchop.Path)
		}
	}

	return err
}

type PatchOp struct {
	Path types.Path
	Rev  string
	Leaf types.Leaf
}
