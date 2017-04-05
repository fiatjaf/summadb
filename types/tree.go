package types

import (
	"bytes"
	"encoding/json"
	"strings"
)

type Tree struct {
	Leaf
	Branches
	Rev     string
	Map     string
	Deleted bool
}

type Branches map[string]*Tree

func NewTree() *Tree { return &Tree{Branches: make(Branches)} }

func (t *Tree) UnmarshalJSON(j []byte) error {
	var v interface{}
	err := json.Unmarshal(j, &v)
	if err != nil {
		return err
	}

	*t = TreeFromInterface(v)
	return nil
}

func TreeFromInterface(v interface{}) Tree {
	t := Tree{}

	switch val := v.(type) {
	case map[string]interface{}:
		if val, ok := val["_val"]; ok {
			t.Leaf = LeafFromInterface(val)
		}
		if rev, ok := val["_rev"]; ok {
			t.Rev = rev.(string)
		}
		if mapf, ok := val["@map"]; ok {
			t.Map = mapf.(string)
		}
		if deleted, ok := val["_deleted"]; ok {
			t.Deleted = deleted.(bool)
		}

		delete(val, "_id")
		delete(val, "_val")
		delete(val, "_rev")
		delete(val, "@map")
		delete(val, "_deleted")
		t.Branches = make(Branches, len(val))
		for k, v := range val {
			subt := TreeFromInterface(v)
			t.Branches[k] = &subt
		}
	default:
		t.Leaf = LeafFromInterface(v)
	}
	return t
}

func (t Tree) MarshalJSON() ([]byte, error) {
	var parts [][]byte

	// current leaf
	if t.Leaf.Kind != NULL {
		jsonLeaf, err := t.Leaf.MarshalJSON()
		if err != nil {
			return nil, err
		}
		buffer := bytes.NewBufferString(`"_val":`)
		buffer.Write(jsonLeaf)
		parts = append(parts, buffer.Bytes())
	}

	// rev
	if t.Rev != "" {
		buffer := bytes.NewBufferString(`"_rev":`)
		buffer.WriteString(`"` + t.Rev + `"`)
		parts = append(parts, buffer.Bytes())
	}

	// map
	if t.Map != "" {
		buffer := bytes.NewBufferString(`"@map":`)
		buffer.WriteString(`"` + strings.Replace(t.Map, `"`, `\"`, -1) + `"`)
		parts = append(parts, buffer.Bytes())
	}

	// deleted
	if t.Deleted {
		buffer := bytes.NewBufferString(`"_deleted":`)
		buffer.WriteString("true")
		parts = append(parts, buffer.Bytes())
	}

	// all branches
	if len(t.Branches) > 0 {
		subts := make([][]byte, len(t.Branches))
		i := 0
		for k, Tree := range t.Branches {
			jsonLeaf, err := Tree.MarshalJSON()
			if err != nil {
				return nil, err
			}
			subts[i] = append([]byte(`"`+strings.Replace(k, `"`, `\"`, -1)+`":`), jsonLeaf...)
			i++
		}
		joinedbranches := bytes.Join(subts, []byte{','})
		parts = append(parts, joinedbranches)
	}

	joined := bytes.Join(parts, []byte{','})
	out := append([]byte{'{'}, joined...)
	out = append(out, '}')
	return out, nil
}

func (t Tree) Recurse(p Path, handle func(Path, Leaf, Tree) bool) {
	proceed := handle(p, t.Leaf, t)
	if proceed {
		for key, t := range t.Branches {
			deeppath := p.Child(key)
			t.Recurse(deeppath, handle)
		}
	}
}
