package goutils

import (
	"bytes"
	"encoding/gob"
)

// Clone uses gob to deep copy objects.
//
//nolint:ireturn
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
