package discovery

import (
	_ "embed"

	"github.com/amp-labs/connectors/tools/fileconv/googledicsovery"
)

var (
	// Static file containing google discovery spec.
	//
	//go:embed calendar-discovery-file.json
	apiCalendarFile []byte

	CalendarFileManager = googledicsovery.NewFileManager(apiCalendarFile) // nolint:gochecknoglobals
)
