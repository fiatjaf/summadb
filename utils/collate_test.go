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
	c.Assert(ToIndexable([]interface{}{"w", "m"}), DeepEquals, []byte{'5', '4', 'w', 0, '4', 'm', 0, 0})

	// collation only supports float64, never integers;
	// and slices and maps more specific than []interface{}, map[interface{}]interface{}
	//   are also forbidden.
	c.Assert(func() { indexify([]string{"xiii"}) }, PanicMatches, ".*does not work.*")
	c.Assert(func() { ToIndexable([]float64{37.23, 16.6, 12}) }, PanicMatches, ".*does not work.*")
	c.Assert(func() { ToIndexable(map[string]interface{}{"xi": "lascou"}) }, PanicMatches, ".*does not work.*")
}
