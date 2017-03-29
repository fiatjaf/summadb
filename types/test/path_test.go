package test

import (
	"github.com/fiatjaf/summadb/types"
	. "gopkg.in/check.v1"
)

func (s *TypesSuite) TestPath(c *C) {
	c.Assert(types.ParsePath("fruits/banana"), DeepEquals, types.Path{"fruits", "banana"})
	c.Assert(types.ParsePath("fruits/banana/color").RelativeTo(types.ParsePath("fruits")), DeepEquals, types.Path{"banana", "color"})
}
