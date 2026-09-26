package fileregistry

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
)

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

	_ = json.Unmarshal(fileData, &data)

	return data
}
