package substitutions

import (
	"errors"

	"github.com/spyzhov/ajson"
)

var ErrSubstitutionFailure = errors.New("failed to resolve substitutions")

type RegistryValue interface {
	string | *ajson.Node
}

type Registry[V RegistryValue] map[string]V

func (r Registry[V]) convertStrMap() map[string]string {
	result := make(map[string]string)
	for k, v := range r {
		result[k] = RegistryValueToString(v)
	}

	return result
}

// Apply substitutions to the struct. Creates side effects.
func (r Registry[V]) Apply(input any) error {
	err := substituteStruct(input, r.convertStrMap())
	if err != nil {
		return errors.Join(err, ErrSubstitutionFailure)
	}

	return nil
}

func RegistryValueToString[V RegistryValue](value V) string {
	var name string
	if v, ok := any(value).(string); ok {
		name = v
	}

	if v, ok := any(value).(*ajson.Node); ok {
		name = v.MustString()
	}

	return name
}
