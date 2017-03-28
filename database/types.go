package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

type tree struct {
	value    *value
	subtrees map[string]tree
}

func Tree(j string) tree {
	t := &tree{}
	t.UnmarshalJSON([]byte(j))
	return *t
}

func (t *tree) UnmarshalJSON(j []byte) error {
	var v interface{}
	err := json.Unmarshal(j, &v)
	if err != nil {
		return err
	}
	return t.fromInterface(v)
}

func (t *tree) fromInterface(v interface{}) error {
	switch val := v.(type) {
	case map[string]interface{}:
		t.subtrees = make(map[string]tree, len(val))
		for k, v := range val {
			subt := &tree{}
			err := subt.fromInterface(v)
			if err != nil {
				return err
			}
			t.subtrees[k] = *subt
		}
	case int:
	case float64:
		t.value = &value{kind: NUMBER, float64: float64(val)}
	case string:
		t.value = &value{kind: STRING, string: val}
	case bool:
		t.value = &value{kind: BOOL, bool: val}
	default:
		if v == nil {
			t.value = &value{kind: NULL}
		} else {
			return errors.New("type not expected.")
		}
	}
	return nil
}

func (t tree) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`{`)

	if t.value != nil {
		jsonvalue, err := t.value.MarshalJSON()
		if err != nil {
			return nil, err
		}
		buffer.WriteString(`"_val":`)
		buffer.Write(jsonvalue)

		// separate "_val" from the other keys
		if len(t.subtrees) > 0 {
			buffer.WriteByte(byte(','))
		}
	}

	subtrees := make([][]byte, len(t.subtrees))
	i := 0
	for k, tree := range t.subtrees {
		jsonvalue, err := tree.MarshalJSON()
		if err != nil {
			return nil, err
		}
		subtrees[i] = append([]byte(`"`+k+`":`), jsonvalue...)
		i++
	}
	buffer.Write(bytes.Join(subtrees, []byte(",")))

	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}

func (t tree) recurse(p path, handle func(path, *value)) {
	handle(p, t.value)
	for k, t := range t.subtrees {
		t.recurse(p.concat(k), handle)
	}
}

const (
	STRING = 's'
	NUMBER = 'n'
	BOOL   = 'b'
	NULL   = 'u'
)

type value struct {
	kind byte
	float64
	string
	bool
}

func (v value) MarshalJSON() ([]byte, error) {
	switch v.kind {
	case STRING:
		return []byte(`"` + v.string + `"`), nil
	case NUMBER:
		return []byte(strconv.FormatFloat(v.float64, 'f', -1, 32)), nil
	case BOOL:
		if v.bool {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	case NULL:
		return []byte("null"), nil
	}
	return []byte("undefined"), nil
}

type path []string

func NewPath(s string) path         { return path(strings.Split(s, "/")) }
func (p path) join() string         { return strings.Join(p, "/") }
func (p path) concat(k string) path { return append(p, k) }
