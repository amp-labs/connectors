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

	CalendarFileManager = api3.NewOpenapiFileManager(calendarAPIFile) // nolint:gochecknoglobals
)
