package paramsbuilder

import (
	"log/slog"

	"github.com/amp-labs/connectors/common/substitutions"
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
)

// NewCatalogVariables converts JSON into supported list of Catalog Variables.
// This enforces type checking.
func NewCatalogVariables[V substitutions.RegistryValue](
	registry substitutions.Registry[V],
) []catalogreplacer.CatalogVariable {
	result := make([]catalogreplacer.CatalogVariable, 0)

	for key, value := range registry {
		name := substitutions.RegistryValueToString(value)

		switch key {
		case catalogreplacer.VariableWorkspace:
			result = append(result, &Workspace{Name: name})
		default:
			slog.Info("unknown substitution SubstitutionPlan for catalog", key, value)
		}
	}

	return result
}
