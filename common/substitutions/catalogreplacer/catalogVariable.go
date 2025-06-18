package catalogreplacer

import (
	"github.com/amp-labs/connectors/common/substitutions"
)

const (
	VariableWorkspace = "workspace"
)

// CatalogVariable allows dynamically to replace variables represented with `{{VAR_NAME}}` string.
type CatalogVariable interface {
	GetSubstitutionPlan() SubstitutionPlan
}

type CatalogVariables []CatalogVariable

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

// CustomCatalogVariable is a variable that can be created on the fly. Just specify the plan of what
// should be replaced with what data, it implements CatalogVariable.
type CustomCatalogVariable struct {
	Plan SubstitutionPlan
}

var _ CatalogVariable = CustomCatalogVariable{}

func (c CustomCatalogVariable) GetSubstitutionPlan() SubstitutionPlan {
	return c.Plan
}

func (c *CatalogVariables) AddDefaults(defaults ...CatalogVariable) {
	registry := make(map[string]CatalogVariable)

	for _, variable := range *c {
		plan := variable.GetSubstitutionPlan()
		registry[plan.From] = variable
	}

	for _, variable := range defaults {
		plan := variable.GetSubstitutionPlan()

		if _, found := registry[plan.From]; !found {
			*c = append(*c, CustomCatalogVariable{Plan: plan})
		}
	}
}
