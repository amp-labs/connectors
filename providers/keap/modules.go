package keap

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/keap/metadata"
)

const (
	// ModuleV1
	// Deprecated.
	ModuleV1 = common.ModuleID(providers.ModuleKeapV1)
	// ModuleV2
	// Deprecated.
	ModuleV2 = common.ModuleID(providers.ModuleKeapV2)
)

// SupportedModules represents currently working and supported modules within the Keap connector.
// Modules are added to schema.json file using OpenAPI script.
var SupportedModules = metadata.Schemas.ModuleRegistry() // nolint: gochecknoglobals
