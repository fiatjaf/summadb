package database

import (
	"strings"

	"github.com/fiatjaf/levelup"
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/summadb/summadb/types"
)

const SEP = "^!~"

func (db *SummaDB) triggerAncestorMapFunctions(p types.Path) {
	// look through ancestors for map functions
	son := p.Copy()
	for parent := son.Parent(); !parent.Equals(son); parent = son.Parent() {
		mapf, _ := db.Get(parent.Child("!map").Join())
		if mapf != "" {
			// grab document
			tree, err := db.Read(son)
			if err != nil {
				log.Error("failed to read document to the map function",
					"err", err,
					"path", parent)
				continue
			}

			// run map function
			docpath := son
			docid := docpath.Last()
			emittedrows := runMap(mapf, tree, docid)
			db.updateEmittedRecordsInTheDatabase(docpath.Parent(), docid, emittedrows)
		}

		son = parent
	}
}

func (db *SummaDB) triggerChildrenMapUpdates(mapf string, p types.Path) {
	tree, err := db.Read(p)
	if err != nil {
		log.Error("failed to fetch parent tree on triggerChildrenMapUpdates.",
			"err", err,
			"path", p)
		return
	}

	// mapf will be an empty string if it has been deleted.
	if mapf == "" {
		// in this case we do an 'update' with no emitted rows. that will clean the map results.
		for docid, _ := range tree.Branches {
			db.updateEmittedRecordsInTheDatabase(p, docid, []types.EmittedRow{})
		}
	} else {
		for docid, doc := range tree.Branches {
			emittedrows := runMap(mapf, *doc, docid)
			db.updateEmittedRecordsInTheDatabase(p, docid, emittedrows)
		}
	}
}

func (db *SummaDB) updateEmittedRecordsInTheDatabase(p types.Path, docid string, emittedrows []types.EmittedRow) {
	allrelativepaths := make([]string, len(emittedrows))

	for i, row := range emittedrows {
		allrelativepaths[i] = row.RelativePath.Join()
	}

	localmetakey := "mapped:" + p.Join() + ":" + docid

	// fetch previous emitted rows for this same map and docid
	prevkeys, err := db.local.Get(localmetakey)
	if err != nil && err != levelup.NotFound {
		log.Error("unexpected error when reading list of previous emitted rows",
			"err", err,
			"localmetakey", localmetakey)
		return
	}

	// remove all these emitted rows from the database
	for _, relativepath := range strings.Split(prevkeys, SEP) {
		if relativepath == "" {
			continue
		}

		err = db.deleteEmittedRow(p, types.ParsePath(relativepath))
		if err != nil {
			log.Error("unexpected error when deleting emitted row from the database.",
				"err", err,
				"relpath", relativepath)
		}
	}

	// store keys emitted by this doc so we can delete/update them later
	if len(emittedrows) > 0 {
		err = db.local.Put(localmetakey, strings.Join(allrelativepaths, SEP))
		if err != nil {
			log.Error("unexpected error when storing list of emitted rows",
				"err", err,
				"localmetakey", localmetakey)
			return
		}

		// save all emitted rows in the database
		for _, row := range emittedrows {
			err = db.saveEmittedRow(p, row.RelativePath, row.Value)
		}
	} else {
		err = db.local.Del(localmetakey)
		if err != nil {
			log.Error("unexpected error when deleting list of emitted rows",
				"err", err,
				"localmetakey", localmetakey)
			return
		}
	}
}

func (db *SummaDB) deleteEmittedRow(base types.Path, relpath types.Path) error {
	var ops []levelup.Operation

	rowpath := append(base.Child("!map"), relpath...)
	iter := db.ReadRange(&slu.RangeOpts{
		Start: rowpath.Join(),
		End:   rowpath.Join() + "~~~",
	})
	defer iter.Release()

	for ; iter.Valid(); iter.Next() {
		ops = append(ops, slu.Del(iter.Key()))
	}
	return db.Batch(ops)
}

func (db *SummaDB) saveEmittedRow(base types.Path, relpath types.Path, value types.Tree) error {
	var ops []levelup.Operation

	rowpath := append(base.Child("!map"), relpath...)
	value.Recurse(rowpath,
		func(p types.Path, leaf types.Leaf, t types.Tree) (proceed bool) {
			if leaf.Kind != types.NULL {
				jsonvalue, _ := leaf.MarshalJSON()
				ops = append(ops, slu.Put(p.Join(), string(jsonvalue)))
			}
			proceed = true
			return
		})
	return db.Batch(ops)
}
