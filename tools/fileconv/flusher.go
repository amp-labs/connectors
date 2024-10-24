package fileconv

import (
	"encoding/json"
	"os"
)

type Flusher struct{}

func (Flusher) ToFile(filename string, object any) error {
	data, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, os.ModePerm) //nolint:gosec
}
