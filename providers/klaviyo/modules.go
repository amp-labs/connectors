package klaviyo

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/klaviyo/metadata"
)

const (
	// Module2024Oct15
	// Deprecated.
	Module2024Oct15 = common.ModuleID(providers.ModuleKlaviyo2024Oct15)
)

// SupportedModules represents currently working and supported modules within the Klaviyo connector.
// Modules are added to schema.json file using OpenAPI script.
var SupportedModules = metadata.Schemas.ModuleRegistry() // nolint: gochecknoglobals
