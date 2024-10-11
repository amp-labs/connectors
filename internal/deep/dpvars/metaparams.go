package dpvars

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type MetadataVariables interface {
	requirements.ConnectorComponent
	FromMap(map[string]string)
	ToMap() map[string]string
	GetSubstitutionPlans() []paramsbuilder.SubstitutionPlan
}

type EmptyMetadataVariables struct{}

var _ MetadataVariables = &EmptyMetadataVariables{}

func (e *EmptyMetadataVariables) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID: requirements.MetadataVariables,
		Constructor: func() *EmptyMetadataVariables {
			return &EmptyMetadataVariables{}
		},
		Interface: new(MetadataVariables),
	}
}

func (e *EmptyMetadataVariables) FromMap(m map[string]string) {
	// no-op
}

func (e *EmptyMetadataVariables) ToMap() map[string]string {
	return nil
}

func (e *EmptyMetadataVariables) GetSubstitutionPlans() []paramsbuilder.SubstitutionPlan {
	return nil
}
