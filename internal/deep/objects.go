package deep

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type ObjectManager interface {
	requirements.Requirement

	IsReadSupported(objectName string) bool
	IsWriteSupported(objectName string) bool
}

var _ ObjectManager = ObjectRegistry{}
var _ ObjectManager = EmptyObjectRegistry{}

type ObjectRegistry struct {
	Read  handy.Set[string]
	Write handy.Set[string]
}

func (o ObjectRegistry) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "objectRegistry",
		Constructor: handy.Returner(o),
		Interface:   new(ObjectManager),
	}
}

func (o ObjectRegistry) IsReadSupported(objectName string) bool {
	if len(o.Read) == 0 {
		return true
	}

	return o.Read.Has(objectName)
}

func (o ObjectRegistry) IsWriteSupported(objectName string) bool {
	if len(o.Write) == 0 {
		return true
	}

	return o.Write.Has(objectName)
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

func (e EmptyObjectRegistry) IsWriteSupported(string) bool {
	return true
}
