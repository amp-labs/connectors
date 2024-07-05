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
	GetSubstitutionPlan() SubstitutionPlan
}

// SubstitutionPlan defines an intent to replace `from` with `to`.
type SubstitutionPlan struct {
	From string
	To   string
}

type RegistryValue interface {
	string | *ajson.Node
}

type SubstitutionRegistry[V RegistryValue] map[string]V

func NewCatalogSubstitutionRegistry(vars []CatalogVariable) SubstitutionRegistry[string] {
	substitutions := make(SubstitutionRegistry[string])

	for _, variable := range vars {
		s := variable.GetSubstitutionPlan()
		substitutions[s.From] = s.To
	}

	return substitutions
}

// NewCatalogVariables converts JSON into supported list of Catalog Variables.
// This enforces type checking.
func NewCatalogVariables[V RegistryValue](registry SubstitutionRegistry[V]) []CatalogVariable {
	result := make([]CatalogVariable, 0)

	for key, value := range registry {
		name := registryValueToString(value)

		switch key {
		case variableWorkspace:
			result = append(result, &Workspace{Name: name})
		default:
			slog.Info("unknown substitution SubstitutionPlan for catalog", key, value)
		}
	}

	return result
}

func registryValueToString[V RegistryValue](value V) string {
	var name string
	if v, ok := any(value).(string); ok {
		name = v
	}

	if v, ok := any(value).(*ajson.Node); ok {
		name = v.MustString()
	}

	return name
}
