package utils

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/kr/pretty"
	check "gopkg.in/check.v1"
)

// --------------------------------------------------------------------
// DeeplyEquals checker.

type deeplyEqualsChecker struct {
	*check.CheckerInfo
}

/* the DeeplyEquals checker differ from DeepEquals[1] because while
   that one just uses reflect.DeepEqual, which doesn't support
   resolving deeply nested pointers, this one uses github.com/kr/pretty,
   a master in the task of pointer resolve.

   [1]: https://github.com/go-check/check/blob/v1/checkers.go#L181-L204
*/

var DeeplyEquals check.Checker = &deeplyEqualsChecker{
	&check.CheckerInfo{
		Name:   "DeeplyEquals",
		Params: []string{"obtained", "expected"},
	},
}

func (checker *deeplyEqualsChecker) Check(params []interface{}, names []string) (result bool, e string) {
	return len(pretty.Diff(params[0], params[1])) == 0, ""
}

// --------------------------------------------------------------------
// JSONEquals checker.

type jsonEqualsChecker struct {
	*check.CheckerInfo
}

/* compares two JSON strings, ignoring key order, spaces etc. */

var JSONEquals check.Checker = &jsonEqualsChecker{
	&check.CheckerInfo{
		Name:   "JSONEquals",
		Params: []string{"obtained", "expected"},
	},
}

func (checker *jsonEqualsChecker) Check(params []interface{}, names []string) (result bool, e string) {
	var a interface{}
	var p1 []byte
	switch p := params[0].(type) {
	case []byte:
		p1 = p
	case string:
		p1 = []byte(p)
	}
	json.Unmarshal(p1, &a)

	var b interface{}
	var p2 []byte
	switch p := params[0].(type) {
	case []byte:
		p2 = p
	case string:
		p2 = []byte(p)
	}
	json.Unmarshal(p2, &b)

	return reflect.DeepEqual(a, b), ""
}

// --------------------------------------------------------------------
// Startswith checker.

type startsWithChecker struct {
	*check.CheckerInfo
}

/* checks if the first argument has the second as its prefix. */

var StartsWith check.Checker = &startsWithChecker{
	&check.CheckerInfo{
		Name:   "StartsWith",
		Params: []string{"string", "prefix"},
	},
}

func (checker *startsWithChecker) Check(params []interface{}, names []string) (result bool, e string) {
	s := params[0].(string)
	prf := params[1].(string)

	return strings.HasPrefix(s, prf), ""
}
