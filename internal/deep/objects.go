package deep

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type ObjectManager interface {
	requirements.Requirement

	IsReadSupported(objectName string) bool
}

var _ ObjectManager = ObjectRegistry{}
var _ ObjectManager = EmptyObjectRegistry{}

type ObjectRegistry struct {
	Read handy.Set[string]
}

func (o ObjectRegistry) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "objectRegistry",
		Constructor: handy.Returner(o),
		Interface:   new(ObjectManager),
	}
}

func (o ObjectRegistry) IsReadSupported(objectName string) bool {
	return o.Read.Has(objectName)
}

type EmptyObjectRegistry struct{}

func (e EmptyObjectRegistry) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "objectRegistry",
		Constructor: handy.Returner(e),
		Interface:   new(ObjectManager),
	}
}

func (e EmptyObjectRegistry) IsReadSupported(string) bool {
	return true
}
