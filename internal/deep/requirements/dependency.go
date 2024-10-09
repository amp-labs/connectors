package requirements

import (
	"errors"
	"github.com/amp-labs/connectors/common/handy"
	"go.uber.org/dig"
)

type Requirement interface {
	Satisfies() Dependency
}

type Dependency struct {
	ID          string
	Constructor any
}

func (d Dependency) apply(container *dig.Container) error {
	return container.Provide(d.Constructor)
}

type Dependencies handy.Map[string, Dependency]

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
