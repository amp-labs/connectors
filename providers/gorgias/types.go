package gorgias

import "github.com/amp-labs/connectors/internal/datautils"

var supportedReadOPbjects = datautils.NewSet( //nolint:gochecknoglobals
	"account",
	"customers",
	"custom-fields",
	"events",
	"integrations",
	"jobs",
	"macros",
	"rules",
	"satisfaction-surveys",
	"tags",
	"teams",
	"tickets",
	"messages",
	"users",
	"views",
	"phone/voice-calls",
	"phone/voice-call-recordings",
	"phone/voice-call-events",
)
