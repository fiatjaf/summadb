package types

func TreeFromJSON(j string) Tree {
	t := &Tree{}
	t.UnmarshalJSON([]byte(j))
	return *t
}
