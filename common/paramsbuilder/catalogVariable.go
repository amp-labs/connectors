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

// ExtractCatalogVariables accepts any struct that embeds one or multiple CatalogVariables.
// It will try to explore all known implementors of CatalogVariable and return them.
func ExtractCatalogVariables(parameters any) []catalogreplacer.CatalogVariable {
	var catalogsVars []catalogreplacer.CatalogVariable

	// Workspace is the only known CatalogVariable
	if workspaceHolder, ok := parameters.(WorkspaceHolder); ok {
		workspace := workspaceHolder.GiveWorkspace()
		catalogsVars = append(catalogsVars, workspace)
	}

	return catalogsVars
}

// CustomCatalogVariable is a variable that can be created on the fly. Just specify the plan of what
// should be replaced with what data, it implements CatalogVariable.
type CustomCatalogVariable struct {
	Plan catalogreplacer.SubstitutionPlan
}

var _ catalogreplacer.CatalogVariable = CustomCatalogVariable{}

func (c CustomCatalogVariable) GetSubstitutionPlan() catalogreplacer.SubstitutionPlan {
	return c.Plan
}
