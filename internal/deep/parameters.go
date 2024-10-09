package deep

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type Parameters[P paramsbuilder.ParamAssurance] struct {
	Params      any
	CatalogVars []paramsbuilder.CatalogVariable
}

func newParametersHolder[P paramsbuilder.ParamAssurance](
	opts []func(params *P),
) (*Parameters[P], error) {
	// Apply option functions to produce final parameters instance.
	var paramsTemplate P
	parameters, err := paramsbuilder.Apply(paramsTemplate, opts)
	if err != nil {
		return nil, err
	}

	// Parameters may have fields that count as Catalog variables.
	// They are stored separately for quick access.
	catalogVariables := paramsbuilder.ExtractCatalogVariables(parameters)

	return &Parameters[P]{
		Params:      parameters,
		CatalogVars: catalogVariables,
	}, nil
}

func (p Parameters[P]) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "parameters",
		Constructor: newParametersHolder[P],
	}
}
