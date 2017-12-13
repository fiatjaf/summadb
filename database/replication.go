package database

import (
	"strings"

	"github.com/fiatjaf/levelup"
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/summadb/summadb/types"
	"github.com/summadb/summadb/utils"
)

type Replicator struct {
	db   *SummaDB
	path types.Path
}

type PathRev struct {
	Path string
	Rev  string
}

func (p PathRev) MarshalJSON() ([]byte, error) {
	b := append([]byte{'['}, utils.JSONString(p.Path)...)
	b = append(b, ',')
	b = append(b, utils.JSONString(p.Rev)...)
	return append(b, ']'), nil
}

// AllRevs() returns a list of all paths currently stored in the database
// along with their revs.
// This list should be sent to any database which tries to replicate with
// this one.
func (r Replicator) AllRevs() (revs []PathRev, err error) {
	iter := r.db.ReadRange(&slu.RangeOpts{
		Start: r.path.Join(),
		End:   r.path.Join() + "~~~",
	})
	defer iter.Release()
	for ; iter.Valid(); iter.Next() {
		if err = iter.Error(); err != nil {
			return
		}

		rawpath := iter.Key()

		// we only want the rows ending in _rev
		if strings.Index(rawpath, "_rev") == -1 {
			continue
		}

		// the record path, without the ending _rev
		relpath := types.ParsePath(rawpath).RelativeTo(r.path)
		recordpath := relpath[:len(relpath)-1]

		// we also do not want the root _rev of the current path
		if len(recordpath) == 0 {
			continue
		}

		revs = append(revs, PathRev{recordpath.Join(), iter.Value()})
	}
	return
}

// RevsDiff() takes a list of []PathRev returned from another database
// (the other database have supposedly called AllRevs() in itself) and
// returns a list of the paths which are not matching the currently
// stored here.
//
// RevsDiff() will not return a match if the currently stored rev has a
// higher preference than the remote rev at the same path.
func (r Replicator) RevsDiff(remoteRevs []PathRev) (paths []string, err error) {
	for _, thatpathrev := range remoteRevs {
		thisrev, err := r.db.Get(thatpathrev.Path + "/_rev")
		thisrevn, _ := revNumber(thisrev)
		thatrevn, _ := revNumber(thatpathrev.Rev)

		if thisrevn < thatrevn || err == levelup.NotFound {
			// the remove rev takes precedence.
			paths = append(paths, thatpathrev.Path)
			continue
		}

		if thisrevn == thatrevn && thisrev != thatpathrev.Rev {
			// a conflict

		}

		if err != nil {
			return nil, err
		}
	}
	return
}
