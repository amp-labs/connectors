package deep

import (
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep/dpobjects"
	"github.com/amp-labs/connectors/internal/deep/dpread"
	"github.com/amp-labs/connectors/internal/deep/dpremove"
	"github.com/amp-labs/connectors/internal/deep/dprequests"
	"github.com/amp-labs/connectors/internal/deep/dpvars"
	"github.com/amp-labs/connectors/internal/deep/dpwrite"
	"github.com/amp-labs/connectors/internal/deep/requirements"
	"github.com/amp-labs/connectors/providers"
	"go.uber.org/dig"
)

// Connector
// TODO documentation. !!!! THIS IS THE MOST CRUCIAL METHOD !!!!
func Connector[C any, P paramsbuilder.ParamAssurance](
	connectorConstructor any,
	provider providers.Provider,
	options []func(params *P),
	reqs ...requirements.ConnectorComponent,
) (*C, error) {
	return ExtendedConnector[C, P, *dpvars.EmptyMetadataVariables](
		connectorConstructor, provider, &dpvars.EmptyMetadataVariables{}, options, reqs...,
	)
}

// ExtendedConnector is the same connector builder as Connector method.
// TODO documentation.
func ExtendedConnector[C any, P paramsbuilder.ParamAssurance, D dpvars.MetadataVariables](
	connectorConstructor any,
	provider providers.Provider,
	metadataVariables D,
	options []func(params *P),
	components ...requirements.ConnectorComponent,
) (*C, error) {
	// This is a default list of dependencies available for a "connectorConstructor" to pick up.
	dependencies := requirements.NewDependencies([]requirements.Dependency{
		{
			// Connector must have Provider name.
			ID: requirements.Provider,
			Constructor: func() providers.Provider {
				return provider
			},
		},
		{
			// Connector is configured using options.
			ID: requirements.Options,
			Constructor: func() []func(params *P) {
				return options
			},
		},
		// Metadata Variables hold connector specific data fields.
		// They are inferred from parameters
		metadataVariables.Satisfies(),

		// Options are realized into parameters.
		// Some parameters can be catalog or metadata variables.
		dpvars.Parameters[P]{}.Satisfies(),
		dpvars.ConnectorData[P, D]{}.Satisfies(),
		dpvars.CatalogVariables[P, D]{}.Satisfies(),

		// Every connector makes requests. Clients holds authenticated HTTP clients
		// capable of doing JSON, XML calls. It needs parameters, catalog vars for proper setup.
		// ErrorHandler would parse error response depending on media type.
		// HeaderSupplements is used to attach headers when performing said requests.
		{
			ID:          requirements.Clients,
			Constructor: dprequests.NewClients[P, D],
		},
		interpreter.ErrorHandler{}.Satisfies(),
		dprequests.HeaderSupplements{}.Satisfies(),

		// Guards against unsupported objects.
		// By default, every object would reach Reader, Writer, etc.
		dpobjects.Registry{}.Satisfies(),

		// Most connectors do no-op on close.
		// *EmptyCloser is available as constructor argument.
		EmptyCloser{}.Satisfies(),

		// READ
		// Default behaviour:
		//  -> first page is not changes and uses what dpobjects.URLResolver has given.
		//  -> next page is empty, hence no pagination
		//  -> read is done using GET operation.
		// *Reader is available as constructor argument.
		Reader{}.Satisfies(),
		dpread.FirstPageBuilder{}.Satisfies(),
		dpread.NextPageBuilder{}.Satisfies(),
		dpread.RequestGet{}.Satisfies(),
		dpread.ResponseLocator{}.Satisfies(),

		// WRITE
		// Default behaviour:
		//  -> create is done using POST operation.
		//  -> update is done using PUT operation.
		//  -> write response is not parsed and returns success.
		// *Writer is available as constructor argument.
		Writer{}.Satisfies(),
		dpwrite.RequestPostPut{}.Satisfies(),
		dpwrite.ResponseBuilder{}.Satisfies(),

		// METADATA
		// Default behaviour:
		//  -> expects dpmetadata.SchemaHolder.
		// *StaticMetadata is available as constructor argument.
		StaticMetadata{}.Satisfies(),

		// DELETE
		// Default behaviour:
		//  -> remove is done using DELETE operation.
		// *Remover is available as constructor argument.
		Remover{}.Satisfies(),
		dpremove.RequestDelete{}.Satisfies(),

		{
			// This is the main constructor which will get all dependencies resolved.
			// It is possible that not all dependencies are needed, this list is exhaustive,
			// which describes all the building blocks that Deep connector may have.
			ID:          requirements.Connector,
			Constructor: connectorConstructor,
		},
	})

	for _, component := range components {
		dependencies.Add(component.Satisfies())
	}

	container := dig.New()
	if err := dependencies.Apply(container); err != nil {
		return nil, err
	}

	return resolveDependencies[C](container)
}

// Tries to invoke connector constructor. The dig.Container will have dependencies that will be injected.
// Produces connector of a specified T type.
func resolveDependencies[T any](container *dig.Container) (*T, error) {
	var result *T
	err := container.Invoke(func(builder *T) {
		result = builder
	})

	return result, err
}
