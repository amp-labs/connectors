package deep

import "github.com/amp-labs/connectors/common/urlbuilder"

type URLResolver interface {
	ResolveURL(baseURL, objectName string) (*urlbuilder.URL, error)
}

type DirectURLResolver struct {
	Resolve func(baseURL, objectName string) (*urlbuilder.URL, error)
}

func (d DirectURLResolver) ResolveURL(baseURL, objectName string) (*urlbuilder.URL, error) {
	return d.Resolve(baseURL, objectName)
}
