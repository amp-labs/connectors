package deep

import "github.com/amp-labs/connectors/common/urlbuilder"

type URLResolver struct {
	Resolve func(baseURL, objectName string) (*urlbuilder.URL, error)
}

func (r URLResolver) Satisfies() Dependency {
	return Dependency{Constructor: returner(r)}
}

func returner[T any](self T) func() *T {
	return func() *T {
		return &self
	}
}
