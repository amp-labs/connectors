package requirements

import (
	"errors"

	"github.com/amp-labs/connectors/common/handy"
	"go.uber.org/dig"
)

// ConnectorComponent is a building block that satisfies connector dependency.
//
// Major connector component captures the logic to perform Read/Write/ListObjectMetadata/Delete operations.
// It is such that satisfies the main immediate goal of connector.
// Such examples include deep.Reader, deep.Writer, deep.StaticMetadata, deep.Remover.
// When building deep connector you would embed Major Component to gain that functionality.
//
// Major component is a template with default behaviour, some steps in the implementation would need customization,
// which is achieved through ConnectorComponent, representing those steps.
// This structure uses the Template Design Pattern via dependency injection.
//
// ConnectorComponent may need multiple other ConnectorComponents,
// they are wired automatically by matching Dependency definition.
//
// Only one ConnectorComponent with identical ID can exist in a pool.
// By providing ConnectorComponent that satisfies existing dependency
// you are in fact replacing one behaviour with the other.
type ConnectorComponent interface {
	Satisfies() Dependency
}

// Dependency holds a factory which produces instances of certain type.
// It creates instances associated with ID.
type Dependency struct {
	ID          ComponentID
	Constructor any
	Interface   any // TODO interface should be implied based on ID
}

// Makes itself available to dig.Container.
func (d Dependency) apply(container *dig.Container) error {
	var options []dig.ProvideOption
	if d.Interface != nil {
		options = append(options, dig.As(d.Interface))
	}

	return container.Provide(d.Constructor, options...)
}

// Dependencies is a map indexed by Dependency.ID.
type Dependencies handy.Map[ComponentID, Dependency]

// NewDependencies is a constructor that expects input to be in order.
// Any duplicates will be eliminated, which means the last in order would win, others ignored.
func NewDependencies(deps []Dependency) Dependencies {
	result := Dependencies{}
	for _, dep := range deps {
		result[dep.ID] = dep
	}

	return result
}

func (d Dependencies) Add(dep Dependency) {
	d[dep.ID] = dep
}

func (d Dependencies) Apply(container *dig.Container) error {
	var err error
	for _, dep := range d {
		err = errors.Join(err, dep.apply(container))
	}

	return err
}
