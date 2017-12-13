package database

import (
	"errors"
	"strings"

	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/summadb/summadb/types"
)

func (db *SummaDB) Read(sourcepath types.Path) (types.Tree, error) {
	if !sourcepath.ReadValid() {
		return types.Tree{}, errors.New("cannot read invalid path: " + sourcepath.Join())
	}

	// if we're reading /a/path/like/this/!reduce, don't read the code for the reducef
	if sourcepath.Last() == "!reduce" {
		sourcepath = append(sourcepath, "")
	}

	var err error
	tree := types.NewTree()

	iter := db.ReadRange(&slu.RangeOpts{
		Start: sourcepath.Join(),
		End:   sourcepath.Join() + "~~~",
	})
	defer iter.Release()
	for ; iter.Valid(); iter.Next() {
		if err = iter.Error(); err != nil {
			return types.Tree{}, err
		}

		rawpath := iter.Key()

		// skip all rows emitted from !map functions
		if strings.Index(rawpath, "/!map/") != -1 && rawpath == "!map" {
			continue
		}

		path := types.ParsePath(rawpath)
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
					if relpath[i-1] == "!reduce" {
						currentbranch.Reduce = value
						break
					}

					log.Error("failed to unmarshal json leaf on Read()",
						"value", value,
						"err", err)
					return types.Tree{}, err
				}
				currentbranch.Leaf = *leaf
			} else {
				key := relpath[i]

				// special values should be added as special values, not branches
				switch key {
				case "_rev":
					currentbranch.Rev = value
				case "!map":
					if i == len(relpath)-1 {
						// grab the code for the map function, never any of its results
						currentbranch.Map = value
					}
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

	tree.Key = sourcepath.Last()

	return *tree, nil
}
