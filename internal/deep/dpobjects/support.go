package dpobjects

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

// Registry lists object names allowed to perform operation.
// By default, everybody can pass.
type Registry struct {
	Read   handy.Set[string]
	Write  handy.Set[string]
	Delete handy.Set[string]
	// Strict is optional.
	// By default, empty set means everybody can pass.
	// Setting this flag to true will prohibit any object passing unless present in the set.
	Strict bool
}

func (o Registry) IsReadSupported(objectName string) bool {
	if len(o.Read) == 0 {
		return o.defaultBehavior()
	}

	return o.Read.Has(objectName)
}

func (o Registry) IsWriteSupported(objectName string) bool {
	if len(o.Write) == 0 {
		return o.defaultBehavior()
	}

	return o.Write.Has(objectName)
}

func (o Registry) IsDeleteSupported(objectName string) bool {
	if len(o.Delete) == 0 {
		return o.defaultBehavior()
	}

	return o.Delete.Has(objectName)
}

func (o Registry) defaultBehavior() bool {
	return !o.Strict
}

func (o Registry) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.ObjectRegistry,
		Constructor: handy.PtrReturner(o),
		Interface:   new(Support),
	}
}
