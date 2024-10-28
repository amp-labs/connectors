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
