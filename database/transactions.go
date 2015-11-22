package database

import (
	"fmt"
	"log"
	"strconv"

	"github.com/fiatjaf/sublevel"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type kind int

const (
	SAVE           kind = 1
	DELETE         kind = 2
	DELETECHILDREN kind = 3
	UNDELETE       kind = 4
	NOTHING        kind = 0

	RESET int = -1
)

type prepared map[string]*op
type op struct {
	kind kind
	val  []byte
	rev  string
	path string // this field is only used in RESET operations
}

func (p prepared) reset(path string) prepared {
	p["RESET"] = &op{
		kind: -1,
		path: path,
	}
	return p
}

func (p prepared) prepare(kind kind, path string, val []byte) prepared {
	p[path] = &op{
		kind: kind,
		val:  val,
	}

	/* add NOTHING ops for each parent key not already being modified
	   this will ensure they will get their revs bumped later */
	pathKeys := SplitKeys(path)
	if kind == SAVE {
		for i := 1; i < len(pathKeys); i++ {
			parentKey := JoinKeys(pathKeys[:i])
			if _, ok := p[parentKey]; !ok {
				p[parentKey] = &op{kind: UNDELETE}
			}
		}
	} else if kind == DELETE {
		for i := 1; i < len(pathKeys); i++ {
			parentKey := JoinKeys(pathKeys[:i])
			if _, ok := p[parentKey]; !ok {
				p[parentKey] = &op{kind: NOTHING}
			}
		}
	}

	return p
}

func (p prepared) commit(db *sublevel.AbstractLevel) (prepared, error) {
	batch := db.NewBatch()

	/* lastseq */
	updateSeq := getUpdateSeq(db)

	/* reseting */
	if op, ok := p["RESET"]; ok {
		deleteChildrenForPathInBatch(db, batch, op.path, &updateSeq)
		batch.Delete(DOC_STORE, []byte(op.path+"/_deleted"))
		delete(p, "RESET")
	}

	for path, op := range p {
		bytepath := []byte(path)
		if op.kind == SAVE {
			batch.Put(DOC_STORE, bytepath, op.val)
			batch.Delete(DOC_STORE, []byte(path+"/_deleted"))
			updateSeq++
			op.rev = bumpPathInBatch(db, batch, path, updateSeq)

		} else if op.kind == DELETE {
			deleteChildrenForPathInBatch(db, batch, path, &updateSeq)

			// now we operate on the "base" key
			batch.Delete(DOC_STORE, bytepath)
			batch.Put(DOC_STORE, []byte(path+"/_deleted"), []byte(nil))
			updateSeq++
			op.rev = bumpPathInBatch(db, batch, path, updateSeq)

		} else if op.kind == UNDELETE {
			batch.Delete(DOC_STORE, []byte(path+"/_deleted"))
			updateSeq++
			op.rev = bumpPathInBatch(db, batch, path, updateSeq)

		} else if op.kind == NOTHING {
			updateSeq++
			op.rev = bumpPathInBatch(db, batch, path, updateSeq)
		}

		/* there's no need to bump parent revs here, since all the parents were already
		   added to the prepared map */
	}

	// bump global update seq
	batch.Put(BY_SEQ, []byte(UPDATE_SEQ_KEY), []byte(strconv.Itoa(int(updateSeq))))

	err := db.Write(batch, nil)
	if err != nil {
		log.Print("commit failed: ", err)
		return nil, err
	}
	return p, nil
}

func deleteChildrenForPathInBatch(
	db *sublevel.AbstractLevel,
	batch *sublevel.SuperBatch,
	path string,
	updateSeq *uint64,
) error {
	/* iterate through all children (/something/here, /something/else/where etc.)
	   do not iterate through this same "base" path.
	   save the fetched paths in a bucket from which we will then bump their revs.
	*/
	toBump := make(map[string]bool)
	docs := db.Sub(DOC_STORE)
	iter := docs.NewIterator(util.BytesPrefix([]byte(path+"/")), nil)
	for iter.Next() {
		subpath := string(iter.Key())

		/* for special values, those starting with "_", we remove the last key of
		   path so we can guarantee that intermediate keys with no value also have
		   their revs bumped. we can trust this because these intermediate paths will
		   always have a _rev (and sometimes a _deleted) */
		pathKeys := SplitKeys(subpath)
		if pathKeys[len(pathKeys)-1][0] == '_' {
			toBump[JoinKeys(pathKeys[:len(pathKeys)-1])] = true
		} else {
			toBump[subpath] = true
		}
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		log.Print("delete children iteration failed: ", err)
		return err
	}

	for subpath := range toBump {
		batch.Delete(DOC_STORE, []byte(subpath))
		batch.Put(DOC_STORE, []byte(subpath+"/_deleted"), []byte(nil))
		*updateSeq++
		bumpPathInBatch(db, batch, subpath, *updateSeq)
	}

	return nil
}

func bumpPathInBatch(
	db *sublevel.AbstractLevel,
	batch *sublevel.SuperBatch,
	path string,
	newseq uint64,
) (newrev string) {
	// bumping rev
	docs := db.Sub(DOC_STORE)
	oldrev := GetRev(docs, path)
	newrev = NewRev(string(oldrev))
	batch.Put(DOC_STORE, []byte(path+"/_rev"), []byte(newrev))

	// bumping seq
	batch.Put(BY_SEQ, []byte(fmt.Sprintf("%s:%016d", path, newseq)), []byte(newrev))

	return newrev
}
