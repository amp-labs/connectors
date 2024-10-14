package dpobjects

import (
	"errors"

	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

// ErrNoMatchingURL is returned when Object is unknown, and we cannot match it with provider API.
var ErrNoMatchingURL = errors.New("cannot match URL for object")

// URLFormat creates URL using custom callback.
type URLFormat struct {
	Produce func(method Method, baseURL, objectName string) (*urlbuilder.URL, error)
}

func (r URLFormat) FindURL(method Method, baseURL, objectName string) (*urlbuilder.URL, error) {
	// Default behaviour. Callback is required.
	if r.Produce == nil {
		return nil, ErrNoMatchingURL
	}

	return r.Produce(method, baseURL, objectName)
}

func (r URLFormat) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.ObjectURLResolver,
		Constructor: handy.PtrReturner(r),
		Interface:   new(URLResolver),
	}
}
