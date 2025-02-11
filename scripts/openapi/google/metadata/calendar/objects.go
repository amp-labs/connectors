package calendar

import (
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/providers/google/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	ignoreEndpoints = []string{ // nolint:gochecknoglobals
		"*/watch",
		"/colors",
	}
	displayNameOverride = map[string]string{ // nolint:gochecknoglobals
		"acl":          "ACL",
		"calendarList": "Calendars",
	}
)

func Objects() []metadatadef.Schema {
	explorer, err := openapi.CalendarFileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewDenyPathStrategy(ignoreEndpoints),
		nil, displayNameOverride,
		func(objectName, fieldName string) bool {
			return fieldName == "items"
		},
	)
	goutils.MustBeNil(err)

	return objects
}
