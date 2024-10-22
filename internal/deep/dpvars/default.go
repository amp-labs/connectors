package dpvars

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

// EmptyMetadataVariables is a default implementation of MetadataVariables.
// Usually connector will have no metadata, so no fields to store.
type EmptyMetadataVariables struct{}

func (e *EmptyMetadataVariables) FromMap(map[string]string) {
	// no-op
}

func (e *EmptyMetadataVariables) ToMap() map[string]string {
	return nil
}

func (e *EmptyMetadataVariables) GetSubstitutionPlans() []paramsbuilder.SubstitutionPlan {
	return nil
}

func (e *EmptyMetadataVariables) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID: requirements.MetadataVariables,
		Constructor: func() *EmptyMetadataVariables {
			return &EmptyMetadataVariables{}
		},
		Interface: new(MetadataVariables),
	}
}
