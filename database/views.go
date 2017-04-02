package database

import (
	"path"
	"strings"

	"github.com/fiatjaf/levelup"
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/fiatjaf/summadb/types"
	"github.com/fiatjaf/summadb/views"
	"github.com/mgutz/logxi/v1"
)

func (db *SummaDB) changed(p types.Path) {
	// look through ancestors for map functions
	son := p.Copy()
	for parent := son.Parent(); !parent.Equals(son); parent = son.Parent() {
		mapf, _ := db.Get(parent.Child("_map").Join())
		if mapf != "" {
			// grab document
			tree, err := db.Read(parent)
			if err != nil {
				log.Error("failed to read document to the map function",
					"err", err,
					"path", parent)
				continue
			}

			// run map function
			go func(docpath types.Path) {
				docid := docpath.Last()
				emittedrows := runMap(mapf, tree, docid)
				db.updateEmittedRowsInTheDatabase(docpath.Parent(), docid, emittedrows)
			}(parent)
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

	for key, doc := range tree.Branches {
		emittedrows := runMap(mapf, *doc, key)
		db.updateEmittedRowsInTheDatabase(p, key, emittedrows)
	}
}

func runMap(mapf string, tree types.Tree, key string) []views.EmittedRow {
	log.Debug("running map", "mapf", mapf, "key", key)
	emittedrows, err := views.Map(mapf, tree, key)
	if err != nil {
		log.Error("views.Map returned error.",
			"err", err,
			"mapf", mapf)
	}
	return emittedrows
}

func (db *SummaDB) updateEmittedRowsInTheDatabase(
	p types.Path, docid string, emittedrows []views.EmittedRow) {

	allkeys := make([]string, len(emittedrows))

	for i, row := range emittedrows {
		allkeys[i] = row.Key
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
	for _, dbrowkey := range strings.Split(prevkeys, "^!~@!") {
		err = db.deleteEmittedRow(p, dbrowkey)
		if err != nil {
			log.Error("unexpected error when deleting emitted row from the database.",
				"err", err,
				"rowkey", dbrowkey)
		}
	}

	// store keys emitted by this doc so we can delete/update them later
	if len(allkeys) > 0 {
		err = db.local.Put(localmetakey, strings.Join(allkeys, "^!~@!"))
		if err != nil {
			log.Error("unexpected error when storing list of emitted rows",
				"err", err,
				"localmetakey", localmetakey)
			return
		}

		// save all emitted rows in the database
		for _, row := range emittedrows {
			err = db.saveEmittedRow(p, row.Key, row.Value)
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

func (db *SummaDB) deleteEmittedRow(base types.Path, key string) error {
	var ops []levelup.Operation

	basepath := base.Child("@map").Child(key)
	iter := db.ReadRange(&slu.RangeOpts{
		Start: basepath.Join(),
		End:   basepath.Join() + "~~~",
	})
	defer iter.Release()

	for ; iter.Valid(); iter.Next() {
		ops = append(ops, slu.Del(path.Join()))
	}
	return db.Batch(ops)
}

func (db *SummaDB) saveEmittedRow(base types.Path, key string, value types.Tree) error {
	var ops []levelup.Operation

	basepath := base.Child("@map").Child(key)
	value.Recurse(basepath,
		func(p types.Path, leaf types.Leaf, t types.Tree) (proceed bool) {
			if leaf.Kind == types.UNDEFINED {
				proceed = true
				return
			}

			jsonvalue, _ := leaf.MarshalJSON()
			ops = append(ops, slu.Put(p.Join(), string(jsonvalue)))
			proceed = true
			return
		})
	return db.Batch(ops)
}
