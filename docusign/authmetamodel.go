package docusign

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
)

const serverKey = "server"

// Metadata fields that must be specified to initialize connector.
var requiredMetadataFields = []string{ // nolint:gochecknoglobals
	serverKey,
}

// AuthMetadataVars is a complete list of authentication metadata associated with connector.
// This model serves as a documentation of map[string]string contents.
type AuthMetadataVars struct {
	Server string
}

// NewAuthMetadataVars parses map into the model.
func NewAuthMetadataVars(dictionary map[string]string) *AuthMetadataVars {
	return &AuthMetadataVars{
		Server: dictionary[serverKey],
	}
}

// AsMap converts model back to the map.
func (v AuthMetadataVars) AsMap() *map[string]string {
	return &map[string]string{
		serverKey: v.Server,
	}
}

// GetSubstitutionPlan allows variable substitution when resolving provider information.
// Only server URL is supported.
func (v AuthMetadataVars) GetSubstitutionPlan() paramsbuilder.SubstitutionPlan {
	return paramsbuilder.SubstitutionPlan{
		From: serverKey,
		To:   v.Server,
	}
}
