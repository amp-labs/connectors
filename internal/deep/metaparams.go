package deep

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

type MetadataVariables interface {
	requirements.Requirement
	ToMap() map[string]string
	FromMap(map[string]string)
	GetSubstitutionPlans() []paramsbuilder.SubstitutionPlan
}

type EmptyMetadataVariables struct {}

var _ MetadataVariables = EmptyMetadataVariables{}

func (e EmptyMetadataVariables) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          "metadataVariables",
		Constructor: handy.Returner(e),
		Interface:   new(MetadataVariables),
	}
}

func (e EmptyMetadataVariables) ToMap() map[string]string {
	return nil
}

func (e EmptyMetadataVariables) FromMap(m map[string]string) {
	// no-op
}

func (e EmptyMetadataVariables) GetSubstitutionPlans() []paramsbuilder.SubstitutionPlan {
	return nil
}
