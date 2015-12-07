package database

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/syndtr/goleveldb/leveldb/util"
)

func GetValueAt(path string) ([]byte, error) {
	docs := db.Sub(DOC_STORE)

	bs, err := docs.Get([]byte(path), nil)
	if err != nil {
		return nil, err
	}

	return bs, nil
}

func GetSpecialKeysAt(basepath string) (specialKeys struct {
	Rev     string
	Deleted bool
}, err error) {
	docs := db.Sub(DOC_STORE)

	basepath = basepath + "/_"
	baseLength := len(basepath)
	iter := docs.NewIterator(util.BytesPrefix([]byte(basepath)), nil)
	for iter.Next() {
		key := string(iter.Key())[baseLength-1:]
		if key == "_rev" {
			specialKeys.Rev = string(iter.Value())
		} else if key == "_deleted" {
			specialKeys.Deleted = true
		}
	}
	err = iter.Error()
	if err != nil {
		return specialKeys, err
	}
	return specialKeys, nil
}

func GetTreeAt(basepath string) (map[string]interface{}, error) {
	/* TODO consider writing a variation of this that will fetch _revs and other special
	   things. */

	docs := db.Sub(DOC_STORE)

	// check if this tree is touched
	_, err := docs.Get([]byte(basepath+"/_rev"), nil)
	if err != nil {
		return nil, err
	}

	baseLength := len(basepath)
	bytebasepath := []byte(basepath)
	baseTree := make(map[string]interface{})

	// fetch the "base" key if exists
	val, err := docs.Get(bytebasepath, nil)
	if err == nil {
		baseTree["_val"] = FromLevel(val)
	}

	// iterate only through subkeys by adding a "/" to the prefix
	iter := docs.NewIterator(util.BytesPrefix(append(bytebasepath, '/')), nil)
	for iter.Next() {
		key := string(iter.Key())[baseLength+1:]
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
					/* this subtree doesn't exist in our response object yet,
					   create it */
					subtree = make(map[string]interface{})
				}
				tree[UnescapeKey(subkey)] = subtree
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

/*
  this just saves the _rev on REV_STORE, it doesn't actually replace the valid document
  (winning rev). it is meant for the replication process, when all missing revs are sent
  and must be kept, but still only the biggest number wins, so we only have to keep it.
*/
func AcknowledgeRevFor(path string, rev string) error {
	revs := db.Sub(REV_STORE)
	return revs.Put([]byte(path+"::"+rev), []byte{}, nil)
}

func SaveValueAt(path string, bs []byte) (newrev string, err error) {
	txn := make(prepared)
	txn.prepare(SAVE, path, bs)
	txn, err = txn.commit()
	if err != nil {
		return "", err
	}
	return txn[path].rev, nil
}

func DeleteAt(path string) (newrev string, err error) {
	txn := make(prepared)
	txn = txn.prepare(DELETE, path, nil)
	txn, err = txn.commit()
	if err != nil {
		return "", err
	}
	return txn[path].rev, nil
}

func ReplaceTreeAt(
	path string,
	tree map[string]interface{},
	forcerev bool,
) (newrev string, err error) {
	txn := make(prepared)

	// first we delete everything under this path
	txn = txn.reset(path)

	// then we proceed just like SaveTreeAt
	txn, err = saveObjectAt(txn, path, tree)
	if err != nil {
		return "", err
	}

	// forcing use of current rev as the actual rev (for new_edits)
	txn[path].forcerev = forcerev

	txn, err = txn.commit()
	if err != nil {
		return "", err
	}

	return txn[path].rev, nil
}

func SaveTreeAt(path string, tree map[string]interface{}) (newrev string, err error) {
	txn := make(prepared)
	txn, err = saveObjectAt(txn, path, tree)
	if err != nil {
		return "", err
	}
	txn, err = txn.commit()
	if err != nil {
		return "", err
	}

	return txn[path].rev, nil
}

func saveObjectAt(txn prepared, base string, o map[string]interface{}) (prepared, error) {
	var err error
	var revtoset string

	for k, v := range o {
		if len(k) > 0 && k[0] == '_' {
			if k == "_val" {
				/* actually set the value at this path */
				txn = txn.prepare(SAVE, base, ToLevel(v))
			} else if k == "_id" {
				/* if there are no other values in this map, there is _id, so we should add something */
				if _, exists := txn[base]; !exists {
					txn = txn.prepare(SAVE, base, nil)
				}
			} else if k == "_rev" {
				/* attempt to use this rev -- will fail if this path already exists in the db */
				revtoset = v.(string)
			}
			/* skip other special values */
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
			txn, err = saveObjectAt(txn, base+"/"+EscapeKey(k), sliceAsTree)
			if err != nil {
				return nil, err
			}
			continue
		}
		if v == nil || rv.Kind() != reflect.Map {
			if v == nil {
				// setting a value to null should delete it
				txn = txn.prepare(DELETE, base+"/"+EscapeKey(k), nil)
			} else {
				/* actually set */
				txn = txn.prepare(SAVE, base+"/"+EscapeKey(k), ToLevel(v))
			}
			continue
		}

		/* it's a map, so proceed to do add more things deeply into the tree */
		txn, err = saveObjectAt(txn, base+"/"+EscapeKey(k), v.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
	}

	// apply the rev we get before (or, if we didn't get, doesn't matter)
	txn[base].currentrev = revtoset

	return txn, nil
}
