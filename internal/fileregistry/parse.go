package fileregistry

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
)

// MustParseJSON parses fileData into type D.
// fileData may contain gzip-compressed JSON.
//
// It panics if fileData cannot be parsed as JSON.
func MustParseJSON[D any](fileData []byte) D {
	var data D

	// Assume JSON data is zipped.
	// Otherwise, fallback to reading plain JSON.
	if reader, err := gzip.NewReader(bytes.NewReader(fileData)); err == nil {
		defer reader.Close()

		if zipData, err := io.ReadAll(reader); err == nil {
			fileData = zipData
		}
	}

	if err := json.Unmarshal(fileData, &data); err != nil {
		panic(err)
	}

	return data
}
