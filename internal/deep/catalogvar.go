package deep

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type CatalogVariables[P paramsbuilder.ParamAssurance, D MetadataVariables] struct {
	List []paramsbuilder.CatalogVariable
}

func newCatalogVariables[P paramsbuilder.ParamAssurance, D MetadataVariables](
	parameters *Parameters[P],
	data *ConnectorData[P, D],
) *CatalogVariables[P, D] {
	variables := paramsbuilder.ExtractCatalogVariables(parameters.Params)

	// Sometimes connector metadata holds CatalogVariables, collect them here.
	plans := data.Metadata.GetSubstitutionPlans()
	for _, plan := range plans {
		variables = append(variables, &paramsbuilder.CustomCatalogVariable{
			Plan: plan,
		})
	}

	return &CatalogVariables[P, D]{
		List: variables,
	}
}

func (c CatalogVariables[P, D]) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "catalogVariables",
		Constructor: newCatalogVariables[P, D],
	}
}
