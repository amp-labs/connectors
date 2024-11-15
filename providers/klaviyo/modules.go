package klaviyo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/klaviyo/metadata"
)

const (
	// Module2024Oct15 is the latest stable version of API as of the date of writing.
	// https://developers.klaviyo.com/en/reference/api_overview
	Module2024Oct15 common.ModuleID = "2024-10-15"
)

// SupportedModules represents currently working and supported modules within the Klaviyo connector.
// Modules are added to schema.json file using OpenAPI script.
var SupportedModules = metadata.Schemas.ModuleRegistry() // nolint: gochecknoglobals
