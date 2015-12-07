package database

import (
	"fmt"
	"strconv"

	log "github.com/Sirupsen/logrus"
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

	RESET kind = -1
)

type prepared map[string]*op
type op struct {
	kind       kind
	val        []byte
	currentrev string
	forcerev   bool
	rev        string
	path       string // this field is only used in RESET operations
}

func (p prepared) reset(path string) prepared {
	p["RESET"] = &op{
		kind: RESET,
		path: path,
	}
	p[path] = &op{kind: NOTHING}
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

func (p prepared) commit() (prepared, error) {
	batch := db.NewBatch()

	/* lastseq */
	updateSeq := getUpdateSeq()

	/* reseting */
	if op, ok := p["RESET"]; ok {
		deleteChildrenForPathInBatch(batch, op.path, &updateSeq, false)
		batch.Delete(DOC_STORE, []byte(op.path+"/_deleted"))
		delete(p, "RESET")
	}

	for path, op := range p {
		bytepath := []byte(path)

		if op.kind == SAVE {
			batch.Put(DOC_STORE, bytepath, op.val)
			batch.Delete(DOC_STORE, []byte(path+"/_deleted"))
			updateSeq++
			op.rev = bumpPathInBatch(batch, path, op.currentrev, op.forcerev, updateSeq)

		} else if op.kind == DELETE {
			deleteChildrenForPathInBatch(batch, path, &updateSeq, true)

			// now we operate on the "base" key
			batch.Delete(DOC_STORE, bytepath)
			batch.Put(DOC_STORE, []byte(path+"/_deleted"), []byte(nil))
			updateSeq++
			op.rev = bumpPathInBatch(batch, path, op.currentrev, op.forcerev, updateSeq)

		} else if op.kind == UNDELETE {
			batch.Delete(DOC_STORE, []byte(path+"/_deleted"))
			updateSeq++
			op.rev = bumpPathInBatch(batch, path, "", op.forcerev, updateSeq)

		} else if op.kind == NOTHING {
			updateSeq++
			op.rev = bumpPathInBatch(batch, path, op.currentrev, op.forcerev, updateSeq)
		}

		/* there's no need to bump parent revs here, since all the parents were already
		   added to the prepared map */
	}

	// bump global update seq
	batch.Put(BY_SEQ, []byte(UPDATE_SEQ_KEY), []byte(strconv.Itoa(int(updateSeq))))

	err := db.Write(batch, nil)
	if err != nil {
		log.Error("commit failed: ", err)
		return nil, err
	}
	return p, nil
}

func deleteChildrenForPathInBatch(
	batch *sublevel.SuperBatch,
	path string,
	updateSeq *uint64,
	bump bool,
) error {
	/* iterate through all children (/something/here, /something/else/where etc.)
	   do not iterate through this same "base" path.
	   save the fetched paths in a bucket from which we will then bump their revs.
	*/
	toDelete := make(map[string]bool)
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
			toDelete[JoinKeys(pathKeys[:len(pathKeys)-1])] = true
		} else {
			toDelete[subpath] = true
		}
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		log.Error("delete children iteration failed: ", err)
		return err
	}

	for subpath := range toDelete {
		batch.Delete(DOC_STORE, []byte(subpath))
		batch.Put(DOC_STORE, []byte(subpath+"/_deleted"), []byte(nil))
	}

	// we will not bump revisions or seqs unless this flag is true
	// it should be false if we are doing a RESET
	if bump {
		for subpath := range toDelete {
			*updateSeq++
			bumpPathInBatch(batch, subpath, "", false, *updateSeq)
		}
	}

	return nil
}

func bumpPathInBatch(
	batch *sublevel.SuperBatch,
	path string,
	currentrev string,
	forcerev bool,
	newseq uint64,
) (newrev string) {
	// bumping rev
	if forcerev {
		newrev = currentrev
	} else {
		if currentrev == "" {
			/* no rev was passed, we will fetch our current rev
			   and bump accordingly */
			currentrev = string(GetRev(path))
		}
		newrev = NewRev(currentrev)
	}

	batch.Put(DOC_STORE, []byte(path+"/_rev"), []byte(newrev))
	batch.Put(REV_STORE, []byte(path+"::"+newrev), []byte{})

	// bumping seq
	pathkeys := SplitKeys(path)
	nkeys := len(pathkeys)
	basekey := JoinKeys(pathkeys[:nkeys-1])
	lastkey := pathkeys[nkeys-1]
	batch.Put(BY_SEQ, []byte(fmt.Sprintf("%s::%016d", basekey, newseq)), []byte(lastkey+"::"+newrev))

	return newrev
}
