package api

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// File reads OpenAPI v3 file data.
type File struct {
	file []byte
}

func NewFile(data []byte) *File {
	return &File{
		file: data,
	}
}

func (m File) Extractor() (*Extractor, error) {
	loader := openapi3.NewLoader()

	data, err := loader.LoadFromData(m.file)
	if err != nil {
		return nil, err
	}

	return NewExtractor(data), nil
}
