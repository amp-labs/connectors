package keap

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/keap/metadata"
)

const (
	// ModuleV1 is a grouping of V1 API endpoints.
	// https://developer.keap.com/docs/rest/
	ModuleV1 common.ModuleID = "version1"
	// ModuleV2 is a grouping of V2 API endpoints.
	// https://developer.keap.com/docs/restv2/
	ModuleV2 common.ModuleID = "version2"
)

// SupportedModules represents currently working and supported modules within the Keap connector.
// Modules are added to schema.json file using OpenAPI script.
var SupportedModules = metadata.Schemas.ModuleRegistry() // nolint: gochecknoglobals
