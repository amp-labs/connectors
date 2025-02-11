package google

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/google/metadata"
)

const ModuleCalendar common.ModuleID = "calendar"

var SupportedModules = metadata.Schemas.ModuleRegistry() // nolint: gochecknoglobals
