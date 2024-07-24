package api3

// Aliases holds alternative name.
// Acts as tautology for unknown names.
// Words are stored in pairs and are associated as synonyms.
type Aliases struct {
	dict    map[string]string
	reverse map[string]string
}

func NewAliases(source map[string]string) Aliases {
	a := Aliases{
		dict:    make(map[string]string),
		reverse: make(map[string]string),
	}

	for k, v := range source {
		a.dict[k] = v
		a.reverse[v] = k
	}

	return a
}

// Synonym provides matching word, works both ways.
func (a Aliases) Synonym(name string) string {
	alias, ok := a.dict[name]
	if ok {
		return alias
	}

	alias, ok = a.reverse[name]
	if ok {
		return alias
	}

	return name
}
