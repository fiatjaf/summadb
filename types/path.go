package types

import "strings"

type Path []string

func ParsePath(s string) Path       { return Path(strings.Split(s, "/")) }
func (p Path) Join() string         { return strings.Join(p, "/") }
func (p Path) Concat(k string) Path { return append(p, k) }
