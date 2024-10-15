package atlassian

import (
	"github.com/amp-labs/connectors/common/substitutions/catalogreplacer"
	"github.com/amp-labs/connectors/internal/deep/dpvars"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

const cloudIdKey = "cloudId"

// AuthMetadataVars is a complete list of authentication metadata associated with connector.
// This model serves as a documentation of map[string]string contents.
type AuthMetadataVars struct {
	CloudId string
}

func (v *AuthMetadataVars) FromMap(dictionary map[string]string) {
	v.CloudId = dictionary[cloudIdKey]
}

func (v *AuthMetadataVars) ToMap() map[string]string {
	return map[string]string{
		cloudIdKey: v.CloudId,
	}
}

func (v *AuthMetadataVars) GetSubstitutionPlans() []catalogreplacer.SubstitutionPlan {
	return nil
}

func NewAuthMetadataVars() *AuthMetadataVars {
	return &AuthMetadataVars{CloudId: ""}
}

var _ dpvars.MetadataVariables = &AuthMetadataVars{}

func (v *AuthMetadataVars) Satisfies() requirements.Dependency {
	return requirements.Dependency{
		ID:          requirements.MetadataVariables,
		Constructor: NewAuthMetadataVars,
		Interface:   new(dpvars.MetadataVariables),
	}
}
