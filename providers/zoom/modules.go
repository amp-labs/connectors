package zoom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/zoom/metadata"
)

const (
	ModuleUser common.ModuleID = "user"

	ModuleMeeting common.ModuleID = "meeting"
)

var SupportedModules = metadata.Schemas.ModuleRegistry() // nolint: gochecknoglobals
