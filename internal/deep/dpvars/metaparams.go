package dpvars

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

var (
	// Implementations.
	_ MetadataVariables = &EmptyMetadataVariables{}
)

// MetadataVariables is a connector component which holds parsed paramsbuilder.Metadata.
// In case you need to define custom connector data you should implement this interface.
type MetadataVariables interface {
	requirements.ConnectorComponent

	// FromMap is a main function that accepts values to initialize itself.
	// When processing paramsbuilder.Metadata its data will be available via this method.
	FromMap(map[string]string)

	// ToMap converts itself into dynamic map.
	ToMap() map[string]string

	// GetSubstitutionPlans returns list of catalog substitutions, should catalog require additional metadata.
	GetSubstitutionPlans() []paramsbuilder.SubstitutionPlan
}
