package google

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/google/metadata"
)

const (
	ModuleCalendar common.ModuleID = "calendar"
)

// SupportedModules represents currently working and supported modules within the Google connector.
// Modules are added to schema.json file using Google Discovery script.
var SupportedModules = metadata.Schemas.ModuleRegistry() // nolint: gochecknoglobals
