package zoom

import (
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoom/metadata"
)

const (
	// ModuleUser
	// Deprecated.
	ModuleUser = providers.ModuleZoomUser
	// ModuleMeeting
	// Deprecated.
	ModuleMeeting = providers.ModuleZoomMeeting
)

var SupportedModules = metadata.Schemas.ModuleRegistry() // nolint: gochecknoglobals
