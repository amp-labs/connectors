package api3

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// OpenapiFileManager locates openapi file.
// Allows to read data of interest.
// Use it when dealing with OpenAPI v3.
type OpenapiFileManager[C any] struct {
	file []byte
}

func NewOpenapiFileManager[C any](file []byte) *OpenapiFileManager[C] {
	return &OpenapiFileManager[C]{
		file: file,
	}
}

func (m OpenapiFileManager[C]) GetExplorer(opts ...Option) (*Explorer[C], error) {
	loader := openapi3.NewLoader()

	data, err := loader.LoadFromData(m.file)
	if err != nil {
		return nil, err
	}

	return NewExplorer[C](data, opts...), nil
}
