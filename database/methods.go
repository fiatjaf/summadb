package database

import (
	"bytes"
	"errors"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
	"reflect"
	"strings"
)

const (
	DOC_STORE = "doc-store"
	BY_SEQ    = "by-seq"
)

func GetValueAt(path string) ([]byte, error) {
	db := Sub(DOC_STORE)
	defer db.Close()

	bs, err := db.Get([]byte(path), nil)
	if err != nil {
		return []byte(nil), err
	}

	return bs, nil
}

func getWithRev(db sublevel, path string) ([]byte, string, error) {
	val, err := db.Get([]byte(path), nil)
	if err != nil { /* key does not exist. we should be able to create it. */
		val = []byte(nil)
	}

	revb, err := db.Get([]byte(path+"/_rev"), nil)
	if err != nil {
		/* if the _rev does not exist, this is probably because the key
		 * also doesn't, but we will not even check that, just create a
		 * new _rev. */
		revb = []byte("0-0")
	}

	rev := string(revb)
	if RevIsOk(rev) {
		return val, rev, nil
	}
	return []byte(nil), "", errors.New("_rev is not in the correct format: " + rev)
}

func GetTreeAt(path string) (map[string]interface{}, error) {
	db := Sub(DOC_STORE)
	defer db.Close()

	baseLength := len(path)
	tree := make(map[string]interface{})
	iter := db.NewIterator(util.BytesPrefix([]byte(path)), nil)
	for iter.Next() {
		key := string(iter.Key())[baseLength:]
		val := iter.Value()

		if key == "" {
			tree["_val"] = FromLevel(val)
		} else {
			pathKeys := strings.Split(key[1:], "/")

			/* skip special values, those starting with "_" */
			if strings.HasPrefix(pathKeys[len(pathKeys)-1], "_") {
				continue
			}

			baseTree := tree // temporarily save reference to the base tree here
			for _, subkey := range pathKeys {
				var subtree map[string]interface{}
				var ok bool
				if subtree, ok = tree[subkey].(map[string]interface{}); !ok {
					subtree = make(map[string]interface{})
					tree[subkey] = subtree
				}
				subtree["_val"] = FromLevel(val)
				tree = subtree
			}
			tree = baseTree // get reference to base tree back before next iteration step
		}
	}
	iter.Release()
	err := iter.Error()

	if err != nil {
		return make(map[string]interface{}), err
	}

	return tree, nil
}

func SaveValueAt(path string, bs []byte) error {
	db := Sub(DOC_STORE)
	defer db.Close()

	old, rev, err := getWithRev(db, path)
	if err != nil {
		return err
	}

	saveRaw(db, path, old, rev, bs)
	return nil
}

func saveRaw(db sublevel, path string, old []byte, oldrev string, bs []byte) {
	// don't save if it is equal.
	if bytes.Equal(old, bs) {
		return
	}

	batch := db.NewBatch()
	batch.Put([]byte(path), ToLevel(bs))
	batch.Put([]byte(path+"/_rev"), []byte(NewRev(oldrev, bs)))
	err := db.Write(batch, nil)
	if err != nil {
		log.Print("saveRaw failed: ", err)
	}

	// TODO notify all upper keys of this change
}

func DeleteAt(path string) error {
	db := Sub(DOC_STORE)
	defer db.Close()

	old, rev, err := getWithRev(db, path)
	if err != nil {
		return err
	}

	deleteRaw(db, path, old, rev)
	return nil
}

func deleteRaw(db sublevel, path string, old []byte, oldrev string) {
	batch := db.NewBatch()
	batch.Delete([]byte(path))
	batch.Put([]byte(path+"/_rev"), []byte(NewRev(oldrev, nil)))
	batch.Put([]byte(path+"/_deleted"), []byte(nil))
	err := db.Write(batch, nil)
	if err != nil {
		log.Print("deleteRaw failed: ", err)
	}

	// TODO notify all upper keys of this change
}

func SaveTreeAt(path string, tree map[string]interface{}) {
	db := Sub(DOC_STORE)
	defer db.Close()

	saveObjectAt(db, path, tree)
}

func saveObjectAt(db sublevel, base string, o map[string]interface{}) {
	for k, v := range o {
		if k[0] == 0x5f /* skip special values, those starting with "_" */ {
			continue
		} else if v == nil || reflect.TypeOf(v).Kind() != reflect.Map {
			k = string(k)
			var path string
			if k == "_val" {
				path = base
			} else {
				path = base + "/" + k
			}

			// setting single values
			old, rev, err := getWithRev(db, path)
			if err != nil {
				log.Print("problem in getWithRev inside saveObjectAt: ", err)
				return
			}
			if v == nil {
				// setting a value to null should delete it
				deleteRaw(db, path, old, rev)
			} else {
				saveRaw(db, path, old, rev, ToLevel(v))
			}
		} else {
			saveObjectAt(db, base+"/"+k, v.(map[string]interface{}))
		}
	}
}
