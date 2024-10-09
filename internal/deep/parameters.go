package deep

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type Parameters[P paramsbuilder.ParamAssurance] struct {
	Params      any
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

	return &Parameters[P]{
		Params:      parameters,
	}, nil
}

func (p Parameters[P]) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "parameters",
		Constructor: newParametersHolder[P],
	}
}
