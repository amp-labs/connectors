package zendesksupport

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/zendesksupport/metadata"
)

const (
	// ModuleTicketing is used for proxying requests through.
	// https://developer.zendesk.com/api-reference/ticketing/introduction/
	ModuleTicketing common.ModuleID = "ticketing"
	// ModuleHelpCenter is Zendesk Help Center.
	// https://developer.zendesk.com/api-reference/help_center/help-center-api/introduction/
	ModuleHelpCenter common.ModuleID = "help-center"
)

// SupportedModules represents currently working and supported modules within the Zendesk connector.
// Modules are added to schema.json file using OpenAPI script.
var SupportedModules = metadata.Schemas.ModuleRegistry() // nolint: gochecknoglobals
