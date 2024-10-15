package gong

import (
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/providers/gong/metadata"
)

const (
	objectNameCalls = "calls"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = handy.NewSetFromList( //nolint:gochecknoglobals
	metadata.Schemas.GetObjectNames(),
)

var supportedObjectsByWrite = handy.NewSet( //nolint:gochecknoglobals
	objectNameCalls,
)
