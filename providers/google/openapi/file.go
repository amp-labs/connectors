package openapi

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	// Static file containing openapi spec.
	//
	//go:embed calendar.json
	calendarAPIFile []byte

	//go:embed people.json
	peopleAPIFile []byte

	CalendarFileManager = api3.NewOpenapiFileManager[any](calendarAPIFile) // nolint:gochecknoglobals
	PeopleFileManager   = api3.NewOpenapiFileManager[any](peopleAPIFile)   // nolint:gochecknoglobals
)
