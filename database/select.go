package database

import (
	"errors"

	"github.com/fiatjaf/levelup"
	"github.com/summadb/summadb/types"
)

// Select() takes a pointer to a Tree. It will fill that tree with values from the database,
// for each branch in the tree. It will not create branches. It will not care about the values
// in the given Tree, they could be anything.
func (db *SummaDB) Select(sourcepath types.Path, requestTree *types.Tree) error {
	if !sourcepath.ReadValid() {
		return errors.New("cannot read invalid path: " + sourcepath.Join())
	}

	var err error
	requestTree.Recurse(sourcepath, func(p types.Path, l types.Leaf, t types.Tree) (proceed bool) {
		relpath := p.RelativeTo(sourcepath)
		subtree := requestTree.DeepPath(relpath)

		if t.RequestLeaf {
			// leaf requested
			val, ierr := db.Get(p.Join())

			if ierr != nil {
				if ierr == levelup.NotFound {
					// not found means null, since this leaf was explictly requested
					// (making it 'undefined' would cause it to not appear in the JSON result)
					subtree.Leaf = types.NullLeaf()
					proceed = true
					return
				} else {
					// an error is an error is an error
					err = ierr
					proceed = true
					return
				}
			}

			// normal value handling
			leaf := &types.Leaf{}
			err = leaf.UnmarshalJSON([]byte(val))
			subtree.Leaf = *leaf
		}

		if t.RequestRev {
			// _rev requested
			rev, ierr := db.Get(p.Child("_rev").Join())
			if ierr != nil {
				err = ierr
			}
			subtree.Rev = rev
		}

		if t.RequestMap {
			// !map requested
			subtree.Map, err = db.Get(p.Child("!map").Join())
		}

		if t.RequestDeleted {
			// _del requested
			_, ierr := db.Get(p.Child("_del").Join())
			if ierr == levelup.NotFound {
				subtree.Deleted = true
			} else if ierr != nil {
				err = ierr
			}
		}

		if t.RequestKey {
			// _key requested
			subtree.Key = p.Last()
		}

		proceed = true
		return
	})

	return err
}
