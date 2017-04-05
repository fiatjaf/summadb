package database

import (
	"errors"

	"github.com/fiatjaf/levelup"
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/fiatjaf/summadb/types"
)

func (db *SummaDB) Delete(p types.Path, rev string) error {
	var ops []levelup.Operation

	// check if the path is valid for mutating
	if !p.Valid() {
		return errors.New("cannot delete invalid path: " + p.Join())
	}

	// check if the toplevel rev matches and cancel everything if it doesn't
	if err := db.checkRev(rev, p); err != nil {
		return err
	}

	// store all revs to bump in a map and bump them all at once
	revsToBump := make(map[string]string)

	alreadyDeleted := make(map[string]bool)

	iter := db.ReadRange(&slu.RangeOpts{
		Start: p.Join(),
		End:   p.Join() + "~~~",
	})
	defer iter.Release()
	for ; iter.Valid(); iter.Next() {
		path := types.ParsePath(iter.Key())

		switch path.Last() {
		case "_rev":
			// mark this rev to bump, unless it was already deleted
			if _, was := alreadyDeleted[path.Parent().Join()]; !was {
				revsToBump[path.Parent().Join()] = iter.Value()
			}
		case "_del":
			// the path was already deleted, so we shouldn't do anything
			alreadyDeleted[path.Parent().Join()] = true
		default:
			// drop the value at this path (it doesn't matter,
			// we're deleting everything besides _rev and _del)
			ops = append(ops, slu.Del(path.Join()))

			if path.Leaf() {
				// mark it as deleted
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

	// finally, regardless of anything else, the source path should be bumped
	rev, _ = db.Get(p.Child("_rev").Join())
	revsToBump[p.Join()] = rev

	// bump revs
	for leafpath, oldrev := range revsToBump {
		p := types.ParsePath(leafpath)
		newrev := bumpRev(oldrev)
		ops = append(ops, slu.Put(p.Child("_rev").Join(), newrev))
	}

	// write
	err = db.Batch(ops)

	if err == nil {
		// if a map is being deleted, trigger a mapf update
		// no value is going to be emitted, since all child rows are deleted
		// but we need to clear everything at @map/
		go db.triggerChildrenMapUpdates("", p)

		// since this is a general subtree modification
		go db.triggerAncestorMapFunctions(p)
	}

	return err
}
