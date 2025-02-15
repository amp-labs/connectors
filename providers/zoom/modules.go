package zoom

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/zoom/metadata"
)

const (
	Users common.ModuleID = "Users"
)

var SupportedModules = metadata.Schemas.ModuleRegistry() // nolint: gochecknoglobals
