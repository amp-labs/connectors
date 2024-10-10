package deep

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type Method string

const (
	ReadMethod   Method = "READ"
	CreateMethod Method = "CREATE"
	UpdateMethod Method = "UPDATE"
	DeleteMethod Method = "DELETE"
)

type ObjectURLResolver interface {
	requirements.ConnectorComponent
	FindURL(method Method, baseURL, objectName string) (*urlbuilder.URL, error)
}

type SingleURLFormat struct {
	Produce func(method Method, baseURL, objectName string) (*urlbuilder.URL, error)
}

func (r SingleURLFormat) FindURL(method Method, baseURL, objectName string) (*urlbuilder.URL, error) {
	return r.Produce(method, baseURL, objectName)
}

var _ ObjectURLResolver = SingleURLFormat{}

func (r SingleURLFormat) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "objectUrlResolver",
		Constructor: handy.Returner(r),
		Interface:   new(ObjectURLResolver),
	}
}
