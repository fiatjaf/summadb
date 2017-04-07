package database

import (
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/fiatjaf/summadb/types"
)

type RowsParams struct {
	KeyStart   string
	KeyEnd     string
	Descending bool
	Limit      int
}

// Rows provide a querying interface similar to CouchDB, in which you can manually specify
// key start and end, starting at a certain "path level".
// in contrast with Read, which returns a big tree of everything under the given path,
// Rows return an array of trees, as the children of the given path.
func (db *SummaDB) Rows(sourcepath types.Path, params RowsParams) (rows []*types.Tree, err error) {
	rangeopts := slu.RangeOpts{
		Start:   sourcepath.Join() + "/",
		End:     sourcepath.Join() + "/~~~",
		Reverse: params.Descending,
	}
	if params.KeyStart != "" {
		rangeopts.Start = sourcepath.Join() + "/" + params.KeyStart
	}
	if params.KeyEnd != "" {
		rangeopts.End = sourcepath.Join() + "/" + params.KeyEnd
	}
	if params.Limit != 0 {
		rangeopts.Limit = params.Limit
	}

	iter := db.ReadRange(&rangeopts)
	defer iter.Release()
	for ; iter.Valid(); iter.Next() {
		if err = iter.Error(); err != nil {
			return
		}

		path := types.ParsePath(iter.Key())
		relpath := path.RelativeTo(sourcepath)
		key := relpath[0]

		// fetch the tree we're currently filling or start a new tree
		var tree *types.Tree
		if len(rows) == 0 || rows[len(rows)-1].Key != key {
			tree = types.NewTree()
			tree.Key = key
			rows = append(rows, tree)
		} else {
			tree = rows[len(rows)-1]
		}

		value := iter.Value()
		if value == "" {
			continue
		}

		// descend into tree filling in the values read from the database
		currentbranch := tree
		for /* start at 1 because 0 is already the _key */ i := 1; i <= len(relpath); i++ {
			if i == len(relpath) {
				// last key of the path
				// add the leaf here
				leaf := &types.Leaf{}
				if err = leaf.UnmarshalJSON([]byte(value)); err != nil {
					log.Error("failed to unmarshal json on Rows()",
						"value", value,
						"err", err)
					return
				}
				currentbranch.Leaf = *leaf
			} else {
				key := relpath[i]

				// special values should be added as special values, not branches
				switch key {
				case "_rev":
					currentbranch.Rev = value
				case "@map":
					currentbranch.Map = value
				case "_del":
					currentbranch.Deleted = true
				default:
					// create a subbranch at this key
					subbranch, exists := currentbranch.Branches[key]
					if !exists {
						subbranch = types.NewTree()
						currentbranch.Branches[key] = subbranch
					}

					// proceed to the next, deeper, branch
					currentbranch = subbranch
					continue
				}
				break // will break if it is a special key, continue if not
			}
		}
	}
	return
}
