package database

import (
	"errors"

	"github.com/summadb/summadb/types"
)

// the most simple function in the world, a helper to the _rev for a given path.
func (db *SummaDB) Rev(p types.Path) (string, error) {
	// check if the path is valid
	if !p.WriteValid() {
		return "", errors.New("cannot get rev for invalid path: " + p.Join())
	}

	return db.Get(p.Child("_rev").Join())
}
