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
	diff := pretty.Diff(params[0], params[1])
	if len(diff) > 0 {
		return false, "differences: \n" + strings.Join(diff, "\n")
	} else {
		return true, ""
	}
}

// --------------------------------------------------------------------
// JSONEquals checker.

type jsonEqualsChecker struct {
	*check.CheckerInfo
}

/* compares two JSON string, []byte or objects whatsoever,
   ignoring key order, spaces etc. */

var JSONEquals check.Checker = &jsonEqualsChecker{
	&check.CheckerInfo{
		Name:   "JSONEquals",
		Params: []string{"obtained", "expected"},
	},
}

func (checker *jsonEqualsChecker) Check(params []interface{}, names []string) (result bool, e string) {
	toObject := func(n interface{}) (res interface{}) {
		switch p := n.(type) {
		case []byte:
			json.Unmarshal(p, &res)
		case string:
			json.Unmarshal([]byte(p), &res)
		default:
			if j, err := json.Marshal(p); err != nil {
				panic(err)
			} else {
				json.Unmarshal(j, &res)
			}
		}
		return
	}

	toJSONString := func(r interface{}) string {
		b, _ := json.Marshal(r)
		return string(b)
	}

	a := toObject(params[0])
	b := toObject(params[1])

	if reflect.DeepEqual(a, b) {
		return true, ""
	} else {
		return false, toJSONString(a) + " != " + toJSONString(b)
	}
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
