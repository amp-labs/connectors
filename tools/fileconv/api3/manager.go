package api3

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// OpenapiFileManager locates openapi file.
// Allows to read data of interest.
type OpenapiFileManager struct {
	file []byte
}

func NewOpenapiFileManager(file []byte) *OpenapiFileManager {
	return &OpenapiFileManager{
		file: file,
	}
}

func (m OpenapiFileManager) GetExplorer(opts ...Option) (*Explorer, error) {
	loader := openapi3.NewLoader()

	data, err := loader.LoadFromData(m.file)
	if err != nil {
		return nil, err
	}

	return &Explorer{
		schema: &Document{
			delegate: data,
		},
		parameters: createParams(opts),
	}, nil
}
