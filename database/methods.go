package database

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fiatjaf/sublevel"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	DOC_STORE = "doc-store"
	BY_SEQ    = "by-seq"
)

func GetValueAt(path string) ([]byte, error) {
	db := Open().MustSub(DOC_STORE)
	defer db.Close()

	bs, err := db.Get([]byte(path), nil)
	if err != nil {
		return []byte(nil), err
	}

	return bs, nil
}

func GetTreeAt(basepath string) (map[string]interface{}, error) {
	db := Open().MustSub(DOC_STORE)
	defer db.Close()

	baseLength := len(basepath)
	bytebasepath := []byte(basepath)
	baseTree := make(map[string]interface{})

	// fetch the "base" key if exists
	val, err := db.Get(bytebasepath, nil)
	if err == nil {
		baseTree["_val"] = FromLevel(val)
	}

	// iterate only through subkeys by adding a "/" to the prefix
	iter := db.NewIterator(util.BytesPrefix(append(bytebasepath, 0x2f)), nil)
	for iter.Next() {
		key := string(iter.Key())[baseLength:]
		val := iter.Value()

		if key == "" {
			baseTree["_val"] = FromLevel(val)
		} else {
			pathKeys := SplitKeys(key)

			/* skip special values, those starting with "_" */
			if strings.HasPrefix(pathKeys[len(pathKeys)-1], "_") {
				continue
			}

			tree := baseTree
			for _, subkey := range pathKeys {
				var subtree map[string]interface{}
				var ok bool
				if subtree, ok = tree[subkey].(map[string]interface{}); !ok {
					/* this subtree doesn't exist in our response object yet, create it */
					subtree = make(map[string]interface{})
					tree[subkey] = subtree
				}
				tree = subtree // descend into that level
			}
			// no more levels to descend into, apply the value to our response object
			tree["_val"] = FromLevel(val)
		}
	}
	iter.Release()
	err = iter.Error()

	if err != nil {
		return make(map[string]interface{}), err
	}

	return baseTree, nil
}

func SaveValueAt(path string, bs []byte) (newrev string, err error) {
	db := Open().MustSub(DOC_STORE)
	defer db.Close()

	normpath := NormalizePath(path)
	txn := make(prepared)
	txn.prepare(SAVE, normpath, bs)
	txn, err = txn.commit(db)
	if err != nil {
		return "", err
	}
	return txn[path].rev, nil
}

func DeleteAt(path string) (newrev string, err error) {
	db := Open().MustSub(DOC_STORE)
	defer db.Close()

	normpath := NormalizePath(path)
	txn := make(prepared)
	txn = txn.prepare(DELETE, normpath, nil)
	txn, err = txn.commit(db)
	if err != nil {
		return "", err
	}
	return txn[path].rev, nil
}

func SaveTreeAt(path string, tree map[string]interface{}) (newrev string, err error) {
	db, err := Open().Sub(DOC_STORE)
	if err != nil {
		return "", err
	}
	defer db.Close()

	normpath := NormalizePath(path)
	txn := make(prepared)
	txn, err = saveObjectAt(db, txn, normpath, tree)
	if err != nil {
		return "", err
	}
	txn, err = txn.commit(db)
	if err != nil {
		return "", err
	}
	return txn[path].rev, nil
}

func saveObjectAt(db *sublevel.Sublevel, txn prepared, base string, o map[string]interface{}) (prepared, error) {
	var err error
	for k, v := range o {
		if k == "_val" {
			/* actually set */
			txn = txn.prepare(SAVE, base, ToLevel(v))
			continue
		}
		if k[0] == 0x5f {
			/* skip secial values, i. e., those starting with "_" */
			continue
		}

		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice {
			/* setting array as a map of numbers to values */
			sliceAsTree := make(map[string]interface{})
			for i := 0; i < rv.Len(); i++ {
				sliceAsTree[fmt.Sprintf("%d", i)] = rv.Index(i).Interface()
			}
			// we proceed as if it were a map
			txn, err = saveObjectAt(db, txn, base+"/"+k, sliceAsTree)
			if err != nil {
				return nil, err
			}
			continue
		}
		if v == nil || rv.Kind() != reflect.Map {
			if v == nil {
				// setting a value to null should delete it
				txn = txn.prepare(DELETE, base+"/"+k, nil)
			} else {
				/* actually set */
				txn = txn.prepare(SAVE, base+"/"+k, ToLevel(v))
			}
			continue
		}

		/* it's a map, so proceed to do add more things deeply into the tree */
		txn, err = saveObjectAt(db, txn, base+"/"+k, v.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
	}
	return txn, nil
}
