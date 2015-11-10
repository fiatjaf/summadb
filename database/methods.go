package database

import (
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
			pathKeys := splitKeys(key)

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

func SaveValueAt(path string, bs []byte) error {
	db := Open().MustSub(DOC_STORE)
	defer db.Close()

	prepared := make(prepared)
	prepare(prepared, SAVE, path, bs)
	return commit(db, prepared)
}

func DeleteAt(path string) error {
	db := Open().MustSub(DOC_STORE)
	defer db.Close()

	prepared := make(prepared)
	prepare(prepared, DELETE, path, nil)
	return commit(db, prepared)
}

func SaveTreeAt(path string, tree map[string]interface{}) error {
	db, err := Open().Sub(DOC_STORE)
	if err != nil {
		return err
	}
	defer db.Close()

	prepared := make(prepared)
	saveObjectAt(db, prepared, path, tree)
	return commit(db, prepared)
}

func saveObjectAt(db *sublevel.Sublevel, prepared prepared, base string, o map[string]interface{}) error {
	for k, v := range o {
		if k[0] == 0x5f && string(k) != "_val" {
			/* skip secial values, i. e., those starting with "_", except for "_val" */
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
			if v == nil {
				// setting a value to null should delete it
				prepare(prepared, DELETE, path, nil)
			} else {
				// where we actually set each single value:
				prepare(prepared, SAVE, path, ToLevel(v))
			}
		} else {
			err := saveObjectAt(db, prepared, base+"/"+k, v.(map[string]interface{}))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
