package types

import (
	"math"
	"strings"
)

type Path []string

func ParsePath(s string) Path {
	splt := strings.Split(s, "/")
	var path Path
	for _, k := range splt {
		trm := strings.TrimSpace(k)
		if trm != "" {
			path = append(path, trm)
		}
	}
	return path
}

func (p Path) Join() string { return strings.Join(p, "/") }

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

func (p Path) Parent() Path {
	size := len(p)
	if size == 0 {
		return p
	}
	return p[:size-1]
}
func (p Path) Child(s string) Path { return append(p, s) }
func (p Path) Last() string {
	size := len(p)
	if size == 0 {
		return ""
	}
	return p[size-1]
}
func (p Path) Special() bool { return strings.HasPrefix(p.Last(), "_") }
func (p Path) Leaf() bool    { return !p.Special() }
func (p Path) Mapped() bool  { return p.Parent().Last() == "@map" }

func (p1 Path) Equals(p2 Path) bool { return p1.Join() == p2.Join() }
func (p Path) Copy() Path {
	var newpath Path
	return append(newpath, p...)
}
