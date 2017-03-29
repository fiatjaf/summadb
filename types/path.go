package types

import (
	"math"
	"strings"
)

type Path []string

func ParsePath(s string) Path { return Path(strings.Split(s, "/")) }
func (p Path) Join() string   { return strings.Join(p, "/") }

func (target Path) RelativeTo(source Path) Path {
	length := int(math.Min(float64(len(target)), float64(len(source))))
	samepartslength := length
	for i := 0; i < length; i++ {
		if source[i] != target[i] {
			samepartslength = i
			break
		}
	}

	out := Path{}
	for i := samepartslength; i < len(source); i++ {
		out = append(out, "..")
	}

	extraparts := target[samepartslength:]
	out = append(out, extraparts...)

	return out
}
