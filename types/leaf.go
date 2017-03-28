package types

import (
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

func BoolLeaf(v bool) Leaf      { return Leaf{Kind: BOOL, bool: v} }
func StringLeaf(v string) Leaf  { return Leaf{Kind: STRING, string: v} }
func NumberLeaf(v float64) Leaf { return Leaf{Kind: NUMBER, float64: v} }
func NullLeaf() Leaf            { return Leaf{Kind: NULL} }
