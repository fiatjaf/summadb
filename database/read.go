package database

import (
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/fiatjaf/summadb/types"
)

func (db *SummaDB) Read(sourcepath types.Path) (t types.Tree, err error) {
	tree := types.NewTree()

	iter := db.ReadRange(&slu.RangeOpts{
		Start: sourcepath.Join(),
		End:   sourcepath.Join() + "~~~",
	})
	defer iter.Release()
	for ; iter.Valid(); iter.Next() {
		if err = iter.Error(); err != nil {
			return
		}

		path := types.ParsePath(iter.Key())
		relpath := path.RelativeTo(sourcepath)

		value := iter.Value()
		if value == "" {
			continue
		}

		currentbranch := tree
		for i := 0; i <= len(relpath); i++ {
			if i == len(relpath) {
				// last key of the path
				// add the leaf here
				leaf := &types.Leaf{}
				if err = leaf.UnmarshalJSON([]byte(value)); err != nil {
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

	return *tree, nil
}
