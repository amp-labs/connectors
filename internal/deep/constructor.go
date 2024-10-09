package deep

import (
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/amp-labs/connectors/providers"
	"go.uber.org/dig"
)

// Connector
// TODO document that it can be a constructor or Dependency object (maybe we want to support DI tagging)
func Connector[C any, P paramsbuilder.ParamAssurance](
	connectorConstructor any,
	provider providers.Provider,
	options []func(params *P),
	reqs ...requirements.Requirement,
) (*C, error) {

	deps := requirements.NewDependencies([]requirements.Dependency{
		{
			// Connector must have Provider name
			ID: "provider",
			Constructor: func() providers.Provider {
				return provider
			},
		},
		{
			// Connector is configured using options.
			ID: "options",
			Constructor: func() []func(params *P) {
				return options
			},
		},
		{
			// HTTP clients use error handler.
			ID: "errorHandler",
			Constructor: func() interpreter.ErrorHandler {
				// Empty by default.
				return interpreter.ErrorHandler{}
			},
		},
		{
			// Connector may choose to be empty closer.
			ID: "closer",
			Constructor: func() *EmptyCloser {
				return &EmptyCloser{}
			},
		},
		EmptyObjectRegistry{}.Satisfies(),
		{
			// Connector will have HTTP clients which can be implied from parameters "P".
			ID:          "clients",
			Constructor: newClients[P],
		},
		{
			// Connector that lists Objects.
			// TODO describe dependencies
			ID:          "reader",
			Constructor: NewReader,
		},
		{
			// Connector that creates new records or updates existing.
			// TODO describe dependencies
			ID:          "writer",
			Constructor: NewWriter,
		},
		{
			// Connector may serve ListObjectMetadata from static file.
			// Note: this requires another dependency of *scrapper.ObjectMetadataResult.
			ID:          "staticMetadata",
			Constructor: NewStaticMetadata,
		},
		{
			// Connector may allow record deletion.
			// TODO describe dependencies
			ID:          "remover",
			Constructor: NewRemover,
		},
		{
			// This is the main constructor which will get all dependencies resolved.
			// It is possible that not all dependencies are needed, this list is exhaustive,
			// which describes all the building blocks that Deep connector may have.
			ID:          "connector",
			Constructor: connectorConstructor,
		},
	})

	for _, requirement := range reqs {
		deps.Add(requirement.Satisfies())
	}

	container := dig.New()
	if err := deps.Apply(container); err != nil {
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
