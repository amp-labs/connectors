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

type URLResolver struct {
	Resolve func(method Method, baseURL, objectName string) (*urlbuilder.URL, error)
}

func (r URLResolver) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "urlResolver",
		Constructor: handy.Returner(r),
	}
}
