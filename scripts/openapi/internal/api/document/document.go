// Package document
// Traverse openapi3 file to acquire object schema, its fields.
// Allows processing openapi3 nodes related to certain object.
package document

import (
	"errors"

	"github.com/getkin/kin-openapi/openapi3"
)

type HTTPOperation string

var ErrUnprocessableObject = errors.New("don't know how to process schema")

// Document is a wrapper of openapi with null checks.
type Document struct {
	delegate *openapi3.T
}

func New(data *openapi3.T) *Document {
	return &Document{
		delegate: data,
	}
}

func (s Document) GetPaths() []PathItem {
	paths := s.delegate.Paths
	if paths == nil {
		return nil
	}

	result := make([]PathItem, 0)
	for urlPath, pathItem := range paths.Map() {
		result = append(result, PathItem{
			urlPath:  urlPath,
			delegate: pathItem,
		})
	}

	return result
}

type PathItem struct {
	urlPath  string
	delegate *openapi3.PathItem
}
