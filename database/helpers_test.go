package database

import (
	. "gopkg.in/check.v1"
)

func (s *DatabaseSuite) TestBumpRev(c *C) {
	n := bumpRev("5-iuqoe")
	c.Assert(int32(n[0]), Equals, '6')
	c.Assert(int32(n[1]), Equals, '-')
	c.Assert(n[3:], Equals, "iuqo")

	n = bumpRev("")
	c.Assert(n, HasLen, 7)
	c.Assert(int32(n[0]), Equals, '1')
	c.Assert(int32(n[1]), Equals, '-')

	n = bumpRev("0-")
	c.Assert(n, HasLen, 7)
	c.Assert(int32(n[0]), Equals, '1')
	c.Assert(int32(n[1]), Equals, '-')

	n = bumpRev("18-f")
	c.Assert(n, HasLen, 8)
	c.Assert(int32(n[0]), Equals, '1')
	c.Assert(int32(n[1]), Equals, '9')
	c.Assert(int32(n[2]), Equals, '-')
	c.Assert(int32(n[7]), Equals, 'f')

	n = bumpRev("1-ucywlskdie")
	c.Assert(int32(n[0]), Equals, '2')
	c.Assert(int32(n[1]), Equals, '-')
	c.Assert(n[3:], Equals, "ucyw")
}

func (s *DatabaseSuite) TestRevFromParents(c *C) {
	c.Assert(revFromParents("5-xabcd", "5-yabcd"), Equals, revFromParents("5-yabcd", "5-xabcd"))
	c.Assert(revFromParents("6-yhxbc", "6-uyrbc"), Equals, "9-0NWbc")
	c.Assert(revFromParents("3-yuiop", "3-yuiop"), Equals, "3-yuiop")
}
