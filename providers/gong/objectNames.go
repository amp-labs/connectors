package gong

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/gong/metadata"
)

const (
	objectNameCalls = "calls"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var supportedObjectsByWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCalls,
)
