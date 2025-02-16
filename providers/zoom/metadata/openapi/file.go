package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

// Static file containing openapi spec.
//

var (
	//go:embed user.json
	usersFile []byte

	//go:embed meeting.json
	meetingFile []byte

	UsersFileManager   = api3.NewOpenapiFileManager(usersFile)   // nolint:gochecknoglobals
	MeetingFileManager = api3.NewOpenapiFileManager(meetingFile) // nolint:gochecknoglobalsw
)
