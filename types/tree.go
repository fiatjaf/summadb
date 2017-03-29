package types

import (
	"bytes"
	"encoding/json"
	"errors"
)

type Tree struct {
	Leaf
	Branches
}

type Branches map[string]*Tree

func NewTree() *Tree { return &Tree{Branches: make(Branches)} }

func (t *Tree) UnmarshalJSON(j []byte) error {
	var v interface{}
	err := json.Unmarshal(j, &v)
	if err != nil {
		return err
	}
	parsedTree, err := TreeFromInterface(v)
	if err != nil {
		return err
	}

	*t = *parsedTree
	return nil
}

func TreeFromInterface(v interface{}) (*Tree, error) {
	t := &Tree{}

	switch val := v.(type) {
	case map[string]interface{}:
		t.Branches = make(Branches, len(val))
		for k, v := range val {
			subt, err := TreeFromInterface(v)
			if err != nil {
				return nil, err
			}
			t.Branches[k] = subt
		}
	case int:
		t.Leaf = NumberLeaf(float64(val))
	case float64:
		t.Leaf = NumberLeaf(val)
	case string:
		t.Leaf = StringLeaf(val)
	case bool:
		t.Leaf = BoolLeaf(val)
	default:
		if v == nil {
			t.Leaf = NullLeaf()
		} else {
			return nil, errors.New("type not expected.")
		}
	}
	return t, nil
}

func (t Tree) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`{`)

	if t.Leaf.Kind != UNDEFINED {
		jsonLeaf, err := t.Leaf.MarshalJSON()
		if err != nil {
			return nil, err
		}
		buffer.WriteString(`"_val":`)
		buffer.Write(jsonLeaf)

		// separate "_val" from the other keys
		if len(t.Branches) > 0 {
			buffer.WriteByte(byte(','))
		}
	}

	subtrees := make([][]byte, len(t.Branches))
	i := 0
	for k, Tree := range t.Branches {
		jsonLeaf, err := Tree.MarshalJSON()
		if err != nil {
			return nil, err
		}
		subtrees[i] = append([]byte(`"`+k+`":`), jsonLeaf...)
		i++
	}
	buffer.Write(bytes.Join(subtrees, []byte(",")))

	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

func (t Tree) Recurse(p Path, handle func(Path, Leaf)) {
	handle(p, t.Leaf)
	for k, t := range t.Branches {
		t.Recurse(append(p, k), handle)
	}
}
