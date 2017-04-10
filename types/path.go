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
func (p Path) Child(s string) Path {
	return append(p.Copy(), s)
}
func (p Path) Last() string {
	size := len(p)
	if size == 0 {
		return ""
	}
	return p[size-1]
}
func (p1 Path) Equals(p2 Path) bool { return p1.Join() == p2.Join() }
func (p Path) Copy() Path {
	var newpath Path
	return append(newpath, p...)
}

func (p Path) ReadValid() bool {
	switch p.Last() {
	case "_rev", "_val", "_del", "@map":
		return false
	}
	return true
}

func (p Path) WriteValid() bool {
	for i, key := range p {
		if key == "" {
			return false
		}
		if key[0] == '_' {
			return false
		}
		if key[0] == '@' && i != len(p)-1 {
			return false
		}
	}
	return true
}

func (p Path) IsLeaf() bool {
	last := p.Last()
	if len(last) > 0 && last[0] != '_' && last[0] != '@' {
		return true
	}
	return false
}
