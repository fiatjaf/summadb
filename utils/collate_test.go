package utils

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestCollate(t *testing.T) {
	TestingT(t)
}

type CollateSuite struct{}

var _ = Suite(&CollateSuite{})

func (s *CollateSuite) TestBasic(c *C) {
	c.Assert(string(ToIndexable(float64(337))), StartsWith, "323263.3")
	c.Assert(string(ToIndexable([]interface{}{
		"bazuca",
		float64(100),
		true,
	})), Equals, "54bazuca\x00323261\x0021\x00\x00")
}
