package database

import (
	"errors"

	"github.com/fiatjaf/levelup"
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/fiatjaf/summadb/types"
)

func (db *SummaDB) Delete(p types.Path, rev string) error {
	var ops []levelup.Operation

	// check if the toplevel rev matches and cancel everything if it doesn't
	toplevelrev, err := db.Get(p.Child("_rev").Join())
	if (toplevelrev != rev) && (err != levelup.NotFound && rev != "") {
		return errors.New("mismatching revs: " + toplevelrev + " and " + rev)
	}

	iter := db.ReadRange(&slu.RangeOpts{
		Start: p.Join(),
		End:   p.Join() + "~~~",
	})
	defer iter.Release()

	alreadyDeleted := make(map[string]bool)
	revsToBump := make(map[string]string)
	for ; iter.Valid(); iter.Next() {
		path := types.ParsePath(iter.Key())

		switch path.Last() {
		case "_rev":
			// mark this rev to bump, unless it was already deleted
			if _, was := alreadyDeleted[path.Parent().Join()]; !was {
				revsToBump[path.Parent().Join()] = iter.Value()
			}
		case "_deleted":
			// the path was already deleted, so we shouldn't do anything
			alreadyDeleted[path.Parent().Join()] = true
		default:
			// drop the value at this path (it doesn't matter,
			// we're deleting everything besides _rev and _deleted)
			ops = append(ops, slu.Del(path.Join()))

			if path.Leaf() {
				// mark it as deleted
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

	// finally, regardless of anything else, the source path should be bumped
	rev, _ = db.Get(p.Child("_rev").Join())
	revsToBump[p.Join()] = rev

	// bump revs
	for leafpath, oldrev := range revsToBump {
		p := types.ParsePath(leafpath)
		newrev := bumpRev(oldrev)
		ops = append(ops, slu.Put(p.Child("_rev").Join(), newrev))
	}

	return db.Batch(ops)
}
