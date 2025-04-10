package meeting

import (
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/internal/metadatadef"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/zoom"
	"github.com/amp-labs/connectors/providers/zoom/metadata/openapi"
	"github.com/amp-labs/connectors/tools/fileconv/api3"
)

var (
	allowedEndpoints = []string{ // nolint:gochecknoglobals
		"/archive_files",
		"/devices/groups",
		"/devices",
		"/meetings/meeting_summaries",
		"/report/billing",
		"/report/activities",
		"/sip_phones/phones",
		"/tracking_fields",
		"/h323/devices",
	}

	objectEndpoints = map[string]string{ // nolint:gochecknoglobals
		"/devices/groups":             "device_groups",
		"/archive_files":              "archive_files",
		"/meetings/meeting_summaries": "meeting_summaries",
		"/report/billing":             "billing_report",
		"/report/activities":          "activities_report",
		"/h323/devices":               "h323_devices",
	}
)

func Objects() []metadatadef.Schema {
	explorer, err := openapi.MeetingFileManager.GetExplorer(
		api3.WithDisplayNamePostProcessors(
			api3.CamelCaseToSpaceSeparated,
			api3.CapitalizeFirstLetterEveryWord,
		),
	)
	goutils.MustBeNil(err)

	objects, err := explorer.ReadObjectsGet(
		api3.NewAllowPathStrategy(allowedEndpoints),
		objectEndpoints, nil,
		api3.CustomMappingObjectCheck(zoom.ObjectNameToResponseField[providers.ModuleZoomMeeting]),
	)

	goutils.MustBeNil(err)

	return objects
}
