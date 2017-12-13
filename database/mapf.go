package database

import (
	"strings"

	"github.com/fiatjaf/levelup"
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/summadb/summadb/types"
	"github.com/summadb/summadb/views"
)

const SEP = "^!~"

func runMap(mapf string, tree types.Tree, key string) []types.EmittedRow {
	emittedrows, err := views.Map(mapf, tree, key)
	if err != nil {
		log.Error("views.Map returned error.",
			"err", err,
			"mapf", mapf,
			"docid", key)
	}
	return emittedrows
}

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
		// in this case we do an update with new emitted rows,
		// that will clean the map results.
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

func (db *SummaDB) updateEmittedRecordsInTheDatabase(
	p types.Path,
	docid string,
	emittedrows []types.EmittedRow,
) {
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

		relpath := types.ParsePath(relativepath)
		deletedRecord, err := db.deleteEmittedRow(p, relpath)
		if err != nil {
			log.Error("unexpected error when deleting emitted row from the database.",
				"err", err,
				"relpath", relativepath)
		} else {
			// run the "remove" reducer directive
			row := types.EmittedRow{
				RelativePath: relpath,
				Value:        deletedRecord,
			}
			err := db.runReduce(p, "remove", row, docid)
			if err != nil {
				log.Error("unexpected error when running 'remove' reduce.",
					"err", err,
					"row", row)
			}
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

			// run the "add" reducer directive
			err := db.runReduce(p, "add", row, docid)
			if err != nil {
				log.Error("unexpected error when running 'add' reduce.",
					"err", err,
					"row", row)
			}
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

func (db *SummaDB) deleteEmittedRow(base types.Path, relpath types.Path) (types.Tree, error) {
	var ops []levelup.Operation

	rpath := append(base.Child("!map"), relpath...)
	record, err := db.Read(rpath)
	if err != nil {
		return record, err
	}

	record.Recurse(rpath, func(np types.Path, _ types.Leaf, _ types.Tree) bool {
		ops = append(ops, slu.Del(np.Join()))
		return true
	})
	return record, db.Batch(ops)
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
