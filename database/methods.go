package database

import (
	"github.com/syndtr/goleveldb/leveldb/util"
	"reflect"
	"strings"
)

const (
	DOC_STORE = "doc-store"
	BY_SEQ    = "by-seq"
)

func GetValueAt(path string) (interface{}, error) {
	db := Sub(DOC_STORE)
	defer db.Close()

	bs, err := db.Get([]byte(path), nil)
	if err != nil {
		return "", err
	}

	return FromLevel(bs), nil
}

func GetTreeAt(path string) (interface{}, error) {
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
			baseTree := tree // temporarily save reference to the base tree here
			for _, subkey := range strings.Split(key[1:], "/") {
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

func SaveValue(path string, bs []byte) {
	db := Sub(DOC_STORE)
	defer db.Close()

	saveRaw(db, path, bs)
}

func saveRaw(db sublevel, path string, bs []byte) {
	db.Put([]byte(path), bs, nil)
}

func SaveTree(path string, tree map[string]interface{}) {
	db := Sub(DOC_STORE)
	defer db.Close()

	saveObjectAt(db, path, tree)
}

func saveObjectAt(db sublevel, base string, o map[string]interface{}) {
	for k, v := range o {
		if v == nil || reflect.TypeOf(v).Kind() != reflect.Map {
			k = string(k)
			var path string
			if k == "_val" {
				path = base
			} else {
				path = base + "/" + k
			}

			if v == nil {
				// setting a value to null should delete it
				deleteRaw(db, path)
			} else {
				saveRaw(db, path, ToLevel(v))
			}
		} else {
			saveObjectAt(db, base+"/"+k, v.(map[string]interface{}))
		}
	}
}

func Delete(path string) {
	db := Sub(DOC_STORE)
	defer db.Close()

	deleteRaw(db, path)
}

func deleteRaw(db sublevel, path string) {
	db.Delete([]byte(path), nil)
}
