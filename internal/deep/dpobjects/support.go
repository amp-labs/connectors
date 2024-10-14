package dpobjects

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

// SupportRegistry lists object names allowed to perform operation.
// By default, everybody can pass.
type SupportRegistry struct {
	Read   handy.Set[string]
	Write  handy.Set[string]
	Delete handy.Set[string]
	// AllowAllObjects is optional.
	// By default, empty set means no object can pass.
	// Setting this flag to true will allow any object passing if the set is empty.
	AllowAllObjects bool
}

func (o SupportRegistry) IsReadSupported(objectName string) bool {
	if len(o.Read) == 0 {
		return o.defaultBehavior()
	}

	return o.Read.Has(objectName)
}

func (o SupportRegistry) IsWriteSupported(objectName string) bool {
	if len(o.Write) == 0 {
		return o.defaultBehavior()
	}

	return o.Write.Has(objectName)
}

func (o SupportRegistry) IsDeleteSupported(objectName string) bool {
	if len(o.Delete) == 0 {
		return o.defaultBehavior()
	}

	return o.Delete.Has(objectName)
}

func (o SupportRegistry) defaultBehavior() bool {
	return o.AllowAllObjects
}

func (o SupportRegistry) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.ObjectRegistry,
		Constructor: handy.PtrReturner(o),
		Interface:   new(Support),
	}
}
