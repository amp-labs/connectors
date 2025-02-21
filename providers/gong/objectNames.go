package gong

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/gong/metadata"
)

const (
	objectNameCalls      = "calls"
	objectNameTranscript = "transcripts"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var supportedObjectsByWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCalls,
)

var postReadObjects = datautils.NewSet(
	objectNameTranscript,
)

var ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectNameTranscript: "callTranscripts",
},
	func(objectName string) (fieldName string) {
		return objectName
	},
)
