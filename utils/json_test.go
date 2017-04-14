package utils

import (
	. "gopkg.in/check.v1"
)

func (s *UtilsSuite) TestJSON(c *C) {
	c.Assert(string(JSONString("alalalá")), DeepEquals, `"alalalá"`)
	c.Assert(string(JSONString(`
x:	"ÿ"
    `)), DeepEquals, `"\nx:\t\"ÿ\"\n    "`)
}
