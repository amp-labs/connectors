package paramsbuilder

import (
	"log/slog"

	"github.com/spyzhov/ajson"
)

const (
	variableWorkspace = "workspace"
)

// CatalogVariable allows dynamically to replace variables represented with `{{VAR_NAME}}` string.
type CatalogVariable interface {
	getSubstitutionPlan() SubstitutionPlan
}

// SubstitutionPlan defines an intent to replace `from` with `to`.
type SubstitutionPlan struct {
	from string
	to   string
}

type SubstitutionRegistry map[string]string

func GetCatalogSubstitutionRegistry(vars []CatalogVariable) SubstitutionRegistry {
	substitutions := make(SubstitutionRegistry)

	for _, variable := range vars {
		s := variable.getSubstitutionPlan()
		substitutions[s.from] = s.to
	}

	return substitutions
}

// NewCatalogVariables converts JSON into supported list of Catalog Variables.
// This enforces type checking.
func NewCatalogVariables(registry map[string]*ajson.Node) []CatalogVariable {
	result := make([]CatalogVariable, 0)

	for key, value := range registry {
		switch key {
		case variableWorkspace:
			result = append(result, &Workspace{Name: value.MustString()})
		default:
			slog.Info("unknown substitution SubstitutionPlan for catalog", key, value)
		}
	}

	return result
}
