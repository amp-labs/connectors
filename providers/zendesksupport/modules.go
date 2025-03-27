package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

const (
	// ModuleTicketing
	// Deprecated.
	ModuleTicketing = common.ModuleID(providers.ModuleZendeskTicketing)
	// ModuleHelpCenter
	// Deprecated.
	ModuleHelpCenter = common.ModuleID(providers.ModuleZendeskHelpCenter)
)

// SupportedModules represents currently working and supported modules within the Zendesk connector.
// Modules are added to schema.json file using OpenAPI script.
var SupportedModules = metadata.Schemas.ModuleRegistry() // nolint: gochecknoglobals
