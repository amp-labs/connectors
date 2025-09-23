package gong

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/gong/metadata"
)

const (
	objectNameCalls              = "calls"
	objectNameTranscript         = "transcripts"
	objectNameMeetings           = "meetings"
	objectNameDigitalInteraction = "digital-interaction"
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var supportedObjectsByWrite = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCalls,
	objectNameMeetings,
	objectNameDigitalInteraction,
)

var postReadObjects = datautils.NewSet( //nolint:gochecknoglobals
	objectNameTranscript,
)

var ObjectNameToResponseField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectNameTranscript: "callTranscripts",
},
	func(objectName string) (fieldName string) {
		return objectName
	},
)

var objectNameToWriteResultIDField = datautils.NewDefaultMap(map[string]string{ //nolint:gochecknoglobals
	objectNameCalls:              "callId",
	objectNameMeetings:           "meetingId",
	objectNameDigitalInteraction: "requestId",
},
	func(objectName string) (fieldName string) {
		return "id"
	},
)
