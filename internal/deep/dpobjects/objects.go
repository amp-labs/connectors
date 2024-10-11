package dpobjects

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type ObjectManager interface {
	requirements.ConnectorComponent

	IsReadSupported(objectName string) bool
	IsWriteSupported(objectName string) bool
	IsDeleteSupported(objectName string) bool
}

var (
	_ ObjectManager = ObjectSupport{}
	_ ObjectManager = EmptyObjectRegistry{}
)

type ObjectSupport struct {
	Read   handy.Set[string]
	Write  handy.Set[string]
	Delete handy.Set[string]
}

func (o ObjectSupport) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.ObjectRegistry,
		Constructor: handy.Returner(o),
		Interface:   new(ObjectManager),
	}
}

func (o ObjectSupport) IsReadSupported(objectName string) bool {
	if len(o.Read) == 0 {
		return true
	}

	return o.Read.Has(objectName)
}

func (o ObjectSupport) IsWriteSupported(objectName string) bool {
	if len(o.Write) == 0 {
		return true
	}

	return o.Write.Has(objectName)
}

func (o ObjectSupport) IsDeleteSupported(objectName string) bool {
	if len(o.Delete) == 0 {
		return true
	}

	return o.Delete.Has(objectName)
}

type EmptyObjectRegistry struct{}

func (e EmptyObjectRegistry) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.ObjectRegistry,
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

func (e EmptyObjectRegistry) IsDeleteSupported(string) bool {
	return true
}

type ObjectData struct {
	URLPath  string
	NodePath string
}
