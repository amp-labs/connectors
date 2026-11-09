package files

import (
	_ "embed"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api"
)

// Static file containing openapi spec.
//

var (
	//go:embed user.json
	usersFile []byte

	//go:embed meeting.json
	meetingFile []byte

	UsersFile   = api.NewFile(usersFile)   // nolint:gochecknoglobals
	MeetingFile = api.NewFile(meetingFile) // nolint:gochecknoglobals

	OutputDir = "providers/zoom/internal/metadata" // nolint:gochecknoglobals
)
