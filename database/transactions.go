package database

import (
	"log"
	"strings"

	"github.com/fiatjaf/sublevel"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type kind int

const (
	SAVE    kind = 1
	DELETE  kind = 2
	NOTHING kind = 0
)

type prepared map[string]op
type op struct {
	kind kind
	val  []byte
}

func prepare(p prepared, kind kind, path string, val []byte) {
	p[path] = op{
		kind: kind,
		val:  val,
	}

	/* add NOTHING ops for each parent key not already being modified
	   this will ensure they will get their revs bumped later */
	pathKeys := SplitKeys(path)
	for i := range pathKeys {
		parentKey := JoinKeys(pathKeys[:i])
		if _, ok := p[parentKey]; !ok {
			p[parentKey] = op{kind: NOTHING}
		}
	}
}

func commit(db *sublevel.Sublevel, prepared prepared) error {
	batch := db.NewBatch()

	for path, op := range prepared {
		val := op.val
		bytepath := []byte(path)
		if op.kind == SAVE {
			batch.Put(bytepath, val)
			bumpRevForPathInBatch(db, batch, path)

		} else if op.kind == DELETE {
			/* iterate through all children (/something/here, /something/else/where etc.)
			   do not iterate through this same "base" path.
			   save the fetched paths in a bucket from which we will then bump their revs.
			*/
			toBump := make(map[string]bool)
			iter := db.NewIterator(util.BytesPrefix(append(bytepath, 0x2f)), nil)
			for iter.Next() {
				subpath := string(iter.Key())

				/* for special values, those starting with "_", we remove the last key of
				   path so we can guarantee that intermediate keys with no value also have
				   their revs bumped. we can trust this because these intermediate paths will
				   always have a _rev (and sometimes a _deleted) */
				pathKeys := SplitKeys(subpath)
				if strings.HasPrefix(pathKeys[len(pathKeys)-1], "_") {
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
				batch.Delete([]byte(subpath))
				batch.Put([]byte(subpath+"/_deleted"), []byte(nil))
				bumpRevForPathInBatch(db, batch, subpath)
			}

			// now we operate on the "base" key
			batch.Delete(bytepath)
			batch.Put([]byte(path+"/_deleted"), []byte(nil))
			bumpRevForPathInBatch(db, batch, path)

		} else if op.kind == NOTHING {
			bumpRevForPathInBatch(db, batch, path)
		}

		/* there's no need to bump parent revs here, since all the parents were already
		   added to the prepared map */
	}

	err := db.Write(batch, nil)
	if err != nil {
		log.Print("commit failed: ", err)
		return err
	}
	return nil
}

func bumpRevForPathInBatch(db *sublevel.Sublevel, batch *sublevel.SubBatch, path string) {
	oldrev := GetRev(db, path)
	batch.Put([]byte(path+"/_rev"), []byte(NewRev(string(oldrev))))
}
