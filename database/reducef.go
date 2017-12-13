package database

import (
	"github.com/fiatjaf/levelup"
	slu "github.com/fiatjaf/levelup/stringlevelup"
	"github.com/summadb/summadb/types"
	"github.com/summadb/summadb/views"
)

func (db *SummaDB) runReduce(
	base types.Path, /* the path of the main record that has a "!reduce" function */
	directive string, /* "add" or "remove" */
	emitted types.EmittedRow, /* the value that is being added, or the value that has already been removed */
	key string, /* the doc key of the document that caused the emitted row above to be emitted */
) error {
	reducepath := base.Child("!reduce")

	// the code for the reduce function
	reducef, _ := db.Get(reducepath.Join())
	if reducef == "" {
		return nil
	}

	// the current reduced value (what's stored until now), transaction needed here
	current, err := db.Read(base.Child("!reduce"))
	if err != nil && err != levelup.NotFound {
		log.Error("failed to fetch current reduced value.",
			"err", err,
			"base", base,
			"reducef", reducef,
			"emitted", emitted,
			"key", key)
		return err
	}

	// actually run the reduce function
	result, err := views.Reduce(reducef, directive, current, emitted, key)
	if err != nil {
		log.Error("views.Reduce returned error.",
			"err", err,
			"base", base,
			"reducef", reducef,
			"emitted", emitted,
			"key", key)
		return err
	}

	return db.updateReduceValueInTheDatabase(reducepath, current, result)
}

func (db *SummaDB) updateReduceValueInTheDatabase(reducepath types.Path, old types.Tree, new types.Tree) error {
	var ops []levelup.Operation

	old.Recurse(reducepath,
		func(p types.Path, leaf types.Leaf, t types.Tree) (proceed bool) {
			ops = append(ops, slu.Del(p.Join()))
			proceed = true
			return
		})

	new.Recurse(reducepath,
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
