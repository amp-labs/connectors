package paramsbuilder

import (
	"log/slog"

	"github.com/amp-labs/connectors/common/substitutions"
)

const (
	variableWorkspace = "workspace"
)

// CatalogVariable allows dynamically to replace variables represented with `{{VAR_NAME}}` string.
type CatalogVariable interface {
	GetSubstitutionPlan() SubstitutionPlan
}

// SubstitutionPlan defines an intent to replace `from` with `to`.
type SubstitutionPlan struct {
	From string
	To   string
}

func NewCatalogSubstitutionRegistry(vars []CatalogVariable) substitutions.Registry[string] {
	subs := make(substitutions.Registry[string])

	for _, variable := range vars {
		s := variable.GetSubstitutionPlan()
		subs[s.From] = s.To
	}

	return subs
}

// NewCatalogVariables converts JSON into supported list of Catalog Variables.
// This enforces type checking.
func NewCatalogVariables[V substitutions.RegistryValue](registry substitutions.Registry[V]) []CatalogVariable {
	result := make([]CatalogVariable, 0)

	for key, value := range registry {
		name := substitutions.RegistryValueToString(value)

		switch key {
		case variableWorkspace:
			result = append(result, &Workspace{Name: name})
		default:
			slog.Info("unknown substitution SubstitutionPlan for catalog", key, value)
		}
	}

	return result
}
