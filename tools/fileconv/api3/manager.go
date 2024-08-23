package api3

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// OpenapiFileManager locates openapi file.
// Allows to read data of interest.
type OpenapiFileManager struct {
	openapi []byte
}

func NewOpenapiFileManager(openapi []byte) *OpenapiFileManager {
	return &OpenapiFileManager{
		openapi: openapi,
	}
}

func (m OpenapiFileManager) GetExplorer(opts ...Option) (*Explorer, error) {
	loader := openapi3.NewLoader()

	data, err := loader.LoadFromData(m.openapi)
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
