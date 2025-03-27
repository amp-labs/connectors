package zoom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoom/metadata"
)

const (
	// ModuleUser
	// Deprecated.
	ModuleUser = common.ModuleID(providers.ModuleZoomUser)
	// ModuleMeeting
	// Deprecated.
	ModuleMeeting = common.ModuleID(providers.ModuleZoomMeeting)
)

var SupportedModules = metadata.Schemas.ModuleRegistry() // nolint: gochecknoglobals
