package goutils

import (
	"bytes"
	"encoding/gob"
)

//nolint:ireturn
func init() {
	gob.Register(map[string]any{})
	gob.Register([]any{})
}

// Clone uses gob to deep copy objects.
func Clone[T any](input T) (T, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	if err := enc.Encode(input); err != nil {
		return input, err
	}

	var replica T
	if err := dec.Decode(&replica); err != nil {
		return input, err
	}

	return replica, nil
}
