package deep

import (
	"errors"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/providers"
	"go.uber.org/dig"
)

type Dependency struct {
	Constructor any
}

func (d Dependency) apply(container *dig.Container) error {
	return container.Provide(d.Constructor)
}

func Connector[C any, P paramsbuilder.ParamAssurance](
	connectorConstructor any,
	provider providers.Provider,
	errorHandler *interpreter.ErrorHandler,
	options []func(params *P),
	dependencies ...Dependency,
) (*C, error) {

	core := []Dependency{
		{
			// Connector must have Provider name
			Constructor: func() providers.Provider {
				return provider
			},
		},
		{
			// Connector is configured using options.
			Constructor: func() []func(params *P) {
				return options
			},
		},
		{
			// HTTP clients use error handler.
			Constructor: func() interpreter.ErrorHandler {
				if errorHandler == nil {
					return interpreter.ErrorHandler{}
				}

				return *errorHandler
			},
		},
		{
			// Connector may choose to be empty closer.
			Constructor: func() *EmptyCloser {
				return &EmptyCloser{}
			},
		},
		{
			// Connector will have HTTP clients which can be implied from parameters "P".
			Constructor: newClients[P],
		},
		{
			// Connector may serve ListObjectMetadata from static file.
			// Note: this requires another dependency of *scrapper.ObjectMetadataResult.
			Constructor: NewStaticMetadata,
		},
		{
			// This is the main constructor which will get all dependencies resolved.
			// It is possible that not all dependencies are needed, this list is exhaustive,
			// which describes all the building blocks that Deep connector may have.
			Constructor: connectorConstructor,
		},
	}

	var err error
	container := dig.New()
	for _, dependency := range append(core, dependencies...) {
		err = errors.Join(err, dependency.apply(container))
	}

	if err != nil {
		return nil, err
	}

	return resolveDependencies[C](container)
}

func resolveDependencies[T any](container *dig.Container) (*T, error) {
	var result *T
	err := container.Invoke(func(builder *T) {
		result = builder
	})
	return result, err
}
