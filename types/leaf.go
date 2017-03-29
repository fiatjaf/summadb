package types

import (
	"encoding/json"
	"errors"
	"strconv"
)

const (
	STRING    = 's'
	NUMBER    = 'n'
	BOOL      = 'b'
	NULL      = 'u'
	UNDEFINED = 0
)

type Leaf struct {
	Kind byte
	float64
	string
	bool
}

func BoolLeaf(v bool) Leaf      { return Leaf{Kind: BOOL, bool: v} }
func StringLeaf(v string) Leaf  { return Leaf{Kind: STRING, string: v} }
func NumberLeaf(v float64) Leaf { return Leaf{Kind: NUMBER, float64: v} }
func NullLeaf() Leaf            { return Leaf{Kind: NULL} }

func (l Leaf) String() string  { return l.string }
func (l Leaf) Number() float64 { return l.float64 }
func (l Leaf) Bool() bool      { return l.bool }

func (l Leaf) MarshalJSON() ([]byte, error) {
	switch l.Kind {
	case STRING:
		return []byte(`"` + l.string + `"`), nil
	case NUMBER:
		return []byte(strconv.FormatFloat(l.float64, 'f', -1, 32)), nil
	case BOOL:
		if l.bool {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	case NULL:
		return []byte("null"), nil
	}
	return nil, errors.New("unexpected type.")
}

func (l *Leaf) UnmarshalJSON(j []byte) error {
	var v interface{}
	err := json.Unmarshal(j, &v)
	if err != nil {
		return err
	}

	switch val := v.(type) {
	case string:
		*l = StringLeaf(val)
	case float64:
		*l = NumberLeaf(float64(val))
	case bool:
		*l = BoolLeaf(val)
	default:
		*l = NullLeaf()
	}

	return nil
}
